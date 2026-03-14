package sidebar

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/ui/common"
)

type panel int

const (
	panelCalendar panel = iota
	panelTasks
	panelStandup
	panelCount // sentinel for wrapping
)

// Model manages the sidebar's three stacked panels.
type Model struct {
	Width       int
	Height      int
	Focused     bool
	ActivePanel panel
}

func New() Model {
	return Model{
		Width:       common.SidebarWidth,
		ActivePanel: panelCalendar,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Focused {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "j", "down":
			m.ActivePanel = (m.ActivePanel + 1) % panelCount
		case "k", "up":
			m.ActivePanel = (m.ActivePanel - 1 + panelCount) % panelCount
		}
	}
	return m, nil
}

func (m Model) View() string {
	panelHeight := (m.Height - 2) / 3 // divide among 3 panels, minus borders
	if panelHeight < 3 {
		panelHeight = 3
	}

	panels := []struct {
		title   string
		content string
		idx     panel
	}{
		{"Calendar", "No events today", panelCalendar},
		{"Tasks", "No tasks", panelTasks},
		{"Standup", "Not filled", panelStandup},
	}

	var rendered []string
	for _, p := range panels {
		style := common.UnfocusedBorder
		if m.Focused && m.ActivePanel == p.idx {
			style = common.FocusedBorder
		}

		content := fmt.Sprintf(" %s\n %s",
			lipgloss.NewStyle().Bold(true).Render(p.title),
			common.MutedStyle.Render(p.content),
		)

		box := style.
			Width(m.Width - 2). // account for border
			Height(panelHeight).
			Render(content)

		rendered = append(rendered, box)
	}

	return strings.Join(rendered, "\n")
}
