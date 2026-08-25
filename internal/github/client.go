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

func FetchEvents(client *Client, url, username string, maxRetries int, baseDelay, maxDelay time.Duration) ([]byte, error) {
	httpClient := &http.Client{
		Timeout: client.Timeout,
	}

	var lastErr error
	var resp *http.Response
	var body []byte

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("Error creating request: %v", err)
		}
		req.Header.Set("User-Agent", "GitHubEventsCLI/1.0")

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
		if status == http.StatusOK {
			return body, nil
		}

		if status == http.StatusNotFound {
			return nil, fmt.Errorf("User '%s' does not exist (404 Not Found).", username)
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
			return nil, fmt.Errorf("Access forbidden (403). Check rate limits or authentication.\nResponse body: %s", string(body))
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

		return nil, fmt.Errorf("Unexpected status %d: %s", status, string(body))
	}

	if lastErr != nil {
		return nil, fmt.Errorf("Failed after %d attempts: %v", maxRetries, lastErr)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get events (status %d).", resp.StatusCode)
	}
	return nil, fmt.Errorf("Unknown failure")
}

func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	delay := base * time.Duration(1<<uint(attempt-1))
	if delay > max {
		delay = max
	}
	jitter := time.Duration(100*attempt) * time.Millisecond
	return delay + jitter
}
