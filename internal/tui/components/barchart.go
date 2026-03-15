package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

// BarItem represents one bar in a horizontal bar chart.
type BarItem struct {
	Label   string
	Value   float64
	Percent float64
	Color   lipgloss.Color
}

// RenderHorizontalBars renders a horizontal bar chart.
func RenderHorizontalBars(items []BarItem, width int, theme common.Theme) string {
	if len(items) == 0 || width < 20 {
		return ""
	}

	maxLabelW := 0
	for _, item := range items {
		if len(item.Label) > maxLabelW {
			maxLabelW = len(item.Label)
		}
	}

	pctW := 5
	barW := max(width-maxLabelW-pctW-4, 5)

	var lines []string
	for _, item := range items {
		label := lipgloss.NewStyle().
			Foreground(theme.Fg).
			Width(maxLabelW).
			Render(item.Label)

		filled := min(int(item.Percent/100*float64(barW)), barW)
		empty := barW - filled

		bar := lipgloss.NewStyle().Foreground(item.Color).Render(strings.Repeat("█", filled))
		bg := lipgloss.NewStyle().Foreground(theme.FgDim).Render(strings.Repeat("░", empty))

		pct := lipgloss.NewStyle().
			Foreground(theme.FgDim).
			Width(pctW).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%.0f%%", item.Percent))

		lines = append(lines, fmt.Sprintf(" %s %s%s %s", label, bar, bg, pct))
	}

	return strings.Join(lines, "\n")
}
