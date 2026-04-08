package tui

import (
	"context"
	"fmt"
	"sort"

	"yummyani/internal/api"
	"yummyani/internal/kodik"
	"yummyani/internal/player"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Async tea messages ---

type searchDoneMsg struct {
	results []api.SearchResult
	err     error
}

type videosDoneMsg struct {
	videos []api.VideoEntry
	title  string
	err    error
}

type extractDoneMsg struct {
	resp *kodik.Response
	err  error
}

type playDoneMsg struct{ err error }

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// --- Async commands ---

func (m Model) doSearch(ctx context.Context, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.anime.Search(ctx, query, m.searchLimit)
		return searchDoneMsg{results: results, err: err}
	}
}

func (m Model) doFetchVideos(ctx context.Context, animeID int) tea.Cmd {
	return func() tea.Msg {
		videos, err := m.anime.GetVideos(ctx, animeID)
		if err != nil {
			return videosDoneMsg{err: err}
		}
		info, err := m.anime.GetAnime(ctx, animeID)
		title := fmt.Sprintf("Anime #%d", animeID)
		if err == nil && info != nil {
			title = info.DisplayName()
		}
		return videosDoneMsg{videos: videos, title: title}
	}
}

func (m Model) doExtract(ctx context.Context, iframeURL string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.linkExtractor.Parse(ctx, iframeURL)
		if err != nil {
			return extractDoneMsg{err: fmt.Errorf("kodik parse: %w", err)}
		}
		if kodik.BestLink(resp) == "" {
			return extractDoneMsg{err: fmt.Errorf("no playable links found")}
		}
		return extractDoneMsg{resp: resp}
	}
}

func (m Model) startPlay(title, url, _ string) tea.Cmd {
	return func() tea.Msg {
		if err := m.player.Play(title, url); err != nil {
			return playDoneMsg{err: err}
		}
		return playDoneMsg{}
	}
}

// --- Conversion helpers ---

func searchToItems(results []api.SearchResult) []FilterItem {
	items := make([]FilterItem, len(results))
	for i, r := range results {
		items[i] = FilterItem{
			ID:    r.AnimeID,
			Label: r.Title,
			Sub:   fmt.Sprintf("%s  %s", r.Year.String(), r.Type.DisplayName()),
		}
	}
	return items
}

func dubbingToItems(groups []api.DubbingGroup) []FilterItem {
	items := make([]FilterItem, len(groups))
	for i, g := range groups {
		items[i] = FilterItem{
			ID:    i,
			Label: g.Name,
			Sub:   fmt.Sprintf("%d эп.", len(g.Episodes)),
		}
	}
	return items
}

func episodesToItems(episodes []api.VideoEntry) []FilterItem {
	items := make([]FilterItem, len(episodes))
	for i, ep := range episodes {
		dur := ""
		if ep.Duration > 0 {
			dur = player.FormatDuration(ep.Duration)
		}
		items[i] = FilterItem{
			ID:    i,
			Label: fmt.Sprintf("Серия %d", ep.Number),
			Sub:   dur,
		}
	}
	return items
}

func qualityToItems(qualities []kodik.Quality) []FilterItem {
	items := make([]FilterItem, len(qualities))
	for i, q := range qualities {
		typ := "MP4"
		if len(q.Links) > 0 {
			typ = player.LinkType(q.Links[0].Type)
		}
		items[i] = FilterItem{
			ID:    i,
			Label: q.Label,
			Sub:   typ,
		}
	}
	return items
}

// sortEpisodes sorts episodes by number in ascending order.
func sortEpisodes(episodes []api.VideoEntry) {
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Number < episodes[j].Number
	})
}
