package common

import "charm.land/lipgloss/v2"

// Color palette.
var (
	ColorGreen     = lipgloss.Color("#04B575")
	ColorRed       = lipgloss.Color("#FF4672")
	ColorMuted     = lipgloss.Color("#626262")
	ColorHighlight = lipgloss.Color("#874BFD")
	ColorWhite     = lipgloss.Color("#FAFAFA")
	ColorBorder    = lipgloss.Color("#3C3C3C")
	ColorHeader    = lipgloss.Color("#874BFD")
)

// Layout constants.
const (
	SidebarWidth = 28
	HeaderHeight = 1
	FooterHeight = 1
)

// Reusable styles.
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorHeader).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	FooterKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	FocusedBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorHighlight)

	UnfocusedBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	GainStyle  = lipgloss.NewStyle().Foreground(ColorGreen)
	LossStyle  = lipgloss.NewStyle().Foreground(ColorRed)
	MutedStyle = lipgloss.NewStyle().Foreground(ColorMuted)
)
