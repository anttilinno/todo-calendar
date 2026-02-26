package main

import (
	"flag"
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

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Show version")
	showStatus := flag.Bool("status", false, "Show today's pending todo count")
	flag.BoolVar(showVersion, "v", false, "Show version")
	flag.BoolVar(showStatus, "s", false, "Show today's pending todo count")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: todo-calendar [flags]\n\nA terminal calendar with todo management.\n\nFlags:\n")
		fmt.Fprintf(os.Stderr, "  -s, --status   Show today's pending todo count\n")
		fmt.Fprintf(os.Stderr, "  -v, --version  Show version\n")
		fmt.Fprintf(os.Stderr, "  -h, --help     Show this help\n")
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

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

	if *showStatus {
		today := time.Now().Format("2006-01-02")
		todos := s.TodosForDateRange(today, today)
		fmt.Print(status.FormatStatus(todos))
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
