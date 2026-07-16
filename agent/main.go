package main

import (
	"fmt"
	"log"
	"os"

	"time"

	"github.com/Mykelsown/txline-sharp/config"
	"github.com/Mykelsown/txline-sharp/feed"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("TxLINE Sharp Movement Detector")
	log.Println("================================")

	// Load config and credentials.
	credPath := "../setup/credentials.json"
	if v := os.Getenv("CREDENTIALS_FILE"); v != "" {
		credPath = v
	}

	cfg, err := config.Load(credPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("Wallet:           %s", cfg.Creds.WalletAddress)
	log.Printf("Service Level:    %d", cfg.Creds.ServiceLevel)
	log.Printf("Activated At:     %s", cfg.Creds.ActivatedAt)
	log.Printf("Poll Interval:    %ds", cfg.PollIntervalSec)
	log.Printf("Movement Threshold: %.0f%%", cfg.MovementThreshold*100)
	log.Println()

	// Connectivity test: fetch World Cup fixtures.
	client := feed.NewClient(cfg.Creds.JWT, cfg.Creds.APIToken)

	log.Println("Testing API connectivity (fetching fixtures)...")
	fixtures, err := client.Fixtures()
	if err != nil {
		log.Fatalf("connectivity test failed: %v", err)
	}

	fmt.Printf("\nWorld Cup fixtures found: %d\n\n", len(fixtures))
	limit := 5
	if len(fixtures) < limit {
		limit = len(fixtures)
	}
	for i, f := range fixtures[:limit] {
		kickoff := time.Unix(f.StartTime, 0).UTC().Format("Jan 02 15:04 UTC")
		fmt.Printf("  %d. %s vs %s (ID: %d, Kickoff: %s, GameState: %d)\n",
			i+1,
			f.HomeTeam(),
			f.AwayTeam(),
			f.FixtureID,
			kickoff,
			f.GameState,
		)
	}
	if len(fixtures) > limit {
		fmt.Printf("  ... and %d more\n", len(fixtures)-limit)
	}

	fmt.Println("\nPhase 2 complete. Feed layer is operational.")
	fmt.Println("Next: detector + memory store (Phase 3).")
}