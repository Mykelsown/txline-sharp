package arena

import (
	"fmt"

	"github.com/Mykelsown/txline-sharp/detector"
)

// MomentumAgent is Agent A. It backs the direction of every sharp movement,
// on the theory that sharp money is informed money and the market is right.
//
// Strategy rules:
//   - LOW signal:    SKIP (not enough conviction)
//   - MEDIUM signal: BACK with standard stake
//   - HIGH signal:   BACK with 1.5x stake (higher conviction)
//
// For SHORTENING outcomes (prob going up), backing means taking the
// current (shorter) price, expecting the outcome to win.
//
// For DRIFTING outcomes (prob going down), backing the drift means
// backing the OPPOSING outcome at a longer price, expecting it to win.
type MomentumAgent struct{}

func NewMomentumAgent() *MomentumAgent { return &MomentumAgent{} }

func (a *MomentumAgent) Name() string { return "Agent-A (Momentum)" }

func (a *MomentumAgent) Decide(sig detector.Signal) Decision {
	id := SignalID(sig)

	if sig.Severity == detector.SeverityLow {
		return Decision{
			AgentName:    a.Name(),
			SignalID:     id,
			DecisionType: DecisionSkip,
			Reasoning:    fmt.Sprintf("LOW severity (%.1f%% delta) below momentum threshold. Skipping.", sig.ProbDelta*100),
			TargetPrice:  sig.PriceAfter,
		}
	}

	stake := DefaultStake
	if sig.Severity == detector.SeverityHigh {
		stake = DefaultStake * 1.5
	}

	reasoning := fmt.Sprintf(
		"Sharp %s detected on %s. Prob shifted %.1f%% -> %.1f%% (delta +%.1f%%). "+
			"Following the smart money %s at price %.3f.",
		sig.Direction,
		sig.OutcomeName,
		sig.ProbBefore*100,
		sig.ProbAfter*100,
		sig.ProbDelta*100,
		sig.Direction,
		sig.PriceAfter,
	)

	return Decision{
		AgentName:    a.Name(),
		SignalID:     id,
		DecisionType: DecisionBack,
		Reasoning:    reasoning,
		Stake:        stake,
		TargetPrice:  sig.PriceAfter,
	}
}