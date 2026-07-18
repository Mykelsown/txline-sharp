package feed

import "time"

// Fixture represents a single match from /api/fixtures/snapshot
type Fixture struct {
	FixtureID          int64  `json:"FixtureId"`
	CompetitionID      int64  `json:"CompetitionId"`
	Participant1       string `json:"Participant1"`
	Participant2       string `json:"Participant2"`
	Participant1IsHome bool   `json:"Participant1IsHome"`
	StartTimeMs        int64  `json:"StartTime"` // Unix timestamp in milliseconds
	GameState          int    `json:"GameState"`
	StatusID           int    `json:"StatusId"`
}

// StartTime returns the fixture start time as a time.Time.
func (f Fixture) StartTime() time.Time {
	return time.Unix(f.StartTimeMs/1000, 0).UTC()
}

// HomeTeam returns the home team name.
func (f Fixture) HomeTeam() string {
	if f.Participant1IsHome {
		return f.Participant1
	}
	return f.Participant2
}

// AwayTeam returns the away team name.
func (f Fixture) AwayTeam() string {
	if f.Participant1IsHome {
		return f.Participant2
	}
	return f.Participant1
}

// IsFootball returns true for FIFA World Cup football fixtures (competition 72).
func (f Fixture) IsFootball() bool {
	return f.CompetitionID == 72
}

// OddsRecord is the raw shape returned by /api/odds/snapshot/:fixtureId.
// Each record contains multiple outcomes (PriceNames/Prices arrays).
type OddsRecord struct {
	FixtureID        int64    `json:"FixtureId"`
	MessageID        string   `json:"MessageId"`
	Ts               int64    `json:"Ts"`
	SuperOddsType    string   `json:"SuperOddsType"`
	MarketPeriod     *string  `json:"MarketPeriod"`
	MarketParameters *string  `json:"MarketParameters"`
	InRunning        bool     `json:"InRunning"`
	PriceNames       []string `json:"PriceNames"`
	Prices           []int64  `json:"Prices"`  // scaled by 1000, divide to get decimal odds
	Pct              []string `json:"Pct"`
	BlockHash        string   `json:"BlockHash"`
}

// OddsEntry is a single flattened outcome extracted from an OddsRecord.
// This is the unit the detector and memory store operate on.
type OddsEntry struct {
	// Identity
	FixtureID    int64   `json:"FixtureId"`
	OutcomeKey   string  `json:"OutcomeKey"`   // unique: MessageId + "_" + PriceName
	MarketName   string  `json:"MarketName"`   // SuperOddsType + MarketParameters
	OutcomeName  string  `json:"OutcomeName"`  // PriceName (e.g. "part1", "draw", "part2")
	MarketPeriod string  `json:"MarketPeriod"` // e.g. "half=1" or "full"

	// Pricing
	Price     float64 `json:"Price"`     // decimal odds (Prices[i] / 1000)
	ProbPct   string  `json:"ProbPct"`   // percentage string from API
	InRunning bool    `json:"InRunning"`

	// On-chain anchor
	Ts        int64  `json:"Ts"`
	BlockHash string `json:"BlockHash"`
}

// ImpliedProbability converts decimal odds to implied probability (0.0 to 1.0).
func (o OddsEntry) ImpliedProbability() float64 {
	if o.Price <= 0 {
		return 0
	}
	return 1.0 / o.Price
}

// FlattenOddsRecords converts a slice of OddsRecord into a flat slice of OddsEntry,
// one entry per outcome per market. Invalid/zero prices are skipped.
func FlattenOddsRecords(records []OddsRecord) []OddsEntry {
	var entries []OddsEntry

	for _, rec := range records {
		if len(rec.PriceNames) == 0 || len(rec.Prices) == 0 {
			continue
		}

		marketName := rec.SuperOddsType
		if rec.MarketParameters != nil && *rec.MarketParameters != "" {
			marketName += " (" + *rec.MarketParameters + ")"
		}

		period := "full"
		if rec.MarketPeriod != nil && *rec.MarketPeriod != "" {
			period = *rec.MarketPeriod
		}

		for i, priceName := range rec.PriceNames {
			if i >= len(rec.Prices) {
				break
			}
			rawPrice := rec.Prices[i]
			if rawPrice <= 0 {
				continue
			}

			// Prices are scaled by 1000 in the TxLINE API.
			decimalPrice := float64(rawPrice) / 1000.0

			pct := ""
			if i < len(rec.Pct) {
				pct = rec.Pct[i]
			}

			// Build a stable unique key for this outcome across polls.
			// We use SuperOddsType + MarketParameters + MarketPeriod + PriceName
			// so the same market/outcome pair maps to the same key each poll.
			key := rec.SuperOddsType + "|" + period + "|"
			if rec.MarketParameters != nil {
				key += *rec.MarketParameters
			}
			key += "|" + priceName

			entries = append(entries, OddsEntry{
				FixtureID:    rec.FixtureID,
				OutcomeKey:   key,
				MarketName:   marketName,
				OutcomeName:  priceName,
				MarketPeriod: period,
				Price:        decimalPrice,
				ProbPct:      pct,
				InRunning:    rec.InRunning,
				Ts:           rec.Ts,
				BlockHash:    rec.BlockHash,
			})
		}
	}

	return entries
}

// OddsSnapshot is a map of OutcomeKey -> OddsEntry for a single fixture.
type OddsSnapshot map[string]OddsEntry

// ScoreEntry represents a single score record from /api/scores/snapshot/:fixtureId
type ScoreEntry struct {
	FixtureID int64          `json:"FixtureId"`
	Action    string         `json:"Action"`
	Period    int            `json:"Period"`
	StatusID  int            `json:"StatusId"`
	Seq       int64          `json:"Seq"`
	Timestamp int64          `json:"Ts"`
	GameState string         `json:"GameState"`
	Stats     map[string]int `json:"Stats"`
}

// IsFinished returns true when the match has reached a terminal state.
func (s ScoreEntry) IsFinished() bool {
	return s.StatusID == 100
}

// SSEMessage is the parsed form of a single Server-Sent Event frame.
type SSEMessage struct {
	ID    string
	Event string
	Data  string
}
