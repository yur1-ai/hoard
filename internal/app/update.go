package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
)

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = msg.Width < minWidth || msg.Height < minHeight
		m.propagateSize()
		return m, nil

	case ErrMsg:
		m.lastErr = msg.Error()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Escape always resets to normal mode and closes overlays
	if key.Matches(msg, m.keys.Escape) {
		m.mode = modeNormal
		m.showHelp = false
		m.lastErr = ""
		return m, nil
	}

	// Help overlay: only ? and Esc work while showing
	if m.showHelp {
		if key.Matches(msg, m.keys.Help) {
			m.showHelp = false
		}
		return m, nil
	}

	// In text input or search mode, don't process global keys
	if m.mode != modeNormal {
		return m, nil
	}

	// Global keys (normal mode)
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp

	case key.Matches(msg, m.keys.View1):
		m.activeView = viewPortfolio
	case key.Matches(msg, m.keys.View2):
		m.activeView = viewWatchlist
	case key.Matches(msg, m.keys.View3):
		m.activeView = viewCharts
	case key.Matches(msg, m.keys.View4):
		m.activeView = viewNews

	case key.Matches(msg, m.keys.Tab):
		m.toggleSidebar()

	default:
		// Route to focused area
		if m.focus == focusSidebar {
			var cmd tea.Cmd
			m.sidebar, cmd = m.sidebar.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *App) toggleSidebar() {
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
	m.sidebar.Focused = m.focus == focusSidebar
	m.propagateSize()
}

func (m *App) propagateSize() {
	marketWidth := m.width
	contentHeight := m.height - 2 // header + footer

	if m.sidebarOpen {
		marketWidth -= m.sidebar.Width
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	m.sidebar.Height = contentHeight

	m.portfolio.SetSize(marketWidth, contentHeight)
	m.watchlist.SetSize(marketWidth, contentHeight)
	m.charts.SetSize(marketWidth, contentHeight)
	m.news.SetSize(marketWidth, contentHeight)
}
