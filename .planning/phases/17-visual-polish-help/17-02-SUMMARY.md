---
phase: 17-visual-polish-help
plan: 02
subsystem: ui
tags: [help, keybindings, bubble-tea, lipgloss, tui]

# Dependency graph
requires:
  - phase: 17-01
    provides: Visual polish foundation (checkbox styling, separators)
provides:
  - Mode-aware help bar with ? toggle
  - Short help (5 keys) and full help (all 15 keys) routing
  - Dynamic help height with content pane resizing
  - FullHelp column layout (groups of 5)
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "help.ShowAll toggle for short/full help switching"
    - "Dynamic help height measurement before content layout"
    - "helpKeyMap.FullHelp() column grouping (groups of 5)"

key-files:
  created: []
  modified:
    - internal/app/keys.go
    - internal/app/model.go
    - internal/todolist/model.go

key-decisions:
  - "HELP-short-keys: Show a/x/d/e// as the 5 most-used keys in short mode (CRUD + filter)"
  - "HELP-column-size: Group expanded bindings into columns of 5 for readability"
  - "HELP-reset-on-tab: Reset ShowAll on pane switch to avoid stale expanded state"

patterns-established:
  - "help.ShowAll toggle: ? key toggles between ShortHelp and FullHelp rendering"
  - "Dynamic help height: compute helpBar first, measure with lipgloss.Height(), subtract from available space"

# Metrics
duration: 2min
completed: 2026-02-07
---

# Phase 17 Plan 02: Mode-Aware Help Bar Summary

**? toggle for help bar: 5-key short mode in normal, full 15-key columnar expansion, Enter/Esc only in input modes**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-07T09:58:56Z
- **Completed:** 2026-02-07T10:00:46Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Normal mode help bar reduced from 19 bindings to ~10 (5 todo + app keys + ?)
- ? key toggles expanded help showing all 15 todo bindings in multi-line column layout
- Input modes show only Enter/Esc regardless of help expansion state
- Dynamic help height: content panes shrink/grow as help bar expands/collapses
- FullHelp styles (FullKey, FullDesc, FullSeparator) set for all 4 themes

## Task Commits

Each task was committed atomically:

1. **Task 1: Add ? keybinding and short/full help methods** - `9c49e7d` (feat)
2. **Task 2: Wire ? toggle, expanded help routing, and dynamic help height** - `1f9071e` (feat)

## Files Created/Modified
- `internal/app/keys.go` - Added Help key.Binding (?) to app KeyMap
- `internal/app/model.go` - ? toggle handler, dynamic help height, expanded help routing, FullHelp column grouping, FullHelp styles
- `internal/todolist/model.go` - HelpBindings() returns 5 short keys, AllHelpBindings() returns all 15

## Decisions Made
- Chose a/x/d/e// as the 5 short-mode keys (add, complete, delete, edit, filter) -- these are the most frequent CRUD and discovery operations
- Column size of 5 for FullHelp layout -- balances readability with terminal width
- Reset ShowAll on Tab pane switch to avoid stale expanded state crossing panes

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 17 (Visual Polish & Help) complete
- Ready for Phase 18 (Responsive Layout)

## Self-Check: PASSED

---
*Phase: 17-visual-polish-help*
*Completed: 2026-02-07*
