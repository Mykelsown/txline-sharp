package detector

import "time"

// Severity classifies how significant a sharp movement signal is.
type Severity string

const (
	SeverityLow    Severity = "LOW"
	SeverityMedium Severity = "MEDIUM"
	SeverityHigh   Severity = "HIGH"
)

// Signal represents a detected sharp odds movement on a single outcome.
type Signal struct {
	// Identity
	FixtureID   int64  `json:"fixture_id"`
	HomeTeam    string `json:"home_team"`
	AwayTeam    string `json:"away_team"`
	MarketName  string `json:"market_name"`
	OutcomeName string `json:"outcome_name"`
	OutcomeKey  string `json:"outcome_key"` // stable key: SuperOddsType|period|params|priceName
	InRunning   bool   `json:"in_running"`

	// Odds movement
	PriceBefore float64 `json:"price_before"`
	PriceAfter  float64 `json:"price_after"`
	ProbBefore  float64 `json:"prob_before"`
	ProbAfter   float64 `json:"prob_after"`
	ProbDelta   float64 `json:"prob_delta"`

	// Classification
	Severity  Severity `json:"severity"`
	Direction string   `json:"direction"` // "SHORTENING" or "DRIFTING"

	// Timing and on-chain anchor
	DetectedAt time.Time `json:"detected_at"`
	BlockHash  string    `json:"block_hash"`

	// Outcome tracking (filled after match ends)
	Resolved          bool    `json:"resolved"`
	PredictionCorrect *bool   `json:"prediction_correct,omitempty"`
	FinalScore        string  `json:"final_score,omitempty"`
	AIAnalysis        string  `json:"ai_analysis,omitempty"`
}

// classifySeverity assigns severity based on absolute probability delta.
func classifySeverity(delta float64) Severity {
	switch {
	case delta >= 0.12:
		return SeverityHigh
	case delta >= 0.07:
		return SeverityMedium
	default:
		return SeverityLow
	}
}
