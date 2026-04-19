package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"yummyani/internal/api"
	"yummyani/internal/kodik"
	"yummyani/internal/player"
)

// handleKey routes key events to the appropriate per-state handler.
func (m Model) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case viewSearch:
		return m.handleSearchKey(key)
	case viewResults:
		return m.handleResultsKey(key)
	case viewDubbing:
		return m.handleDubbingKey(key)
	case viewEpisodes:
		return m.handleEpisodesKey(key)
	case viewQuality:
		return m.handleQualityKey(key)
	case viewPlaying:
		return m.handlePlayingKey(key)
	case viewError:
		return m.handleErrorKey(key)
	default:
		return m, nil
	}
}

// handleSearchKey processes key events in the search view.
func (m Model) handleSearchKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(key)

	if key.String() == "enter" {
		if q := m.filter.Value(); q != "" {
			m.query = q
			m.state = viewExtracting
			m.status = "Поиск..."
			return m, tea.Batch(cmd, m.doSearch(m.ctx, q))
		}
	}
	return m, cmd
}

// handleResultsKey processes key events in the results view.
func (m Model) handleResultsKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(key)

	switch key.String() {
	case "enter":
		if sel := m.filter.Selected(); sel != nil {
			m.animeID = sel.ID
			m.state = viewExtracting
			m.status = "Загрузка видео..."
			return m, tea.Batch(cmd, m.doFetchVideos(m.ctx, m.animeID))
		}
	case "esc":
		m.transitionToSearch()
		return m, m.filter.Init()
	}
	return m, cmd
}

func (m Model) handleDubbingKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(key)

	switch key.String() {
	case "enter":
		if sel := m.filter.Selected(); sel != nil {
			idx := sel.ID
			if idx < 0 || idx >= len(m.groups) {
				return m, nil
			}
			m.groupIdx = idx
			m.episodes = m.groups[idx].Episodes
			sortEpisodes(m.episodes)
			m.filter.Reset(episodesToItems(m.episodes))
			m.filter.SetPlaceholder("Фильтр...")
			m.filter.SetPrompt("📽 ")
			m.state = viewEpisodes

			if len(m.episodes) > 0 {
				urls := make([]string, 0, 3)
				for i := 0; i < len(m.episodes) && i < 3; i++ {
					urls = append(urls, kodik.EnsureHTTPS(m.episodes[i].IframeURL))
				}
				go m.linkExtractor.Prefetch(m.ctx, urls)
			}
			return m, cmd
		}
	case "esc":
		m.filter.Reset(searchToItems(m.results))
		m.filter.SetPrompt("🔎 ")
		m.state = viewResults
		return m, cmd
	}
	return m, cmd
}

// handleEpisodesKey processes key events in the episodes view.
func (m Model) handleEpisodesKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(key)

	switch key.String() {
	case "enter":
		if sel := m.filter.Selected(); sel != nil {
			idx := sel.ID
			if idx < 0 || idx >= len(m.episodes) {
				return m, nil
			}
			m.epIdx = idx
			ep := m.episodes[idx]
			iframeURL := kodik.EnsureHTTPS(ep.IframeURL)
			m.state = viewExtracting
			m.status = fmt.Sprintf("Извлечение ссылки (Серия %d)...", ep.Number)
			return m, tea.Batch(cmd, m.doExtract(m.ctx, iframeURL))
		}
	case "esc":
		m.filter.Reset(dubbingToItems(m.groups))
		m.filter.SetPrompt("🎙 ")
		m.state = viewDubbing
		return m, cmd
	}
	return m, cmd
}

// handleQualityKey processes key events in the quality selection view.
func (m Model) handleQualityKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(key)

	switch key.String() {
	case "enter":
		if sel := m.filter.Selected(); sel != nil {
			idx := sel.ID
			if idx < 0 || idx >= len(m.availableQualities) {
				return m, nil
			}
			q := m.availableQualities[idx]
			m.selectedQuality = q
			m.state = viewPlaying
			m.status = fmt.Sprintf("%s [%s]", q.Label, player.LinkType(q.Links[0].Type))
			return m, tea.Batch(cmd, m.startPlay(
				m.playTitle(epAt(m.episodes, m.epIdx)),
				q.Links[0].Src,
				q.Label,
			))
		}
	case "esc":
		m.filter.Reset(episodesToItems(m.episodes))
		m.filter.SetPrompt("📽 ")
		m.state = viewEpisodes
		return m, cmd
	}
	return m, cmd
}

// handlePlayingKey processes key events during playback.
func (m Model) handlePlayingKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "esc" {
		m.filter.Reset(episodesToItems(m.episodes))
		m.filter.SetPrompt("📽 ")
		m.state = viewEpisodes
	}
	return m, nil
}

// handleErrorKey processes key events in the error view.
func (m Model) handleErrorKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		m.err = nil
		m.transitionToSearch()
		return m, m.filter.Init()
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

// --- Async result handlers ---

func (m Model) handleSearchDone(msg searchDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = viewError
		return m, nil
	}
	m.results = msg.results
	if len(m.results) == 0 {
		m.err = fmt.Errorf("ничего не найдено")
		m.state = viewError
		return m, nil
	}
	m.filter.Reset(searchToItems(m.results))
	m.filter.SetPlaceholder("Фильтр...")
	m.filter.SetPrompt("🔎 ")
	m.state = viewResults
	return m, nil
}

func (m Model) handleVideosDone(msg videosDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = viewError
		return m, nil
	}
	m.animeTitle = msg.title
	m.groups = api.GroupByDubbing(msg.videos)
	if len(m.groups) == 0 {
		m.err = fmt.Errorf("нет Kodik озвучек")
		m.state = viewError
		return m, nil
	}
	m.filter.Reset(dubbingToItems(m.groups))
	m.filter.SetPlaceholder("Фильтр...")
	m.filter.SetPrompt("🎙 ")
	m.state = viewDubbing
	return m, nil
}

func (m Model) handleExtractDone(msg extractDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = viewError
		return m, nil
	}
	m.kodikResp = msg.resp
	m.availableQualities = kodik.AvailableQualities(msg.resp)
	if len(m.availableQualities) == 0 {
		m.err = fmt.Errorf("нет доступных качеств видео")
		m.state = viewError
		return m, nil
	}
	// If only one quality, skip selection and play immediately.
	if len(m.availableQualities) == 1 {
		q := m.availableQualities[0]
		m.selectedQuality = q
		m.state = viewPlaying
		m.status = fmt.Sprintf("%s [%s]", q.Label, player.LinkType(q.Links[0].Type))
		return m, m.startPlay(
			m.playTitle(epAt(m.episodes, m.epIdx)),
			q.Links[0].Src,
			q.Label,
		)
	}
	m.filter.Reset(qualityToItems(m.availableQualities))
	m.filter.SetPlaceholder("Фильтр...")
	m.filter.SetPrompt("🎬 ")
	m.state = viewQuality
	return m, nil
}

func (m Model) handlePlayDone(msg playDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = viewError
		return m, nil
	}
	m.state = viewEpisodes
	m.status = ""
	m.kodikResp = nil
	m.availableQualities = nil
	return m, nil
}

// --- Transition helpers ---

func (m *Model) transitionToSearch() {
	m.filter.Reset(nil)
	m.filter.SetPrompt("🔍 ")
	m.filter.SetPlaceholder("Название аниме...")
	m.state = viewSearch
}

func (m Model) playTitle(ep api.VideoEntry) string {
	group := ""
	if m.groupIdx >= 0 && m.groupIdx < len(m.groups) {
		group = m.groups[m.groupIdx].Name
	}
	return fmt.Sprintf("%s — Серия %d [%s]", m.animeTitle, ep.Number, group)
}

func epAt(episodes []api.VideoEntry, idx int) api.VideoEntry {
	if idx >= 0 && idx < len(episodes) {
		return episodes[idx]
	}
	return api.VideoEntry{}
}
