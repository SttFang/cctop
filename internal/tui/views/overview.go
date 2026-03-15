package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/metrics"
	"github.com/fanghanjun/cctop/internal/tui/common"
	"github.com/fanghanjun/cctop/internal/tui/components"
)

type OverviewView struct {
	Data   *metrics.Overview
	Layout common.Layout
	Theme  common.Theme
}

func (v *OverviewView) View() string {
	if v.Data == nil {
		return lipgloss.NewStyle().
			Foreground(v.Theme.FgDim).
			Width(v.Layout.ContentW).
			Height(v.Layout.ContentH).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Loading...")
	}

	w := v.Layout.ContentW
	h := v.Layout.ContentH

	// Row 1: Stat cards (fixed 7 lines - room for 2-line subtitle)
	cardH := 7
	cardRow := v.renderCards(w, cardH)

	// Row 3: Model distribution (fixed 6 lines)
	modelH := min(len(v.Data.ModelDistrib)+3, 7)
	modelRow := v.renderModelPanel(w, modelH)

	// Row 2: Charts (fill remaining)
	chartH := h - cardH - modelH
	if chartH < 5 {
		chartH = 5
	}
	chartRow := v.renderChartRow(w, chartH)

	return components.VStack([]string{cardRow, chartRow, modelRow})
}

func (v *OverviewView) renderCards(totalW, h int) string {
	cardW := totalW / 4
	lastW := totalW - cardW*3 // absorb rounding

	dim := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	fg := lipgloss.NewStyle().Foreground(v.Theme.Fg).Bold(true)
	green := lipgloss.NewStyle().Foreground(v.Theme.Green)

	innerW := func(w int) int { return w - 4 }

	makeCard := func(title, value string, subLines []string, w int) string {
		iw := innerW(w)
		var lines []string
		lines = append(lines, dim.Width(iw).Align(lipgloss.Center).Render(title))
		lines = append(lines, fg.Width(iw).Align(lipgloss.Center).Render(value))
		for _, s := range subLines {
			lines = append(lines, dim.Width(iw).MaxWidth(iw).Align(lipgloss.Center).Render(s))
		}
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(v.Theme.FgDim).
			Width(w - 2).
			MaxWidth(w).
			Height(h - 2).
			Padding(0, 1).
			Render(strings.Join(lines, "\n"))
	}

	// Token breakdown: 2 lines showing all 4 categories
	tokenLine1 := fmt.Sprintf("in %-6s  out %-6s",
		metrics.FormatTokens(v.Data.TotalInput),
		metrics.FormatTokens(v.Data.TotalOutput))
	tokenLine2 := fmt.Sprintf("cR %-6s  cW %-6s",
		metrics.FormatTokens(v.Data.TotalCacheR),
		metrics.FormatTokens(v.Data.TotalCacheW))

	// Today cost
	var todayLine string
	if v.Data.TodayCost > 0 {
		todayLine = green.Render("▲") + " " + metrics.FormatCost(v.Data.TodayCost) + " today"
	} else {
		todayLine = "no data today"
	}

	cards := []string{
		makeCard("EST. COST", metrics.FormatCost(v.Data.TotalCost), []string{todayLine, dim.Render("(API pricing)")}, cardW),
		makeCard("TOKENS", metrics.FormatTokens(v.Data.TotalTokens), []string{tokenLine1, tokenLine2}, cardW),
		makeCard("SESSIONS", formatInt(v.Data.TotalSessions), []string{fmt.Sprintf("today %s", formatInt(v.Data.TodaySessions))}, cardW),
		makeCard("MESSAGES", formatInt(v.Data.TotalMessages), []string{fmt.Sprintf("today %s", formatInt(v.Data.TodayMessages))}, lastW),
	}

	return components.HStack(cards)
}

func (v *OverviewView) renderChartRow(totalW, h int) string {
	if v.Layout.Mode <= common.LayoutStandard {
		// Standard: full-width cost trend only
		return v.renderCostChart(totalW, h)
	}

	// Comfort/Wide: cost trend 60% + hourly 40%
	leftW := totalW * 60 / 100
	rightW := totalW - leftW
	left := v.renderCostChart(leftW, h)
	right := v.renderHourlyPanel(rightW, h)
	return components.HStack([]string{left, right})
}

func (v *OverviewView) renderCostChart(w, h int) string {
	var values []float64
	var labels []string
	for _, dc := range v.Data.DailyCosts {
		values = append(values, dc.Cost)
		labels = append(labels, dc.Date)
	}

	chartData := []components.LineChartData{
		{Label: "Cost", Values: values, Color: v.Theme.Primary},
	}
	innerW := w - 4
	innerH := h - 4
	chart := components.RenderLineChart(chartData, labels, innerW, innerH, v.Theme)
	return components.Panel("Daily Cost Trend ($)", chart, w, h, v.Theme)
}

func (v *OverviewView) renderHourlyPanel(w, h int) string {
	innerW := w - 4
	innerH := h - 4
	histo := components.RenderHourlyHistogram(v.Data.HourCounts, innerW, innerH, v.Theme)
	return components.Panel("Hourly Activity", histo, w, h, v.Theme)
}

func (v *OverviewView) renderModelPanel(totalW, h int) string {
	innerW := totalW - 6 // border + padding
	dim := lipgloss.NewStyle().Foreground(v.Theme.FgDim)

	limit := min(len(v.Data.ModelDistrib), 4)

	labelW := 12
	costW := 12
	pctW := 7
	barW := innerW - labelW - costW - pctW - 3
	if barW < 5 {
		barW = 5
	}

	var lines []string
	for i := 0; i < limit; i++ {
		m := v.Data.ModelDistrib[i]
		if m.Percent < 0.05 {
			continue
		}

		color := v.Theme.ModelColor(m.Model)
		filled := min(int(m.Percent/100*float64(barW)), barW)
		if filled < 1 && m.Percent > 0 {
			filled = 1
		}
		empty := barW - filled

		label := lipgloss.NewStyle().Foreground(color).Width(labelW).MaxWidth(labelW).Render(m.Model)
		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
			dim.Render(strings.Repeat("░", empty))
		pct := dim.Width(pctW).Align(lipgloss.Right).Render(fmt.Sprintf("%.1f%%", m.Percent))
		cost := lipgloss.NewStyle().Foreground(v.Theme.Fg).Width(costW).Align(lipgloss.Right).Render(metrics.FormatCost(m.Cost))

		// Clip entire line
		line := label + " " + bar + " " + pct + " " + cost
		lines = append(lines, lipgloss.NewStyle().MaxWidth(innerW).Render(line))
	}

	content := strings.Join(lines, "\n")
	return components.Panel("Model Distribution", content, totalW, h, v.Theme)
}

func formatInt(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return metrics.FormatTokens(int64(n))
}
