---
phase: 15-markdown-templates
plan: 02
subsystem: ui
tags: [glamour, markdown, preview, viewport, overlay, body-indicator]

# Dependency graph
requires:
  - phase: 15-01
    provides: Todo.Body field, HasBody() method, store layer with body support
provides:
  - Markdown preview overlay with glamour-rendered styled output
  - Body indicator [+] in todo list for todos with non-empty body
  - Preview keybinding (p) wired through todolist -> app -> preview overlay
affects: [15-03 (template creation flow uses preview for body viewing)]

# Tech tracking
tech-stack:
  added: [github.com/charmbracelet/glamour v0.10.0]
  patterns: [glamour TermRenderer with theme-matched StyleConfig, viewport-based overlay]

key-files:
  created: [internal/preview/model.go, internal/preview/keys.go, internal/preview/styles.go]
  modified: [internal/todolist/model.go, internal/todolist/keys.go, internal/todolist/styles.go, internal/app/model.go, go.mod, go.sum]

key-decisions:
  - "Glamour base style selected by theme name: LightStyleConfig for 'light', DarkStyleConfig for all others"
  - "Document.Margin zeroed to prevent glamour adding its own padding (app handles padding via lipgloss)"
  - "Preview overlay follows search/settings pattern: showPreview bool + preview field + CloseMsg handler"
  - "Preview only opens for todos with non-empty body (p on empty-body todo is a no-op)"

patterns-established:
  - "Preview overlay pattern: identical to search/settings (message routing, view rendering, help bindings)"

# Metrics
duration: 3min
completed: 2026-02-06
---

# Phase 15 Plan 02: Preview Overlay and Body Indicator Summary

**Glamour-rendered markdown preview overlay with viewport scrolling, body indicator [+] in todo list, and preview keybinding wired through app**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-06T21:24:02Z
- **Completed:** 2026-02-06T21:27:02Z
- **Tasks:** 2
- **Files created:** 3
- **Files modified:** 5

## Accomplishments
- Created internal/preview package with Model (viewport + glamour), KeyMap (scroll/close), and Styles (theme-integrated renderer)
- Added glamour v0.10.0 dependency for terminal markdown rendering with styled headings, lists, and code blocks
- Added BodyIndicator style and [+] marker to todolist for todos with non-empty body
- Wired preview overlay into app following search/settings overlay pattern (routing, view, help)
- Preview keybinding (p) emits PreviewMsg from todolist, app creates preview and opens overlay

## Task Commits

Each task was committed atomically:

1. **Task 1: Create preview package with glamour markdown rendering** - `97699c6` (feat)
2. **Task 2: Wire preview overlay into app and add body indicator** - `6d5ad03` (feat)

## Files Created/Modified
- `internal/preview/keys.go` - KeyMap with Up/Down/PageUp/PageDown/Close bindings
- `internal/preview/styles.go` - NewMarkdownRenderer with glamour theme integration, Styles struct
- `internal/preview/model.go` - Preview Model with viewport, glamour renderer, CloseMsg, resize handling
- `internal/todolist/styles.go` - Added BodyIndicator style field
- `internal/todolist/model.go` - Added PreviewMsg, [+] indicator in renderTodo, Preview key handler
- `internal/todolist/keys.go` - Added Preview binding (p key) to KeyMap
- `internal/app/model.go` - Added preview import, showPreview/preview fields, routing, view, help
- `go.mod` / `go.sum` - Added glamour v0.10.0 and transitive dependencies

## Decisions Made
- Glamour base style chosen by theme name (light vs dark) rather than terminal auto-detection, matching the app's explicit theme system.
- Document.Margin set to zero to prevent glamour double-padding; app controls padding via lipgloss Border style.
- Preview overlay placed in routing priority after settings but before search, consistent with overlay layering.
- Preview is a no-op when the selected todo has no body (avoids showing empty preview).

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Preview overlay complete and functional for viewing todo bodies
- Phase 15-03 (template creation flow) can now show previews of template-generated bodies
- Body indicator provides visual affordance for which todos have bodies

## Self-Check: PASSED
