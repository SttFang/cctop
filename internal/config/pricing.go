package config

import "strings"

// ModelPricing holds per-million-token prices in USD.
type ModelPricing struct {
	InputPrice       float64 // $/M input tokens
	OutputPrice      float64 // $/M output tokens
	CacheReadPrice   float64 // $/M cache read tokens
	CacheCreatePrice float64 // $/M cache creation tokens
}

// PricingTable maps model ID patterns to pricing.
var PricingTable = map[string]ModelPricing{
	"opus-4-6": {
		InputPrice:       15.0,
		OutputPrice:      75.0,
		CacheReadPrice:   1.5,
		CacheCreatePrice: 18.75,
	},
	"opus-4-5": {
		InputPrice:       15.0,
		OutputPrice:      75.0,
		CacheReadPrice:   1.5,
		CacheCreatePrice: 18.75,
	},
	"sonnet-4-6": {
		InputPrice:       3.0,
		OutputPrice:      15.0,
		CacheReadPrice:   0.30,
		CacheCreatePrice: 3.75,
	},
	"sonnet-4-5": {
		InputPrice:       3.0,
		OutputPrice:      15.0,
		CacheReadPrice:   0.30,
		CacheCreatePrice: 3.75,
	},
	"haiku-4-5": {
		InputPrice:       0.80,
		OutputPrice:      4.0,
		CacheReadPrice:   0.08,
		CacheCreatePrice: 1.0,
	},
	"haiku-3-5": {
		InputPrice:       0.80,
		OutputPrice:      4.0,
		CacheReadPrice:   0.08,
		CacheCreatePrice: 1.0,
	},
}

// defaultPricing is used when no model match is found (sonnet pricing).
var defaultPricing = PricingTable["sonnet-4-6"]

// normalizeForLookup normalizes model ID for pricing lookup.
func normalizeForLookup(modelID string) string {
	modelID = strings.TrimPrefix(modelID, "anthropic/")
	modelID = strings.ReplaceAll(modelID, ".", "-")
	// Handle "claude-3-5-haiku" → contains "3-5-haiku" → match "haiku-3-5" won't work
	// So we normalize "3-5-haiku" patterns too
	return modelID
}

// LookupPricing finds the pricing for a model ID by matching known patterns.
func LookupPricing(modelID string) ModelPricing {
	normalized := normalizeForLookup(modelID)
	for pattern, pricing := range PricingTable {
		if strings.Contains(normalized, pattern) {
			return pricing
		}
	}
	// Check reversed haiku pattern: "3-5-haiku" in "claude-3-5-haiku-..."
	if strings.Contains(normalized, "3-5-haiku") {
		return PricingTable["haiku-3-5"]
	}
	return defaultPricing
}

// ComputeCost calculates the cost in USD for given token counts and model.
func ComputeCost(modelID string, input, output, cacheRead, cacheCreate int64) float64 {
	p := LookupPricing(modelID)
	cost := (float64(input)*p.InputPrice +
		float64(output)*p.OutputPrice +
		float64(cacheRead)*p.CacheReadPrice +
		float64(cacheCreate)*p.CacheCreatePrice) / 1_000_000
	return cost
}
