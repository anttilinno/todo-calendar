---
phase: 16-external-editor
plan: 01
subsystem: editor
tags: [editor, bubbletea, exec-process, temp-file, keybinding]
dependency-graph:
  requires: [15-body-infrastructure]
  provides: [external-editor-integration]
  affects: []
tech-stack:
  added: []
  patterns: [tea.ExecProcess for shell-out, editing flag for View guard]
key-files:
  created:
    - internal/editor/editor.go
  modified:
    - internal/todolist/keys.go
    - internal/todolist/model.go
    - internal/app/model.go
decisions:
  - "POSIX vi fallback (not vim) for $VISUAL/$EDITOR chain"
  - "# title heading in temp file for context; body parsed from below heading"
  - "editing bool flag in app model prevents View() scrollback leak"
  - "Temp file cleaned up after ReadResult, not before"
metrics:
  duration: "2 min"
  completed: "2026-02-06"
---

# Phase 16 Plan 01: External Editor Integration Summary

**External editor lifecycle: 'o' key -> temp .md file -> $VISUAL/$EDITOR/vi -> content diff -> conditional save -> TUI resume with no terminal artifacts.**

## What Was Done

### Task 1: Create editor package (c471913)
Created `internal/editor/editor.go` with:
- **ResolveEditor()**: Checks `$VISUAL`, then `$EDITOR`, falls back to `"vi"`
- **Open()**: Creates temp `.md` file with `# title` heading + body, splits editor string on whitespace for multi-arg support (`code --wait`), returns `tea.ExecProcess` command
- **ReadResult()**: Reads temp file after editor exits, parses body (everything after `# ` heading), compares to original, returns changed flag
- **EditorFinishedMsg**: Carries TodoID, TempPath, OriginalBody, Err for the callback

### Task 2: Wire editor into todolist and app (69a687f)
- Added `OpenEditor` keybinding (`o`) to todolist KeyMap with help text
- Added `OpenEditorMsg` type and emission on `o` keypress (fetches fresh todo from store)
- Added `editing bool` field to app Model
- Added `View()` guard: returns empty string when `editing=true` to prevent alt-screen teardown leak
- Added `OpenEditorMsg` handler: sets `editing=true`, calls `editor.Open()`
- Added `EditorFinishedMsg` handler: restores rendering, reads result, removes temp file, conditionally updates body via `store.UpdateBody()`, refreshes calendar indicators

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| vi fallback (not vim) | POSIX standard; available on all Unix systems |
| # heading in temp file | Gives user context; body parsed from below heading |
| editing flag View guard | Prevents Bubble Tea from rendering TUI content to normal terminal buffer during editor |
| strings.Fields for editor split | Supports `EDITOR="code --wait"` and similar multi-arg editors |
| Temp cleanup after ReadResult | Must read file before deleting it |

## Task Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create editor package | c471913 | internal/editor/editor.go |
| 2 | Wire editor into todolist and app | 69a687f | internal/todolist/keys.go, internal/todolist/model.go, internal/app/model.go |

## Deviations from Plan

None -- plan executed exactly as written.

## Verification Results

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- OpenEditor keybinding added to KeyMap, ShortHelp, FullHelp, HelpBindings
- editing flag guard is first check in View() (before !m.ready)
- EditorFinishedMsg reads temp file before cleanup
- Calendar indicators refresh after body edit

## Next Phase Readiness

This completes v1.4 milestone. All 16 phases (28 plans) are done.
No blockers or concerns.

## Self-Check: PASSED
