package main

import (
	"fmt"
	"os"

	"github.com/antti/todo-calendar/internal/app"
	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/google"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/recurring"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/api/calendar/v3"
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

	dbPath, err := config.DBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database path error: %v\n", err)
		os.Exit(1)
	}

	s, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Store error: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	recurring.AutoCreate(s)

	authState := google.CheckAuthState()

	var calSvc *calendar.Service
	if authState == google.AuthReady {
		calSvc, _ = google.NewCalendarService()
	}

	t := theme.ForName(cfg.Theme)
	model := app.New(provider, cfg.MondayStart(), s, t, cfg, authState, calSvc)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
