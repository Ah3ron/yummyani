// Package tui implements a BubbleTea TUI for YummyAnime.
package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"yummyani/internal/api"
	"yummyani/internal/kodik"
	"yummyani/internal/player"
)

// viewState represents the current screen in the TUI.
type viewState int

const (
	viewSearch viewState = iota
	viewResults
	viewDubbing
	viewEpisodes
	viewQuality
	viewExtracting
	viewPlaying
	viewError
)

// Model holds all state for the BubbleTea TUI application.
//
// All external dependencies (API client, Kodik parser, player) are
// injected through the constructor, enabling testability and DI.
type Model struct {
	ctx context.Context

	state  viewState
	width  int
	height int
	filter FzfFilter
	err    error

	// Injected dependencies (interfaces).
	anime         api.AnimeProvider
	linkExtractor kodik.LinkExtractor
	player        player.Player

	spinner spinner.Model

	// Search flow.
	query   string
	results []api.SearchResult
	animeID int

	// Anime info.
	animeTitle string

	// Episode flow.
	groups   []api.DubbingGroup
	groupIdx int
	episodes []api.VideoEntry
	epIdx    int

	// Quality selection.
	kodikResp          *kodik.Response
	availableQualities []kodik.Quality
	selectedQuality    kodik.Quality

	status      string
	searchLimit int
}

// Option configures a [Model].
type Option func(*Model)

// WithSearchLimit sets the maximum number of search results.
func WithSearchLimit(n int) Option {
	return func(m *Model) { m.searchLimit = n }
}

// WithMaxVisibleLines sets the default visible lines for the filter.
func WithMaxVisibleLines(n int) Option {
	return func(m *Model) { m.filter.SetMaxLines(n) }
}

// NewModel creates a Model with the given dependencies and options.
func NewModel(
	ctx context.Context,
	animeProvider api.AnimeProvider,
	linkExtractor kodik.LinkExtractor,
	p player.Player,
	opts ...Option,
) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	f := NewFzfFilter("🔍 ")
	f.SetPlaceholder("Название аниме...")

	m := Model{
		ctx:           ctx,
		state:         viewSearch,
		spinner:       s,
		filter:        f,
		anime:         animeProvider,
		linkExtractor: linkExtractor,
		player:        p,
		searchLimit:   10,
	}

	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// Init returns the initial commands (spinner tick + cursor blink).
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.filter.Init())
}

// Update dispatches BubbleTea messages to the appropriate handler.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Window resize.
	if w, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = w.Width
		m.height = w.Height
		m.filter.SetMaxLines(w.Height - 6)
		m.filter.SetWidth(w.Width - 4)
		return m, nil
	}

	// Spinner tick.
	if tick, ok := msg.(spinner.TickMsg); ok {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(tick)
		return m, cmd
	}

	// Async results.
	switch msg := msg.(type) {
	case searchDoneMsg:
		return m.handleSearchDone(msg)
	case videosDoneMsg:
		return m.handleVideosDone(msg)
	case extractDoneMsg:
		return m.handleExtractDone(msg)
	case playDoneMsg:
		return m.handlePlayDone(msg)
	}

	// Key handling.
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if key.String() == "ctrl+c" {
		return m, tea.Quit
	}

	return m.handleKey(key)
}
