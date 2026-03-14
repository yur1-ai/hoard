package header

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/yur1-ai/hoard/internal/ui/common"
)

// Data holds the values displayed in the header bar.
type Data struct {
	TotalValue float64
	DayChange  float64
	DayPct     float64
	AIProvider string
	AIOnline   bool
}

// Render returns a single-line header bar sized to width.
func Render(d Data, width int) string {
	title := common.HeaderStyle.Render("HOARD")

	var parts []string
	parts = append(parts, title)

	if d.TotalValue > 0 {
		parts = append(parts, fmt.Sprintf("Portfolio: $%s", formatNumber(d.TotalValue)))
	}

	if d.DayChange != 0 {
		changeStr := fmt.Sprintf("Day: %s (%s)",
			formatMoney(d.DayChange), formatPercent(d.DayPct))
		if d.DayChange >= 0 {
			parts = append(parts, common.GainStyle.Render(changeStr))
		} else {
			parts = append(parts, common.LossStyle.Render(changeStr))
		}
	}

	if d.AIProvider != "" {
		dot := "●"
		if !d.AIOnline {
			dot = "○"
		}
		parts = append(parts, fmt.Sprintf("AI: %s %s", d.AIProvider, dot))
	}

	now := time.Now().Format("15:04")
	parts = append(parts, now)

	content := strings.Join(parts, "  ")

	bar := lipgloss.NewStyle().
		Width(width).
		Background(common.ColorHeader).
		Foreground(common.ColorWhite).
		Padding(0, 1)

	return bar.Render(content)
}

func formatNumber(v float64) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("%.2fM", v/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func formatMoney(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+$%.2f", v)
	}
	return fmt.Sprintf("-$%.2f", -v)
}

func formatPercent(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+%.2f%%", v)
	}
	return fmt.Sprintf("%.2f%%", v)
}
