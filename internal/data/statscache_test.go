package data

import (
	"os"
	"testing"
)

func TestParseStatsCache(t *testing.T) {
	data, err := os.ReadFile("../../testdata/stats-cache.json")
	if err != nil {
		t.Fatal(err)
	}

	sc, err := ParseStatsCache(data)
	if err != nil {
		t.Fatalf("ParseStatsCache error: %v", err)
	}

	// Verify dailyActivity
	if len(sc.DailyActivity) != 2 {
		t.Errorf("DailyActivity length = %d, want 2", len(sc.DailyActivity))
	}
	if sc.DailyActivity[0].Date != "2026-03-05" {
		t.Errorf("first DailyActivity date = %q, want 2026-03-05", sc.DailyActivity[0].Date)
	}
	if sc.DailyActivity[0].MessageCount != 19390 {
		t.Errorf("first DailyActivity messageCount = %d, want 19390", sc.DailyActivity[0].MessageCount)
	}

	// Verify modelUsage
	if len(sc.ModelUsage) != 3 {
		t.Errorf("ModelUsage length = %d, want 3", len(sc.ModelUsage))
	}
	opus := sc.ModelUsage["claude-opus-4-6"]
	if opus.InputTokens != 188964877 {
		t.Errorf("opus InputTokens = %d, want 188964877", opus.InputTokens)
	}
	if opus.CacheReadInputTokens != 4967705724 {
		t.Errorf("opus CacheReadInputTokens = %d, want 4967705724", opus.CacheReadInputTokens)
	}

	// Verify hourCounts
	if len(sc.HourCounts) != 5 {
		t.Errorf("HourCounts length = %d, want 5", len(sc.HourCounts))
	}
	if sc.HourCounts["14"] != 3993 {
		t.Errorf("HourCounts[14] = %d, want 3993", sc.HourCounts["14"])
	}

	// Verify totals
	if sc.TotalSessions != 54995 {
		t.Errorf("TotalSessions = %d, want 54995", sc.TotalSessions)
	}
	if sc.TotalMessages != 376399 {
		t.Errorf("TotalMessages = %d, want 376399", sc.TotalMessages)
	}
}

func TestParseStatsCache_FileNotFound(t *testing.T) {
	_, err := LoadStatsCache("/nonexistent/path/stats-cache.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseStatsCache_Malformed(t *testing.T) {
	_, err := ParseStatsCache([]byte(`{invalid json`))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseStatsCache_Empty(t *testing.T) {
	sc, err := ParseStatsCache([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.TotalSessions != 0 {
		t.Errorf("expected 0 sessions for empty data, got %d", sc.TotalSessions)
	}
}
