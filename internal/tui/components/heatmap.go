package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

var hBlocks = []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// RenderHourlyHistogram renders a 24-bar histogram of hourly activity.
func RenderHourlyHistogram(counts [24]int, width, height int, theme common.Theme) string {
	if height < 3 {
		return ""
	}

	maxVal := 0
	for _, c := range counts {
		if c > maxVal {
			maxVal = c
		}
	}
	if maxVal == 0 {
		return lipgloss.NewStyle().Foreground(theme.FgDim).Render("No activity data")
	}

	barHeight := height - 2 // room for X labels + padding
	if barHeight < 1 {
		barHeight = 1
	}

	// Each hour gets barChars characters
	barChars := (width - 4) / 24
	if barChars < 1 {
		barChars = 1
	}
	if barChars > 3 {
		barChars = 3
	}

	// Build rows from top to bottom
	var rows []string
	for row := barHeight; row >= 1; row-- {
		var line strings.Builder
		line.WriteString(" ")
		for h := 0; h < 24; h++ {
			ratio := float64(counts[h]) / float64(maxVal)
			cellLevel := ratio * float64(barHeight)

			var ch string
			if cellLevel >= float64(row) {
				ch = "█"
			} else if cellLevel > float64(row-1) {
				frac := cellLevel - float64(row-1)
				idx := int(frac * 8)
				if idx > 8 {
					idx = 8
				}
				if idx < 0 {
					idx = 0
				}
				ch = hBlocks[idx]
			} else {
				ch = " "
			}

			// Color based on intensity
			color := theme.Green
			if ratio > 0.7 {
				color = theme.Red
			} else if ratio > 0.4 {
				color = theme.Yellow
			}

			bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat(ch, barChars))
			line.WriteString(bar)
		}
		rows = append(rows, line.String())
	}

	// X-axis labels
	dimStyle := lipgloss.NewStyle().Foreground(theme.FgDim)
	var labelLine strings.Builder
	labelLine.WriteString(" ")

	labelStep := 6
	if barChars >= 2 {
		labelStep = 3
	}

	for h := 0; h < 24; h++ {
		if h%labelStep == 0 {
			label := fmt.Sprintf("%-*d", barChars, h)
			labelLine.WriteString(dimStyle.Render(label))
		} else {
			labelLine.WriteString(strings.Repeat(" ", barChars))
		}
	}
	rows = append(rows, labelLine.String())

	return strings.Join(rows, "\n")
}
