---
phase: 33-oauth-offline-guard
plan: 01
subsystem: auth
tags: [oauth2, google, pkce, token-persistence, bubbletea]

# Dependency graph
requires: []
provides:
  - "internal/google package with OAuth 2.0 auth core"
  - "AuthState enum for UI state detection"
  - "TokenSource for authenticated HTTP clients"
  - "StartAuthFlow tea.Cmd for Bubble Tea integration"
affects: [34-calendar-fetch, 35-ui-integration]

# Tech tracking
tech-stack:
  added: [golang.org/x/oauth2, cloud.google.com/go/compute/metadata]
  patterns: [persistingTokenSource-wrapper, atomic-token-write-0600, loopback-pkce-auth-flow]

key-files:
  created: [internal/google/auth.go, internal/google/auth_test.go]
  modified: [go.mod, go.sum]

key-decisions:
  - "PKCE with S256 challenge for desktop OAuth security"
  - "Ephemeral port (127.0.0.1:0) for loopback redirect"
  - "Token hash comparison (AccessToken+Expiry) to detect refresh"
  - "checkAuthStateAt unexported helper for testability"

patterns-established:
  - "persistingTokenSource: wrap oauth2.TokenSource to auto-save on refresh"
  - "Atomic token write: CreateTemp + Write + Sync + Close + Chmod 0600 + Rename"

# Metrics
duration: 2min
completed: 2026-02-14
---

# Phase 33 Plan 01: OAuth Auth Core Summary

**OAuth 2.0 auth package with PKCE loopback flow, atomic token persistence (0600), and Bubble Tea integration via AuthResultMsg**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-14T11:32:05Z
- **Completed:** 2026-02-14T11:33:46Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Complete OAuth 2.0 implementation with PKCE in internal/google/auth.go
- Token persistence with atomic writes and 0600 permissions
- Auth state detection (NotConfigured/NeedsLogin/Ready/Revoked) without network calls
- 6 unit tests covering persistence, permissions, and state detection

## Task Commits

Each task was committed atomically:

1. **Task 1: Create internal/google/auth.go with OAuth core** - `cdfe03b` (feat)
2. **Task 2: Add unit tests for token persistence and auth state** - `fa40548` (test)

## Files Created/Modified
- `internal/google/auth.go` - OAuth config loading, token persistence, PKCE auth flow, AuthState, TokenSource, StartAuthFlow
- `internal/google/auth_test.go` - 6 tests for save/load, permissions, auth state, invalid config
- `go.mod` - Added golang.org/x/oauth2 dependency
- `go.sum` - Updated checksums

## Decisions Made
- Used PKCE with S256 challenge option for desktop app OAuth security
- Ephemeral port on 127.0.0.1:0 for loopback redirect (avoids port conflicts)
- Token hash uses AccessToken+Expiry string to detect token refresh
- Added unexported checkAuthStateAt helper accepting explicit paths for testability

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
- internal/google package ready for Phase 34 to call TokenSource() for authenticated HTTP clients
- AuthState enum ready for UI integration in Phase 35
- StartAuthFlow() tea.Cmd ready for Bubble Tea model usage

---
*Phase: 33-oauth-offline-guard*
*Completed: 2026-02-14*
