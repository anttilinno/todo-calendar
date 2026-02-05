# Phase 4 Plan 02: Wire Indicators Through Calendar Model and Update App Layout

**One-liner:** Connected store query to calendar model for live bracket indicators and widened app layout to accommodate the 34-char grid.

## Metadata

- **Phase:** 04-calendar-enhancements
- **Plan:** 02
- **Duration:** ~2 min
- **Completed:** 2026-02-05

## What Was Done

### Task 1: Wire store into calendar model for indicator data
- Added `store *store.Store` and `indicators map[int]int` fields to calendar Model struct
- Updated `New` constructor to accept `*store.Store` as third parameter and initialize indicators
- Added indicator refresh alongside holidays on PrevMonth/NextMonth navigation
- Added `RefreshIndicators()` public method for external callers (app model)
- Replaced `nil` indicators in View() with `m.indicators`
- **Commit:** `a23812c`

### Task 2: Update app model layout width and wire store to calendar
- Changed `calendar.New(provider, mondayStart)` to `calendar.New(provider, mondayStart, s)` to pass store
- Updated `calendarInnerWidth` from 24 to 38 (34-char grid + 4 chars extra space)
- Added `m.calendar.RefreshIndicators()` at end of every Update cycle for immediate indicator updates
- Also added `RefreshIndicators()` in Tab handler to catch pane-switch edge case
- **Commit:** `a0919fd`

## Key Files

### Created
None.

### Modified
- `internal/calendar/model.go` -- Store reference, indicators field, RefreshIndicators method, live data to RenderGrid
- `internal/app/model.go` -- Store passed to calendar, layout width 38, indicator refresh on every update

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Refresh indicators on every Update cycle | Negligible cost (iterating tens of todos), guarantees indicators are always fresh after any mutation |
| Also refresh on Tab switch | Early return in Tab handler bypasses end-of-Update refresh; ensures indicators update when switching panes |
| calendarInnerWidth = 38 | Grid is 34 chars; add 4 chars for consistent inner padding matching previous 24 (was 20 + 4) |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Tab handler skips indicator refresh**
- **Found during:** Task 2
- **Issue:** The Tab key handler returns early (`return m, nil`) before the end-of-Update `RefreshIndicators()` call, meaning pane switches would not refresh indicators
- **Fix:** Added `m.calendar.RefreshIndicators()` call in the Tab handler before the early return
- **Files modified:** `internal/app/model.go`
- **Commit:** `a0919fd`

## Verification

- `go build ./...` passes
- `go vet ./...` passes
- Calendar model has store reference and indicators field
- `calendar.New` accepts store as third parameter
- Indicators refresh on month navigation, pane switch, and after every update cycle
- `calendarInnerWidth` is 38 to accommodate 34-char grid
- RenderGrid receives actual indicator data instead of nil

## Success Criteria Met

- INDI-01: Calendar dates with incomplete todos display bracket indicators `[N]` (store query feeds indicators to RenderGrid)
- INDI-02: Dates with only completed todos show no indicator (IncompleteTodosPerDay omits zero-count days)
- INDI-03: Calendar grid alignment maintained (uniform 4-char cells from Plan 01)
- FDOW-01: User can set `first_day_of_week` in config.toml (from Plan 01)
- FDOW-02: Calendar grid renders with configured first day (mondayStart flows through)
- FDOW-03: Day-of-week header reflects configured start day (from Plan 01)

## Phase 4 Complete

All Phase 4 requirements are satisfied across Plans 01 and 02:
- Plan 01: Config migration, store query method, grid rendering overhaul
- Plan 02: End-to-end wiring of indicators and layout update

Ready for Phase 5 (Todo Editing).
