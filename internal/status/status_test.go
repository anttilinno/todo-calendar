package status

import (
	"testing"

	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

func TestFormatStatus_EmptySlice(t *testing.T) {
	got := FormatStatus(nil)
	if got != "" {
		t.Errorf("FormatStatus(nil) = %q, want empty string", got)
	}
}

func TestFormatStatus_AllCompleted(t *testing.T) {
	todos := []store.Todo{
		{Text: "done task", Done: true, Priority: 1},
	}
	got := FormatStatus(todos)
	if got != "" {
		t.Errorf("FormatStatus(all done) = %q, want empty string", got)
	}
}

func TestFormatStatus_SinglePending(t *testing.T) {
	todos := []store.Todo{
		{Text: "buy milk", Done: false, Priority: 0},
	}
	got := FormatStatus(todos)
	if got != "1" {
		t.Errorf("FormatStatus(1 pending) = %q, want %q", got, "1")
	}
}

func TestFormatStatus_MultiplePending(t *testing.T) {
	todos := []store.Todo{
		{Text: "urgent", Done: false, Priority: 1},
		{Text: "low", Done: false, Priority: 3},
	}
	got := FormatStatus(todos)
	if got != "2" {
		t.Errorf("FormatStatus(2 pending) = %q, want %q", got, "2")
	}
}

func TestFormatStatus_CompletedTodoIgnored(t *testing.T) {
	todos := []store.Todo{
		{Text: "active", Done: false, Priority: 2},
		{Text: "done", Done: true, Priority: 1},
	}
	got := FormatStatus(todos)
	if got != "1" {
		t.Errorf("FormatStatus(1 pending, 1 done) = %q, want %q", got, "1")
	}
}

func TestFormatStatus_MixedPriorityAndNoPriority(t *testing.T) {
	todos := []store.Todo{
		{Text: "no prio", Done: false, Priority: 0},
		{Text: "P3", Done: false, Priority: 3},
		{Text: "also no prio", Done: false, Priority: 0},
	}
	got := FormatStatus(todos)
	if got != "3" {
		t.Errorf("FormatStatus(3 pending) = %q, want %q", got, "3")
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
