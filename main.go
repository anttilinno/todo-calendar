package main

import (
	"fmt"
	"os"

	"github.com/antti/todo-calendar/internal/app"
	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	provider, err := holidays.NewProvider(cfg.Country)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Holiday provider error: %v\n", err)
		os.Exit(1)
	}

	todosPath, err := store.TodosPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Todos path error: %v\n", err)
		os.Exit(1)
	}

	s, err := store.NewStore(todosPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Store error: %v\n", err)
		os.Exit(1)
	}

	model := app.New(provider, cfg.MondayStart(), s)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
