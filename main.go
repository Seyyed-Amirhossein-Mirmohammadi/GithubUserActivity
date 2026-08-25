package main

import (
	"fmt"
	"github-activity/internal/service"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: go run main.go <github-username>")
		os.Exit(1)
	}
	username := args[0]

	if err := service.FetchAndSaveEvents(username); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
