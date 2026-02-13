---
phase: 31-priority-data-layer
plan: 01
subsystem: database
tags: [sqlite, migration, priority, store-interface]

# Dependency graph
requires: []
provides:
  - "Priority int field on Todo struct with HasPriority() and PriorityLabel() helpers"
  - "TodoStore interface Add/Update with priority int parameter"
  - "SQLite schema v7 with priority INTEGER NOT NULL DEFAULT 0 column"
  - "Full priority roundtrip: Add -> Find -> Update -> Todos"
affects: [32-priority-ui]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Migration v7 follows v6 (date_precision) precedent: ALTER TABLE + PRAGMA user_version"
    - "Priority 0 = no priority, 1-4 = valid priority levels"

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
  - "Priority 0 used as placeholder in callers until Phase 32 wires actual priority"
  - "Existing test Add/Update calls updated in Task 1 (not Task 2) to satisfy go vet"

patterns-established:
  - "Priority field pattern: 0=none, 1-4=valid, checked by HasPriority()"

# Metrics
duration: 3min
completed: 2026-02-13
---

# Phase 31 Plan 01: Priority Data Layer Summary

**SQLite schema v7 with priority column, Todo struct extension, and full store interface update for priority persistence**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-13T18:43:31Z
- **Completed:** 2026-02-13T18:46:53Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- SQLite migration v7 adds priority INTEGER NOT NULL DEFAULT 0 column
- Todo struct extended with Priority field plus HasPriority() and PriorityLabel() helpers
- TodoStore interface Add() and Update() extended with priority int parameter
- All callers updated (todolist model passes 0, recurring fakeStore stubs match)
- 3 new tests verify full priority roundtrip and helper edge cases
- Full test suite passes with 0 failures

## Task Commits

Each task was committed atomically:

1. **Task 1: Schema migration v7, struct extension, and store implementation** - `9dc7855` (feat)
2. **Task 2: Priority roundtrip and helper tests** - `c766bea` (test)

## Files Created/Modified
- `internal/store/todo.go` - Priority field, HasPriority(), PriorityLabel() helpers
- `internal/store/iface.go` - Extended Add/Update interface signatures with priority int
- `internal/store/sqlite.go` - Migration v7, todoColumns, scanTodo, Add, Update, AddScheduledTodo
- `internal/store/sqlite_test.go` - 3 new priority tests + existing tests updated for new signatures
- `internal/todolist/model.go` - saveAdd() and saveEdit() pass priority 0
- `internal/recurring/generate_test.go` - fakeStore Add/Update stubs match new interface

## Decisions Made
- Existing test calls updated in Task 1 (instead of Task 2) because go vet runs type-checking on test files, making it a blocking issue for Task 1 verification
- Priority 0 used as placeholder in todolist model callers until Phase 32 wires actual priority from the edit form

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Moved test Add/Update signature updates from Task 2 to Task 1**
- **Found during:** Task 1 (verification step)
- **Issue:** `go vet ./...` failed because sqlite_test.go still called Add() with 3 args instead of 4; vet runs type checking on test files
- **Fix:** Updated all existing s.Add() calls in sqlite_test.go to pass priority 0, as part of Task 1 instead of Task 2
- **Files modified:** internal/store/sqlite_test.go
- **Verification:** `go vet ./...` passes cleanly
- **Committed in:** 9dc7855 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary for Task 1 verification to pass. No scope creep -- the same changes were planned for Task 2, just executed earlier.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Priority data layer complete and tested, ready for Phase 32 (Priority UI)
- Phase 32 will wire actual priority values from the edit form instead of hardcoded 0
- Priority colors (P1=red, P2=orange, P3=blue, P4=grey) can be mapped in the UI layer

## Self-Check: PASSED

- All 6 modified files exist on disk
- Both commits (9dc7855, c766bea) found in git log
- SUMMARY.md created at expected path

---
*Phase: 31-priority-data-layer*
*Completed: 2026-02-13*
