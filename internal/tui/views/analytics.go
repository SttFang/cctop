package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/config"
	"github.com/fanghanjun/cctop/internal/data"
	"github.com/fanghanjun/cctop/internal/metrics"
	"github.com/fanghanjun/cctop/internal/tui/common"
	"github.com/fanghanjun/cctop/internal/tui/components"
)

type AnalyticsView struct {
	Layout common.Layout
	Theme  common.Theme
	Stats  *data.StatsCache
	Period string
}

func (v *AnalyticsView) View() string {
	if v.Stats == nil {
		return lipgloss.NewStyle().
			Foreground(v.Theme.FgDim).
			Width(v.Layout.ContentW).
			Height(v.Layout.ContentH).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Loading analytics data...")
	}

	w := v.Layout.ContentW
	h := v.Layout.ContentH

	// Row 0: Period selector (1 line)
	periodBar := v.renderPeriodBar(w)

	// Available height after period bar
	contentH := h - 2

	// Row 1: Token trend (top 40%)
	topH := contentH * 40 / 100
	if topH < 6 {
		topH = 6
	}
	tokenTrend := v.renderTokenTrend(w, topH)

	// Row 2: Cost trend + Cache rate (middle 30%)
	midH := contentH * 30 / 100
	if midH < 5 {
		midH = 5
	}
	midRow := v.renderMidRow(w, midH)

	// Row 3: Hourly + Model breakdown (bottom)
	botH := contentH - topH - midH
	if botH < 5 {
		botH = 5
	}
	botRow := v.renderBotRow(w, botH)

	return components.VStack([]string{periodBar, tokenTrend, midRow, botRow})
}

func (v *AnalyticsView) renderPeriodBar(w int) string {
	dim := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	active := lipgloss.NewStyle().Background(v.Theme.Primary).Foreground(v.Theme.Bg).Padding(0, 1)
	inactive := lipgloss.NewStyle().Foreground(v.Theme.FgDim).Padding(0, 1)

	periods := []string{"D", "W", "M", "A"}
	var parts []string
	parts = append(parts, dim.Render("  Period: "))
	for _, p := range periods {
		if v.Period == p {
			parts = append(parts, active.Render(p))
		} else {
			parts = append(parts, inactive.Render(p))
		}
	}

	line := strings.Join(parts, "")
	return lipgloss.NewStyle().Width(w).MaxWidth(w).Render(line) + "\n" +
		dim.Render(strings.Repeat("─", w))
}

func (v *AnalyticsView) filterTokens() []data.DailyModelTokens {
	now := time.Now()
	var start string
	switch v.Period {
	case "D":
		start = now.Format("2006-01-02")
	case "W":
		start = now.AddDate(0, 0, -7).Format("2006-01-02")
	case "M":
		start = now.AddDate(0, -1, 0).Format("2006-01-02")
	default:
		start = ""
	}

	var result []data.DailyModelTokens
	for _, dmt := range v.Stats.DailyModelTokens {
		if start == "" || dmt.Date >= start {
			result = append(result, dmt)
		}
	}
	return result
}

func (v *AnalyticsView) renderTokenTrend(w, h int) string {
	filtered := v.filterTokens()

	// Build per-model series
	modelData := make(map[string][]float64)
	var dates []string
	for _, dmt := range filtered {
		dates = append(dates, dmt.Date)
		for model, tokens := range dmt.TokensByModel {
			name := metrics.NormalizeModelName(model)
			modelData[name] = append(modelData[name], float64(tokens))
		}
	}

	// Pad series to same length
	for name, vals := range modelData {
		for len(vals) < len(dates) {
			vals = append([]float64{0}, vals...)
		}
		modelData[name] = vals
	}

	colors := map[string]lipgloss.Color{
		"opus-4-6": v.Theme.Primary, "opus-4-5": v.Theme.Primary,
		"sonnet-4-6": v.Theme.Secondary, "sonnet-4-5": v.Theme.Secondary,
		"haiku-4-5": v.Theme.Orange, "haiku-3-5": v.Theme.Orange,
	}

	var series []components.LineChartData
	for name, vals := range modelData {
		color := v.Theme.Cyan
		if c, ok := colors[name]; ok {
			color = c
		}
		series = append(series, components.LineChartData{Label: name, Values: vals, Color: color})
	}

	chart := components.RenderLineChart(series, dates, w-6, h-4, v.Theme)
	return components.Panel("Token Trend (by Model)", chart, w, h, v.Theme)
}

func (v *AnalyticsView) renderMidRow(totalW, h int) string {
	leftW := totalW / 2
	rightW := totalW - leftW

	// Cost trend
	filtered := v.filterTokens()
	var costVals []float64
	var costDates []string
	for _, dmt := range filtered {
		var dayCost float64
		for model, tokens := range dmt.TokensByModel {
			p := config.LookupPricing(model)
			// Use blended cost estimate
			mu := v.Stats.ModelUsage[model]
			total := mu.InputTokens + mu.OutputTokens + mu.CacheReadInputTokens + mu.CacheCreationInputTokens
			if total > 0 {
				fullCost := config.ComputeCost(model, mu.InputTokens, mu.OutputTokens, mu.CacheReadInputTokens, mu.CacheCreationInputTokens)
				dayCost += float64(tokens) * fullCost / float64(total)
			} else {
				dayCost += float64(tokens) * p.InputPrice / 1_000_000
			}
		}
		costVals = append(costVals, dayCost)
		costDates = append(costDates, dmt.Date)
	}

	costSeries := []components.LineChartData{{Label: "Cost", Values: costVals, Color: v.Theme.Green}}
	costChart := components.RenderLineChart(costSeries, costDates, leftW-6, h-4, v.Theme)
	leftPanel := components.Panel("Cost Trend ($)", costChart, leftW, h, v.Theme)

	// Cache hit rate
	var rates []float64
	var rateDates []string
	for _, dmt := range filtered {
		var cacheR, input int64
		for model, tokens := range dmt.TokensByModel {
			mu := v.Stats.ModelUsage[model]
			t := mu.InputTokens + mu.CacheReadInputTokens
			if t > 0 {
				ratio := float64(mu.CacheReadInputTokens) / float64(t)
				cacheR += int64(float64(tokens) * ratio)
				input += int64(float64(tokens) * (1 - ratio))
			}
		}
		rates = append(rates, metrics.CacheHitRate(cacheR, input))
		rateDates = append(rateDates, dmt.Date)
	}

	cacheSeries := []components.LineChartData{{Label: "Cache%", Values: rates, Color: v.Theme.Cyan}}
	cacheChart := components.RenderLineChart(cacheSeries, rateDates, rightW-6, h-4, v.Theme)
	rightPanel := components.Panel("Cache Hit Rate (%)", cacheChart, rightW, h, v.Theme)

	return components.HStack([]string{leftPanel, rightPanel})
}

func (v *AnalyticsView) renderBotRow(totalW, h int) string {
	leftW := totalW / 2
	rightW := totalW - leftW

	// Hourly histogram
	overview := metrics.ComputeOverview(v.Stats)
	histo := components.RenderHourlyHistogram(overview.HourCounts, leftW-6, h-4, v.Theme)
	leftPanel := components.Panel("Hourly Activity", histo, leftW, h, v.Theme)

	// Model cost breakdown - compact table
	dim := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	fg := lipgloss.NewStyle().Foreground(v.Theme.Fg).Bold(true)
	innerW := rightW - 6

	var lines []string
	for _, m := range overview.ModelDistrib {
		if m.Percent < 0.05 {
			continue
		}
		color := v.Theme.ModelColor(m.Model)
		barW := innerW - 28
		if barW < 3 {
			barW = 3
		}
		filled := min(int(m.Percent/100*float64(barW)), barW)
		if filled < 1 && m.Percent > 0 {
			filled = 1
		}
		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled))

		name := lipgloss.NewStyle().Foreground(color).Width(11).MaxWidth(11).Render(m.Model)
		cost := fg.Width(11).Align(lipgloss.Right).Render(metrics.FormatCost(m.Cost))
		pct := dim.Width(5).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f%%", m.Percent))

		line := name + " " + cost + "\n" + bar + " " + pct
		lines = append(lines, lipgloss.NewStyle().MaxWidth(innerW).Render(line))
	}

	content := strings.Join(lines, "\n")
	rightPanel := components.Panel("Model Cost Breakdown", content, rightW, h, v.Theme)

	return components.HStack([]string{leftPanel, rightPanel})
}
