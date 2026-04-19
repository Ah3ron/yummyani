package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	inner := m.renderContent()
	w := m.width
	h := m.height
	if w < 40 {
		w = 40
	}
	if h < 15 {
		h = 15
	}
	frameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 1).
		Width(w - 2).
		Height(h - 2)
	return frameStyle.Render(inner)
}

func (m Model) renderContent() string {
	switch m.state {
	case viewSearch:
		return m.viewSearch()
	case viewResults:
		return m.viewResults()
	case viewDubbing:
		return m.viewDubbing()
	case viewEpisodes:
		return m.viewEpisodes()
	case viewQuality:
		return m.viewQuality()
	case viewExtracting:
		return m.viewExtracting()
	case viewPlaying:
		return m.viewPlaying()
	case viewError:
		return m.viewError()
	default:
		return ""
	}
}

func (m Model) viewSearch() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, accentStyle.Render("◉"), headerStyle.Render(" YummyAnime")),
		dimStyle.Render(strings.Repeat("─", 24)),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("enter — поиск  |  ctrl+c — выход"),
	)
}

func (m Model) viewResults() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, accentStyle.Render("◉"), headerStyle.Render(fmt.Sprintf(" Результаты: %s", m.query))),
		dimStyle.Render(strings.Repeat("─", 74)),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — выбрать  |  esc — назад"),
	)
}

func (m Model) viewDubbing() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, accentStyle.Render("◉"), headerStyle.Render(fmt.Sprintf(" %s", m.animeTitle))),
		dimStyle.Render(strings.Repeat("─", 74)),
		sectionStyle.Render(fmt.Sprintf("Озвучки: %d", len(m.groups))),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — выбрать  |  esc — назад"),
	)
}

func (m Model) viewEpisodes() string {
	group := m.selectedGroup()
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, accentStyle.Render("◉"), headerStyle.Render(fmt.Sprintf(" %s", m.animeTitle))),
		dimStyle.Render(strings.Repeat("─", 74)),
		sectionStyle.Render(fmt.Sprintf("[%s]  %d серий", group, len(m.episodes))),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — играть  |  esc — назад"),
	)
}

func (m Model) viewQuality() string {
	ep := epAt(m.episodes, m.epIdx)
	group := m.selectedGroup()
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, accentStyle.Render("◉"), headerStyle.Render(fmt.Sprintf(" %s", m.animeTitle))),
		dimStyle.Render(strings.Repeat("─", 74)),
		qualityBadgeStyle.Render(fmt.Sprintf("Серия %d  [%s]", ep.Number, group)),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — играть  |  esc — назад"),
	)
}

func (m Model) viewExtracting() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				spinnerStyle.Render(m.spinner.View()),
				dimStyle.Render(m.status),
			)),
	)
}

func (m Model) viewPlaying() string {
	ep := epAt(m.episodes, m.epIdx)
	group := m.selectedGroup()
	title := fmt.Sprintf("%s — Серия %d [%s]", m.animeTitle, ep.Number, group)

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, accentStyle.Render("▶"), headerStyle.Render(" Воспроизведение")),
		dimStyle.Render(strings.Repeat("─", 74)),
		normalStyle.Render(title),
		dimStyle.Render("MPV запущен..."),
		helpStyle.Render("esc — назад"),
	)
}

func (m Model) viewError() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginTop(6).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				errorStyle.Render("✗"),
				normalStyle.Render(m.err.Error()),
			)),
	)
}

func (m Model) selectedGroup() string {
	if m.groupIdx >= 0 && m.groupIdx < len(m.groups) {
		return m.groups[m.groupIdx].Name
	}
	return ""
}
