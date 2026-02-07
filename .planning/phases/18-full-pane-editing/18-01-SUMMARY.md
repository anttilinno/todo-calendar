# Phase 18 Plan 01: Full-Pane Edit Infrastructure Summary

**One-liner:** Mode-branched View() with vertically centered editView() for add/edit todo, plus SetSize pane-aware dimensions and Tab field-switching key binding.

## Task Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add edit styles, dateInput, editField, SetSize, Tab key | a72e124 | styles.go, keys.go, model.go, app/model.go |
| 2 | Refactor View() for mode-branched full-pane rendering | 9b46214 | model.go |

## What Was Built

### Edit Styles (styles.go)
- `EditTitle`: Bold, AccentFg -- used for "Add Todo" / "Edit Todo" headings
- `FieldLabel`: Bold, NormalFg -- used for "Title", "Date" labels
- `EditHint`: MutedFg -- used for "Enter confirm | Esc cancel" hint text

### Key Binding (keys.go)
- `SwitchField`: Tab key for switching between title/date fields in dated add mode

### Model Infrastructure (model.go)
- `dateInput textinput.Model`: Second text input for date field (placeholder "YYYY-MM-DD", prompt "Date: ", limit 10 chars)
- `editField int`: Field focus tracker (0=title, 1=date)
- `SetSize(w, h)`: Pane-aware dimensions replacing WindowSizeMsg broadcast
- `editView()`: Vertically centered form with EditTitle heading, FieldLabel, input, and EditHint
- `normalView()`: Extracted existing list rendering unchanged
- `View()`: Mode-branched dispatch -- inputMode/dateInputMode/editTextMode/editDateMode route to editView(), everything else to normalView()
- `HelpBindings()`/`AllHelpBindings()`: Mode-aware, includes SwitchField for dated modes

### App Integration (app/model.go)
- `syncTodoSize()`: Computes pane dimensions (accounts for help bar, frame, calendar width) and calls todoList.SetSize()
- All WindowSizeMsg handlers updated to call syncTodoSize() instead of broadcasting WindowSizeMsg to todolist
- Removed redundant todolist WindowSizeMsg broadcast from main handler and overlay handlers

## Decisions Made

| ID | Decision | Rationale |
|----|----------|-----------|
| EDIT-vertical-center | Use topPad = (height - lines) / 3 for edit form positioning | Places form in upper-third for comfortable reading, consistent with settings/search overlays but slightly elevated |
| EDIT-single-field-edit | Edit title (e) and edit date (E) show single-field forms, not combined | Matches current UX expectations; combined form deferred to potential future enhancement |
| EDIT-remove-wsm | Remove WindowSizeMsg handler from todolist Update | Prevents terminal-level dimensions from overriding pane-level SetSize values |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed dead WindowSizeMsg handler from todolist Update**
- **Found during:** Task 2
- **Issue:** The todolist Update() had a WindowSizeMsg handler that stored terminal-level dimensions, which would conflict with the new pane-aware SetSize() values if ever triggered
- **Fix:** Removed the handler since app now calls SetSize() directly
- **Files modified:** internal/todolist/model.go
- **Commit:** 9b46214

## Verification Results

1. `go build ./...` -- PASS
2. `go vet ./...` -- PASS
3. EditTitle present in styles.go -- PASS
4. SwitchField present in keys.go -- PASS
5. editView() present in model.go -- PASS
6. SetSize method present in model.go -- PASS
7. todoList.SetSize call present in app/model.go -- PASS
8. Mode-branched View() dispatch present -- PASS

## Duration

~3 minutes

## Self-Check: PASSED
