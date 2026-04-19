package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent   = lipgloss.Color("#7aa2f7")
	colorError    = lipgloss.Color("#f7768e")
	colorDim      = lipgloss.Color("#565f89")
	colorMuted    = lipgloss.Color("#a9b1d6")
	colorBright   = lipgloss.Color("#c0caf5")
	colorBorder   = lipgloss.Color("#3b4261")
	colorPurple   = lipgloss.Color("#9d7cd8")
	colorCyan     = lipgloss.Color("#7dcfff")
	colorFilterBg = lipgloss.Color("#16161e")
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

	filterStyle = lipgloss.NewStyle().
			Background(colorFilterBg)

	qualityBadgeStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)
)
