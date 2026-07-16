import * as anchor from "@coral-xyz/anchor";
import type { Txoracle } from "./types/txoracle";
import txoracleIdl from "./idl/txoracle.json";
import {
  ASSOCIATED_TOKEN_PROGRAM_ID,
  TOKEN_2022_PROGRAM_ID,
  getAssociatedTokenAddressSync,
} from "@solana/spl-token";
import { Connection, PublicKey, SystemProgram } from "@solana/web3.js";
import axios from "axios";
import nacl from "tweetnacl";
import * as fs from "fs";
import * as path from "path";
import * as dotenv from "dotenv";

dotenv.config();

// ── Config ────────────────────────────────────────────────────────────────────

const API_ORIGIN = "https://txline.txodds.com";
const API_BASE   = `${API_ORIGIN}/api`;
const RPC_URL    = process.env.ANCHOR_PROVIDER_URL ?? "https://api.mainnet-beta.solana.com";
const KEYPAIR_PATH = process.env.ANCHOR_WALLET ?? "";

const PROGRAM_ID   = new PublicKey("9ExbZjAapQww1vfcisDmrngPinHTEfpjYRWMunJgcKaA");
const TXL_MINT     = new PublicKey("Zhw9TVKp68a1QrftncMSd6ELXKDtpVMNuMGr1jNwdeL");

// Service level 12 = real-time World Cup data (free tier, no TxL required)
const SERVICE_LEVEL_ID = 12;
const DURATION_WEEKS   = 4;
const SELECTED_LEAGUES: number[] = [];

// ── Wallet ────────────────────────────────────────────────────────────────────

function loadWallet(keypairPath: string): anchor.web3.Keypair {
  const raw = fs.readFileSync(path.resolve(keypairPath.replace("~", process.env.HOME ?? "")), "utf8");
  const bytes = JSON.parse(raw) as number[];
  return anchor.web3.Keypair.fromSecretKey(Uint8Array.from(bytes));
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function main() {
  console.log("TxLINE Activation Script");
  console.log("========================\n");

  // 1. Load wallet
  if (!KEYPAIR_PATH) {
    throw new Error("ANCHOR_WALLET env var not set");
  }
  const payer = loadWallet(KEYPAIR_PATH);
  console.log(`Wallet:  ${payer.publicKey.toBase58()}`);

  // 2. Set up Anchor provider
  const connection = new Connection(RPC_URL, "confirmed");
  const wallet     = new anchor.Wallet(payer);
  const provider   = new anchor.AnchorProvider(connection, wallet, { commitment: "confirmed" });
  anchor.setProvider(provider);

  const program = new anchor.Program<Txoracle>(txoracleIdl as Txoracle, provider);

  if (!program.programId.equals(PROGRAM_ID)) {
    throw new Error(
      `IDL program ${program.programId.toBase58()} does not match mainnet program ${PROGRAM_ID.toBase58()}`
    );
  }
  console.log(`Program: ${program.programId.toBase58()}`);

  // 3. Check SOL balance
  const balance = await connection.getBalance(payer.publicKey);
  console.log(`Balance: ${balance / anchor.web3.LAMPORTS_PER_SOL} SOL\n`);
  if (balance < 5_000_000) {
    throw new Error("Insufficient SOL. Need at least 0.005 SOL for fees and rent.");
  }

  // 4. Derive PDAs
  const [tokenTreasuryPda] = PublicKey.findProgramAddressSync(
    [Buffer.from("token_treasury_v2")],
    program.programId
  );

  const tokenTreasuryVault = getAssociatedTokenAddressSync(
    TXL_MINT,
    tokenTreasuryPda,
    true,
    TOKEN_2022_PROGRAM_ID,
    ASSOCIATED_TOKEN_PROGRAM_ID
  );

  const [pricingMatrixPda] = PublicKey.findProgramAddressSync(
    [Buffer.from("pricing_matrix")],
    program.programId
  );

  const userTokenAccount = getAssociatedTokenAddressSync(
    TXL_MINT,
    provider.wallet.publicKey,
    false,
    TOKEN_2022_PROGRAM_ID,
    ASSOCIATED_TOKEN_PROGRAM_ID
  );

  // 5. Subscribe on-chain
  console.log(`Subscribing (Service Level ${SERVICE_LEVEL_ID}, ${DURATION_WEEKS} weeks)...`);
  let txSig: string;
  try {
    txSig = await program.methods
      .subscribe(SERVICE_LEVEL_ID, DURATION_WEEKS)
      .accounts({
        user: provider.wallet.publicKey,
        pricingMatrix: pricingMatrixPda,
        tokenMint: TXL_MINT,
        userTokenAccount,
        tokenTreasuryVault,
        tokenTreasuryPda,
        tokenProgram: TOKEN_2022_PROGRAM_ID,
        associatedTokenProgram: ASSOCIATED_TOKEN_PROGRAM_ID,
        systemProgram: SystemProgram.programId,
      })
      .rpc();
  } catch (err: any) {
    // If already subscribed, the program may throw. Surface the message clearly.
    throw new Error(`Subscribe transaction failed: ${err?.message ?? err}`);
  }

  console.log(`Subscribe tx: ${txSig}`);
  console.log("Waiting for confirmation...");
  await connection.confirmTransaction(txSig, "confirmed");
  console.log("Confirmed.\n");

  // 6. Get guest JWT
  console.log("Fetching guest JWT...");
  const authResponse = await axios.post(`${API_ORIGIN}/auth/guest/start`);
  const jwt: string = authResponse.data.token;
  console.log("JWT obtained.\n");

  // 7. Sign activation message
  // For SELECTED_LEAGUES = [], message is: `${txSig}::${jwt}`
  const messageString = `${txSig}:${SELECTED_LEAGUES.join(",")}:${jwt}`;
  const message       = new TextEncoder().encode(messageString);
  const signatureBytes = nacl.sign.detached(message, payer.secretKey);
  const walletSignature = Buffer.from(signatureBytes).toString("base64");

  // 8. Activate API token
  console.log("Activating API token...");
  const activationResponse = await axios.post(
    `${API_BASE}/token/activate`,
    { txSig, walletSignature, leagues: SELECTED_LEAGUES },
    { headers: { Authorization: `Bearer ${jwt}` } }
  );

  const apiToken: string = activationResponse.data.token ?? activationResponse.data;
  console.log("API token activated.\n");

  // 9. Write credentials to file (gitignored)
  const credentials = {
    jwt,
    apiToken,
    walletAddress: payer.publicKey.toBase58(),
    serviceLevel: SERVICE_LEVEL_ID,
    activatedAt: new Date().toISOString(),
    txSig,
  };

  const outPath = path.join(__dirname, "credentials.json");
  fs.writeFileSync(outPath, JSON.stringify(credentials, null, 2));
  console.log(`Credentials saved to: ${outPath}`);

  // 10. Quick sanity check: fetch fixtures
  console.log("\nRunning sanity check (fetching fixtures)...");
  const fixturesRes = await axios.get(`${API_BASE}/fixtures`, {
    headers: {
      Authorization: `Bearer ${jwt}`,
      "X-Api-Token": apiToken,
    },
  });

  const fixtures = fixturesRes.data;
  const count = Array.isArray(fixtures) ? fixtures.length : Object.keys(fixtures).length;
  console.log(`Fixtures received: ${count} items`);
  console.log("\nActivation complete. You are ready to run the Go agent.");
}

main().catch((err) => {
  console.error("\nActivation failed:", err.message ?? err);
  process.exit(1);
});