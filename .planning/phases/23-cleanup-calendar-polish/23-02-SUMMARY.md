---
phase: 23
plan: 02
subsystem: todolist
tags: [cleanup, keybindings, dead-code-removal]
depends_on:
  requires: []
  provides: ["clean todolist model without AddDated/TemplateUse flows"]
  affects: ["phase-24 (unified add form builds on clean model)"]
tech-stack:
  added: []
  patterns: []
key-files:
  created: []
  modified:
    - internal/todolist/keys.go
    - internal/todolist/model.go
decisions:
  - id: CLN-02-IMPL
    decision: "Removed tmpl import entirely since only template-use flow used it; template-create flow (T key) uses store.AddTemplate directly"
    rationale: "No remaining callers of tmpl.ExtractPlaceholders or tmpl.ExecuteTemplate in todolist package"
metrics:
  duration: "4 min"
  completed: "2026-02-07"
---

# Phase 23 Plan 02: Remove AddDated and TemplateUse Keybindings Summary

Removed A (AddDated) and t (TemplateUse) key bindings plus 328 lines of dead code from the todolist package, clearing the way for Phase 24's unified add form.

## What Was Done

### Task 1: Remove AddDated and TemplateUse keybindings and dead code

**keys.go changes:**
- Removed `AddDated` and `TemplateUse` fields from `KeyMap` struct
- Removed from `ShortHelp()` and `FullHelp()` return slices
- Removed initialization blocks from `DefaultKeyMap()`

**model.go mode enum cleanup:**
- Removed `dateInputMode`, `templateSelectMode`, `placeholderInputMode` constants
- Remaining modes: normalMode, inputMode, editMode, filterMode, templateNameMode, templateContentMode

**model.go Model struct cleanup (10 fields removed):**
- `addingDated bool`
- `pendingText string`
- `templates []store.Template`
- `templateCursor int`
- `pendingTemplate *store.Template`
- `placeholderNames []string`
- `placeholderIndex int`
- `placeholderValues map[string]string`
- `pendingBody string`
- `fromTemplate bool`

**Removed functions (3 entire handler functions):**
- `updateDateInputMode` (~35 lines)
- `updateTemplateSelectMode` (~65 lines)
- `updatePlaceholderInputMode` (~30 lines)

**Simplified existing code:**
- `updateNormalMode`: removed AddDated and TemplateUse case blocks
- `updateInputMode`: removed all addingDated branches and fromTemplate checks; now only handles simple floating add
- `HelpBindings()` and `AllHelpBindings()`: removed dead mode cases and AddDated/TemplateUse from normalMode
- `Update()`: removed dateInputMode/templateSelectMode/placeholderInputMode dispatch
- `View()`: removed dead modes from edit view switch
- `editView()`: removed templateSelectMode, placeholderInputMode, and addingDated rendering blocks
- `clearTemplateState()`: simplified to only reset `pendingTemplateName`

**Import cleanup:**
- Removed `tmpl` import (only used by removed template-use flow)

**Net result:** -328 lines, 2 key bindings removed, 3 modes removed, 10 struct fields removed, 3 functions removed.

## Preserved Flows

- **T (TemplateCreate):** templateNameMode + templateContentMode flow intact
- **a (Add):** inputMode flow intact (simplified, floating-only)
- **e (Edit):** editMode with title+date+body intact

## Task Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Remove AddDated and TemplateUse keybindings and dead code | 87ba1d4 | internal/todolist/keys.go, internal/todolist/model.go |

## Verification Results

- `go build ./...` -- passes
- `go vet ./...` -- passes
- `go test ./...` -- passes
- No references to AddDated, TemplateUse, dateInputMode, templateSelectMode, placeholderInputMode, addingDated, or fromTemplate remain
- TemplateCreate (T key) flow preserved (templateNameMode + templateContentMode)
- Add (a key) flow preserved (inputMode)
- Edit (e key) flow preserved (editMode)

## Deviations from Plan

None -- plan executed exactly as written.

## Next Phase Readiness

Phase 24 (unified add form) can now build on a clean todolist model without legacy AddDated or TemplateUse flows to conflict with. The `a` key currently opens a simple title-only input; Phase 24 will replace this with a multi-field form.

## Self-Check: PASSED
