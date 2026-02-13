---
phase: 32-priority-ui-theme
plan: 01
subsystem: ui
tags: [lipgloss, bubbletea, priority, theme, sqlite, tui]

# Dependency graph
requires:
  - phase: 31-priority-data-layer
    provides: "Priority column in todos table, Todo.HasPriority/PriorityLabel helpers, Add/Update with priority param"
provides:
  - "PriorityP1Fg through PriorityP4Fg theme colors for all 4 themes"
  - "HighestPriorityPerDay store method for calendar integration"
  - "Priority badge styles and priorityBadgeStyle helper in todolist"
  - "Priority edit field (inline selector) in add and edit forms"
  - "Colored [P1]-[P4] badge rendering in todo list with fixed 5-char alignment"
  - "Named field constants for editField (fieldTitle, fieldDate, fieldPriority, fieldBody, fieldTemplate)"
affects: [32-02-priority-calendar-search, calendar, search]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Named editField constants replacing magic numbers"
    - "Inline selector field (left/right arrows) for non-text form fields"
    - "Fixed-width badge slot for column alignment in todo rendering"
    - "priorityBadgeStyle helper method on Styles struct"

key-files:
  created: []
  modified:
    - "internal/theme/theme.go"
    - "internal/store/iface.go"
    - "internal/store/sqlite.go"
    - "internal/store/sqlite_test.go"
    - "internal/todolist/styles.go"
    - "internal/todolist/model.go"
    - "internal/recurring/generate_test.go"

key-decisions:
  - "Priority field uses inline selector (left/right arrows) not textinput"
  - "Priority badge uses fixed 5-char slot ([P1] + space or 5 spaces) for column alignment"
  - "Named field constants (fieldTitle=0..fieldTemplate=4) replace all magic editField numbers"
  - "Completed todos keep colored priority badge with grey strikethrough text"
  - "HighestPriorityPerDay uses MIN(priority) GROUP BY day for efficient single-query lookup"

patterns-established:
  - "Named field constants: fieldTitle, fieldDate, fieldPriority, fieldBody, fieldTemplate"
  - "Inline selector pattern: left/right arrows cycle options, no textinput needed"
  - "Priority badge always occupies 5 chars for alignment consistency"

# Metrics
duration: 6min
completed: 2026-02-13
---

# Phase 32 Plan 01: Priority UI + Theme Summary

**Priority theme colors for 4 themes, HighestPriorityPerDay store query, inline priority selector in edit forms, and colored [P1]-[P4] badge rendering with fixed-width alignment**

## Performance

- **Duration:** 6 min
- **Started:** 2026-02-13T19:15:33Z
- **Completed:** 2026-02-13T19:21:16Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Added PriorityP1Fg through PriorityP4Fg color fields to Theme struct with appropriate hex values for Dark, Light, Nord, and Solarized themes
- Implemented HighestPriorityPerDay SQL query returning day-to-priority mapping for calendar integration (Plan 02)
- Wired priority field into both add and edit forms with inline selector (none/P1/P2/P3/P4, left/right arrows)
- Added colored priority badge rendering in todo list with fixed 5-char slot for column alignment
- Replaced all hardcoded editField magic numbers with named constants throughout model.go

## Task Commits

Each task was committed atomically:

1. **Task 1: Theme priority colors, store HighestPriorityPerDay, and todolist priority styles** - `67ba1c3` (feat)
2. **Task 2: Priority edit field and badge rendering in todolist** - `16dbfb4` (feat)

## Files Created/Modified
- `internal/theme/theme.go` - Added PriorityP1Fg-P4Fg fields to Theme struct and all 4 theme functions
- `internal/store/iface.go` - Added HighestPriorityPerDay method to TodoStore interface
- `internal/store/sqlite.go` - Implemented HighestPriorityPerDay with MIN/GROUP BY query
- `internal/store/sqlite_test.go` - Added TestHighestPriorityPerDay with priority ranking, completed exclusion, empty month tests
- `internal/todolist/styles.go` - Added PriorityP1-P4 styles and priorityBadgeStyle helper method
- `internal/todolist/model.go` - Added editPriority field, named field constants, priority selector, badge rendering, save wiring
- `internal/recurring/generate_test.go` - Added HighestPriorityPerDay stub to fakeStore

## Decisions Made
- Used inline selector (left/right arrows) for priority field rather than textinput -- faster UX for 5 fixed options, prevents invalid input
- Fixed 5-char badge slot ("[P1] " or "     ") ensures text alignment regardless of priority presence
- Named constants replace all magic editField numbers for maintainability and safety during future field additions
- Priority badge keeps its color on completed todos (strikethrough only applies to text, not badge)
- HighestPriorityPerDay query uses `MIN(priority) WHERE priority BETWEEN 1 AND 4` for efficient single-pass lookup

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Priority colors, store query, and todolist rendering are complete
- Plan 02 (calendar priority indicators and search badge rendering) can proceed
- HighestPriorityPerDay method is ready for calendar grid integration

## Self-Check: PASSED

All 7 modified files verified on disk. Both task commits (67ba1c3, 16dbfb4) verified in git log. SUMMARY.md exists at expected path.

---
*Phase: 32-priority-ui-theme*
*Completed: 2026-02-13*
