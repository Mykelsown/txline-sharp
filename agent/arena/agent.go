package arena

import (
	"fmt"
	"time"

	"github.com/Mykelsown/txline-sharp/detector"
)

// Strategy defines how an agent responds to a sharp movement signal.
type Strategy interface {
	Name() string
	Decide(sig detector.Signal) Decision
}

// DecisionType is the action an agent takes on a signal.
type DecisionType string

const (
	DecisionBack  DecisionType = "BACK"  // bet in the direction of the move
	DecisionFade  DecisionType = "FADE"  // bet against the direction of the move
	DecisionSkip  DecisionType = "SKIP"  // signal not strong enough, no action
)

// Decision is an agent's response to a single signal.
type Decision struct {
	AgentName    string       `json:"agent_name"`
	SignalID     string       `json:"signal_id"` // fixture_id + outcome_id + timestamp
	DecisionType DecisionType `json:"decision_type"`
	Reasoning    string       `json:"reasoning"`
	Stake        float64      `json:"stake"`       // hypothetical stake in USDT
	TargetPrice  float64      `json:"target_price"` // price agent is betting at
	DecidedAt    time.Time    `json:"decided_at"`

	// Settled after match ends
	Settled       bool     `json:"settled"`
	Won           bool     `json:"won"`
	PnL           float64  `json:"pnl"`           // profit or loss on this decision
	FinalScore    string   `json:"final_score,omitempty"`
}

// SignalID builds a unique string ID for a signal.
func SignalID(sig detector.Signal) string {
	return fmt.Sprintf("%d_%s_%s",
		sig.FixtureID,
		sig.OutcomeKey,
		sig.DetectedAt.Format("20060102T150405"),
	)
}

// DefaultStake is the hypothetical stake each agent places per decision.
const DefaultStake = 100.0 // USDT
