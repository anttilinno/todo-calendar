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
	}
}
