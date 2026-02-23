package search

import (
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds themed lipgloss styles for search overlay rendering.
type Styles struct {
	Title          lipgloss.Style
	ResultText     lipgloss.Style
	ResultDate     lipgloss.Style
	SelectedResult lipgloss.Style
	SelectedDate   lipgloss.Style
	Hint           lipgloss.Style
	Empty          lipgloss.Style
	PriorityP1     lipgloss.Style
	PriorityP2     lipgloss.Style
	PriorityP3     lipgloss.Style
	PriorityP4     lipgloss.Style
	PriorityMuted  lipgloss.Style
}

// NewStyles builds search styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		Title:          lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		ResultText:     lipgloss.NewStyle().Foreground(t.NormalFg),
		ResultDate:     lipgloss.NewStyle().Foreground(t.MutedFg),
		SelectedResult: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		SelectedDate:   lipgloss.NewStyle().Foreground(t.AccentFg),
		Hint:           lipgloss.NewStyle().Foreground(t.MutedFg),
		Empty:          lipgloss.NewStyle().Foreground(t.MutedFg),
		PriorityP1:     lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP1Fg),
		PriorityP2:     lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP2Fg),
		PriorityP3:     lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP3Fg),
		PriorityP4:     lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP4Fg),
		PriorityMuted:  lipgloss.NewStyle().Foreground(t.MutedFg),
	}
}

// priorityBadgeStyle returns the appropriate priority badge style for the given level (1-3).
func (s Styles) priorityBadgeStyle(level int) lipgloss.Style {
	switch level {
	case 1:
		return s.PriorityP1
	case 2:
		return s.PriorityP2
	case 3:
		return s.PriorityP3
	default:
		return s.PriorityP3
	}
}
