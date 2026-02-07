package tmplmgr

import (
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds themed lipgloss styles for the template management overlay.
type Styles struct {
	Title              lipgloss.Style
	TemplateName       lipgloss.Style
	SelectedName       lipgloss.Style
	Separator          lipgloss.Style
	Content            lipgloss.Style
	Hint               lipgloss.Style
	Error              lipgloss.Style
	Empty              lipgloss.Style
	ScheduleSuffix     lipgloss.Style
	ScheduleActive     lipgloss.Style
	ScheduleInactive   lipgloss.Style
	ScheduleDay        lipgloss.Style
	ScheduleDaySelected lipgloss.Style
	SchedulePrompt     lipgloss.Style
}

// NewStyles builds template management styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		Title:        lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		TemplateName: lipgloss.NewStyle().Foreground(t.NormalFg),
		SelectedName: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		Separator:    lipgloss.NewStyle().Foreground(t.MutedFg),
		Content:      lipgloss.NewStyle().Foreground(t.NormalFg),
		Hint:         lipgloss.NewStyle().Foreground(t.MutedFg),
		Error:           lipgloss.NewStyle().Foreground(t.HolidayFg),
		Empty:              lipgloss.NewStyle().Foreground(t.MutedFg),
		ScheduleSuffix:     lipgloss.NewStyle().Foreground(t.MutedFg),
		ScheduleActive:     lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		ScheduleInactive:   lipgloss.NewStyle().Foreground(t.MutedFg),
		ScheduleDay:        lipgloss.NewStyle().Foreground(t.NormalFg),
		ScheduleDaySelected: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		SchedulePrompt:     lipgloss.NewStyle().Foreground(t.NormalFg),
	}
}
