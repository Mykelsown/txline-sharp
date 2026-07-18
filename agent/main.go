package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/Mykelsown/txline-sharp/arena"
	"github.com/Mykelsown/txline-sharp/config"
	"github.com/Mykelsown/txline-sharp/detector"
	"github.com/Mykelsown/txline-sharp/feed"
	"github.com/Mykelsown/txline-sharp/store"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	// Load .env file if it exists. Silently ignored if not found,
	// so environment variables set externally still work.
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded .env file")
	}

	log.Println("TxLINE Sharp Movement Detector + Arena")
	log.Println("=======================================")

	credPath := "../setup/credentials.json"
	if v := os.Getenv("CREDENTIALS_FILE"); v != "" {
		credPath = v
	}

	cfg, err := config.Load(credPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("Wallet:             %s", cfg.Creds.WalletAddress)
	log.Printf("Service Level:      %d", cfg.Creds.ServiceLevel)
	log.Printf("Poll Interval:      %ds", cfg.PollIntervalSec)
	log.Printf("Movement Threshold: %.0f%%", cfg.MovementThreshold*100)
	log.Printf("Signals File:       %s", cfg.SignalsFile)
	log.Printf("Arena Results:      %s", cfg.ArenaResultsFile)

	// AI interpreter (optional, requires ANTHROPIC_API_KEY in .env or environment).
	interpreter := arena.NewInterpreter()
	if interpreter != nil {
		log.Println("AI Interpreter:     enabled")
	} else {
		log.Println("AI Interpreter:     disabled (set ANTHROPIC_API_KEY in .env to enable)")
	}
	log.Println()

	// Initialize core components.
	client  := feed.NewClient(cfg.Creds.JWT, cfg.Creds.APIToken)
	memory  := store.NewMemory()
	detect  := detector.New(cfg.MovementThreshold)
	tracker := store.NewOutcomeTracker(cfg.SignalsFile)
	engine  := arena.NewEngine(cfg.ArenaResultsFile)

	// Open persist layer (creates signals.jsonl if it doesn't exist).
	persist, err := store.NewPersist(cfg.SignalsFile)
	if err != nil {
		log.Fatalf("persist: %v", err)
	}
	defer persist.Close()

	// Load any signals logged in previous runs.
	existing, err := store.LoadAll(cfg.SignalsFile)
	if err != nil {
		log.Fatalf("load existing signals: %v", err)
	}
	log.Printf("Loaded %d existing signal(s) from previous runs.", len(existing))

	// Fetch football fixtures once at startup.
	log.Println("Fetching World Cup fixtures...")
	allFixtures, err := client.Fixtures()
	if err != nil {
		log.Fatalf("fixtures: %v", err)
	}

	var fixtures []feed.Fixture
	for _, f := range allFixtures {
		if f.IsFootball() {
			fixtures = append(fixtures, f)
		}
	}

	if len(fixtures) == 0 {
		log.Fatal("No World Cup football fixtures found in bundle.")
	}

	fmt.Printf("\nTracking %d World Cup fixture(s):\n", len(fixtures))
	for _, f := range fixtures {
		fmt.Printf("  - %s vs %s (ID: %d, Kickoff: %s)\n",
			f.HomeTeam(), f.AwayTeam(),
			f.FixtureID,
			f.StartTime().Format("2006-01-02 15:04 UTC"),
		)
	}
	fmt.Println()

	// Graceful shutdown on SIGINT / SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(time.Duration(cfg.PollIntervalSec) * time.Second)
	defer ticker.Stop()

	log.Printf("Agent running. Polling every %ds. Press Ctrl+C to stop.", cfg.PollIntervalSec)

	// Run one poll immediately before waiting for the ticker.
	poll(client, memory, detect, persist, tracker, engine, interpreter, fixtures)

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received. Stopping agent.")
			engine.PrintSummary()
			if err := engine.Save(); err != nil {
				log.Printf("arena save error: %v", err)
			}
			printSummary(cfg.SignalsFile)
			return
		case <-ticker.C:
			poll(client, memory, detect, persist, tracker, engine, interpreter, fixtures)
		}
	}
}

// poll fetches fresh odds and scores for every tracked fixture,
// runs the detector, persists signals, routes them to the arena,
// and resolves finished matches.
func poll(
	client      *feed.Client,
	memory      *store.Memory,
	detect      *detector.Detector,
	persist     *store.Persist,
	tracker     *store.OutcomeTracker,
	engine      *arena.Engine,
	interpreter *arena.Interpreter,
	fixtures    []feed.Fixture,
) {
	for _, fixture := range fixtures {
		// Check scores and resolve finished matches.
		scores, err := client.ScoresSnapshot(fixture.FixtureID)
		if err != nil {
			log.Printf("scores fetch error (fixture %d): %v", fixture.FixtureID, err)
		}

		for _, score := range scores {
			if score.IsFinished() {
				homeScore := score.Stats["home_score"]
				awayScore := score.Stats["away_score"]
				finalScore := fmt.Sprintf("%s %d - %d %s",
					fixture.HomeTeam(), homeScore, awayScore, fixture.AwayTeam())

				result := "DRAW"
				if homeScore > awayScore {
					result = "HOME_WIN"
				} else if awayScore > homeScore {
					result = "AWAY_WIN"
				}

				if err := tracker.Resolve(fixture, score); err != nil {
					log.Printf("outcome resolve error: %v", err)
				}
				engine.Settle(fixture.FixtureID, result,
					fixture.HomeTeam(), fixture.AwayTeam(), finalScore)
				break
			}
		}

		// Fetch current odds snapshot.
		entries, err := client.OddsSnapshot(fixture.FixtureID)
		if err != nil {
			log.Printf("odds fetch error (fixture %d): %v", fixture.FixtureID, err)
			continue
		}

		if len(entries) == 0 {
			log.Printf("fixture %d (%s vs %s): no odds available yet",
				fixture.FixtureID, fixture.HomeTeam(), fixture.AwayTeam())
			continue
		}

		curr := store.BuildSnapshot(entries)
		prev, hasPrev := memory.Get(fixture.FixtureID)
		memory.Set(fixture.FixtureID, curr)

		if !hasPrev {
			log.Printf("fixture %d (%s vs %s): baseline stored (%d outcomes)",
				fixture.FixtureID, fixture.HomeTeam(), fixture.AwayTeam(), len(curr))
			continue
		}

		signals := detect.Diff(fixture, prev, curr)
		if len(signals) == 0 {
			log.Printf("fixture %d (%s vs %s): %d outcomes checked, no movement",
				fixture.FixtureID, fixture.HomeTeam(), fixture.AwayTeam(), len(curr))
			continue
		}

		for _, sig := range signals {
			// 1. Persist to signals.jsonl.
			if err := persist.Append(sig); err != nil {
				log.Printf("persist error: %v", err)
			}

			// 2. Print signal to terminal.
			printSignal(sig)

			// 3. AI interpretation (if enabled).
			if interpreter != nil {
				log.Println("Requesting AI interpretation...")
				commentary := interpreter.Interpret(sig)
				fmt.Printf("  AI Analysis: %s\n\n", commentary)
			}

			// 4. Route to arena agents.
			engine.Process(sig)
		}
	}
}

// printSignal formats and prints a single signal to the terminal.
func printSignal(sig detector.Signal) {
	fmt.Printf("\n[%s SIGNAL] %s vs %s\n", sig.Severity, sig.HomeTeam, sig.AwayTeam)
	fmt.Printf("  Market:     %s\n", sig.MarketName)
	fmt.Printf("  Outcome:    %s\n", sig.OutcomeName)
	fmt.Printf("  Direction:  %s\n", sig.Direction)
	fmt.Printf("  Price:      %.3f -> %.3f\n", sig.PriceBefore, sig.PriceAfter)
	fmt.Printf("  Prob:       %.1f%% -> %.1f%% (delta: +%.1f%%)\n",
		sig.ProbBefore*100, sig.ProbAfter*100, sig.ProbDelta*100)
	fmt.Printf("  BlockHash:  %s\n", sig.BlockHash)
	fmt.Printf("  DetectedAt: %s\n", sig.DetectedAt.Format(time.RFC3339))
}

// printSummary prints a final signal accuracy table on shutdown.
func printSummary(filePath string) {
	signals, err := store.LoadAll(filePath)
	if err != nil || len(signals) == 0 {
		log.Println("No signals logged this session.")
		return
	}

	total, resolved, correct := len(signals), 0, 0
	for _, s := range signals {
		if s.Resolved {
			resolved++
			if s.PredictionCorrect != nil && *s.PredictionCorrect {
				correct++
			}
		}
	}

	fmt.Println("\n========== SIGNAL SUMMARY ==========")
	fmt.Printf("Total signals logged: %d\n", total)
	fmt.Printf("Resolved:             %d\n", resolved)
	if resolved > 0 {
		fmt.Printf("Correct predictions:  %d / %d (%.0f%%)\n",
			correct, resolved, float64(correct)/float64(resolved)*100)
	}
	fmt.Println("=====================================")
}