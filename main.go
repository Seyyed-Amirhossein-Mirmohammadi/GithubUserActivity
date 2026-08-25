package main

import (
	"encoding/json"
	"fmt"
	"github-activity/internal/cache"
	"github-activity/internal/github"
	"github-activity/internal/storage"
	"net/http"
	"os"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: go run main.go <github-username>")
		os.Exit(1)
	}
	username := args[0]

	cacher, err := cache.NewCache("./cache")
	if err != nil {
		fmt.Printf("Cache init failed: %v\n", err)
		os.Exit(1)
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
		newBody, newETag, status, err := github.FetchEventsWithETag(
			client, url, username, cachedETag, maxRetries, baseDelay, maxDelay,
		)
		if err != nil {
			fmt.Printf("Conditional request failed: %v\n", err)
			os.Exit(1)
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
			fmt.Printf("Fetched fresh events (new ETag %s)\n", newETag)
		} else {
			fmt.Printf("Unexpected status %d\n", status)
			os.Exit(1)
		}
	} else {
		body, newETag, status, err := github.FetchEventsWithETag(
			client, url, username, "", maxRetries, baseDelay, maxDelay,
		)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		if status != http.StatusOK {
			fmt.Printf("Unexpected status %d\n", status)
			os.Exit(1)
		}
		fromCache = false
		_ = cacher.Set(username, body, newETag)
		fmt.Printf("Fetched fresh events (ETag %s)\n", newETag)
	}

	var events interface{}
	if err := json.Unmarshal(body, &events); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	filename, err := storage.SaveEvents(events, username)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	if fromCache {
		fmt.Printf("Successfully saved cached events for user '%s' to %s\n", username, filename)
	} else {
		fmt.Printf("Successfully saved fresh events for user '%s' to %s\n", username, filename)
	}
}
