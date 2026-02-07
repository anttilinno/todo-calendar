package tmplmgr

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

// viewMode tracks the current interaction mode of the overlay.
type viewMode int

const (
	listMode   viewMode = iota
	renameMode
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
}

// New creates a new template management overlay model.
func New(s store.TodoStore, t theme.Theme) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.CharLimit = 80

	m := Model{
		store:  s,
		keys:   DefaultKeyMap(),
		styles: NewStyles(t),
		input:  ti,
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
	if m.mode == renameMode {
		return []key.Binding{m.keys.Confirm, m.keys.Cancel}
	}
	return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Delete, m.keys.Rename, m.keys.Edit, m.keys.Cancel}
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
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// updateListMode handles key messages in list mode.
func (m Model) updateListMode(msg tea.KeyMsg) (Model, tea.Cmd) {
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

// View renders the template management overlay.
func (m Model) View() string {
	var b strings.Builder

	title := m.styles.Title.Render("Templates")
	b.WriteString(title)
	b.WriteString("\n\n")

	if len(m.templates) == 0 {
		b.WriteString(m.styles.Empty.Render("(no templates)"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.Hint.Render("  esc close"))
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
		if i == m.cursor {
			if m.mode == renameMode {
				b.WriteString("> ")
				b.WriteString(m.input.View())
			} else {
				line := "> " + t.Name
				b.WriteString(m.styles.SelectedName.Render(line))
			}
		} else {
			line := "  " + t.Name
			b.WriteString(m.styles.TemplateName.Render(line))
		}
		b.WriteString("\n")
	}

	// Error message for rename.
	if m.err != "" {
		b.WriteString(m.styles.Error.Render("  " + m.err))
		b.WriteString("\n")
	}

	// Separator.
	sep := strings.Repeat("-", 40)
	b.WriteString("\n")
	b.WriteString(m.styles.Separator.Render(sep))
	b.WriteString("\n\n")

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

	// Hint bar.
	b.WriteString("\n\n")
	if m.mode == renameMode {
		b.WriteString(m.styles.Hint.Render("  enter confirm  |  esc cancel"))
	} else {
		b.WriteString(m.styles.Hint.Render("  j/k navigate  |  r rename  |  d delete  |  e edit  |  esc close"))
	}

	return m.verticalCenter(b.String())
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
