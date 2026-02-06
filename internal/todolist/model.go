package todolist

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

// PreviewMsg is emitted when the user wants to preview a todo's body.
type PreviewMsg struct {
	Todo store.Todo
}

// mode represents the current input state of the todo list.
type mode int

const (
	normalMode          mode = iota
	inputMode                // typing todo text
	dateInputMode            // typing date for a dated todo
	editTextMode             // editing existing todo text
	editDateMode             // editing existing todo date
	filterMode               // inline filter narrowing visible todos
	templateSelectMode       // browsing/selecting a template
	placeholderInputMode     // filling in placeholder values one at a time
	templateNameMode         // entering a name for a new template
	templateContentMode      // entering content for a new template (multi-line)
)

// itemKind classifies a visible row in the rendered list.
type itemKind int

const (
	headerItem itemKind = iota
	todoItem
	emptyItem
)

// visibleItem is a single row in the combined todo list display.
type visibleItem struct {
	kind  itemKind
	label string      // display text for headers/empty
	todo  *store.Todo // non-nil only for todoItem
}

// Model represents the todo list pane.
type Model struct {
	focused     bool
	width       int
	height      int
	mode        mode
	cursor      int // index into selectable items (todo items only)
	input       textinput.Model
	store       store.TodoStore
	viewYear    int
	viewMonth   time.Month
	addingDated bool   // true if current add will produce a dated todo
	pendingText string // text saved during dateInputMode
	editingID       int    // ID of the todo being edited
	filterQuery     string // current filter text (empty = no filter)
	dateLayout      string // Go time layout for date display/input
	datePlaceholder string // human-readable date placeholder
	keys            KeyMap
	styles          Styles

	// Template workflow fields
	templates        []store.Template   // cached template list for selection
	templateCursor   int                // selection cursor in template list
	pendingTemplate  *store.Template    // selected template during placeholder flow
	placeholderNames []string           // extracted placeholder names from template
	placeholderIndex int                // which placeholder we're currently prompting for
	placeholderValues map[string]string  // collected values so far
	pendingTemplateName string           // name for template being created
	templateTextarea textarea.Model      // multi-line textarea for template content entry
	pendingBody      string             // template body to attach after todo creation
	fromTemplate     bool               // true when creating a todo from a template
}

// New creates a new todo list model backed by the given store.
func New(s store.TodoStore, t theme.Theme) Model {
	ti := textinput.New()
	ti.Placeholder = "What needs doing?"
	ti.CharLimit = 120
	ti.Prompt = "> "

	ta := textarea.New()
	ta.Placeholder = "Template content (use {{.VarName}} for placeholders)"
	ta.ShowLineNumbers = false

	now := time.Now()
	return Model{
		store:            s,
		input:            ti,
		templateTextarea: ta,
		viewYear:         now.Year(),
		viewMonth:        now.Month(),
		dateLayout:       "2006-01-02",
		datePlaceholder:  "YYYY-MM-DD",
		keys:             DefaultKeyMap(),
		styles:           NewStyles(t),
	}
}

// SetFocused sets whether this pane is focused.
func (m *Model) SetFocused(f bool) {
	m.focused = f
}

// SetViewMonth updates the filtered month for dated todos.
func (m *Model) SetViewMonth(year int, month time.Month) {
	m.viewYear = year
	m.viewMonth = month
	m.filterQuery = ""
	if m.mode == filterMode {
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
	}
}

// IsInputting returns true when the todo list is in text entry mode.
// The app uses this to suppress the quit keybinding.
func (m Model) IsInputting() bool {
	return m.mode != normalMode
}

// HelpBindings returns context-appropriate key bindings for the help bar.
func (m Model) HelpBindings() []key.Binding {
	if m.mode != normalMode {
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	}
	return []key.Binding{m.keys.Up, m.keys.Down, m.keys.MoveUp, m.keys.MoveDown, m.keys.Add, m.keys.AddDated, m.keys.Edit, m.keys.EditDate, m.keys.Toggle, m.keys.Delete, m.keys.Filter, m.keys.Preview, m.keys.TemplateUse, m.keys.TemplateCreate}
}

// visibleItems builds the combined display list of headers, todos, and empty placeholders.
func (m Model) visibleItems() []visibleItem {
	var items []visibleItem

	// Month section header
	monthLabel := fmt.Sprintf("%s %d", m.viewMonth.String(), m.viewYear)
	items = append(items, visibleItem{kind: headerItem, label: monthLabel})

	// Dated todos for the viewed month
	dated := m.store.TodosForMonth(m.viewYear, m.viewMonth)
	if len(dated) == 0 {
		items = append(items, visibleItem{kind: emptyItem, label: "(no todos this month)"})
	} else {
		for i := range dated {
			items = append(items, visibleItem{kind: todoItem, todo: &dated[i]})
		}
	}

	// Floating section header
	items = append(items, visibleItem{kind: headerItem, label: "Floating"})

	// Floating todos
	floating := m.store.FloatingTodos()
	if len(floating) == 0 {
		items = append(items, visibleItem{kind: emptyItem, label: "(no floating todos)"})
	} else {
		for i := range floating {
			items = append(items, visibleItem{kind: todoItem, todo: &floating[i]})
		}
	}

	// Apply inline filter when active
	if m.filterQuery != "" {
		query := strings.ToLower(m.filterQuery)
		var filtered []visibleItem
		for _, item := range items {
			switch item.kind {
			case headerItem:
				filtered = append(filtered, item)
			case todoItem:
				if strings.Contains(strings.ToLower(item.todo.Text), query) {
					filtered = append(filtered, item)
				}
			// Skip emptyItem entries -- they are misleading when filtered
			}
		}
		// Post-process: add "(no matches)" after headers with no following todos
		var result []visibleItem
		for i, item := range filtered {
			result = append(result, item)
			if item.kind == headerItem {
				// Check if next item is a todoItem or another header/end
				nextIsTodo := i+1 < len(filtered) && filtered[i+1].kind == todoItem
				if !nextIsTodo {
					result = append(result, visibleItem{kind: emptyItem, label: "(no matches)"})
				}
			}
		}
		return result
	}

	return items
}

// selectableIndices returns the indices of visible items that are selectable (todo items).
func selectableIndices(items []visibleItem) []int {
	var indices []int
	for i, item := range items {
		if item.kind == todoItem {
			indices = append(indices, i)
		}
	}
	return indices
}

// Update handles messages for the todo list pane.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch m.mode {
		case inputMode:
			return m.updateInputMode(msg)
		case dateInputMode:
			return m.updateDateInputMode(msg)
		case editTextMode:
			return m.updateEditTextMode(msg)
		case editDateMode:
			return m.updateEditDateMode(msg)
		case filterMode:
			return m.updateFilterMode(msg)
		default:
			return m.updateNormalMode(msg)
		}
	}

	return m, nil
}

// updateNormalMode handles key events in normal navigation mode.
func (m Model) updateNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	items := m.visibleItems()
	selectable := selectableIndices(items)

	switch {
	case key.Matches(msg, m.keys.Down):
		if len(selectable) > 0 && m.cursor < len(selectable)-1 {
			m.cursor++
		}

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, m.keys.MoveUp):
		if len(selectable) > 0 && m.cursor > 0 && m.cursor < len(selectable) {
			curIdx := selectable[m.cursor]
			prevIdx := selectable[m.cursor-1]
			curTodo := items[curIdx].todo
			prevTodo := items[prevIdx].todo
			if curTodo != nil && prevTodo != nil &&
				curTodo.HasDate() == prevTodo.HasDate() {
				m.store.SwapOrder(curTodo.ID, prevTodo.ID)
				m.cursor--
			}
		}

	case key.Matches(msg, m.keys.MoveDown):
		if len(selectable) > 0 && m.cursor >= 0 && m.cursor < len(selectable)-1 {
			curIdx := selectable[m.cursor]
			nextIdx := selectable[m.cursor+1]
			curTodo := items[curIdx].todo
			nextTodo := items[nextIdx].todo
			if curTodo != nil && nextTodo != nil &&
				curTodo.HasDate() == nextTodo.HasDate() {
				m.store.SwapOrder(curTodo.ID, nextTodo.ID)
				m.cursor++
			}
		}

	case key.Matches(msg, m.keys.Add):
		m.mode = inputMode
		m.addingDated = false
		m.input.Placeholder = "What needs doing?"
		m.input.Prompt = "> "
		m.input.SetValue("")
		return m, m.input.Focus()

	case key.Matches(msg, m.keys.AddDated):
		m.mode = inputMode
		m.addingDated = true
		m.input.Placeholder = "What needs doing?"
		m.input.Prompt = "> "
		m.input.SetValue("")
		return m, m.input.Focus()

	case key.Matches(msg, m.keys.Toggle):
		if len(selectable) > 0 && m.cursor < len(selectable) {
			idx := selectable[m.cursor]
			if items[idx].todo != nil {
				m.store.Toggle(items[idx].todo.ID)
			}
		}

	case key.Matches(msg, m.keys.Delete):
		if len(selectable) > 0 && m.cursor < len(selectable) {
			idx := selectable[m.cursor]
			if items[idx].todo != nil {
				m.store.Delete(items[idx].todo.ID)
			}
			// Clamp cursor after deletion
			newSelectable := selectableIndices(m.visibleItems())
			if m.cursor >= len(newSelectable) {
				m.cursor = max(0, len(newSelectable)-1)
			}
		}

	case key.Matches(msg, m.keys.Edit):
		if len(selectable) > 0 && m.cursor < len(selectable) {
			todo := items[selectable[m.cursor]].todo
			m.editingID = todo.ID
			m.mode = editTextMode
			m.input.Placeholder = "Edit todo text"
			m.input.Prompt = "> "
			m.input.SetValue(todo.Text)
			m.input.CursorEnd()
			return m, m.input.Focus()
		}

	case key.Matches(msg, m.keys.EditDate):
		if len(selectable) > 0 && m.cursor < len(selectable) {
			todo := items[selectable[m.cursor]].todo
			m.editingID = todo.ID
			m.mode = editDateMode
			m.input.Placeholder = m.datePlaceholder + " (empty = floating)"
			m.input.Prompt = "Date: "
			m.input.SetValue(config.FormatDate(todo.Date, m.dateLayout))
			m.input.CursorEnd()
			return m, m.input.Focus()
		}

	case key.Matches(msg, m.keys.Filter):
		m.mode = filterMode
		m.filterQuery = ""
		m.input.Placeholder = "Filter todos..."
		m.input.Prompt = "/ "
		m.input.SetValue("")
		return m, m.input.Focus()

	case key.Matches(msg, m.keys.Preview):
		if len(selectable) > 0 && m.cursor < len(selectable) {
			todo := items[selectable[m.cursor]].todo
			if todo != nil && todo.HasBody() {
				t := *todo
				return m, func() tea.Msg { return PreviewMsg{Todo: t} }
			}
		}
	}

	return m, nil
}

// updateFilterMode handles key events in filter mode.
func (m Model) updateFilterMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		// Esc clears filter and returns to normal mode
		m.mode = normalMode
		m.filterQuery = ""
		m.input.Blur()
		m.input.SetValue("")
		// Clamp cursor after filter removal restores all items
		selectable := selectableIndices(m.visibleItems())
		if m.cursor >= len(selectable) {
			m.cursor = max(0, len(selectable)-1)
		}
		return m, nil
	}
	// Forward all other keys to the text input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.filterQuery = m.input.Value()
	// Clamp cursor after filter narrows visible items
	selectable := selectableIndices(m.visibleItems())
	if m.cursor >= len(selectable) {
		m.cursor = max(0, len(selectable)-1)
	}
	return m, cmd
}

// updateInputMode handles key events while typing todo text.
func (m Model) updateInputMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			// Don't add empty todos -- stay in input mode
			return m, nil
		}
		if m.addingDated {
			m.pendingText = text
			m.mode = dateInputMode
			m.input.Placeholder = m.datePlaceholder
			m.input.Prompt = "Date: "
			m.input.SetValue("")
			return m, nil
		}
		m.store.Add(text, "")
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		return m, nil

	case key.Matches(msg, m.keys.Cancel):
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		m.pendingText = ""
		return m, nil
	}

	// Forward all other keys to the text input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// updateDateInputMode handles key events while typing a date.
func (m Model) updateDateInputMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		date := strings.TrimSpace(m.input.Value())
		if date == "" {
			return m, nil
		}
		// Parse in user's configured format, convert to ISO for storage
		isoDate, err := config.ParseUserDate(date, m.dateLayout)
		if err != nil {
			// Invalid date -- stay in date input mode
			return m, nil
		}
		m.store.Add(m.pendingText, isoDate)
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		m.pendingText = ""
		return m, nil

	case key.Matches(msg, m.keys.Cancel):
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		m.pendingText = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// updateEditTextMode handles key events while editing an existing todo's text.
func (m Model) updateEditTextMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			// Don't save empty text
			return m, nil
		}
		// Get current todo to preserve its date
		todo := m.store.Find(m.editingID)
		if todo != nil {
			m.store.Update(m.editingID, text, todo.Date)
		}
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		return m, nil

	case key.Matches(msg, m.keys.Cancel):
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// updateEditDateMode handles key events while editing an existing todo's date.
func (m Model) updateEditDateMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		date := strings.TrimSpace(m.input.Value())
		// Empty date is valid -- means "make floating"
		isoDate := ""
		if date != "" {
			var err error
			isoDate, err = config.ParseUserDate(date, m.dateLayout)
			if err != nil {
				// Invalid date -- stay in edit mode
				return m, nil
			}
		}
		// Get current todo to preserve its text
		todo := m.store.Find(m.editingID)
		if todo != nil {
			m.store.Update(m.editingID, todo.Text, isoDate)
		}
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		// Clamp cursor -- todo may have moved between sections
		newSelectable := selectableIndices(m.visibleItems())
		if m.cursor >= len(newSelectable) {
			m.cursor = max(0, len(newSelectable)-1)
		}
		return m, nil

	case key.Matches(msg, m.keys.Cancel):
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the todo list pane content.
func (m Model) View() string {
	items := m.visibleItems()
	selectable := selectableIndices(items)

	var b strings.Builder
	selectableIdx := 0

	for _, item := range items {
		switch item.kind {
		case headerItem:
			b.WriteString(m.styles.SectionHeader.Render(item.label))
			b.WriteString("\n")

		case emptyItem:
			b.WriteString("  " + m.styles.Empty.Render(item.label))
			b.WriteString("\n")

		case todoItem:
			isSelected := selectableIdx < len(selectable) && selectableIdx == m.cursor && m.focused
			m.renderTodo(&b, item.todo, isSelected)
			selectableIdx++
		}
	}

	// Show input field when in input/date mode
	if m.mode != normalMode {
		b.WriteString("\n")
		b.WriteString(m.input.View())
	}

	return b.String()
}

// renderTodo writes a single todo line to the builder.
func (m Model) renderTodo(b *strings.Builder, t *store.Todo, selected bool) {
	// Cursor indicator
	if selected {
		b.WriteString(m.styles.Cursor.Render("> "))
	} else {
		b.WriteString("  ")
	}

	// Checkbox
	check := "[ ] "
	if t.Done {
		check = "[x] "
	}

	// Text with optional body indicator and date (display in user's configured format)
	text := t.Text
	if t.HasBody() {
		text += " " + m.styles.BodyIndicator.Render("[+]")
	}
	if t.HasDate() {
		text += " " + m.styles.Date.Render(config.FormatDate(t.Date, m.dateLayout))
	}

	if t.Done {
		b.WriteString(m.styles.Completed.Render(check + text))
	} else {
		b.WriteString(check + text)
	}
	b.WriteString("\n")
}

// SetTheme replaces the todolist styles with ones built from the given theme.
// This preserves all model state (cursor, mode, input).
func (m *Model) SetTheme(t theme.Theme) {
	m.styles = NewStyles(t)
}

// SetDateFormat updates the date display layout and input placeholder.
func (m *Model) SetDateFormat(layout, placeholder string) {
	m.dateLayout = layout
	m.datePlaceholder = placeholder
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
