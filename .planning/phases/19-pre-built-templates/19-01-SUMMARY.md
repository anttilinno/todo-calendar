---
phase: 19-pre-built-templates
plan: 01
subsystem: database
tags: [sqlite, migration, templates, seeding]

# Dependency graph
requires:
  - phase: 15-markdown-templates
    provides: templates table, AddTemplate/ListTemplates/DeleteTemplate CRUD
provides:
  - 7 pre-built templates seeded on first launch (3 general + 4 dev)
  - Version-3 SQLite migration for template seeding
  - Seed template tests (seeding, idempotency, deletion, placeholder validation)
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Migration-based seeding: seed data via PRAGMA user_version migration, not runtime checks"
    - "INSERT OR IGNORE for idempotent seeding with UNIQUE constraints"

key-files:
  created:
    - internal/store/seed.go
    - internal/store/sqlite_test.go
  modified:
    - internal/store/sqlite.go

key-decisions:
  - "SEED-migration-v3: Seed templates via version-3 migration, not runtime count checks"
  - "SEED-insert-ignore: Use INSERT OR IGNORE to handle existing user templates with same name"

patterns-established:
  - "Seed data pattern: define seed content in separate file, insert via migration"

# Metrics
duration: 1min
completed: 2026-02-07
---

# Phase 19 Plan 01: Pre-Built Templates Summary

**7 seed templates (Meeting Notes, Checklist, Daily Plan, Bug Report, Feature Spec, PR Checklist, Code Review) via version-3 SQLite migration with INSERT OR IGNORE idempotency**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-07T10:42:07Z
- **Completed:** 2026-02-07T10:43:14Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- 7 pre-built templates seeded on first launch (3 general-purpose, 4 dev-focused)
- Each template has 0-3 valid placeholders using {{.PascalCase}} syntax
- Version-3 migration ensures seeding runs exactly once per database
- 4 comprehensive tests covering seeding, idempotency, permanent deletion, and placeholder counts

## Task Commits

Each task was committed atomically:

1. **Task 1: Create seed templates and version-3 migration** - `867c0ae` (feat)
2. **Task 2: Add seed template tests** - `c105976` (test)

## Files Created/Modified
- `internal/store/seed.go` - Template content constants and defaultTemplates() function (7 templates)
- `internal/store/sqlite.go` - Version-3 migration block that seeds templates with INSERT OR IGNORE
- `internal/store/sqlite_test.go` - 4 tests: seeding, idempotency, permanent deletion, placeholder counts

## Decisions Made
- Used version-3 migration (not runtime template count check) to ensure seeding runs exactly once and deleted templates never return
- Used INSERT OR IGNORE to gracefully handle the case where a user already has a template with the same name as a seed template

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed incorrect module path in test import**
- **Found during:** Task 2 (Add seed template tests)
- **Issue:** Plan specified import path `github.com/anttiz/todo-calendar/internal/tmpl` but actual module is `github.com/antti/todo-calendar`
- **Fix:** Corrected import to `github.com/antti/todo-calendar/internal/tmpl`
- **Files modified:** internal/store/sqlite_test.go
- **Verification:** `go test ./internal/store/ -v` passes all 4 tests
- **Committed in:** c105976 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Trivial module path correction. No scope creep.

## Issues Encountered
None beyond the import path typo noted above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All 7 templates are seeded and tested
- Templates are immediately available in the template selection UI on first launch
- Phase 19 is complete (single-plan phase)

## Self-Check: PASSED

---
*Phase: 19-pre-built-templates*
*Completed: 2026-02-07*
