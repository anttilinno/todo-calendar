package status

import (
	"testing"
	"time"

	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

// mockStore implements the subset of store.TodoStore needed by FormatStatus.
type mockStore struct {
	store.TodoStore
	daily   []store.Todo
	monthly []store.Todo
	yearly  []store.Todo
}

func (m *mockStore) TodosForDateRange(_, _ string) []store.Todo { return m.daily }
func (m *mockStore) MonthTodos(_ int, _ time.Month) []store.Todo { return m.monthly }
func (m *mockStore) YearTodos(_ int) []store.Todo { return m.yearly }

func TestFormatStatus_Empty(t *testing.T) {
	got := FormatStatus(&mockStore{})
	want := `{"daily":0,"monthly":0,"yearly":0}`
	if got != want {
		t.Errorf("FormatStatus(empty) = %q, want %q", got, want)
	}
}

func TestFormatStatus_DailyOnly(t *testing.T) {
	s := &mockStore{
		daily: []store.Todo{{Text: "task", Done: false}},
	}
	got := FormatStatus(s)
	want := `{"daily":1,"monthly":0,"yearly":0}`
	if got != want {
		t.Errorf("FormatStatus = %q, want %q", got, want)
	}
}

func TestFormatStatus_AllCategories(t *testing.T) {
	s := &mockStore{
		daily:   []store.Todo{{Done: false}, {Done: true}},
		monthly: []store.Todo{{Done: false}, {Done: false}},
		yearly:  []store.Todo{{Done: false}, {Done: false}, {Done: false}},
	}
	got := FormatStatus(s)
	want := `{"daily":1,"monthly":2,"yearly":3}`
	if got != want {
		t.Errorf("FormatStatus = %q, want %q", got, want)
	}
}

func TestFormatStatus_AllDone(t *testing.T) {
	s := &mockStore{
		daily:   []store.Todo{{Done: true}},
		monthly: []store.Todo{{Done: true}},
		yearly:  []store.Todo{{Done: true}},
	}
	got := FormatStatus(s)
	want := `{"daily":0,"monthly":0,"yearly":0}`
	if got != want {
		t.Errorf("FormatStatus(all done) = %q, want %q", got, want)
	}
}

func TestPriorityColorHex(t *testing.T) {
	th := theme.Dark()
	tests := []struct {
		priority int
		want     string
	}{
		{1, "#FF5F5F"},
		{2, "#FFAF5F"},
		{3, "#5F87FF"},
		{4, "#808080"},
		{0, "#5F5FD7"},  // fallback to AccentFg
		{99, "#5F5FD7"}, // unknown also falls back
	}
	for _, tt := range tests {
		got := th.PriorityColorHex(tt.priority)
		if got != tt.want {
			t.Errorf("PriorityColorHex(%d) = %q, want %q", tt.priority, got, tt.want)
		}
	}
}
