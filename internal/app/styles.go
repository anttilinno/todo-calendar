package app

import "github.com/charmbracelet/lipgloss"

var (
	focusedBorderColor   = lipgloss.Color("62")  // purple
	unfocusedBorderColor = lipgloss.Color("240") // gray

	focusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(focusedBorderColor).
			Padding(0, 1)

	unfocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(unfocusedBorderColor).
			Padding(0, 1)
)

// paneStyle returns the appropriate lipgloss style based on focus state.
func paneStyle(focused bool) lipgloss.Style {
	if focused {
		return focusedStyle
	}
	return unfocusedStyle
}
