---
phase: 36-status-subcommand
plan: 02
subsystem: cli
tags: [polybar, subcommand, status, main]

# Dependency graph
requires:
  - "36-01: FormatStatus, WriteStatusFile, PriorityColorHex"
provides:
  - "todo-calendar status subcommand that queries DB, formats Polybar output, writes state file, and exits"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: [subcommand-routing, early-exit-before-tui]

key-files:
  created: []
  modified:
    - main.go

key-decisions:
  - "Restructured main() so config+db load before subcommand branch, avoiding duplicate setup in runStatus"
  - "runStatus takes pre-opened config and store as parameters rather than re-loading them"
  - "Status subcommand skips holiday provider, Google Calendar, recurring tasks, and Bubble Tea entirely"

patterns-established:
  - "Subcommand routing: check os.Args before TUI setup, call dedicated function, return early"

requirements-completed: [BAR-01, BAR-02, BAR-03]

# Metrics
duration: 1min
completed: 2026-02-23
---

# Phase 36 Plan 02: Status Subcommand Routing Summary

**Wired `todo-calendar status` subcommand into main.go with early exit before TUI, querying today's todos and writing Polybar state file**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-23T14:08:30Z
- **Completed:** 2026-02-23T14:10:00Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Restructured main() to load config+db first, enabling subcommand branching before TUI setup
- Added runStatus() function that queries today's todos, formats Polybar output, and writes state file atomically
- Status subcommand initializes only config, store, theme, and status -- no holiday provider, Google Calendar, recurring tasks, or Bubble Tea

## Task Commits

Each task was committed atomically:

1. **Task 1: Add status subcommand routing to main.go** - `696ca72` (feat)

## Files Created/Modified
- `main.go` - Added status subcommand routing with runStatus() function, restructured config+db loading before TUI setup

## Decisions Made
- Restructured main() so config+db load before subcommand branch, avoiding duplicate setup in runStatus
- runStatus takes pre-opened config and store as parameters rather than re-loading them
- Status subcommand skips holiday provider, Google Calendar, recurring tasks, and Bubble Tea entirely

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Polybar status integration fully complete: `todo-calendar status` is the user-facing entry point
- Users can add `todo-calendar status` to Polybar config or cron jobs
- Phase 36 complete -- all plans executed

## Self-Check: PASSED

- main.go modified and compiles cleanly
- Commit 696ca72 found in git log
- `todo-calendar status` runs, writes /tmp/.todo_status, exits with code 0
- All existing tests pass

---
*Phase: 36-status-subcommand*
*Completed: 2026-02-23*
