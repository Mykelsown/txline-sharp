package store

import (
	"fmt"
	"log"
	"strings"

	"github.com/Mykelsown/txline-sharp/detector"
	"github.com/Mykelsown/txline-sharp/feed"
)

// OutcomeTracker resolves unresolved signals once a match has finished.
type OutcomeTracker struct {
	filePath string
}

// NewOutcomeTracker constructs a tracker pointing at the given signals file.
func NewOutcomeTracker(filePath string) *OutcomeTracker {
	return &OutcomeTracker{filePath: filePath}
}

// Resolve loads all signals for the given fixture, determines whether each
// signal's direction correctly predicted the match outcome, updates the
// prediction_correct and final_score fields, and rewrites the file.
//
// score is the latest ScoreEntry for the finished fixture.
// fixture provides team name context.
func (t *OutcomeTracker) Resolve(fixture feed.Fixture, score feed.ScoreEntry) error {
	signals, err := LoadAll(t.filePath)
	if err != nil {
		return fmt.Errorf("load signals: %w", err)
	}

	if len(signals) == 0 {
		log.Printf("outcome tracker: no signals to resolve for fixture %d", fixture.FixtureID)
		return nil
	}

	// Build final score string from Stats map.
	// TxLINE Stats keys: "home_score" and "away_score"
	homeScore := score.Stats["home_score"]
	awayScore := score.Stats["away_score"]
	finalScore := fmt.Sprintf("%s %d - %d %s",
		fixture.HomeTeam(), homeScore,
		awayScore, fixture.AwayTeam(),
	)

	// Determine match result from final scores.
	// result is "HOME_WIN", "AWAY_WIN", or "DRAW"
	result := matchResult(homeScore, awayScore)

	resolved := 0
	for i, sig := range signals {
		if sig.FixtureID != fixture.FixtureID || sig.Resolved {
			continue
		}

		correct := predictedCorrectly(sig, result)
		signals[i].Resolved = true
		signals[i].PredictionCorrect = &correct
		signals[i].FinalScore = finalScore
		resolved++
	}

	if resolved == 0 {
		log.Printf("outcome tracker: all signals for fixture %d already resolved", fixture.FixtureID)
		return nil
	}

	if err := WriteAll(t.filePath, signals); err != nil {
		return fmt.Errorf("write resolved signals: %w", err)
	}

	log.Printf("outcome tracker: resolved %d signal(s) for fixture %d. Final: %s",
		resolved, fixture.FixtureID, finalScore)
	return nil
}

// matchResult derives the match outcome from final home and away scores.
func matchResult(homeScore, awayScore int) string {
	switch {
	case homeScore > awayScore:
		return "HOME_WIN"
	case awayScore > homeScore:
		return "AWAY_WIN"
	default:
		return "DRAW"
	}
}

// predictedCorrectly checks whether a signal's direction implied the correct winner.
//
// Logic:
//   - A SHORTENING signal on an outcome means money backed that outcome (prob went up).
//   - A DRIFTING signal means money left that outcome (prob went down).
//   - We check if the shortened outcome actually won.
//
// For "Match Winner" markets, outcome names typically contain the team name.
// We match case-insensitively against home/away team names.
func predictedCorrectly(sig detector.Signal, result string) bool {
	name := strings.ToLower(sig.OutcomeName)

	// Identify which side the outcome refers to.
	// "draw" or "tie" is treated as DRAW.
	outcomeResult := ""
	switch {
	case strings.Contains(name, "draw") || strings.Contains(name, "tie"):
		outcomeResult = "DRAW"
	case strings.Contains(name, strings.ToLower(sig.HomeTeam)):
		outcomeResult = "HOME_WIN"
	case strings.Contains(name, strings.ToLower(sig.AwayTeam)):
		outcomeResult = "AWAY_WIN"
	default:
		// Cannot determine which side, treat as unresolvable.
		return false
	}

	// SHORTENING means the market moved toward this outcome winning.
	// Correct prediction = outcome that shortened actually happened.
	if sig.Direction == "SHORTENING" {
		return outcomeResult == result
	}

	// DRIFTING means money left this outcome.
	// Correct prediction = outcome that drifted did NOT happen.
	return outcomeResult != result
}