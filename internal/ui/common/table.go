package common

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

// NewStyledTable returns a table.Model with Hoard's default styling.
func NewStyledTable(columns []table.Column, rows []table.Row, height int) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(height),
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorMuted).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(ColorWhite).
		Background(ColorHighlight)
	t.SetStyles(s)
	return t
}

// FormatMoney returns a colored money string.
func FormatMoney(value float64) string {
	s := fmt.Sprintf("$%.2f", value)
	if value < 0 {
		s = fmt.Sprintf("-$%.2f", -value)
		return LossStyle.Render(s)
	}
	return s
}

// FormatPercent returns a colored percentage string.
func FormatPercent(value float64) string {
	s := fmt.Sprintf("%.2f%%", value)
	if value > 0 {
		return GainStyle.Render("+" + s)
	}
	if value < 0 {
		return LossStyle.Render(s)
	}
	return s
}

// FormatChange returns "▲ +2.3%" or "▼ -1.2%" with appropriate color.
func FormatChange(value float64) string {
	if value > 0 {
		return GainStyle.Render(fmt.Sprintf("▲ +%.2f%%", value))
	}
	if value < 0 {
		return LossStyle.Render(fmt.Sprintf("▼ %.2f%%", value))
	}
	return MutedStyle.Render("● 0.00%")
}
