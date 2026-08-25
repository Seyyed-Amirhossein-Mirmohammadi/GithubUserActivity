package service

import (
	"encoding/json"
	"fmt"
	"github-activity/internal/cache"
	"github-activity/internal/github"
	"github-activity/internal/storage"
	"net/http"
	"time"
)

func FetchAndSaveEvents(username string) error {
	cacher, err := cache.NewCache("./cache")
	if err != nil {
		return fmt.Errorf("cache init failed: %v", err)
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	client := &github.Client{Timeout: 30 * time.Second}
	maxRetries := 3
	baseDelay := 1 * time.Second
	maxDelay := 10 * time.Second

	var body []byte
	var fromCache bool

	cachedBody, cachedETag, ok := cacher.Get(username)

	if ok {
		newBody, newETag, status, err := client.FetchEventsWithETag(
			url, username, cachedETag, maxRetries, baseDelay, maxDelay,
		)
		if err != nil {
			return err
		}

		if status == http.StatusNotModified {
			body = cachedBody
			fromCache = true
			_ = cacher.Set(username, body, cachedETag)
			fmt.Printf("Using cached events (unchanged, ETag %s)\n", cachedETag)
		} else if status == http.StatusOK {
			body = newBody
			fromCache = false
			_ = cacher.Set(username, body, newETag)
			fmt.Printf("Fetched fresh events (ETag %s)\n", newETag)
		} else {
			return fmt.Errorf("unexpected status %d", status)
		}
	} else {
		newBody, newETag, status, err := client.FetchEventsWithETag(
			url, username, "", maxRetries, baseDelay, maxDelay,
		)
		if err != nil {
			return err
		}
		if status != http.StatusOK {
			return fmt.Errorf("unexpected status %d", status)
		}
		body = newBody
		fromCache = false
		_ = cacher.Set(username, body, newETag)
		fmt.Printf("Fetched fresh events (ETag %s)\n", newETag)
	}

	if len(body) == 0 {
		return fmt.Errorf("empty response body")
	}

	var events interface{}
	if err := json.Unmarshal(body, &events); err != nil {
		return fmt.Errorf("failed to parse JSON: %v\nResponse body (first 100 chars): %s", err, string(body[:min(len(body), 100)]))
	}

	filename, err := storage.SaveEvents(events, username)
	if err != nil {
		return fmt.Errorf("error saving events: %v", err)
	}

	if fromCache {
		fmt.Printf("Successfully saved cached events for user '%s' to %s\n", username, filename)
	} else {
		fmt.Printf("Successfully saved fresh events for user '%s' to %s\n", username, filename)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
