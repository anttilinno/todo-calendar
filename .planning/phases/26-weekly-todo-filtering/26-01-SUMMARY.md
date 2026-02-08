---
phase: 26-weekly-todo-filtering
plan: 01
subsystem: ui
tags: [bubbletea, sqlite, todolist, calendar, filtering]

# Dependency graph
requires:
  - phase: 23-cleanup-calendar-polish
    provides: weekly view toggle and navigation (h/l/w keys)
provides:
  - TodosForDateRange store method for date-range todo queries
  - Week-aware todo panel filtering synced to calendar view mode
  - syncTodoView app wiring that coordinates calendar and todolist state
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "syncTodoView pattern: centralized view-mode-aware todo sync replacing scattered SetViewMonth calls"
    - "Conditional visibleItems: week filter state selects between date-range and month queries"

key-files:
  created: []
  modified:
    - internal/store/iface.go
    - internal/store/sqlite.go
    - internal/store/todo.go
    - internal/calendar/model.go
    - internal/todolist/model.go
    - internal/app/model.go
    - internal/recurring/generate_test.go

key-decisions:
  - "Week filter uses ISO date strings (not time.Time) to match existing store patterns"
  - "syncTodoView replaces scattered SetViewMonth calls for single point of view-mode logic"
  - "Search jump clears week filter to return to monthly context"

patterns-established:
  - "syncTodoView: centralized calendar-to-todolist sync respecting view mode"

# Metrics
duration: 3min
completed: 2026-02-08
---

# Phase 26 Plan 01: Weekly Todo Filtering Summary

**Date-range store query with week-aware todo panel filtering synced to calendar weekly/monthly view mode toggle**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-08T10:05:54Z
- **Completed:** 2026-02-08T10:08:56Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- TodosForDateRange method on store interface and SQLite implementation for week-scoped queries
- Week filter state in todolist model with conditional visibleItems logic (week header vs month header)
- syncTodoView helper in app model replacing 3 scattered SetViewMonth calls with view-mode-aware sync
- Search jump clears week filter to maintain monthly context on navigation

## Task Commits

Each task was committed atomically:

1. **Task 1: Add date-range query and model support** - `7957a04` (feat)
2. **Task 2: Wire app model to sync weekly view with todo filtering** - `d0a0a86` (feat)

## Files Created/Modified
- `internal/store/iface.go` - Added TodosForDateRange to TodoStore interface
- `internal/store/sqlite.go` - SQLite implementation of date-range query
- `internal/store/todo.go` - InDateRange helper on Todo struct
- `internal/calendar/model.go` - WeekStart() public getter for week boundary
- `internal/todolist/model.go` - weekFilterStart/End fields, SetWeekFilter/ClearWeekFilter, conditional visibleItems
- `internal/app/model.go` - syncTodoView helper, replaced 3 SetViewMonth sites, ClearWeekFilter on search jump
- `internal/recurring/generate_test.go` - Added TodosForDateRange stub to fakeStore

## Decisions Made
- Week filter uses ISO date strings to match existing store date patterns (not time.Time)
- syncTodoView centralizes view-mode logic instead of duplicating conditionals at each call site
- Search jump explicitly clears week filter since it navigates by month

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added TodosForDateRange stub to fakeStore in tests**
- **Found during:** Task 1
- **Issue:** Adding TodosForDateRange to the TodoStore interface broke the fakeStore in generate_test.go
- **Fix:** Added stub method `func (f *fakeStore) TodosForDateRange(startDate, endDate string) []store.Todo { return nil }`
- **Files modified:** internal/recurring/generate_test.go
- **Verification:** `go test ./...` passes
- **Committed in:** 7957a04 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary for interface satisfaction. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Weekly todo filtering complete and functional
- All 4 success criteria met: week filtering, floating todos visible, h/l updates, w toggle restore

---
*Phase: 26-weekly-todo-filtering*
*Completed: 2026-02-08*
