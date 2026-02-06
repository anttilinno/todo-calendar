package calendar

import (
	"fmt"
	"strings"
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

// View renders the calendar pane content including the overview section.
func (m Model) View() string {
	todayDay := 0
	now := time.Now()
	if now.Year() == m.year && now.Month() == m.month {
		todayDay = now.Day()
	}

	grid := RenderGrid(m.year, m.month, todayDay, m.holidays, m.mondayStart, m.indicators, m.styles)
	return grid + m.renderOverview()
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
	paddedLabel := fmt.Sprintf(" %-16s", "Unknown")
	b.WriteString(m.styles.OverviewCount.Render(paddedLabel))
	b.WriteString(m.styles.OverviewPending.Render(fmt.Sprintf("%d", fc.Pending)))
	b.WriteString("  ")
	b.WriteString(m.styles.OverviewCompleted.Render(fmt.Sprintf("%d", fc.Completed)))
	b.WriteString("\n")

	return b.String()
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

// Keys returns the calendar key bindings (for help bar aggregation).
func (m Model) Keys() KeyMap { return m.keys }
