---
phase: 07-todo-reordering
plan: 02
subsystem: ui
tags: [keybindings, reordering, bubble-tea, help-bar, section-boundary]

# Dependency graph
requires:
  - phase: 07-todo-reordering
    provides: "SortOrder field, SwapOrder method, updated sort logic from plan 01"
provides:
  - "MoveUp (K) and MoveDown (J) keybindings for todo reordering"
  - "Section boundary enforcement preventing cross-section swaps"
  - "Cursor-follows-item behavior after move"
  - "Help bar integration showing move keybindings in normal mode"
affects: [08-settings-overlay]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Shift+letter as variant keybinding (K=move up, J=move down alongside k=up, j=down)"
    - "Section boundary check via HasDate comparison before swap"

key-files:
  created: []
  modified:
    - "internal/todolist/keys.go"
    - "internal/todolist/model.go"

key-decisions:
  - "MoveUp/MoveDown placed after Down in KeyMap struct, matching navigation-then-action ordering"
  - "Section boundary uses HasDate equality check -- simple and sufficient"

patterns-established:
  - "Move handlers follow same guard pattern as Toggle/Delete (check selectable length and cursor bounds)"
  - "Cursor adjustment after swap (decrement on move-up, increment on move-down)"

# Metrics
duration: 1min
completed: 2026-02-06
---

# Phase 7 Plan 02: Wire MoveUp/MoveDown Keybindings Summary

**Shift+K/J keybindings for todo reordering with section boundary enforcement and help bar integration**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-06T07:46:51Z
- **Completed:** 2026-02-06T07:48:02Z
- **Tasks:** 1
- **Files modified:** 2

## Accomplishments
- Added MoveUp (K) and MoveDown (J) key bindings to KeyMap with help text
- Implemented move handlers in updateNormalMode with HasDate section boundary checks
- Cursor follows the moved todo after swap for seamless multi-position moves
- Updated HelpBindings, ShortHelp, and FullHelp to include move keybindings

## Task Commits

Each task was committed atomically:

1. **Task 1: Add MoveUp/MoveDown keybindings and handlers** - `cff17e9` (feat)

## Files Created/Modified
- `internal/todolist/keys.go` - Added MoveUp and MoveDown fields to KeyMap struct, initialized in DefaultKeyMap, included in ShortHelp/FullHelp
- `internal/todolist/model.go` - Added MoveUp/MoveDown case handlers in updateNormalMode with SwapOrder calls and section boundary checks; updated HelpBindings

## Decisions Made
None - followed plan as specified. All implementation details matched the research patterns exactly.

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 7 (Todo Reordering) is fully complete
- All v1.2 reordering success criteria are met:
  1. User can move selected todo up via Shift+K
  2. User can move selected todo down via Shift+J
  3. Custom order persists across restart (via SortOrder from plan 01)
  4. K/J keybindings visible in help bar during normal mode
- Ready to proceed to Phase 8 (Settings Overlay)
- No blockers or concerns

## Self-Check: PASSED

---
*Phase: 07-todo-reordering*
*Completed: 2026-02-06*
