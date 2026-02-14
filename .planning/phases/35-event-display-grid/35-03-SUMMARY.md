---
phase: 35-event-display-grid
plan: 03
subsystem: ui
tags: [google-calendar, calendar-grid, settings, bubbletea]

requires:
  - phase: 35-event-display-grid
    provides: EndDate field, ExpandMultiDay helper, EventFg theme color, GoogleCalendarEnabled config toggle
  - phase: 35-event-display-grid
    provides: eventItem kind, SetCalendarEvents on todolist, event rendering in todo panel
provides:
  - Calendar grid bracket indicators on days with Google Calendar events
  - hasEventsPerDay helper for per-day event lookup in calendar model
  - Enabled/Disabled cycling toggle for Google Calendar in settings
  - GoogleCalendarEnabled gating in app model for all event data paths
affects: [event-rendering, calendar-grid, settings]

tech-stack:
  added: []
  patterns: [event indicators reuse existing bracket and style pattern, settings cycling toggle for connected services]

key-files:
  created: []
  modified:
    - internal/calendar/model.go
    - internal/calendar/grid.go
    - internal/settings/model.go
    - internal/app/model.go

key-decisions:
  - "Event-only days use default Indicator style (no priority coloring)"
  - "Today with events uses TodayIndicator style (consistent with todo indicators)"
  - "Settings toggle only appears when AuthReady; otherwise action row preserved"

patterns-established:
  - "Service toggle pattern: cycling Enabled/Disabled when connected, action row when not"

duration: 3min
completed: 2026-02-14
---

# Phase 35 Plan 03: Event Grid Indicators and Settings Toggle Summary

**Calendar grid bracket indicators for event days using ExpandMultiDay, and Enabled/Disabled settings toggle gating event display across todo list and grid**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-14T13:12:49Z
- **Completed:** 2026-02-14T13:15:38Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Calendar grid shows bracket indicators on days with Google Calendar events, even without todos
- Event-only days use default Indicator style; today+event uses TodayIndicator
- Settings shows Enabled/Disabled cycling toggle when Google Calendar is AuthReady
- Settings preserves action row (Sign in/Reconnect) behavior when not connected
- App model gates all SetCalendarEvents calls on GoogleCalendarEnabled
- Toggling immediately shows/hides events from both todo list and calendar grid
- Events continue to be fetched in background even when display is disabled

## Task Commits

Each task was committed atomically:

1. **Task 1: Add event indicators to calendar grid** - `1037aaf` (feat)
2. **Task 2: Add settings toggle and app wiring for GoogleCalendarEnabled** - `33e5d92` (feat)

## Files Created/Modified
- `internal/calendar/model.go` - calendarEvents field, SetCalendarEvents setter, hasEventsPerDay helper, hasEvents wiring to grid renderers
- `internal/calendar/grid.go` - hasEvents parameter on RenderGrid and RenderWeekGrid, bracket and style logic for event-only days
- `internal/settings/model.go` - Cycling Enabled/Disabled toggle when AuthReady, Config() includes GoogleCalendarEnabled, SetGoogleAuthState toggle switch
- `internal/app/model.go` - GoogleCalendarEnabled gating in EventsFetchedMsg, syncTodoView, and SettingChangedMsg handlers

## Decisions Made
- Event-only days use default Indicator style (no priority coloring needed since events have no priority concept)
- Today with events uses TodayIndicator style for consistency with existing today+todo indicator behavior
- Settings toggle only appears when AuthReady; action row pattern preserved for Sign in/Reconnect states

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All three plans of Phase 35 complete: event foundations, todo list rendering, and grid indicators
- Google Calendar event display is fully functional with settings toggle control
- No blockers

---
*Phase: 35-event-display-grid*
*Completed: 2026-02-14*
