package feed

import "time"

// Fixture represents a single match from /api/fixtures/snapshot
type Fixture struct {
	FixtureID          int64   `json:"FixtureId"`
	CompetitionID      int64   `json:"CompetitionId"`
	Participant1       string  `json:"Participant1"`
	Participant2       string  `json:"Participant2"`
	Participant1IsHome bool    `json:"Participant1IsHome"`
	StartTimeMs        int64   `json:"StartTime"` // Unix timestamp in milliseconds
	GameState          int     `json:"GameState"`
	StatusID           int     `json:"StatusId"`
}

// StartTime returns the fixture start time as a time.Time.
func (f Fixture) StartTime() time.Time {
	return time.Unix(f.StartTimeMs/1000, 0).UTC()
}

// HomeTeam returns the home team name based on the feed designation.
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

// OddsEntry represents a single odds record from /api/odds/snapshot/:fixtureId
type OddsEntry struct {
	FixtureID   int64   `json:"FixtureId"`
	MarketID    int64   `json:"MarketId"`
	OutcomeID   int64   `json:"OutcomeId"`
	MarketName  string  `json:"MarketName"`
	OutcomeName string  `json:"OutcomeName"`
	Price       float64 `json:"Price"`
	Timestamp   string  `json:"Timestamp"`
	BlockHash   string  `json:"BlockHash"`
}

// ImpliedProbability converts a decimal odds price to implied probability (0.0 to 1.0).
func (o OddsEntry) ImpliedProbability() float64 {
	if o.Price <= 0 {
		return 0
	}
	return 1.0 / o.Price
}

// ScoreEntry represents a single score record from /api/scores/snapshot/:fixtureId
type ScoreEntry struct {
	FixtureID int64          `json:"FixtureId"`
	Action    string         `json:"Action"`
	Period    int            `json:"Period"`
	StatusID  int            `json:"StatusId"`
	Seq       int64          `json:"Seq"`
	Timestamp int64         `json:"Ts"`
	GameState string         `json:"GameState"` // API returns string e.g. "1" not int
	Stats     map[string]int `json:"Stats"`
}

// IsFinished returns true when the match has reached a terminal state.
// StatusID 100 = game_finalised per TxLINE soccer feed docs.
func (s ScoreEntry) IsFinished() bool {
	return s.StatusID == 100
}

// OddsSnapshot is a map of outcomeID -> OddsEntry for a single fixture,
// used by the detector to diff against the previous snapshot.
type OddsSnapshot map[int64]OddsEntry

// SSEMessage is the parsed form of a single Server-Sent Event frame.
type SSEMessage struct {
	ID    string
	Event string
	Data  string
}