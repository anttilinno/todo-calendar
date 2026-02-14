---
phase: 35-event-display-grid
plan: 02
subsystem: ui
tags: [google-calendar, event-display, todolist, bubbletea]

requires:
  - phase: 35-event-display-grid
    provides: EndDate field, ExpandMultiDay helper, EventTime/EventText styles, EventFg theme color
  - phase: 34-event-fetching
    provides: CalendarEvent type, FetchEvents, MergeEvents, EventsFetchedMsg
provides:
  - eventItem kind in todolist for rendering calendar events
  - Event rendering with 24h time prefix or "all day" label
  - Event insertion before todos in dated section (week and month views)
  - Event data flow from app model to todolist via SetCalendarEvents
affects: [35-03, calendar-grid, event-rendering]

tech-stack:
  added: []
  patterns: [non-selectable item kind in todolist, event data piping via setter method]

key-files:
  created: []
  modified:
    - internal/todolist/model.go
    - internal/app/model.go

key-decisions:
  - "Events inserted before todos in dated section for visual priority"
  - "Events hidden during filter mode (non-interactive items are noise when filtering)"

patterns-established:
  - "Non-selectable item kinds skip selectableIndices automatically (no code changes needed)"
  - "SetCalendarEvents called both on fetch completion and on view navigation"

duration: 2min
completed: 2026-02-14
---

# Phase 35 Plan 02: Event Display in Todo List Summary

**eventItem kind with 24h time rendering, non-selectable events above todos in dated section, and app-to-todolist event data wiring**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-14T13:08:45Z
- **Completed:** 2026-02-14T13:10:44Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- eventItem kind added to todolist with CalendarEvent pointer in visibleItem struct
- renderEvent shows "all day" (bold teal) or "HH:MM" 24h format prefix before event summary
- Events inserted before todos in both week-filtered and month views, with correct date filtering
- Empty placeholder only shows when both events and todos are absent for a section
- Events automatically non-selectable (selectableIndices only returns todoItem indices)
- Events hidden during filter mode for clean filtering experience
- App model wires events to todolist on every fetch and every view navigation

## Task Commits

Each task was committed atomically:

1. **Task 1: Add eventItem kind and event rendering to todolist** - `d0fb62a` (feat)
2. **Task 2: Wire event data from app model to todolist** - `ef8afcb` (feat)

## Files Created/Modified
- `internal/todolist/model.go` - eventItem kind, CalendarEvent field on visibleItem, calendarEvents on Model, SetCalendarEvents, renderEvent, event insertion in visibleItems, event case in normalView, filter exclusion
- `internal/app/model.go` - SetCalendarEvents calls in EventsFetchedMsg handler and syncTodoView

## Decisions Made
- Events inserted before todos in dated section for visual priority (events are calendar-driven, todos are user-driven)
- Events hidden during filter mode since they are non-interactive and would be noise when filtering todos

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Events render in todo panel with distinct styling, ready for Plan 03 (calendar grid event indicators)
- No blockers

---
*Phase: 35-event-display-grid*
*Completed: 2026-02-14*
