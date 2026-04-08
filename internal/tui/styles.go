// Package tui implements the BubbleTea TUI for the YummyAnime player.
//
// It provides a multi-screen interface: search → results → dubbing →
// episodes → quality → playback, with fuzzy filtering on each screen.
package tui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	colorAccent    = lipgloss.Color("#6C9CFA")
	colorSuccess   = lipgloss.Color("#6BCB77")
	colorWarning   = lipgloss.Color("#FFD93D")
	colorDim       = lipgloss.Color("#888888")
	colorBright    = lipgloss.Color("#FFFFFF")
	colorMuted     = lipgloss.Color("#AAAAAA")
	colorError     = lipgloss.Color("#FF6B6B")
	colorSection   = lipgloss.Color("#BB86FC")
	colorQuality   = lipgloss.Color("#03DAC6")
	colorQualityBg = lipgloss.Color("#1a1a2e")
	colorFilterBg  = lipgloss.Color("#1e1e2e")
)

// Pre-built lipgloss styles.
var (
	pageStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBright).
			MarginBottom(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	accentStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSection).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	qualityBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorQuality).
				Background(colorQualityBg).
				Padding(0, 1)

	listStyle = lipgloss.NewStyle().
			MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	filterStyle = lipgloss.NewStyle().
			Background(colorFilterBg)
)
