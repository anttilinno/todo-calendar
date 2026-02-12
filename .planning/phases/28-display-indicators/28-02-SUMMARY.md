---
phase: 28-display-indicators
plan: 02
subsystem: ui
tags: [lipgloss, calendar, indicators, fuzzy-date]

# Dependency graph
requires:
  - phase: 27-date-precision-input
    provides: "MonthTodos/YearTodos store methods, date_precision column"
provides:
  - "Circle indicators on calendar title line for month/year fuzzy todo status"
  - "FuzzyPending and FuzzyDone styles in calendar.Styles"
  - "fuzzyStatus helper function for todo slice status evaluation"
affects: [29-smart-listing]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Circle indicator rendering reusing existing theme roles"]

key-files:
  created: []
  modified:
    - "internal/calendar/styles.go"
    - "internal/calendar/grid.go"
    - "internal/calendar/model.go"

key-decisions:
  - "Reuse PendingFg/CompletedCountFg theme roles for fuzzy indicators (no new theme roles)"
  - "nil-safe store parameter in RenderGrid for backward compatibility"
  - "Visible-width centering using pre-computed character count rather than lipgloss.Width()"

patterns-established:
  - "fuzzyStatus helper: reusable pattern for evaluating todo slice completion status"

# Metrics
duration: 2min
completed: 2026-02-12
---

# Phase 28 Plan 02: Calendar Circle Indicators Summary

**Circle indicators on calendar title showing month-todo (left) and year-todo (right) fuzzy status using red/green coloring**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-12T12:01:32Z
- **Completed:** 2026-02-12T12:03:08Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Added FuzzyPending (red) and FuzzyDone (green) styles reusing existing theme color roles
- Calendar title line renders colored circle indicators for month and year fuzzy-date todo status
- Circles absent when no fuzzy todos exist, red when pending, green when all done
- Weekly view intentionally unmodified (fuzzy todos excluded per VIEW-01)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add fuzzy indicator styles to calendar** - `3821367` (feat)
2. **Task 2: Render circle indicators on calendar title line** - `1973471` (feat)

## Files Created/Modified
- `internal/calendar/styles.go` - Added FuzzyPending and FuzzyDone styles to Styles struct and NewStyles constructor
- `internal/calendar/grid.go` - Added fuzzyStatus helper, updated RenderGrid with store param and circle indicator rendering
- `internal/calendar/model.go` - Updated View() call site to pass m.store to RenderGrid

## Decisions Made
- Reused PendingFg and CompletedCountFg theme roles rather than adding new theme color roles (follows Phase 10 pattern)
- Made store parameter nil-safe in RenderGrid for backward compatibility
- Used pre-computed visible character count for centering rather than lipgloss.Width() for simplicity

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Circle indicators render correctly on calendar title line
- All 4 themes supported via existing PendingFg/CompletedCountFg roles
- Ready for Phase 29 (smart listing) which will display fuzzy todos in the todo list

## Self-Check: PASSED

- All 3 modified files exist and contain expected content
- Both task commits verified (3821367, 1973471)
- `go build ./...` passes
- `go test ./...` passes

---
*Phase: 28-display-indicators*
*Completed: 2026-02-12*
