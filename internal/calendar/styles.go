package calendar

import "github.com/charmbracelet/lipgloss"

// Lip Gloss styles for calendar grid cells.
// Unexported -- used only by grid.go within this package.
var (
	headerStyle      = lipgloss.NewStyle().Bold(true)
	weekdayHdrStyle  = lipgloss.NewStyle().Faint(true)
	normalStyle      = lipgloss.NewStyle()
	todayStyle       = lipgloss.NewStyle().Bold(true).Reverse(true)
	holidayStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	indicatorStyle   = lipgloss.NewStyle().Bold(true)
)
