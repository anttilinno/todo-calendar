---
phase: 15-markdown-templates
plan: 01
subsystem: database
tags: [sqlite, templates, text-template, store-interface, migration]

# Dependency graph
requires:
  - phase: 14-sqlite-backend
    provides: SQLiteStore with TodoStore interface, PRAGMA user_version migration pattern
provides:
  - Todo.Body field in struct and all SQL queries
  - Template type and CRUD (AddTemplate, ListTemplates, FindTemplate, DeleteTemplate)
  - UpdateBody method on both store backends
  - Migration v2 creating templates table with UNIQUE name constraint
  - internal/tmpl package with ExtractPlaceholders and ExecuteTemplate utilities
affects: [15-02 (preview overlay needs HasBody), 15-03 (template creation needs CRUD + tmpl utils)]

# Tech tracking
tech-stack:
  added: [text/template, text/template/parse]
  patterns: [template AST walking for placeholder extraction, migration versioning v2]

key-files:
  created: [internal/tmpl/tmpl.go]
  modified: [internal/store/todo.go, internal/store/store.go, internal/store/sqlite.go]

key-decisions:
  - "Body field empty on Add(); template flow uses UpdateBody() separately"
  - "JSON Store gets stub implementations for template methods (not supported)"
  - "Templates stored in SQLite with UNIQUE name constraint, not filesystem"
  - "ExtractPlaceholders uses text/template/parse AST walk, not regex"
  - "ExecuteTemplate uses missingkey=zero for safe missing placeholder handling"

patterns-established:
  - "Migration v2 pattern: version < 2 block in migrate() for new tables"
  - "Template AST walk pattern: walkFields recursive with node type switch"

# Metrics
duration: 3min
completed: 2026-02-06
---

# Phase 15 Plan 01: Store Layer Foundation Summary

**Todo.Body field, templates table (migration v2), template CRUD on both store backends, and text/template placeholder extraction/execution utilities**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-06T21:18:31Z
- **Completed:** 2026-02-06T21:21:09Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments
- Todo struct extended with Body field, included in all SQL queries (todoColumns, scanTodo, INSERT)
- Templates table created via migration v2 with UNIQUE name constraint and full CRUD on SQLiteStore
- New internal/tmpl package providing ExtractPlaceholders (AST walk) and ExecuteTemplate (map-based filling)
- TodoStore interface extended with 5 new methods, both Store and SQLiteStore satisfy it

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Body field to Todo struct and update SQL queries** - `a70deb9` (feat)
2. **Task 2: Add templates table migration and template CRUD methods** - `b543fbf` (feat)
3. **Task 3: Create template placeholder extraction and execution utilities** - `7cc0dc4` (feat)

## Files Created/Modified
- `internal/store/todo.go` - Added Body field, HasBody() method, Template type
- `internal/store/store.go` - Extended TodoStore interface (+5 methods), JSON Store stubs, UpdateBody impl
- `internal/store/sqlite.go` - Migration v2, body in columns/scan/INSERT, UpdateBody, template CRUD
- `internal/tmpl/tmpl.go` - New package: ExtractPlaceholders and ExecuteTemplate utilities

## Decisions Made
- Body is always empty on initial Add(); the template creation flow calls UpdateBody() separately to set the body. This avoids changing the Add() signature.
- JSON Store template methods are stubs (return error/nil/no-op) since JSON store does not support templates. Main app uses SQLite exclusively.
- ExtractPlaceholders uses text/template/parse AST walking rather than regex, handling all node types (If, Range, With, etc.) correctly.
- ExecuteTemplate uses Option("missingkey=zero") so missing placeholders produce empty strings rather than errors.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Store layer complete: Todo.Body, Template CRUD, and tmpl utilities all ready
- Phase 15-02 (preview overlay) can use HasBody() for indicators and Body field for rendering
- Phase 15-03 (template creation flow) can use AddTemplate, ListTemplates, ExtractPlaceholders, ExecuteTemplate

## Self-Check: PASSED

---
*Phase: 15-markdown-templates*
*Completed: 2026-02-06*
