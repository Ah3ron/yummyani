// Package tui implements the terminal UI using bubbletea.
package tui

import (
        "fmt"
        "os"
        "os/exec"
        "strings"

        "github.com/charmbracelet/bubbles/spinner"
        "github.com/charmbracelet/bubbles/textinput"
        tea "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/lipgloss"

        "yummyani/internal/api"
        "yummyani/internal/kodik"

        "net/http"
)

// --- State machine ---

type viewState int

const (
        viewSearch viewState = iota
        viewResults
        viewDubbing
        viewEpisodes
        viewExtracting
        viewPlaying
        viewError
)

// --- Model ---

// Model is the main bubbletea model.
type Model struct {
        state    viewState
        width    int
        height   int
        client   *api.Client
        kodikCLI *http.Client // HTTP client for kodik
        spinner  spinner.Model
        input    textinput.Model
        err      error

        // Data flow.
        query     string
        results   []api.SearchResult
        animeID   int
        animeTitle string
        groups    []api.DubbingGroup
        groupIdx  int
        episodes  []api.VideoEntry
        epIdx     int
        cursor    int

        // Status.
        status string
}

// --- Messages ---

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
        url string
        err error
}

type playDoneMsg struct {
        err error
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// --- Init ---

// NewModel creates the initial TUI model.
func NewModel() Model {
        s := spinner.New()
        s.Spinner = spinner.Dot
        s.Style = lipgloss.NewStyle().Foreground(accent)

        ti := textinput.New()
        ti.Placeholder = "Название аниме..."
        ti.Focus()
        ti.CharLimit = 200
        ti.Width = 50

        return Model{
                state:   viewSearch,
                spinner: s,
                input:   ti,
                client:  api.NewClient(),
                kodikCLI: kodik.NewHTTPClient(),
        }
}

func (m Model) Init() tea.Cmd {
        return tea.Batch(
                m.spinner.Tick,
                textinput.Blink,
        )
}

// --- Update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
        case tea.WindowSizeMsg:
                m.width = msg.Width
                m.height = msg.Height
                return m, nil

        case tea.KeyMsg:
                return m.handleKey(msg)

        case spinner.TickMsg:
                var cmd tea.Cmd
                m.spinner, cmd = m.spinner.Update(msg)
                return m, cmd

        case searchDoneMsg:
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
                m.state = viewResults
                m.cursor = 0
                return m, nil

        case videosDoneMsg:
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
                m.state = viewDubbing
                m.cursor = 0
                return m, nil

        case extractDoneMsg:
                if msg.err != nil {
                        m.err = msg.err
                        m.state = viewError
                        return m, nil
                }
                // Start playback.
                m.state = viewPlaying
                m.status = msg.url
                return m, m.startPlay(msg.url)

        case playDoneMsg:
                if msg.err != nil {
                        m.err = msg.err
                        m.state = viewError
                        return m, nil
                }
                // After playback finishes, go back to episodes.
                m.state = viewEpisodes
                m.status = ""
                return m, nil

        case errMsg:
                m.err = msg.err
                m.state = viewError
                return m, nil
        }

        return m, nil
}

// --- View ---

func (m Model) View() string {
        switch m.state {
        case viewSearch:
                return m.viewSearch()
        case viewResults:
                return m.viewResults()
        case viewDubbing:
                return m.viewDubbing()
        case viewEpisodes:
                return m.viewEpisodes()
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

// --- Key handling ---

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch m.state {
        case viewSearch:
                return m.handleSearchKey(msg)
        case viewResults:
                return m.handleResultsKey(msg)
        case viewDubbing:
                return m.handleDubbingKey(msg)
        case viewEpisodes:
                return m.handleEpisodesKey(msg)
        case viewError:
                if msg.String() == "q" || msg.String() == "esc" {
                        return m, tea.Quit
                }
                if msg.String() == "enter" {
                        m.state = viewSearch
                        m.err = nil
                        return m, nil
                }
        case viewPlaying:
                if msg.String() == "q" {
                        return m, tea.Quit
                }
        }

        // Global: ctrl+c quits.
        if msg.String() == "ctrl+c" {
                return m, tea.Quit
        }

        return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        if msg.String() == "enter" {
                q := strings.TrimSpace(m.input.Value())
                if q == "" {
                        return m, nil
                }
                m.query = q
                m.state = viewExtracting
                m.status = "Поиск..."
                return m, m.doSearch(q)
        }

        if msg.String() == "ctrl+c" {
                return m, tea.Quit
        }

        var cmd tea.Cmd
        m.input, cmd = m.input.Update(msg)
        return m, cmd
}

func (m Model) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "up", "k":
                if m.cursor > 0 {
                        m.cursor--
                }
        case "down", "j":
                if m.cursor < len(m.results)-1 {
                        m.cursor++
                }
        case "enter":
                selected := m.results[m.cursor]
                m.animeID = selected.AnimeID
                m.state = viewExtracting
                m.status = "Загрузка видео..."
                return m, m.doFetchVideos(selected.AnimeID)
        case "q", "esc":
                m.state = viewSearch
                m.results = nil
                return m, nil
        }
        return m, nil
}

func (m Model) handleDubbingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "up", "k":
                if m.cursor > 0 {
                        m.cursor--
                }
        case "down", "j":
                if m.cursor < len(m.groups)-1 {
                        m.cursor++
                }
        case "enter":
                m.groupIdx = m.cursor
                m.episodes = m.groups[m.cursor].Episodes
                // Sort episodes by number.
                sortEpisodes(m.episodes)
                m.state = viewEpisodes
                m.cursor = 0
                return m, nil
        case "q", "esc":
                m.state = viewResults
                m.groups = nil
                return m, nil
        }
        return m, nil
}

func (m Model) handleEpisodesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "up", "k":
                if m.cursor > 0 {
                        m.cursor--
                }
        case "down", "j":
                if m.cursor < len(m.episodes)-1 {
                        m.cursor++
                }
        case "enter":
                ep := m.episodes[m.cursor]
                iframeURL := ep.IframeURL
                if strings.HasPrefix(iframeURL, "//") {
                        iframeURL = "https:" + iframeURL
                }
                m.epIdx = m.cursor
                m.state = viewExtracting
                m.status = fmt.Sprintf("Извлечение ссылки (Серия %d)...", ep.Number)
                return m, m.doExtract(iframeURL)
        case "q", "esc":
                m.state = viewDubbing
                m.episodes = nil
                m.cursor = m.groupIdx
                return m, nil
        case "b":
                m.state = viewDubbing
                m.episodes = nil
                m.cursor = m.groupIdx
                return m, nil
        }
        return m, nil
}

// --- Async commands ---

func (m Model) doSearch(query string) tea.Cmd {
        return func() tea.Msg {
                results, err := m.client.Search(query, 10)
                return searchDoneMsg{results: results, err: err}
        }
}

func (m Model) doFetchVideos(animeID int) tea.Cmd {
        return func() tea.Msg {
                var title string
                videos, err := m.client.GetVideos(animeID)
                if err != nil {
                        return videosDoneMsg{err: err}
                }

                info, err := m.client.GetAnime(animeID)
                if err == nil {
                        title = info.Title
                        if title == "" {
                                title = info.Name
                        }
                }
                if title == "" {
                        title = fmt.Sprintf("Anime #%d", animeID)
                }

                return videosDoneMsg{videos: videos, title: title}
        }
}

func (m Model) doExtract(iframeURL string) tea.Cmd {
        return func() tea.Msg {
                resp, err := kodik.Parse(m.kodikCLI, iframeURL)
                if err != nil {
                        return extractDoneMsg{err: fmt.Errorf("kodik parse: %w", err)}
                }
                link := kodik.BestLink(resp)
                if link == "" {
                        return extractDoneMsg{err: fmt.Errorf("no playable links found")}
                }
                return extractDoneMsg{url: link}
        }
}

func (m Model) startPlay(url string) tea.Cmd {
        return func() tea.Msg {
                title := fmt.Sprintf("%s — Серия %d [%s]",
                        m.animeTitle,
                        m.episodes[m.epIdx].Number,
                        m.groups[m.groupIdx].Name)

                cmd := exec.Command("mpv",
                        "--title="+title,
                        "--force-media-title="+title,
                        "--no-terminal",
                        "--msg-level=all=error",
                        "--force-seekable=yes",
                        "--http-header-fields=Referer: https://kodikplayer.com/",
                        url,
                )
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                cmd.Stdin = os.Stdin

                err := cmd.Run()
                return playDoneMsg{err: err}
        }
}

// --- Sort helper ---

func sortEpisodes(episodes []api.VideoEntry) {
        // Simple insertion sort by number.
        for i := 1; i < len(episodes); i++ {
                key := episodes[i]
                j := i - 1
                for j >= 0 && episodes[j].Number > key.Number {
                        episodes[j+1] = episodes[j]
                        j--
                }
                episodes[j+1] = key
        }
}
