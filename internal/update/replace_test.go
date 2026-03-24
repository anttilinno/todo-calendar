package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceBinary_Success(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "todo-calendar")

	original := []byte("original-binary-content")
	if err := os.WriteFile(binPath, original, 0755); err != nil {
		t.Fatal(err)
	}

	newContent := []byte("new-binary-content-v2")
	if err := ReplaceBinary(binPath, newContent); err != nil {
		t.Fatalf("ReplaceBinary: %v", err)
	}

	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(newContent) {
		t.Errorf("content = %q, want %q", got, newContent)
	}

	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0755 {
		t.Errorf("permissions = %o, want 0755", perm)
	}
}

func TestReplaceBinary_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "todo-calendar")

	if err := os.WriteFile(binPath, []byte("old"), 0700); err != nil {
		t.Fatal(err)
	}

	if err := ReplaceBinary(binPath, []byte("new")); err != nil {
		t.Fatalf("ReplaceBinary: %v", err)
	}

	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0700 {
		t.Errorf("permissions = %o, want 0700", perm)
	}
}

func TestReplaceBinary_NonExistentPath(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "nonexistent")

	err := ReplaceBinary(binPath, []byte("data"))
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func TestReplaceBinary_ReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "todo-calendar")

	if err := os.WriteFile(binPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	// Make directory read-only so temp file creation fails.
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	err := ReplaceBinary(binPath, []byte("new"))
	if err == nil {
		t.Fatal("expected error for read-only directory")
	}
}

func TestReplaceBinary_EmptyBinary(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "todo-calendar")

	if err := os.WriteFile(binPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	err := ReplaceBinary(binPath, []byte{})
	if err == nil {
		t.Fatal("expected error for empty binary")
	}
}
