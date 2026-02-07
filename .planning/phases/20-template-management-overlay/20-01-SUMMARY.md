---
phase: 20-template-management-overlay
plan: 01
subsystem: ui
tags: [bubble-tea, overlay, template-management, lipgloss, textinput]

# Dependency graph
requires:
  - phase: 19-pre-built-templates
    provides: "Templates table with seed data, template CRUD on store"
provides:
  - "UpdateTemplate method on TodoStore interface (rename + content update with UNIQUE error)"
  - "tmplmgr overlay package (Model, Update, View, keys, styles)"
  - "CloseMsg, EditTemplateMsg, TemplateUpdatedMsg message types for app integration"
affects: [20-02 (wiring into app.Model), 21 (schedule schema extends templates)]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "tmplmgr overlay pattern: separate package with Model/Update/View following search/settings/preview"
    - "UpdateTemplate returns error for UNIQUE constraint handling in UI"

key-files:
  created:
    - "internal/tmplmgr/model.go"
    - "internal/tmplmgr/keys.go"
    - "internal/tmplmgr/styles.go"
  modified:
    - "internal/store/store.go"
    - "internal/store/sqlite.go"

key-decisions:
  - "Template content shown as raw text (not glamour-rendered) to reveal placeholder syntax"
  - "No delete confirmation dialog, matching existing todo delete pattern"
  - "Rename error shown inline rather than in a separate dialog"

patterns-established:
  - "tmplmgr follows established overlay pattern: separate package, CloseMsg, SetSize, SetTheme, HelpBindings"

# Metrics
duration: 2min
completed: 2026-02-07
---

# Phase 20 Plan 01: Store Extension and Template Manager Overlay Summary

**UpdateTemplate on store interface with UNIQUE error return; tmplmgr overlay package with list navigation, raw content preview, delete, and rename with duplicate name handling**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-07T12:18:58Z
- **Completed:** 2026-02-07T12:21:20Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Extended TodoStore interface with UpdateTemplate(id, name, content) that returns error for UNIQUE constraint violations
- Created tmplmgr package following the established overlay pattern (search, settings, preview)
- Overlay supports list mode (j/k nav, d delete, r rename, e edit) and rename mode (enter confirm, esc cancel)
- Raw template content preview below list (shows placeholder syntax per REQ-21)
- Inline error display for duplicate template name on rename (per Pitfall 7)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add UpdateTemplate to store interface and implementations** - `4aa445a` (feat)
2. **Task 2: Create internal/tmplmgr package with overlay model** - `20b070a` (feat)

## Files Created/Modified
- `internal/tmplmgr/model.go` - Core overlay model with list/rename modes, cursor navigation, delete, rename, edit message emission (239 lines)
- `internal/tmplmgr/keys.go` - KeyMap with j/k/d/r/e/enter/esc bindings and DefaultKeyMap constructor
- `internal/tmplmgr/styles.go` - Themed styles (Title, TemplateName, SelectedName, Separator, Content, Hint, Error, Empty) with NewStyles constructor
- `internal/store/store.go` - Added UpdateTemplate to TodoStore interface and JSON Store stub
- `internal/store/sqlite.go` - Implemented UpdateTemplate with UNIQUE constraint error propagation

## Decisions Made
- Template content is shown as raw text, not glamour-rendered, so users can see `{{.Placeholder}}` syntax (per REQ-21)
- No confirmation dialog on delete, matching the existing todo delete pattern (per REQ-22)
- Rename input is pre-filled with current name and cursor positioned at end (per REQ-23)
- Duplicate name error shown inline below the rename input in HolidayFg (red) color (per Pitfall 7)
- EditTemplateMsg emitted for external editor launch; actual editor wiring deferred to Plan 02

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- tmplmgr package is complete and compiles independently
- Plan 20-02 can wire the overlay into app.Model using the exported message types
- EditTemplateMsg is ready for external editor integration in Plan 02
- UpdateTemplate error return enables proper UNIQUE constraint handling in the UI

## Self-Check: PASSED

---
*Phase: 20-template-management-overlay*
*Completed: 2026-02-07*
