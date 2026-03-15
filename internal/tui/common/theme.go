package common

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme for the TUI.
type Theme struct {
	Bg          lipgloss.Color
	BgSurface   lipgloss.Color
	BgHighlight lipgloss.Color
	Fg          lipgloss.Color
	FgDim       lipgloss.Color
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Green       lipgloss.Color
	Red         lipgloss.Color
	Yellow      lipgloss.Color
	Cyan        lipgloss.Color
	Orange      lipgloss.Color
}

var DarkTheme = Theme{
	Bg:          lipgloss.Color("#1a1b26"),
	BgSurface:   lipgloss.Color("#24283b"),
	BgHighlight: lipgloss.Color("#2f3549"),
	Fg:          lipgloss.Color("#c0caf5"),
	FgDim:       lipgloss.Color("#565f89"),
	Primary:     lipgloss.Color("#7aa2f7"),
	Secondary:   lipgloss.Color("#bb9af7"),
	Green:       lipgloss.Color("#9ece6a"),
	Red:         lipgloss.Color("#f7768e"),
	Yellow:      lipgloss.Color("#e0af68"),
	Cyan:        lipgloss.Color("#7dcfff"),
	Orange:      lipgloss.Color("#ff9e64"),
}

var LightTheme = Theme{
	Bg:          lipgloss.Color("#d5d6db"),
	BgSurface:   lipgloss.Color("#e9e9ec"),
	BgHighlight: lipgloss.Color("#c4c8da"),
	Fg:          lipgloss.Color("#343b58"),
	FgDim:       lipgloss.Color("#9699a3"),
	Primary:     lipgloss.Color("#2e7de9"),
	Secondary:   lipgloss.Color("#7847bd"),
	Green:       lipgloss.Color("#587539"),
	Red:         lipgloss.Color("#c64343"),
	Yellow:      lipgloss.Color("#8c6c3e"),
	Cyan:        lipgloss.Color("#007197"),
	Orange:      lipgloss.Color("#b15c00"),
}

// ModelColor returns the theme color for a normalized model name.
func (t Theme) ModelColor(model string) lipgloss.Color {
	switch {
	case containsStr(model, "opus"):
		return t.Primary
	case containsStr(model, "sonnet"):
		return t.Secondary
	case containsStr(model, "haiku"):
		return t.Orange
	default:
		return t.Cyan
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// CostColor returns a color based on cost magnitude.
func (t Theme) CostColor(cost float64) lipgloss.Color {
	switch {
	case cost >= 5:
		return t.Red
	case cost >= 1:
		return t.Yellow
	default:
		return t.Green
	}
}
