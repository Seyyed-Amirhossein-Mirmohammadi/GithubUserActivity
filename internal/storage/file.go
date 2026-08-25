package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func SaveEvents(events interface{}, username string) (string, error) {
	outputDir := "./output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	prettyJSON, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %v", err)
	}

	filename := filepath.Join(outputDir, fmt.Sprintf("%s_events.json", username))
	if err := os.WriteFile(filename, prettyJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	return filename, nil
}
