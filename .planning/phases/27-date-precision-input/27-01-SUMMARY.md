---
phase: 27-date-precision-input
plan: 01
subsystem: database
tags: [sqlite, migration, date-precision, fuzzy-dates]

# Dependency graph
requires: []
provides:
  - "Todo struct with DatePrecision field and helper methods (IsFuzzy, IsMonthPrecision, IsYearPrecision)"
  - "Migration 6 adding date_precision column to todos table"
  - "Add/Update store methods accepting datePrecision parameter"
  - "MonthTodos and YearTodos query methods for fuzzy-date retrieval"
  - "Day-level queries exclude fuzzy-date todos from calendar indicators"
affects: [27-02-PLAN, 28-display-sections, calendar-indicators]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "date_precision column convention: 'day', 'month', 'year', or '' (floating)"
    - "Fuzzy dates stored as first-of-period: month='YYYY-MM-01', year='YYYY-01-01'"

key-files:
  created: []
  modified:
    - "internal/store/todo.go"
    - "internal/store/iface.go"
    - "internal/store/sqlite.go"
    - "internal/store/sqlite_test.go"
    - "internal/todolist/model.go"
    - "internal/recurring/generate_test.go"

key-decisions:
  - "Fuzzy-date todos excluded from InDateRange (weekly view) at store level per VIEW-01 requirement"
  - "Floating todos get date_precision='' (empty string), not 'day'"
  - "Interface and caller updates pulled into Task 1 for compilation correctness"

patterns-established:
  - "date_precision column: 'day' for dated, 'month'/'year' for fuzzy, '' for floating"
  - "Day-level queries filter with AND date_precision = 'day' to exclude fuzzy"

# Metrics
duration: 4min
completed: 2026-02-12
---

# Phase 27 Plan 01: Date Precision Storage Summary

**SQLite migration adding date_precision column with precision-aware CRUD, MonthTodos/YearTodos queries, and day-level query filtering**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-12T11:18:15Z
- **Completed:** 2026-02-12T11:22:19Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Todo struct extended with DatePrecision field and IsFuzzy/IsMonthPrecision/IsYearPrecision helpers
- Migration 6 adds date_precision column with 'day' default; fixes floating todos to ''
- Add/Update methods accept datePrecision parameter; all callers updated
- Day-level queries (IncompleteTodosPerDay, TotalTodosPerDay, TodosForMonth, TodosForDateRange) exclude fuzzy-date todos
- MonthTodos and YearTodos query methods for retrieving fuzzy-date subsets
- Four new test functions verify precision storage, fuzzy queries, and day-query exclusion

## Task Commits

Each task was committed atomically:

1. **Task 1: Add date_precision column and update Todo struct** - `929eb3a` (feat)
2. **Task 2: Add tests for date precision storage and querying** - `75c8bff` (test)

## Files Created/Modified
- `internal/store/todo.go` - DatePrecision field, helper methods, updated InMonth/InDateRange
- `internal/store/iface.go` - Updated Add/Update signatures, added MonthTodos/YearTodos
- `internal/store/sqlite.go` - Migration 6, precision-aware CRUD, new query methods, day-level filters
- `internal/store/sqlite_test.go` - TestDatePrecision, TestMonthTodosQuery, TestYearTodosQuery, TestDayQueriesExcludeFuzzy
- `internal/todolist/model.go` - Updated saveAdd/saveEdit to pass datePrecision
- `internal/recurring/generate_test.go` - Updated fakeStore for new interface signatures

## Decisions Made
- Fuzzy-date todos excluded from InDateRange at store level (per VIEW-01 requirement, implemented early)
- Floating todos use empty string for date_precision, not 'day'
- Interface and caller updates pulled from Task 2 into Task 1 to ensure compilation (see Deviations)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Interface and caller updates pulled into Task 1**
- **Found during:** Task 1 (verification step)
- **Issue:** Plan placed interface updates in Task 2, but the compile-time check `var _ TodoStore = (*SQLiteStore)(nil)` failed after changing Add/Update signatures in sqlite.go
- **Fix:** Updated iface.go (Add/Update signatures, MonthTodos/YearTodos), todolist/model.go (callers), and recurring/generate_test.go (fakeStore) as part of Task 1
- **Files modified:** internal/store/iface.go, internal/todolist/model.go, internal/recurring/generate_test.go
- **Verification:** `go build ./...` and `go test ./...` both pass
- **Committed in:** 929eb3a (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Task 2 scope reduced to test-only since interface/caller work moved to Task 1. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Store layer fully supports date precision: day, month, year, and floating
- Ready for Plan 02 (segmented date input UI) to compute precision from user input
- Ready for Phase 28 to add display sections for month/year todos

---
*Phase: 27-date-precision-input*
*Completed: 2026-02-12*
