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

// OpenEditorMsg is emitted when the user wants to edit a todo's body in an external editor.
type OpenEditorMsg struct {
	Todo store.Todo
}

// mode represents the current input state of the todo list.
type mode int

const (
	normalMode          mode = iota
	inputMode                // typing todo text
	editMode                 // editing existing todo (title + date + body)
	filterMode               // inline filter narrowing visible todos
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
	editingID       int    // ID of the todo being edited
	filterQuery     string // current filter text (empty = no filter)
	dateLayout      string // Go time layout for date display/input
	datePlaceholder string // human-readable date placeholder
	keys            KeyMap
	styles          Styles

	// Full-pane edit fields
	dateInput     textinput.Model // separate input for date field in full-pane mode
	bodyTextarea  textarea.Model  // textarea for body editing in edit mode
	editField     int             // 0 = title, 1 = date, 2 = body, 3 = template
	templateInput textinput.Model // placeholder input for template field (Phase 25 adds picker)

	// Template workflow fields
	pendingTemplateName string           // name for template being created
	templateTextarea textarea.Model      // multi-line textarea for template content entry
}

// New creates a new todo list model backed by the given store.
func New(s store.TodoStore, t theme.Theme) Model {
	ti := textinput.New()
	ti.Placeholder = "What needs doing?"
	ti.CharLimit = 120
	ti.Prompt = "> "

	di := textinput.New()
	di.Placeholder = "YYYY-MM-DD"
	di.Prompt = "Date: "
	di.CharLimit = 10

	ta := textarea.New()
	ta.Placeholder = "Template content (use {{.VarName}} for placeholders)"
	ta.ShowLineNumbers = false

	ba := textarea.New()
	ba.Placeholder = "Body text (markdown supported)"
	ba.ShowLineNumbers = false

	tmplInput := textinput.New()
	tmplInput.Placeholder = "Press Enter to select template"
	tmplInput.Prompt = "> "
	tmplInput.CharLimit = 0 // Read-only placeholder for Phase 25

	now := time.Now()
	return Model{
		store:            s,
		input:            ti,
		dateInput:        di,
		bodyTextarea:     ba,
		templateInput:    tmplInput,
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

// SetSize sets the available pane dimensions for layout calculations.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
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

// HelpBindings returns a short list of context-appropriate key bindings for the help bar.
// In normal mode, shows max 5 most-used keys (HELP-01).
func (m Model) HelpBindings() []key.Binding {
	switch m.mode {
	case inputMode:
		if m.editField == 2 || m.editField == 3 {
			return []key.Binding{m.keys.SwitchField, m.keys.Save, m.keys.Cancel}
		}
		return []key.Binding{m.keys.SwitchField, m.keys.Confirm, m.keys.Cancel}
	case editMode:
		if m.editField == 2 {
			return []key.Binding{m.keys.SwitchField, m.keys.Save, m.keys.Cancel}
		}
		return []key.Binding{m.keys.SwitchField, m.keys.Confirm, m.keys.Cancel}
	case normalMode:
		return []key.Binding{m.keys.Add, m.keys.Toggle, m.keys.Delete, m.keys.Edit, m.keys.Filter}
	case templateContentMode:
		return []key.Binding{m.keys.Save, m.keys.Cancel}
	default:
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	}
}

// AllHelpBindings returns all key bindings for the expanded help view.
// In non-normal modes, still returns only Confirm/Cancel (HELP-02).
func (m Model) AllHelpBindings() []key.Binding {
	switch m.mode {
	case inputMode:
		if m.editField == 2 || m.editField == 3 {
			return []key.Binding{m.keys.SwitchField, m.keys.Save, m.keys.Cancel}
		}
		return []key.Binding{m.keys.SwitchField, m.keys.Confirm, m.keys.Cancel}
	case editMode:
		if m.editField == 2 {
			return []key.Binding{m.keys.SwitchField, m.keys.Save, m.keys.Cancel}
		}
		return []key.Binding{m.keys.SwitchField, m.keys.Confirm, m.keys.Cancel}
	case normalMode:
		return []key.Binding{
			m.keys.Up, m.keys.Down, m.keys.MoveUp, m.keys.MoveDown,
			m.keys.Add, m.keys.Edit,
			m.keys.Toggle, m.keys.Delete, m.keys.Filter,
			m.keys.Preview, m.keys.OpenEditor, m.keys.TemplateCreate,
		}
	case templateContentMode:
		return []key.Binding{m.keys.Save, m.keys.Cancel}
	default:
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	}
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
	// templateContentMode needs all message types (not just KeyMsg) for textarea blink/tick.
	if m.mode == templateContentMode && m.focused {
		return m.updateTemplateContentMode(msg)
	}

	// Forward blink/tick messages to focused text input in edit modes.
	switch m.mode {
	case inputMode, editMode:
		if _, ok := msg.(tea.KeyMsg); !ok {
			if _, ok := msg.(tea.WindowSizeMsg); !ok {
				var cmd tea.Cmd
				switch m.editField {
				case 2:
					m.bodyTextarea, cmd = m.bodyTextarea.Update(msg)
				case 3:
					m.templateInput, cmd = m.templateInput.Update(msg)
				case 1:
					m.dateInput, cmd = m.dateInput.Update(msg)
				default:
					m.input, cmd = m.input.Update(msg)
				}
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch m.mode {
		case inputMode:
			return m.updateInputMode(msg)
		case editMode:
			return m.updateEditMode(msg)
		case filterMode:
			return m.updateFilterMode(msg)
		case templateNameMode:
			return m.updateTemplateNameMode(msg)
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
		m.editField = 0
		m.input.Placeholder = "What needs doing?"
		m.input.Prompt = "> "
		m.input.SetValue("")
		m.dateInput.SetValue("")
		m.dateInput.Placeholder = m.datePlaceholder + " (empty = floating)"
		m.bodyTextarea.SetValue("")
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
			// Fetch fresh from store to get current body content
			fresh := m.store.Find(todo.ID)
			if fresh == nil {
				return m, nil
			}
			m.editingID = fresh.ID
			m.mode = editMode
			m.editField = 0
			m.input.Placeholder = "Todo title"
			m.input.Prompt = "> "
			m.input.SetValue(fresh.Text)
			m.input.CursorEnd()
			m.dateInput.Placeholder = m.datePlaceholder + " (empty = floating)"
			m.dateInput.Prompt = "> "
			m.dateInput.SetValue(config.FormatDate(fresh.Date, m.dateLayout))
			m.bodyTextarea.SetValue(fresh.Body)
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
			if todo != nil {
				t := *todo
				return m, func() tea.Msg { return PreviewMsg{Todo: t} }
			}
		}

	case key.Matches(msg, m.keys.OpenEditor):
		if len(selectable) > 0 && m.cursor < len(selectable) {
			todo := items[selectable[m.cursor]].todo
			if todo != nil {
				// Fetch fresh from store to get current body content.
				fresh := m.store.Find(todo.ID)
				if fresh != nil {
					t := *fresh
					return m, func() tea.Msg { return OpenEditorMsg{Todo: t} }
				}
			}
		}

	case key.Matches(msg, m.keys.TemplateCreate):
		m.mode = templateNameMode
		m.input.Placeholder = "Template name"
		m.input.Prompt = "> "
		m.input.SetValue("")
		return m, m.input.Focus()
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

// updateInputMode handles key events in the 4-field add form (title, date, body, template).
func (m Model) updateInputMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Save):
		// Ctrl+D saves from any field
		return m.saveAdd()

	case key.Matches(msg, m.keys.Confirm):
		// Body field uses Enter for newlines
		if m.editField == 2 {
			var cmd tea.Cmd
			m.bodyTextarea, cmd = m.bodyTextarea.Update(msg)
			return m, cmd
		}
		// Template field forwards Enter (Phase 25 will open picker)
		if m.editField == 3 {
			return m, nil
		}
		// Title/Date: Enter saves
		return m.saveAdd()

	case key.Matches(msg, m.keys.SwitchField):
		// Cycle: title(0) -> date(1) -> body(2) -> template(3) -> title(0)
		switch m.editField {
		case 0:
			m.editField = 1
			m.input.Blur()
			return m, m.dateInput.Focus()
		case 1:
			m.editField = 2
			m.dateInput.Blur()
			return m, m.bodyTextarea.Focus()
		case 2:
			m.editField = 3
			m.bodyTextarea.Blur()
			return m, m.templateInput.Focus()
		case 3:
			m.editField = 0
			m.templateInput.Blur()
			return m, m.input.Focus()
		}
		return m, nil

	case key.Matches(msg, m.keys.Cancel):
		if m.editField == 2 || m.editField == 3 {
			// Esc in body/template goes back to title field instead of cancelling
			m.editField = 0
			m.bodyTextarea.Blur()
			m.templateInput.Blur()
			return m, m.input.Focus()
		}
		// Esc in title/date cancels entirely
		m.mode = normalMode
		m.input.Blur()
		m.dateInput.Blur()
		m.bodyTextarea.Blur()
		m.templateInput.Blur()
		m.input.SetValue("")
		m.dateInput.SetValue("")
		m.bodyTextarea.SetValue("")
		m.editField = 0
		return m, nil
	}

	// Forward key events to the focused field
	var cmd tea.Cmd
	switch m.editField {
	case 0:
		m.input, cmd = m.input.Update(msg)
	case 1:
		m.dateInput, cmd = m.dateInput.Update(msg)
	case 2:
		m.bodyTextarea, cmd = m.bodyTextarea.Update(msg)
	case 3:
		m.templateInput, cmd = m.templateInput.Update(msg)
	}
	return m, cmd
}

// updateEditMode handles key events while editing an existing todo (title + date + body).
func (m Model) updateEditMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Save):
		// Ctrl+D saves from any field (including body)
		return m.saveEdit()

	case key.Matches(msg, m.keys.Confirm):
		// Body field uses Enter for newlines; use Ctrl+D or Tab away to save
		if m.editField == 2 {
			// In body textarea, Enter inserts a newline — forward to textarea
			var cmd tea.Cmd
			m.bodyTextarea, cmd = m.bodyTextarea.Update(msg)
			return m, cmd
		}
		// Save all three fields
		return m.saveEdit()

	case key.Matches(msg, m.keys.SwitchField):
		// Cycle: title(0) → date(1) → body(2) → title(0)
		switch m.editField {
		case 0:
			m.editField = 1
			m.input.Blur()
			return m, m.dateInput.Focus()
		case 1:
			m.editField = 2
			m.dateInput.Blur()
			return m, m.bodyTextarea.Focus()
		case 2:
			m.editField = 0
			m.bodyTextarea.Blur()
			return m, m.input.Focus()
		}
		return m, nil

	case key.Matches(msg, m.keys.Cancel):
		if m.editField == 2 {
			// Esc in body field goes back to title field instead of cancelling
			m.editField = 0
			m.bodyTextarea.Blur()
			return m, m.input.Focus()
		}
		m.mode = normalMode
		m.input.Blur()
		m.dateInput.Blur()
		m.bodyTextarea.Blur()
		m.input.SetValue("")
		m.dateInput.SetValue("")
		m.bodyTextarea.SetValue("")
		m.editField = 0
		return m, nil
	}

	// Forward key events to the focused field
	var cmd tea.Cmd
	switch m.editField {
	case 0:
		m.input, cmd = m.input.Update(msg)
	case 1:
		m.dateInput, cmd = m.dateInput.Update(msg)
	case 2:
		m.bodyTextarea, cmd = m.bodyTextarea.Update(msg)
	}
	return m, cmd
}

// saveEdit persists all three fields and returns to normal mode.
func (m Model) saveEdit() (Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return m, nil
	}

	date := strings.TrimSpace(m.dateInput.Value())
	isoDate := ""
	if date != "" {
		var err error
		isoDate, err = config.ParseUserDate(date, m.dateLayout)
		if err != nil {
			// Invalid date — focus date field
			m.editField = 1
			m.input.Blur()
			return m, m.dateInput.Focus()
		}
	}

	body := m.bodyTextarea.Value()

	m.store.Update(m.editingID, text, isoDate)
	m.store.UpdateBody(m.editingID, body)

	m.mode = normalMode
	m.input.Blur()
	m.dateInput.Blur()
	m.bodyTextarea.Blur()
	m.input.SetValue("")
	m.dateInput.SetValue("")
	m.bodyTextarea.SetValue("")
	m.editField = 0

	// Clamp cursor — todo may have moved between sections
	newSelectable := selectableIndices(m.visibleItems())
	if m.cursor >= len(newSelectable) {
		m.cursor = max(0, len(newSelectable)-1)
	}
	return m, nil
}

// saveAdd persists a new todo from the 4-field add form and returns to normal mode.
func (m Model) saveAdd() (Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return m, nil
	}

	date := strings.TrimSpace(m.dateInput.Value())
	isoDate := ""
	if date != "" {
		var err error
		isoDate, err = config.ParseUserDate(date, m.dateLayout)
		if err != nil {
			// Invalid date -- focus date field
			m.editField = 1
			m.input.Blur()
			m.bodyTextarea.Blur()
			m.templateInput.Blur()
			return m, m.dateInput.Focus()
		}
	}

	todo := m.store.Add(text, isoDate)

	body := m.bodyTextarea.Value()
	if strings.TrimSpace(body) != "" {
		m.store.UpdateBody(todo.ID, body)
	}

	m.mode = normalMode
	m.input.Blur()
	m.dateInput.Blur()
	m.bodyTextarea.Blur()
	m.templateInput.Blur()
	m.input.SetValue("")
	m.dateInput.SetValue("")
	m.bodyTextarea.SetValue("")
	m.editField = 0
	return m, nil
}

// updateTemplateNameMode handles key events while entering a template name.
func (m Model) updateTemplateNameMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		name := strings.TrimSpace(m.input.Value())
		if name == "" {
			return m, nil
		}
		m.pendingTemplateName = name
		m.mode = templateContentMode
		m.input.Blur()
		m.templateTextarea.Reset()
		return m, m.templateTextarea.Focus()

	case key.Matches(msg, m.keys.Cancel):
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
		m.clearTemplateState()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// updateTemplateContentMode handles messages while entering template content (multi-line).
func (m Model) updateTemplateContentMode(msg tea.Msg) (Model, tea.Cmd) {
	saveKey := key.NewBinding(key.WithKeys("ctrl+d"))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, saveKey):
			content := strings.TrimSpace(m.templateTextarea.Value())
			if content == "" {
				return m, nil
			}
			m.store.AddTemplate(m.pendingTemplateName, content)
			m.mode = normalMode
			m.templateTextarea.Reset()
			m.clearTemplateState()
			return m, nil

		case key.Matches(msg, m.keys.Cancel):
			m.mode = normalMode
			m.templateTextarea.Reset()
			m.clearTemplateState()
			return m, nil
		}
	}

	// Forward all other messages to textarea (handles typing, cursor, blink, etc.)
	var cmd tea.Cmd
	m.templateTextarea, cmd = m.templateTextarea.Update(msg)
	return m, cmd
}

// View renders the todo list pane content.
func (m Model) View() string {
	switch m.mode {
	case inputMode, editMode, templateNameMode, templateContentMode:
		return m.editView()
	default:
		return m.normalView()
	}
}

// editView renders a vertically centered full-pane form for add/edit modes.
func (m Model) editView() string {
	var b strings.Builder

	// Heading
	var title string
	switch m.mode {
	case editMode:
		title = "Edit Todo"
	case templateNameMode:
		title = "New Template"
	case templateContentMode:
		title = "Template Content: " + m.pendingTemplateName
	default:
		title = "Add Todo"
	}
	b.WriteString(m.styles.EditTitle.Render(title))
	b.WriteString("\n\n")

	// Field(s)
	switch m.mode {
	case editMode:
		// Three fields: Title, Date, Body
		b.WriteString(m.styles.FieldLabel.Render("Title"))
		b.WriteString("\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
		b.WriteString(m.styles.FieldLabel.Render("Date"))
		b.WriteString("\n")
		b.WriteString(m.dateInput.View())
		b.WriteString("\n\n")
		b.WriteString(m.styles.FieldLabel.Render("Body"))
		b.WriteString("\n")
		b.WriteString(m.bodyTextarea.View())
		b.WriteString("\n")

	case templateContentMode:
		b.WriteString(m.templateTextarea.View())
		b.WriteString("\n")

	case templateNameMode:
		b.WriteString(m.styles.FieldLabel.Render("Name"))
		b.WriteString("\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")

	case inputMode:
		// Four fields: Title, Date, Body, Template
		b.WriteString(m.styles.FieldLabel.Render("Title"))
		b.WriteString("\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
		b.WriteString(m.styles.FieldLabel.Render("Date"))
		b.WriteString("\n")
		b.WriteString(m.dateInput.View())
		b.WriteString("\n\n")
		b.WriteString(m.styles.FieldLabel.Render("Body"))
		b.WriteString("\n")
		b.WriteString(m.bodyTextarea.View())
		b.WriteString("\n\n")
		b.WriteString(m.styles.FieldLabel.Render("Template"))
		b.WriteString("\n")
		b.WriteString(m.templateInput.View())
		b.WriteString("\n")
	}

	// Vertical centering (skip for modes with textareas)
	content := b.String()
	if m.height > 0 && m.mode != editMode && m.mode != inputMode && m.mode != templateContentMode {
		lines := strings.Count(content, "\n") + 1
		topPad := (m.height - lines) / 3
		if topPad > 0 {
			content = strings.Repeat("\n", topPad) + content
		}
	}

	return content
}

// normalView renders the standard todo list with headers, items, and inline controls.
func (m Model) normalView() string {
	items := m.visibleItems()
	selectable := selectableIndices(items)

	var b strings.Builder
	selectableIdx := 0

	for _, item := range items {
		switch item.kind {
		case headerItem:
			if b.Len() > 0 {
				b.WriteString("\n") // VIS-02: spacing before non-first headers
			}
			b.WriteString(m.styles.SectionHeader.Render(item.label))
			b.WriteString("\n")
			sep := strings.Repeat("─", len(item.label))
			b.WriteString(m.styles.Separator.Render(sep))
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

	// Show mode-specific UI below the todo list
	switch m.mode {
	case filterMode:
		b.WriteString("\n")
		b.WriteString(m.input.View())

	case normalMode:
		// No extra UI in normal mode
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

	// Styled checkbox (VIS-03)
	if t.Done {
		b.WriteString(m.styles.CheckboxDone.Render("[x]"))
	} else {
		b.WriteString(m.styles.Checkbox.Render("[ ]"))
	}
	b.WriteString(" ")

	// Text content -- styled separately from checkbox
	text := t.Text
	if t.Done {
		text = m.styles.Completed.Render(text)
	}
	b.WriteString(text)

	// Body indicator (after text, not affected by completed styling)
	if t.HasBody() {
		b.WriteString(" " + m.styles.BodyIndicator.Render("[+]"))
	}

	// Recurring indicator (after body indicator, before date)
	if t.ScheduleID > 0 {
		b.WriteString(" " + m.styles.RecurringIndicator.Render("[R]"))
	}

	// Date (after text, not affected by completed styling)
	if t.HasDate() {
		b.WriteString(" " + m.styles.Date.Render(config.FormatDate(t.Date, m.dateLayout)))
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

// clearTemplateState resets all template workflow fields.
func (m *Model) clearTemplateState() {
	m.pendingTemplateName = ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
