package main

import (
	"encoding/json"
	"fmt"
	"github-activity/internal/github"
	"github-activity/internal/storage"
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

	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	client := &github.Client{
		Timeout: 30 * time.Second,
	}

	maxRetries := 3
	baseDelay := 1 * time.Second
	maxDelay := 10 * time.Second

	body, err := github.FetchEvents(client, url, username, maxRetries, baseDelay, maxDelay)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
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

	fmt.Printf("Successfully saved events for user '%s' to %s\n", username, filename)
}
