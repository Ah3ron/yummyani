package tui

import "github.com/charmbracelet/lipgloss"

// Colors.
var (
	accent  = lipgloss.Color("#6C9CFA") // blue
	success = lipgloss.Color("#6BCB77") // green
	warning = lipgloss.Color("#FFD93D") // yellow
	dim     = lipgloss.Color("#888888")
	bright  = lipgloss.Color("#FFFFFF")
	muted   = lipgloss.Color("#AAAAAA")
)

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(bright)

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(success)

	selectedStyle = lipgloss.NewStyle().
			Foreground(bright)

	normalStyle = lipgloss.NewStyle().
			Foreground(muted)

	dimStyle = lipgloss.NewStyle().
			Foreground(dim)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(dim).
			MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(accent)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#BB86FC")).
			MarginBottom(0)

	qualityBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#03DAC6")).
				Background(lipgloss.Color("#1a1a2e")).
				Padding(0, 1)
)
