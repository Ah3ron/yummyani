package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func sepLine(w int) string {
	if w < 10 {
		w = 10
	}
	return dimStyle.Render(strings.Repeat("─", w))
}

func renderHeader(icon, title string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left, accentStyle.Render(icon), headerStyle.Render(title))
}

func helpLine(txt string) string {
	return helpStyle.Render(txt)
}
