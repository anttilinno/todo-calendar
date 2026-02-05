package app

import (
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds themed lipgloss styles for the app-level pane borders.
type Styles struct {
	Focused   lipgloss.Style
	Unfocused lipgloss.Style
}

// NewStyles builds app styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		Focused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.BorderFocused).
			Padding(0, 1),
		Unfocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.BorderUnfocused).
			Padding(0, 1),
	}
}

// Pane returns the appropriate lipgloss style based on focus state.
func (s Styles) Pane(focused bool) lipgloss.Style {
	if focused {
		return s.Focused
	}
	return s.Unfocused
}
