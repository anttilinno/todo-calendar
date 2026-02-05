package main

import (
	"fmt"
	"os"

	"github.com/antti/todo-calendar/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model := app.New()
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
