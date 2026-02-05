package todolist

import tea "github.com/charmbracelet/bubbletea"

// Model represents the todo list pane.
type Model struct {
	focused bool
	width   int
	height  int
}

// New creates a new todo list model.
func New() Model {
	return Model{}
}

// Update handles messages for the todo list pane.
// Returns concrete Model type, not tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the todo list pane content.
func (m Model) View() string {
	if m.focused {
		return "Todo List (focused)"
	}
	return "Todo List"
}

// SetFocused sets whether this pane is focused.
func (m *Model) SetFocused(f bool) {
	m.focused = f
}
