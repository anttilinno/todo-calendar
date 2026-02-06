package settings

import (
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds themed lipgloss styles for settings overlay rendering.
type Styles struct {
	Title         lipgloss.Style
	Label         lipgloss.Style
	Value         lipgloss.Style
	SelectedLabel lipgloss.Style
	SelectedValue lipgloss.Style
	Hint          lipgloss.Style
}

// NewStyles builds settings styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		Title:         lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		Label:         lipgloss.NewStyle().Foreground(t.NormalFg),
		Value:         lipgloss.NewStyle().Foreground(t.MutedFg),
		SelectedLabel: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		SelectedValue: lipgloss.NewStyle().Foreground(t.AccentFg),
		Hint:          lipgloss.NewStyle().Foreground(t.MutedFg),
	}
}
