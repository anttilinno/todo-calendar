---
phase: 38-update-engine
plan: 01
subsystem: update
tags: [github-api, sha256, semver, net-http]

requires:
  - phase: none
    provides: standalone module
provides:
  - CheckForUpdate function querying GitHub Releases API
  - DownloadAsset function with 5s HTTP timeout
  - VerifyChecksum SHA256 verification
  - compareVersions integer-based semver comparison
affects: [38-02 atomic binary replacement, 40 launch wiring]

tech-stack:
  added: []
  patterns: [httptest mock server for API testing, apiURL parameter override for testability]

key-files:
  created: [internal/update/check.go, internal/update/check_test.go]
  modified: []

key-decisions:
  - "CheckForUpdate returns *Release struct (Tag, BinaryURL, ChecksumURL) instead of just version string for downstream use"
  - "apiURL parameter override instead of package-level var for test isolation"
  - "parseChecksumFile accepts both 'hash  filename' and bare hex formats"

patterns-established:
  - "update package: httptest.NewServer with JSON handler for GitHub API mocking"
  - "apiURL parameter override pattern for testable HTTP clients"

requirements-completed: [R1, R2, NR1, NR2, NR3]

duration: 3min
completed: 2026-03-24
---

# Phase 38 Plan 01: GitHub API Client, Downloader, and SHA256 Checksum Verifier Summary

**GitHub Releases API client with semver comparison, asset downloader with 5s timeout, and SHA256 checksum verification using only net/http stdlib**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-24T14:57:39Z
- **Completed:** 2026-03-24T15:01:00Z
- **Tasks:** 1 (TDD: RED + GREEN)
- **Files modified:** 2

## Accomplishments
- CheckForUpdate queries GitHub Releases API, compares semver, returns Release with asset URLs when newer version exists
- DownloadAsset fetches arbitrary URL content with 5-second HTTP timeout
- VerifyChecksum validates SHA256 against checksum files in both "hash  filename" and bare hex formats
- 13 test cases passing including version comparison, network errors, invalid JSON, 404 responses, checksum mismatch

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): Failing tests for update engine** - `2189152` (test)
2. **Task 1 (GREEN): Implement update engine** - `6c049f4` (feat)

## Files Created/Modified
- `internal/update/check.go` - GitHub API client, asset downloader, SHA256 checksum verifier, semver comparator
- `internal/update/check_test.go` - 13 test cases with httptest mock servers

## Decisions Made
- CheckForUpdate returns *Release struct with Tag/BinaryURL/ChecksumURL rather than just a version string, making downstream consumption trivial
- Used apiURL parameter override for testability rather than unexported package variable
- parseChecksumFile handles both "hash  filename" format (sha256sum output) and bare hex hash

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- update package exports CheckForUpdate, DownloadAsset, VerifyChecksum ready for 38-02 (atomic binary replacement)
- Release struct provides BinaryURL and ChecksumURL for download orchestration
- No new dependencies added; only stdlib net/http, crypto/sha256, encoding/json

---
*Phase: 38-update-engine*
*Completed: 2026-03-24*
