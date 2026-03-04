package todolist

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

func TestAddTodoCtrlDFromBody(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewSQLiteStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	th := theme.Dark()
	m := New(s, th)
	m.SetDateFormat("eu", "02.01.2006", "DD.MM.YYYY")
	m.SetFocused(true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	send := func(msg tea.Msg) {
		m, _ = m.Update(msg)
	}
	keyRune := func(r rune) tea.KeyMsg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	}

	// Press 'a' to enter add mode
	send(keyRune('a'))
	if m.mode != inputMode {
		t.Fatalf("expected inputMode, got %d", m.mode)
	}

	// Type title "A"
	send(keyRune('A'))
	t.Logf("title value: %q", m.input.Value())

	// Tab to date field (segment 0 = day)
	send(tea.KeyMsg{Type: tea.KeyTab})
	t.Logf("editField=%d dateSegFocus=%d", m.editField, m.dateSegFocus)

	// Tab to skip day (leave empty), go to month (segment 1)
	send(tea.KeyMsg{Type: tea.KeyTab})
	t.Logf("editField=%d dateSegFocus=%d", m.editField, m.dateSegFocus)

	// Type "03" for month (auto-advances to year after 2nd digit)
	send(keyRune('0'))
	send(keyRune('3'))
	t.Logf("month value: %q dateSegFocus=%d (should auto-advance to 2)", m.dateSegMonth.Value(), m.dateSegFocus)

	// Type "2026" directly (auto-advance already moved to year segment)
	send(keyRune('2'))
	send(keyRune('0'))
	send(keyRune('2'))
	send(keyRune('6'))
	t.Logf("year value: %q editField=%d dateSegFocus=%d", m.dateSegYear.Value(), m.editField, m.dateSegFocus)

	// Tab to priority
	send(tea.KeyMsg{Type: tea.KeyTab})
	t.Logf("editField=%d (should be %d=priority)", m.editField, fieldPriority)

	// Set high priority (right arrow)
	send(tea.KeyMsg{Type: tea.KeyRight})
	t.Logf("editPriority=%d", m.editPriority)

	// Tab to body
	send(tea.KeyMsg{Type: tea.KeyTab})
	t.Logf("editField=%d (should be %d=body)", m.editField, fieldBody)

	// Verify state before save
	t.Logf("Before Ctrl+D: title=%q day=%q month=%q year=%q priority=%d",
		m.input.Value(), m.dateSegDay.Value(), m.dateSegMonth.Value(), m.dateSegYear.Value(), m.editPriority)

	// Press Ctrl+D to save
	send(tea.KeyMsg{Type: tea.KeyCtrlD})
	t.Logf("After Ctrl+D: mode=%d editError=%q", m.mode, m.editError)

	// Check if todo was added
	todos := s.Todos()
	if len(todos) == 0 {
		t.Fatalf("no todos added! mode=%d editError=%q", m.mode, m.editError)
	}
	t.Logf("Todo added: text=%q date=%q datePrecision=%q priority=%d",
		todos[0].Text, todos[0].Date, todos[0].DatePrecision, todos[0].Priority)
}

// TestAddTodoTabAfterAutoAdvanceSkipsYear reproduces the bug where pressing Tab
// after month auto-advances to year causes the year to be skipped entirely.
func TestAddTodoTabAfterAutoAdvanceSkipsYear(t *testing.T) {
	dir := t.TempDir()
	s, err := store.NewSQLiteStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	th := theme.Dark()
	m := New(s, th)
	m.SetDateFormat("eu", "02.01.2006", "DD.MM.YYYY")
	m.SetFocused(true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})

	send := func(msg tea.Msg) {
		m, _ = m.Update(msg)
	}
	keyRune := func(r rune) tea.KeyMsg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	}

	// Enter add mode, type title
	send(keyRune('a'))
	send(keyRune('A'))

	// Tab to date, skip day, Tab to month
	send(tea.KeyMsg{Type: tea.KeyTab}) // title → date seg0 (day)
	send(tea.KeyMsg{Type: tea.KeyTab}) // day → month (seg1)

	// Type "03" - auto-advances to year (seg2)
	send(keyRune('0'))
	send(keyRune('3'))
	t.Logf("After typing 03: editField=%d dateSegFocus=%d", m.editField, m.dateSegFocus)

	// User presses Tab thinking they need to advance (but auto-advance already did)
	// BUG: This skips year and goes to priority
	send(tea.KeyMsg{Type: tea.KeyTab})
	t.Logf("After extra Tab: editField=%d dateSegFocus=%d", m.editField, m.dateSegFocus)

	// If editField went to priority (2) instead of staying in date with year focused,
	// the user types "2026" into priority (which ignores runes) and year stays empty
	if m.editField != fieldDate {
		t.Errorf("BUG: Tab after auto-advance skipped year! editField=%d (expected %d=fieldDate)", m.editField, fieldDate)
	}
}
