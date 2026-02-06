---
phase: 07-todo-reordering
plan: 01
subsystem: store
tags: [sort-order, reordering, json-persistence, backwards-compatibility]

# Dependency graph
requires:
  - phase: 06-color-themes
    provides: "existing store and todolist component"
provides:
  - "SortOrder field on Todo struct with JSON persistence"
  - "EnsureSortOrder migration for legacy data"
  - "SwapOrder method for reordering two todos"
  - "Updated sort logic using SortOrder as primary key"
  - "Add method assigns SortOrder to new todos"
affects: [07-02 (wire MoveUp/MoveDown keybindings), 08-settings-overlay]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "SortOrder as primary sort key with date/ID tiebreakers"
    - "EnsureSortOrder migration pattern for backwards-compatible field additions"
    - "Sparse sort order values (increments of 10) for future insertion flexibility"

key-files:
  created: []
  modified:
    - "internal/store/todo.go"
    - "internal/store/store.go"

key-decisions:
  - "SortOrder placed after CreatedAt in Todo struct with omitempty for clean legacy JSON"
  - "EnsureSortOrder uses (i+1)*10 spacing for legacy migration"
  - "SwapOrder is silent no-op on missing IDs, consistent with Toggle/Delete pattern"

patterns-established:
  - "Sort order migration on load: EnsureSortOrder called in NewStore after load"
  - "New todos get maxOrder+10, placing them at end of list"

# Metrics
duration: 2min
completed: 2026-02-06
---

# Phase 7 Plan 01: Add SortOrder Field and Store Methods Summary

**SortOrder int field on Todo struct with EnsureSortOrder migration, SwapOrder method, and updated sort comparators using SortOrder as primary key**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-06T07:43:06Z
- **Completed:** 2026-02-06T07:44:50Z
- **Tasks:** 1
- **Files modified:** 2

## Accomplishments
- Added SortOrder field to Todo struct with backwards-compatible JSON tag
- EnsureSortOrder migration assigns unique values to legacy todos on load
- TodosForMonth and FloatingTodos sort by SortOrder first, then existing tiebreakers
- Add method assigns maxOrder+10 so new todos appear at end
- SwapOrder method swaps sort order of two todos by ID and persists

## Task Commits

Each task was committed atomically:

1. **Task 1: Add SortOrder field and update sort logic** - `5823244` (feat)

## Files Created/Modified
- `internal/store/todo.go` - Added SortOrder int field with json:"sort_order,omitempty" tag
- `internal/store/store.go` - Added EnsureSortOrder, SwapOrder methods; updated sort logic in TodosForMonth and FloatingTodos; updated Add to assign SortOrder

## Decisions Made
None - followed plan as specified. All implementation details matched the research patterns exactly.

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- SortOrder field and SwapOrder method are ready for Plan 07-02 to wire MoveUp/MoveDown keybindings
- The todolist model can call store.SwapOrder(id1, id2) when the user presses move keys
- No blockers or concerns

## Self-Check: PASSED

---
*Phase: 07-todo-reordering*
*Completed: 2026-02-06*
