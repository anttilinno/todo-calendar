package store

import (
	"path/filepath"
	"testing"

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
