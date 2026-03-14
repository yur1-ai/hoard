package help

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/ui/common"
)

type binding struct {
	key  string
	desc string
}

var globalBindings = []binding{
	{"1-4", "Switch market views"},
	{"Tab", "Toggle sidebar"},
	{"/", "Search symbols"},
	{"?", "Toggle this help"},
	{"q", "Quit"},
}

var portfolioBindings = []binding{
	{"a", "Add position"},
	{"e", "Edit position"},
	{"d", "Delete position"},
	{"f", "Filter by account"},
	{"r", "Refresh prices"},
	{"j/k", "Navigate up/down"},
	{"s", "Sort column"},
}

var sidebarBindings = []binding{
	{"a", "Add task / edit standup"},
	{"x", "Toggle task done"},
	{"d", "Delete task"},
	{"e", "Edit standup"},
	{"j/k", "Switch panels"},
}

// Render returns the help overlay content sized to width×height.
func Render(width, height int) string {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorHighlight).
		Width(8)

	descStyle := lipgloss.NewStyle().
		Foreground(common.ColorWhite)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorWhite).
		MarginBottom(1)

	var b strings.Builder
	b.WriteString(titleStyle.Render("  KEYBOARD SHORTCUTS"))
	b.WriteString("\n\n")

	writeSection := func(title string, bindings []binding) {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(common.ColorMuted).Render("  "+title))
		b.WriteString("\n")
		for _, bind := range bindings {
			fmt.Fprintf(&b, "    %s %s\n",
				keyStyle.Render(bind.key),
				descStyle.Render(bind.desc))
		}
		b.WriteString("\n")
	}

	writeSection("Global", globalBindings)
	writeSection("Portfolio", portfolioBindings)
	writeSection("Sidebar", sidebarBindings)

	b.WriteString(common.MutedStyle.Render("  Press ? or Esc to close"))

	overlay := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorHighlight).
		Padding(1, 2).
		Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay)
}
