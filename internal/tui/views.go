package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current view based on the model's state.
func (m Model) View() string {
	switch m.state {
	case viewSearch:
		return pageStyle.Render(m.viewSearch())
	case viewResults:
		return pageStyle.Render(m.viewResults())
	case viewDubbing:
		return pageStyle.Render(m.viewDubbing())
	case viewEpisodes:
		return pageStyle.Render(m.viewEpisodes())
	case viewQuality:
		return pageStyle.Render(m.viewQuality())
	case viewExtracting:
		return pageStyle.Render(m.viewExtracting())
	case viewPlaying:
		return pageStyle.Render(m.viewPlaying())
	case viewError:
		return pageStyle.Render(m.viewError())
	default:
		return ""
	}
}

func (m Model) viewSearch() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("🎮 YummyAnime Player"),
		subtitleStyle.Render("Введите название аниме для поиска:"),
		m.filter.View(),
		helpStyle.Render("enter — искать  |  ctrl+c — выход"),
	)
}

func (m Model) viewResults() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(fmt.Sprintf("🔍 Результаты: %s", m.query)),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — выбрать  |  esc — назад"),
	)
}

func (m Model) viewDubbing() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(fmt.Sprintf("🎙 %s — Озвучки", m.animeTitle)),
		sectionStyle.Render(fmt.Sprintf("Найдено озвучек: %d", len(m.groups))),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — выбрать  |  esc — назад"),
	)
}

func (m Model) viewEpisodes() string {
	group := m.selectedGroup()
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(fmt.Sprintf("📽 %s [%s]", m.animeTitle, group)),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — играть  |  esc — назад"),
	)
}

func (m Model) viewQuality() string {
	ep := epAt(m.episodes, m.epIdx)
	group := m.selectedGroup()
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(fmt.Sprintf("🎬 %s — Серия %d [%s]", m.animeTitle, ep.Number, group)),
		qualityBadgeStyle.Render("Выберите качество видео:"),
		m.filter.View(),
		m.filter.ViewItems(),
		helpStyle.Render("↑↓ — навигация  |  enter — играть  |  esc — назад"),
	)
}

func (m Model) viewExtracting() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("⏳ Загрузка"),
		lipgloss.NewStyle().
			MarginTop(1).
			Render(fmt.Sprintf("  %s %s", m.spinner.View(), warningStyle.Render(m.status))),
	)
}

func (m Model) viewPlaying() string {
	ep := epAt(m.episodes, m.epIdx)
	group := m.selectedGroup()
	title := fmt.Sprintf("%s — Серия %d [%s]", m.animeTitle, ep.Number, group)

	return lipgloss.JoinVertical(lipgloss.Left,
		successStyle.Render("▶ Воспроизведение"),
		normalStyle.Render(title),
		dimStyle.Render(m.status),
		helpStyle.Render("esc — назад"),
	)
}

func (m Model) viewError() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		errorStyle.Render("✗ Ошибка"),
		normalStyle.Render(m.err.Error()),
		helpStyle.Render("enter — назад к поиску  |  esc — выход"),
	)
}

// selectedGroup returns the name of the currently selected dubbing group.
func (m Model) selectedGroup() string {
	if m.groupIdx >= 0 && m.groupIdx < len(m.groups) {
		return m.groups[m.groupIdx].Name
	}
	return ""
}
