package portfolio

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/ui/common"
)

type Model struct {
	width, height int
}

func New() Model { return Model{} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	content := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		Render("Portfolio — positions will appear here")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) SetSize(w, h int) { m.width = w; m.height = h }
