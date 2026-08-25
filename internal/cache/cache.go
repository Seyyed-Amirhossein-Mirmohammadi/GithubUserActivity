package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type meta struct {
	Timestamp int64  `json:"timestamp"`
	ETag      string `json:"etag"`
}

type Cache struct {
	Dir string
}

func NewCache(dir string) (*Cache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &Cache{Dir: dir}, nil
}

func (c *Cache) Get(username string) ([]byte, string, bool) {
	base := filepath.Join(c.Dir, username)

	metaFile := base + ".meta"
	data, err := os.ReadFile(metaFile)
	if err != nil {
		return nil, "", false
	}
	var m meta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, "", false
	}

	body, err := os.ReadFile(base + ".json")
	if err != nil {
		return nil, "", false
	}
	return body, m.ETag, true
}

func (c *Cache) Set(username string, body []byte, etag string) error {
	base := filepath.Join(c.Dir, username)

	if err := os.WriteFile(base+".json", body, 0644); err != nil {
		return fmt.Errorf("failed to write cache body: %w", err)
	}

	m := meta{
		Timestamp: time.Now().Unix(),
		ETag:      etag,
	}
	metaData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}
	if err := os.WriteFile(base+".meta", metaData, 0644); err != nil {
		return fmt.Errorf("failed to write cache meta: %w", err)
	}
	return nil
}
