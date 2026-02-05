---
phase: "06-themes"
plan: "01"
subsystem: "theme"
tags: ["lipgloss", "theme", "colors", "config"]
dependency_graph:
  requires: []
  provides: ["Theme struct", "4 preset themes", "ForName selector", "Config.Theme field"]
  affects: ["06-02 (wire themes into UI)"]
tech_stack:
  added: []
  patterns: ["semantic color roles", "preset constructor pattern"]
key_files:
  created: ["internal/theme/theme.go"]
  modified: ["internal/config/config.go"]
decisions:
  - id: "06-01-A"
    decision: "14 semantic color fields cover all current UI elements"
    rationale: "Named by role (BorderFocused, HolidayFg) not component (calendarBorder)"
  - id: "06-01-B"
    decision: "Empty string means terminal default"
    rationale: "Dark theme leaves most colors empty to respect user terminal palette"
metrics:
  duration: "1 min"
  completed: "2026-02-05"
---

# Phase 06 Plan 01: Theme Data Layer Summary

Theme struct with 14 semantic lipgloss.Color fields, four preset constructors (Dark/Light/Nord/Solarized), and ForName selector defaulting to Dark.

## What Was Done

### Task 1: Create theme package with struct and 4 presets
- Created `internal/theme/theme.go` with `Theme` struct defining 14 semantic color roles
- Four preset constructors: `Dark()`, `Light()`, `Nord()`, `Solarized()`
- `ForName(name string) Theme` selector using case-insensitive matching, defaults to Dark
- All colors use hex strings; empty string means terminal default
- Commit: `3e28127`

### Task 2: Add Theme field to config
- Added `Theme string` field with `toml:"theme"` tag to Config struct
- Default value `"dark"` in `DefaultConfig()`
- Existing config files without theme field automatically get dark theme
- Commit: `e52e52e`

## Deviations from Plan

None -- plan executed exactly as written.

## Decisions Made

| ID | Decision | Rationale |
|----|----------|-----------|
| 06-01-A | 14 semantic color fields cover all UI elements | Named by role not component for flexibility |
| 06-01-B | Empty string = terminal default | Dark theme respects user terminal palette |

## Verification

- `go build ./internal/theme/` -- pass
- `go build ./internal/config/` -- pass
- `go vet ./internal/...` -- pass

## Next Phase Readiness

Plan 06-02 can proceed immediately. It will:
- Import `theme.ForName(cfg.Theme)` in the app model
- Replace hardcoded color values in calendar and todolist with theme fields
- Pass Theme through to all style definitions
