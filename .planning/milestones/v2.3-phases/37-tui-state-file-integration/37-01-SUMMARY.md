---
phase: 37-tui-state-file-integration
plan: 01
subsystem: ui
tags: [bubbletea, polybar, status-file, tui]

# Dependency graph
requires:
  - phase: 36-status-subcommand
    provides: "status.FormatStatus and status.WriteStatusFile functions"
provides:
  - "refreshStatusFile method on app.Model triggered on startup and every todo mutation"
  - "theme field on app.Model for status file formatting"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: ["status file refresh as side-effect in TUI update cycle"]

key-files:
  created: []
  modified:
    - internal/app/model.go
    - internal/status/status_test.go

key-decisions:
  - "Value receiver for refreshStatusFile matching Init/Update pattern -- reads model fields only"
  - "Silent error handling for status file writes -- best-effort side effect, Polybar shows stale data on failure"
  - "Single refreshStatusFile call at bottom of Update catches all todoPane mutations without per-action wiring"

patterns-established:
  - "Status file refresh pattern: query store, format, write as side-effect in TUI lifecycle"

requirements-completed: [BAR-04, BAR-05]

# Metrics
duration: 2min
completed: 2026-02-23
---

# Phase 37 Plan 01: TUI State File Integration Summary

**refreshStatusFile method on app.Model wired into Init and every todo mutation path for real-time Polybar sync**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-23T14:44:06Z
- **Completed:** 2026-02-23T14:46:21Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Added `refreshStatusFile` method to app.Model that queries today's todos, formats via `status.FormatStatus`, writes via `status.WriteStatusFile`
- Wired refreshStatusFile at 4 call sites: Init (startup), Update bottom (all todoPane mutations), EditorFinishedMsg (body edits), SettingChangedMsg (theme changes)
- Added 2 integration tests validating the FormatStatus + writeStatusFileTo end-to-end pipeline

## Task Commits

Each task was committed atomically:

1. **Task 1: Add refreshStatusFile method and wire into app.Model startup** - `aeaba5b` (feat)
2. **Task 2: Add integration test for refreshStatusFile wiring** - `8ecffd9` (test)

## Files Created/Modified
- `internal/app/model.go` - Added theme field, refreshStatusFile method, 4 call sites (Init, Update, EditorFinishedMsg, SettingChangedMsg)
- `internal/status/status_test.go` - Added TestRefreshStatusFileEndToEnd and TestRefreshStatusFileAllDone integration tests

## Decisions Made
- Used value receiver `(m Model)` for refreshStatusFile to match Init/Update pattern -- it only reads model fields and calls external functions
- Silent error handling (`_ = status.WriteStatusFile()`) -- status file is best-effort; failure just means stale Polybar display
- Single call at bottom of Update() method catches all todoPane mutations (add, toggle, delete, edit, reorder) without needing per-action wiring

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 37 has only this one plan; TUI state file integration is complete
- Polybar can now display real-time todo status via `/tmp/.todo_status` while the TUI is running

---
*Phase: 37-tui-state-file-integration*
*Completed: 2026-02-23*
