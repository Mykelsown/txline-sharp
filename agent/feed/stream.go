package feed

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// StreamOdds opens the TxLINE odds SSE stream and sends parsed messages
// to the returned channel. It reconnects automatically on error.
// Cancel the context to stop the stream.
func (c *Client) StreamOdds(ctx context.Context) (<-chan SSEMessage, <-chan error) {
	return c.stream(ctx, "/odds/stream")
}

// StreamScores opens the TxLINE scores SSE stream.
func (c *Client) StreamScores(ctx context.Context) (<-chan SSEMessage, <-chan error) {
	return c.stream(ctx, "/scores/stream")
}

// stream is the shared SSE connection loop used by StreamOdds and StreamScores.
func (c *Client) stream(ctx context.Context, path string) (<-chan SSEMessage, <-chan error) {
	msgs := make(chan SSEMessage, 64)
	errs := make(chan error, 4)

	go func() {
		defer close(msgs)
		defer close(errs)

		backoff := 2 * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := c.connectSSE(ctx, path, msgs); err != nil {
				select {
				case errs <- fmt.Errorf("stream %s: %w", path, err):
				default:
				}

				// Renew JWT before reconnecting in case of auth expiry.
				_ = c.renewJWT()

				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
					if backoff < 30*time.Second {
						backoff *= 2
					}
				}
				continue
			}
			backoff = 2 * time.Second
		}
	}()

	return msgs, errs
}

// connectSSE opens a single SSE connection and reads messages until the
// connection closes or the context is cancelled.
func (c *Client) connectSSE(ctx context.Context, path string, out chan<- SSEMessage) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, APIBase+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	c.mu.Lock()
	jwt := c.jwt
	c.mu.Unlock()

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("X-Api-Token", c.apiToken)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	httpClient := &http.Client{Timeout: 0} // no timeout for streaming
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("401 unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	return c.parseSSE(ctx, resp, out)
}

// parseSSE reads lines from an open SSE response body and emits SSEMessages.
func (c *Client) parseSSE(ctx context.Context, resp *http.Response, out chan<- SSEMessage) error {
	scanner := bufio.NewScanner(resp.Body)

	var current SSEMessage

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line := scanner.Text()

		// Blank line = message boundary, dispatch if non-empty.
		if line == "" {
			if current.Data != "" || current.Event != "" {
				select {
				case out <- current:
				case <-ctx.Done():
					return nil
				}
			}
			current = SSEMessage{}
			continue
		}

		// Comment / heartbeat lines start with ":".
		if strings.HasPrefix(line, ":") {
			continue
		}

		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}

		field := line[:idx]
		value := strings.TrimPrefix(line[idx+1:], " ")

		switch field {
		case "id":
			current.ID = value
		case "event":
			current.Event = value
		case "data":
			if current.Data != "" {
				current.Data += "\n"
			}
			current.Data += value
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner: %w", err)
	}
	return nil
}