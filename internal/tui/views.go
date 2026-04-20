package tui

import (
	"fmt"

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
	return lipgloss.NewStyle().Width(w - 2).Height(h - 2).Render(inner)
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
		renderHeader("◉", " YummyAnime"),
		sepLine(m.width-4),
		lipgloss.NewStyle().MarginTop(1).Render(m.filter.View()),
		m.filter.ViewItems(),
		helpLine("enter — поиск  |  ctrl+c — выход"),
	)
}

func (m Model) viewResults() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		renderHeader("◉", fmt.Sprintf(" Результаты: %s", m.query)),
		sepLine(m.width-4),
		lipgloss.NewStyle().MarginTop(1).Render(m.filter.View()),
		m.filter.ViewItems(),
		helpLine("↑↓ — навигация  |  enter — выбрать  |  esc — назад"),
	)
}

func (m Model) viewDubbing() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		renderHeader("◉", fmt.Sprintf(" %s", m.animeTitle)),
		sepLine(m.width-4),
		lipgloss.NewStyle().MarginTop(1).Render(sectionStyle.Render(fmt.Sprintf("Озвучки: %d", len(m.groups)))),
		lipgloss.NewStyle().MarginTop(1).Render(m.filter.View()),
		m.filter.ViewItems(),
		helpLine("↑↓ — навигация  |  enter — выбрать  |  esc — назад"),
	)
}

func (m Model) viewEpisodes() string {
	group := m.selectedGroup()
	return lipgloss.JoinVertical(lipgloss.Left,
		renderHeader("◉", fmt.Sprintf(" %s", m.animeTitle)),
		sepLine(m.width-4),
		lipgloss.NewStyle().MarginTop(1).Render(sectionStyle.Render(fmt.Sprintf("[%s]  %d серий", group, len(m.episodes)))),
		lipgloss.NewStyle().MarginTop(1).Render(m.filter.View()),
		m.filter.ViewItems(),
		helpLine("↑↓ — навигация  |  enter — играть  |  esc — назад"),
	)
}

func (m Model) viewQuality() string {
	ep := epAt(m.episodes, m.epIdx)
	group := m.selectedGroup()
	return lipgloss.JoinVertical(lipgloss.Left,
		renderHeader("◉", fmt.Sprintf(" %s", m.animeTitle)),
		sepLine(m.width-4),
		lipgloss.NewStyle().MarginTop(1).Render(qualityBadgeStyle.Render(fmt.Sprintf("Серия %d  [%s]", ep.Number, group))),
		lipgloss.NewStyle().MarginTop(1).Render(m.filter.View()),
		m.filter.ViewItems(),
		helpLine("↑↓ — навигация  |  enter — играть  |  esc — назад"),
	)
}

func (m Model) viewExtracting() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginTop(5).Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				spinnerStyle.Render(m.spinner.View()),
				dimStyle.Render(m.status),
			),
		),
	)
}

func (m Model) viewPlaying() string {
	ep := epAt(m.episodes, m.epIdx)
	group := m.selectedGroup()
	title := fmt.Sprintf("%s — Серия %d [%s]", m.animeTitle, ep.Number, group)

	return lipgloss.JoinVertical(lipgloss.Left,
		renderHeader("▶", " Воспроизведение"),
		sepLine(m.width-4),
		lipgloss.NewStyle().MarginTop(1).Render(normalStyle.Render(title)),
		dimStyle.Render("MPV запущен..."),
		helpLine("esc — назад"),
	)
}

func (m Model) viewError() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginTop(5).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				errorStyle.Render("✗"),
				normalStyle.Render(m.err.Error()),
			),
		),
	)
}

func (m Model) selectedGroup() string {
	if m.groupIdx >= 0 && m.groupIdx < len(m.groups) {
		return m.groups[m.groupIdx].Name
	}
	return ""
}
