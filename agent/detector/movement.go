package detector

import (
	"math"
	"time"

	"github.com/Mykelsown/txline-sharp/feed"
)

// Detector diffs consecutive odds snapshots and emits Signals when a
// significant implied-probability shift is detected.
type Detector struct {
	threshold float64 // minimum absolute prob delta to trigger (e.g. 0.04)
}

// New constructs a Detector with the given movement threshold.
func New(threshold float64) *Detector {
	return &Detector{threshold: threshold}
}

// Diff compares a previous odds snapshot against a new one for a single fixture
// and returns all Signals whose implied-probability shift exceeds the threshold.
//
// fixture provides match context (team names) for the signal.
// prev is the stored snapshot from the last poll cycle.
// curr is the freshly fetched snapshot.
func (d *Detector) Diff(
	fixture feed.Fixture,
	prev feed.OddsSnapshot,
	curr feed.OddsSnapshot,
) []Signal {
	var signals []Signal
	now := time.Now().UTC()

	for outcomeID, newEntry := range curr {
		oldEntry, existed := prev[outcomeID]
		if !existed {
			// New outcome appeared in the market. Skip, no baseline to diff.
			continue
		}

		probBefore := oldEntry.ImpliedProbability()
		probAfter := newEntry.ImpliedProbability()

		if probBefore == 0 || probAfter == 0 {
			continue
		}

		delta := math.Abs(probAfter - probBefore)
		if delta < d.threshold {
			continue
		}

		direction := "SHORTENING"
		if probAfter < probBefore {
			direction = "DRIFTING"
		}

		sig := Signal{
			FixtureID:   fixture.FixtureID,
			HomeTeam:    fixture.HomeTeam(),
			AwayTeam:    fixture.AwayTeam(),
			MarketName:  newEntry.MarketName,
			OutcomeName: newEntry.OutcomeName,
			OutcomeID:   outcomeID,
			PriceBefore: oldEntry.Price,
			PriceAfter:  newEntry.Price,
			ProbBefore:  probBefore,
			ProbAfter:   probAfter,
			ProbDelta:   delta,
			Severity:    classifySeverity(delta),
			Direction:   direction,
			DetectedAt:  now,
			BlockHash:   newEntry.BlockHash,
			Resolved:    false,
		}

		signals = append(signals, sig)
	}

	return signals
}