package app

import (
	"github.com/antti/todo-calendar/internal/calendar"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/todolist"
	"github.com/charmbracelet/bubbles/help"
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

// helpKeyMap adapts a slice of key.Binding to the help.KeyMap interface.
type helpKeyMap struct {
	bindings []key.Binding
}

func (h helpKeyMap) ShortHelp() []key.Binding  { return h.bindings }
func (h helpKeyMap) FullHelp() [][]key.Binding  { return [][]key.Binding{h.bindings} }

// Model is the root application model.
type Model struct {
	calendar   calendar.Model
	todoList   todolist.Model
	activePane pane
	width      int
	height     int
	ready      bool
	keys       KeyMap
	help       help.Model
}

// New creates a new root application model with the given dependencies.
func New(provider *holidays.Provider, mondayStart bool, s *store.Store) Model {
	cal := calendar.New(provider, mondayStart)
	cal.SetFocused(true)

	tl := todolist.New(s)
	tl.SetViewMonth(cal.Year(), cal.Month())

	return Model{
		calendar:   cal,
		todoList:   tl,
		activePane: calendarPane,
		keys:       DefaultKeyMap(),
		help:       help.New(),
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
		// In input mode, only ctrl+c quits (let 'q' go to textinput)
		isInputting := m.activePane == todoPane && m.todoList.IsInputting()

		switch {
		case key.Matches(msg, m.keys.Quit) && !isInputting:
			return m, tea.Quit
		case isInputting && msg.String() == "ctrl+c":
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab) && !isInputting:
			if m.activePane == calendarPane {
				m.activePane = todoPane
			} else {
				m.activePane = calendarPane
			}
			m.calendar.SetFocused(m.activePane == calendarPane)
			m.todoList.SetFocused(m.activePane == todoPane)
			m.todoList.SetViewMonth(m.calendar.Year(), m.calendar.Month())
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.help.Width = msg.Width

		// Sync todo view month on first ready
		m.todoList.SetViewMonth(m.calendar.Year(), m.calendar.Month())

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
		// Sync todo list when calendar month changes
		m.todoList.SetViewMonth(m.calendar.Year(), m.calendar.Month())
	case todoPane:
		m.todoList, cmd = m.todoList.Update(msg)
	}

	return m, cmd
}

// currentHelpKeys returns an aggregated help KeyMap based on the active pane.
func (m Model) currentHelpKeys() helpKeyMap {
	var bindings []key.Binding

	switch m.activePane {
	case calendarPane:
		calKeys := m.calendar.Keys()
		bindings = append(bindings, calKeys.PrevMonth, calKeys.NextMonth)
	case todoPane:
		bindings = append(bindings, m.todoList.HelpBindings()...)
	}

	bindings = append(bindings, m.keys.Tab, m.keys.Quit)
	return helpKeyMap{bindings: bindings}
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

	m.help.Width = m.width
	helpBar := m.help.View(m.currentHelpKeys())

	return lipgloss.JoinVertical(lipgloss.Left, top, helpBar)
}
