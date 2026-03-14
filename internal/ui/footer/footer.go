package footer

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/ui/common"
)

// Render returns a single-line footer with key hints and optional error.
func Render(lastErr string, width int) string {
	hints := []struct{ key, label string }{
		{"1", "Portfolio"},
		{"2", "Watchlist"},
		{"3", "Charts"},
		{"4", "News"},
		{"Tab", "Sidebar"},
		{"a", "add"},
		{"?", "Help"},
		{"q", "Quit"},
	}

	var content string
	for i, h := range hints {
		if i > 0 {
			content += "  "
		}
		content += fmt.Sprintf("[%s]%s", common.FooterKeyStyle.Render(h.key), h.label)
	}

	if lastErr != "" {
		errMsg := common.ErrorStyle.Render(fmt.Sprintf(" [!] %s", lastErr))
		content = errMsg + "  " + content
	}

	bar := lipgloss.NewStyle().
		Width(width).
		Foreground(common.ColorMuted).
		Padding(0, 1)

	return bar.Render(content)
}
