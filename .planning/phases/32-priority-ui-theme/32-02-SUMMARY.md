---
phase: 32-priority-ui-theme
plan: 02
subsystem: ui
tags: [lipgloss, bubbletea, priority, calendar, search, tui]

# Dependency graph
requires:
  - phase: 32-01-priority-ui-theme
    provides: "PriorityP1Fg-P4Fg theme colors, HighestPriorityPerDay store method, priorityBadgeStyle pattern"
provides:
  - "Priority-colored calendar day indicators (IndicatorP1-P4, TodayIndicatorP1-P4)"
  - "Priority-aware monthly and weekly grid rendering"
  - "Priority badge rendering in search results with fixed 5-char alignment"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Priority cache pattern in RenderWeekGrid for cross-month week spans"
    - "Fixed-width badge slot reused in search (same pattern as todolist)"

key-files:
  created: []
  modified:
    - "internal/calendar/styles.go"
    - "internal/calendar/grid.go"
    - "internal/calendar/model.go"
    - "internal/search/styles.go"
    - "internal/search/model.go"

key-decisions:
  - "RenderWeekGrid uses priority cache (like indicator cache) rather than a parameter, since it queries store directly for cross-month spans"
  - "Search badge uses same rendering order as todolist: cursor > badge > checkbox > text"

patterns-established:
  - "Priority cache pattern: getPriorities closure with monthKey cache, same as getIndicators/getTotals"
  - "Priority badge in search: same fixed 5-char slot pattern as todolist"

# Metrics
duration: 3min
completed: 2026-02-13
---

# Phase 32 Plan 02: Priority Calendar Indicators and Search Badges Summary

**Priority-colored calendar day brackets (P1=red, P2=orange, P3=blue, P4=grey) in both monthly and weekly grids, plus [P1]-[P4] badge rendering in search results**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-13T19:24:32Z
- **Completed:** 2026-02-13T19:27:55Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Added 8 priority indicator styles (IndicatorP1-P4 and TodayIndicatorP1-P4) to calendar Styles struct
- Updated RenderGrid and RenderWeekGrid to color day brackets by highest-priority incomplete todo, falling through to default indicator color for non-prioritized days
- Added priorities map field to calendar Model, refreshed at all 6 mutation points (New, RefreshIndicators, ToggleWeek, PrevMonth, NextMonth, SetYearMonth)
- Added priority badge styles and priorityBadgeStyle helper to search Styles
- Rendered colored [P1]-[P4] badges in search results with fixed 5-char slot alignment matching todolist

## Task Commits

Each task was committed atomically:

1. **Task 1: Calendar priority-aware indicator styles and grid rendering** - `1b07fda` (feat)
2. **Task 2: Search results priority badge rendering** - `a730bd7` (feat)

## Files Created/Modified
- `internal/calendar/styles.go` - Added IndicatorP1-P4 and TodayIndicatorP1-P4 styles initialized from theme priority colors
- `internal/calendar/grid.go` - Priority-aware day cell styling in RenderGrid (via parameter) and RenderWeekGrid (via cache)
- `internal/calendar/model.go` - Added priorities field, refreshed at all 6 data refresh points, passed to RenderGrid
- `internal/search/styles.go` - Added PriorityP1-P4 styles and priorityBadgeStyle helper method
- `internal/search/model.go` - Priority badge rendering in View() result loop with cursor > badge > checkbox > text order

## Decisions Made
- RenderWeekGrid uses a priority cache (getPriorities closure with monthKey) rather than accepting a parameter, because it spans 2 months and queries the store directly -- consistent with existing getIndicators/getTotals pattern
- Search badge follows the same cursor > badge > checkbox > text rendering order as todolist for visual consistency

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Priority visual system is now complete across all views: todo list, calendar (month + week), and search
- All 4 themes (dark, light, nord, solarized) have correct priority colors propagated to calendar and search styles
- Phase 32 is fully complete -- the v2.1 Priorities milestone is done

## Self-Check: PASSED

All 5 modified files verified on disk. Both task commits (1b07fda, a730bd7) verified in git log. All 5 must-have artifact patterns confirmed present. SUMMARY.md exists at expected path.

---
*Phase: 32-priority-ui-theme*
*Completed: 2026-02-13*
