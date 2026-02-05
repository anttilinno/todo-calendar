package app

import (
	"github.com/antti/todo-calendar/internal/calendar"
	"github.com/antti/todo-calendar/internal/todolist"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// pane identifies which pane is active.
type pane int

const (
	calendarPane pane = iota
	todoPane
)

// Model is the root application model.
type Model struct {
	calendar   calendar.Model
	todoList   todolist.Model
	activePane pane
	width      int
	height     int
	ready      bool
	keys       KeyMap
}

// New creates a new root application model.
func New() Model {
	cal := calendar.New()
	cal.SetFocused(true)

	return Model{
		calendar:   cal,
		todoList:   todolist.New(),
		activePane: calendarPane,
		keys:       DefaultKeyMap(),
	}
}

// Init returns the initial command for the root model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the root model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			if m.activePane == calendarPane {
				m.activePane = todoPane
			} else {
				m.activePane = calendarPane
			}
			m.calendar.SetFocused(m.activePane == calendarPane)
			m.todoList.SetFocused(m.activePane == todoPane)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Broadcast to all children
		var calCmd, todoCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)
		m.todoList, todoCmd = m.todoList.Update(msg)
		return m, tea.Batch(calCmd, todoCmd)
	}

	// Route to focused child only
	var cmd tea.Cmd
	switch m.activePane {
	case calendarPane:
		m.calendar, cmd = m.calendar.Update(msg)
	case todoPane:
		m.todoList, cmd = m.todoList.Update(msg)
	}

	return m, cmd
}

// View renders the root model.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Calculate frame overhead from pane style
	frameH, frameV := paneStyle(true).GetFrameSize()

	helpHeight := 1
	contentHeight := m.height - helpHeight - frameV
	if contentHeight < 1 {
		contentHeight = 1
	}

	calendarInnerWidth := 24
	todoInnerWidth := m.width - calendarInnerWidth - (frameH * 2)

	// Guard against impossibly narrow terminals
	if todoInnerWidth < 1 {
		return "Terminal too small"
	}

	calStyle := paneStyle(m.activePane == calendarPane).
		Width(calendarInnerWidth).
		Height(contentHeight)

	todoStyle := paneStyle(m.activePane == todoPane).
		Width(todoInnerWidth).
		Height(contentHeight)

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		calStyle.Render(m.calendar.View()),
		todoStyle.Render(m.todoList.View()),
	)

	statusBar := "q: quit | Tab: switch pane"

	return lipgloss.JoinVertical(lipgloss.Left, top, statusBar)
}
