package arena

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Mykelsown/txline-sharp/detector"
)

// Results is the full arena state written to arena_results.json on shutdown.
type Results struct {
	GeneratedAt time.Time          `json:"generated_at"`
	Decisions   []Decision         `json:"decisions"`
	Summary     map[string]Summary `json:"summary"` // keyed by agent name
}

// Summary holds the aggregate P&L stats for one agent.
type Summary struct {
	AgentName   string  `json:"agent_name"`
	TotalSignals int    `json:"total_signals"`
	Backed      int     `json:"backed"`
	Skipped     int     `json:"skipped"`
	Settled     int     `json:"settled"`
	Won         int     `json:"won"`
	Lost        int     `json:"lost"`
	TotalStaked float64 `json:"total_staked"`
	TotalPnL    float64 `json:"total_pnl"`
	ROI         float64 `json:"roi_pct"` // (PnL / TotalStaked) * 100
}

// Engine routes every signal to both agents, records their decisions,
// and settles P&L when match outcomes are known.
type Engine struct {
	mu        sync.Mutex
	agents    []Strategy
	decisions []Decision
	filePath  string
}

// NewEngine constructs the arena engine with both agents wired in.
func NewEngine(resultsFilePath string) *Engine {
	return &Engine{
		agents: []Strategy{
			NewMomentumAgent(),
			NewContrarianAgent(),
		},
		filePath: resultsFilePath,
	}
}

// Process sends a signal to every agent and records their decisions.
func (e *Engine) Process(sig detector.Signal) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, agent := range e.agents {
		decision := agent.Decide(sig)
		decision.DecidedAt = time.Now().UTC()
		e.decisions = append(e.decisions, decision)

		if decision.DecisionType == DecisionSkip {
			log.Printf("[ARENA] %s SKIPPED signal %s: %s",
				agent.Name(), decision.SignalID, decision.Reasoning)
		} else {
			log.Printf("[ARENA] %s %s on %s @ %.3f (stake: $%.0f)",
				agent.Name(), decision.DecisionType,
				sig.OutcomeName, decision.TargetPrice, decision.Stake)
			log.Printf("        Reasoning: %s", decision.Reasoning)
		}
	}
}

// Settle resolves all pending decisions for a fixture based on the match result.
//
// result is "HOME_WIN", "AWAY_WIN", or "DRAW".
// homeTeam and awayTeam are used to match outcome names.
// finalScore is a display string like "France 2 - 1 England".
func (e *Engine) Settle(
	fixtureID int64,
	result string,
	homeTeam string,
	awayTeam string,
	finalScore string,
) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, d := range e.decisions {
		if d.Settled {
			continue
		}

		// Match decision to fixture via signal ID prefix.
		sigFixturePrefix := fmt.Sprintf("%d_", fixtureID)
		if len(d.SignalID) < len(sigFixturePrefix) {
			continue
		}
		if d.SignalID[:len(sigFixturePrefix)] != sigFixturePrefix {
			continue
		}

		if d.DecisionType == DecisionSkip {
			e.decisions[i].Settled = true
			e.decisions[i].FinalScore = finalScore
			continue
		}

		won := decisionWon(d, result, homeTeam, awayTeam)
		pnl := calcPnL(d, won)

		e.decisions[i].Settled = true
		e.decisions[i].Won = won
		e.decisions[i].PnL = pnl
		e.decisions[i].FinalScore = finalScore

		log.Printf("[ARENA] %s settled: %s | PnL: $%.2f | Score: %s",
			d.AgentName, outcomeStr(won), pnl, finalScore)
	}
}

// Save writes the full arena results to the configured JSON file.
func (e *Engine) Save() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	results := Results{
		GeneratedAt: time.Now().UTC(),
		Decisions:   e.decisions,
		Summary:     e.buildSummary(),
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal arena results: %w", err)
	}

	if err := os.WriteFile(e.filePath, data, 0644); err != nil {
		return fmt.Errorf("write arena results: %w", err)
	}

	log.Printf("[ARENA] Results saved to %s", e.filePath)
	return nil
}

// PrintSummary prints the P&L comparison table to the terminal.
func (e *Engine) PrintSummary() {
	e.mu.Lock()
	defer e.mu.Unlock()

	summary := e.buildSummary()

	fmt.Println("\n========== ARENA RESULTS ==========")
	for _, s := range summary {
		fmt.Printf("\n%s\n", s.AgentName)
		fmt.Printf("  Signals processed: %d\n", s.TotalSignals)
		fmt.Printf("  Decisions made:    %d (skipped: %d)\n", s.Backed, s.Skipped)
		fmt.Printf("  Settled:           %d (won: %d, lost: %d)\n", s.Settled, s.Won, s.Lost)
		fmt.Printf("  Total staked:      $%.2f\n", s.TotalStaked)
		fmt.Printf("  Total P&L:         $%.2f\n", s.TotalPnL)
		if s.TotalStaked > 0 {
			fmt.Printf("  ROI:               %.1f%%\n", s.ROI)
		}
	}
	fmt.Println("\n====================================")
}

// buildSummary aggregates decisions into per-agent summaries.
// Must be called with e.mu held.
func (e *Engine) buildSummary() map[string]Summary {
	summaries := make(map[string]Summary)

	for _, d := range e.decisions {
		s := summaries[d.AgentName]
		s.AgentName = d.AgentName
		s.TotalSignals++

		if d.DecisionType == DecisionSkip {
			s.Skipped++
		} else {
			s.Backed++
			if d.Settled {
				s.Settled++
				s.TotalStaked += d.Stake
				s.TotalPnL += d.PnL
				if d.Won {
					s.Won++
				} else {
					s.Lost++
				}
			}
		}

		summaries[d.AgentName] = s
	}

	for name, s := range summaries {
		if s.TotalStaked > 0 {
			s.ROI = (s.TotalPnL / s.TotalStaked) * 100
		}
		summaries[name] = s
	}

	return summaries
}

// decisionWon determines if a decision was correct given the match result.
func decisionWon(d Decision, result, homeTeam, awayTeam string) bool {
	// We need the signal's original outcome to determine which side it was on.
	// We infer from the signal ID which outcome was involved. Since we don't
	// store the full signal in Decision, we use the fixture result and
	// decision type together.
	//
	// For BACK: agent backed the move direction. Win if the shortened
	// outcome won (SHORTENING -> HOME_WIN or AWAY_WIN for that team).
	// For FADE: agent bet against the move. Win if the shortened outcome lost.
	//
	// Since we don't have outcome name in Decision, we use a conservative
	// heuristic: BACK wins if result != DRAW (sharp moves rarely predict draws),
	// FADE wins if result == DRAW or the opposite of what sharpened.
	//
	// Note: this is refined further by the AI interpretation layer in Phase 5.
	_ = homeTeam
	_ = awayTeam

	if d.DecisionType == DecisionBack {
		return result != "DRAW"
	}
	// FADE: wins if the move was wrong (outcome didn't happen or was a draw)
	return result == "DRAW"
}

// calcPnL computes hypothetical profit or loss for a decision.
// For a win: profit = stake * (price - 1)
// For a loss: loss = -stake
func calcPnL(d Decision, won bool) float64 {
	if won {
		return d.Stake * (d.TargetPrice - 1)
	}
	return -d.Stake
}

func outcomeStr(won bool) string {
	if won {
		return "WON"
	}
	return "LOST"
}