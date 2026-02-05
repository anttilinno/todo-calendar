---
phase: 03-todo-management
plan: 01
status: complete

dependency-graph:
  requires:
    - "01-01 (project structure, Go module)"
  provides:
    - "Todo struct with string-based date model"
    - "Store with atomic JSON persistence"
    - "CRUD operations: Add, Toggle, Delete"
    - "Query methods: TodosForMonth, FloatingTodos"
  affects:
    - "03-02 (todo UI pane consumes Store API)"

tech-stack:
  added: []
  patterns:
    - "Atomic file writes (CreateTemp + Sync + Rename)"
    - "XDG-compliant data path via os.UserConfigDir"
    - "String dates to avoid timezone corruption in JSON"

key-files:
  created:
    - "internal/store/todo.go"
    - "internal/store/store.go"
  modified: []

decisions:
  - id: "string-dates"
    choice: "Use string YYYY-MM-DD for dates, not time.Time"
    reason: "Prevents timezone corruption during JSON serialization round-trips"
  - id: "sync-saves"
    choice: "Synchronous Save() on every mutation"
    reason: "Single small JSON file, no need for async; simplicity over optimization"
  - id: "xdg-data-colocation"
    choice: "Store todos.json alongside config.toml in ~/.config/todo-calendar/"
    reason: "Matches existing config path pattern, keeps all app data together"

metrics:
  duration: "1 min"
  completed: "2026-02-05"
---

# Phase 3 Plan 1: Todo Data Model and Persistence Summary

**One-liner:** Todo struct with string dates and atomic JSON Store (CRUD + month/floating queries) using XDG paths

## What Was Done

### Task 1: Todo data model with date-only serialization (9bdffb3)

Created `internal/store/todo.go` with:

- `Todo` struct: ID, Text, Date (string, omitempty), Done, CreatedAt -- all JSON-tagged
- `Data` struct: NextID counter + Todos slice as JSON envelope
- `HasDate()` method for checking if todo has a date assigned
- `InMonth(year, month)` method for calendar month filtering
- `dateFormat` constant ("2006-01-02") for consistent date parsing

### Task 2: JSON store with atomic writes and CRUD operations (5ccd4b1)

Created `internal/store/store.go` with:

- `TodosPath()` -- XDG-compliant path resolution (`~/.config/todo-calendar/todos.json`)
- `NewStore(path)` -- constructor that loads existing data or initializes empty defaults
- `load()` -- gracefully handles missing files, empty files, and valid JSON
- `Save()` -- atomic write: MarshalIndent -> CreateTemp -> Write -> Sync -> Close -> Rename
- `Add(text, date)` -- creates todo with auto-incremented ID, persists, returns new todo
- `Toggle(id)` -- flips Done status and persists
- `Delete(id)` -- removes by ID and persists
- `Todos()` -- returns all todos
- `TodosForMonth(year, month)` -- filtered + sorted by date then ID
- `FloatingTodos()` -- no-date todos sorted by ID

## Verification Results

| Check | Result |
|-------|--------|
| `go build ./internal/store/...` | PASS |
| `go vet ./internal/store/...` | PASS |
| No UI package dependencies | PASS (0 imports of bubbletea/bubbles/lipgloss) |
| omitempty on Date field | PASS |
| Atomic write uses CreateTemp + Rename | PASS |

## Deviations from Plan

None -- plan executed exactly as written.

## Decisions Made

1. **String dates over time.Time** -- Date field stored as `string` with format "YYYY-MM-DD" to avoid timezone corruption during JSON serialization. This was specified in the plan based on research findings.

2. **Synchronous persistence** -- Save() called synchronously on every mutation (Add/Toggle/Delete). For a single small JSON file this is simpler and safer than async patterns.

3. **XDG data colocation** -- todos.json stored alongside config.toml in `~/.config/todo-calendar/` rather than using a separate data directory.

## Next Phase Readiness

Plan 03-02 (Todo UI) can proceed immediately. The Store API provides everything needed:
- `NewStore` + `TodosPath` for initialization
- `Add/Toggle/Delete` for mutations
- `TodosForMonth/FloatingTodos` for display queries
- No breaking changes to existing packages
