package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewSearch() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🎮 YummyAnime Player"))
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render("Введите название аниме для поиска:"))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter — искать  |  ctrl+c — выход"))

	return boxStyle.Render(b.String())
}

func (m Model) viewResults() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("🔍 Результаты: %s", m.query)))
	b.WriteString("\n")

	for i, r := range m.results {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▶ "
			style = selectedStyle
		}

		statusIcon := statusIcon(r.AnimeStatus.Title)
		year := r.Year.String()
		typ := r.Type.Shortname
		if typ == "" {
			typ = r.Type.Name
		}

		line := fmt.Sprintf("%s%s %s %s  %s  %s",
			cursor,
			style.Render(fmt.Sprintf("%-6d", r.AnimeID)),
			statusIcon,
			style.Render(r.Title),
			dimStyle.Render(year),
			dimStyle.Render(typ),
		)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("↑↓ — навигация  |  enter — выбрать  |  q — назад"))

	return boxStyle.Render(b.String())
}

func (m Model) viewDubbing() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("🎙 %s — Озвучки", m.animeTitle)))
	b.WriteString("\n")

	for i, g := range m.groups {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▶ "
			style = selectedStyle
		}

		count := len(g.Episodes)
		line := fmt.Sprintf("%s%s  %s",
			cursor,
			style.Render(g.Name),
			dimStyle.Render(fmt.Sprintf("(%d эп.)", count)),
		)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("↑↓ — навигация  |  enter — выбрать  |  q — назад"))

	return boxStyle.Render(b.String())
}

func (m Model) viewEpisodes() string {
	var b strings.Builder

	groupName := ""
	if m.groupIdx < len(m.groups) {
		groupName = m.groups[m.groupIdx].Name
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("📽 %s [%s]", m.animeTitle, groupName)))
	b.WriteString("\n")

	for i, ep := range m.episodes {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▶ "
			style = selectedStyle
		}

		dur := ""
		if ep.Duration > 0 {
			mins := ep.Duration / 60
			secs := ep.Duration % 60
			dur = dimStyle.Render(fmt.Sprintf("  (%d:%02d)", mins, secs))
		}

		line := fmt.Sprintf("%s%sСерия %d%s",
			cursor,
			style.Render(fmt.Sprintf("%-5d", ep.Number)),
			ep.Number,
			dur,
		)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("↑↓ — навигация  |  enter — играть  |  b/q — назад"))

	return boxStyle.Render(b.String())
}

func (m Model) viewQuality() string {
	var b strings.Builder

	ep := m.episodes[m.epIdx]
	groupName := ""
	if m.groupIdx < len(m.groups) {
		groupName = m.groups[m.groupIdx].Name
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("🎬 %s — Серия %d [%s]", m.animeTitle, ep.Number, groupName)))
	b.WriteString("\n\n")
	b.WriteString(sectionStyle.Render("Выберите качество видео:"))
	b.WriteString("\n")

	for i, q := range m.availableQualities {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "▶ "
			style = selectedStyle
		}

		qualityBadge := qualityBadgeStyle.Render(fmt.Sprintf(" %-5s", q.Label))
		codec := linkType(q.Links[0].Type)

		line := fmt.Sprintf("%s%s %s",
			cursor,
			qualityBadge,
			style.Render(codec),
		)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("↑↓ — навигация  |  enter — играть  |  q — назад к сериям"))

	return boxStyle.Render(b.String())
}

func (m Model) viewExtracting() string {
	return fmt.Sprintf("\n  %s %s\n", m.spinner.View(), statusStyle.Render(m.status))
}

func (m Model) viewPlaying() string {
	var b strings.Builder
	ep := m.episodes[m.epIdx]
	groupName := ""
	if m.groupIdx < len(m.groups) {
		groupName = m.groups[m.groupIdx].Name
	}
	title := fmt.Sprintf("%s — Серия %d [%s]", m.animeTitle, ep.Number, groupName)

	b.WriteString(titleStyle.Render("▶ Воспроизведение"))
	b.WriteString("\n")
	b.WriteString(normalStyle.Render(title))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render(m.status))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("q — выход"))

	return boxStyle.Render(b.String())
}

func (m Model) viewError() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render("✗ Ошибка"))
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render(m.err.Error()))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter — назад к поиску  |  q — выход"))

	return boxStyle.Render(b.String())
}

// statusIcon returns an emoji for the anime airing status.
func statusIcon(status string) string {
	s := strings.ToLower(status)
	switch {
	case strings.Contains(s, "завершён"), strings.Contains(s, "вышел"):
		return "✅"
	case strings.Contains(s, "онгоинг"), strings.Contains(s, "выходит"):
		return "🔄"
	case strings.Contains(s, "анонс"):
		return "📢"
	default:
		return "?"
	}
}
