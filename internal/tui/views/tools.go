package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/store"
	"github.com/fanghanjun/cctop/internal/tui/common"
	"github.com/fanghanjun/cctop/internal/tui/components"
)

type ToolCategory struct {
	Name    string
	Tools   []string
	Percent float64
	Total   int
}

type ToolsView struct {
	Layout    common.Layout
	Theme     common.Theme
	ToolFreqs []store.ToolFreq
	Period    string
}

var toolCategories = []struct {
	name  string
	tools []string
}{
	{"Read Ops", []string{"Read", "Grep", "Glob"}},
	{"Write Ops", []string{"Edit", "Write"}},
	{"Exec Ops", []string{"Bash", "Agent"}},
	{"Search Ops", []string{"WebSearch", "WebFetch", "LSP"}},
}

func (v *ToolsView) View() string {
	if len(v.ToolFreqs) == 0 {
		return lipgloss.NewStyle().
			Foreground(v.Theme.FgDim).
			Width(v.Layout.ContentW).
			Height(v.Layout.ContentH).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No tool data yet. Sessions need to be indexed first.")
	}

	w := v.Layout.ContentW
	h := v.Layout.ContentH

	// Top: tool ranking (55%)
	topH := h * 55 / 100
	if topH < 8 {
		topH = 8
	}
	topPanel := v.renderRanking(w, topH)

	// Bottom: category + placeholder
	botH := h - topH
	if botH < 5 {
		botH = 5
	}
	botPanel := v.renderBottom(w, botH)

	return components.VStack([]string{topPanel, botPanel})
}

func (v *ToolsView) renderRanking(w, h int) string {
	header := lipgloss.NewStyle().Foreground(v.Theme.Primary).Bold(true)
	dim := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	fg := lipgloss.NewStyle().Foreground(v.Theme.Fg)
	innerW := w - 6

	var lines []string

	// Header row
	barW := innerW - 40
	if barW < 5 {
		barW = 5
	}
	hdr := fmt.Sprintf("  %-3s %-12s %8s %6s  %-*s", "#", "Tool", "Calls", "Pct", barW, "Distribution")
	lines = append(lines, header.MaxWidth(innerW).Render(hdr))
	lines = append(lines, dim.Render(strings.Repeat("─", innerW)))

	maxCalls := 0
	if len(v.ToolFreqs) > 0 {
		maxCalls = v.ToolFreqs[0].TotalCalls
	}

	maxRows := h - 5
	for i, tf := range v.ToolFreqs {
		if i >= maxRows {
			break
		}

		filled := 0
		if maxCalls > 0 {
			filled = min(tf.TotalCalls*barW/maxCalls, barW)
		}
		bar := lipgloss.NewStyle().Foreground(v.Theme.Primary).Render(strings.Repeat("█", filled))

		row := fmt.Sprintf("  %-3d %-12s %8d %5.1f%%  ",
			i+1, tf.Name, tf.TotalCalls, tf.Percent)
		line := fg.Render(row) + bar
		lines = append(lines, lipgloss.NewStyle().MaxWidth(innerW).Render(line))
	}

	content := strings.Join(lines, "\n")
	return components.Panel("Tool Usage Ranking", content, w, h, v.Theme)
}

func (v *ToolsView) renderBottom(totalW, h int) string {
	leftW := totalW * 45 / 100
	rightW := totalW - leftW

	// Categories
	cats := v.computeCategories()
	dim := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	innerW := leftW - 6
	barW := innerW - 20
	if barW < 5 {
		barW = 5
	}

	colors := []lipgloss.Color{v.Theme.Primary, v.Theme.Secondary, v.Theme.Orange, v.Theme.Cyan}

	var catLines []string
	for i, cat := range cats {
		if cat.Percent < 0.1 {
			continue
		}
		color := colors[0]
		if i < len(colors) {
			color = colors[i]
		}
		filled := min(int(cat.Percent/100*float64(barW)), barW)
		if filled < 1 {
			filled = 1
		}
		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled))
		label := lipgloss.NewStyle().Foreground(color).Width(10).Render(cat.Name)
		pct := dim.Width(5).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f%%", cat.Percent))
		detail := dim.Render("  (" + strings.Join(cat.Tools, "+") + ")")

		line := label + " " + bar + " " + pct + "\n" + detail
		catLines = append(catLines, lipgloss.NewStyle().MaxWidth(innerW).Render(line))
	}

	leftPanel := components.Panel("Tool Category", strings.Join(catLines, "\n"), leftW, h, v.Theme)

	// Right: top 5 summary
	var summLines []string
	for i, tf := range v.ToolFreqs {
		if i >= 5 {
			break
		}
		summLines = append(summLines,
			fmt.Sprintf("  %d. %-10s %6d calls (%.1f%%)", i+1, tf.Name, tf.TotalCalls, tf.Percent))
	}
	rightPanel := components.Panel("Top 5 Tools", strings.Join(summLines, "\n"), rightW, h, v.Theme)

	return components.HStack([]string{leftPanel, rightPanel})
}

func (v *ToolsView) computeCategories() []ToolCategory {
	freqMap := make(map[string]int)
	var total int
	for _, tf := range v.ToolFreqs {
		freqMap[tf.Name] = tf.TotalCalls
		total += tf.TotalCalls
	}

	var cats []ToolCategory
	var categorized int
	for _, def := range toolCategories {
		cat := ToolCategory{Name: def.name, Tools: def.tools}
		for _, tool := range def.tools {
			cat.Total += freqMap[tool]
		}
		categorized += cat.Total
		cats = append(cats, cat)
	}

	if uncategorized := total - categorized; uncategorized > 0 {
		cats = append(cats, ToolCategory{Name: "Other", Tools: []string{"..."}, Total: uncategorized})
	}

	for i := range cats {
		if total > 0 {
			cats[i].Percent = float64(cats[i].Total) / float64(total) * 100
		}
	}
	return cats
}
