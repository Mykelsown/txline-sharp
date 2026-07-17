package arena

import (
	"fmt"

	"github.com/Mykelsown/txline-sharp/detector"
)

// ContrarianAgent is Agent B. It fades every sharp movement, on the theory
// that odds overreact to sharp money and will revert, or that the public
// side of the market is undervalued after a sharp move.
//
// Strategy rules:
//   - LOW signal:    SKIP (not worth fading weak moves)
//   - MEDIUM signal: FADE with standard stake
//   - HIGH signal:   FADE with 1.5x stake (bigger overreaction to exploit)
//
// Fading a SHORTENING outcome means backing AGAINST it (backing the draw
// or opposing team at a now-inflated price).
//
// Fading a DRIFTING outcome means backing the drifted outcome itself,
// now available at a longer, potentially overblown price.
type ContrarianAgent struct{}

func NewContrarianAgent() *ContrarianAgent { return &ContrarianAgent{} }

func (a *ContrarianAgent) Name() string { return "Agent-B (Contrarian)" }

func (a *ContrarianAgent) Decide(sig detector.Signal) Decision {
	id := SignalID(sig)

	if sig.Severity == detector.SeverityLow {
		return Decision{
			AgentName:    a.Name(),
			SignalID:     id,
			DecisionType: DecisionSkip,
			Reasoning:    fmt.Sprintf("LOW severity (%.1f%% delta) not worth fading. Skipping.", sig.ProbDelta*100),
			TargetPrice:  sig.PriceAfter,
		}
	}

	stake := DefaultStake
	if sig.Severity == detector.SeverityHigh {
		stake = DefaultStake * 1.5
	}

	reasoning := fmt.Sprintf(
		"Fading sharp %s on %s. Prob shifted %.1f%% -> %.1f%% (delta +%.1f%%). "+
			"Market overreacted. Taking the other side at price %.3f (before move: %.3f).",
		sig.Direction,
		sig.OutcomeName,
		sig.ProbBefore*100,
		sig.ProbAfter*100,
		sig.ProbDelta*100,
		sig.PriceBefore,
		sig.PriceAfter,
	)

	return Decision{
		AgentName:    a.Name(),
		SignalID:     id,
		DecisionType: DecisionFade,
		Reasoning:    reasoning,
		Stake:        stake,
		TargetPrice:  sig.PriceBefore, // fading at the pre-move price
	}
}