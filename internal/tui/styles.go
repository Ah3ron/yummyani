package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent     = lipgloss.Color("#7FD4FF")
	colorBright     = lipgloss.Color("#D9E0EE")
	colorDim        = lipgloss.Color("#9AA5B1")
	colorMuted      = lipgloss.Color("#6E7A86")
	colorPurple     = lipgloss.Color("#A67EE9")
	colorCyan       = lipgloss.Color("#66E1FF")
	colorError      = lipgloss.Color("#FF6B6B")
	colorFilterBg   = lipgloss.Color("#0B1014")
	colorFilterFg   = lipgloss.Color("#C8D0DA")
	colorBackground = lipgloss.Color("#07090B")
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	normalStyle = lipgloss.NewStyle().
			Foreground(colorBright)

	accentStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	sectionStyle = lipgloss.NewStyle().
			Foreground(colorPurple).
			Bold(true)

	itemSelStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	itemNormStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Background(colorFilterBg).
			Foreground(colorFilterFg)

	qualityBadgeStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)
)
