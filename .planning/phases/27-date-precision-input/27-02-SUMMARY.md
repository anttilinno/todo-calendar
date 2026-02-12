---
phase: 27-date-precision-input
plan: 02
subsystem: ui
tags: [textinput, segmented-input, date-precision, fuzzy-dates, bubble-tea]

# Dependency graph
requires:
  - "27-01: date_precision column, Add/Update with datePrecision, MonthTodos/YearTodos queries"
provides:
  - "Segmented 3-field date input (dd/mm/yyyy) replacing single date text field"
  - "Format-aware segment ordering (ISO/EU/US) via dateSegmentOrder helper"
  - "Precision derivation from empty segments: year-only, month+year, or full date"
  - "Fuzzy date display: 'March 2026' for month, '2026' for year"
  - "Tab navigation between date segments with auto-advance on full input"
affects: [28-display-sections, settings-date-format]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Segmented input: 3 textinput.Model fields with visual-to-semantic position mapping"
    - "dateSegOrder [3]int maps visual position to semantic (0=day, 1=month, 2=year)"
    - "focusDateSegment/blurAllDateSegments helpers for segment focus management"
    - "deriveDateFromSegments returns (isoDate, precision, errSegPos) tuple"

key-files:
  created: []
  modified:
    - "internal/todolist/model.go"
    - "internal/todolist/styles.go"
    - "internal/app/model.go"

key-decisions:
  - "Segment auto-advance when char limit reached for fluid typing experience"
  - "Backspace on empty segment navigates to previous segment"
  - "Separator characters (- . /) blocked in segment input since they are visual only"
  - "Fuzzy dates shown as human-readable text (March 2026) rather than partial ISO"

patterns-established:
  - "renderFuzzyDate helper for precision-aware date display"
  - "updateDateSegment intercepts keys for segment-level behavior"

# Metrics
duration: 4min
completed: 2026-02-12
---

# Phase 27 Plan 02: Segmented Date Input Summary

**3-segment date input with format-aware ordering, Tab navigation between segments, and precision derivation from empty segments for fuzzy-date creation**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-12T11:24:24Z
- **Completed:** 2026-02-12T11:29:08Z
- **Tasks:** 1
- **Files modified:** 3

## Accomplishments
- Replaced single date text field with 3 segmented inputs (day, month, year) with format-aware visual ordering
- Tab cycles through date segments within the date field before advancing to the next form field
- Empty segments derive correct date precision: all empty = floating, year only = year, year+month = month, all filled = day
- Fuzzy dates display as "March 2026" (month) or "2026" (year) in the todo list
- Auto-advance to next segment when current reaches character limit for fluid number entry
- Hint text below date field explains precision behavior

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement segmented date input with format-aware ordering** - `bdbdae7` (feat)

## Files Created/Modified
- `internal/todolist/model.go` - Replaced dateInput with dateSegDay/Month/Year, added segment helpers, precision derivation, fuzzy date rendering, updated all form handlers
- `internal/todolist/styles.go` - Added DateSeparator style for muted separator chars
- `internal/app/model.go` - Updated SetDateFormat calls to pass format name as first arg

## Decisions Made
- Auto-advance to next segment when char limit reached (2 for day/month, 4 for year) for fluid typing
- Backspace on empty segment navigates backward to previous segment
- Separator characters blocked in segment input since separators are rendered visually between segments
- Fuzzy dates shown as human-readable "March 2026" and "2026" rather than partial ISO dates
- Edit mode pre-populates segments based on existing todo's DatePrecision (day=all 3, month=year+month, year=year only)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 27 complete: date precision storage + segmented input both shipped
- Ready for Phase 28 to add display sections for month/year todos in the todo list view
- All existing tests pass, app compiles cleanly

## Self-Check: PASSED

- All 3 modified files exist on disk
- Task commit bdbdae7 verified in git log
- Key patterns (dateSegDay, DateSeparator, cfg.DateFormat) found in respective files
- `go build ./...` passes with zero errors
- `go test ./...` passes all tests

---
*Phase: 27-date-precision-input*
*Completed: 2026-02-12*
