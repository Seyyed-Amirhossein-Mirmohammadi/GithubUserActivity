package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	Timeout time.Duration
}

func (c *Client) FetchEventsWithETag(url, username, etag string, maxRetries int, baseDelay, maxDelay time.Duration) ([]byte, string, int, error) {
	httpClient := &http.Client{
		Timeout: c.Timeout,
	}

	var lastErr error
	var resp *http.Response
	var body []byte

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, "", 0, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("User-Agent", "GitHubEventsCLI/1.0")
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}

		resp, err = httpClient.Do(req)
		if err != nil {
			lastErr = err
			fmt.Printf("Attempt %d failed (network/timeout): %v\n", attempt, err)
			if attempt == maxRetries {
				break
			}
			delay := backoffDelay(attempt, baseDelay, maxDelay)
			fmt.Printf("Retrying in %v...\n", delay)
			time.Sleep(delay)
			continue
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			fmt.Printf("Attempt %d failed (reading body): %v\n", attempt, err)
			if attempt == maxRetries {
				break
			}
			delay := backoffDelay(attempt, baseDelay, maxDelay)
			fmt.Printf("Retrying in %v...\n", delay)
			time.Sleep(delay)
			continue
		}

		status := resp.StatusCode

		if status == http.StatusNotModified {
			return nil, etag, status, nil
		}

		if status == http.StatusOK {
			newEtag := resp.Header.Get("ETag")
			return body, newEtag, status, nil
		}

		if status == http.StatusNotFound {
			return nil, "", status, fmt.Errorf("user '%s' does not exist (404 Not Found)", username)
		}

		if status == http.StatusForbidden {
			remaining := resp.Header.Get("X-RateLimit-Remaining")
			reset := resp.Header.Get("X-RateLimit-Reset")
			if remaining == "0" && reset != "" {
				resetTime, err := strconv.ParseInt(reset, 10, 64)
				if err == nil {
					wait := time.Until(time.Unix(resetTime, 0))
					if wait > 0 {
						fmt.Printf("Rate limit exceeded. Waiting %v before retry.\n", wait.Truncate(time.Second))
						time.Sleep(wait + 2*time.Second)
						continue
					}
				}
			}
			return nil, "", status, fmt.Errorf("access forbidden (403) – check rate limits or authentication\nResponse: %s", string(body))
		}

		if status >= 500 && status <= 599 {
			fmt.Printf("Attempt %d failed with server error %d\n", attempt, status)
			if attempt == maxRetries {
				lastErr = fmt.Errorf("server returned %d", status)
				break
			}
			delay := backoffDelay(attempt, baseDelay, maxDelay)
			fmt.Printf("Retrying in %v...\n", delay)
			time.Sleep(delay)
			continue
		}

		return nil, "", status, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	if lastErr != nil {
		return nil, "", 0, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
	}
	if resp == nil {
		return nil, "", 0, fmt.Errorf("unknown failure – no response")
	}
	return nil, "", resp.StatusCode, fmt.Errorf("failed to get events (status %d)", resp.StatusCode)
}

func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	delay := base * time.Duration(1<<uint(attempt-1))
	if delay > max {
		delay = max
	}
	jitter := time.Duration(100*attempt) * time.Millisecond
	return delay + jitter
}
