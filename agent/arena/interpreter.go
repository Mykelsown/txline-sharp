package arena

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Mykelsown/txline-sharp/detector"
)

const anthropicURL = "https://api.anthropic.com/v1/messages"

// Interpreter uses Claude to generate human-readable commentary on signals.
type Interpreter struct {
	apiKey string
	client *http.Client
}

// NewInterpreter constructs an Interpreter. Returns nil if no API key is set,
// so the caller can skip interpretation gracefully.
func NewInterpreter() *Interpreter {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil
	}
	return &Interpreter{
		apiKey: key,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

// Interpret sends a signal to Claude and returns a short analytical commentary.
// If the API call fails, it returns a fallback string rather than an error,
// so the agent continues running even if interpretation is unavailable.
func (i *Interpreter) Interpret(sig detector.Signal) string {
	if i == nil {
		return ""
	}

	prompt := buildPrompt(sig)

	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4-6",
		"max_tokens": 300,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"system": `You are a sharp sports betting analyst. You receive real-time odds movement 
data from TxLINE and provide concise, professional commentary on what the movement 
likely means. Be analytical and specific. Maximum 3 sentences. No hedging phrases 
like "it's possible that". State your analysis directly.`,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Sprintf("[interpretation unavailable: %v]", err)
	}

	req, err := http.NewRequest(http.MethodPost, anthropicURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Sprintf("[interpretation unavailable: %v]", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", i.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Sprintf("[interpretation unavailable: %v]", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("[interpretation unavailable: %v]", err)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return "[interpretation unavailable: parse error]"
	}

	if result.Error != nil {
		return fmt.Sprintf("[interpretation unavailable: %s]", result.Error.Message)
	}

	for _, block := range result.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text
		}
	}

	return "[interpretation unavailable: empty response]"
}

// buildPrompt constructs the analysis prompt for Claude.
func buildPrompt(sig detector.Signal) string {
	return fmt.Sprintf(`Sharp odds movement detected on a World Cup match:

Match: %s vs %s
Market: %s
Outcome: %s
Direction: %s
Price movement: %.3f -> %.3f
Implied probability shift: %.1f%% -> %.1f%% (delta: +%.1f%%)
Severity: %s
Detected at: %s
On-chain block hash: %s

Analyze what this movement likely indicates about where professional money is going 
and what it suggests about the likely match outcome. Consider the severity and direction.`,
		sig.HomeTeam, sig.AwayTeam,
		sig.MarketName,
		sig.OutcomeName,
		sig.Direction,
		sig.PriceBefore, sig.PriceAfter,
		sig.ProbBefore*100, sig.ProbAfter*100, sig.ProbDelta*100,
		sig.Severity,
		sig.DetectedAt.Format(time.RFC3339),
		sig.BlockHash,
	)
}