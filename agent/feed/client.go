package feed

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	BaseURL      = "https://txline.txodds.com"
	APIBase      = BaseURL + "/api"
	GuestAuthURL = BaseURL + "/auth/guest/start"
)

// Client wraps the TxLINE HTTP API with automatic JWT renewal.
type Client struct {
	http     *http.Client
	jwt      string
	apiToken string
	mu       sync.Mutex
}

// NewClient constructs a Client from the activated credentials.
func NewClient(jwt, apiToken string) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		jwt:      jwt,
		apiToken: apiToken,
	}
}

// renewJWT fetches a fresh guest JWT from the auth endpoint.
func (c *Client) renewJWT() error {
	resp, err := c.http.Post(GuestAuthURL, "application/json", nil)
	if err != nil {
		return fmt.Errorf("jwt renewal request failed: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("jwt renewal decode failed: %w", err)
	}
	if body.Token == "" {
		return fmt.Errorf("jwt renewal returned empty token")
	}

	c.mu.Lock()
	c.jwt = body.Token
	c.mu.Unlock()
	return nil
}

// get performs a GET request, renewing the JWT once on 401.
func (c *Client) get(path string) ([]byte, error) {
	for attempt := 0; attempt < 2; attempt++ {
		c.mu.Lock()
		jwt := c.jwt
		c.mu.Unlock()

		req, err := http.NewRequest(http.MethodGet, APIBase+path, nil)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+jwt)
		req.Header.Set("X-Api-Token", c.apiToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http get %s: %w", path, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			if err := c.renewJWT(); err != nil {
				return nil, fmt.Errorf("jwt renewal: %w", err)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("http %d from %s: %s", resp.StatusCode, path, string(body))
		}

		return io.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("get %s: exhausted retries", path)
}

// Fixtures fetches all fixtures in the active subscription bundle.
// The free World Cup tier returns all covered World Cup fixtures.
func (c *Client) Fixtures() ([]Fixture, error) {
	data, err := c.get("/fixtures/snapshot")
	if err != nil {
		return nil, fmt.Errorf("fixtures: %w", err)
	}
	var fixtures []Fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("fixtures decode: %w", err)
	}
	return fixtures, nil
}

// OddsSnapshot fetches the current odds for a single fixture.
func (c *Client) OddsSnapshot(fixtureID int64) ([]OddsEntry, error) {
	data, err := c.get(fmt.Sprintf("/odds/snapshot/%d", fixtureID))
	if err != nil {
		return nil, fmt.Errorf("odds snapshot %d: %w", fixtureID, err)
	}
	var entries []OddsEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("odds snapshot decode: %w", err)
	}
	return entries, nil
}

// ScoresSnapshot fetches the latest score state for a single fixture.
func (c *Client) ScoresSnapshot(fixtureID int64) ([]ScoreEntry, error) {
	data, err := c.get(fmt.Sprintf("/scores/snapshot/%d", fixtureID))
	if err != nil {
		return nil, fmt.Errorf("scores snapshot %d: %w", fixtureID, err)
	}
	var entries []ScoreEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("scores snapshot decode: %w", err)
	}
	return entries, nil
}

// JWT returns the current guest JWT (safe for SSE stream headers).
func (c *Client) JWT() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.jwt
}

// APIToken returns the activated API token.
func (c *Client) APIToken() string {
	return c.apiToken
}