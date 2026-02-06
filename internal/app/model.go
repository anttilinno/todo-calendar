package app

import (
	"github.com/antti/todo-calendar/internal/calendar"
	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/search"
	"github.com/antti/todo-calendar/internal/settings"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
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
	calendar     calendar.Model
	todoList     todolist.Model
	activePane   pane
	width        int
	height       int
	ready        bool
	keys         KeyMap
	help         help.Model
	styles       Styles
	showSettings bool
	settings     settings.Model
	showSearch   bool
	search       search.Model
	store        store.TodoStore
	cfg          config.Config
	savedConfig  config.Config
}

// New creates a new root application model with the given dependencies.
func New(provider *holidays.Provider, mondayStart bool, s store.TodoStore, t theme.Theme, cfg config.Config) Model {
	cal := calendar.New(provider, mondayStart, s, t)
	cal.SetFocused(true)

	tl := todolist.New(s, t)
	tl.SetDateFormat(cfg.DateLayout(), cfg.DatePlaceholder())
	tl.SetViewMonth(cal.Year(), cal.Month())

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(t.AccentFg)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)

	return Model{
		calendar:   cal,
		todoList:   tl,
		activePane: calendarPane,
		keys:       DefaultKeyMap(),
		help:       h,
		styles:     NewStyles(t),
		store:      s,
		cfg:        cfg,
	}
}

// Init returns the initial command for the root model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the root model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle settings-specific messages regardless of showSettings state,
	// because they arrive as commands from the settings model in the next
	// Update cycle.
	switch msg := msg.(type) {
	case settings.ThemeChangedMsg:
		m.applyTheme(msg.Theme)
		return m, nil

	case settings.SaveMsg:
		m.showSettings = false
		oldCountry := m.cfg.Country
		m.cfg = msg.Cfg
		_ = config.Save(m.cfg)
		// Apply the saved theme (may differ from live preview if user cycled
		// theme multiple times before saving).
		m.applyTheme(theme.ForName(msg.Cfg.Theme))
		// Rebuild provider if country changed
		if msg.Cfg.Country != oldCountry {
			if p, err := holidays.NewProvider(msg.Cfg.Country); err == nil {
				m.calendar.SetProvider(p)
			}
		}
		m.calendar.SetMondayStart(msg.Cfg.MondayStart())
		m.todoList.SetDateFormat(m.cfg.DateLayout(), m.cfg.DatePlaceholder())
		m.calendar.RefreshIndicators()
		return m, nil

	case settings.CancelMsg:
		m.showSettings = false
		m.cfg = m.savedConfig
		m.applyTheme(theme.ForName(m.savedConfig.Theme))
		return m, nil

	case search.JumpMsg:
		m.showSearch = false
		m.calendar.SetYearMonth(msg.Year, msg.Month)
		m.todoList.SetViewMonth(msg.Year, msg.Month)
		m.calendar.RefreshIndicators()
		return m, nil

	case search.CloseMsg:
		m.showSearch = false
		return m, nil
	}

	// When settings overlay is open, route most messages there.
	if m.showSettings {
		return m.updateSettings(msg)
	}

	// When search overlay is open, route most messages there.
	if m.showSearch {
		return m.updateSearch(msg)
	}

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
			m.calendar.RefreshIndicators()
			return m, nil
		case key.Matches(msg, m.keys.Settings) && !isInputting:
			m.savedConfig = m.cfg
			m.settings = settings.New(m.cfg, theme.ForName(m.cfg.Theme))
			m.settings.SetSize(m.width, m.height)
			m.showSettings = true
			return m, nil
		case key.Matches(msg, m.keys.Search) && !isInputting:
			m.search = search.New(m.store, theme.ForName(m.cfg.Theme), m.cfg)
			m.search.SetSize(m.width, m.height)
			m.showSearch = true
			return m, m.search.Init()
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

	// Refresh calendar indicators after every update cycle so that
	// todo mutations (add/toggle/delete) are reflected immediately.
	m.calendar.RefreshIndicators()

	return m, cmd
}

// updateSettings routes messages to the settings model when the overlay is open.
func (m Model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Also propagate window resize to all children so the app resizes correctly.
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.ready = true
		m.help.Width = wsm.Width
		m.settings.SetSize(wsm.Width, wsm.Height)

		var calCmd, todoCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)
		m.todoList, todoCmd = m.todoList.Update(msg)
		return m, tea.Batch(calCmd, todoCmd)
	}

	var cmd tea.Cmd
	m.settings, cmd = m.settings.Update(msg)
	return m, cmd
}

// updateSearch routes messages to the search model when the overlay is open.
func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Also propagate window resize to all children so the app resizes correctly.
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.ready = true
		m.help.Width = wsm.Width
		m.search.SetSize(wsm.Width, wsm.Height)

		var calCmd, todoCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)
		m.todoList, todoCmd = m.todoList.Update(msg)
		return m, tea.Batch(calCmd, todoCmd)
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	return m, cmd
}

// applyTheme updates all component styles with the given theme.
func (m *Model) applyTheme(t theme.Theme) {
	m.styles = NewStyles(t)
	m.calendar.SetTheme(t)
	m.todoList.SetTheme(t)
	m.settings.SetTheme(t)
	m.search.SetTheme(t)
	m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(t.AccentFg)
	m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
	m.help.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)
}

// currentHelpKeys returns an aggregated help KeyMap based on the active pane.
func (m Model) currentHelpKeys() helpKeyMap {
	if m.showSearch {
		return helpKeyMap{bindings: m.search.HelpBindings()}
	}
	if m.showSettings {
		return helpKeyMap{bindings: m.settings.HelpBindings()}
	}

	var bindings []key.Binding

	switch m.activePane {
	case calendarPane:
		calKeys := m.calendar.Keys()
		bindings = append(bindings, calKeys.PrevMonth, calKeys.NextMonth, calKeys.ToggleWeek)
	case todoPane:
		bindings = append(bindings, m.todoList.HelpBindings()...)
	}

	bindings = append(bindings, m.keys.Tab, m.keys.Settings, m.keys.Search, m.keys.Quit)
	return helpKeyMap{bindings: bindings}
}

// View renders the root model.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.showSettings {
		m.help.Width = m.width
		helpBar := m.help.View(m.currentHelpKeys())
		return lipgloss.JoinVertical(lipgloss.Left, m.settings.View(), helpBar)
	}

	if m.showSearch {
		m.help.Width = m.width
		helpBar := m.help.View(m.currentHelpKeys())
		return lipgloss.JoinVertical(lipgloss.Left, m.search.View(), helpBar)
	}

	// Calculate frame overhead from pane style
	frameH, frameV := m.styles.Pane(true).GetFrameSize()

	helpHeight := 1
	contentHeight := m.height - helpHeight - frameV
	if contentHeight < 1 {
		contentHeight = 1
	}

	calendarInnerWidth := 38
	todoInnerWidth := m.width - calendarInnerWidth - (frameH * 2)

	// Guard against impossibly narrow terminals
	if todoInnerWidth < 1 {
		return "Terminal too small"
	}

	calStyle := m.styles.Pane(m.activePane == calendarPane).
		Width(calendarInnerWidth).
		Height(contentHeight)

	todoStyle := m.styles.Pane(m.activePane == todoPane).
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
