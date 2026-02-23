package status

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/theme"
)

func TestFormatStatus_EmptySlice(t *testing.T) {
	got := FormatStatus(nil, theme.Dark())
	if got != "" {
		t.Errorf("FormatStatus(nil) = %q, want empty string", got)
	}
}

func TestFormatStatus_AllCompleted(t *testing.T) {
	todos := []store.Todo{
		{Text: "done task", Done: true, Priority: 1},
	}
	got := FormatStatus(todos, theme.Dark())
	if got != "" {
		t.Errorf("FormatStatus(all done) = %q, want empty string", got)
	}
}

func TestFormatStatus_SinglePendingNoPriority(t *testing.T) {
	todos := []store.Todo{
		{Text: "buy milk", Done: false, Priority: 0},
	}
	th := theme.Dark()
	got := FormatStatus(todos, th)
	// No priority -> AccentFg (#5F5FD7 for dark theme)
	want := "%{F#5F5FD7}\uf46d 1%{F-}"
	if got != want {
		t.Errorf("FormatStatus(no priority) = %q, want %q", got, want)
	}
}

func TestFormatStatus_MultiplePendingHighestPriority(t *testing.T) {
	todos := []store.Todo{
		{Text: "urgent", Done: false, Priority: 1},
		{Text: "low", Done: false, Priority: 3},
	}
	th := theme.Dark()
	got := FormatStatus(todos, th)
	// P1 is highest (lowest number) -> PriorityP1Fg (#FF5F5F for dark theme)
	want := "%{F#FF5F5F}\uf46d 2%{F-}"
	if got != want {
		t.Errorf("FormatStatus(P1+P3) = %q, want %q", got, want)
	}
}

func TestFormatStatus_CompletedTodoIgnored(t *testing.T) {
	todos := []store.Todo{
		{Text: "active P2", Done: false, Priority: 2},
		{Text: "done P1", Done: true, Priority: 1},
	}
	th := theme.Dark()
	got := FormatStatus(todos, th)
	// Only pending count=1, highest pending priority=2 -> PriorityP2Fg (#FFAF5F)
	want := "%{F#FFAF5F}\uf46d 1%{F-}"
	if got != want {
		t.Errorf("FormatStatus(P2 pending, P1 done) = %q, want %q", got, want)
	}
}

func TestFormatStatus_P3Color(t *testing.T) {
	todos := []store.Todo{
		{Text: "medium", Done: false, Priority: 3},
	}
	th := theme.Dark()
	got := FormatStatus(todos, th)
	want := "%{F#5F87FF}\uf46d 1%{F-}"
	if got != want {
		t.Errorf("FormatStatus(P3) = %q, want %q", got, want)
	}
}

func TestFormatStatus_P4Color(t *testing.T) {
	todos := []store.Todo{
		{Text: "low", Done: false, Priority: 4},
	}
	th := theme.Dark()
	got := FormatStatus(todos, th)
	want := "%{F#808080}\uf46d 1%{F-}"
	if got != want {
		t.Errorf("FormatStatus(P4) = %q, want %q", got, want)
	}
}

func TestFormatStatus_MixedPriorityAndNoPriority(t *testing.T) {
	todos := []store.Todo{
		{Text: "no prio", Done: false, Priority: 0},
		{Text: "P3", Done: false, Priority: 3},
		{Text: "also no prio", Done: false, Priority: 0},
	}
	th := theme.Dark()
	got := FormatStatus(todos, th)
	// Highest priority among pending is P3 (priority=3), count=3
	want := "%{F#5F87FF}\uf46d 3%{F-}"
	if got != want {
		t.Errorf("FormatStatus(mixed) = %q, want %q", got, want)
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

func TestWriteStatusFile(t *testing.T) {
	// Use a temp directory to avoid writing to /tmp/.todo_status during tests
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".todo_status")

	content := "%{F#FF5F5F}\uf46d 3%{F-}"
	err := writeStatusFileTo(content, path)
	if err != nil {
		t.Fatalf("writeStatusFileTo() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

func TestWriteStatusFile_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".todo_status")

	// Write initial content
	if err := writeStatusFileTo("old", path); err != nil {
		t.Fatalf("first write error = %v", err)
	}

	// Overwrite
	if err := writeStatusFileTo("new", path); err != nil {
		t.Fatalf("second write error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "new" {
		t.Errorf("file content = %q, want %q", string(got), "new")
	}
}

func TestWriteStatusFile_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".todo_status")

	if err := writeStatusFileTo("", path); err != nil {
		t.Fatalf("writeStatusFileTo() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "" {
		t.Errorf("file content = %q, want empty string", string(got))
	}
}
