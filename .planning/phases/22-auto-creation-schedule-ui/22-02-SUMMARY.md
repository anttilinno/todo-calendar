---
phase: 22-auto-creation-schedule-ui
plan: 02
subsystem: ui
tags: [lipgloss, recurring, indicators, template-overlay]

# Dependency graph
requires:
  - phase: 21-schedule-schema-crud
    provides: "Schedule struct, ListSchedulesForTemplate, ScheduleID on Todo"
  - phase: 20-template-overlay
    provides: "tmplmgr overlay with template list View()"
provides:
  - "[R] recurring indicator on auto-created todos"
  - "Schedule cadence suffix display in template overlay list"
affects: [22-03, 22-04]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Recurring indicator pattern: check ScheduleID > 0 to show [R]"
    - "Schedule label pattern: parse cadence via recurring.ParseRule for display"

key-files:
  created: []
  modified:
    - internal/todolist/model.go
    - internal/todolist/styles.go
    - internal/tmplmgr/model.go
    - internal/tmplmgr/styles.go

key-decisions:
  - "Schedule suffix placed outside SelectedName style so it stays dimmed even on cursor line"
  - "ordinalSuffix handles 11th/12th/13th special cases for correct English"
  - "Fallback to raw CadenceType if ParseRule fails"

patterns-established:
  - "Indicator ordering: text, [+] body, [R] recurring, date"
  - "Schedule display: scheduleLabel() helper for human-readable cadence"

# Metrics
duration: 3min
completed: 2026-02-07
---

# Phase 22 Plan 02: Recurring Indicators Summary

**[R] indicator on recurring todos and schedule cadence suffix (daily/weekdays/Mon-Wed/15th) in template overlay**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-07T13:13:07Z
- **Completed:** 2026-02-07T13:16:00Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Todos with ScheduleID > 0 now show a muted [R] indicator after text and [+], before date
- Template overlay list shows dimmed cadence suffix next to template names with schedules
- All four cadence types display correctly: (daily), (weekdays), (Mon/Wed/Fri), (15th of month)
- Templates without schedules show no suffix

## Task Commits

Each task was committed atomically:

1. **Task 1: [R] recurring indicator on todos** - `7527af9` (feat)
2. **Task 2: Schedule cadence suffix in template overlay list** - `476d8cf` (feat)

## Files Created/Modified
- `internal/todolist/styles.go` - Added RecurringIndicator style (MutedFg)
- `internal/todolist/model.go` - Added [R] rendering in renderTodo() when ScheduleID > 0
- `internal/tmplmgr/styles.go` - Added ScheduleSuffix style (MutedFg)
- `internal/tmplmgr/model.go` - Added scheduleLabel(), ordinalSuffix() helpers and suffix rendering in View()

## Decisions Made
- Schedule suffix placed outside SelectedName style so it remains dimmed even when cursor is on the line
- ordinalSuffix handles teens (11th, 12th, 13th) as special cases for correct English ordinals
- Fallback to raw CadenceType string in parentheses if ParseRule fails (defensive)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Recurring indicators are visible, ready for Plan 03 (schedule CRUD UI in overlay)
- scheduleLabel() helper is in place for any future schedule display needs

---
*Phase: 22-auto-creation-schedule-ui*
*Completed: 2026-02-07*

## Self-Check: PASSED
