package components

// StatCard rendering is now done inline in overview.go using Panel primitives.
// This file is kept for backwards compatibility with tests.

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

// StatCardProps configures a stat card.
type StatCardProps struct {
	Title   string
	Value   string
	SubText string
	DeltaUp bool
	Width   int
	Height  int
	Theme   common.Theme
}

// RenderStatCard renders a stat card component.
func RenderStatCard(p StatCardProps) string {
	if p.Width < 8 {
		return ""
	}

	innerW := max(p.Width-4, 4)

	dim := lipgloss.NewStyle().Foreground(p.Theme.FgDim).Width(innerW).Align(lipgloss.Center)
	bold := lipgloss.NewStyle().Foreground(p.Theme.Fg).Bold(true).Width(innerW).Align(lipgloss.Center)

	var lines []string
	lines = append(lines, dim.Render(p.Title))
	lines = append(lines, bold.Render(p.Value))
	if p.SubText != "" {
		prefix := ""
		if p.DeltaUp {
			prefix = lipgloss.NewStyle().Foreground(p.Theme.Green).Render("▲ ")
		}
		lines = append(lines, dim.Render(prefix+p.SubText))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Theme.FgDim).
		Padding(0, 1).
		Width(p.Width - 2).
		MaxWidth(p.Width).
		Render(strings.Join(lines, "\n"))
}

// RenderStatCards renders a row of stat cards.
func RenderStatCards(cards []StatCardProps) string {
	rendered := make([]string, len(cards))
	for i, c := range cards {
		rendered[i] = RenderStatCard(c)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
