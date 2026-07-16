package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mykelsown/txline-sharp/config"
	"github.com/Mykelsown/txline-sharp/detector"
	"github.com/Mykelsown/txline-sharp/feed"
	"github.com/Mykelsown/txline-sharp/store"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("TxLINE Sharp Movement Detector")
	log.Println("================================")

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
	log.Println()

	client  := feed.NewClient(cfg.Creds.JWT, cfg.Creds.APIToken)
	memory  := store.NewMemory()
	detect  := detector.New(cfg.MovementThreshold)

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

	log.Printf("Agent running. Polling every %ds. Press Ctrl+C to stop.\n", cfg.PollIntervalSec)

	// Run one poll immediately before waiting for the ticker.
	poll(client, memory, detect, fixtures)

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received. Stopping agent.")
			return
		case <-ticker.C:
			poll(client, memory, detect, fixtures)
		}
	}
}

// poll fetches fresh odds for every tracked fixture and runs the detector.
func poll(
	client  *feed.Client,
	memory  *store.Memory,
	detect  *detector.Detector,
	fixtures []feed.Fixture,
) {
	for _, fixture := range fixtures {
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

		// Print signals to terminal (persist layer comes in Phase 4).
		for _, sig := range signals {
			fmt.Printf("\n[SIGNAL] %s | %s vs %s\n", sig.Severity, sig.HomeTeam, sig.AwayTeam)
			fmt.Printf("  Market:    %s\n", sig.MarketName)
			fmt.Printf("  Outcome:   %s\n", sig.OutcomeName)
			fmt.Printf("  Direction: %s\n", sig.Direction)
			fmt.Printf("  Price:     %.3f -> %.3f\n", sig.PriceBefore, sig.PriceAfter)
			fmt.Printf("  Prob:      %.1f%% -> %.1f%% (delta: %.1f%%)\n",
				sig.ProbBefore*100, sig.ProbAfter*100, sig.ProbDelta*100)
			fmt.Printf("  BlockHash: %s\n", sig.BlockHash)
			fmt.Printf("  DetectedAt: %s\n\n", sig.DetectedAt.Format(time.RFC3339))
		}
	}
}