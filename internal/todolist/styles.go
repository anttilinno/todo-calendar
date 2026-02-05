package todolist

import "github.com/charmbracelet/lipgloss"

var (
	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("62"))

	completedStyle = lipgloss.NewStyle().
			Faint(true).
			Strikethrough(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62"))

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	emptyStyle = lipgloss.NewStyle().
			Faint(true)
)
