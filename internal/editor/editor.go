package editor

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// EditorFinishedMsg is returned by the ExecProcess callback after the editor exits.
type EditorFinishedMsg struct {
	TodoID       int
	TempPath     string
	OriginalBody string
	Err          error
}

// ResolveEditor returns the user's preferred editor by checking $VISUAL,
// then $EDITOR, falling back to "vi" (POSIX default).
func ResolveEditor() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	return "vi"
}

// Open creates a temp .md file with the todo content and launches the user's
// editor via tea.ExecProcess. The returned command suspends the TUI, runs the
// editor, and emits EditorFinishedMsg when the editor exits.
func Open(todoID int, title string, body string) tea.Cmd {
	// Write temp file with .md extension for syntax highlighting.
	f, err := os.CreateTemp("", "todo-calendar-*.md")
	if err != nil {
		return func() tea.Msg {
			return EditorFinishedMsg{TodoID: todoID, Err: err}
		}
	}

	// File content: # title heading, then blank line, then body.
	content := "# " + title + "\n\n" + body
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return func() tea.Msg {
			return EditorFinishedMsg{TodoID: todoID, Err: err}
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return func() tea.Msg {
			return EditorFinishedMsg{TodoID: todoID, Err: err}
		}
	}

	tempPath := f.Name()
	originalBody := body

	// Resolve editor and split on whitespace to support "code --wait" style values.
	parts := strings.Fields(ResolveEditor())
	args := append(parts[1:], tempPath)
	cmd := exec.Command(parts[0], args...)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return EditorFinishedMsg{
			TodoID:       todoID,
			TempPath:     tempPath,
			OriginalBody: originalBody,
			Err:          err,
		}
	})
}

// ReadResult reads the temp file after the editor exits, parses the body
// (everything after the # heading), and reports whether it changed.
// The caller is responsible for removing msg.TempPath after calling this.
func ReadResult(msg EditorFinishedMsg) (newBody string, changed bool, err error) {
	if msg.Err != nil {
		return "", false, msg.Err
	}

	data, err := os.ReadFile(msg.TempPath)
	if err != nil {
		return "", false, err
	}

	body := parseBody(string(data))

	if body == msg.OriginalBody {
		return body, false, nil
	}
	return body, true, nil
}

// parseBody extracts the body from editor file content.
// It skips the first line that starts with "# " (the title heading)
// and returns the rest, trimmed of leading blank lines and trailing whitespace.
// If no heading is found, the entire content is treated as body.
func parseBody(content string) string {
	lines := strings.Split(content, "\n")
	foundHeading := false
	startIdx := 0

	for i, line := range lines {
		if !foundHeading && strings.HasPrefix(line, "# ") {
			foundHeading = true
			startIdx = i + 1
			break
		}
	}

	if !foundHeading {
		return strings.TrimSpace(content)
	}

	// Join everything after the heading line.
	remaining := strings.Join(lines[startIdx:], "\n")

	// Trim leading blank lines and trailing whitespace.
	remaining = strings.TrimLeft(remaining, "\n")
	remaining = strings.TrimRight(remaining, " \t\n")

	return remaining
}
