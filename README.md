# TxLINE Sharp · World Cup 2026 Sharp Movement Detector

> A fully autonomous real-time odds intelligence agent for World Cup 2026, powered by TxLINE's cryptographically anchored data feed and Solana.

**Live Demo:** https://txline-sharp.vercel.app  
**API Endpoint:** https://txline-sharp-production.up.railway.app  
**Built for:** Superteam × TxODDS Hackathon 2026

---

## What It Does

TxLINE Sharp monitors World Cup 2026 odds every 60 seconds via TxLINE's Service Level 12 real-time feed. When any outcome's implied probability shifts by more than a configurable threshold, the system:

1. Flags it as a sharp movement signal with severity classification (LOW / MEDIUM / HIGH)
2. Anchors the signal to the Solana block hash from TxLINE's on-chain data, creating a cryptographically provable timestamp before the outcome is known
3. Sends the signal to Claude (Anthropic) for professional market analysis
4. Routes it to two competing arena agents running opposite strategies
5. Persists everything to an append-only signal log for outcome tracking

Once deployed, the system runs entirely autonomously with zero human intervention.

---

## Core Idea

Sharp money moves odds before outcomes happen. Professional bettors and syndicates place large positions that compress or expand prices in ways that are statistically detectable. TxLINE Sharp surfaces those movements in real time, interprets them with AI, and tracks whether they predicted the result.

The key innovation is the on-chain anchor. Every signal carries a Solana block hash from TxLINE's cryptographic data layer, proving the signal was recorded at a specific point in time before the match ended. This makes the entire signal history tamper-proof and auditable, something no traditional sports analytics tool provides.

---

## Architecture

```
TxLINE API (Service Level 12, real-time)
         │
         ▼
   Go Agent (Railway)
         │
   ┌─────┴──────┐
   ▼            ▼
Detector      Score Tracker
(implied      (resolves signals
prob diff)     after match ends)
   │
   ├── signals.jsonl (append-only log)
   │
   ├── Claude API (signal interpretation)
   │
   └── Arena Engine
         ├── Agent A: Momentum (follows sharp money)
         └── Agent B: Contrarian (fades the move)
              │
              └── arena_results.json
                       │
                       ▼
              REST API (:8081)
                       │
                       ▼
              React Frontend (Vercel)
```

### Stack

| Layer | Technology |
|---|---|
| On-chain subscription | TypeScript, Anchor, @solana/web3.js |
| Detection engine | Go 1.22 |
| AI interpretation | Claude Sonnet 4.6 (Anthropic SDK) |
| HTTP API | Go net/http |
| Frontend | React 18, TypeScript, Vite |
| Agent deployment | Railway (Docker) |
| Frontend deployment | Vercel |
| Data anchor | Solana mainnet, TxLINE block hashes |

---

## TxLINE Endpoints Used

| Endpoint | Purpose |
|---|---|
| `POST /auth/guest/start` | Guest JWT for API access |
| `GET /api/fixtures/snapshot` | Fetch all World Cup fixtures in bundle |
| `GET /api/odds/snapshot/:fixtureId` | Current odds for a fixture (polled every 60s) |
| `GET /api/scores/snapshot/:fixtureId` | Match score and status for outcome resolution |

On-chain: `subscribe(serviceLevel=12, durationWeeks=4)` on the TxOracle program (`9ExbZjAapQww1vfcisDmrngPinHTEfpjYRWMunJgcKaA`) on Solana mainnet.

---

## Signal Detection Logic

For each poll cycle, the agent:

1. Fetches the full odds snapshot for every live fixture
2. Flattens the multi-outcome market records into individual outcome entries
3. Computes implied probability for each outcome: `prob = 1 / decimalOdds`
4. Diffs against the previous snapshot stored in memory
5. Flags any outcome where `|probAfter - probBefore| >= threshold`

**Severity classification:**
- `LOW`: 4-7% probability shift
- `MEDIUM`: 7-12% probability shift
- `HIGH`: 12%+ probability shift

**Direction:**
- `SHORTENING`: implied probability increased (money backing this outcome)
- `DRIFTING`: implied probability decreased (money leaving this outcome)

---

## Arena Agents

Two agents run in parallel, receiving every signal simultaneously:

**Agent A (Momentum):** Follows the sharp move. Backs whichever outcome shortened, on the theory that sharp money is informed money. Skips LOW severity signals. Stakes $100 on MEDIUM, $150 on HIGH.

**Agent B (Contrarian):** Fades the sharp move. Takes the opposing position, on the theory that odds overreact to sharp positioning. Same severity thresholds and stake sizing as Agent A.

Both agents record their reasoning for each decision. P&L is tracked hypothetically across all settled matches and the results are written to `arena_results.json` on every poll cycle.

---

## API Reference

All endpoints return JSON. CORS is open for judge review.

```
GET /health              → liveness probe
GET /api/status          → agent runtime state
GET /api/fixtures        → tracked World Cup fixtures
GET /api/signals         → all logged signals (from signals.jsonl)
GET /api/arena           → arena agent decisions and P&L summary
```

**Example `/api/status` response:**
```json
{
  "wallet": "9GLP7Ja585pHi2pisCtWQVokX37AnJCNL5G6EZf7JQth",
  "service_level": 12,
  "activated_at": "2026-07-16T10:59:55.175Z",
  "poll_interval_sec": 60,
  "movement_threshold": 0.04,
  "is_running": true,
  "ai_interpreter_enabled": true,
  "total_signals": 47,
  "last_poll": "2026-07-19T18:45:00Z"
}
```

---

## Running Locally

### Prerequisites
- Go 1.22+
- Node.js 20+
- Solana CLI
- Funded Solana mainnet wallet

### Step 1: Activate TxLINE subscription (one-time)

```bash
cd setup
npm install
cp .env.example .env
# Fill in ANCHOR_WALLET and ANCHOR_PROVIDER_URL
npx ts-node activate.ts
```

This submits one Solana transaction (~0.003 SOL) and writes `credentials.json`.

### Step 2: Run the Go agent

```bash
cd agent
cp .env.example .env
# Fill in ANTHROPIC_API_KEY
go mod tidy
go run .
```

Agent starts polling and exposes the API at `http://localhost:8081`.

### Step 3: Run the frontend

```bash
cd txline-frontend
npm install
# Create .env.local with VITE_API_URL=http://localhost:8081
npm run dev
```

Open `http://localhost:5173`.

### Docker (Linux / production)

```bash
# From project root
docker compose up --build
```

Frontend at `http://localhost:3000`, agent API at `http://localhost:8081`.

---

## Environment Variables

### Agent (`agent/.env`)

| Variable | Description | Default |
|---|---|---|
| `ANTHROPIC_API_KEY` | Claude API key for signal interpretation | required |
| `POLL_INTERVAL_SEC` | How often to poll odds in seconds | `60` |
| `MOVEMENT_THRESHOLD` | Minimum implied prob delta to trigger signal | `0.04` |
| `CREDENTIALS_JSON` | Full credentials JSON (Railway/cloud deployments) | optional |
| `SIGNALS_FILE` | Path to signals log | `signals.jsonl` |
| `ARENA_RESULTS_FILE` | Path to arena results | `arena_results.json` |
| `API_ADDR` | HTTP server bind address | `:8081` |

### Frontend (`txline-frontend/.env.local`)

| Variable | Description |
|---|---|
| `VITE_API_URL` | Go agent base URL (e.g. `https://your-agent.railway.app`) |

---

## Project Structure

```
txline-sharp/
  setup/                    TypeScript: one-time Solana activation
    activate.ts
    create-token-account.ts
    credentials.json         (gitignored)

  agent/                    Go: autonomous detection agent
    main.go
    config/config.go         env + credentials loader
    feed/
      client.go              TxLINE HTTP client with JWT renewal
      stream.go              SSE stream reader
      types.go               Go structs matching TxLINE schema
    detector/
      movement.go            implied probability diff logic
      signal.go              Signal struct + severity classification
    store/
      memory.go              in-memory odds snapshot ring buffer
      persist.go             append-only signals.jsonl writer
      outcome.go             post-match signal resolution
    arena/
      agent.go               Strategy interface + Decision struct
      momentum.go            Agent A: follows sharp money
      contrarian.go          Agent B: fades the move
      engine.go              routes signals, tracks P&L, saves results
      interpreter.go         Claude AI signal interpretation
    api/
      server.go              REST API server

  txline-frontend/          React + TypeScript frontend
    src/
      components/
        TopBar.tsx            agent status + live/demo indicator
        FixturePanel.tsx      fixture list + match timer
        SignalFeed.tsx        filterable signal stream
        SignalCard.tsx        individual signal with probability bar
        ArenaPanel.tsx        agent P&L comparison
        ProbabilityBar.tsx    animated odds movement visualization
      hooks/
        useLiveData.ts        live API polling hook
      types/index.ts          TypeScript types matching Go structs
      data/mock.ts            demo fallback data

  docker-compose.yml        one-command production deployment
```

---

## TxLINE API Feedback

**What worked well:**
The normalized JSON schema across all competition types is genuinely excellent. The `SuperOddsType` + `PriceNames` + `Prices` array pattern is clean and consistent, making it straightforward to flatten into per-outcome records for analysis. The cryptographic block hash on every odds record is the standout feature, it's what makes this project's audit trail possible and is something I haven't seen from any other sports data provider.

The Service Level 12 real-time feed performed reliably throughout the tournament window with sub-second update latency during live matches.

**Where friction appeared:**
The odds snapshot endpoint wraps its response in `{"value": [...], "Count": N}` while the documentation examples showed a raw array. This caused an initial parsing failure that took time to debug. Consistent response envelope across all endpoints would help.

The `Prices` field uses integer scaling by 1000 (so `1500` means `1.500` decimal odds) but this isn't documented explicitly. It had to be inferred from the `Pct` percentage values. A note in the schema docs would save builders time.

Guest JWT expiry behavior during long-running agents could also be more explicit in the docs. The renewal-on-401 pattern works but the TTL isn't documented.

---

## Submission Checklist

- [x] Demo video
- [x] Public GitHub repo: github.com/Mykelsown/txline-sharp
- [x] Live deployment: https://txline-sharp.vercel.app
- [x] Functional API endpoint: https://txline-sharp-production.up.railway.app/health
- [x] Technical documentation (this README)
- [x] TxLINE API feedback (above)

---

*Built by Samuel Micheal Pelumi ([@Mykelsown](https://github.com/Mykelsown)) for the Superteam × TxODDS World Cup 2026 Hackathon.*