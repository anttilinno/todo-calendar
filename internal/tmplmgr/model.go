package tmplmgr

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/antti/todo-calendar/internal/recurring"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/antti/todo-calendar/internal/tmpl"
)

// viewMode tracks the current interaction mode of the overlay.
type viewMode int

const (
	listMode              viewMode = iota
	renameMode
	scheduleMode
	placeholderDefaultsMode
)

// CloseMsg is emitted when the user presses Esc to close the overlay.
type CloseMsg struct{}

// EditTemplateMsg is emitted when the user presses 'e' to edit a template
// in an external editor. The app layer handles the editor launch.
type EditTemplateMsg struct{ Template store.Template }

// TemplateUpdatedMsg is emitted after rename or delete so the app can
// refresh any cached template data if needed.
type TemplateUpdatedMsg struct{}

// Model represents the template management overlay.
type Model struct {
	templates []store.Template
	cursor    int
	mode      viewMode
	store     store.TodoStore
	width     int
	height    int
	keys      KeyMap
	styles    Styles
	input     textinput.Model
	err       string

	// Schedule picker state
	cadenceTypes    []string
	cadenceIndex    int
	weeklyDays      [7]bool
	weekdayCursor   int
	monthlyInput    textinput.Model
	editingSchedule *store.Schedule

	// Placeholder defaults state
	pendingCadenceType  string
	pendingCadenceValue string
	placeholderNames    []string
	placeholderIndex    int
	placeholderValues   map[string]string
	defaultsInput       textinput.Model
}

// New creates a new template management overlay model.
func New(s store.TodoStore, t theme.Theme) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.CharLimit = 80

	mi := textinput.New()
	mi.Prompt = "> "
	mi.CharLimit = 2
	mi.Placeholder = "1-31"

	di := textinput.New()
	di.Prompt = "> "
	di.CharLimit = 200

	m := Model{
		store:         s,
		keys:          DefaultKeyMap(),
		styles:        NewStyles(t),
		input:         ti,
		cadenceTypes:  []string{"none", "daily", "weekdays", "weekly", "monthly"},
		monthlyInput:  mi,
		defaultsInput: di,
	}
	m.RefreshTemplates()
	return m
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

// HelpBindings returns overlay-specific key bindings for help bar display.
func (m Model) HelpBindings() []key.Binding {
	switch m.mode {
	case renameMode:
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	case scheduleMode:
		bindings := []key.Binding{m.keys.Left, m.keys.Right, m.keys.Confirm, m.keys.Cancel}
		if m.cadenceTypes[m.cadenceIndex] == "weekly" {
			bindings = []key.Binding{m.keys.Left, m.keys.Right, m.keys.Up, m.keys.Down, m.keys.Toggle, m.keys.Confirm, m.keys.Cancel}
		}
		return bindings
	case placeholderDefaultsMode:
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	default:
		return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Delete, m.keys.Rename, m.keys.Edit, m.keys.Schedule, m.keys.Cancel}
	}
}

// SetError sets an error message to display in the overlay.
func (m *Model) SetError(msg string) {
	m.err = msg
}

// RefreshTemplates reloads templates from the store.
func (m *Model) RefreshTemplates() {
	m.templates = m.store.ListTemplates()
	if m.cursor >= len(m.templates) {
		m.cursor = len(m.templates) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// selected returns the currently selected template, or nil if none.
func (m Model) selected() *store.Template {
	if len(m.templates) == 0 || m.cursor < 0 || m.cursor >= len(m.templates) {
		return nil
	}
	t := m.templates[m.cursor]
	return &t
}

// Update handles messages for the template management overlay.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case listMode:
			return m.updateListMode(msg)
		case renameMode:
			return m.updateRenameMode(msg)
		case scheduleMode:
			return m.updateScheduleMode(msg)
		case placeholderDefaultsMode:
			return m.updatePlaceholderDefaultsMode(msg)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// updateListMode handles key messages in list mode.
func (m Model) updateListMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	m.err = ""
	switch {
	case key.Matches(msg, m.keys.Cancel):
		return m, func() tea.Msg { return CloseMsg{} }

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.templates)-1 {
			m.cursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		if sel := m.selected(); sel != nil {
			m.store.DeleteTemplate(sel.ID)
			m.RefreshTemplates()
			return m, func() tea.Msg { return TemplateUpdatedMsg{} }
		}
		return m, nil

	case key.Matches(msg, m.keys.Rename):
		if sel := m.selected(); sel != nil {
			m.mode = renameMode
			m.input.SetValue(sel.Name)
			m.input.Focus()
			m.input.CursorEnd()
			m.err = ""
		}
		return m, nil

	case key.Matches(msg, m.keys.Edit):
		if sel := m.selected(); sel != nil {
			t := *sel
			return m, func() tea.Msg { return EditTemplateMsg{Template: t} }
		}
		return m, nil

	case key.Matches(msg, m.keys.Schedule):
		if sel := m.selected(); sel != nil {
			m.mode = scheduleMode
			m.cadenceIndex = 0
			m.weeklyDays = [7]bool{}
			m.weekdayCursor = 0
			m.monthlyInput.SetValue("")
			m.monthlyInput.Blur()
			m.editingSchedule = nil
			m.err = ""

			// Load existing schedule if any.
			scheds := m.store.ListSchedulesForTemplate(sel.ID)
			if len(scheds) > 0 {
				sched := scheds[0]
				m.editingSchedule = &sched

				ruleStr := sched.CadenceType
				if sched.CadenceValue != "" {
					ruleStr += ":" + sched.CadenceValue
				}
				rule, err := recurring.ParseRule(ruleStr)
				if err == nil {
					switch rule.Type {
					case "daily":
						m.cadenceIndex = 1
					case "weekdays":
						m.cadenceIndex = 2
					case "weekly":
						m.cadenceIndex = 3
						dayIndex := map[string]int{
							"mon": 0, "tue": 1, "wed": 2, "thu": 3,
							"fri": 4, "sat": 5, "sun": 6,
						}
						for _, d := range rule.Days {
							if idx, ok := dayIndex[d]; ok {
								m.weeklyDays[idx] = true
							}
						}
					case "monthly":
						m.cadenceIndex = 4
						m.monthlyInput.SetValue(strconv.Itoa(rule.DayOfMonth))
						m.monthlyInput.Focus()
					}
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// updateRenameMode handles key messages in rename mode.
func (m Model) updateRenameMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = listMode
		m.err = ""
		m.input.Blur()
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		sel := m.selected()
		if sel == nil {
			m.mode = listMode
			m.input.Blur()
			return m, nil
		}

		newName := strings.TrimSpace(m.input.Value())
		if newName == "" || newName == sel.Name {
			m.mode = listMode
			m.err = ""
			m.input.Blur()
			return m, nil
		}

		err := m.store.UpdateTemplate(sel.ID, newName, sel.Content)
		if err != nil {
			m.err = "Name already exists"
			return m, nil
		}

		m.mode = listMode
		m.err = ""
		m.input.Blur()
		m.RefreshTemplates()
		return m, func() tea.Msg { return TemplateUpdatedMsg{} }
	}

	// Forward other keys to the text input.
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// dayNames maps index 0-6 to lowercase day abbreviations (Mon=0, Sun=6).
var dayNames = [7]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

// updateScheduleMode handles key messages in schedule picker mode.
func (m Model) updateScheduleMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = listMode
		m.err = ""
		m.monthlyInput.Blur()
		return m, nil

	case key.Matches(msg, m.keys.Left):
		wasMonthly := m.cadenceTypes[m.cadenceIndex] == "monthly"
		m.cadenceIndex = (m.cadenceIndex - 1 + len(m.cadenceTypes)) % len(m.cadenceTypes)
		if wasMonthly {
			m.monthlyInput.Blur()
		}
		if m.cadenceTypes[m.cadenceIndex] == "monthly" {
			m.monthlyInput.Focus()
		}
		return m, nil

	case key.Matches(msg, m.keys.Right):
		wasMonthly := m.cadenceTypes[m.cadenceIndex] == "monthly"
		m.cadenceIndex = (m.cadenceIndex + 1) % len(m.cadenceTypes)
		if wasMonthly {
			m.monthlyInput.Blur()
		}
		if m.cadenceTypes[m.cadenceIndex] == "monthly" {
			m.monthlyInput.Focus()
		}
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.cadenceTypes[m.cadenceIndex] == "weekly" && m.weekdayCursor > 0 {
			m.weekdayCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.cadenceTypes[m.cadenceIndex] == "weekly" && m.weekdayCursor < 6 {
			m.weekdayCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Toggle):
		if m.cadenceTypes[m.cadenceIndex] == "weekly" {
			m.weeklyDays[m.weekdayCursor] = !m.weeklyDays[m.weekdayCursor]
		}
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		sel := m.selected()
		if sel == nil {
			m.mode = listMode
			m.monthlyInput.Blur()
			return m, nil
		}

		var cadenceType, cadenceValue string

		switch m.cadenceTypes[m.cadenceIndex] {
		case "none":
			if m.editingSchedule != nil {
				m.store.DeleteSchedule(m.editingSchedule.ID)
			}
			m.mode = listMode
			m.err = ""
			m.monthlyInput.Blur()
			m.RefreshTemplates()
			return m, func() tea.Msg { return TemplateUpdatedMsg{} }

		case "daily":
			cadenceType = "daily"
			cadenceValue = ""

		case "weekdays":
			cadenceType = "weekdays"
			cadenceValue = ""

		case "weekly":
			var selected []string
			for i := 0; i < 7; i++ {
				if m.weeklyDays[i] {
					selected = append(selected, dayNames[i])
				}
			}
			if len(selected) == 0 {
				m.err = "Select at least one day"
				return m, nil
			}
			cadenceType = "weekly"
			cadenceValue = strings.Join(selected, ",")

		case "monthly":
			dayStr := strings.TrimSpace(m.monthlyInput.Value())
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 31 {
				m.err = "Enter a day 1-31"
				return m, nil
			}
			cadenceType = "monthly"
			cadenceValue = dayStr
		}

		// Check if template has placeholders that need defaults.
		placeholders, pErr := tmpl.ExtractPlaceholders(sel.Content)
		if pErr == nil && len(placeholders) > 0 {
			m.pendingCadenceType = cadenceType
			m.pendingCadenceValue = cadenceValue
			m.placeholderNames = placeholders
			m.placeholderIndex = 0
			m.placeholderValues = make(map[string]string)
			if m.editingSchedule != nil && m.editingSchedule.PlaceholderDefaults != "" {
				json.Unmarshal([]byte(m.editingSchedule.PlaceholderDefaults), &m.placeholderValues)
			}
			m.defaultsInput.SetValue(m.placeholderValues[placeholders[0]])
			m.defaultsInput.Focus()
			m.defaultsInput.CursorEnd()
			m.monthlyInput.Blur()
			m.mode = placeholderDefaultsMode
			return m, nil
		}

		defaults := "{}"
		if m.editingSchedule != nil {
			if m.editingSchedule.PlaceholderDefaults != "" {
				defaults = m.editingSchedule.PlaceholderDefaults
			}
			m.store.UpdateSchedule(m.editingSchedule.ID, cadenceType, cadenceValue, defaults)
		} else {
			m.store.AddSchedule(sel.ID, cadenceType, cadenceValue, defaults)
		}
		m.mode = listMode
		m.err = ""
		m.monthlyInput.Blur()
		m.RefreshTemplates()
		return m, func() tea.Msg { return TemplateUpdatedMsg{} }
	}

	// Forward to monthlyInput when in monthly mode.
	if m.cadenceTypes[m.cadenceIndex] == "monthly" {
		var cmd tea.Cmd
		m.monthlyInput, cmd = m.monthlyInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// updatePlaceholderDefaultsMode handles key messages in placeholder defaults mode.
func (m Model) updatePlaceholderDefaultsMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.mode = listMode
		m.err = ""
		m.defaultsInput.Blur()
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		m.placeholderValues[m.placeholderNames[m.placeholderIndex]] = m.defaultsInput.Value()

		if m.placeholderIndex < len(m.placeholderNames)-1 {
			m.placeholderIndex++
			m.defaultsInput.SetValue(m.placeholderValues[m.placeholderNames[m.placeholderIndex]])
			m.defaultsInput.CursorEnd()
			return m, nil
		}

		// Last placeholder -- save the schedule.
		defaultsJSON, _ := json.Marshal(m.placeholderValues)
		defaults := string(defaultsJSON)

		sel := m.selected()
		if m.editingSchedule != nil {
			m.store.UpdateSchedule(m.editingSchedule.ID, m.pendingCadenceType, m.pendingCadenceValue, defaults)
		} else if sel != nil {
			m.store.AddSchedule(sel.ID, m.pendingCadenceType, m.pendingCadenceValue, defaults)
		}
		m.mode = listMode
		m.err = ""
		m.defaultsInput.Blur()
		m.RefreshTemplates()
		return m, func() tea.Msg { return TemplateUpdatedMsg{} }
	}

	var cmd tea.Cmd
	m.defaultsInput, cmd = m.defaultsInput.Update(msg)
	return m, cmd
}

// View renders the template management overlay.
func (m Model) View() string {
	var b strings.Builder

	title := m.styles.Title.Render("Templates")
	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.templates) == 0 {
		b.WriteString(m.styles.Empty.Render("(no templates)"))
		return m.verticalCenter(b.String())
	}

	// Template list.
	maxListItems := m.height - 12
	if maxListItems < 1 {
		maxListItems = 1
	}
	visible := len(m.templates)
	if visible > maxListItems {
		visible = maxListItems
	}

	for i := 0; i < visible; i++ {
		t := m.templates[i]
		suffix := m.scheduleLabel(t.ID)
		styledSuffix := ""
		if suffix != "" {
			styledSuffix = " " + m.styles.ScheduleSuffix.Render(suffix)
		}

		if i == m.cursor {
			if m.mode == renameMode {
				b.WriteString("> ")
				b.WriteString(m.input.View())
			} else {
				line := "> " + t.Name
				b.WriteString(m.styles.SelectedName.Render(line) + styledSuffix)
			}
		} else {
			line := "  " + t.Name
			b.WriteString(m.styles.TemplateName.Render(line) + styledSuffix)
		}
		b.WriteString("\n")
	}

	// Error message for list/rename mode (schedule errors are shown inside the picker).
	if m.err != "" && (m.mode == listMode || m.mode == renameMode) {
		b.WriteString(m.styles.Error.Render("  " + m.err))
		b.WriteString("\n")
	}

	// Separator.
	sep := strings.Repeat("-", 40)
	b.WriteString("\n")
	b.WriteString(m.styles.Separator.Render(sep))
	b.WriteString("\n\n")

	if m.mode == scheduleMode {
		// Schedule picker.
		b.WriteString(m.renderSchedulePicker())
	} else if m.mode == placeholderDefaultsMode {
		// Placeholder defaults prompting.
		prompt := fmt.Sprintf("Set default for %q (%d/%d):",
			m.placeholderNames[m.placeholderIndex],
			m.placeholderIndex+1,
			len(m.placeholderNames))
		b.WriteString(m.styles.SchedulePrompt.Render(prompt))
		b.WriteString("\n")
		b.WriteString(m.defaultsInput.View())
	} else {
		// Content preview of selected template (raw text).
		if sel := m.selected(); sel != nil {
			maxContentLines := m.height - visible - 10
			if maxContentLines < 1 {
				maxContentLines = 1
			}
			lines := strings.Split(sel.Content, "\n")
			if len(lines) > maxContentLines {
				lines = lines[:maxContentLines]
			}
			content := strings.Join(lines, "\n")
			b.WriteString(m.styles.Content.Render(content))
		}
	}

	return m.verticalCenter(b.String())
}

// renderSchedulePicker renders the schedule picker UI below the separator.
func (m Model) renderSchedulePicker() string {
	var b strings.Builder

	// Cadence type bar: < None  Daily  Weekdays  Weekly  Monthly >
	displayNames := [5]string{"None", "Daily", "Weekdays", "Weekly", "Monthly"}
	b.WriteString("Schedule: < ")
	for i, name := range displayNames {
		if i == m.cadenceIndex {
			b.WriteString(m.styles.ScheduleActive.Render(name))
		} else {
			b.WriteString(m.styles.ScheduleInactive.Render(name))
		}
		if i < len(displayNames)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString(" >")
	b.WriteString("\n")

	cadence := m.cadenceTypes[m.cadenceIndex]

	switch cadence {
	case "weekly":
		b.WriteString("\n")
		weekdayLabels := [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
		for i := 0; i < 7; i++ {
			cursor := "  "
			if i == m.weekdayCursor {
				cursor = "> "
			}
			check := "[ ]"
			if m.weeklyDays[i] {
				check = "[x]"
			}
			var label string
			if m.weeklyDays[i] {
				label = m.styles.ScheduleDaySelected.Render(weekdayLabels[i])
			} else {
				label = m.styles.ScheduleDay.Render(weekdayLabels[i])
			}
			b.WriteString(fmt.Sprintf("  %s%s  %s\n", cursor, label, check))
		}

	case "monthly":
		b.WriteString("\n")
		b.WriteString("  Day of month: ")
		b.WriteString(m.monthlyInput.View())
		b.WriteString("\n")
	}

	// Error message.
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.Error.Render("  " + m.err))
	}

	return b.String()
}

// scheduleLabel returns a display suffix for the schedule attached to the
// given template, e.g. "(daily)", "(Mon/Wed/Fri)", "(15th of month)".
// Returns "" if no schedule is attached.
func (m Model) scheduleLabel(templateID int) string {
	schedules := m.store.ListSchedulesForTemplate(templateID)
	if len(schedules) == 0 {
		return ""
	}
	sched := schedules[0]

	// Build rule string: cadenceType alone or cadenceType:cadenceValue.
	ruleStr := sched.CadenceType
	if sched.CadenceValue != "" {
		ruleStr += ":" + sched.CadenceValue
	}

	rule, err := recurring.ParseRule(ruleStr)
	if err != nil {
		// Fallback: show raw cadence type.
		return "(" + sched.CadenceType + ")"
	}

	switch rule.Type {
	case "daily":
		return "(daily)"
	case "weekdays":
		return "(weekdays)"
	case "weekly":
		dayLabels := make([]string, len(rule.Days))
		for i, d := range rule.Days {
			// Capitalize first letter: "mon" -> "Mon"
			dayLabels[i] = strings.ToUpper(d[:1]) + d[1:]
		}
		return "(" + strings.Join(dayLabels, "/") + ")"
	case "monthly":
		return fmt.Sprintf("(%d%s of month)", rule.DayOfMonth, ordinalSuffix(rule.DayOfMonth))
	default:
		return "(" + sched.CadenceType + ")"
	}
}

// ordinalSuffix returns the English ordinal suffix for a number (st, nd, rd, th).
func ordinalSuffix(n int) string {
	if n >= 11 && n <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

// verticalCenter centers the content vertically within the available height.
func (m Model) verticalCenter(content string) string {
	if m.height > 0 {
		lines := strings.Count(content, "\n") + 1
		topPad := (m.height - lines) / 2
		if topPad > 0 {
			content = strings.Repeat("\n", topPad) + content
		}
	}
	return content
}
