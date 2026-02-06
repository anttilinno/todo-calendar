---
phase: 15-markdown-templates
plan: 03
subsystem: ui
tags: [templates, todolist, placeholder, textarea, template-workflow]

# Dependency graph
requires:
  - phase: 15-01
    provides: Template CRUD (AddTemplate, ListTemplates, DeleteTemplate), UpdateBody, tmpl utilities
  - phase: 15-02
    provides: Preview overlay for viewing template-generated body content
provides:
  - Template selection mode with cursor-based browsing and deletion
  - Placeholder input mode with sequential prompting for {{.Variable}} values
  - Template creation flow (name + multi-line content via textarea)
  - fromTemplate flag wiring UpdateBody() after Add() in both dated and undated flows
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: [textarea.Model for multi-line input, clearTemplateState helper for workflow reset]

key-files:
  created: []
  modified: [internal/todolist/model.go, internal/todolist/keys.go]

key-decisions:
  - "Template selection uses cursor-based navigation (j/k) not text input, matching todo list pattern"
  - "Template content entry uses textarea.Model (multi-line) with Ctrl+D to save, not Enter"
  - "Placeholder values allowed to be empty (empty string fills the placeholder)"
  - "Template use always prompts for todo text after template selection, even for no-placeholder templates"
  - "templateContentMode routes all msg types (not just KeyMsg) for textarea blink/tick support"

patterns-established:
  - "clearTemplateState() pattern: centralized state reset for multi-step workflows"
  - "fromTemplate + pendingBody pattern: deferred body attachment after todo creation"

# Metrics
duration: 3min
completed: 2026-02-06
---

# Phase 15 Plan 03: Template Creation and Usage Flow Summary

**Template selection mode with placeholder prompting, template creation with multi-line textarea, and fromTemplate body attachment wired through Add+UpdateBody**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-06T21:29:48Z
- **Completed:** 2026-02-06T21:33:18Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Added TemplateUse (t) and TemplateCreate (T) keybindings with help text
- Added four new modes: templateSelectMode, placeholderInputMode, templateNameMode, templateContentMode
- Template selection shows cursor-navigable list with name and truncated content preview
- Placeholder input prompts sequentially for each {{.Variable}} with progress indicator (N/M)
- Template creation flow: name via textinput, content via multi-line textarea with Ctrl+D save
- fromTemplate flag triggers UpdateBody() after store.Add() in both dated and undated code paths
- Template deletion in selection mode with cursor clamping after removal
- All template modes cancellable with Esc, clearing state via clearTemplateState()

## Task Commits

Each task was committed atomically:

1. **Task 1: Add template keybindings and new modes to todolist** - `eb52ccd` (feat)
2. **Task 2: Implement template selection, placeholder input, and creation flows** - `c662201` (feat)

## Files Created/Modified
- `internal/todolist/keys.go` - Added TemplateUse and TemplateCreate bindings to KeyMap, ShortHelp, FullHelp, DefaultKeyMap
- `internal/todolist/model.go` - Added 4 new modes, 11 template workflow fields, textarea initialization, 4 update functions (templateSelect, placeholderInput, templateName, templateContent), modified inputMode/dateInputMode for fromTemplate, added View rendering for all new modes, clearTemplateState helper

## Decisions Made
- Template selection uses cursor-based navigation (j/k and Up/Down) rather than text search, matching the existing todo list navigation pattern.
- Multi-line template content uses textarea.Model with Ctrl+D to save, since Enter needs to insert newlines in markdown content.
- Placeholder values are allowed to be empty -- the user can press Enter to skip a placeholder, which fills it with an empty string.
- Template use always requires entering todo text (inputMode) after template selection, ensuring every todo has descriptive text regardless of template body.
- templateContentMode is handled at the top of Update() before the KeyMsg type switch, since textarea needs tick/blink messages.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 15 (Markdown Templates) is now complete: store layer, preview overlay, and template workflow all functional
- Templates can be created (T), selected (t), filled with placeholders, and attached as todo bodies
- Template-generated bodies are viewable via preview (p) from phase 15-02

## Self-Check: PASSED
