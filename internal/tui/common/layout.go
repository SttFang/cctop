package common

// LayoutMode defines the terminal size category.
type LayoutMode int

const (
	LayoutCompact  LayoutMode = iota // < 80 cols or < 24 rows
	LayoutStandard                   // 80-119 cols, 24-39 rows
	LayoutComfort                    // 120-159 cols, 40+ rows
	LayoutWide                       // 160+ cols
)

// Layout holds computed dimensions for all UI regions.
type Layout struct {
	Mode          LayoutMode
	Width, Height int

	// Fixed regions
	HeaderH    int // always 1
	StatusBarH int // always 1

	// Available content area
	ContentW int
	ContentH int

	// Overview sub-regions
	StatCardW   int
	StatCardH   int
	ChartRowH   int
	ChartLeftW  int
	ChartRightW int
	BarRowH     int

	// Sessions sub-regions
	FilterBarH int
	TableH     int
	PreviewH   int

	// Analytics sub-regions
	TopChartH int
	MidChartH int
	BotChartH int
}

// ComputeLayout calculates layout dimensions based on terminal size and active tab.
func ComputeLayout(w, h, tab int) Layout {
	l := Layout{
		Width: w, Height: h,
		HeaderH: 1, StatusBarH: 1,
	}
	l.ContentW = w
	l.ContentH = h - l.HeaderH - l.StatusBarH

	switch {
	case w < 80 || h < 24:
		l.Mode = LayoutCompact
	case w < 120:
		l.Mode = LayoutStandard
	case w < 160:
		l.Mode = LayoutComfort
	default:
		l.Mode = LayoutWide
	}

	switch tab {
	case 0:
		l.computeOverview()
	case 1:
		l.computeSessions()
	case 2:
		l.computeAnalytics()
	case 3:
		l.computeTools()
	}

	return l
}

func (l *Layout) computeOverview() {
	numCards := 4
	if l.Mode == LayoutCompact {
		numCards = 2
	}
	l.StatCardW = l.ContentW / numCards
	if l.Mode == LayoutCompact {
		l.StatCardH = 3
	} else {
		l.StatCardH = 4
	}

	remaining := l.ContentH - l.StatCardH
	if l.Mode == LayoutCompact {
		remaining -= l.StatCardH
	}

	l.BarRowH = 5
	remaining -= l.BarRowH

	l.ChartRowH = max(remaining, 0)

	switch l.Mode {
	case LayoutCompact:
		l.ChartLeftW = 0
		l.ChartRightW = 0
		l.ChartRowH = 0
	case LayoutStandard:
		l.ChartLeftW = l.ContentW
		l.ChartRightW = 0
	default:
		l.ChartLeftW = l.ContentW * 60 / 100
		l.ChartRightW = l.ContentW - l.ChartLeftW
	}
}

func (l *Layout) computeSessions() {
	l.FilterBarH = 2
	switch l.Mode {
	case LayoutCompact:
		l.FilterBarH = 1
		l.PreviewH = 0
	case LayoutStandard:
		l.PreviewH = 3
	default:
		l.PreviewH = 4
	}
	l.TableH = max(l.ContentH-l.FilterBarH-l.PreviewH, 3)
}

func (l *Layout) computeAnalytics() {
	l.TopChartH = l.ContentH * 40 / 100
	l.MidChartH = l.ContentH * 30 / 100
	l.BotChartH = l.ContentH - l.TopChartH - l.MidChartH
}

func (l *Layout) computeTools() {
	l.TopChartH = l.ContentH * 55 / 100
	l.BotChartH = l.ContentH - l.TopChartH
}
