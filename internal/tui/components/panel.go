package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

// Panel renders a bordered panel with a title, strict width/height constraints.
// This is the core layout primitive - content NEVER overflows.
func Panel(title string, content string, width, height int, theme common.Theme) string {
	if width < 4 || height < 3 {
		return ""
	}

	innerW := width - 2 // border left + right
	innerH := height - 2 // border top + bottom

	// Title line
	titleH := 0
	if title != "" {
		titleH = 1
	}
	contentH := innerH - titleH

	// Clip content lines to innerW and contentH
	lines := strings.Split(content, "\n")
	var clipped []string
	for i, line := range lines {
		if i >= contentH {
			break
		}
		// Hard clip line width
		clipped = append(clipped, clipToWidth(line, innerW-2)) // -2 for padding
	}
	// Pad to fill height
	for len(clipped) < contentH {
		clipped = append(clipped, "")
	}

	body := strings.Join(clipped, "\n")

	if title != "" {
		titleStr := lipgloss.NewStyle().
			Foreground(theme.FgDim).
			Bold(true).
			Render(title)
		body = titleStr + "\n" + body
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.FgDim).
		Padding(0, 1).
		Width(innerW).
		MaxWidth(width).
		Height(innerH).
		MaxHeight(height)

	return boxStyle.Render(body)
}

// clipToWidth hard-clips a string to maxW visible characters.
func clipToWidth(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= maxW {
		return s
	}
	// Strip ANSI, clip runes, but we lose styling.
	// Better approach: use MaxWidth style
	return lipgloss.NewStyle().MaxWidth(maxW).Render(s)
}

// HStack places panels side by side, each with exact width.
func HStack(panels []string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

// VStack places panels vertically.
func VStack(panels []string) string {
	return lipgloss.JoinVertical(lipgloss.Left, panels...)
}
