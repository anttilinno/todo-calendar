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
	"github.com/antti/todo-calendar/internal/fuzzy"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/antti/todo-calendar/internal/tmpl"
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
	normalMode mode = iota
	inputMode       // typing todo text
	editMode        // editing existing todo (title + date + body)
	filterMode      // inline filter narrowing visible todos
)

// itemKind classifies a visible row in the rendered list.
type itemKind int

const (
	headerItem itemKind = iota
	todoItem
	emptyItem
)

// sectionID identifies which section a visible item belongs to.
type sectionID int

const (
	sectionDated    sectionID = iota // dated todos (month or week)
	sectionMonth                     // "This Month" fuzzy-date todos
	sectionYear                      // "This Year" fuzzy-date todos
	sectionFloating                  // floating (undated) todos
)

// visibleItem is a single row in the combined todo list display.
type visibleItem struct {
	kind    itemKind
	label   string      // display text for headers/empty
	todo    *store.Todo // non-nil only for todoItem
	section sectionID   // which section this item belongs to
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
	bodyTextarea  textarea.Model  // textarea for body editing in edit mode
	editField     int             // 0 = title, 1 = date, 2 = body, 3 = template
	templateInput textinput.Model // placeholder input for template field (Phase 25 adds picker)

	// Segmented date input (replaces dateInput)
	dateSegDay   textinput.Model
	dateSegMonth textinput.Model
	dateSegYear  textinput.Model
	dateSegFocus int    // 0, 1, 2 = which segment is focused (left to right)
	dateSegOrder [3]int // maps visual position to semantic: 0=day, 1=month, 2=year
	dateFormat   string // "iso", "eu", or "us"

	// Week filter state (empty = no filter, set by app model when calendar is in weekly view)
	weekFilterStart string
	weekFilterEnd   string

	// Template picker sub-state (within inputMode)
	pickingTemplate         bool
	pickerTemplates         []store.Template
	pickerCursor            int
	promptingPlaceholders   bool
	pickerPlaceholderNames  []string
	pickerPlaceholderIndex  int
	pickerPlaceholderValues map[string]string
	pickerSelectedTemplate  *store.Template

}

// New creates a new todo list model backed by the given store.
func New(s store.TodoStore, t theme.Theme) Model {
	ti := textinput.New()
	ti.Placeholder = "What needs doing?"
	ti.CharLimit = 120
	ti.Prompt = "> "

	segDay := textinput.New()
	segDay.Placeholder = "dd"
	segDay.CharLimit = 2
	segDay.Width = 4
	segDay.Prompt = ""

	segMonth := textinput.New()
	segMonth.Placeholder = "mm"
	segMonth.CharLimit = 2
	segMonth.Width = 4
	segMonth.Prompt = ""

	segYear := textinput.New()
	segYear.Placeholder = "yyyy"
	segYear.CharLimit = 4
	segYear.Width = 6
	segYear.Prompt = ""

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
		dateSegDay:       segDay,
		dateSegMonth:     segMonth,
		dateSegYear:      segYear,
		dateSegOrder:     dateSegmentOrder("iso"),
		dateFormat:       "iso",
		bodyTextarea:     ba,
		templateInput:    tmplInput,
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

// SetWeekFilter sets the week date range filter for visibleItems.
// When active, dated todos are filtered to [startDate, endDate] instead of the full month.
func (m *Model) SetWeekFilter(startDate, endDate string) {
	m.weekFilterStart = startDate
	m.weekFilterEnd = endDate
	m.cursor = 0
	m.filterQuery = ""
	if m.mode == filterMode {
		m.mode = normalMode
		m.input.Blur()
		m.input.SetValue("")
	}
}

// ClearWeekFilter removes the week date range filter, reverting to full month display.
func (m *Model) ClearWeekFilter() {
	m.weekFilterStart = ""
	m.weekFilterEnd = ""
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
		if m.pickingTemplate {
			return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Confirm, m.keys.Cancel}
		}
		if m.promptingPlaceholders {
			return []key.Binding{m.keys.Confirm, m.keys.Cancel}
		}
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
	default:
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	}
}

// AllHelpBindings returns all key bindings for the expanded help view.
// In non-normal modes, still returns only Confirm/Cancel (HELP-02).
func (m Model) AllHelpBindings() []key.Binding {
	switch m.mode {
	case inputMode:
		if m.pickingTemplate {
			return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Confirm, m.keys.Cancel}
		}
		if m.promptingPlaceholders {
			return []key.Binding{m.keys.Confirm, m.keys.Cancel}
		}
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
			m.keys.Preview, m.keys.OpenEditor,
		}
	default:
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	}
}

// visibleItems builds the combined display list of headers, todos, and empty placeholders.
func (m Model) visibleItems() []visibleItem {
	var items []visibleItem

	// Section header and dated todos: week-filtered or full month
	if m.weekFilterStart != "" {
		// Week filter active: show "Week of {date}" header and date-range query
		startDate, err := time.Parse("2006-01-02", m.weekFilterStart)
		headerLabel := "Week of " + m.weekFilterStart
		if err == nil {
			headerLabel = fmt.Sprintf("Week of %s %d", startDate.Month().String(), startDate.Day())
		}
		items = append(items, visibleItem{kind: headerItem, label: headerLabel, section: sectionDated})

		dated := m.store.TodosForDateRange(m.weekFilterStart, m.weekFilterEnd)
		if len(dated) == 0 {
			items = append(items, visibleItem{kind: emptyItem, label: "(no todos this week)", section: sectionDated})
		} else {
			for i := range dated {
				items = append(items, visibleItem{kind: todoItem, todo: &dated[i], section: sectionDated})
			}
		}
	} else {
		// Month section header
		monthLabel := fmt.Sprintf("%s %d", m.viewMonth.String(), m.viewYear)
		items = append(items, visibleItem{kind: headerItem, label: monthLabel, section: sectionDated})

		// Dated todos for the viewed month
		dated := m.store.TodosForMonth(m.viewYear, m.viewMonth)
		if len(dated) == 0 {
			items = append(items, visibleItem{kind: emptyItem, label: "(no todos this month)", section: sectionDated})
		} else {
			for i := range dated {
				items = append(items, visibleItem{kind: todoItem, todo: &dated[i], section: sectionDated})
			}
		}
	}

	// This Month and This Year sections: only in monthly view (not weekly)
	if m.weekFilterStart == "" {
		// This Month section
		items = append(items, visibleItem{kind: headerItem, label: "This Month", section: sectionMonth})
		monthTodos := m.store.MonthTodos(m.viewYear, m.viewMonth)
		if len(monthTodos) == 0 {
			items = append(items, visibleItem{kind: emptyItem, label: "(no month todos)", section: sectionMonth})
		} else {
			for i := range monthTodos {
				items = append(items, visibleItem{kind: todoItem, todo: &monthTodos[i], section: sectionMonth})
			}
		}

		// This Year section
		items = append(items, visibleItem{kind: headerItem, label: "This Year", section: sectionYear})
		yearTodos := m.store.YearTodos(m.viewYear)
		if len(yearTodos) == 0 {
			items = append(items, visibleItem{kind: emptyItem, label: "(no year todos)", section: sectionYear})
		} else {
			for i := range yearTodos {
				items = append(items, visibleItem{kind: todoItem, todo: &yearTodos[i], section: sectionYear})
			}
		}
	}

	// Floating section header
	items = append(items, visibleItem{kind: headerItem, label: "Floating", section: sectionFloating})

	// Floating todos
	floating := m.store.FloatingTodos()
	if len(floating) == 0 {
		items = append(items, visibleItem{kind: emptyItem, label: "(no floating todos)", section: sectionFloating})
	} else {
		for i := range floating {
			items = append(items, visibleItem{kind: todoItem, todo: &floating[i], section: sectionFloating})
		}
	}

	// Apply inline filter when active
	if m.filterQuery != "" {
		var filtered []visibleItem
		for _, item := range items {
			switch item.kind {
			case headerItem:
				filtered = append(filtered, item)
			case todoItem:
				if matched, _ := fuzzy.Match(m.filterQuery, item.todo.Text); matched {
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
					result = append(result, visibleItem{kind: emptyItem, label: "(no matches)", section: item.section})
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
	// Forward blink/tick messages to focused text input in edit modes.
	switch m.mode {
	case inputMode, editMode:
		if _, ok := msg.(tea.KeyMsg); !ok {
			if _, ok := msg.(tea.WindowSizeMsg); !ok {
				var cmd tea.Cmd
				// During placeholder prompting, always forward to m.input
				if m.promptingPlaceholders {
					m.input, cmd = m.input.Update(msg)
					return m, cmd
				}
				switch m.editField {
				case 2:
					m.bodyTextarea, cmd = m.bodyTextarea.Update(msg)
				case 3:
					m.templateInput, cmd = m.templateInput.Update(msg)
				case 1:
					seg := m.dateSegmentByPos(m.dateSegFocus)
					*seg, cmd = seg.Update(msg)
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
			curItem := items[curIdx]
			prevItem := items[prevIdx]
			if curItem.todo != nil && prevItem.todo != nil &&
				curItem.section == prevItem.section {
				m.store.SwapOrder(curItem.todo.ID, prevItem.todo.ID)
				m.cursor--
			}
		}

	case key.Matches(msg, m.keys.MoveDown):
		if len(selectable) > 0 && m.cursor >= 0 && m.cursor < len(selectable)-1 {
			curIdx := selectable[m.cursor]
			nextIdx := selectable[m.cursor+1]
			curItem := items[curIdx]
			nextItem := items[nextIdx]
			if curItem.todo != nil && nextItem.todo != nil &&
				curItem.section == nextItem.section {
				m.store.SwapOrder(curItem.todo.ID, nextItem.todo.ID)
				m.cursor++
			}
		}

	case key.Matches(msg, m.keys.Add):
		m.mode = inputMode
		m.editField = 0
		m.input.Placeholder = "What needs doing?"
		m.input.Prompt = "> "
		m.input.SetValue("")
		m.clearAllDateSegments()
		m.blurAllDateSegments()
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
			// Populate date segments from existing todo
			m.clearAllDateSegments()
			m.blurAllDateSegments()
			if fresh.Date != "" {
				parts := strings.SplitN(fresh.Date, "-", 3)
				if len(parts) == 3 {
					switch fresh.DatePrecision {
					case "day":
						m.dateSegYear.SetValue(parts[0])
						m.dateSegMonth.SetValue(parts[1])
						m.dateSegDay.SetValue(parts[2])
					case "month":
						m.dateSegYear.SetValue(parts[0])
						m.dateSegMonth.SetValue(parts[1])
					case "year":
						m.dateSegYear.SetValue(parts[0])
					}
				}
			}
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
	// Template picker sub-states intercept all keys
	if m.pickingTemplate {
		return m.updateTemplatePicker(msg)
	}
	if m.promptingPlaceholders {
		return m.updatePlaceholderPrompting(msg)
	}

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
		// Template field: Enter opens template picker
		if m.editField == 3 {
			templates := m.store.ListTemplates()
			if len(templates) == 0 {
				return m, nil // No templates available
			}
			m.pickingTemplate = true
			m.pickerTemplates = templates
			m.pickerCursor = 0
			return m, nil
		}
		// Title/Date: Enter saves
		return m.saveAdd()

	case key.Matches(msg, m.keys.SwitchField):
		// Cycle: title(0) -> date segments(1) -> body(2) -> template(3) -> title(0)
		switch m.editField {
		case 0:
			m.editField = 1
			m.input.Blur()
			return m, m.focusDateSegment(0)
		case 1:
			if m.dateSegFocus < 2 {
				// Advance to next date segment
				return m, m.focusDateSegment(m.dateSegFocus + 1)
			}
			// Past last segment -> body
			m.editField = 2
			m.blurAllDateSegments()
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
			m.blurAllDateSegments()
			m.pickingTemplate = false
			m.promptingPlaceholders = false
			m.pickerSelectedTemplate = nil
			m.pickerTemplates = nil
			m.pickerPlaceholderNames = nil
			m.pickerPlaceholderValues = nil
			return m, m.input.Focus()
		}
		// Esc in title/date cancels entirely
		m.mode = normalMode
		m.input.Blur()
		m.blurAllDateSegments()
		m.bodyTextarea.Blur()
		m.templateInput.Blur()
		m.input.SetValue("")
		m.clearAllDateSegments()
		m.bodyTextarea.SetValue("")
		m.templateInput.SetValue("")
		m.pickingTemplate = false
		m.promptingPlaceholders = false
		m.pickerSelectedTemplate = nil
		m.pickerTemplates = nil
		m.pickerPlaceholderNames = nil
		m.pickerPlaceholderValues = nil
		m.editField = 0
		return m, nil
	}

	// Forward key events to the focused field
	var cmd tea.Cmd
	switch m.editField {
	case 0:
		m.input, cmd = m.input.Update(msg)
	case 1:
		return m.updateDateSegment(msg)
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
		// Cycle: title(0) → date segments(1) → body(2) → title(0)
		switch m.editField {
		case 0:
			m.editField = 1
			m.input.Blur()
			return m, m.focusDateSegment(0)
		case 1:
			if m.dateSegFocus < 2 {
				return m, m.focusDateSegment(m.dateSegFocus + 1)
			}
			m.editField = 2
			m.blurAllDateSegments()
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
		m.blurAllDateSegments()
		m.bodyTextarea.Blur()
		m.input.SetValue("")
		m.clearAllDateSegments()
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
		return m.updateDateSegment(msg)
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

	isoDate, precision, errPos := m.deriveDateFromSegments()
	if errPos >= 0 {
		// Invalid/incomplete date -- focus the problematic segment
		m.editField = 1
		m.input.Blur()
		m.bodyTextarea.Blur()
		return m, m.focusDateSegment(errPos)
	}

	body := m.bodyTextarea.Value()

	m.store.Update(m.editingID, text, isoDate, precision)
	m.store.UpdateBody(m.editingID, body)

	m.mode = normalMode
	m.input.Blur()
	m.blurAllDateSegments()
	m.bodyTextarea.Blur()
	m.input.SetValue("")
	m.clearAllDateSegments()
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

	isoDate, precision, errPos := m.deriveDateFromSegments()
	if errPos >= 0 {
		// Invalid/incomplete date -- focus the problematic segment
		m.editField = 1
		m.input.Blur()
		m.bodyTextarea.Blur()
		m.templateInput.Blur()
		return m, m.focusDateSegment(errPos)
	}

	todo := m.store.Add(text, isoDate, precision)

	body := m.bodyTextarea.Value()
	if strings.TrimSpace(body) != "" {
		m.store.UpdateBody(todo.ID, body)
	}

	m.mode = normalMode
	m.input.Blur()
	m.blurAllDateSegments()
	m.bodyTextarea.Blur()
	m.templateInput.Blur()
	m.input.SetValue("")
	m.clearAllDateSegments()
	m.bodyTextarea.SetValue("")
	m.templateInput.SetValue("")
	m.pickingTemplate = false
	m.promptingPlaceholders = false
	m.pickerSelectedTemplate = nil
	m.pickerTemplates = nil
	m.pickerPlaceholderNames = nil
	m.pickerPlaceholderValues = nil
	m.editField = 0
	return m, nil
}


// View renders the todo list pane content.
func (m Model) View() string {
	switch m.mode {
	case inputMode, editMode:
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
	default:
		title = "Add Todo"
	}
	b.WriteString(m.styles.EditTitle.Render(title))
	b.WriteString("\n\n")

	// Field(s)
	switch m.mode {
	case editMode:
		// Three fields: Title, Date (segmented), Body
		b.WriteString(m.styles.FieldLabel.Render("Title"))
		b.WriteString("\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
		b.WriteString(m.styles.FieldLabel.Render("Date"))
		b.WriteString("\n")
		b.WriteString(m.renderDateSegments())
		b.WriteString("\n")
		b.WriteString(m.styles.EditHint.Render("(leave day blank for month todo, leave day+month blank for year todo)"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.FieldLabel.Render("Body"))
		b.WriteString("\n")
		b.WriteString(m.bodyTextarea.View())
		b.WriteString("\n")

	case inputMode:
		if m.pickingTemplate {
			// Render "Select Template" heading + template list with cursor
			b.Reset()
			b.WriteString(m.styles.EditTitle.Render("Select Template"))
			b.WriteString("\n\n")
			for i, t := range m.pickerTemplates {
				if i == m.pickerCursor {
					b.WriteString(m.styles.Cursor.Render("> "))
				} else {
					b.WriteString("  ")
				}
				b.WriteString(t.Name)
				// Brief inline preview (40 chars, single line)
				preview := t.Content
				if len(preview) > 40 {
					preview = preview[:40] + "..."
				}
				preview = strings.ReplaceAll(preview, "\n", " ")
				b.WriteString("  " + m.styles.Empty.Render(preview))
				b.WriteString("\n")
			}
		} else if m.promptingPlaceholders {
			// Render placeholder prompt heading + input field
			b.Reset()
			pTitle := fmt.Sprintf("Fill Placeholder (%d/%d)",
				m.pickerPlaceholderIndex+1, len(m.pickerPlaceholderNames))
			b.WriteString(m.styles.EditTitle.Render(pTitle))
			b.WriteString("\n\n")
			b.WriteString(m.styles.FieldLabel.Render(
				m.pickerPlaceholderNames[m.pickerPlaceholderIndex]))
			b.WriteString("\n")
			b.WriteString(m.input.View())
			b.WriteString("\n")
		} else {
			// Normal 4-field form: Title, Date (segmented), Body, Template
			b.WriteString(m.styles.FieldLabel.Render("Title"))
			b.WriteString("\n")
			b.WriteString(m.input.View())
			b.WriteString("\n\n")
			b.WriteString(m.styles.FieldLabel.Render("Date"))
			b.WriteString("\n")
			b.WriteString(m.renderDateSegments())
			b.WriteString("\n")
			b.WriteString(m.styles.EditHint.Render("(leave day blank for month todo, leave day+month blank for year todo)"))
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
	}

	// Vertical centering (skip for modes with textareas)
	content := b.String()
	if m.height > 0 && m.mode != editMode && m.mode != inputMode {
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
		b.WriteString(" " + m.styles.Date.Render(renderFuzzyDate(t, m.dateLayout)))
	}

	b.WriteString("\n")
}

// SetTheme replaces the todolist styles with ones built from the given theme.
// This preserves all model state (cursor, mode, input).
func (m *Model) SetTheme(t theme.Theme) {
	m.styles = NewStyles(t)
}

// SetDateFormat updates the date display layout, input placeholder, and segment ordering.
func (m *Model) SetDateFormat(format, layout, placeholder string) {
	m.dateFormat = format
	m.dateLayout = layout
	m.datePlaceholder = placeholder
	m.dateSegOrder = dateSegmentOrder(format)
}

// updateTemplatePicker handles key events in the template picker sub-state.
func (m Model) updateTemplatePicker(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.pickerCursor < len(m.pickerTemplates)-1 {
			m.pickerCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		selected := m.pickerTemplates[m.pickerCursor]
		m.pickerSelectedTemplate = &selected
		names, err := tmpl.ExtractPlaceholders(selected.Content)
		if err != nil || len(names) == 0 {
			// No placeholders -- render and pre-fill immediately
			body, _ := tmpl.ExecuteTemplate(selected.Content, map[string]string{})
			return m.prefillFromTemplate(&selected, body), m.input.Focus()
		}
		// Has placeholders -- enter prompting sub-state
		m.promptingPlaceholders = true
		m.pickingTemplate = false
		m.pickerPlaceholderNames = names
		m.pickerPlaceholderIndex = 0
		m.pickerPlaceholderValues = make(map[string]string)
		m.input.SetValue("")
		m.input.Placeholder = names[0]
		m.input.Prompt = names[0] + ": "
		return m, m.input.Focus()

	case key.Matches(msg, m.keys.Cancel):
		m.pickingTemplate = false
		m.pickerTemplates = nil
		m.pickerCursor = 0
		// Return to Template field (editField=3)
		return m, m.templateInput.Focus()
	}
	return m, nil
}

// updatePlaceholderPrompting handles key events while prompting for template placeholder values.
func (m Model) updatePlaceholderPrompting(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		value := strings.TrimSpace(m.input.Value())
		m.pickerPlaceholderValues[m.pickerPlaceholderNames[m.pickerPlaceholderIndex]] = value
		m.pickerPlaceholderIndex++
		if m.pickerPlaceholderIndex < len(m.pickerPlaceholderNames) {
			// More placeholders remain
			name := m.pickerPlaceholderNames[m.pickerPlaceholderIndex]
			m.input.Placeholder = name
			m.input.Prompt = name + ": "
			m.input.SetValue("")
			return m, nil
		}
		// All placeholders filled -- render and pre-fill
		body, _ := tmpl.ExecuteTemplate(
			m.pickerSelectedTemplate.Content,
			m.pickerPlaceholderValues,
		)
		return m.prefillFromTemplate(m.pickerSelectedTemplate, body), m.input.Focus()

	case key.Matches(msg, m.keys.Cancel):
		// Go back to template picker
		m.promptingPlaceholders = false
		m.pickingTemplate = true
		m.input.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// prefillFromTemplate sets form fields from a selected template and returns to the title field.
func (m Model) prefillFromTemplate(t *store.Template, renderedBody string) Model {
	m.pickingTemplate = false
	m.promptingPlaceholders = false
	m.input.SetValue(t.Name)
	m.input.Placeholder = "What needs doing?"
	m.input.Prompt = "> "
	m.input.CursorEnd()
	m.bodyTextarea.SetValue(renderedBody)
	m.templateInput.SetValue(t.Name)
	m.editField = 0
	// Clear picker state
	m.pickerTemplates = nil
	m.pickerCursor = 0
	m.pickerSelectedTemplate = nil
	m.pickerPlaceholderNames = nil
	m.pickerPlaceholderIndex = 0
	m.pickerPlaceholderValues = nil
	return m
}

// renderDateSegments renders the three date segment inputs in format-aware order with separators.
func (m Model) renderDateSegments() string {
	sep := m.styles.DateSeparator.Render(" " + m.dateSegSeparator() + " ")
	var parts []string
	for i := 0; i < 3; i++ {
		seg := m.dateSegmentByPos(i)
		parts = append(parts, seg.View())
	}
	return parts[0] + sep + parts[1] + sep + parts[2]
}

// renderFuzzyDate formats a todo's date for display, respecting its precision level.
// Day-precision: formatted per user's date format. Month-precision: "March 2026". Year-precision: "2026".
func renderFuzzyDate(t *store.Todo, dateLayout string) string {
	if t.Date == "" {
		return ""
	}
	switch t.DatePrecision {
	case "year":
		parsed, err := time.Parse("2006-01-02", t.Date)
		if err != nil {
			return t.Date
		}
		return fmt.Sprintf("%d", parsed.Year())
	case "month":
		parsed, err := time.Parse("2006-01-02", t.Date)
		if err != nil {
			return t.Date
		}
		return fmt.Sprintf("%s %d", parsed.Month().String(), parsed.Year())
	default:
		return config.FormatDate(t.Date, dateLayout)
	}
}

// deriveDateFromSegments reads the three date segment values and derives the ISO date
// string and date precision. Returns (isoDate, precision, errSegPos) where errSegPos >= 0
// indicates which visual segment needs attention (-1 means success).
func (m Model) deriveDateFromSegments() (string, string, int) {
	day := strings.TrimSpace(m.dateSegDay.Value())
	month := strings.TrimSpace(m.dateSegMonth.Value())
	year := strings.TrimSpace(m.dateSegYear.Value())

	// All empty: floating todo
	if year == "" && month == "" && day == "" {
		return "", "", -1
	}

	// Year is required for any dated todo
	if year == "" {
		// Find the visual position of the year segment
		for i := 0; i < 3; i++ {
			if m.dateSegOrder[i] == 2 {
				return "", "", i
			}
		}
		return "", "", 0
	}

	// Validate year is 4 digits
	if len(year) != 4 {
		for i := 0; i < 3; i++ {
			if m.dateSegOrder[i] == 2 {
				return "", "", i
			}
		}
	}
	for _, c := range year {
		if c < '0' || c > '9' {
			for i := 0; i < 3; i++ {
				if m.dateSegOrder[i] == 2 {
					return "", "", i
				}
			}
		}
	}

	// Year only: year precision
	if month == "" && day == "" {
		return year + "-01-01", "year", -1
	}

	// Day filled but no month: invalid
	if month == "" && day != "" {
		for i := 0; i < 3; i++ {
			if m.dateSegOrder[i] == 1 {
				return "", "", i
			}
		}
	}

	// Validate month
	monthNum := 0
	for _, c := range month {
		if c < '0' || c > '9' {
			for i := 0; i < 3; i++ {
				if m.dateSegOrder[i] == 1 {
					return "", "", i
				}
			}
		}
		monthNum = monthNum*10 + int(c-'0')
	}
	if monthNum < 1 || monthNum > 12 {
		for i := 0; i < 3; i++ {
			if m.dateSegOrder[i] == 1 {
				return "", "", i
			}
		}
	}
	paddedMonth := fmt.Sprintf("%02d", monthNum)

	// Year + month only: month precision
	if day == "" {
		return year + "-" + paddedMonth + "-01", "month", -1
	}

	// All three filled: day precision - validate as real date
	dayNum := 0
	for _, c := range day {
		if c < '0' || c > '9' {
			for i := 0; i < 3; i++ {
				if m.dateSegOrder[i] == 0 {
					return "", "", i
				}
			}
		}
		dayNum = dayNum*10 + int(c-'0')
	}
	paddedDay := fmt.Sprintf("%02d", dayNum)
	isoDate := year + "-" + paddedMonth + "-" + paddedDay

	// Validate the full date
	_, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		for i := 0; i < 3; i++ {
			if m.dateSegOrder[i] == 0 {
				return "", "", i
			}
		}
	}

	return isoDate, "day", -1
}

// updateDateSegment handles key events forwarded to the focused date segment.
// It intercepts separator chars, handles auto-advance on full segment, and backspace navigation.
func (m Model) updateDateSegment(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Block separator characters (handled visually)
	if key == "-" || key == "." || key == "/" {
		return m, nil
	}

	seg := m.dateSegmentByPos(m.dateSegFocus)
	prevLen := len(seg.Value())

	// Backspace on empty segment: move back to previous segment
	if key == "backspace" && seg.Value() == "" && m.dateSegFocus > 0 {
		return m, m.focusDateSegment(m.dateSegFocus - 1)
	}

	// Forward the key to the focused segment
	var cmd tea.Cmd
	*seg, cmd = seg.Update(msg)

	// Auto-advance: if segment just reached its char limit, move to next
	newLen := len(seg.Value())
	limit := m.dateSegCharLimit(m.dateSegFocus)
	if newLen >= limit && newLen > prevLen && m.dateSegFocus < 2 {
		cmd2 := m.focusDateSegment(m.dateSegFocus + 1)
		return m, tea.Batch(cmd, cmd2)
	}

	return m, cmd
}

// dateSegmentOrder returns the visual-to-semantic mapping for date segments.
// Semantic: 0=day, 1=month, 2=year. The returned array maps visual positions (left-to-right) to semantic meaning.
func dateSegmentOrder(format string) [3]int {
	switch format {
	case "eu":
		return [3]int{0, 1, 2} // dd mm yyyy
	case "us":
		return [3]int{1, 0, 2} // mm dd yyyy
	default: // iso
		return [3]int{2, 1, 0} // yyyy mm dd
	}
}

// dateSegmentByPos returns the textinput for a visual position using dateSegOrder.
func (m *Model) dateSegmentByPos(pos int) *textinput.Model {
	switch m.dateSegOrder[pos] {
	case 0:
		return &m.dateSegDay
	case 1:
		return &m.dateSegMonth
	case 2:
		return &m.dateSegYear
	default:
		return &m.dateSegDay
	}
}

// dateSegSeparator returns the separator character for the configured date format.
func (m *Model) dateSegSeparator() string {
	switch m.dateFormat {
	case "eu":
		return "."
	case "us":
		return "/"
	default:
		return "-"
	}
}

// dateSegPlaceholderByPos returns the placeholder text for a visual position.
func (m *Model) dateSegPlaceholderByPos(pos int) string {
	switch m.dateSegOrder[pos] {
	case 0:
		return "dd"
	case 1:
		return "mm"
	case 2:
		return "yyyy"
	default:
		return "dd"
	}
}

// focusDateSegment focuses the segment at the given visual position and blurs all others.
func (m *Model) focusDateSegment(pos int) tea.Cmd {
	m.dateSegDay.Blur()
	m.dateSegMonth.Blur()
	m.dateSegYear.Blur()
	m.dateSegFocus = pos
	return m.dateSegmentByPos(pos).Focus()
}

// blurAllDateSegments blurs all three date segments.
func (m *Model) blurAllDateSegments() {
	m.dateSegDay.Blur()
	m.dateSegMonth.Blur()
	m.dateSegYear.Blur()
}

// clearAllDateSegments clears all three date segments and resets focus.
func (m *Model) clearAllDateSegments() {
	m.dateSegDay.SetValue("")
	m.dateSegMonth.SetValue("")
	m.dateSegYear.SetValue("")
	m.dateSegFocus = 0
}

// dateSegCharLimit returns the character limit for the segment at the given visual position.
func (m *Model) dateSegCharLimit(pos int) int {
	switch m.dateSegOrder[pos] {
	case 2:
		return 4 // year
	default:
		return 2 // day or month
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
