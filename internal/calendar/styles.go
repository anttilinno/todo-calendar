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
	IndicatorP1      lipgloss.Style // P1 priority pending indicator
	IndicatorP2      lipgloss.Style // P2 priority pending indicator
	IndicatorP3      lipgloss.Style // P3 priority pending indicator
	IndicatorP4      lipgloss.Style // P4 priority pending indicator
	TodayIndicatorP1 lipgloss.Style // P1 on today's date
	TodayIndicatorP2 lipgloss.Style // P2 on today's date
	TodayIndicatorP3 lipgloss.Style // P3 on today's date
	TodayIndicatorP4 lipgloss.Style // P4 on today's date
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
		IndicatorP1:      lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP1Fg),
		IndicatorP2:      lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP2Fg),
		IndicatorP3:      lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP3Fg),
		IndicatorP4:      lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP4Fg),
		TodayIndicatorP1: lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP1Fg).Background(t.TodayBg),
		TodayIndicatorP2: lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP2Fg).Background(t.TodayBg),
		TodayIndicatorP3: lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP3Fg).Background(t.TodayBg),
		TodayIndicatorP4: lipgloss.NewStyle().Bold(true).Foreground(t.PriorityP4Fg).Background(t.TodayBg),
		OverviewHeader:    lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
		OverviewCount:     lipgloss.NewStyle().Foreground(t.MutedFg),
		OverviewActive:    lipgloss.NewStyle().Bold(true).Foreground(t.NormalFg),
		OverviewPending:   lipgloss.NewStyle().Foreground(t.PendingFg),
		OverviewCompleted: lipgloss.NewStyle().Foreground(t.CompletedCountFg),
		FuzzyPending:      lipgloss.NewStyle().Foreground(t.PendingFg),
		FuzzyDone:         lipgloss.NewStyle().Foreground(t.CompletedCountFg),
	}
}
