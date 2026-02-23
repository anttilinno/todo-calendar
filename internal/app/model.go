package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/antti/todo-calendar/internal/calendar"
	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/editor"
	"github.com/antti/todo-calendar/internal/google"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/preview"
	"github.com/antti/todo-calendar/internal/search"
	"github.com/antti/todo-calendar/internal/settings"
	"github.com/antti/todo-calendar/internal/status"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/antti/todo-calendar/internal/tmplmgr"
	"github.com/antti/todo-calendar/internal/todolist"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gcal "google.golang.org/api/calendar/v3"
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

func (h helpKeyMap) ShortHelp() []key.Binding { return h.bindings }
func (h helpKeyMap) FullHelp() [][]key.Binding {
	if len(h.bindings) <= 5 {
		return [][]key.Binding{h.bindings}
	}
	var groups [][]key.Binding
	for i := 0; i < len(h.bindings); i += 5 {
		end := i + 5
		if end > len(h.bindings) {
			end = len(h.bindings)
		}
		groups = append(groups, h.bindings[i:end])
	}
	return groups
}

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
	showPreview   bool
	preview       preview.Model
	showTmplMgr   bool
	tmplMgr       tmplmgr.Model
	editing         bool
	editingTmplID   int
	editorErr       string
	store           store.TodoStore
	theme           theme.Theme
	cfg             config.Config
	googleAuthState google.AuthState
	calendarSvc     *gcal.Service
	calendarEvents  []google.CalendarEvent
	eventsSyncToken string
	eventsFetchErr  error
}

// New creates a new root application model with the given dependencies.
func New(provider *holidays.Provider, mondayStart bool, s store.TodoStore, t theme.Theme, cfg config.Config, authState google.AuthState, calSvc *gcal.Service) Model {
	cal := calendar.New(provider, mondayStart, s, t)
	cal.SetFocused(true)
	cal.SetShowFuzzySections(cfg.ShowMonthTodos, cfg.ShowYearTodos)

	tl := todolist.New(s, t)
	tl.SetDateFormat(cfg.DateFormat, cfg.DateLayout(), cfg.DatePlaceholder())
	tl.SetShowFuzzySections(cfg.ShowMonthTodos, cfg.ShowYearTodos)
	tl.SetPriorityStyle(cfg.PriorityStyle)
	tl.SetViewMonth(cal.Year(), cal.Month())

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(t.AccentFg)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(t.AccentFg)
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)

	return Model{
		calendar:        cal,
		todoList:        tl,
		activePane:      calendarPane,
		keys:            DefaultKeyMap(),
		help:            h,
		styles:          NewStyles(t),
		store:           s,
		theme:           t,
		cfg:             cfg,
		googleAuthState: authState,
		calendarSvc:     calSvc,
	}
}

// Init returns the initial command for the root model.
func (m Model) Init() tea.Cmd {
	m.refreshStatusFile()
	if m.googleAuthState == google.AuthNotConfigured {
		return nil
	}
	var cmds []tea.Cmd
	if m.calendarSvc != nil {
		cmds = append(cmds, google.FetchEventsCmd(m.calendarSvc, ""))
	}
	cmds = append(cmds, google.ScheduleEventTick())
	return tea.Batch(cmds...)
}

// Update handles messages for the root model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle settings-specific messages regardless of showSettings state,
	// because they arrive as commands from the settings model in the next
	// Update cycle.
	switch msg := msg.(type) {
	case settings.SettingChangedMsg:
		oldCountry := m.cfg.Country
		m.cfg = msg.Cfg
		_ = config.Save(m.cfg)
		m.applyTheme(theme.ForName(msg.Cfg.Theme))
		if msg.Cfg.Country != oldCountry {
			if p, err := holidays.NewProvider(msg.Cfg.Country); err == nil {
				m.calendar.SetProvider(p)
			}
		}
		m.calendar.SetMondayStart(msg.Cfg.MondayStart())
		m.todoList.SetDateFormat(m.cfg.DateFormat, m.cfg.DateLayout(), m.cfg.DatePlaceholder())
		m.todoList.SetShowFuzzySections(msg.Cfg.ShowMonthTodos, msg.Cfg.ShowYearTodos)
		m.todoList.SetPriorityStyle(msg.Cfg.PriorityStyle)
		m.calendar.SetShowFuzzySections(msg.Cfg.ShowMonthTodos, msg.Cfg.ShowYearTodos)
		if m.cfg.GoogleCalendarEnabled {
			m.todoList.SetCalendarEvents(m.calendarEvents)
			m.calendar.SetCalendarEvents(m.calendarEvents)
		} else {
			m.todoList.SetCalendarEvents(nil)
			m.calendar.SetCalendarEvents(nil)
		}
		m.calendar.RefreshIndicators()
		m.refreshStatusFile()
		return m, nil

	case settings.CloseMsg:
		m.showSettings = false
		return m, nil

	case google.EventsFetchedMsg:
		if msg.Err != nil {
			m.eventsFetchErr = msg.Err
			// Keep last known calendarEvents intact, schedule retry
			return m, google.ScheduleEventTick()
		}
		m.eventsFetchErr = nil
		if m.eventsSyncToken == "" {
			// Full sync: replace all events
			m.calendarEvents = msg.Events
		} else {
			// Incremental sync: merge changes
			m.calendarEvents = google.MergeEvents(m.calendarEvents, msg.Events)
		}
		m.eventsSyncToken = msg.SyncToken
		if m.cfg.GoogleCalendarEnabled {
			m.todoList.SetCalendarEvents(m.calendarEvents)
			m.calendar.SetCalendarEvents(m.calendarEvents)
		} else {
			m.todoList.SetCalendarEvents(nil)
			m.calendar.SetCalendarEvents(nil)
		}
		return m, google.ScheduleEventTick()

	case google.EventTickMsg:
		if m.calendarSvc == nil || m.googleAuthState != google.AuthReady {
			return m, google.ScheduleEventTick()
		}
		return m, google.FetchEventsCmd(m.calendarSvc, m.eventsSyncToken)

	case google.AuthResultMsg:
		if msg.Success {
			m.googleAuthState = google.AuthReady
			m.settings.SetGoogleAuthState(google.AuthReady)
			// Create calendar service and trigger first fetch
			if m.calendarSvc == nil {
				if svc, err := google.NewCalendarService(); err == nil {
					m.calendarSvc = svc
					return m, tea.Batch(
						google.FetchEventsCmd(svc, ""),
						google.ScheduleEventTick(),
					)
				}
			}
		} else {
			m.googleAuthState = google.AuthNeedsLogin
			m.settings.SetGoogleAuthState(google.AuthNeedsLogin)
		}
		return m, nil

	case settings.StartGoogleAuthMsg:
		return m, google.StartAuthFlow()

	case search.JumpMsg:
		m.showSearch = false
		m.calendar.SetYearMonth(msg.Year, msg.Month)
		m.todoList.SetViewMonth(msg.Year, msg.Month)
		m.todoList.ClearWeekFilter()
		m.calendar.RefreshIndicators()
		return m, nil

	case search.CloseMsg:
		m.showSearch = false
		return m, nil

	case preview.CloseMsg:
		m.showPreview = false
		return m, nil

	case tmplmgr.CloseMsg:
		m.showTmplMgr = false
		return m, nil

	case tmplmgr.EditTemplateMsg:
		m.editing = true
		m.editingTmplID = msg.Template.ID
		tpl := msg.Template
		return m, editorOpenTemplateContent(tpl.Content)

	case tmplmgr.TemplateUpdatedMsg:
		return m, nil

	case todolist.PreviewMsg:
		m.preview = preview.New(msg.Todo.Text, msg.Todo.Body, m.cfg.Theme, theme.ForName(m.cfg.Theme), m.width, m.height)
		m.showPreview = true
		return m, nil

	case todolist.OpenEditorMsg:
		m.editing = true
		todo := msg.Todo
		return m, editor.Open(todo.ID, todo.Text, todo.Body)

	case editor.EditorFinishedMsg:
		m.editing = false
		if m.editingTmplID != 0 {
			templateID := m.editingTmplID
			m.editingTmplID = 0
			defer os.Remove(msg.TempPath)
			if msg.Err != nil {
				m.tmplMgr.SetError(fmt.Sprintf("Could not open editor: %v", msg.Err))
				return m, nil
			}
			data, err := os.ReadFile(msg.TempPath)
			if err != nil {
				return m, nil
			}
			newContent := strings.TrimRight(string(data), " \t\n")
			if newContent != msg.OriginalBody {
				tpl := m.store.FindTemplate(templateID)
				if tpl != nil {
					m.store.UpdateTemplate(templateID, tpl.Name, newContent)
				}
			}
			m.tmplMgr.RefreshTemplates()
			return m, nil
		}
		newBody, changed, err := editor.ReadResult(msg)
		// Always clean up temp file after reading to prevent /tmp accumulation.
		os.Remove(msg.TempPath)
		if err != nil {
			m.editorErr = fmt.Sprintf("Could not open editor: %v", err)
			return m, nil
		}
		if changed {
			m.store.UpdateBody(msg.TodoID, newBody)
		}
		m.calendar.RefreshIndicators()
		m.refreshStatusFile()
		return m, nil
	}

	// When settings overlay is open, route most messages there.
	if m.showSettings {
		return m.updateSettings(msg)
	}

	// When template manager overlay is open, route most messages there.
	if m.showTmplMgr {
		return m.updateTmplMgr(msg)
	}

	// When preview overlay is open, route most messages there.
	if m.showPreview {
		return m.updatePreview(msg)
	}

	// When search overlay is open, route most messages there.
	if m.showSearch {
		return m.updateSearch(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.editorErr = ""
		// In input mode, only ctrl+c quits (let 'q' go to textinput)
		isInputting := m.activePane == todoPane && m.todoList.IsInputting()

		switch {
		case key.Matches(msg, m.keys.Quit) && !isInputting:
			if m.help.ShowAll {
				m.help.ShowAll = false
				return m, nil
			}
			return m, tea.Quit
		case m.help.ShowAll && msg.String() == "esc":
			m.help.ShowAll = false
			return m, nil
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
			m.syncTodoView()
			m.calendar.RefreshIndicators()
			m.help.ShowAll = false
			return m, nil
		case key.Matches(msg, m.keys.Settings) && !isInputting:
			m.settings = settings.New(m.cfg, theme.ForName(m.cfg.Theme), m.googleAuthState)
			m.settings.SetSize(m.width, m.height)
			m.showSettings = true
			return m, nil
		case key.Matches(msg, m.keys.Search) && !isInputting:
			m.search = search.New(m.store, theme.ForName(m.cfg.Theme), m.cfg)
			m.search.SetSize(m.width, m.height)
			m.showSearch = true
			return m, m.search.Init()
		case key.Matches(msg, m.keys.Templates) && !isInputting:
			m.tmplMgr = tmplmgr.New(m.store, theme.ForName(m.cfg.Theme))
			m.tmplMgr.SetSize(m.width, m.height)
			m.showTmplMgr = true
			return m, nil
		case key.Matches(msg, m.keys.Help) && !isInputting:
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.help.Width = msg.Width

		// Sync todo view on first ready
		m.syncTodoView()

		// Broadcast to calendar
		var calCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)

		// Set pane dimensions on todolist explicitly
		m.syncTodoSize()

		return m, calCmd
	}

	// Route to focused child only
	var cmd tea.Cmd
	switch m.activePane {
	case calendarPane:
		m.calendar, cmd = m.calendar.Update(msg)
		// Sync todo list when calendar view changes (month or week navigation)
		m.syncTodoView()
	case todoPane:
		m.todoList, cmd = m.todoList.Update(msg)
	}

	// Refresh calendar indicators after every update cycle so that
	// todo mutations (add/toggle/delete) are reflected immediately.
	m.calendar.RefreshIndicators()
	m.refreshStatusFile()

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

		var calCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)
		m.syncTodoSize()
		return m, calCmd
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

		var calCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)
		m.syncTodoSize()
		return m, calCmd
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	return m, cmd
}

// updatePreview routes messages to the preview model when the overlay is open.
func (m Model) updatePreview(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Also propagate window resize to all children so the app resizes correctly.
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.ready = true
		m.help.Width = wsm.Width

		var calCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)
		m.syncTodoSize()

		var prevCmd tea.Cmd
		m.preview, prevCmd = m.preview.Update(msg)
		return m, tea.Batch(calCmd, prevCmd)
	}

	var cmd tea.Cmd
	m.preview, cmd = m.preview.Update(msg)
	return m, cmd
}

// updateTmplMgr routes messages to the tmplmgr model when the overlay is open.
func (m Model) updateTmplMgr(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Also propagate window resize to all children so the app resizes correctly.
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.ready = true
		m.help.Width = wsm.Width
		m.tmplMgr.SetSize(wsm.Width, wsm.Height)

		var calCmd tea.Cmd
		m.calendar, calCmd = m.calendar.Update(msg)
		m.syncTodoSize()
		return m, calCmd
	}

	var cmd tea.Cmd
	m.tmplMgr, cmd = m.tmplMgr.Update(msg)
	return m, cmd
}

// editorOpenTemplateContent opens the user's editor with the template content.
// Unlike editor.Open, it writes raw content without a # heading, since templates
// are raw content with placeholders, not todo bodies.
func editorOpenTemplateContent(content string) tea.Cmd {
	f, err := os.CreateTemp("", "todo-calendar-template-*.md")
	if err != nil {
		return func() tea.Msg {
			return editor.EditorFinishedMsg{Err: err}
		}
	}

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return func() tea.Msg {
			return editor.EditorFinishedMsg{Err: err}
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return func() tea.Msg {
			return editor.EditorFinishedMsg{Err: err}
		}
	}

	tempPath := f.Name()
	parts := strings.Fields(editor.ResolveEditor())
	args := append(parts[1:], tempPath)
	cmd := exec.Command(parts[0], args...)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editor.EditorFinishedMsg{
			TempPath:     tempPath,
			OriginalBody: content,
			Err:          err,
		}
	})
}

// CalendarEvents returns the current Google Calendar events for use by other components.
func (m Model) CalendarEvents() []google.CalendarEvent {
	return m.calendarEvents
}

// syncTodoView sets the todolist view month and conditionally applies/clears
// the week filter based on the calendar's current view mode.
func (m *Model) syncTodoView() {
	m.todoList.SetViewMonth(m.calendar.Year(), m.calendar.Month())
	if m.cfg.GoogleCalendarEnabled {
		m.todoList.SetCalendarEvents(m.calendarEvents)
		m.calendar.SetCalendarEvents(m.calendarEvents)
	} else {
		m.todoList.SetCalendarEvents(nil)
		m.calendar.SetCalendarEvents(nil)
	}
	if m.calendar.GetViewMode() == calendar.WeekView {
		ws := m.calendar.WeekStart()
		we := ws.AddDate(0, 0, 6)
		m.todoList.SetWeekFilter(
			ws.Format("2006-01-02"),
			we.Format("2006-01-02"),
		)
	} else {
		m.todoList.ClearWeekFilter()
	}
}

// syncTodoSize computes the pane dimensions and passes them to the todolist model.
func (m *Model) syncTodoSize() {
	frameH, frameV := m.styles.Pane(true).GetFrameSize()

	// Render help to measure its height
	helpBar := m.help.View(m.currentHelpKeys())
	helpHeight := lipgloss.Height(helpBar)
	if helpHeight < 1 {
		helpHeight = 1
	}

	contentHeight := m.height - helpHeight - frameV
	if contentHeight < 1 {
		contentHeight = 1
	}

	calendarInnerWidth := 38
	todoInnerWidth := m.width - calendarInnerWidth - (frameH * 2)
	if todoInnerWidth < 1 {
		todoInnerWidth = 1
	}

	m.todoList.SetSize(todoInnerWidth, contentHeight)

	// Calendar content width = pane width minus horizontal padding (not border).
	paneStyle := m.styles.Pane(true)
	hPad := paneStyle.GetPaddingLeft() + paneStyle.GetPaddingRight()
	m.calendar.SetContentWidth(calendarInnerWidth - hPad)
}

// applyTheme updates all component styles with the given theme.
func (m *Model) applyTheme(t theme.Theme) {
	m.theme = t
	m.styles = NewStyles(t)
	m.calendar.SetTheme(t)
	m.todoList.SetTheme(t)
	m.settings.SetTheme(t)
	m.search.SetTheme(t)
	m.preview.SetTheme(t)
	m.tmplMgr.SetTheme(t)
	m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(t.AccentFg)
	m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
	m.help.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)
	m.help.Styles.FullKey = lipgloss.NewStyle().Foreground(t.AccentFg)
	m.help.Styles.FullDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
	m.help.Styles.FullSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)
}

// refreshStatusFile writes the current Polybar status to the state file.
// It queries today's todos, formats via status.FormatStatus, and writes via
// status.WriteStatusFile. Errors are silently ignored — the status file is a
// best-effort side effect.
func (m Model) refreshStatusFile() {
	today := time.Now().Format("2006-01-02")
	todos := m.store.TodosForDateRange(today, today)
	output := status.FormatStatus(todos, m.theme)
	_ = status.WriteStatusFile(output)
}

// currentHelpKeys returns an aggregated help KeyMap based on the active pane.
func (m Model) currentHelpKeys() helpKeyMap {
	if m.showPreview {
		return helpKeyMap{bindings: m.preview.HelpBindings()}
	}
	if m.showTmplMgr {
		return helpKeyMap{bindings: m.tmplMgr.HelpBindings()}
	}
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
		if m.help.ShowAll {
			bindings = append(bindings, m.todoList.AllHelpBindings()...)
		} else {
			bindings = append(bindings, m.todoList.HelpBindings()...)
		}
	}

	if m.help.ShowAll {
		bindings = append(bindings, m.keys.Tab, m.keys.Settings, m.keys.Search, m.keys.Templates, m.keys.Quit)
	}
	// Show ? in help bar except during input modes (HELP-02)
	if !m.todoList.IsInputting() || m.activePane == calendarPane {
		bindings = append(bindings, m.keys.Help)
	}
	return helpKeyMap{bindings: bindings}
}

// View renders the root model.
func (m Model) View() string {
	// When an external editor is running, return empty string to prevent
	// Bubble Tea from leaking TUI content to terminal scrollback during
	// alt-screen teardown.
	if m.editing {
		return ""
	}

	if !m.ready {
		return "Initializing..."
	}

	if m.showSettings {
		m.help.Width = m.width
		helpBar := m.help.View(m.currentHelpKeys())
		return lipgloss.JoinVertical(lipgloss.Left, m.settings.View(), helpBar)
	}

	if m.showPreview {
		m.help.Width = m.width
		helpBar := m.help.View(m.currentHelpKeys())
		return lipgloss.JoinVertical(lipgloss.Left, m.preview.View(), helpBar)
	}

	if m.showTmplMgr {
		m.help.Width = m.width
		helpBar := m.help.View(m.currentHelpKeys())
		return lipgloss.JoinVertical(lipgloss.Left, m.tmplMgr.View(), helpBar)
	}

	if m.showSearch {
		m.help.Width = m.width
		helpBar := m.help.View(m.currentHelpKeys())
		return lipgloss.JoinVertical(lipgloss.Left, m.search.View(), helpBar)
	}

	// Calculate help bar first so we can measure its height
	m.help.Width = m.width
	helpBar := m.help.View(m.currentHelpKeys())
	helpHeight := lipgloss.Height(helpBar)
	if helpHeight < 1 {
		helpHeight = 1
	}

	// Calculate frame overhead from pane style
	frameH, frameV := m.styles.Pane(true).GetFrameSize()
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
		Height(contentHeight).
		MaxHeight(contentHeight + frameV) // truncate overflow including border

	todoStyle := m.styles.Pane(m.activePane == todoPane).
		Width(todoInnerWidth).
		Height(contentHeight).
		MaxHeight(contentHeight + frameV) // truncate overflow including border

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		calStyle.Render(m.calendar.View()),
		todoStyle.Render(m.todoList.View()),
	)

	if m.editorErr != "" {
		errLine := m.styles.Error.Render(m.editorErr)
		return lipgloss.JoinVertical(lipgloss.Left, top, errLine, helpBar)
	}
	if m.eventsFetchErr != nil && m.cfg.GoogleCalendarEnabled {
		errMsg := m.eventsFetchErr.Error()
		if len(errMsg) > 80 {
			errMsg = errMsg[:80] + "..."
		}
		errLine := m.styles.Error.Render("Calendar: " + errMsg)
		return lipgloss.JoinVertical(lipgloss.Left, top, errLine, helpBar)
	}
	return lipgloss.JoinVertical(lipgloss.Left, top, helpBar)
}
