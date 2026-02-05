package calendar

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

// Model represents the calendar pane.
type Model struct {
	focused     bool
	width       int
	height      int
	year        int
	month       time.Month
	today       time.Time
	holidays    map[int]bool
	indicators  map[int]int
	provider    *holidays.Provider
	store       *store.Store
	keys        KeyMap
	mondayStart bool
	styles      Styles
}

// New creates a new calendar model with the given holiday provider,
// week-start preference, and store for indicator data.
func New(provider *holidays.Provider, mondayStart bool, s *store.Store, t theme.Theme) Model {
	now := time.Now()
	y, m, _ := now.Date()

	return Model{
		year:        y,
		month:       m,
		today:       now,
		holidays:    provider.HolidaysInMonth(y, m),
		indicators:  s.IncompleteTodosPerDay(y, m),
		provider:    provider,
		store:       s,
		keys:        DefaultKeyMap(),
		mondayStart: mondayStart,
		styles:      NewStyles(t),
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
		case key.Matches(msg, m.keys.PrevMonth):
			m.month--
			if m.month < time.January {
				m.month = time.December
				m.year--
			}
			m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
			m.indicators = m.store.IncompleteTodosPerDay(m.year, m.month)

		case key.Matches(msg, m.keys.NextMonth):
			m.month++
			if m.month > time.December {
				m.month = time.January
				m.year++
			}
			m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
			m.indicators = m.store.IncompleteTodosPerDay(m.year, m.month)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the calendar pane content.
func (m Model) View() string {
	todayDay := 0
	now := time.Now()
	if now.Year() == m.year && now.Month() == m.month {
		todayDay = now.Day()
	}

	return RenderGrid(m.year, m.month, todayDay, m.holidays, m.mondayStart, m.indicators, m.styles)
}

// RefreshIndicators recomputes the indicator data for the current month.
// Call this after todo mutations to keep the calendar display in sync.
func (m *Model) RefreshIndicators() {
	m.indicators = m.store.IncompleteTodosPerDay(m.year, m.month)
}

// SetFocused sets whether this pane is focused.
func (m *Model) SetFocused(f bool) {
	m.focused = f
}

// Year returns the currently viewed year.
func (m Model) Year() int { return m.year }

// Month returns the currently viewed month.
func (m Model) Month() time.Month { return m.month }

// Keys returns the calendar key bindings (for help bar aggregation).
func (m Model) Keys() KeyMap { return m.keys }
