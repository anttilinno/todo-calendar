package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/antti/todo-calendar/internal/tmpl"
)

func TestSeedTemplates(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	templates := s.ListTemplates()
	if len(templates) != 7 {
		t.Fatalf("expected 7 templates, got %d", len(templates))
	}

	for _, tpl := range templates {
		if tpl.Name == "" {
			t.Error("template has empty name")
		}
		if tpl.Content == "" {
			t.Errorf("template %q has empty content", tpl.Name)
		}
		_, err := tmpl.ExtractPlaceholders(tpl.Content)
		if err != nil {
			t.Errorf("template %q has invalid placeholder syntax: %v", tpl.Name, err)
		}
	}
}

func TestSeedTemplates_Idempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s1, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	count1 := len(s1.ListTemplates())
	s1.Close()

	s2, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	count2 := len(s2.ListTemplates())
	s2.Close()

	if count1 != count2 {
		t.Errorf("template count changed: %d -> %d", count1, count2)
	}
}

func TestSeedTemplates_DeletionPermanent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s1, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	templates := s1.ListTemplates()
	for _, tpl := range templates {
		s1.DeleteTemplate(tpl.ID)
	}
	if len(s1.ListTemplates()) != 0 {
		t.Error("templates not deleted")
	}
	s1.Close()

	s2, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	if len(s2.ListTemplates()) != 0 {
		t.Error("deleted templates reappeared after reopen")
	}
	s2.Close()
}

func TestSeedTemplates_PlaceholderCounts(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	templates := s.ListTemplates()
	counts := make(map[string]int)
	for _, tpl := range templates {
		placeholders, err := tmpl.ExtractPlaceholders(tpl.Content)
		if err != nil {
			t.Fatalf("template %q: %v", tpl.Name, err)
		}
		counts[tpl.Name] = len(placeholders)
	}

	expected := map[string]int{
		"Meeting Notes": 2,
		"Checklist":     1,
		"Daily Plan":    0,
		"Bug Report":    2,
		"Feature Spec":  1,
		"PR Checklist":  1,
		"Code Review":   2,
	}

	for name, want := range expected {
		got, ok := counts[name]
		if !ok {
			t.Errorf("template %q not found", name)
			continue
		}
		if got != want {
			t.Errorf("template %q: expected %d placeholders, got %d", name, want, got)
		}
	}
}

func TestScheduleCRUD(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	// Create a template to link schedules to.
	tpl, err := s.AddTemplate("Weekly Review", "# Weekly Review\n\n- [ ] Check calendar")
	if err != nil {
		t.Fatalf("add template: %v", err)
	}

	// AddSchedule
	sc, err := s.AddSchedule(tpl.ID, "weekly", "monday", `{"assignee":"me"}`)
	if err != nil {
		t.Fatalf("add schedule: %v", err)
	}
	if sc.ID == 0 {
		t.Error("schedule ID should be non-zero")
	}
	if sc.TemplateID != tpl.ID {
		t.Errorf("expected template_id %d, got %d", tpl.ID, sc.TemplateID)
	}
	if sc.CadenceType != "weekly" {
		t.Errorf("expected cadence_type 'weekly', got %q", sc.CadenceType)
	}
	if sc.CadenceValue != "monday" {
		t.Errorf("expected cadence_value 'monday', got %q", sc.CadenceValue)
	}
	if sc.PlaceholderDefaults != `{"assignee":"me"}` {
		t.Errorf("expected placeholder_defaults, got %q", sc.PlaceholderDefaults)
	}

	// ListSchedules
	all := s.ListSchedules()
	if len(all) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(all))
	}
	if all[0].ID != sc.ID {
		t.Errorf("list returned wrong schedule ID: %d", all[0].ID)
	}

	// ListSchedulesForTemplate
	forTpl := s.ListSchedulesForTemplate(tpl.ID)
	if len(forTpl) != 1 {
		t.Fatalf("expected 1 schedule for template, got %d", len(forTpl))
	}

	// ListSchedulesForTemplate with wrong template ID returns empty.
	forOther := s.ListSchedulesForTemplate(9999)
	if len(forOther) != 0 {
		t.Errorf("expected 0 schedules for non-existent template, got %d", len(forOther))
	}

	// UpdateSchedule
	err = s.UpdateSchedule(sc.ID, "daily", "", `{}`)
	if err != nil {
		t.Fatalf("update schedule: %v", err)
	}
	updated := s.ListSchedules()
	if len(updated) != 1 {
		t.Fatalf("expected 1 schedule after update, got %d", len(updated))
	}
	if updated[0].CadenceType != "daily" {
		t.Errorf("expected cadence_type 'daily' after update, got %q", updated[0].CadenceType)
	}
	if updated[0].CadenceValue != "" {
		t.Errorf("expected empty cadence_value after update, got %q", updated[0].CadenceValue)
	}
	if updated[0].PlaceholderDefaults != `{}` {
		t.Errorf("expected '{}' placeholder_defaults after update, got %q", updated[0].PlaceholderDefaults)
	}

	// DeleteSchedule
	s.DeleteSchedule(sc.ID)
	afterDelete := s.ListSchedules()
	if len(afterDelete) != 0 {
		t.Errorf("expected 0 schedules after delete, got %d", len(afterDelete))
	}
}

func TestScheduleDeduplication(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	tpl, err := s.AddTemplate("Dedup Test", "content")
	if err != nil {
		t.Fatalf("add template: %v", err)
	}

	sc, err := s.AddSchedule(tpl.ID, "daily", "", "{}")
	if err != nil {
		t.Fatalf("add schedule: %v", err)
	}

	// Add a scheduled todo for a specific date.
	todo := s.AddScheduledTodo("Daily task", "2026-02-07", "body", sc.ID)
	if todo.ID == 0 {
		t.Fatal("expected non-zero todo ID")
	}

	// TodoExistsForSchedule should return true for that date.
	if !s.TodoExistsForSchedule(sc.ID, "2026-02-07") {
		t.Error("TodoExistsForSchedule should return true for existing schedule+date")
	}

	// TodoExistsForSchedule should return false for a different date.
	if s.TodoExistsForSchedule(sc.ID, "2026-02-08") {
		t.Error("TodoExistsForSchedule should return false for different date")
	}

	// Adding a second todo for the same schedule+date should fail (UNIQUE constraint).
	dup := s.AddScheduledTodo("Duplicate task", "2026-02-07", "body2", sc.ID)
	if dup.ID != 0 {
		t.Error("duplicate scheduled todo should fail (return zero ID)")
	}
}

func TestScheduleCascadeOnTemplateDelete(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	tpl, err := s.AddTemplate("Cascade Test", "content")
	if err != nil {
		t.Fatalf("add template: %v", err)
	}

	_, err = s.AddSchedule(tpl.ID, "weekly", "friday", "{}")
	if err != nil {
		t.Fatalf("add schedule: %v", err)
	}

	// Verify schedule exists.
	if len(s.ListSchedules()) != 1 {
		t.Fatal("expected 1 schedule before delete")
	}

	// Delete template -- should cascade to delete schedules.
	s.DeleteTemplate(tpl.ID)

	// Verify schedule is gone.
	remaining := s.ListSchedules()
	if len(remaining) != 0 {
		t.Errorf("expected 0 schedules after template delete, got %d", len(remaining))
	}
}

func TestScheduleSetNullOnDelete(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	tpl, err := s.AddTemplate("SetNull Test", "content")
	if err != nil {
		t.Fatalf("add template: %v", err)
	}

	sc, err := s.AddSchedule(tpl.ID, "daily", "", "{}")
	if err != nil {
		t.Fatalf("add schedule: %v", err)
	}

	// Add a scheduled todo.
	todo := s.AddScheduledTodo("Linked task", "2026-03-01", "body", sc.ID)
	if todo.ID == 0 {
		t.Fatal("expected non-zero todo ID")
	}
	if todo.ScheduleID != sc.ID {
		t.Errorf("expected schedule_id %d, got %d", sc.ID, todo.ScheduleID)
	}

	// Delete the schedule -- should SET NULL on the todo's schedule_id.
	s.DeleteSchedule(sc.ID)

	// Verify todo still exists but ScheduleID is 0 (NULL).
	found := s.Find(todo.ID)
	if found == nil {
		t.Fatal("todo should still exist after schedule delete")
	}
	if found.ScheduleID != 0 {
		t.Errorf("expected schedule_id 0 after schedule delete, got %d", found.ScheduleID)
	}
}

func TestAddScheduledTodo(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	tpl, err := s.AddTemplate("Scheduled Test", "content")
	if err != nil {
		t.Fatalf("add template: %v", err)
	}

	sc, err := s.AddSchedule(tpl.ID, "monthly", "1", "{}")
	if err != nil {
		t.Fatalf("add schedule: %v", err)
	}

	// Add a scheduled todo for February 2026.
	todo := s.AddScheduledTodo("Monthly review", "2026-02-01", "Review body", sc.ID)
	if todo.ID == 0 {
		t.Fatal("expected non-zero todo ID")
	}

	// Verify it appears in TodosForMonth.
	todos := s.TodosForMonth(2026, time.February)
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo for Feb 2026, got %d", len(todos))
	}

	found := todos[0]
	if found.Text != "Monthly review" {
		t.Errorf("expected text 'Monthly review', got %q", found.Text)
	}
	if found.Body != "Review body" {
		t.Errorf("expected body 'Review body', got %q", found.Body)
	}
	if found.Date != "2026-02-01" {
		t.Errorf("expected date '2026-02-01', got %q", found.Date)
	}
	if found.ScheduleID != sc.ID {
		t.Errorf("expected schedule_id %d, got %d", sc.ID, found.ScheduleID)
	}
	if found.ScheduleDate != "2026-02-01" {
		t.Errorf("expected schedule_date '2026-02-01', got %q", found.ScheduleDate)
	}
	if found.Done {
		t.Error("scheduled todo should not be done")
	}
	if found.SortOrder == 0 {
		t.Error("scheduled todo should have non-zero sort_order")
	}
}

func TestDatePrecision(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	// Day-level todo
	day := s.Add("Day task", "2026-03-15", "day", 0)
	if day.DatePrecision != "day" {
		t.Errorf("day todo: want precision 'day', got %q", day.DatePrecision)
	}

	// Month-level todo
	month := s.Add("Month task", "2026-03-01", "month", 0)
	if month.DatePrecision != "month" {
		t.Errorf("month todo: want precision 'month', got %q", month.DatePrecision)
	}
	if month.Date != "2026-03-01" {
		t.Errorf("month todo: want date '2026-03-01', got %q", month.Date)
	}

	// Year-level todo
	year := s.Add("Year task", "2026-01-01", "year", 0)
	if year.DatePrecision != "year" {
		t.Errorf("year todo: want precision 'year', got %q", year.DatePrecision)
	}
	if year.Date != "2026-01-01" {
		t.Errorf("year todo: want date '2026-01-01', got %q", year.Date)
	}

	// Floating todo
	floating := s.Add("Float", "", "", 0)
	if floating.DatePrecision != "" {
		t.Errorf("floating todo: want precision '', got %q", floating.DatePrecision)
	}
	if floating.Date != "" {
		t.Errorf("floating todo: want date '', got %q", floating.Date)
	}

	// Verify round-trip through Find
	found := s.Find(month.ID)
	if found == nil {
		t.Fatal("month todo not found via Find")
	}
	if found.DatePrecision != "month" {
		t.Errorf("found month todo: want precision 'month', got %q", found.DatePrecision)
	}
}

func TestMonthTodosQuery(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	// Add month-precision todo for March 2026
	s.Add("Month task", "2026-03-01", "month", 0)
	// Add day-precision todo for March 15 2026
	s.Add("Day task", "2026-03-15", "day", 0)

	// MonthTodos should return only the month-precision todo
	monthTodos := s.MonthTodos(2026, time.March)
	if len(monthTodos) != 1 {
		t.Fatalf("MonthTodos(2026, March): want 1, got %d", len(monthTodos))
	}
	if monthTodos[0].Text != "Month task" {
		t.Errorf("MonthTodos: want 'Month task', got %q", monthTodos[0].Text)
	}

	// MonthTodos for April should return empty
	aprilTodos := s.MonthTodos(2026, time.April)
	if len(aprilTodos) != 0 {
		t.Errorf("MonthTodos(2026, April): want 0, got %d", len(aprilTodos))
	}
}

func TestYearTodosQuery(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	// Add year-precision todo for 2026
	s.Add("Year task", "2026-01-01", "year", 0)
	// Add day-precision todo for 2026-03-15
	s.Add("Day task", "2026-03-15", "day", 0)

	// YearTodos should return only the year-precision todo
	yearTodos := s.YearTodos(2026)
	if len(yearTodos) != 1 {
		t.Fatalf("YearTodos(2026): want 1, got %d", len(yearTodos))
	}
	if yearTodos[0].Text != "Year task" {
		t.Errorf("YearTodos: want 'Year task', got %q", yearTodos[0].Text)
	}

	// YearTodos for 2027 should return empty
	nextYear := s.YearTodos(2027)
	if len(nextYear) != 0 {
		t.Errorf("YearTodos(2027): want 0, got %d", len(nextYear))
	}
}

func TestDayQueriesExcludeFuzzy(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	// Add day-level todo in March 2026
	s.Add("Day task", "2026-03-15", "day", 0)
	// Add month-level todo in March 2026
	s.Add("Month task", "2026-03-01", "month", 0)
	// Add year-level todo in 2026
	s.Add("Year task", "2026-01-01", "year", 0)

	// TodosForMonth should return only day-level todo
	monthTodos := s.TodosForMonth(2026, time.March)
	if len(monthTodos) != 1 {
		t.Fatalf("TodosForMonth: want 1, got %d", len(monthTodos))
	}
	if monthTodos[0].Text != "Day task" {
		t.Errorf("TodosForMonth: want 'Day task', got %q", monthTodos[0].Text)
	}

	// IncompleteTodosPerDay should count only day-level todo
	incomplete := s.IncompleteTodosPerDay(2026, time.March)
	if incomplete[15] != 1 {
		t.Errorf("IncompleteTodosPerDay[15]: want 1, got %d", incomplete[15])
	}
	// Day 1 should not be counted (month-precision todo lives on day 1 but should be excluded)
	if incomplete[1] != 0 {
		t.Errorf("IncompleteTodosPerDay[1]: want 0, got %d", incomplete[1])
	}

	// TotalTodosPerDay should count only day-level todo
	total := s.TotalTodosPerDay(2026, time.March)
	if total[15] != 1 {
		t.Errorf("TotalTodosPerDay[15]: want 1, got %d", total[15])
	}
	if total[1] != 0 {
		t.Errorf("TotalTodosPerDay[1]: want 0, got %d", total[1])
	}

	// TodosForDateRange should exclude fuzzy todos
	rangeTodos := s.TodosForDateRange("2026-01-01", "2026-12-31")
	if len(rangeTodos) != 1 {
		t.Fatalf("TodosForDateRange: want 1, got %d", len(rangeTodos))
	}
	if rangeTodos[0].Text != "Day task" {
		t.Errorf("TodosForDateRange: want 'Day task', got %q", rangeTodos[0].Text)
	}
}

func TestHighestPriorityPerDay(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	// Day 10: P2 and P3 -- highest (lowest number) should be P2
	s.Add("Task A", "2026-04-10", "day", 2)
	s.Add("Task B", "2026-04-10", "day", 3)

	// Day 12: P1 only
	s.Add("Task C", "2026-04-12", "day", 1)

	// Day 15: P4 only
	s.Add("Task D", "2026-04-15", "day", 4)

	// Day 20: completed P1 -- should be excluded
	done := s.Add("Task E", "2026-04-20", "day", 1)
	s.Toggle(done.ID)

	// Day 20: also has a non-prioritized todo -- should be excluded (priority 0)
	s.Add("Task F", "2026-04-20", "day", 0)

	// Day 25: non-prioritized todo only -- should not appear
	s.Add("Task G", "2026-04-25", "day", 0)

	result := s.HighestPriorityPerDay(2026, time.April)

	// Day 10: P2 wins (lower number = higher priority)
	if result[10] != 2 {
		t.Errorf("day 10: want priority 2, got %d", result[10])
	}

	// Day 12: P1
	if result[12] != 1 {
		t.Errorf("day 12: want priority 1, got %d", result[12])
	}

	// Day 15: P4
	if result[15] != 4 {
		t.Errorf("day 15: want priority 4, got %d", result[15])
	}

	// Day 20: completed P1 excluded, non-prioritized excluded -> no entry
	if _, ok := result[20]; ok {
		t.Errorf("day 20: should not have entry (completed P1 + non-prioritized), got %d", result[20])
	}

	// Day 25: non-prioritized only -> no entry
	if _, ok := result[25]; ok {
		t.Errorf("day 25: should not have entry (non-prioritized only), got %d", result[25])
	}

	// Empty month: no entries
	empty := s.HighestPriorityPerDay(2026, time.January)
	if len(empty) != 0 {
		t.Errorf("empty month: want 0 entries, got %d", len(empty))
	}
}

func TestPriorityRoundtrip(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	// Add with priority 1
	todo := s.Add("Urgent task", "2026-03-15", "day", 1)
	if todo.Priority != 1 {
		t.Errorf("Add returned Priority=%d, want 1", todo.Priority)
	}

	// Find verifies DB roundtrip
	found := s.Find(todo.ID)
	if found == nil {
		t.Fatal("Find returned nil")
	}
	if found.Priority != 1 {
		t.Errorf("Find after Add: Priority=%d, want 1", found.Priority)
	}

	// Update priority to 3
	s.Update(todo.ID, "Urgent task", "2026-03-15", "day", 3)
	found = s.Find(todo.ID)
	if found == nil {
		t.Fatal("Find after Update returned nil")
	}
	if found.Priority != 3 {
		t.Errorf("Find after Update: Priority=%d, want 3", found.Priority)
	}

	// Todos() also returns correct priority
	all := s.Todos()
	var match *Todo
	for i := range all {
		if all[i].ID == todo.ID {
			match = &all[i]
			break
		}
	}
	if match == nil {
		t.Fatal("todo not found in Todos()")
	}
	if match.Priority != 3 {
		t.Errorf("Todos() Priority=%d, want 3", match.Priority)
	}
}

func TestPriorityDefaultZero(t *testing.T) {
	s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer s.Close()

	todo := s.Add("Normal task", "2026-03-15", "day", 0)
	if todo.Priority != 0 {
		t.Errorf("Add with priority 0: got Priority=%d", todo.Priority)
	}

	found := s.Find(todo.ID)
	if found == nil {
		t.Fatal("Find returned nil")
	}
	if found.Priority != 0 {
		t.Errorf("Find: Priority=%d, want 0", found.Priority)
	}
}

func TestPriorityHelpers(t *testing.T) {
	tests := []struct {
		priority      int
		hasPriority   bool
		priorityLabel string
	}{
		{0, false, ""},
		{1, true, "P1"},
		{2, true, "P2"},
		{3, true, "P3"},
		{4, false, ""},
		{-1, false, ""},
		{5, false, ""},
	}

	for _, tc := range tests {
		todo := Todo{Priority: tc.priority}
		if got := todo.HasPriority(); got != tc.hasPriority {
			t.Errorf("Priority %d: HasPriority()=%v, want %v", tc.priority, got, tc.hasPriority)
		}
		if got := todo.PriorityLabel(); got != tc.priorityLabel {
			t.Errorf("Priority %d: PriorityLabel()=%q, want %q", tc.priority, got, tc.priorityLabel)
		}
	}
}
