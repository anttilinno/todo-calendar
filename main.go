package main

import (
	"fmt"
	"os"
	"time"

	"github.com/antti/todo-calendar/internal/app"
	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/google"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/recurring"
	"github.com/antti/todo-calendar/internal/status"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	gcal "google.golang.org/api/calendar/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
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

	// Subcommand routing: branch before TUI setup.
	if len(os.Args) >= 2 && os.Args[1] == "status" {
		runStatus(s)
		return
	}

	provider, err := holidays.NewProvider(cfg.Country)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Holiday provider error: %v\n", err)
		os.Exit(1)
	}

	recurring.AutoCreate(s)

	authState := google.CheckAuthState()

	var calSvc *gcal.Service
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

// runStatus queries today's todos, formats the count, and writes the state file.
func runStatus(s *store.SQLiteStore) {
	today := time.Now().Format("2006-01-02")
	todos := s.TodosForDateRange(today, today)

	output := status.FormatStatus(todos)

	if err := status.WriteStatusFile(output); err != nil {
		fmt.Fprintf(os.Stderr, "Status write error: %v\n", err)
		os.Exit(1)
	}
}
