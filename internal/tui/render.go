package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fanghanjun/cctop/internal/tui/components"
	"github.com/fanghanjun/cctop/internal/tui/views"
)

// RenderApp produces the full screen output.
func RenderApp(m AppModel) string {
	if m.Layout.Width == 0 {
		return "Initializing..."
	}

	if m.Layout.Width < 60 || m.Layout.Height < 16 {
		return "Terminal too small. Minimum: 60x16\nPress q to quit."
	}

	if m.ShowHelp {
		return renderHelp(m)
	}

	var parts []string

	header := components.RenderHeader(m.ActiveTab, m.Layout.Width, m.Config.RefreshSecs, m.Theme)
	parts = append(parts, header)

	content := renderContent(m)
	parts = append(parts, content)

	var extraHints string
	switch m.ActiveTab {
	case 1:
		ks := lipgloss.NewStyle().Foreground(m.Theme.Yellow)
		ds := lipgloss.NewStyle().Foreground(m.Theme.FgDim)
		extraHints = ks.Render("/") + ds.Render(":Search") + "  " +
			ks.Render("s") + ds.Render(":Sort") + "  " +
			ks.Render("n/p") + ds.Render(":Page") + "  " +
			ks.Render("Enter") + ds.Render(":Detail")
	case 2, 3:
		extraHints = lipgloss.NewStyle().Foreground(m.Theme.Yellow).Render("D/W/M/A") +
			lipgloss.NewStyle().Foreground(m.Theme.FgDim).Render(":Period")
	}
	statusBar := components.RenderStatusBar(m.ActiveTab, m.Layout.Width, extraHints, m.Theme)
	parts = append(parts, statusBar)

	return strings.Join(parts, "\n")
}

func renderContent(m AppModel) string {
	switch m.ActiveTab {
	case 0:
		v := &views.OverviewView{
			Data:   m.Overview,
			Layout: m.Layout,
			Theme:  m.Theme,
		}
		return v.View()
	case 1:
		pageSize := max(m.Layout.TableH-3, 10)
		v := &views.SessionsView{
			Layout:    m.Layout,
			Theme:     m.Theme,
			Sessions:  m.Sessions,
			Total:     m.SessionTotal,
			Cursor:    m.SessionCursor,
			Page:      m.SessionPage,
			PageSize:  pageSize,
			SortBy:    m.SessionSort,
			SortDesc:  m.SessionDesc,
			Search:    m.SearchText,
			Searching: m.Searching,
			Indexing:  m.Indexing,
		}
		if m.ShowDetail && m.SessionCursor < len(m.Sessions) {
			sess := m.Sessions[m.SessionCursor]
			var toolCalls map[string]int
			if m.Store != nil {
				toolCalls, _ = m.Store.GetSessionToolCalls(sess.SessionID)
			}
			v.ShowDetail = true
			v.DetailData = &views.SessionDetailData{
				Session:   sess,
				ToolCalls: toolCalls,
			}
		}
		return v.View()
	case 2:
		v := &views.AnalyticsView{
			Layout: m.Layout,
			Theme:  m.Theme,
			Stats:  m.StatsCache,
			Period: m.AnalyticsPeriod,
		}
		return v.View()
	case 3:
		v := &views.ToolsView{
			Layout:    m.Layout,
			Theme:     m.Theme,
			ToolFreqs: m.ToolFreqs,
			Period:    m.AnalyticsPeriod,
		}
		return v.View()
	default:
		return ""
	}
}

func renderHelp(m AppModel) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(m.Theme.Primary).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(m.Theme.Yellow).
		Width(15)

	descStyle := lipgloss.NewStyle().
		Foreground(m.Theme.Fg)

	dimStyle := lipgloss.NewStyle().
		Foreground(m.Theme.FgDim)

	lines := []string{
		titleStyle.Render("CCtop - Claude Code Monitor"),
		"",
		dimStyle.Render("Navigation"),
		keyStyle.Render("  1/2/3/4") + descStyle.Render("Switch tabs"),
		keyStyle.Render("  Tab") + descStyle.Render("Next tab"),
		keyStyle.Render("  Shift-Tab") + descStyle.Render("Previous tab"),
		keyStyle.Render("  j/k ↑/↓") + descStyle.Render("Navigate up/down"),
		keyStyle.Render("  Enter") + descStyle.Render("Open detail"),
		keyStyle.Render("  Esc") + descStyle.Render("Go back"),
		"",
		dimStyle.Render("Actions"),
		keyStyle.Render("  r") + descStyle.Render("Refresh data"),
		keyStyle.Render("  /") + descStyle.Render("Search/filter"),
		keyStyle.Render("  s") + descStyle.Render("Cycle sort"),
		keyStyle.Render("  t") + descStyle.Render("Toggle theme"),
		keyStyle.Render("  D/W/M/A") + descStyle.Render("Time period"),
		keyStyle.Render("  ?") + descStyle.Render("This help"),
		keyStyle.Render("  q") + descStyle.Render("Quit"),
		"",
		dimStyle.Render("Press any key to close help"),
	}

	content := strings.Join(lines, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.Theme.Primary).
		Padding(1, 2).
		Width(50)

	box := boxStyle.Render(content)

	return lipgloss.Place(m.Layout.Width, m.Layout.Height,
		lipgloss.Center, lipgloss.Center, box)
}
