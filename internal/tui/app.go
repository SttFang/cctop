package tui

import (
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fanghanjun/cctop/internal/config"
	"github.com/fanghanjun/cctop/internal/data"
	"github.com/fanghanjun/cctop/internal/metrics"
	"github.com/fanghanjun/cctop/internal/store"
	"github.com/fanghanjun/cctop/internal/tui/common"
)

// AppModel is the root Bubbletea model.
type AppModel struct {
	Config    config.Config
	ActiveTab int
	Theme     common.Theme
	IsDark    bool
	Layout    common.Layout
	ShowHelp  bool

	// Data
	StatsCache *data.StatsCache
	Overview   *metrics.Overview
	Store      *store.Store

	// Sessions tab state
	Sessions     []store.SessionRow
	SessionTotal int
	SessionPage  int
	SessionSort  string
	SessionDesc  bool
	SessionCursor int
	SearchText   string
	Searching    bool
	ShowDetail   bool

	// Tools tab state
	ToolFreqs []store.ToolFreq

	// Analytics state
	AnalyticsPeriod string

	// Sync state
	Indexing bool
	IndexPct float64

	// Timing
	lastRefresh time.Time
}

// Messages
type TickMsg time.Time

type DataLoadedMsg struct {
	Stats *data.StatsCache
	Err   error
}

type SyncDoneMsg struct {
	Result store.SyncResult
	Err    error
}

type SessionsLoadedMsg struct {
	Sessions []store.SessionRow
	Total    int
}

type ToolsLoadedMsg struct {
	ToolFreqs []store.ToolFreq
}

// NewApp creates a new app model.
func NewApp(cfg config.Config) AppModel {
	theme := common.DarkTheme
	isDark := true
	if cfg.Theme == "light" {
		theme = common.LightTheme
		isDark = false
	}

	// Initialize store
	dbPath := filepath.Join(cfg.ClaudeDir, "cctop.db")
	st, _ := store.New(dbPath)

	return AppModel{
		Config:          cfg,
		Theme:           theme,
		IsDark:          isDark,
		Store:           st,
		SessionSort:     "time",
		SessionDesc:     true,
		AnalyticsPeriod: "A",
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		loadData(m.Config),
		syncSessions(m.Store, m.Config),
		tickCmd(m.Config.RefreshInterval()),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.ShowHelp {
			m.ShowHelp = false
			return m, nil
		}

		// Search mode
		if m.Searching {
			return m.handleSearchKey(msg)
		}

		if common.IsQuit(msg) {
			if m.Store != nil {
				m.Store.Close()
			}
			return m, tea.Quit
		}

		// Session detail back
		if m.ShowDetail && common.IsEscape(msg) {
			m.ShowDetail = false
			return m, nil
		}

		if tab := common.TabFromKey(msg); tab >= 0 {
			m.ActiveTab = tab
			m.Layout = common.ComputeLayout(m.Layout.Width, m.Layout.Height, m.ActiveTab)
			return m, m.loadTabData()
		}

		if common.IsTab(msg) {
			m.ActiveTab = (m.ActiveTab + 1) % 4
			m.Layout = common.ComputeLayout(m.Layout.Width, m.Layout.Height, m.ActiveTab)
			return m, m.loadTabData()
		}

		if common.IsShiftTab(msg) {
			m.ActiveTab = (m.ActiveTab + 3) % 4
			m.Layout = common.ComputeLayout(m.Layout.Width, m.Layout.Height, m.ActiveTab)
			return m, m.loadTabData()
		}

		if common.IsThemeToggle(msg) {
			m.IsDark = !m.IsDark
			if m.IsDark {
				m.Theme = common.DarkTheme
			} else {
				m.Theme = common.LightTheme
			}
			return m, nil
		}

		if common.IsRefresh(msg) {
			return m, tea.Batch(loadData(m.Config), syncSessions(m.Store, m.Config))
		}

		if common.IsHelp(msg) {
			m.ShowHelp = true
			return m, nil
		}

		// Tab-specific keys
		return m.handleTabKey(msg)

	case tea.WindowSizeMsg:
		m.Layout = common.ComputeLayout(msg.Width, msg.Height, m.ActiveTab)
		return m, nil

	case DataLoadedMsg:
		if msg.Err == nil && msg.Stats != nil {
			m.StatsCache = msg.Stats
			overview := metrics.ComputeOverview(msg.Stats)
			m.Overview = &overview
		}
		m.lastRefresh = time.Now()
		return m, nil

	case SyncDoneMsg:
		m.Indexing = false
		if msg.Err == nil {
			return m, m.loadTabData()
		}
		return m, nil

	case SessionsLoadedMsg:
		m.Sessions = msg.Sessions
		m.SessionTotal = msg.Total
		return m, nil

	case ToolsLoadedMsg:
		m.ToolFreqs = msg.ToolFreqs
		return m, nil

	case TickMsg:
		return m, tea.Batch(
			loadData(m.Config),
			syncSessions(m.Store, m.Config),
			tickCmd(m.Config.RefreshInterval()),
		)
	}

	return m, nil
}

func (m AppModel) View() string {
	return RenderApp(m)
}

func (m *AppModel) handleTabKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.ActiveTab {
	case 1: // Sessions
		pageSize := max(m.Layout.TableH-3, 10)
		maxPage := max((m.SessionTotal+pageSize-1)/pageSize-1, 0)
		switch {
		case common.IsSearch(msg):
			m.Searching = true
			return m, nil
		case common.IsSort(msg):
			sorts := []string{"time", "cost", "tokens", "duration"}
			for i, s := range sorts {
				if s == m.SessionSort {
					m.SessionSort = sorts[(i+1)%len(sorts)]
					break
				}
			}
			m.SessionPage = 0
			m.SessionCursor = 0
			return m, m.loadSessions()
		case common.IsEnter(msg) && !m.ShowDetail:
			if len(m.Sessions) > 0 && m.SessionCursor < len(m.Sessions) {
				m.ShowDetail = true
			}
			return m, nil
		case msg.String() == "j" || msg.Type == tea.KeyDown:
			if m.SessionCursor < len(m.Sessions)-1 {
				m.SessionCursor++
			} else if m.SessionPage < maxPage {
				// Auto-advance to next page
				m.SessionPage++
				m.SessionCursor = 0
				return m, m.loadSessions()
			}
			return m, nil
		case msg.String() == "k" || msg.Type == tea.KeyUp:
			if m.SessionCursor > 0 {
				m.SessionCursor--
			} else if m.SessionPage > 0 {
				// Auto-go to previous page
				m.SessionPage--
				m.SessionCursor = pageSize - 1
				return m, m.loadSessions()
			}
			return m, nil
		// Page navigation
		case msg.String() == "n" || msg.Type == tea.KeyPgDown || msg.Type == tea.KeyRight:
			if m.SessionPage < maxPage {
				m.SessionPage++
				m.SessionCursor = 0
				return m, m.loadSessions()
			}
			return m, nil
		case msg.String() == "p" || msg.Type == tea.KeyPgUp || msg.Type == tea.KeyLeft:
			if m.SessionPage > 0 {
				m.SessionPage--
				m.SessionCursor = 0
				return m, m.loadSessions()
			}
			return m, nil
		case msg.String() == "g": // go to first page
			m.SessionPage = 0
			m.SessionCursor = 0
			return m, m.loadSessions()
		case msg.String() == "G": // go to last page
			m.SessionPage = maxPage
			m.SessionCursor = 0
			return m, m.loadSessions()
		}
	case 2, 3: // Analytics, Tools
		switch msg.String() {
		case "d", "D":
			m.AnalyticsPeriod = "D"
			return m, m.loadTabData()
		case "w", "W":
			m.AnalyticsPeriod = "W"
			return m, m.loadTabData()
		case "m", "M":
			m.AnalyticsPeriod = "M"
			return m, m.loadTabData()
		case "a", "A":
			m.AnalyticsPeriod = "A"
			return m, m.loadTabData()
		}
	}
	return m, nil
}

func (m *AppModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case common.IsEscape(msg):
		m.Searching = false
		return m, nil
	case common.IsEnter(msg):
		m.Searching = false
		return m, m.loadSessions()
	case msg.Type == tea.KeyBackspace:
		if len(m.SearchText) > 0 {
			m.SearchText = m.SearchText[:len(m.SearchText)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.SearchText += msg.String()
		}
		return m, nil
	}
}

func (m *AppModel) loadTabData() tea.Cmd {
	switch m.ActiveTab {
	case 1:
		return m.loadSessions()
	case 3:
		return m.loadTools()
	default:
		return nil
	}
}

func (m *AppModel) pageSize() int {
	return max(m.Layout.TableH-3, 10)
}

func (m *AppModel) loadSessions() tea.Cmd {
	if m.Store == nil {
		return nil
	}
	st := m.Store
	ps := m.pageSize()
	filter := store.SessionFilter{
		SortBy:   m.SessionSort,
		SortDesc: m.SessionDesc,
		Search:   m.SearchText,
		Limit:    ps,
		Offset:   m.SessionPage * ps,
	}
	return func() tea.Msg {
		sessions, total, _ := st.GetSessions(filter)
		return SessionsLoadedMsg{Sessions: sessions, Total: total}
	}
}

func (m *AppModel) loadTools() tea.Cmd {
	if m.Store == nil {
		return nil
	}
	st := m.Store
	return func() tea.Msg {
		tools, _ := st.GetToolFrequency("", "")
		return ToolsLoadedMsg{ToolFreqs: tools}
	}
}

func loadData(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		stats, err := data.LoadStatsCache(cfg.StatsPath())
		return DataLoadedMsg{Stats: stats, Err: err}
	}
}

func syncSessions(st *store.Store, cfg config.Config) tea.Cmd {
	if st == nil {
		return nil
	}
	projDir := cfg.ProjectsDir()
	return func() tea.Msg {
		if _, err := os.Stat(projDir); err != nil {
			return SyncDoneMsg{Err: err}
		}
		result, err := st.SyncSessions(projDir)
		return SyncDoneMsg{Result: result, Err: err}
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
