---
phase: 14-database-backend
plan: 01
subsystem: database
tags: [go, interface, store, dependency-injection, decoupling]

# Dependency graph
requires:
  - phase: 13-search-filter
    provides: search package consuming store
provides:
  - TodoStore interface in store package
  - All consumers depend on interface, not concrete *Store
affects: [14-02 SQLite storage, 14-03 migration]

# Tech tracking
tech-stack:
  added: []
  patterns: [interface-based dependency injection for store layer]

key-files:
  created: []
  modified:
    - internal/store/store.go
    - internal/app/model.go
    - internal/calendar/model.go
    - internal/calendar/grid.go
    - internal/todolist/model.go
    - internal/search/model.go

key-decisions:
  - "TodoStore interface placed in store package (same package as implementation) for simplicity"
  - "main.go unchanged -- *Store satisfies TodoStore implicitly at call site"

patterns-established:
  - "TodoStore interface: all store consumers accept store.TodoStore, not *store.Store"
  - "Compile-time interface check: var _ TodoStore = (*Store)(nil)"

# Metrics
duration: 1min
completed: 2026-02-06
---

# Phase 14 Plan 01: Store Interface Extraction Summary

**TodoStore interface extracted from concrete Store struct; all 5 consumer packages updated to depend on interface**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-06T20:41:06Z
- **Completed:** 2026-02-06T20:42:06Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Defined TodoStore interface with 16 methods matching all public *Store methods
- Added compile-time interface satisfaction check
- Updated all consumer packages (app, calendar, todolist, search) to use store.TodoStore
- Application compiles and runs identically with zero behavioral changes

## Task Commits

Each task was committed atomically:

1. **Task 1: Define TodoStore interface in store package** - `3db8448` (feat)
2. **Task 2: Change all consumers from *store.Store to store.TodoStore** - `de2276c` (refactor)

## Files Created/Modified
- `internal/store/store.go` - Added TodoStore interface and compile-time check
- `internal/app/model.go` - store field and New() param: *store.Store -> store.TodoStore
- `internal/calendar/model.go` - store field and New() param: *store.Store -> store.TodoStore
- `internal/calendar/grid.go` - RenderWeekGrid st param: *store.Store -> store.TodoStore
- `internal/todolist/model.go` - store field and New() param: *store.Store -> store.TodoStore
- `internal/search/model.go` - store field and New() param: *store.Store -> store.TodoStore

## Decisions Made
- TodoStore interface placed in the store package alongside the concrete implementation, keeping imports simple
- main.go did not need changes since *Store satisfies TodoStore implicitly at the call site where it is passed to app.New()

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- TodoStore interface is ready for SQLite implementation in plan 14-02
- Any new backend only needs to implement store.TodoStore to be a drop-in replacement
- No blockers or concerns

## Self-Check: PASSED

---
*Phase: 14-database-backend*
*Completed: 2026-02-06*
