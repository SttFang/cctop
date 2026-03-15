package data

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadStatsCache reads and parses the stats-cache.json file.
func LoadStatsCache(path string) (*StatsCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading stats cache: %w", err)
	}
	return ParseStatsCache(data)
}

// ParseStatsCache parses raw JSON bytes into a StatsCache struct.
func ParseStatsCache(data []byte) (*StatsCache, error) {
	var sc StatsCache
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("parsing stats cache: %w", err)
	}
	return &sc, nil
}
