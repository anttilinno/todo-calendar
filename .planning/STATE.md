# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-06)

**Core value:** See your month at a glance -- calendar with holidays and todos in one terminal screen.
**Current focus:** v1.4 Data & Editing -- Phase 16 External Editor

## Current Position

Phase: 16 of 16 (External Editor)
Plan: --
Status: Ready to plan
Last activity: 2026-02-06 -- Phase 15 verified and complete

Progress: █████████████████████████████░ 97% (27/28 plans estimated through v1.4)

## Performance Metrics

**Velocity:**
- Total plans completed: 27 (through 15-03)
- Average duration: 2 min
- Total execution time: ~1.1 hours

## Accumulated Context

### Decisions

See PROJECT.md Key Decisions table for complete log.

Recent decisions affecting current work:
- v1.4: SQLite via modernc.org/sqlite (pure Go, no CGo) chosen over mattn/go-sqlite3
- v1.4: No JSON-to-SQLite migration needed (user has no existing data)
- v1.4: Query-on-read (no in-memory cache) -- dataset too small to warrant caching
- 14-01: TodoStore interface in store package; all consumers depend on interface not concrete type
- 14-02: Hand-written SQL with scan helpers instead of sqlc (single table, simple CRUD)
- 14-02: PRAGMA user_version for schema versioning instead of dbmate (lightweight desktop app)
- 14-02: WAL journal mode, MaxOpenConns(1), body column in schema but excluded from SELECTs
- 15-01: Body empty on Add(); template flow uses UpdateBody() separately
- 15-01: JSON Store gets stub template methods (not supported); SQLite is primary
- 15-01: ExtractPlaceholders uses text/template/parse AST walk, not regex
- 15-02: Glamour base style by theme name (light vs dark), Document.Margin zeroed
- 15-02: Preview overlay follows search/settings pattern (showPreview + CloseMsg)
- 15-03: Template selection uses cursor-based navigation, textarea for multi-line content
- 15-03: fromTemplate + pendingBody deferred body attachment after todo creation

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-02-06
Stopped at: Completed 15-03-PLAN.md (Phase 15 complete)
Resume file: None
