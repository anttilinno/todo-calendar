package search

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/fuzzy"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

// JumpMsg is emitted when the user selects a dated search result.
// The parent model uses Year and Month to navigate the calendar.
type JumpMsg struct {
	Year  int
	Month time.Month
}

// CloseMsg is emitted when the user presses Esc to close the search overlay.
type CloseMsg struct{}

// Model represents the search overlay.
type Model struct {
	input      textinput.Model
	results    []store.Todo
	cursor     int
	store      store.TodoStore
	allTodos   []store.Todo
	dateLayout    string
	priorityStyle string
	width         int
	height     int
	keys       KeyMap
	styles     Styles
}

// New creates a new search overlay model.
func New(s store.TodoStore, t theme.Theme, cfg config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Search all todos..."
	ti.Prompt = "? "
	ti.Focus()

	return Model{
		input:         ti,
		store:         s,
		allTodos:      s.Todos(),
		dateLayout:    cfg.DateLayout(),
		priorityStyle: cfg.PriorityStyle,
		keys:          DefaultKeyMap(),
		styles:        NewStyles(t),
	}
}

// Init returns the initial command (starts cursor blinking).
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize stores dimensions for layout.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// SetTheme replaces the styles with ones built from the given theme.
func (m *Model) SetTheme(t theme.Theme) {
	m.styles = NewStyles(t)
}

// Update handles messages for the search overlay.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			return m, func() tea.Msg { return CloseMsg{} }

		case key.Matches(msg, m.keys.Select):
			if len(m.results) > 0 && m.cursor >= 0 && m.cursor < len(m.results) {
				r := m.results[m.cursor]
				if r.HasDate() {
					d, err := time.Parse("2006-01-02", r.Date)
					if err == nil {
						year, month := d.Year(), d.Month()
						return m, func() tea.Msg {
							return JumpMsg{Year: year, Month: month}
						}
					}
				}
				// Floating todo -- no month to jump to, just close
				return m, func() tea.Msg { return CloseMsg{} }
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Forward to textinput and update results
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.results = m.fuzzySearch(m.input.Value())
	// Clamp cursor
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	return m, cmd
}

// View renders the search overlay.
func (m Model) View() string {
	var b strings.Builder

	title := m.styles.Title.Render("Search")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	query := m.input.Value()

	if query == "" {
		b.WriteString(m.styles.Empty.Render("Type to search across all months"))
	} else if len(m.results) == 0 {
		b.WriteString(m.styles.Empty.Render("(no matches)"))
	} else {
		maxVisible := m.height - 8
		if maxVisible < 1 {
			maxVisible = 1
		}
		visible := len(m.results)
		if visible > maxVisible {
			visible = maxVisible
		}
		for i := 0; i < visible; i++ {
			r := m.results[i]

			// Checkbox
			check := "[ ]"
			if r.Done {
				check = "[x]"
			}

			// Priority indicator -- signal bars or nerd icon
			badge := renderPriorityBars(r.Priority, m.priorityStyle, m.styles)

			// Date display
			dateStr := "No date"
			if r.HasDate() {
				dateStr = config.FormatDate(r.Date, m.dateLayout)
			}

			if i == m.cursor {
				b.WriteString(m.styles.SelectedResult.Render("> "))
				b.WriteString(badge)
				b.WriteString(m.styles.SelectedResult.Render(check + " " + r.Text))
				b.WriteString("  ")
				b.WriteString(m.styles.SelectedDate.Render(dateStr))
			} else {
				b.WriteString("  ")
				b.WriteString(badge)
				b.WriteString(m.styles.ResultText.Render(check + " " + r.Text))
				b.WriteString("  ")
				b.WriteString(m.styles.ResultDate.Render(dateStr))
			}
			b.WriteString("\n")
		}
	}

	content := b.String()

	// Center vertically if we have height information.
	if m.height > 0 {
		lines := strings.Count(content, "\n") + 1
		topPad := (m.height - lines) / 2
		if topPad > 0 {
			content = strings.Repeat("\n", topPad) + content
		}
	}

	return content
}

// barChars are the ascending block characters for the signal-strength meter.
var barChars = [3]rune{'▁', '▃', '▅'}

// nerdIcons maps priority level (1-3) to Nerd Font MDI signal cellular icons.
var nerdIcons = [4]string{
	"",           // 0 = no priority
	"\U000F08BE", // P1 = nf-md-signal_cellular_3
	"\U000F08BD", // P2 = nf-md-signal_cellular_2
	"\U000F08BC", // P3 = nf-md-signal_cellular_1
}

// renderPriorityBars returns the priority indicator string for a search result.
func renderPriorityBars(priority int, style string, s Styles) string {
	// Clamp legacy P4 values to P3
	if priority > 3 {
		priority = 3
	}
	if style == "nerd" {
		if priority >= 1 && priority <= 3 {
			return s.priorityBadgeStyle(priority).Render(nerdIcons[priority]) + " "
		}
		return "  " // 1 icon-width + 1 space
	}
	// Default "bars" mode
	if priority < 1 || priority > 3 {
		return "    " // 3 bar-width + 1 space
	}
	filled := 4 - priority // P1=3, P2=2, P3=1
	colorStyle := s.priorityBadgeStyle(priority)
	var result string
	for i, ch := range barChars {
		if i < filled {
			result += colorStyle.Render(string(ch))
		} else {
			result += " "
		}
	}
	return result + " "
}

// fuzzySearch filters allTodos by fuzzy match and sorts by score (best first).
func (m Model) fuzzySearch(query string) []store.Todo {
	if query == "" {
		return nil
	}

	type scored struct {
		todo  store.Todo
		score int
	}

	var matches []scored
	for _, t := range m.allTodos {
		if matched, score := fuzzy.Match(query, t.Text); matched {
			matches = append(matches, scored{todo: t, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].todo.Date > matches[j].todo.Date
	})

	results := make([]store.Todo, len(matches))
	for i, m := range matches {
		results[i] = m.todo
	}
	return results
}

// HelpBindings returns search-specific key bindings for help bar display.
func (m Model) HelpBindings() []key.Binding {
	return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Select, m.keys.Cancel}
}
