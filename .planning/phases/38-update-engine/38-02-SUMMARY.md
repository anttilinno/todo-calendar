---
phase: 38-update-engine
plan: 02
subsystem: update
tags: [atomic-write, binary-replacement, os-rename]

requires:
  - phase: 38-01
    provides: update package with CheckForUpdate, DownloadAsset, VerifyChecksum
provides:
  - ReplaceBinary function for atomic binary replacement with permission preservation
affects: [40 launch wiring, update orchestration]

tech-stack:
  added: []
  patterns: [atomic temp+sync+rename replacement following config.go pattern]

key-files:
  created: [internal/update/replace.go, internal/update/replace_test.go]
  modified: []

key-decisions:
  - "Follows exact atomic write pattern from config.go Save: CreateTemp+Write+Sync+Close+Chmod+Rename"
  - "Temp file created in same directory as target to ensure atomic rename (same filesystem)"

patterns-established:
  - "ReplaceBinary: generic atomic file replacement with permission preservation"

requirements-completed: [R3, NR2]

duration: 2min
completed: 2026-03-24
---

# Phase 38 Plan 02: Atomic Binary Replacement Summary

**Atomic binary replacement using temp+sync+rename with original permission preservation and empty-data safety check**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-24T15:07:06Z
- **Completed:** 2026-03-24T15:10:00Z
- **Tasks:** 1 (TDD: RED + GREEN)
- **Files modified:** 2

## Accomplishments
- ReplaceBinary atomically replaces binary at given path using temp file + sync + rename
- Original file permissions preserved via os.Stat + os.Chmod
- Empty binary data rejected with clear error message
- 5 test cases covering success, permission preservation, non-existent path, read-only dir, and empty binary

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): Failing tests for binary replacement** - `da12e90` (test)
2. **Task 1 (GREEN): Implement atomic binary replacement** - `613215f` (feat)

## Files Created/Modified
- `internal/update/replace.go` - ReplaceBinary function with atomic temp+sync+rename pattern
- `internal/update/replace_test.go` - 5 test cases for binary replacement

## Decisions Made
- Followed exact same atomic write pattern as config.go Save function for codebase consistency
- Temp file created in same directory as target binary to guarantee atomic rename on same filesystem

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- update package now exports CheckForUpdate, DownloadAsset, VerifyChecksum (plan 01) and ReplaceBinary (plan 02)
- Ready for launch wiring phase to orchestrate the full update flow: check -> download -> verify -> replace

---
*Phase: 38-update-engine*
*Completed: 2026-03-24*
