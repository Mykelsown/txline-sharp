import * as anchor from "@coral-xyz/anchor";
import {
  ASSOCIATED_TOKEN_PROGRAM_ID,
  TOKEN_2022_PROGRAM_ID,
  getAssociatedTokenAddressSync,
  createAssociatedTokenAccountInstruction,
  getAccount,
} from "@solana/spl-token";
import { Connection, PublicKey, Transaction, sendAndConfirmTransaction } from "@solana/web3.js";
import * as fs from "fs";
import * as path from "path";
import * as dotenv from "dotenv";

dotenv.config();

const RPC_URL   = process.env.ANCHOR_PROVIDER_URL ?? "https://api.mainnet-beta.solana.com";
const TXL_MINT  = new PublicKey("Zhw9TVKp68a1QrftncMSd6ELXKDtpVMNuMGr1jNwdeL");

function loadKeypair(p: string): anchor.web3.Keypair {
  const raw = fs.readFileSync(path.resolve(p.replace("~", process.env.HOME ?? "")), "utf8");
  return anchor.web3.Keypair.fromSecretKey(Uint8Array.from(JSON.parse(raw) as number[]));
}

async function main() {
  const payer      = loadKeypair(process.env.ANCHOR_WALLET ?? "");
  const connection = new Connection(RPC_URL, "confirmed");

  const ata = getAssociatedTokenAddressSync(
    TXL_MINT,
    payer.publicKey,
    false,
    TOKEN_2022_PROGRAM_ID,
    ASSOCIATED_TOKEN_PROGRAM_ID
  );

  console.log(`Wallet:  ${payer.publicKey.toBase58()}`);
  console.log(`ATA:     ${ata.toBase58()}`);

  // Check if it already exists
  try {
    await getAccount(connection, ata, "confirmed", TOKEN_2022_PROGRAM_ID);
    console.log("Token account already exists. Nothing to do.");
    return;
  } catch {
    console.log("Token account does not exist. Creating...");
  }

  const ix = createAssociatedTokenAccountInstruction(
    payer.publicKey,
    ata,
    payer.publicKey,
    TXL_MINT,
    TOKEN_2022_PROGRAM_ID,
    ASSOCIATED_TOKEN_PROGRAM_ID
  );

  const tx = new Transaction().add(ix);
  const sig = await sendAndConfirmTransaction(connection, tx, [payer], { commitment: "confirmed" });
  console.log(`Created. Tx: ${sig}`);
  console.log("Token account ready. Run activate.ts now.");
}

main().catch((err) => {
  console.error("Failed:", err.message ?? err);
  process.exit(1);
});
