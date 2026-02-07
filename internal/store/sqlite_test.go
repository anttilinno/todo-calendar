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
