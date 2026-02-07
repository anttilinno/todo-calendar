---
phase: 23
plan: 01
subsystem: store, calendar
tags: [cleanup, dead-code, calendar-styles, today-indicator]
depends_on: []
provides: [clean-store-package, blended-today-styles]
affects: [23-02]
tech_stack:
  added: []
  patterns: [interface-extraction, blended-cell-styles]
key_files:
  created:
    - internal/store/iface.go
  modified:
    - internal/store/todo.go
    - internal/calendar/styles.go
    - internal/calendar/grid.go
  deleted:
    - internal/store/store.go
decisions:
  - "TodoStore interface extracted to iface.go; MonthCount and FloatingCount moved alongside it"
  - "Data struct removed from todo.go (only used by JSON store)"
  - "Blended today styles use indicator/done foreground with today background for status-at-a-glance"
metrics:
  duration: 2m 15s
  completed: 2026-02-07
---

# Phase 23 Plan 01: Store Cleanup and Today Indicator Blending Summary

Removed the dead JSON store implementation (471 lines deleted) and added blended today+indicator calendar cell styling so users see pending/done status on today's date at a glance.

## Task Commits

| Task | Name | Commit | Key Changes |
|------|------|--------|-------------|
| 1 | Extract TodoStore interface and remove JSON store | 54e50a7 | Created iface.go, deleted store.go, removed Data struct |
| 2 | Blend today highlight with indicator status | 58b25d6 | Added TodayIndicator/TodayDone styles, updated both grid renderers |

## What Was Done

### Task 1: Extract TodoStore interface and remove JSON store
- Created `internal/store/iface.go` with the `TodoStore` interface, `MonthCount` struct, and `FloatingCount` struct
- Deleted `internal/store/store.go` entirely (JSON `Store` struct, `NewStore`, `TodosPath`, all methods, stub template/schedule methods)
- Removed the `Data` struct from `internal/store/todo.go` (JSON envelope only used by the deleted store)
- Verified `SQLiteStore` already has compile-time interface check in `sqlite.go`
- Net result: 471 lines of dead code removed

### Task 2: Blend today highlight with indicator status
- Added `TodayIndicator` style: bold, indicator foreground color, today background -- shows when today has pending todos
- Added `TodayDone` style: bold, completed count foreground color, today background -- shows when today's todos are all done
- Updated `RenderGrid` style priority: today+pending > today+done > today > holiday > indicator > done > normal
- Updated `RenderWeekGrid` with identical blended priority logic
- Dark theme (empty TodayBg): bold + foreground color alone distinguishes today from non-today indicators

## Deviations from Plan

None -- plan executed exactly as written.

Note: Pre-existing uncommitted changes in other packages (todolist, search, settings, tmplmgr, preview) caused `go build ./...` to fail on unrelated code. Build and vet were verified on the specific modified packages (`internal/store`, `internal/calendar`) which compiled cleanly. Store tests pass.

## Decisions Made

1. **Interface file naming**: Used `iface.go` as specified in the plan for the extracted interface and shared types
2. **Blending strategy**: Foreground color communicates status (yellow=pending, green=done), background keeps today highlight, bold distinguishes from non-today cells

## Next Phase Readiness

- Store package is clean: single SQLite implementation with interface in separate file
- Calendar grid renderers support blended today+status styles
- Ready for 23-02 (remaining cleanup/polish tasks)

## Self-Check: PASSED
