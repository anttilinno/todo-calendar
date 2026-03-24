# Requirements: v2.4 Auto-Update

## Goal

Automatically update the binary on app launch when a newer GitHub release exists, controlled by a settings toggle.

## Functional Requirements

### R1: Version Check on Launch
When `auto_update` is enabled and `version != "dev"`, check GitHub Releases API (`/repos/antti/todo-calendar/releases/latest`) for the latest version tag. Compare against compiled-in `version`. Skip update if current version >= latest.

### R2: Binary Download & Verification
Download `todo-calendar-linux-amd64` and `todo-calendar-linux-amd64.sha256` from the latest release. Verify SHA256 checksum matches. Abort update on mismatch.

### R3: Binary Replacement
Replace the binary at the configured `binary_path` using atomic write (temp file + rename). Preserve file permissions from the original binary. Continue launching normally with current version — the new binary takes effect on next launch.

### R4: Settings — auto_update Toggle
Add `auto_update` bool to Config (default: `false`). Expose in settings overlay as a cycling toggle (Enabled/Disabled).

### R5: Settings — binary_path
Add `binary_path` string to Config (default: `""`). When empty, resolve from `os.Executable()` with symlink resolution at runtime. Expose in settings overlay as a read-only display field showing the effective path.

## Non-Functional Requirements

### NR1: Timeout
HTTP requests must timeout after 5 seconds. A slow or unreachable GitHub API must not delay app launch noticeably.

### NR2: Silent Failure
Network errors, API errors, and checksum mismatches log to stderr but do not prevent app launch. The update is best-effort.

### NR3: No New Dependencies
Use net/http from stdlib. No third-party HTTP or update libraries.

## Out of Scope

- Multi-architecture support (only linux-amd64)
- Rollback mechanism
- Progress indicator during download
- In-app notification of update result
- Changelog display
