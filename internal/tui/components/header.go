package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

var tabNames = []string{"Overview", "Sessions", "Analytics", "Tools"}

// RenderHeader renders the top header bar.
func RenderHeader(activeTab, width, refreshSecs int, theme common.Theme) string {
	titleStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(theme.FgDim)
	tabStyle := lipgloss.NewStyle().Foreground(theme.Fg).Bold(true)

	title := titleStyle.Render(" CCtop")
	sep := dimStyle.Render(" ── ")
	tabName := tabStyle.Render(tabNames[activeTab])
	date := dimStyle.Render(time.Now().Format("2006-01-02"))
	refresh := dimStyle.Render(fmt.Sprintf("%ds", refreshSecs))

	left := title + sep + tabName
	right := date + dimStyle.Render(" ─ ") + refresh + " "

	fillW := max(width-lipgloss.Width(left)-lipgloss.Width(right), 0)
	fill := dimStyle.Render(repeatChar('─', fillW))

	return left + fill + right
}

func repeatChar(ch rune, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}
