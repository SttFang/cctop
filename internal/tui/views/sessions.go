package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/metrics"
	"github.com/fanghanjun/cctop/internal/store"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

// SessionsView renders the sessions browser.
type SessionsView struct {
	Layout    common.Layout
	Theme     common.Theme
	Sessions  []store.SessionRow
	Total     int
	Cursor    int
	Page      int
	PageSize  int
	SortBy    string
	SortDesc  bool
	Search    string
	Searching bool
	Indexing  bool
	IndexPct  float64

	// Detail view
	ShowDetail bool
	DetailData *SessionDetailData
}

// SessionDetailData holds data for the session detail view.
type SessionDetailData struct {
	Session   store.SessionRow
	ToolCalls map[string]int
}

func (v *SessionsView) View() string {
	if v.ShowDetail && v.DetailData != nil {
		return v.renderDetail()
	}
	return v.renderList()
}

func (v *SessionsView) renderList() string {
	var sections []string

	// Filter bar
	sections = append(sections, v.renderFilterBar())

	// Table
	sections = append(sections, v.renderTable())

	// Preview
	if v.Layout.PreviewH > 0 && len(v.Sessions) > 0 && v.Cursor < len(v.Sessions) {
		sections = append(sections, v.renderPreview())
	}

	return strings.Join(sections, "\n")
}

func (v *SessionsView) renderFilterBar() string {
	dimStyle := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	primaryStyle := lipgloss.NewStyle().Foreground(v.Theme.Primary)

	var line strings.Builder
	line.WriteString("  ")

	// Sort indicator
	sortLabel := "Time"
	switch v.SortBy {
	case "cost":
		sortLabel = "Cost"
	case "tokens":
		sortLabel = "Tokens"
	case "duration":
		sortLabel = "Duration"
	}
	sortDir := "▼"
	if !v.SortDesc {
		sortDir = "▲"
	}
	line.WriteString(dimStyle.Render("Sort: "))
	line.WriteString(primaryStyle.Render(sortLabel + " " + sortDir))
	line.WriteString("  ")

	// Search
	if v.Searching {
		line.WriteString(primaryStyle.Render("Search: "))
		line.WriteString(lipgloss.NewStyle().Foreground(v.Theme.Fg).Render(v.Search + "█"))
	} else if v.Search != "" {
		line.WriteString(dimStyle.Render("Search: "))
		line.WriteString(lipgloss.NewStyle().Foreground(v.Theme.Fg).Render(v.Search))
	}

	// Status
	line.WriteString(strings.Repeat(" ", max(v.Layout.ContentW-lipgloss.Width(line.String())-20, 0)))
	if v.Indexing {
		line.WriteString(lipgloss.NewStyle().Foreground(v.Theme.Yellow).Render(fmt.Sprintf("Indexing %.0f%%", v.IndexPct)))
	} else {
		line.WriteString(dimStyle.Render(fmt.Sprintf("%d sessions", v.Total)))
	}

	result := line.String()

	if v.Layout.FilterBarH >= 2 {
		return result + "\n" + dimStyle.Render(strings.Repeat("─", v.Layout.ContentW))
	}
	return result
}

func (v *SessionsView) renderTable() string {
	if len(v.Sessions) == 0 {
		style := lipgloss.NewStyle().
			Foreground(v.Theme.FgDim).
			Width(v.Layout.ContentW).
			Height(v.Layout.TableH).
			Align(lipgloss.Center, lipgloss.Center)
		if v.Indexing {
			return style.Render("Indexing sessions...")
		}
		return style.Render("No sessions indexed yet")
	}

	headerStyle := lipgloss.NewStyle().Foreground(v.Theme.Primary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(v.Theme.FgDim)

	// Column widths
	w := v.Layout.ContentW
	projW, modelW, timeW, durW, tokW, costW := 16, 12, 16, 6, 8, 8

	// Determine visible columns based on width
	showTokens := w >= 100
	showModel := w >= 90

	var header strings.Builder
	header.WriteString("  ")
	header.WriteString(headerStyle.Render(padRight("Project", projW)))
	if showModel {
		header.WriteString(headerStyle.Render(padRight("Model", modelW)))
	}
	header.WriteString(headerStyle.Render(padRight("Time", timeW)))
	header.WriteString(headerStyle.Render(padRight("Dur", durW)))
	if showTokens {
		header.WriteString(headerStyle.Render(padLeft("Tokens", tokW)))
	}
	header.WriteString(headerStyle.Render(padLeft("Cost", costW)))

	var lines []string
	lines = append(lines, header.String())
	lines = append(lines, dimStyle.Render(" "+strings.Repeat("─", w-2)))

	maxRows := v.Layout.TableH - 3 // header + separator + summary
	for i, sess := range v.Sessions {
		if i >= maxRows {
			break
		}

		isSelected := i == v.Cursor
		var row strings.Builder

		if isSelected {
			row.WriteString(lipgloss.NewStyle().Foreground(v.Theme.Primary).Render("▸ "))
		} else {
			row.WriteString("  ")
		}

		rowStyle := lipgloss.NewStyle().Foreground(v.Theme.Fg)
		if isSelected {
			rowStyle = rowStyle.Bold(true)
		}

		row.WriteString(rowStyle.Render(padRight(truncStr(sess.Project, projW-1), projW)))
		if showModel {
			modelStyle := lipgloss.NewStyle().Foreground(v.Theme.ModelColor(metrics.NormalizeModelName(sess.Model)))
			row.WriteString(modelStyle.Render(padRight(truncStr(metrics.NormalizeModelName(sess.Model), modelW-1), modelW)))
		}

		timeStr := formatSessionTime(sess.StartTime)
		row.WriteString(rowStyle.Render(padRight(timeStr, timeW)))
		row.WriteString(rowStyle.Render(padRight(metrics.FormatDuration(sess.DurationMs), durW)))
		if showTokens {
			row.WriteString(rowStyle.Render(padLeft(metrics.FormatTokens(sess.TotalTokens), tokW)))
		}

		costStyle := lipgloss.NewStyle().Foreground(v.Theme.CostColor(sess.Cost))
		row.WriteString(costStyle.Render(padLeft(metrics.FormatCost(sess.Cost), costW)))

		if isSelected {
			lines = append(lines, lipgloss.NewStyle().Background(v.Theme.BgHighlight).Render(
				row.String()+strings.Repeat(" ", max(w-lipgloss.Width(row.String()), 0))))
		} else {
			lines = append(lines, row.String())
		}
	}

	// Pad remaining rows
	for len(lines) < v.Layout.TableH-1 {
		lines = append(lines, "")
	}

	// Summary line
	totalPages := (v.Total + v.PageSize - 1) / v.PageSize
	if v.PageSize <= 0 {
		totalPages = 1
	}
	summary := dimStyle.Render(fmt.Sprintf(" %d sessions  Page %d/%d", v.Total, v.Page+1, totalPages))
	lines = append(lines, dimStyle.Render(strings.Repeat("─", w-2)))
	lines = append(lines, summary)

	return strings.Join(lines, "\n")
}

func (v *SessionsView) renderPreview() string {
	if v.Cursor >= len(v.Sessions) {
		return ""
	}
	sess := v.Sessions[v.Cursor]

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.Theme.FgDim).
		Width(v.Layout.ContentW - 2)

	dimStyle := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	fgStyle := lipgloss.NewStyle().Foreground(v.Theme.Fg)

	title := dimStyle.Render("Preview ── ") +
		lipgloss.NewStyle().Foreground(v.Theme.Primary).Render(sess.Project) +
		dimStyle.Render(" ── ") +
		lipgloss.NewStyle().Foreground(v.Theme.ModelColor(metrics.NormalizeModelName(sess.Model))).Render(metrics.NormalizeModelName(sess.Model))

	prompt := truncStr(sess.FirstPrompt, v.Layout.ContentW-10)
	content := title + "\n" + fgStyle.Render(fmt.Sprintf("%q", prompt)) +
		dimStyle.Render(fmt.Sprintf("  %s  %s", metrics.FormatTokens(sess.TotalTokens), metrics.FormatCost(sess.Cost)))

	return borderStyle.Render(content)
}

func (v *SessionsView) renderDetail() string {
	d := v.DetailData
	sess := d.Session

	var sections []string

	// Header
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.Theme.FgDim)
	dimStyle := lipgloss.NewStyle().Foreground(v.Theme.FgDim)
	fgStyle := lipgloss.NewStyle().Foreground(v.Theme.Fg)

	header := lipgloss.NewStyle().Foreground(v.Theme.Primary).Bold(true).Render(sess.Project) +
		dimStyle.Render(" / ") +
		lipgloss.NewStyle().Foreground(v.Theme.ModelColor(metrics.NormalizeModelName(sess.Model))).Render(metrics.NormalizeModelName(sess.Model)) +
		dimStyle.Render("      ") +
		fgStyle.Render(formatSessionTime(sess.StartTime)) +
		dimStyle.Render("  Duration: ") +
		fgStyle.Render(metrics.FormatDuration(sess.DurationMs))

	prompt := truncStr(sess.FirstPrompt, v.Layout.ContentW-10)
	headerBox := borderStyle.Width(v.Layout.ContentW - 2).Render(header + "\n" + fgStyle.Render(fmt.Sprintf("%q", prompt)))
	sections = append(sections, headerBox)

	// Token and Tools side by side
	tokenW := v.Layout.ContentW/2 - 2
	toolW := v.Layout.ContentW - tokenW - 4

	// Token bars
	maxTok := max64(sess.InputTokens, sess.OutputTokens, sess.CacheRead, sess.CacheCreate)
	tokenLines := []string{
		renderTokenBar("Input", sess.InputTokens, maxTok, tokenW, v.Theme),
		renderTokenBar("Output", sess.OutputTokens, maxTok, tokenW, v.Theme),
		renderTokenBar("CacheRd", sess.CacheRead, maxTok, tokenW, v.Theme),
		renderTokenBar("CacheWr", sess.CacheCreate, maxTok, tokenW, v.Theme),
		"",
		dimStyle.Render(fmt.Sprintf("  Total: %s  Cost: %s",
			metrics.FormatTokens(sess.TotalTokens), metrics.FormatCost(sess.Cost))),
	}
	tokenBox := borderStyle.Width(tokenW).Render(
		dimStyle.Render("Tokens") + "\n" + strings.Join(tokenLines, "\n"))

	// Tool bars
	var toolLines []string
	if d.ToolCalls != nil {
		sortedTools := sortMapByValue(d.ToolCalls)
		maxCalls := 0
		if len(sortedTools) > 0 {
			maxCalls = sortedTools[0].count
		}
		for _, t := range sortedTools {
			barW := toolW - 20
			if barW < 3 {
				barW = 3
			}
			filled := 0
			if maxCalls > 0 {
				filled = t.count * barW / maxCalls
			}
			bar := lipgloss.NewStyle().Foreground(v.Theme.Yellow).Render(strings.Repeat("█", filled))
			toolLines = append(toolLines,
				fmt.Sprintf("  %-8s %s %d", t.name, bar, t.count))
		}
	}
	toolBox := borderStyle.Width(toolW).Render(
		dimStyle.Render("Tools") + "\n" + strings.Join(toolLines, "\n"))

	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top, tokenBox, toolBox))

	// Footer hint
	sections = append(sections, dimStyle.Render("  Esc:Back"))

	return strings.Join(sections, "\n")
}

func renderTokenBar(label string, value, maxVal int64, width int, theme common.Theme) string {
	barW := width - 22
	if barW < 3 {
		barW = 3
	}
	filled := 0
	if maxVal > 0 {
		filled = int(float64(value) / float64(maxVal) * float64(barW))
	}
	empty := barW - filled

	bar := lipgloss.NewStyle().Foreground(theme.Cyan).Render(strings.Repeat("█", filled))
	bg := lipgloss.NewStyle().Foreground(theme.FgDim).Render(strings.Repeat("▒", empty))

	return fmt.Sprintf("  %-8s %s%s %s", label, bar, bg,
		lipgloss.NewStyle().Foreground(theme.Fg).Render(metrics.FormatTokens(value)))
}

func formatSessionTime(isoTime string) string {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, isoTime)
		if err != nil {
			return isoTime
		}
	}
	return t.Format("01-02 15:04")
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func padLeft(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return strings.Repeat(" ", w-len(s)) + s
}

func truncStr(s string, maxW int) string {
	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}
	if maxW <= 3 {
		return string(runes[:maxW])
	}
	return string(runes[:maxW-3]) + "..."
}

func max64(vals ...int64) int64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

type nameCount struct {
	name  string
	count int
}

func sortMapByValue(m map[string]int) []nameCount {
	result := make([]nameCount, 0, len(m))
	for k, v := range m {
		result = append(result, nameCount{k, v})
	}
	// Sort descending by count
	for i := range result {
		for j := i + 1; j < len(result); j++ {
			if result[j].count > result[i].count {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}
