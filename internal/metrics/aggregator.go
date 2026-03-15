package metrics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/fanghanjun/cctop/internal/config"
	"github.com/fanghanjun/cctop/internal/data"
)

// Overview holds all computed metrics for the Overview tab.
type Overview struct {
	TotalCost     float64
	TodayCost     float64
	TotalTokens   int64
	TotalInput    int64
	TotalOutput   int64
	TotalCacheR   int64
	TotalCacheW   int64
	TotalSessions int
	TotalMessages int
	TodaySessions int
	TodayMessages int
	ModelDistrib  []ModelShare
	DailyCosts    []DailyCost
	HourCounts    [24]int
}

type ModelShare struct {
	Model   string
	Tokens  int64
	Percent float64
	Cost    float64
}

type DailyCost struct {
	Date string
	Cost float64
}

// modelCostPerToken computes the blended cost-per-token for a model
// based on the actual ratio of input/output/cacheRead/cacheWrite from modelUsage.
func modelCostPerToken(modelID string, usage data.ModelUsageEntry) float64 {
	total := usage.InputTokens + usage.OutputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens
	if total == 0 {
		return 0
	}
	cost := config.ComputeCost(modelID, usage.InputTokens, usage.OutputTokens, usage.CacheReadInputTokens, usage.CacheCreationInputTokens)
	return cost / float64(total)
}

// ComputeOverview derives all overview metrics from stats cache data.
func ComputeOverview(sc *data.StatsCache) Overview {
	o := Overview{
		TotalSessions: sc.TotalSessions,
		TotalMessages: sc.TotalMessages,
	}

	today := time.Now().Format("2006-01-02")

	// Merge models with same normalized name
	merged := make(map[string]*ModelShare)
	costPerToken := make(map[string]float64) // rawModelID -> blended $/token

	for modelID, usage := range sc.ModelUsage {
		cost := config.ComputeCost(modelID, usage.InputTokens, usage.OutputTokens, usage.CacheReadInputTokens, usage.CacheCreationInputTokens)
		tokens := usage.InputTokens + usage.OutputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens
		o.TotalCost += cost
		o.TotalTokens += tokens
		o.TotalInput += usage.InputTokens
		o.TotalOutput += usage.OutputTokens
		o.TotalCacheR += usage.CacheReadInputTokens
		o.TotalCacheW += usage.CacheCreationInputTokens

		// Store per-token cost for daily cost calculation
		costPerToken[modelID] = modelCostPerToken(modelID, usage)

		name := NormalizeModelName(modelID)
		if m, ok := merged[name]; ok {
			m.Tokens += tokens
			m.Cost += cost
		} else {
			merged[name] = &ModelShare{Model: name, Tokens: tokens, Cost: cost}
		}
	}

	// Build sorted distribution
	for _, m := range merged {
		if o.TotalTokens > 0 {
			m.Percent = float64(m.Tokens) / float64(o.TotalTokens) * 100
		}
		o.ModelDistrib = append(o.ModelDistrib, *m)
	}
	sort.Slice(o.ModelDistrib, func(i, j int) bool {
		return o.ModelDistrib[i].Tokens > o.ModelDistrib[j].Tokens
	})

	// Today's activity
	for _, da := range sc.DailyActivity {
		if da.Date == today {
			o.TodaySessions = da.SessionCount
			o.TodayMessages = da.MessageCount
			break
		}
	}

	// Daily costs: use blended cost-per-token from modelUsage ratios
	for _, dmt := range sc.DailyModelTokens {
		var dayCost float64
		for modelID, tokens := range dmt.TokensByModel {
			cpt := costPerToken[modelID]
			if cpt == 0 {
				// Fallback: model not in modelUsage, use input price as rough estimate
				p := config.LookupPricing(modelID)
				cpt = p.InputPrice / 1_000_000
			}
			dayCost += float64(tokens) * cpt
		}
		o.DailyCosts = append(o.DailyCosts, DailyCost{Date: dmt.Date, Cost: dayCost})
		if dmt.Date == today {
			o.TodayCost = dayCost
		}
	}

	// Hour counts
	for hourStr, count := range sc.HourCounts {
		var hour int
		fmt.Sscanf(hourStr, "%d", &hour)
		if hour >= 0 && hour < 24 {
			o.HourCounts[hour] = count
		}
	}

	return o
}

// NormalizeModelName simplifies model IDs for display.
func NormalizeModelName(modelID string) string {
	// Remove "anthropic/" prefix and normalize dots to dashes
	modelID = strings.TrimPrefix(modelID, "anthropic/")
	normalized := strings.ReplaceAll(modelID, ".", "-")

	// Map known patterns to short names
	switch {
	case strings.Contains(normalized, "opus-4-6"):
		return "opus-4-6"
	case strings.Contains(normalized, "opus-4-5"):
		return "opus-4-5"
	case strings.Contains(normalized, "sonnet-4-6"):
		return "sonnet-4-6"
	case strings.Contains(normalized, "sonnet-4-5"):
		return "sonnet-4-5"
	case strings.Contains(normalized, "haiku-4-5"):
		return "haiku-4-5"
	case strings.Contains(normalized, "3-5-haiku") || strings.Contains(normalized, "haiku-3-5"):
		return "haiku-3-5"
	default:
		return modelID
	}
}

// FormatTokens formats a token count for display.
func FormatTokens(n int64) string {
	switch {
	case n < 1000:
		return fmt.Sprintf("%d", n)
	case n < 1_000_000:
		v := float64(n) / 1000
		if v >= 100 {
			return fmt.Sprintf("%.0fK", v)
		}
		if v >= 10 {
			return fmt.Sprintf("%.1fK", v)
		}
		return fmt.Sprintf("%.1fK", v)
	case n < 1_000_000_000:
		v := float64(n) / 1_000_000
		if v >= 100 {
			return fmt.Sprintf("%.0fM", v)
		}
		if v >= 10 {
			return fmt.Sprintf("%.1fM", v)
		}
		return fmt.Sprintf("%.2fM", v)
	default:
		v := float64(n) / 1_000_000_000
		if v >= 100 {
			return fmt.Sprintf("%.0fB", v)
		}
		if v >= 10 {
			return fmt.Sprintf("%.1fB", v)
		}
		return fmt.Sprintf("%.2fB", v)
	}
}

// FormatCost formats a dollar amount for display.
func FormatCost(cost float64) string {
	if cost < 0.01 && cost > 0 {
		return "$0.01"
	}
	if cost < 1 {
		return fmt.Sprintf("$%.2f", cost)
	}
	// Add thousands separator
	whole := int64(cost)
	frac := cost - float64(whole)
	wholeStr := formatWithCommas(whole)
	return fmt.Sprintf("$%s.%02d", wholeStr, int(math.Round(frac*100)))
}

// FormatDuration formats milliseconds to a human-readable duration.
func FormatDuration(ms int64) string {
	if ms < 0 {
		return "0s"
	}
	secs := ms / 1000
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("%dm", mins)
	}
	hours := mins / 60
	remainMins := mins % 60
	if remainMins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remainMins)
}

func formatWithCommas(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// CacheHitRate computes the cache hit rate percentage.
func CacheHitRate(cacheRead, input int64) float64 {
	total := cacheRead + input
	if total == 0 {
		return 0
	}
	return float64(cacheRead) / float64(total) * 100
}
