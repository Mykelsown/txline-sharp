package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Credentials mirrors the shape written by setup/activate.ts.
type Credentials struct {
	JWT          string `json:"jwt"`
	APIToken     string `json:"apiToken"`
	WalletAddress string `json:"walletAddress"`
	ServiceLevel int    `json:"serviceLevel"`
	ActivatedAt  string `json:"activatedAt"`
	TxSig        string `json:"txSig"`
}

// Config holds all runtime configuration for the agent.
type Config struct {
	// Credentials loaded from credentials.json
	Creds Credentials

	// PollIntervalSec is how often the feed poller snapshots odds.
	// Defaults to 60 seconds; set POLL_INTERVAL_SEC to override.
	PollIntervalSec int

	// MovementThreshold is the minimum implied probability shift (0.0 to 1.0)
	// required to trigger a sharp movement signal.
	// Defaults to 0.04 (4%); set MOVEMENT_THRESHOLD to override.
	MovementThreshold float64

	// SignalsFile is the path to the append-only JSONL signal log.
	SignalsFile string

	// ArenaResultsFile is written on shutdown with the arena P&L summary.
	ArenaResultsFile string

	// DiscordWebhookURL is optional. If set, HIGH signals are POSTed there.
	DiscordWebhookURL string
}

// Load reads credentials from credentialsPath and applies env var overrides.
func Load(credentialsPath string) (*Config, error) {
	absPath, err := filepath.Abs(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("resolve credentials path: %w", err)
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials file %s: %w", absPath, err)
	}

	var creds Credentials
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("decode credentials: %w", err)
	}

	if creds.JWT == "" || creds.APIToken == "" {
		return nil, fmt.Errorf("credentials file is missing jwt or apiToken")
	}

	cfg := &Config{
		Creds:             creds,
		PollIntervalSec:   60,
		MovementThreshold: 0.04,
		SignalsFile:       "signals.jsonl",
		ArenaResultsFile:  "arena_results.json",
	}

	if v := os.Getenv("POLL_INTERVAL_SEC"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 5 {
			return nil, fmt.Errorf("POLL_INTERVAL_SEC must be an integer >= 5, got %q", v)
		}
		cfg.PollIntervalSec = n
	}

	if v := os.Getenv("MOVEMENT_THRESHOLD"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f <= 0 || f >= 1 {
			return nil, fmt.Errorf("MOVEMENT_THRESHOLD must be a float between 0 and 1, got %q", v)
		}
		cfg.MovementThreshold = f
	}

	if v := os.Getenv("SIGNALS_FILE"); v != "" {
		cfg.SignalsFile = v
	}

	if v := os.Getenv("ARENA_RESULTS_FILE"); v != "" {
		cfg.ArenaResultsFile = v
	}

	cfg.DiscordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")

	return cfg, nil
}