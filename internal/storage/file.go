package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

func SaveEvents(events interface{}, username string) (string, error) {
	prettyJSON, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return "", fmt.Errorf("Error formatting JSON: %v", err)
	}
	filename := fmt.Sprintf("%s_events.json", username)
	if err := os.WriteFile(filename, prettyJSON, 0644); err != nil {
		return "", fmt.Errorf("Error writing file: %v", err)
	}
	return filename, nil
}
