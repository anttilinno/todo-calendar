package todolist

import (
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds themed lipgloss styles for todo list rendering.
type Styles struct {
	SectionHeader lipgloss.Style
	Completed     lipgloss.Style
	Cursor        lipgloss.Style
	Date          lipgloss.Style
	Empty         lipgloss.Style
	BodyIndicator      lipgloss.Style
	RecurringIndicator lipgloss.Style
	Separator          lipgloss.Style
	Checkbox      lipgloss.Style
	CheckboxDone  lipgloss.Style
	EditTitle     lipgloss.Style
	FieldLabel    lipgloss.Style
	EditHint      lipgloss.Style
	DateSeparator lipgloss.Style
	PriorityP1    lipgloss.Style
	PriorityP2    lipgloss.Style
	PriorityP3    lipgloss.Style
	PriorityP4    lipgloss.Style
	PriorityMuted lipgloss.Style
	EventTime     lipgloss.Style
	EventText     lipgloss.Style
}

// NewStyles builds todo list styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		SectionHeader: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		Completed:     lipgloss.NewStyle().Strikethrough(true).Foreground(t.CompletedFg),
		Cursor:        lipgloss.NewStyle().Foreground(t.AccentFg),
		Date:          lipgloss.NewStyle().Foreground(t.MutedFg),
		Empty:         lipgloss.NewStyle().Foreground(t.EmptyFg),
		BodyIndicator:      lipgloss.NewStyle().Foreground(t.MutedFg),
		RecurringIndicator: lipgloss.NewStyle().Foreground(t.MutedFg),
		Separator:          lipgloss.NewStyle().Foreground(t.MutedFg),
		Checkbox:      lipgloss.NewStyle().Foreground(t.AccentFg),
		CheckboxDone:  lipgloss.NewStyle().Foreground(t.CompletedCountFg),
		EditTitle:     lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		FieldLabel:    lipgloss.NewStyle().Bold(true).Foreground(t.NormalFg),
		EditHint:      lipgloss.NewStyle().Foreground(t.MutedFg),
		DateSeparator: lipgloss.NewStyle().Foreground(t.MutedFg),
		PriorityP1:    lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP1Fg),
		PriorityP2:    lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP2Fg),
		PriorityP3:    lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP3Fg),
		PriorityP4:    lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP4Fg),
		PriorityMuted: lipgloss.NewStyle().Foreground(t.MutedFg),
		EventTime:     lipgloss.NewStyle().Foreground(t.EventFg).Bold(true),
		EventText:     lipgloss.NewStyle().Foreground(t.EventFg),
	}
}

// priorityBadgeStyle returns the appropriate priority badge style for the given level (1-4).
func (s Styles) priorityBadgeStyle(level int) lipgloss.Style {
	switch level {
	case 1:
		return s.PriorityP1
	case 2:
		return s.PriorityP2
	case 3:
		return s.PriorityP3
	case 4:
		return s.PriorityP4
	default:
		return s.PriorityP4
	}
}
