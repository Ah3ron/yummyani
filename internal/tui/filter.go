package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"yummyani/pkg/fuzzy"
)

type FilterItem struct {
	ID    int
	Label string
	Sub   string
}

type FzfFilter struct {
	items    []FilterItem
	filtered []FilterItem
	origIdx  []int
	cursor   int
	input    textinput.Model
	maxLines int
	width    int
}

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

func (f *FzfFilter) SetPrompt(p string) {
	f.input.Prompt = p
}

func (f *FzfFilter) SetPlaceholder(p string) {
	f.input.Placeholder = p
}

func (f *FzfFilter) Value() string {
	return strings.TrimSpace(f.input.Value())
}

func (f *FzfFilter) Selected() *FilterItem {
	if len(f.filtered) == 0 || f.cursor < 0 || f.cursor >= len(f.filtered) {
		return nil
	}
	return &f.filtered[f.cursor]
}

func (f *FzfFilter) SetMaxLines(n int) {
	if n > 0 {
		f.maxLines = n
	}
}

func (f *FzfFilter) SetWidth(w int) {
	f.width = w
}

func (f FzfFilter) Init() tea.Cmd {
	return textinput.Blink
}

func (f FzfFilter) Update(msg tea.Msg) (FzfFilter, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return f, nil
	}
	return f.handleKey(key)
}

func (f FzfFilter) handleKey(key tea.KeyMsg) (FzfFilter, tea.Cmd) {
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
	f.input, cmd = f.input.Update(key)
	f.refilter()

	if len(f.filtered) > 0 && f.cursor >= len(f.filtered) {
		f.cursor = len(f.filtered) - 1
	}

	return f, cmd
}

func (f FzfFilter) View() string {
	input := f.input.View()
	if f.width > 0 {
		return inputStyle.Width(f.width).Render(input)
	}
	return inputStyle.Render(input)
}

func (f FzfFilter) ViewItems() string {
	if len(f.filtered) == 0 {
		if len(f.items) == 0 {
			return ""
		}
		return "\n" + dimStyle.Render("  нет результатов")
	}

	maxL := f.maxLines
	if maxL <= 0 {
		maxL = 20
	}

	offset := 0
	if f.cursor >= maxL {
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
			b.WriteString(accentStyle.Render("▸ ") + accentStyle.Render(num) + " " + itemSelStyle.Render(item.Label) + sub + "\n")
		} else {
			b.WriteString("  " + dimStyle.Render(num) + " " + itemNormStyle.Render(item.Label) + sub + "\n")
		}
	}

	if len(f.filtered) != len(f.items) && len(f.items) > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", len(f.filtered), len(f.items))))
	}

	return b.String()
}

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
