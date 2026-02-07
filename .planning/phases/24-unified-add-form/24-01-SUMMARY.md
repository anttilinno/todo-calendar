---
phase: 24-unified-add-form
plan: 01
subsystem: ui
tags: [bubbletea, textinput, textarea, form, todo-add]

# Dependency graph
requires:
  - phase: 23-cleanup-calendar-polish
    provides: "Clean codebase with removed obsolete keybindings (A, t)"
provides:
  - "4-field add form (title, date, body, template) in inputMode"
  - "saveAdd() method with store.Add + optional UpdateBody"
  - "Tab cycling across 4 fields with field-aware save/cancel semantics"
  - "templateInput placeholder field ready for Phase 25 picker integration"
affects: [25-template-picker, 26-recurring-from-add]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "inputMode mirrors editMode's multi-field pattern with Tab cycling"
    - "saveAdd() parallels saveEdit() for add-specific flow"

key-files:
  created: []
  modified:
    - "internal/todolist/model.go"

key-decisions:
  - "templateInput uses CharLimit=0 as read-only placeholder for Phase 25"
  - "inputMode 4-field cycle (0-1-2-3-0) extends editMode's 3-field cycle (0-1-2-0)"
  - "Esc from body/template returns to title (matching editMode body Esc pattern)"
  - "Blink forwarding unified for inputMode and editMode via editField switch"

patterns-established:
  - "4-field inputMode form: title(0) -> date(1) -> body(2) -> template(3)"
  - "saveAdd() with store.Add + conditional UpdateBody for body content"

# Metrics
duration: 2min
completed: 2026-02-07
---

# Phase 24 Plan 01: Unified Add Form Summary

**4-field add form replacing single-field inputMode with title/date/body/template Tab cycling, field-aware save semantics, and templateInput placeholder for Phase 25**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-07T15:49:43Z
- **Completed:** 2026-02-07T15:51:46Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Extended inputMode from single title field to 4-field form (title, date, body, template)
- Added saveAdd() method that creates todos via store.Add() with optional UpdateBody for body content
- Tab cycles through all 4 fields; Enter saves from title/date; Ctrl+D saves from body/template
- Help bar dynamically shows confirm vs save based on active field
- Vertical centering disabled for inputMode (now has textarea like editMode)
- templateInput initialized as read-only placeholder for Phase 25 template picker

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend inputMode to 4-field form with save logic** - `d69875a` (feat)
2. **Task 2: Update view rendering and help bindings for multi-field inputMode** - `1450704` (feat)

## Files Created/Modified
- `internal/todolist/model.go` - Extended inputMode with 4-field form, saveAdd(), Tab cycling, view rendering, help bindings

## Decisions Made
- templateInput uses CharLimit=0 as read-only placeholder -- Phase 25 will replace this with a template picker overlay
- Blink/tick forwarding unified for inputMode and editMode using the same editField switch (case 3 is harmless for editMode which only has fields 0-2)
- Esc behavior matches editMode pattern: Esc from deeper fields (body/template) returns to title; Esc from title/date cancels entirely

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- inputMode now supports all 4 fields, ready for Phase 25 template picker integration on field 3
- saveAdd() is in place for Phase 25 to extend with template application logic
- All existing editMode functionality unchanged

## Self-Check: PASSED

---
*Phase: 24-unified-add-form*
*Completed: 2026-02-07*
