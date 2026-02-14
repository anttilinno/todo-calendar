---
phase: 33-oauth-offline-guard
plan: 02
subsystem: ui
tags: [bubbletea, oauth2, settings, tui, google-calendar]

# Dependency graph
requires:
  - phase: 33-oauth-offline-guard/01
    provides: "internal/google package with AuthState, StartAuthFlow, CheckAuthState"
provides:
  - "Google Calendar settings row with dynamic auth status display"
  - "OAuth flow triggered from TUI settings via Enter key"
  - "Graceful offline guard (app works without credentials)"
  - "AuthResultMsg handling in app model for auth state updates"
affects: [34-calendar-fetch, 35-ui-integration]

# Tech tracking
tech-stack:
  added: []
  patterns: [settings-action-row, auth-state-driven-ui, startup-auth-check]

key-files:
  created: []
  modified: [internal/settings/model.go, internal/app/model.go, main.go]

key-decisions:
  - "Google Calendar row is action-only (Enter trigger), not a cycling option"
  - "Auth state checked at startup via CheckAuthState (file existence, no network)"
  - "authFlowActive flag prevents duplicate OAuth launches"

patterns-established:
  - "Action row in settings: special row type with Enter trigger instead of Left/Right cycling"
  - "Startup auth check: main.go calls CheckAuthState before app.New to seed initial state"

# Metrics
duration: 4min
completed: 2026-02-14
---

# Phase 33 Plan 02: Settings UI & OAuth Wiring Summary

**Google Calendar action row in TUI settings with auth-state-driven display, Enter-triggered OAuth flow, and graceful offline guard at startup**

## Performance

- **Duration:** 4 min (includes human verification checkpoint)
- **Started:** 2026-02-14
- **Completed:** 2026-02-14
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Google Calendar row in settings overlay showing dynamic status (Not configured / Sign in / Connected / Reconnect)
- OAuth flow triggered via Enter key on settings row, with "Waiting for browser..." feedback
- App launches cleanly without credentials.json (graceful offline guard AUTH-04)
- Token persists across restarts, showing "Connected" immediately on relaunch

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Google Calendar row to settings and wire auth into app** - `4108e5d` (feat)
2. **Task 2: Verify OAuth flow end-to-end** - human-verify checkpoint, approved

## Files Created/Modified
- `internal/settings/model.go` - Google Calendar action row with auth state display, Enter trigger, authFlowActive guard
- `internal/app/model.go` - AuthResultMsg handling, googleAuthState tracking, StartAuthFlow dispatch
- `main.go` - CheckAuthState at startup, passed to app.New

## Decisions Made
- Google Calendar row uses action-row pattern (Enter trigger) instead of cycling Left/Right like other settings
- Auth state checked at startup via file existence only (no network call, no delay)
- authFlowActive boolean prevents launching duplicate OAuth flows

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

**External services require manual configuration.** Users must:
- Create a GCP project with Calendar API enabled
- Create OAuth 2.0 Desktop client credentials
- Download credentials.json to ~/.config/todo-calendar/credentials.json

## Next Phase Readiness
- Settings UI complete for Google Calendar integration
- Auth flow end-to-end verified (AUTH-01 through AUTH-04, CONF-02)
- Ready for Phase 34 to implement calendar event fetching using TokenSource
- Ready for Phase 35 to integrate events into the calendar UI

## Self-Check: PASSED

- FOUND: internal/settings/model.go
- FOUND: internal/app/model.go
- FOUND: main.go
- FOUND: commit 4108e5d

---
*Phase: 33-oauth-offline-guard*
*Completed: 2026-02-14*
