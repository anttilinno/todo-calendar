---
phase: 21-schedule-schema-crud
plan: 02
subsystem: database
tags: [sqlite, migrations, schedules, recurring, crud, fk-cascade, deduplication]

# Dependency graph
requires:
  - phase: 15-sqlite-backend
    provides: "SQLiteStore with TodoStore interface, migration pattern"
  - phase: 19-markdown-templates
    provides: "Template struct and template CRUD methods"
provides:
  - "Schedule struct with cadence and placeholder defaults"
  - "Todo.ScheduleID and Todo.ScheduleDate fields"
  - "TodoStore interface with 7 schedule methods"
  - "SQLite migrations v4 (schedules table) and v5 (todo schedule columns)"
  - "FK CASCADE: template delete cascades to schedules"
  - "FK SET NULL: schedule delete nullifies todo.schedule_id"
  - "UNIQUE dedup index on (schedule_id, schedule_date)"
affects: [22-schedule-ui-autocreation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Schedule FK CASCADE/SET NULL pattern for parent-child-grandchild relationships"
    - "UNIQUE index for deduplication of auto-created todos"
    - "Nullable column scanning with sql.NullInt64/NullString for schedule fields"

key-files:
  modified:
    - "internal/store/todo.go"
    - "internal/store/store.go"
    - "internal/store/sqlite.go"
    - "internal/store/sqlite_test.go"

key-decisions:
  - "Schedule struct uses string CadenceType/CadenceValue for flexibility (weekly/monday, monthly/1, daily/empty)"
  - "PlaceholderDefaults stored as JSON string for arbitrary key-value pairs"
  - "AddScheduledTodo sets schedule_date = date (display date is the dedup key)"

patterns-established:
  - "scanSchedule/scanSchedules helpers mirror scanTodo/scanTodos pattern"
  - "Schedule methods grouped after template methods in interface"

# Metrics
duration: 4min
completed: 2026-02-07
---

# Phase 21 Plan 02: Schedule Schema CRUD Summary

**SQLite schedule table with FK CASCADE to templates, todo dedup index, and 7 CRUD methods for recurring schedule data layer**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-07T12:43:15Z
- **Completed:** 2026-02-07T12:46:50Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Schedule struct and Todo schedule fields defined for recurring todo support
- TodoStore interface extended with 7 schedule methods (AddSchedule, ListSchedules, ListSchedulesForTemplate, DeleteSchedule, UpdateSchedule, TodoExistsForSchedule, AddScheduledTodo)
- SQLite migrations v4 (schedules table with FK CASCADE) and v5 (schedule_id/schedule_date on todos with UNIQUE dedup index)
- All FK behaviors verified: CASCADE deletes schedules when template deleted, SET NULL nullifies schedule_id on todos when schedule deleted
- 5 integration tests covering CRUD, deduplication, FK CASCADE, FK SET NULL, and scheduled todo creation

## Task Commits

Each task was committed atomically:

1. **Task 1: Schedule struct, Todo fields, and interface extension** - `d3a7192` (feat)
2. **Task 2: SQLite migrations v4+v5 and schedule method implementations** - `81bde14` (feat)

## Files Created/Modified
- `internal/store/todo.go` - Added Schedule struct, Todo.ScheduleID and Todo.ScheduleDate fields
- `internal/store/store.go` - Extended TodoStore interface with 7 schedule methods, JSON store stubs
- `internal/store/sqlite.go` - Migrations v4+v5, updated todoColumns/scanTodo, 7 schedule method implementations with scanSchedule/scanSchedules helpers
- `internal/store/sqlite_test.go` - 5 integration tests (TestScheduleCRUD, TestScheduleDeduplication, TestScheduleCascadeOnTemplateDelete, TestScheduleSetNullOnDelete, TestAddScheduledTodo)

## Decisions Made
- Schedule struct uses string CadenceType/CadenceValue for flexibility (e.g., weekly/monday, monthly/1, daily/empty string)
- PlaceholderDefaults stored as JSON string to allow arbitrary key-value pairs without schema changes
- AddScheduledTodo uses the date parameter as both the todo display date and schedule_date dedup key

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Complete schedule data layer ready for Phase 22 to build UI and auto-creation logic
- All 7 TodoStore schedule methods implemented and tested
- No blockers or concerns

## Self-Check: PASSED

---
*Phase: 21-schedule-schema-crud*
*Completed: 2026-02-07*
