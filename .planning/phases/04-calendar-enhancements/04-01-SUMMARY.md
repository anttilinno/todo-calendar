# Phase 4 Plan 01: Config Migration, Store Query, and Grid Overhaul Summary

**One-liner:** Migrated config from bool to string first-day-of-week, added per-day incomplete todo counts to store, and widened calendar grid to uniform 4-char cells with bracket indicator support.

## Metadata

- **Phase:** 04-calendar-enhancements
- **Plan:** 01
- **Duration:** ~2.5 min
- **Completed:** 2026-02-05

## What Was Done

### Task 1: Config migration + store query + main.go update
- Replaced `MondayStart bool` field with `FirstDayOfWeek string` field in Config struct
- Added `MondayStart()` convenience method so downstream code works unchanged
- Updated `main.go` to call `cfg.MondayStart()` method instead of field access
- Added `IncompleteTodosPerDay(year, month)` method to Store returning `map[int]int`
- **Commit:** `f8d644b`

### Task 2: Calendar grid overhaul to 4-char cells with indicator support
- Changed RenderGrid signature to accept `indicators map[int]int` parameter
- Widened all cells from 2 to 4 characters (`" %2d "` or `"[%2d]"`)
- Updated grid width from 20 to 34 chars (7 x 4 + 6 x 1)
- Aligned weekday headers to 34-char grid width
- Added `indicatorStyle` (bold) to styles.go
- Style priority: today > holiday > indicator > normal
- Passed `nil` for indicators in model.go View() to keep project compilable
- **Commit:** `4291e73`

## Key Files

### Created
None.

### Modified
- `internal/config/config.go` -- FirstDayOfWeek string field with MondayStart() method
- `internal/store/store.go` -- IncompleteTodosPerDay method
- `internal/calendar/grid.go` -- 4-char cell grid with indicator support, 34-char width
- `internal/calendar/styles.go` -- indicatorStyle added
- `internal/calendar/model.go` -- Pass nil indicators to RenderGrid
- `main.go` -- cfg.MondayStart() method call

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Break backward compat on config field | Personal-use app; document change rather than add migration code |
| Uniform 4-char cells for all dates | Prevents column misalignment when mixing bracketed and non-bracketed dates |
| Pass nil indicators temporarily | Keeps project compilable between Plan 01 and Plan 02 |
| Bold-only indicator style | Simple placeholder; Phase 6 (Themes) will overhaul all styles |

## Deviations from Plan

None -- plan executed exactly as written.

## Verification

- `go build ./...` passes
- `go vet ./...` passes
- Config struct uses `FirstDayOfWeek` string with `MondayStart()` method
- Store has `IncompleteTodosPerDay` returning day-to-count map
- RenderGrid renders 4-char cells with indicator bracket support
- Weekday headers align with 34-char grid

## Next Plan Readiness

Plan 04-02 will:
- Wire indicators through the calendar model (add store reference, compute indicators)
- Update `calendarInnerWidth` in app/model.go from 24 to ~38
- Replace the `nil` indicators in model.go View() with actual data
- All building blocks are in place for Plan 02 to connect them
