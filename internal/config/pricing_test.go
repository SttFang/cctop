package config

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestLookupPricing(t *testing.T) {
	tests := []struct {
		modelID string
		want    string // expected pricing pattern match
	}{
		{"claude-opus-4-6", "opus-4-6"},
		{"claude-opus-4-5-20251101", "opus-4-5"},
		{"claude-sonnet-4-6", "sonnet-4-6"},
		{"claude-sonnet-4-5-20250929", "sonnet-4-5"},
		{"claude-haiku-4-5-20251001", "haiku-4-5"},
		{"claude-3-5-haiku-20241022", "haiku-3-5"},
		{"anthropic/claude-opus-4.6", "opus-4-6"},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			p := LookupPricing(tt.modelID)
			expected := PricingTable[tt.want]
			if p.InputPrice != expected.InputPrice {
				t.Errorf("LookupPricing(%q) InputPrice = %f, want %f", tt.modelID, p.InputPrice, expected.InputPrice)
			}
		})
	}
}

func TestLookupPricing_UnknownModel(t *testing.T) {
	p := LookupPricing("unknown-model-xyz")
	// Should fall back to sonnet pricing
	if p.InputPrice != 3.0 {
		t.Errorf("unknown model should use sonnet pricing, got InputPrice=%f", p.InputPrice)
	}
}

func TestComputeCost_Opus(t *testing.T) {
	// opus-4-6: input=15, output=75, cacheRead=1.5, cacheCreate=18.75 per M
	cost := ComputeCost("claude-opus-4-6", 1_000_000, 100_000, 5_000_000, 500_000)
	// = (1M*15 + 100K*75 + 5M*1.5 + 500K*18.75) / 1M
	// = (15,000,000 + 7,500,000 + 7,500,000 + 9,375,000) / 1M
	// = 39,375,000 / 1M = 39.375
	expected := 39.375
	if !almostEqual(cost, expected, 0.01) {
		t.Errorf("ComputeCost opus = %f, want %f", cost, expected)
	}
}

func TestComputeCost_Sonnet(t *testing.T) {
	// sonnet-4-6: input=3, output=15, cacheRead=0.30, cacheCreate=3.75 per M
	cost := ComputeCost("claude-sonnet-4-6", 1_000_000, 100_000, 5_000_000, 500_000)
	// = (1M*3 + 100K*15 + 5M*0.30 + 500K*3.75) / 1M
	// = (3,000,000 + 1,500,000 + 1,500,000 + 1,875,000) / 1M
	// = 7,875,000 / 1M = 7.875
	expected := 7.875
	if !almostEqual(cost, expected, 0.01) {
		t.Errorf("ComputeCost sonnet = %f, want %f", cost, expected)
	}
}

func TestComputeCost_ZeroTokens(t *testing.T) {
	cost := ComputeCost("claude-opus-4-6", 0, 0, 0, 0)
	if cost != 0 {
		t.Errorf("ComputeCost with zero tokens = %f, want 0", cost)
	}
}

func TestComputeCost_RealData(t *testing.T) {
	// From real stats-cache.json opus-4-6 usage
	cost := ComputeCost("claude-opus-4-6",
		188964877,   // input
		12971923,    // output
		4967705724,  // cacheRead
		338172247,   // cacheCreate
	)
	// Should be a large number
	if cost < 1000 {
		t.Errorf("real data cost should be > $1000, got %f", cost)
	}
	// Manual: (188M*15 + 13M*75 + 4968M*1.5 + 338M*18.75)/1M
	// ≈ 2834 + 973 + 7452 + 6341 = 17,600
	if !almostEqual(cost, 17600, 500) {
		t.Logf("real opus cost = $%.2f (expected ~$17,600 ± $500)", cost)
	}
}
