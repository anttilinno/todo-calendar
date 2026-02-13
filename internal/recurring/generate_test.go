package recurring

import (
	"fmt"
	"testing"
	"time"

	"github.com/antti/todo-calendar/internal/store"
)

// fakeStore implements store.TodoStore with only the methods AutoCreate needs.
// All other methods are stubs to satisfy the interface.
type fakeStore struct {
	schedules  []store.Schedule
	templates  map[int]*store.Template
	existing   map[string]bool // key: "scheduleID:date"
	added      []addedTodo
}

type addedTodo struct {
	text       string
	date       string
	body       string
	scheduleID int
}

func (f *fakeStore) ListSchedules() []store.Schedule          { return f.schedules }
func (f *fakeStore) FindTemplate(id int) *store.Template      { return f.templates[id] }
func (f *fakeStore) TodoExistsForSchedule(scheduleID int, date string) bool {
	key := fakeKey(scheduleID, date)
	return f.existing[key]
}
func (f *fakeStore) AddScheduledTodo(text, date, body string, scheduleID int) store.Todo {
	f.added = append(f.added, addedTodo{text: text, date: date, body: body, scheduleID: scheduleID})
	// Mark as existing so dedup works within same run
	if f.existing == nil {
		f.existing = make(map[string]bool)
	}
	f.existing[fakeKey(scheduleID, date)] = true
	return store.Todo{}
}

func fakeKey(scheduleID int, date string) string {
	return fmt.Sprintf("%d:%s", scheduleID, date)
}

// Stub methods to satisfy store.TodoStore interface.
func (f *fakeStore) Add(text, date, datePrecision string, priority int) store.Todo { return store.Todo{} }
func (f *fakeStore) Toggle(id int)                                     {}
func (f *fakeStore) Delete(id int)                                     {}
func (f *fakeStore) Find(id int) *store.Todo                           { return nil }
func (f *fakeStore) Update(id int, text, date, datePrecision string, priority int) {}
func (f *fakeStore) Todos() []store.Todo                               { return nil }
func (f *fakeStore) TodosForMonth(y int, m time.Month) []store.Todo    { return nil }
func (f *fakeStore) TodosForDateRange(startDate, endDate string) []store.Todo { return nil }
func (f *fakeStore) MonthTodos(y int, m time.Month) []store.Todo       { return nil }
func (f *fakeStore) YearTodos(y int) []store.Todo                      { return nil }
func (f *fakeStore) FloatingTodos() []store.Todo                       { return nil }
func (f *fakeStore) IncompleteTodosPerDay(y int, m time.Month) map[int]int { return nil }
func (f *fakeStore) TotalTodosPerDay(y int, m time.Month) map[int]int { return nil }
func (f *fakeStore) TodoCountsByMonth() []store.MonthCount            { return nil }
func (f *fakeStore) FloatingTodoCounts() store.FloatingCount          { return store.FloatingCount{} }
func (f *fakeStore) UpdateBody(id int, body string)                   {}
func (f *fakeStore) AddTemplate(name, content string) (store.Template, error) {
	return store.Template{}, nil
}
func (f *fakeStore) ListTemplates() []store.Template              { return nil }
func (f *fakeStore) DeleteTemplate(id int)                        {}
func (f *fakeStore) UpdateTemplate(id int, name, content string) error { return nil }
func (f *fakeStore) AddSchedule(templateID int, cadenceType, cadenceValue, placeholderDefaults string) (store.Schedule, error) {
	return store.Schedule{}, nil
}
func (f *fakeStore) ListSchedulesForTemplate(templateID int) []store.Schedule { return nil }
func (f *fakeStore) DeleteSchedule(id int)                                    {}
func (f *fakeStore) UpdateSchedule(id int, cadenceType, cadenceValue, placeholderDefaults string) error {
	return nil
}
func (f *fakeStore) HighestPriorityPerDay(y int, m time.Month) map[int]int { return nil }
func (f *fakeStore) SwapOrder(id1, id2 int)      {}
func (f *fakeStore) SearchTodos(query string) []store.Todo { return nil }
func (f *fakeStore) EnsureSortOrder()             {}
func (f *fakeStore) Save() error                  { return nil }

func TestAutoCreateDailySchedule(t *testing.T) {
	// A daily schedule should create todos for all 7 days in the window.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 1, TemplateID: 10, CadenceType: "daily", CadenceValue: "", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{
			10: {ID: 10, Name: "Daily Standup", Content: "standup notes"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC) // Monday
	AutoCreateForDate(fs, today)

	if len(fs.added) != 7 {
		t.Fatalf("daily schedule: want 7 todos, got %d", len(fs.added))
	}
	for i, a := range fs.added {
		expected := today.AddDate(0, 0, i).Format("2006-01-02")
		if a.date != expected {
			t.Errorf("todo %d: want date %s, got %s", i, expected, a.date)
		}
		if a.text != "Daily Standup" {
			t.Errorf("todo %d: want text 'Daily Standup', got %q", i, a.text)
		}
		if a.body != "standup notes" {
			t.Errorf("todo %d: want body 'standup notes', got %q", i, a.body)
		}
		if a.scheduleID != 1 {
			t.Errorf("todo %d: want scheduleID 1, got %d", i, a.scheduleID)
		}
	}
}

func TestAutoCreateWeeklySchedule(t *testing.T) {
	// Weekly schedule for mon,wed should create 2 todos in a Mon-Sun window.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 2, TemplateID: 20, CadenceType: "weekly", CadenceValue: "mon,wed", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{
			20: {ID: 20, Name: "Team Sync", Content: "sync agenda"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC) // Monday
	AutoCreateForDate(fs, today)

	if len(fs.added) != 2 {
		t.Fatalf("weekly schedule: want 2 todos, got %d", len(fs.added))
	}
	// Monday = 2026-02-09, Wednesday = 2026-02-11
	if fs.added[0].date != "2026-02-09" {
		t.Errorf("first weekly todo: want 2026-02-09, got %s", fs.added[0].date)
	}
	if fs.added[1].date != "2026-02-11" {
		t.Errorf("second weekly todo: want 2026-02-11, got %s", fs.added[1].date)
	}
}

func TestAutoCreateMonthlySchedule(t *testing.T) {
	// Monthly schedule for day 10 -- in window starting Feb 9 (7 days: 9-15), day 10 is in range.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 3, TemplateID: 30, CadenceType: "monthly", CadenceValue: "10", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{
			30: {ID: 30, Name: "Monthly Report", Content: "report content"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	AutoCreateForDate(fs, today)

	if len(fs.added) != 1 {
		t.Fatalf("monthly schedule: want 1 todo, got %d", len(fs.added))
	}
	if fs.added[0].date != "2026-02-10" {
		t.Errorf("monthly todo: want 2026-02-10, got %s", fs.added[0].date)
	}
}

func TestAutoCreateMonthlyNotInWindow(t *testing.T) {
	// Monthly schedule for day 20 -- window is Feb 9-15, day 20 is NOT in range.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 4, TemplateID: 30, CadenceType: "monthly", CadenceValue: "20", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{
			30: {ID: 30, Name: "Monthly Review", Content: "review"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	AutoCreateForDate(fs, today)

	if len(fs.added) != 0 {
		t.Fatalf("monthly schedule (out of window): want 0 todos, got %d", len(fs.added))
	}
}

func TestAutoCreateDedup(t *testing.T) {
	// Calling AutoCreate twice should not produce extra todos.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 5, TemplateID: 50, CadenceType: "daily", CadenceValue: "", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{
			50: {ID: 50, Name: "Daily", Content: "daily body"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	AutoCreateForDate(fs, today)
	count1 := len(fs.added)

	AutoCreateForDate(fs, today)
	count2 := len(fs.added)

	if count1 != 7 {
		t.Fatalf("first run: want 7 todos, got %d", count1)
	}
	if count2 != 7 {
		t.Fatalf("second run should not add more: want 7 total, got %d", count2)
	}
}

func TestAutoCreatePlaceholderDefaults(t *testing.T) {
	// Template with placeholders should be filled from schedule's PlaceholderDefaults.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{
				ID:                  6,
				TemplateID:          60,
				CadenceType:         "daily",
				CadenceValue:        "",
				PlaceholderDefaults: `{"Project":"Alpha","Owner":"Alice"}`,
			},
		},
		templates: map[int]*store.Template{
			60: {ID: 60, Name: "Status Update", Content: "Project: {{.Project}}\nOwner: {{.Owner}}"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	AutoCreateForDate(fs, today)

	if len(fs.added) != 7 {
		t.Fatalf("want 7 todos, got %d", len(fs.added))
	}
	expected := "Project: Alpha\nOwner: Alice"
	if fs.added[0].body != expected {
		t.Errorf("body with defaults: want %q, got %q", expected, fs.added[0].body)
	}
}

func TestAutoCreateEmptyPlaceholderDefaults(t *testing.T) {
	// Template with placeholders but empty defaults should produce empty values.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 7, TemplateID: 70, CadenceType: "daily", CadenceValue: "", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{
			70: {ID: 70, Name: "Review", Content: "Project: {{.Project}}"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	AutoCreateForDate(fs, today)

	if len(fs.added) != 7 {
		t.Fatalf("want 7 todos, got %d", len(fs.added))
	}
	// With missingkey=zero, missing key produces empty string.
	expected := "Project: "
	if fs.added[0].body != expected {
		t.Errorf("body with empty defaults: want %q, got %q", expected, fs.added[0].body)
	}
}

func TestAutoCreateMissingTemplate(t *testing.T) {
	// Schedule with missing template (orphan) should be skipped without panic.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 8, TemplateID: 999, CadenceType: "daily", CadenceValue: "", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{}, // No template with ID 999
		existing:  make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	AutoCreateForDate(fs, today) // Should not panic

	if len(fs.added) != 0 {
		t.Fatalf("missing template: want 0 todos, got %d", len(fs.added))
	}
}

func TestAutoCreateBadCadence(t *testing.T) {
	// Schedule with unparseable cadence should be skipped without panic.
	fs := &fakeStore{
		schedules: []store.Schedule{
			{ID: 9, TemplateID: 90, CadenceType: "bogus", CadenceValue: "xyz", PlaceholderDefaults: "{}"},
		},
		templates: map[int]*store.Template{
			90: {ID: 90, Name: "Bad", Content: "content"},
		},
		existing: make(map[string]bool),
	}

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	AutoCreateForDate(fs, today) // Should not panic

	if len(fs.added) != 0 {
		t.Fatalf("bad cadence: want 0 todos, got %d", len(fs.added))
	}
}
