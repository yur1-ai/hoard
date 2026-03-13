package app

import (
	"database/sql"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/yurishevtsov/hoard/internal/config"
)

type inputMode int

const (
	modeNormal inputMode = iota
	modeTextInput
	modeSearch
)

type activeView int

const (
	viewPortfolio activeView = iota
	viewWatchlist
	viewCharts
	viewNews
)

type focusArea int

const (
	focusMarket focusArea = iota
	focusSidebar
)

// App is the root Bubble Tea model.
type App struct {
	cfg    config.Config
	db     *sql.DB
	width  int
	height int

	mode        inputMode
	activeView  activeView
	focus       focusArea
	sidebarOpen bool

}

func New(cfg config.Config, db *sql.DB) App {
	sidebarOpen := cfg.SidebarDefault == "open"
	return App{
		cfg:         cfg,
		db:          db,
		mode:        modeNormal,
		activeView:  viewPortfolio,
		focus:       focusMarket,
		sidebarOpen: sidebarOpen,
	}
}

func (m App) Init() tea.Cmd {
	return nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if msg.Code == tea.KeyEscape {
			m.mode = modeNormal
			return m, nil
		}
		if m.mode == modeNormal {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "1":
				m.activeView = viewPortfolio
			case "2":
				m.activeView = viewWatchlist
			case "3":
				m.activeView = viewCharts
			case "4":
				m.activeView = viewNews
			case "tab":
				// Three-state sidebar toggle (v3 fix)
				switch {
				case !m.sidebarOpen:
					m.sidebarOpen = true
					m.focus = focusSidebar
				case m.focus == focusMarket:
					m.focus = focusSidebar
				default:
					m.sidebarOpen = false
					m.focus = focusMarket
				}
			}
		}
		return m, nil
	}
	return m, nil
}

func (m App) View() tea.View {
	viewNames := [4]string{"Portfolio", "Watchlist", "Charts", "News"}
	content := fmt.Sprintf("HOARD  %s", viewNames[m.activeView])
	content += "\n\nPress [1-4] to switch views, [Tab] to toggle sidebar, [q] to quit"

	if m.sidebarOpen {
		focusLabel := "market"
		if m.focus == focusSidebar {
			focusLabel = "sidebar"
		}
		content += fmt.Sprintf("\n\n[Sidebar: open, focus: %s]", focusLabel)
	}

	return tea.View{Content: content, AltScreen: true}
}
