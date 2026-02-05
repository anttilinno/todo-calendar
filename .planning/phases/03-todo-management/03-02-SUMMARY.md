---
phase: "03"
plan: "02"
subsystem: "todo-ui"
tags: ["bubbletea", "todolist", "help-bar", "textinput", "crud", "month-sync"]
dependency-graph:
  requires: ["03-01"]
  provides: ["todo-crud-ui", "help-bar", "calendar-todo-sync"]
  affects: []
tech-stack:
  added: ["bubbles/textinput", "bubbles/help"]
  patterns: ["mode-based-input", "help-keymap-adapter", "month-sync-on-navigation"]
file-tracking:
  key-files:
    created:
      - "internal/todolist/keys.go"
      - "internal/todolist/styles.go"
    modified:
      - "internal/todolist/model.go"
      - "internal/calendar/model.go"
      - "internal/app/model.go"
      - "internal/app/keys.go"
      - "main.go"
      - "go.mod"
      - "go.sum"
decisions:
  - id: "mode-input-isolation"
    description: "Three-mode state machine (normal/input/dateInput) with Enter/Esc intercepted before textinput"
  - id: "quit-suppression"
    description: "During input mode, 'q' goes to textinput; only ctrl+c quits. Tab also suppressed during input."
  - id: "help-keymap-adapter"
    description: "helpKeyMap adapter type aggregates pane-specific + app bindings for help.Model.View()"
  - id: "cursor-over-selectables"
    description: "Cursor index tracks selectable items only, skipping headers and empty placeholders"
metrics:
  duration: "3 min"
  completed: "2026-02-05"
---

# Phase 3 Plan 2: Todo List UI Summary

Full todo management UI with CRUD operations, two-section rendering, context-sensitive help bar, and calendar-todo month synchronization.

## What Was Done

### Task 1: Rewrite todolist package with full CRUD, modes, and rendering
**Commit:** `7c222cf`

- Rewrote `internal/todolist/model.go` from placeholder to full Bubble Tea model with three modes (normal/input/dateInput)
- Model backed by `*store.Store` for Add, Toggle, Delete operations
- Two-section display: dated todos filtered by viewed month under month header, floating todos under "Floating" header
- Cursor navigation over selectable items only, skipping headers and empty placeholders
- Text input via `bubbles/textinput` with Enter/Esc intercepted before forwarding to textinput
- Created `internal/todolist/keys.go` with KeyMap implementing `help.KeyMap` interface
- Created `internal/todolist/styles.go` with styles for completed, cursor, headers, dates, empty states
- Added `Year()`, `Month()`, `Keys()` accessors to `internal/calendar/model.go`

### Task 2: Wire todo UI into app with help bar, store init, and month sync
**Commit:** `dd03aee`

- Updated `main.go` to initialize store from XDG path and pass to `app.New`
- Changed `app.New` signature to accept `*store.Store`, passes to `todolist.New`
- Added `help.Model` to app for context-sensitive help bar rendering
- Created `helpKeyMap` adapter type aggregating pane-specific + app-level bindings
- Calendar month navigation syncs to todo list via `SetViewMonth`
- Input mode isolation: `q` goes to textinput during input, only `ctrl+c` quits
- Tab switch suppressed during text input mode
- Replaced hardcoded status bar string with dynamic help bar

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Three-mode state machine | normalMode for navigation, inputMode for text, dateInputMode for date -- cleanly separates key handling |
| Quit suppression via isInputting check | Simpler than SetEnabled toggling; single check at app level before matching quit binding |
| helpKeyMap adapter | Lightweight struct satisfying help.KeyMap interface; avoids coupling pane keymaps to app |
| Cursor over selectables only | Cursor index maps to todo items, skipping headers/empty; prevents cursor landing on non-interactive rows |
| Enter/Esc intercepted before textinput | Prevents textinput from swallowing confirm/cancel keys; critical for correct mode transitions |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Missing textinput transitive dependency**
- **Found during:** Task 1 build verification
- **Issue:** `bubbles/textinput` imports `github.com/atotto/clipboard` which was not in go.sum
- **Fix:** Ran `go get github.com/charmbracelet/bubbles/textinput@v0.21.1` to add transitive dep
- **Files modified:** go.mod, go.sum
- **Commit:** `7c222cf`

## Verification

- `go build ./...` passes
- `go vet ./...` passes
- App starts with store loaded from disk
- Right pane shows month header + "Floating" header with empty placeholders
- Help bar shows context-sensitive keybindings per focused pane
- Calendar month navigation updates todo filter

## Next Phase Readiness

This is the final plan of the project. All three phases are complete:
- Phase 1: Scaffold with dual-pane layout
- Phase 2: Calendar with holidays
- Phase 3: Todo management with persistence

The application is functionally complete with all planned features delivered.
