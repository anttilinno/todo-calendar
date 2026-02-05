---
phase: 02-calendar-holidays
plan: 01
subsystem: calendar
tags: [toml, config, holidays, calendar-grid, lipgloss, rickar-cal]

# Dependency graph
requires:
  - phase: 01-tui-scaffold
    provides: "Split-pane Bubble Tea app with calendar placeholder model"
provides:
  - "TOML config with country and monday_start (internal/config)"
  - "Holiday provider with 10-country registry (internal/holidays)"
  - "Pure RenderGrid function with per-cell styling (internal/calendar)"
  - "Calendar KeyMap with PrevMonth/NextMonth bindings"
affects: [02-02-integration, 03-todo-management]

# Tech tracking
tech-stack:
  added: [BurntSushi/toml v1.6.0, rickar/cal/v2 v2.1.27]
  patterns: [pure-rendering-functions, registry-pattern, xdg-config-paths]

key-files:
  created:
    - internal/config/config.go
    - internal/config/paths.go
    - internal/holidays/registry.go
    - internal/holidays/provider.go
    - internal/calendar/grid.go
    - internal/calendar/styles.go
    - internal/calendar/keys.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "weekdayHdrStyle shortened from weekdayHeaderStyle to avoid long line in styles.go"
  - "go mod tidy removes rickar/cal until holidays package exists; re-added in Task 2"

patterns-established:
  - "Pure rendering: RenderGrid has no side effects, takes all data as params"
  - "Format before style: fmt.Sprintf before lipgloss.Render to preserve alignment"
  - "Noon construction: time.Date with hour=12 for holiday checks to avoid TZ edge cases"
  - "Registry pattern: map[string][]*cal.Holiday for extensible country support"

# Metrics
duration: 3min
completed: 2026-02-05
---

# Phase 2 Plan 1: Building Block Packages Summary

**TOML config with XDG paths, holiday provider with 10-country rickar/cal registry, pure calendar grid renderer with Lip Gloss per-cell styling and month navigation key bindings**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-05T10:05:57Z
- **Completed:** 2026-02-05T10:08:39Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments
- Config package with TOML loading that gracefully handles missing files (returns defaults) and reports malformed TOML
- Holiday provider backed by rickar/cal with 10-country registry (de, dk, es, fi, fr, gb, it, no, se, us)
- Pure RenderGrid function producing 20-char wide calendar with today/holiday/normal cell styling
- Calendar key bindings (left/h, right/l) with help.KeyMap interface for Phase 3

## Task Commits

Each task was committed atomically:

1. **Task 1: Install dependencies and create config package** - `afd2b32` (feat)
2. **Task 2: Create holiday provider and country registry** - `c973c3b` (feat)
3. **Task 3: Create calendar styles, key bindings, and grid renderer** - `43f4b84` (feat)

**Plan metadata:** `5706131` (docs: complete plan)

## Files Created/Modified
- `internal/config/config.go` - Config struct with Load and DefaultConfig; TOML overlay on defaults
- `internal/config/paths.go` - XDG config path resolution via os.UserConfigDir
- `internal/holidays/registry.go` - Country code to *cal.Holiday slice mapping for 10 countries
- `internal/holidays/provider.go` - Provider with NewProvider and HolidaysInMonth using noon dates
- `internal/calendar/grid.go` - Pure RenderGrid function with title, weekday header, styled day cells
- `internal/calendar/styles.go` - 5 Lip Gloss styles (header, weekdayHdr, normal, today, holiday)
- `internal/calendar/keys.go` - KeyMap with PrevMonth/NextMonth bindings and help.KeyMap interface
- `go.mod` - Added BurntSushi/toml v1.6.0 and rickar/cal/v2 v2.1.27
- `go.sum` - Updated checksums

## Decisions Made
- Shortened `weekdayHeaderStyle` to `weekdayHdrStyle` for conciseness in styles.go
- Go module tidied between tasks since rickar/cal was initially indirect-only until holidays package existed

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All three packages (config, holidays, calendar) compile and pass vet alongside existing code
- Plan 02-02 can import these packages directly to wire into the Bubble Tea model
- RenderGrid accepts all parameters needed by the model: year, month, today, holidays map, mondayStart
- KeyMap ready for key.Matches in model Update method
- No blockers for integration

---
*Phase: 02-calendar-holidays*
*Completed: 2026-02-05*
