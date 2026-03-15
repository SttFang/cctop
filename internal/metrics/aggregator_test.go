package metrics

import (
	"os"
	"testing"

	"github.com/fanghanjun/cctop/internal/data"
)

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1500, "1.5K"},
		{12345, "12.3K"},
		{123456, "123K"},
		{1234567, "1.23M"},
		{12345678, "12.3M"},
		{123456789, "123M"},
		{5670000000, "5.67B"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatTokens(tt.input)
			if got != tt.want {
				t.Errorf("FormatTokens(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.0, "$0.00"},
		{0.03, "$0.03"},
		{0.50, "$0.50"},
		{4.20, "$4.20"},
		{42.50, "$42.50"},
		{1284.32, "$1,284.32"},
		{12345.67, "$12,345.67"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatCost(tt.input)
			if got != tt.want {
				t.Errorf("FormatCost(%f) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0s"},
		{-100, "0s"},
		{45000, "45s"},
		{59000, "59s"},
		{60000, "1m"},
		{720000, "12m"},
		{3600000, "1h"},
		{3720000, "1h2m"},
		{7380000, "2h3m"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatDuration(tt.input)
			if got != tt.want {
				t.Errorf("FormatDuration(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCacheHitRate(t *testing.T) {
	tests := []struct {
		cacheRead, input int64
		want             float64
	}{
		{80, 20, 80.0},
		{0, 100, 0.0},
		{0, 0, 0.0}, // avoid division by zero
		{100, 0, 100.0},
	}
	for _, tt := range tests {
		got := CacheHitRate(tt.cacheRead, tt.input)
		if got != tt.want {
			t.Errorf("CacheHitRate(%d, %d) = %f, want %f", tt.cacheRead, tt.input, got, tt.want)
		}
	}
}

func TestNormalizeModelName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"claude-opus-4-6", "opus-4-6"},
		{"claude-opus-4-5-20251101", "opus-4-5"},
		{"claude-sonnet-4-6", "sonnet-4-6"},
		{"claude-sonnet-4-5-20250929", "sonnet-4-5"},
		{"claude-haiku-4-5-20251001", "haiku-4-5"},
		{"claude-3-5-haiku-20241022", "haiku-3-5"},
		{"anthropic/claude-opus-4.6", "opus-4-6"},
		{"unknown-model", "unknown-model"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeModelName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeModelName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestComputeOverview(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/stats-cache.json")
	if err != nil {
		t.Fatal(err)
	}
	sc, err := data.ParseStatsCache(raw)
	if err != nil {
		t.Fatal(err)
	}

	o := ComputeOverview(sc)

	if o.TotalSessions != 54995 {
		t.Errorf("TotalSessions = %d, want 54995", o.TotalSessions)
	}
	if o.TotalMessages != 376399 {
		t.Errorf("TotalMessages = %d, want 376399", o.TotalMessages)
	}
	if o.TotalCost <= 0 {
		t.Errorf("TotalCost = %f, should be > 0", o.TotalCost)
	}
	if o.TotalTokens <= 0 {
		t.Errorf("TotalTokens = %d, should be > 0", o.TotalTokens)
	}
	if len(o.ModelDistrib) != 3 {
		t.Errorf("ModelDistrib length = %d, want 3", len(o.ModelDistrib))
	}
	// opus should be first (most tokens)
	if o.ModelDistrib[0].Model != "opus-4-6" {
		t.Errorf("first model = %q, want opus-4-6", o.ModelDistrib[0].Model)
	}
	if len(o.DailyCosts) != 2 {
		t.Errorf("DailyCosts length = %d, want 2", len(o.DailyCosts))
	}
	// Verify hour counts are populated
	if o.HourCounts[14] != 3993 {
		t.Errorf("HourCounts[14] = %d, want 3993", o.HourCounts[14])
	}
}
