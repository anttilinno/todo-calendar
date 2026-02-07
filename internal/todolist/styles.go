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
	BodyIndicator lipgloss.Style
	Separator     lipgloss.Style
	Checkbox      lipgloss.Style
	CheckboxDone  lipgloss.Style
	EditTitle     lipgloss.Style
	FieldLabel    lipgloss.Style
	EditHint      lipgloss.Style
}

// NewStyles builds todo list styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		SectionHeader: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		Completed:     lipgloss.NewStyle().Strikethrough(true).Foreground(t.CompletedFg),
		Cursor:        lipgloss.NewStyle().Foreground(t.AccentFg),
		Date:          lipgloss.NewStyle().Foreground(t.MutedFg),
		Empty:         lipgloss.NewStyle().Foreground(t.EmptyFg),
		BodyIndicator: lipgloss.NewStyle().Foreground(t.MutedFg),
		Separator:     lipgloss.NewStyle().Foreground(t.MutedFg),
		Checkbox:      lipgloss.NewStyle().Foreground(t.AccentFg),
		CheckboxDone:  lipgloss.NewStyle().Foreground(t.CompletedCountFg),
		EditTitle:     lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		FieldLabel:    lipgloss.NewStyle().Bold(true).Foreground(t.NormalFg),
		EditHint:      lipgloss.NewStyle().Foreground(t.MutedFg),
	}
}
