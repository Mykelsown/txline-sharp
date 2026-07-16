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
	OutcomeID   int64  `json:"outcome_id"`

	// Odds movement
	PriceBefore float64 `json:"price_before"`
	PriceAfter  float64 `json:"price_after"`
	ProbBefore  float64 `json:"prob_before"`  // implied probability before
	ProbAfter   float64 `json:"prob_after"`   // implied probability after
	ProbDelta   float64 `json:"prob_delta"`   // absolute shift (positive = shortening)

	// Classification
	Severity  Severity `json:"severity"`
	Direction string   `json:"direction"` // "SHORTENING" or "DRIFTING"

	// Timing and on-chain anchor
	DetectedAt time.Time `json:"detected_at"`
	BlockHash  string    `json:"block_hash"` // cryptographic anchor from TxLINE

	// Outcome tracking (filled in after match ends)
	Resolved        bool   `json:"resolved"`
	PredictionCorrect *bool `json:"prediction_correct,omitempty"`
	FinalScore      string `json:"final_score,omitempty"`
}

// classifySeverity assigns a severity based on the absolute probability delta.
//   < 4%  : should not reach here (filtered by threshold)
//   4-7%  : LOW
//   7-12% : MEDIUM
//   > 12% : HIGH
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