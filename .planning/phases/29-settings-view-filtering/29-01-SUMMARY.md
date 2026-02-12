---
phase: 29-settings-view-filtering
plan: 01
subsystem: ui
tags: [bubbletea, settings, toml, config, toggle]

# Dependency graph
requires:
  - phase: 28-display-indicators
    provides: "Circle indicators on calendar title line and This Month/This Year sections in todolist"
provides:
  - "ShowMonthTodos and ShowYearTodos config fields with TOML persistence"
  - "Settings overlay with 6 options including Show/Hide toggles for fuzzy-date sections"
  - "Visibility gating of This Month and This Year sections in todolist"
  - "Visibility gating of circle indicators on calendar title line"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Boolean toggle options in settings overlay using boolIndex helper"
    - "SetShowFuzzySections setter pattern shared by todolist and calendar models"

key-files:
  created: []
  modified:
    - "internal/config/config.go"
    - "internal/settings/model.go"
    - "internal/todolist/model.go"
    - "internal/app/model.go"
    - "internal/calendar/grid.go"
    - "internal/calendar/model.go"

key-decisions:
  - "Boolean toggles use Show/Hide display labels with true/false config values"
  - "Visibility gating applied at visibleItems() level (todolist) and RenderGrid() level (calendar)"

patterns-established:
  - "boolIndex() helper for mapping bool to option index in settings overlay"
  - "SetShowFuzzySections() setter shared by todolist and calendar models"

# Metrics
duration: 3min
completed: 2026-02-12
---

# Phase 29 Plan 01: Settings View Filtering Summary

**Show/Hide toggles for month and year fuzzy-date sections in settings overlay with live preview and TOML persistence**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-12T13:22:41Z
- **Completed:** 2026-02-12T13:25:29Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Settings overlay now has 6 options: Theme, Country, First Day of Week, Date Format, Show Month Todos, Show Year Todos
- Toggling Show Month Todos to Hide removes the "This Month" section from the todo panel and hides the left circle indicator on the calendar
- Toggling Show Year Todos to Hide removes the "This Year" section from the todo panel and hides the right circle indicator on the calendar
- Settings persist to config.toml and take effect immediately on save

## Task Commits

Each task was committed atomically:

1. **Task 1: Add visibility config fields and settings toggles** - `610e677` (feat)
2. **Task 2: Gate sections in todolist and calendar, wire from app** - `399922d` (feat)

## Files Created/Modified
- `internal/config/config.go` - Added ShowMonthTodos/ShowYearTodos bool fields with TOML tags and true defaults
- `internal/settings/model.go` - Added 2 toggle options with Show/Hide cycling and boolIndex helper
- `internal/todolist/model.go` - Added showMonthTodos/showYearTodos fields, SetShowFuzzySections setter, gating in visibleItems()
- `internal/app/model.go` - Wired visibility settings on init and settings save to both todolist and calendar
- `internal/calendar/grid.go` - Added showMonthTodos/showYearTodos params to RenderGrid, gating circle indicators
- `internal/calendar/model.go` - Added showMonthTodos/showYearTodos fields, SetShowFuzzySections setter, updated RenderGrid call

## Decisions Made
- Boolean toggles use Show/Hide display labels mapping to true/false config values via boolIndex helper
- Visibility gating applied at visibleItems() level for todolist and RenderGrid() level for calendar circles
- Both todolist and calendar models expose identical SetShowFuzzySections(showMonth, showYear bool) setter

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 29 is the last phase in v1.9 milestone (Settings & View Filtering)
- All 3 success criteria met: SET-01 (month toggle), SET-02 (year toggle), SET-03 (accessible in existing overlay with live preview)

## Self-Check: PASSED

All 6 modified files verified present. Both task commits (610e677, 399922d) verified in git log.

---
*Phase: 29-settings-view-filtering*
*Completed: 2026-02-12*
