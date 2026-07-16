package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Mykelsown/txline-sharp/config"
	"github.com/Mykelsown/txline-sharp/feed"
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
	log.Printf("Activated At:       %s", cfg.Creds.ActivatedAt)
	log.Printf("Poll Interval:      %ds", cfg.PollIntervalSec)
	log.Printf("Movement Threshold: %.0f%%", cfg.MovementThreshold*100)
	log.Println()

	client := feed.NewClient(cfg.Creds.JWT, cfg.Creds.APIToken)

	log.Println("Testing API connectivity (fetching fixtures)...")
	fixtures, err := client.Fixtures()
	if err != nil {
		log.Fatalf("connectivity test failed: %v", err)
	}

	// Filter to football (World Cup, competition 72) only
	var footballFixtures []feed.Fixture
	for _, f := range fixtures {
		if f.IsFootball() {
			footballFixtures = append(footballFixtures, f)
		}
	}

	fmt.Printf("\nTotal fixtures in bundle: %d\n", len(fixtures))
	fmt.Printf("World Cup football fixtures: %d\n\n", len(footballFixtures))

	for i, f := range footballFixtures {
		fmt.Printf("  %d. %s vs %s\n", i+1, f.HomeTeam(), f.AwayTeam())
		fmt.Printf("     ID: %d | Kickoff: %s | GameState: %d\n",
			f.FixtureID,
			f.StartTime().Format("2006-01-02 15:04 UTC"),
			f.GameState,
		)
	}

	fmt.Println("\nPhase 2 complete. Feed layer is operational.")
	fmt.Println("Next: detector + memory store (Phase 3).")
}