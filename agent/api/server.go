package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Mykelsown/txline-sharp/feed"
	"github.com/Mykelsown/txline-sharp/store"
)

// AgentState is the shared state the API reads from.
// It is written by the main polling loop and read by HTTP handlers.
type AgentState struct {
	mu sync.RWMutex

	WalletAddress       string
	ServiceLevel        int
	ActivatedAt         string
	PollIntervalSec     int
	MovementThreshold   float64
	IsRunning           bool
	AIInterpreterEnabled bool
	TotalSignals        int
	LastPoll            time.Time
	Fixtures            []feed.Fixture
}

// Update sets the last poll time and signal count atomically.
func (s *AgentState) Update(totalSignals int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastPoll = time.Now()
	s.TotalSignals = totalSignals
}

// SetFixtures stores the fetched fixture list.
func (s *AgentState) SetFixtures(fixtures []feed.Fixture) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Fixtures = fixtures
}

// Server is the HTTP API server.
type Server struct {
	state        *AgentState
	signalsFile  string
	arenaFile    string
	mux          *http.ServeMux
}

// NewServer constructs the API server with the given shared state.
func NewServer(state *AgentState, signalsFile, arenaFile string) *Server {
	s := &Server{
		state:       state,
		signalsFile: signalsFile,
		arenaFile:   arenaFile,
		mux:         http.NewServeMux(),
	}
	s.routes()
	return s
}

// routes registers all API endpoints.
func (s *Server) routes() {
	s.mux.HandleFunc("/api/status",   s.handleStatus)
	s.mux.HandleFunc("/api/fixtures", s.handleFixtures)
	s.mux.HandleFunc("/api/signals",  s.handleSignals)
	s.mux.HandleFunc("/api/arena",    s.handleArena)
	s.mux.HandleFunc("/health",       s.handleHealth)
}

// Start begins listening on the given address (e.g. ":8080").
func (s *Server) Start(addr string) error {
	log.Printf("[API] Listening on http://localhost%s", addr)
	return http.ListenAndServe(addr, cors(s.mux))
}

// handleStatus returns the agent's current runtime state.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.state.mu.RLock()
	resp := map[string]interface{}{
		"wallet":                 s.state.WalletAddress,
		"service_level":          s.state.ServiceLevel,
		"activated_at":           s.state.ActivatedAt,
		"poll_interval_sec":      s.state.PollIntervalSec,
		"movement_threshold":     s.state.MovementThreshold,
		"is_running":             s.state.IsRunning,
		"ai_interpreter_enabled": s.state.AIInterpreterEnabled,
		"total_signals":          s.state.TotalSignals,
		"last_poll":              s.state.LastPoll.Format(time.RFC3339),
	}
	s.state.mu.RUnlock()

	writeJSON(w, resp)
}

// handleFixtures returns the tracked World Cup fixtures.
func (s *Server) handleFixtures(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.state.mu.RLock()
	fixtures := s.state.Fixtures
	s.state.mu.RUnlock()

	type fixtureResp struct {
		ID        int64  `json:"id"`
		HomeTeam  string `json:"home_team"`
		AwayTeam  string `json:"away_team"`
		Kickoff   string `json:"kickoff"`
		Status    string `json:"status"`
		GameState int    `json:"game_state"`
	}

	var resp []fixtureResp
	now := time.Now()
	for _, f := range fixtures {
		status := "upcoming"
		if f.StartTime().Before(now) {
			status = "live"
		}
		resp = append(resp, fixtureResp{
			ID:        f.FixtureID,
			HomeTeam:  f.HomeTeam(),
			AwayTeam:  f.AwayTeam(),
			Kickoff:   f.StartTime().Format(time.RFC3339),
			Status:    status,
			GameState: f.GameState,
		})
	}

	if resp == nil {
		resp = []fixtureResp{}
	}

	writeJSON(w, resp)
}

// handleSignals reads signals.jsonl and returns all signals as JSON.
func (s *Server) handleSignals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	signals, err := store.LoadAll(s.signalsFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("load signals: %v", err), http.StatusInternalServerError)
		return
	}

	if signals == nil {
		writeJSON(w, []struct{}{})
		return
	}

	writeJSON(w, signals)
}

// handleArena reads arena_results.json and returns it as JSON.
func (s *Server) handleArena(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := os.ReadFile(s.arenaFile)
	if os.IsNotExist(err) {
		// Arena results not yet written, return empty structure.
		writeJSON(w, map[string]interface{}{
			"generated_at": time.Now().Format(time.RFC3339),
			"decisions":    []interface{}{},
			"summary":      map[string]interface{}{},
		})
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("read arena file: %v", err), http.StatusInternalServerError)
		return
	}

	// Pass through raw JSON without re-encoding.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// handleHealth is a simple liveness probe for Docker and deployment.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

// writeJSON encodes v as JSON and writes it to w.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[API] encode error: %v", err)
	}
}

// cors wraps a handler with permissive CORS headers for local development.
// In production, tighten the Allow-Origin to your deployed frontend domain.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
