package status

import (
	"os"
	"path/filepath"
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

func TestWriteStatusFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".todo_status")

	content := "3"
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

	if err := writeStatusFileTo("old", path); err != nil {
		t.Fatalf("first write error = %v", err)
	}

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

func TestRefreshStatusFileEndToEnd(t *testing.T) {
	todos := []store.Todo{
		{Text: "Buy milk", Date: "2026-02-23", Done: false, Priority: 2},
		{Text: "Call dentist", Date: "2026-02-23", Done: true, Priority: 1},
	}
	output := FormatStatus(todos)

	tmpFile := filepath.Join(t.TempDir(), ".todo_status")
	err := writeStatusFileTo(output, tmpFile)
	if err != nil {
		t.Fatalf("writeStatusFileTo failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}

	if string(data) != "1" {
		t.Errorf("file content = %q, want %q", string(data), "1")
	}
}

func TestRefreshStatusFileAllDone(t *testing.T) {
	todos := []store.Todo{
		{Text: "Done task", Date: "2026-02-23", Done: true, Priority: 1},
	}
	output := FormatStatus(todos)

	tmpFile := filepath.Join(t.TempDir(), ".todo_status")
	err := writeStatusFileTo(output, tmpFile)
	if err != nil {
		t.Fatalf("writeStatusFileTo failed: %v", err)
	}

	data, _ := os.ReadFile(tmpFile)
	if string(data) != "" {
		t.Errorf("expected empty file for all-done todos, got %q", string(data))
	}
}
