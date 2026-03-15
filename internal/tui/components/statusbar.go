package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

type tabDef struct {
	key  string
	name string
}

var tabs = []tabDef{
	{"1", "Overview"},
	{"2", "Sessions"},
	{"3", "Analytics"},
	{"4", "Tools"},
}

// RenderStatusBar renders the bottom status bar.
func RenderStatusBar(activeTab, width int, extraHints string, theme common.Theme) string {
	activeStyle := lipgloss.NewStyle().
		Background(theme.Primary).
		Foreground(theme.Bg).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(theme.FgDim).
		Padding(0, 1)

	keyStyle := lipgloss.NewStyle().Foreground(theme.Yellow)
	hintStyle := lipgloss.NewStyle().Foreground(theme.FgDim)

	var sb strings.Builder
	for i, t := range tabs {
		label := "[" + t.key + "]" + t.name
		if i == activeTab {
			sb.WriteString(activeStyle.Render(label))
		} else {
			sb.WriteString(inactiveStyle.Render(label))
		}
	}

	sb.WriteString("  ")
	sb.WriteString(keyStyle.Render("q"))
	sb.WriteString(hintStyle.Render(":Quit"))
	sb.WriteString("  ")
	sb.WriteString(keyStyle.Render("r"))
	sb.WriteString(hintStyle.Render(":Refresh"))
	sb.WriteString("  ")
	sb.WriteString(keyStyle.Render("t"))
	sb.WriteString(hintStyle.Render(":Theme"))
	sb.WriteString("  ")
	sb.WriteString(keyStyle.Render("?"))
	sb.WriteString(hintStyle.Render(":Help"))

	if extraHints != "" {
		sb.WriteString("  ")
		sb.WriteString(extraHints)
	}

	content := sb.String()
	contentW := lipgloss.Width(content)
	if contentW < width {
		content += strings.Repeat(" ", width-contentW)
	}
	return content
}
