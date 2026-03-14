package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/ui/footer"
	"github.com/yur1-ai/hoard/internal/ui/header"
	"github.com/yur1-ai/hoard/internal/ui/help"
)

func (m App) View() tea.View {
	if m.tooSmall {
		msg := fmt.Sprintf(
			"Terminal too small (%d×%d)\nMinimum: %d×%d\nPlease resize.",
			m.width, m.height, minWidth, minHeight,
		)
		content := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
		return tea.View{Content: content, AltScreen: true}
	}

	if m.showHelp {
		return tea.View{Content: help.Render(m.width, m.height), AltScreen: true}
	}

	// Header
	headerBar := header.Render(m.headerData, m.width)

	// Market area (active view)
	var marketContent string
	switch m.activeView {
	case viewPortfolio:
		marketContent = m.portfolio.View()
	case viewWatchlist:
		marketContent = m.watchlist.View()
	case viewCharts:
		marketContent = m.charts.View()
	case viewNews:
		marketContent = m.news.View()
	}

	// Body = market + optional sidebar
	var body string
	if m.sidebarOpen {
		body = lipgloss.JoinHorizontal(lipgloss.Top, marketContent, m.sidebar.View())
	} else {
		body = marketContent
	}

	// Footer
	footerBar := footer.Render(m.lastErr, m.width)

	// Compose full layout
	content := lipgloss.JoinVertical(lipgloss.Left, headerBar, body, footerBar)

	return tea.View{Content: content, AltScreen: true}
}
