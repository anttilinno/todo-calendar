package calendar

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/antti/todo-calendar/internal/google"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

// ViewMode controls whether the calendar shows a full month or a single week.
type ViewMode int

const (
	// MonthView shows the full month grid (default).
	MonthView ViewMode = iota
	// WeekView shows a single 7-day week.
	WeekView
)

// weekStartFor returns the date of the first day of the week containing t.
// If mondayStart is true, weeks start on Monday; otherwise Sunday.
func weekStartFor(t time.Time, mondayStart bool) time.Time {
	wd := int(t.Weekday()) // Sunday=0 .. Saturday=6
	if mondayStart {
		offset := (wd + 6) % 7 // Monday=0 .. Sunday=6
		return time.Date(t.Year(), t.Month(), t.Day()-offset, 0, 0, 0, 0, time.Local)
	}
	return time.Date(t.Year(), t.Month(), t.Day()-wd, 0, 0, 0, 0, time.Local)
}

// Model represents the calendar pane.
type Model struct {
	focused     bool
	year        int
	month       time.Month
	today       time.Time
	holidays    map[int]bool
	indicators  map[int]int
	totals      map[int]int
	priorities  map[int]int
	provider    *holidays.Provider
	store       store.TodoStore
	keys        KeyMap
	mondayStart bool
	styles         Styles
	viewMode       ViewMode
	weekStart      time.Time
	showMonthTodos bool
	showYearTodos  bool
	contentWidth   int // pane text content width (pane width minus padding)
	calendarEvents []google.CalendarEvent
}

// New creates a new calendar model with the given holiday provider,
// week-start preference, and store for indicator data.
func New(provider *holidays.Provider, mondayStart bool, s store.TodoStore, t theme.Theme) Model {
	now := time.Now()
	y, m, _ := now.Date()

	return Model{
		year:           y,
		month:          m,
		today:          now,
		holidays:       provider.HolidaysInMonth(y, m),
		indicators:     s.IncompleteTodosPerDay(y, m),
		totals:         s.TotalTodosPerDay(y, m),
		priorities:     s.HighestPriorityPerDay(y, m),
		provider:       provider,
		store:          s,
		keys:           DefaultKeyMap(),
		mondayStart:    mondayStart,
		styles:         NewStyles(t),
		showMonthTodos: true,
		showYearTodos:  true,
	}
}

// Update handles messages for the calendar pane.
// Returns concrete Model type, not tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.ToggleWeek):
			if m.viewMode == MonthView {
				m.viewMode = WeekView
				m.weekStart = weekStartFor(time.Now(), m.mondayStart)
				m.year = m.weekStart.Year()
				m.month = m.weekStart.Month()
			} else {
				m.viewMode = MonthView
				// m.year and m.month already track the week's month
			}
			m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
			m.indicators = m.store.IncompleteTodosPerDay(m.year, m.month)
			m.totals = m.store.TotalTodosPerDay(m.year, m.month)
			m.priorities = m.store.HighestPriorityPerDay(m.year, m.month)

		case key.Matches(msg, m.keys.PrevMonth):
			if m.viewMode == WeekView {
				m.weekStart = m.weekStart.AddDate(0, 0, -7)
				m.year = m.weekStart.Year()
				m.month = m.weekStart.Month()
			} else {
				m.month--
				if m.month < time.January {
					m.month = time.December
					m.year--
				}
			}
			m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
			m.indicators = m.store.IncompleteTodosPerDay(m.year, m.month)
			m.totals = m.store.TotalTodosPerDay(m.year, m.month)
			m.priorities = m.store.HighestPriorityPerDay(m.year, m.month)

		case key.Matches(msg, m.keys.NextMonth):
			if m.viewMode == WeekView {
				m.weekStart = m.weekStart.AddDate(0, 0, 7)
				m.year = m.weekStart.Year()
				m.month = m.weekStart.Month()
			} else {
				m.month++
				if m.month > time.December {
					m.month = time.January
					m.year++
				}
			}
			m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
			m.indicators = m.store.IncompleteTodosPerDay(m.year, m.month)
			m.totals = m.store.TotalTodosPerDay(m.year, m.month)
			m.priorities = m.store.HighestPriorityPerDay(m.year, m.month)
		}

	}

	return m, nil
}

// View renders the calendar pane content including the overview section.
func (m Model) View() string {
	hasEvents := m.hasEventsPerDay(m.year, m.month)
	var content string
	if m.viewMode == WeekView {
		grid := RenderWeekGrid(m.weekStart, time.Now(), m.provider, m.mondayStart, m.store, hasEvents, m.styles)
		content = grid + m.renderOverview()
	} else {
		todayDay := 0
		now := time.Now()
		if now.Year() == m.year && now.Month() == m.month {
			todayDay = now.Day()
		}

		grid := RenderGrid(m.year, m.month, todayDay, m.holidays, m.mondayStart, m.indicators, m.totals, m.priorities, m.store, m.showMonthTodos, m.showYearTodos, m.contentWidth, hasEvents, m.styles)
		content = grid + m.renderOverview()
	}

	return content
}

// renderOverview builds the overview section showing per-month todo counts
// (pending in red-family, completed in green-family) and the floating (undated)
// todo counts. It is computed fresh from the store on every render to guarantee
// live updates without cache invalidation.
func (m Model) renderOverview() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.styles.OverviewHeader.Render("Overview"))
	b.WriteString("\n")

	months := m.store.TodoCountsByMonth()
	for _, mc := range months {
		label := mc.Month.String()
		if mc.Year != m.year {
			label = fmt.Sprintf("%s %d", mc.Month.String(), mc.Year)
		}

		paddedLabel := fmt.Sprintf(" %-16s", label)
		pending := m.styles.OverviewPending.Render(fmt.Sprintf("%d", mc.Pending))
		completed := m.styles.OverviewCompleted.Render(fmt.Sprintf("%d", mc.Completed))

		if mc.Year == m.year && mc.Month == m.month {
			b.WriteString(m.styles.OverviewActive.Render(paddedLabel))
		} else {
			b.WriteString(m.styles.OverviewCount.Render(paddedLabel))
		}
		b.WriteString(pending)
		b.WriteString("  ")
		b.WriteString(completed)
		b.WriteString("\n")
	}

	fc := m.store.FloatingTodoCounts()
	if fc.Pending > 0 || fc.Completed > 0 {
		paddedLabel := fmt.Sprintf(" %-16s", "Unknown")
		b.WriteString(m.styles.OverviewCount.Render(paddedLabel))
		b.WriteString(m.styles.OverviewPending.Render(fmt.Sprintf("%d", fc.Pending)))
		b.WriteString("  ")
		b.WriteString(m.styles.OverviewCompleted.Render(fmt.Sprintf("%d", fc.Completed)))
		b.WriteString("\n")
	}

	return b.String()
}

// RefreshIndicators recomputes the indicator data for the current month.
// Call this after todo mutations to keep the calendar display in sync.
func (m *Model) RefreshIndicators() {
	m.indicators = m.store.IncompleteTodosPerDay(m.year, m.month)
	m.totals = m.store.TotalTodosPerDay(m.year, m.month)
	m.priorities = m.store.HighestPriorityPerDay(m.year, m.month)
}

// SetFocused sets whether this pane is focused.
func (m *Model) SetFocused(f bool) {
	m.focused = f
}

// SetContentWidth sets the pane text content width (pane width minus padding).
func (m *Model) SetContentWidth(w int) {
	m.contentWidth = w
}

// Year returns the currently viewed year.
func (m Model) Year() int { return m.year }

// Month returns the currently viewed month.
func (m Model) Month() time.Month { return m.month }

// SetTheme replaces the calendar styles with ones built from the given theme.
// This preserves all model state (year, month, cursor, provider).
func (m *Model) SetTheme(t theme.Theme) {
	m.styles = NewStyles(t)
}

// SetProvider replaces the holiday provider and refreshes holidays for the current month.
func (m *Model) SetProvider(p *holidays.Provider) {
	m.provider = p
	m.holidays = p.HolidaysInMonth(m.year, m.month)
}

// SetMondayStart sets whether the week starts on Monday.
func (m *Model) SetMondayStart(v bool) {
	m.mondayStart = v
}

// SetShowFuzzySections controls visibility of fuzzy-date circle indicators.
func (m *Model) SetShowFuzzySections(showMonth, showYear bool) {
	m.showMonthTodos = showMonth
	m.showYearTodos = showYear
}

// SetCalendarEvents stores the Google Calendar events for grid indicator display.
func (m *Model) SetCalendarEvents(events []google.CalendarEvent) {
	m.calendarEvents = events
}

// hasEventsPerDay computes a map of day numbers that have calendar events
// in the given year/month.
func (m Model) hasEventsPerDay(year int, month time.Month) map[int]bool {
	result := make(map[int]bool)
	expanded := google.ExpandMultiDay(m.calendarEvents)
	prefix := fmt.Sprintf("%04d-%02d-", year, int(month))
	for _, e := range expanded {
		if strings.HasPrefix(e.Date, prefix) {
			day, err := strconv.Atoi(e.Date[8:10])
			if err == nil {
				result[day] = true
			}
		}
	}
	return result
}

// Keys returns the calendar key bindings (for help bar aggregation).
// Help text is contextual: in WeekView, PrevMonth/NextMonth show "prev week"/"next week"
// and ToggleWeek shows "monthly view".
func (m Model) Keys() KeyMap {
	k := m.keys
	if m.viewMode == WeekView {
		k.PrevMonth = key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("<-/h", "prev week"))
		k.NextMonth = key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("->/l", "next week"))
		k.ToggleWeek = key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "monthly view"))
	}
	return k
}

// GetViewMode returns the current view mode (MonthView or WeekView).
func (m Model) GetViewMode() ViewMode { return m.viewMode }

// WeekStart returns the start date of the currently viewed week.
func (m Model) WeekStart() time.Time { return m.weekStart }

// SetYearMonth navigates directly to the specified year and month,
// refreshing holidays and indicators.
func (m *Model) SetYearMonth(year int, month time.Month) {
	m.year = year
	m.month = month
	m.holidays = m.provider.HolidaysInMonth(year, month)
	m.indicators = m.store.IncompleteTodosPerDay(year, month)
	m.totals = m.store.TotalTodosPerDay(year, month)
	m.priorities = m.store.HighestPriorityPerDay(year, month)
}
