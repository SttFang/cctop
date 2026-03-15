package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

// LineChartData represents a data series for the line chart.
type LineChartData struct {
	Label  string
	Values []float64
	Color  lipgloss.Color
}

// sparkBlocks are vertical bar characters for sparkline rendering.
var sparkBlocks = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderLineChart renders a sparkline-style chart with Y-axis labels.
func RenderLineChart(series []LineChartData, labels []string, width, height int, theme common.Theme) string {
	if len(series) == 0 || width < 10 || height < 3 {
		return lipgloss.NewStyle().Foreground(theme.FgDim).Render("No data")
	}

	yLabelW := 8
	chartW := width - yLabelW - 1
	if chartW < 5 {
		chartW = 5
	}
	chartH := height - 2 // room for X labels and legend
	if chartH < 1 {
		chartH = 1
	}

	// Find global min/max
	minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
	for _, s := range series {
		for _, v := range s.Values {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if minVal == maxVal {
		maxVal = minVal + 1
	}
	// Start Y from 0 if all values positive
	if minVal > 0 {
		minVal = 0
	}
	valRange := maxVal - minVal

	dimStyle := lipgloss.NewStyle().Foreground(theme.FgDim)

	// For each series, resample to chartW data points
	type cellData struct {
		level int // 0-8 block level within this row
		color lipgloss.Color
	}

	// Grid: chartH rows × chartW cols
	grid := make([][]cellData, chartH)
	for r := range grid {
		grid[r] = make([]cellData, chartW)
	}

	for _, s := range series {
		if len(s.Values) == 0 {
			continue
		}

		// Resample values to chartW points
		resampled := resample(s.Values, chartW)

		for x, v := range resampled {
			// Map value to total rows (each row has 8 sub-levels)
			totalLevels := chartH * 8
			level := int((v - minVal) / valRange * float64(totalLevels-1))
			if level < 0 {
				level = 0
			}
			if level >= totalLevels {
				level = totalLevels - 1
			}

			// Fill from bottom up
			fullRows := level / 8
			partialLevel := level % 8

			// Fill complete rows
			for r := 0; r < fullRows && r < chartH; r++ {
				row := chartH - 1 - r
				if grid[row][x].level < 8 {
					grid[row][x] = cellData{level: 8, color: s.Color}
				}
			}
			// Partial top row
			if fullRows < chartH {
				row := chartH - 1 - fullRows
				if partialLevel > 0 && grid[row][x].level < partialLevel {
					grid[row][x] = cellData{level: partialLevel, color: s.Color}
				}
			}
		}
	}

	// Render grid
	var rows []string
	for r := 0; r < chartH; r++ {
		// Y-axis label
		var yLabel string
		switch {
		case r == 0:
			yLabel = formatAxisVal(maxVal)
		case r == chartH/2:
			yLabel = formatAxisVal(minVal + valRange/2)
		case r == chartH-1:
			yLabel = formatAxisVal(minVal)
		}
		yStr := dimStyle.Render(fmt.Sprintf("%*s", yLabelW-1, yLabel)) + dimStyle.Render("│")

		var line strings.Builder
		line.WriteString(yStr)
		for c := 0; c < chartW; c++ {
			cell := grid[r][c]
			if cell.level == 0 {
				line.WriteRune(' ')
			} else {
				ch := sparkBlocks[cell.level]
				line.WriteString(lipgloss.NewStyle().Foreground(cell.color).Render(string(ch)))
			}
		}
		rows = append(rows, line.String())
	}

	// X-axis line
	rows = append(rows, dimStyle.Render(strings.Repeat(" ", yLabelW)+"└"+strings.Repeat("─", chartW)))

	// X-axis labels
	if len(labels) > 0 {
		xLine := make([]byte, chartW)
		for i := range xLine {
			xLine[i] = ' '
		}

		numLabels := chartW / 8
		if numLabels < 2 {
			numLabels = 2
		}
		if numLabels > len(labels) {
			numLabels = len(labels)
		}
		step := max(len(labels)/numLabels, 1)

		for i := 0; i < len(labels); i += step {
			pos := i * chartW / len(labels)
			lbl := labels[i]
			if len(lbl) > 5 {
				lbl = lbl[5:] // strip year "2026-"
			}
			if pos+len(lbl) <= chartW {
				copy(xLine[pos:], lbl)
			}
		}
		rows = append(rows, strings.Repeat(" ", yLabelW+1)+dimStyle.Render(string(xLine)))
	}

	// Legend
	if len(series) > 1 {
		var legend strings.Builder
		legend.WriteString(strings.Repeat(" ", yLabelW+1))
		for i, s := range series {
			if i > 0 {
				legend.WriteString("  ")
			}
			legend.WriteString(lipgloss.NewStyle().Foreground(s.Color).Render("━━"))
			legend.WriteString(" ")
			legend.WriteString(dimStyle.Render(s.Label))
		}
		rows = append(rows, legend.String())
	}

	return strings.Join(rows, "\n")
}

// resample resamples values to targetLen points using linear interpolation.
func resample(values []float64, targetLen int) []float64 {
	if len(values) == 0 {
		return nil
	}
	if len(values) == 1 {
		result := make([]float64, targetLen)
		for i := range result {
			result[i] = values[0]
		}
		return result
	}
	if len(values) >= targetLen {
		// Downsample: pick evenly spaced points
		result := make([]float64, targetLen)
		for i := range result {
			idx := i * (len(values) - 1) / (targetLen - 1)
			result[i] = values[idx]
		}
		return result
	}
	// Upsample: linear interpolation
	result := make([]float64, targetLen)
	for i := range result {
		t := float64(i) / float64(targetLen-1) * float64(len(values)-1)
		lo := int(t)
		hi := lo + 1
		if hi >= len(values) {
			hi = len(values) - 1
		}
		frac := t - float64(lo)
		result[i] = values[lo]*(1-frac) + values[hi]*frac
	}
	return result
}

func formatAxisVal(v float64) string {
	switch {
	case v >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", v/1_000_000_000)
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.0fK", v/1_000)
	case v >= 10:
		return fmt.Sprintf("%.0f", v)
	case v >= 1:
		return fmt.Sprintf("%.1f", v)
	default:
		return fmt.Sprintf("%.2f", v)
	}
}
