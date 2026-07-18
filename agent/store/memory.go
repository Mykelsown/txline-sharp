package store

import (
	"sync"

	"github.com/Mykelsown/txline-sharp/feed"
)

// OddsSnapshot is a map of OutcomeKey -> OddsEntry for one fixture.
type OddsSnapshot = feed.OddsSnapshot

// Memory holds the last known odds snapshot for every tracked fixture.
type Memory struct {
	mu        sync.RWMutex
	snapshots map[int64]feed.OddsSnapshot
}

// NewMemory constructs an empty Memory store.
func NewMemory() *Memory {
	return &Memory{
		snapshots: make(map[int64]feed.OddsSnapshot),
	}
}

// Get returns the previous odds snapshot for a fixture, and whether one exists.
func (m *Memory) Get(fixtureID int64) (feed.OddsSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	snap, ok := m.snapshots[fixtureID]
	return snap, ok
}

// Set stores the latest odds snapshot for a fixture.
func (m *Memory) Set(fixtureID int64, snap feed.OddsSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[fixtureID] = snap
}

// BuildSnapshot converts a flat slice of OddsEntry into an OddsSnapshot
// map keyed by OutcomeKey for O(1) diff lookups.
func BuildSnapshot(entries []feed.OddsEntry) feed.OddsSnapshot {
	snap := make(feed.OddsSnapshot, len(entries))
	for _, e := range entries {
		snap[e.OutcomeKey] = e
	}
	return snap
}
