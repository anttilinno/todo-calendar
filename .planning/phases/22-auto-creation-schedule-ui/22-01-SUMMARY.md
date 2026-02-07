---
phase: 22-auto-creation-schedule-ui
plan: 01
subsystem: recurring
tags: [go, tdd, recurring, auto-create, templates, schedules]

# Dependency graph
requires:
  - phase: 21-schedule-schema-crud
    provides: "Schedule CRUD, TodoExistsForSchedule, AddScheduledTodo store methods"
  - phase: 20-template-engine
    provides: "tmpl.ExecuteTemplate for placeholder fill"
provides:
  - "AutoCreate engine generating scheduled todos on app launch"
  - "AutoCreateForDate testable entry point for date-pinned testing"
  - "buildRuleString and parseDefaults helpers"
affects: [22-02, future schedule UI phases]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Testable time injection via AutoCreateForDate(store, today)"]

key-files:
  created:
    - internal/recurring/generate.go
    - internal/recurring/generate_test.go
  modified:
    - main.go

key-decisions:
  - "Template body executed once per schedule (not per date) for efficiency since defaults are constant"
  - "fakeStore in test implements full TodoStore interface with stubs for unused methods"

patterns-established:
  - "AutoCreateForDate pattern: inject time.Time for deterministic test dates"
  - "buildRuleString assembles cadenceType:cadenceValue for ParseRule compatibility"

# Metrics
duration: 2min
completed: 2026-02-07
---

# Phase 22 Plan 01: AutoCreate Engine Summary

**TDD-driven AutoCreate engine iterating schedules with 7-day window, cadence matching, dedup, and placeholder fill wired into main.go startup**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-07T13:12:52Z
- **Completed:** 2026-02-07T13:14:43Z
- **Tasks:** 1 (TDD: test + feat commits)
- **Files modified:** 3

## Accomplishments
- AutoCreate generates scheduled todos for 7-day rolling window (today + 6 days)
- Deduplication via TodoExistsForSchedule prevents duplicate todos on repeated launches
- Template placeholders filled from schedule's stored JSON defaults
- Graceful handling of missing templates and unparseable cadences (skip, no panic)
- 9 test cases covering daily/weekly/monthly matching, dedup, placeholders, edge cases
- Wired into main.go between store init and app.New for startup execution

## Task Commits

Each task was committed atomically:

1. **RED: Failing tests for AutoCreate** - `b40130c` (test)
2. **GREEN: Implement AutoCreate + main.go wiring** - `d935407` (feat)

_TDD task: RED-GREEN cycle, REFACTOR skipped (code already clean)_

## Files Created/Modified
- `internal/recurring/generate.go` - AutoCreate/AutoCreateForDate with buildRuleString, parseDefaults helpers
- `internal/recurring/generate_test.go` - 9 test cases with fakeStore implementing TodoStore
- `main.go` - Added `recurring.AutoCreate(s)` call at line 42 between store init and app.New

## Decisions Made
- Template body is executed once per schedule rather than per date, since placeholder defaults are constant across the window -- avoids redundant template parsing
- fakeStore's AddScheduledTodo marks the entry as existing in its map so within-run dedup works correctly in the fake

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- AutoCreate engine complete, runs on every app launch
- Ready for 22-02 (schedule UI overlay) to provide user-facing schedule management
- All store methods used by AutoCreate already implemented in SQLiteStore from Phase 21

## Self-Check: PASSED

---
*Phase: 22-auto-creation-schedule-ui*
*Completed: 2026-02-07*
