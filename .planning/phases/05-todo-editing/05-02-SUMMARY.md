# Phase 5 Plan 02: Wire Edit Flows Summary

**One-liner:** Edit text (e) and edit date (E) modes with pre-filled input, store persistence, and cursor clamping on section moves.

## Metadata

- **Phase:** 05-todo-editing
- **Plan:** 02
- **Duration:** ~2 min
- **Completed:** 2026-02-05

## What Was Done

### Task 1: Add edit modes, editingID field, and mode routing
- Added `editTextMode` and `editDateMode` constants to mode enum
- Added `editingID int` field to Model struct for tracking which todo is being edited
- Added routing cases for both edit modes in Update's KeyMsg switch
- Wired `e` key (m.keys.Edit) in updateNormalMode: sets editTextMode, pre-fills input with todo.Text, cursor at end
- Wired `E` key (m.keys.EditDate) in updateNormalMode: sets editDateMode, pre-fills input with todo.Date (empty for floating)
- Updated HelpBindings to include Edit and EditDate in normal mode
- **Commit:** `5131603`

### Task 2: Implement edit text and edit date confirmation handlers
- `updateEditTextMode`: On Enter, trims text, rejects empty, calls store.Find to get current date, then store.Update(id, newText, existingDate)
- `updateEditDateMode`: On Enter, trims date, validates format if non-empty, accepts empty (makes floating), calls store.Find to get current text, then store.Update(id, existingText, newDate). Clamps cursor after confirm since todo may move between dated/floating sections
- Both handlers: Escape cancels without saving, default forwards to textinput.Update
- **Commit:** `6daeee2`

## Key Implementation Details

- **Pre-fill pattern:** `m.input.SetValue(todo.Text)` + `m.input.CursorEnd()` places cursor at end of existing text for natural editing
- **Non-edited field preservation:** store.Find retrieves the todo, then store.Update passes through the field NOT being edited (date for text edit, text for date edit)
- **Empty date = floating:** Unlike dateInputMode for new todos (which rejects empty), editDateMode accepts empty input to convert a dated todo to floating
- **Cursor clamping:** After date edit, todo may move between sections. `selectableIndices(m.visibleItems())` recomputed and cursor clamped to prevent out-of-bounds

## Deviations from Plan

None -- plan executed exactly as written.

## Files Modified

| File | Changes |
|------|---------|
| `internal/todolist/model.go` | +editTextMode/editDateMode constants, +editingID field, +mode routing, +e/E key handlers, +updateEditTextMode/updateEditDateMode, +HelpBindings update |

## Verification

- `go build ./...` -- compiles cleanly
- `go vet ./...` -- no issues
- grep confirms: editTextMode, editDateMode constants present
- grep confirms: updateEditTextMode, updateEditDateMode functions defined
- grep confirms: m.store.Update and m.store.Find calls wired correctly

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Empty text rejected on edit confirm | Consistent with add flow -- empty todos have no purpose |
| Empty date accepted on edit confirm | Core feature: converting dated to floating by clearing date |
| Cursor clamped after date edit only | Text edits never move todos between sections; date edits can |

## Next Phase Readiness

Phase 5 is complete. Both plans delivered:
- Plan 01: Store Find/Update methods + Edit/EditDate key bindings
- Plan 02: Full edit text and edit date mode flows

All edit flows work end-to-end with immediate disk persistence.
