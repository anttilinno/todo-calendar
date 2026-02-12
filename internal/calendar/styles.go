package calendar

import (
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds themed lipgloss styles for calendar grid rendering.
type Styles struct {
	Header         lipgloss.Style
	WeekdayHdr     lipgloss.Style
	Normal         lipgloss.Style
	Today          lipgloss.Style
	Holiday        lipgloss.Style
	Indicator      lipgloss.Style
	IndicatorDone  lipgloss.Style
	TodayIndicator lipgloss.Style
	TodayDone      lipgloss.Style
	OverviewHeader    lipgloss.Style
	OverviewCount     lipgloss.Style
	OverviewActive    lipgloss.Style
	OverviewPending   lipgloss.Style
	OverviewCompleted lipgloss.Style
	FuzzyPending      lipgloss.Style
	FuzzyDone         lipgloss.Style
}

// NewStyles builds calendar styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		Header:         lipgloss.NewStyle().Bold(true).Foreground(t.HeaderFg),
		WeekdayHdr:     lipgloss.NewStyle().Foreground(t.WeekdayFg),
		Normal:         lipgloss.NewStyle().Foreground(t.NormalFg),
		Today:          lipgloss.NewStyle().Bold(true).Foreground(t.TodayFg).Background(t.TodayBg),
		Holiday:        lipgloss.NewStyle().Foreground(t.HolidayFg),
		Indicator:      lipgloss.NewStyle().Bold(true).Foreground(t.IndicatorFg),
		IndicatorDone:  lipgloss.NewStyle().Foreground(t.CompletedCountFg),
		TodayIndicator: lipgloss.NewStyle().Bold(true).Foreground(t.IndicatorFg).Background(t.TodayBg),
		TodayDone:      lipgloss.NewStyle().Bold(true).Foreground(t.CompletedCountFg).Background(t.TodayBg),
		OverviewHeader:    lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		OverviewCount:     lipgloss.NewStyle().Foreground(t.MutedFg),
		OverviewActive:    lipgloss.NewStyle().Bold(true).Foreground(t.NormalFg),
		OverviewPending:   lipgloss.NewStyle().Foreground(t.PendingFg),
		OverviewCompleted: lipgloss.NewStyle().Foreground(t.CompletedCountFg),
		FuzzyPending:      lipgloss.NewStyle().Foreground(t.PendingFg),
		FuzzyDone:         lipgloss.NewStyle().Foreground(t.CompletedCountFg),
	}
}
