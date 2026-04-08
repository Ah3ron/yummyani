package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"yummyani/pkg/fuzzy"
)

// FilterItem is a single selectable item in the filter list.
type FilterItem struct {
	ID    int
	Label string
	Sub   string
}

// FzfFilter is a reusable fzf-like incremental filter component.
//
// Features:
//   - Type to filter items in real time (fuzzy match)
//   - Arrow keys / Ctrl+J,K,P,N to navigate
//   - Ctrl+U to clear input
//   - Esc passed through to parent (for back navigation)
//   - Scrollable list when items exceed visible area
//   - Original numbering preserved when filtering
type FzfFilter struct {
	items    []FilterItem
	filtered []FilterItem
	origIdx  []int
	cursor   int
	input    textinput.Model
	maxLines int
	width    int
}

// NewFzfFilter creates a filter with the given prompt string.
func NewFzfFilter(prompt string) FzfFilter {
	ti := textinput.New()
	ti.Prompt = prompt
	ti.PromptStyle = accentStyle.Background(colorFilterBg)
	ti.TextStyle = normalStyle.Background(colorFilterBg)
	ti.PlaceholderStyle = dimStyle.Background(colorFilterBg)
	ti.CharLimit = 200
	ti.Focus()

	return FzfFilter{
		input:    ti,
		maxLines: 20,
	}
}

// Reset atomically clears the input and sets new items.
func (f *FzfFilter) Reset(items []FilterItem) {
	f.input.SetValue("")
	f.items = items
	f.cursor = 0
	f.filtered = make([]FilterItem, len(items))
	copy(f.filtered, items)
	f.origIdx = make([]int, len(items))
	for i := range f.origIdx {
		f.origIdx[i] = i
	}
}

// SetPrompt changes the input prompt.
func (f *FzfFilter) SetPrompt(p string) {
	f.input.Prompt = p
}

// SetPlaceholder changes the placeholder text.
func (f *FzfFilter) SetPlaceholder(p string) {
	f.input.Placeholder = p
}

// Value returns the trimmed input text.
func (f *FzfFilter) Value() string {
	return strings.TrimSpace(f.input.Value())
}

// Selected returns the currently highlighted item, or nil.
func (f *FzfFilter) Selected() *FilterItem {
	if len(f.filtered) == 0 || f.cursor < 0 || f.cursor >= len(f.filtered) {
		return nil
	}
	return &f.filtered[f.cursor]
}

// Len returns the number of currently filtered items.
func (f *FzfFilter) Len() int {
	return len(f.filtered)
}

// SetMaxLines sets how many items are visible at once.
func (f *FzfFilter) SetMaxLines(n int) {
	if n > 0 {
		f.maxLines = n
	}
}

// SetWidth sets the terminal width for background rendering.
func (f *FzfFilter) SetWidth(w int) {
	f.width = w
}

// Init returns the initial command (cursor blink).
func (f FzfFilter) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles key events for the filter.
func (f FzfFilter) Update(msg tea.Msg) (FzfFilter, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return f, nil
	}

	switch key.String() {
	case "up", "ctrl+k", "ctrl+p":
		if f.cursor > 0 {
			f.cursor--
		}
		return f, nil

	case "down", "ctrl+j", "ctrl+n":
		if f.cursor < len(f.filtered)-1 {
			f.cursor++
		}
		return f, nil

	case "ctrl+u":
		f.input.SetValue("")
		f.refilter()
		return f, nil

	case "esc":
		return f, nil
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	f.refilter()

	if len(f.filtered) > 0 && f.cursor >= len(f.filtered) {
		f.cursor = len(f.filtered) - 1
	}

	return f, cmd
}

// View renders the textinput line with a full-width background.
func (f FzfFilter) View() string {
	input := f.input.View()
	if f.width <= 0 {
		return filterStyle.Render(input)
	}
	// Fill the entire width with background so no gaps remain.
	return filterStyle.Width(f.width).Render(input)
}

// ViewItems renders the scrollable filtered list with cursor indicator.
func (f FzfFilter) ViewItems() string {
	if len(f.filtered) == 0 {
		if len(f.items) == 0 {
			return ""
		}
		return "  " + dimStyle.Render("нет результатов")
	}

	maxL := f.maxLines
	if maxL <= 0 {
		maxL = 20
	}

	// Scrolling.
	offset := 0
	if f.cursor >= offset+maxL {
		offset = f.cursor - maxL + 1
	}
	if f.cursor < offset {
		offset = f.cursor
	}
	end := min(offset+maxL, len(f.filtered))

	numW := max(len(fmt.Sprintf("%d", len(f.items))), 2)

	var b strings.Builder
	b.WriteString("\n")

	for i := offset; i < end; i++ {
		item := f.filtered[i]
		orig := f.origIdx[i] + 1
		num := fmt.Sprintf("%*d", numW, orig)
		sub := ""
		if item.Sub != "" {
			sub = "  " + dimStyle.Render(item.Sub)
		}

		if i == f.cursor {
			b.WriteString("  " + accentStyle.Render("▸") + " " + accentStyle.Render(num) + " " + accentStyle.Render(item.Label) + sub + "\n")
		} else {
			b.WriteString("    " + dimStyle.Render(num) + " " + normalStyle.Render(item.Label) + sub + "\n")
		}
	}

	if len(f.filtered) != len(f.items) && len(f.items) > 0 {
		b.WriteString("\n  " + dimStyle.Render(fmt.Sprintf("%d/%d", len(f.filtered), len(f.items))))
	}

	return listStyle.Render(b.String())
}

// refilter rebuilds f.filtered based on the current input value.
func (f *FzfFilter) refilter() {
	query := strings.ToLower(strings.TrimSpace(f.input.Value()))
	if query == "" {
		f.filtered = make([]FilterItem, len(f.items))
		copy(f.filtered, f.items)
		f.origIdx = make([]int, len(f.items))
		for i := range f.origIdx {
			f.origIdx[i] = i
		}
		return
	}

	f.filtered = f.filtered[:0]
	f.origIdx = f.origIdx[:0]
	for i, item := range f.items {
		searchable := fmt.Sprintf("%d %s %s", i+1, item.Label, item.Sub)
		if fuzzy.Match(query, strings.ToLower(searchable)) {
			f.filtered = append(f.filtered, item)
			f.origIdx = append(f.origIdx, i)
		}
	}
}
