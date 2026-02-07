---
phase: 21-schedule-schema-crud
verified: 2026-02-07T12:50:49Z
status: passed
score: 16/16 must-haves verified
---

# Phase 21: Schedule Schema & CRUD Verification Report

**Phase Goal:** The data layer supports recurring schedule definitions attached to templates with deduplication tracking
**Verified:** 2026-02-07T12:50:49Z
**Status:** PASSED
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | ParseRule correctly parses all four cadence formats: daily, weekdays, weekly:days, monthly:N | ✓ VERIFIED | ParseRule() in rule.go handles all 4 types with validation. Tests pass for daily, weekdays, weekly:mon,fri, monthly:15 |
| 2 | MatchesDate returns true for dates matching the cadence and false otherwise | ✓ VERIFIED | MatchesDate() implemented with switch on Type. Tests verify Monday-Friday for weekdays, specific days for weekly, etc. |
| 3 | Monthly day clamping works for short months (e.g., monthly:31 matches Feb 28 in non-leap year) | ✓ VERIFIED | MatchesDate() uses lastDayOfMonth() helper to clamp. TestMatchesMonthly31Clamping verifies Feb 28, Apr 30 |
| 4 | String() round-trips back to parseable format | ✓ VERIFIED | String() method formats back to "daily", "weekly:mon,fri", etc. TestStringRoundTrip verifies ParseRule(r.String()) == r |
| 5 | ParseRule returns descriptive error for invalid input | ✓ VERIFIED | Tests verify errors for empty, invalid day names, out-of-range monthly, etc. Error messages are descriptive |
| 6 | Migration v4 creates schedules table with FK CASCADE to templates | ✓ VERIFIED | sqlite.go lines 109-126: CREATE TABLE schedules with REFERENCES templates(id) ON DELETE CASCADE. Index on template_id created. |
| 7 | Migration v5 adds schedule_id and schedule_date columns to todos table with unique dedup index | ✓ VERIFIED | sqlite.go lines 128-141: ALTER TABLE adds schedule_id (FK SET NULL), schedule_date, UNIQUE INDEX on (schedule_id, schedule_date) |
| 8 | AddSchedule persists a schedule linked to a template and returns it with an ID | ✓ VERIFIED | sqlite.go line 540: INSERT INTO schedules, returns Schedule with LastInsertId. TestScheduleCRUD verifies |
| 9 | ListSchedules returns all schedules ordered by ID | ✓ VERIFIED | sqlite.go line 561: SELECT ... ORDER BY id. Returns []Schedule |
| 10 | ListSchedulesForTemplate returns only schedules for a given template | ✓ VERIFIED | sqlite.go line 572: SELECT WHERE template_id = ? ORDER BY id |
| 11 | DeleteSchedule removes a schedule by ID | ✓ VERIFIED | sqlite.go line 586: DELETE FROM schedules WHERE id = ? |
| 12 | UpdateSchedule modifies cadence and placeholder defaults | ✓ VERIFIED | sqlite.go line 591: UPDATE schedules SET cadence_type, cadence_value, placeholder_defaults. Returns error on failure |
| 13 | TodoExistsForSchedule correctly detects duplicates by schedule_id + date | ✓ VERIFIED | sqlite.go line 603: SELECT 1 WHERE schedule_id = ? AND schedule_date = ?. TestScheduleDeduplication verifies true/false behavior |
| 14 | AddScheduledTodo creates a todo linked to a schedule with schedule_date set | ✓ VERIFIED | sqlite.go line 613: INSERT with schedule_id and schedule_date. Returns Todo with fields populated. TestAddScheduledTodo verifies |
| 15 | Deleting a template cascades to delete its schedules | ✓ VERIFIED | FK CASCADE in migration v4. TestScheduleCascadeOnTemplateDelete verifies schedules gone after DeleteTemplate |
| 16 | Deleting a schedule sets schedule_id to NULL on linked todos (SET NULL) | ✓ VERIFIED | FK SET NULL in migration v5. TestScheduleSetNullOnDelete verifies todo.ScheduleID = 0 after DeleteSchedule |

**Score:** 16/16 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/recurring/rule.go` | ScheduleRule type with ParseRule, MatchesDate, String | ✓ VERIFIED | 134 lines, exports ScheduleRule and ParseRule, all methods implemented |
| `internal/recurring/rule_test.go` | Comprehensive test coverage for all cadence types | ✓ VERIFIED | 328 lines, 27 tests covering parse, match, errors, round-trip. All pass |
| `internal/store/todo.go` | Schedule struct and Todo.ScheduleID/ScheduleDate fields | ✓ VERIFIED | Schedule struct lines 37-44, Todo fields lines 19-20 |
| `internal/store/store.go` | TodoStore interface with 7 schedule methods | ✓ VERIFIED | Lines 35-41: AddSchedule, ListSchedules, ListSchedulesForTemplate, DeleteSchedule, UpdateSchedule, TodoExistsForSchedule, AddScheduledTodo. JSON stubs lines 439-470 |
| `internal/store/sqlite.go` | Migration v4+v5 and SQLite schedule method implementations | ✓ VERIFIED | Migration v4 lines 109-126, v5 lines 128-141. All 7 methods lines 540-646. todoColumns updated line 152. scanTodo updated lines 159-174 |
| `internal/store/sqlite_test.go` | Integration tests for schedule CRUD and FK behavior | ✓ VERIFIED | 5 tests: TestScheduleCRUD, TestScheduleDeduplication, TestScheduleCascadeOnTemplateDelete, TestScheduleSetNullOnDelete, TestAddScheduledTodo. All pass |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| internal/recurring/rule_test.go | internal/recurring/rule.go | import and test all exported functions | ✓ WIRED | Tests call ParseRule, MatchesDate, String. All tests pass |
| internal/store/sqlite.go | internal/store/store.go | SQLiteStore implements TodoStore schedule methods | ✓ WIRED | All 7 schedule methods implemented. Compile-time interface check passes (line 15) |
| internal/store/sqlite.go | internal/store/todo.go | Uses Schedule struct in return types | ✓ WIRED | AddSchedule returns Schedule{}, scanSchedule uses Schedule struct |
| internal/store/store.go | internal/store/todo.go | Interface references Schedule type | ✓ WIRED | AddSchedule signature returns Schedule, ListSchedules returns []Schedule |
| internal/store/sqlite.go (migrations) | internal/store/sqlite.go (methods) | Migrations create schema used by CRUD methods | ✓ WIRED | Migration v4 creates schedules table. Migration v5 adds FK columns. CRUD methods query these tables |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
|-------------|--------|----------------|
| REQ-26: Schedule schema (migrations v4+v5) | ✓ SATISFIED | None. Migration v4 creates schedules table with FK CASCADE. Migration v5 adds schedule_id/schedule_date with dedup index |
| REQ-27: Schedule rule types (daily/weekdays/weekly/monthly) | ✓ SATISFIED | None. All 4 cadence types parse and match correctly. Monthly clamping verified |
| REQ-28: Schedule CRUD in store | ✓ SATISFIED | None. All 7 methods in TodoStore interface, SQLite implemented, JSON stubbed |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None | - | - | - | All implementations substantive, no stubs |

**Notes:**
- "placeholder" string occurrences in sqlite.go are legitimate field names (placeholder_defaults), not stub patterns
- All error handling present (ParseRule returns errors, UpdateSchedule returns error)
- No TODO/FIXME comments
- No empty returns or console.log patterns

### Human Verification Required

None. All verification automated via tests and code inspection.

### Summary

Phase 21 goal **ACHIEVED**. All must-haves verified:

**Plan 21-01 (ScheduleRule TDD):**
- ✓ ParseRule parses all 4 cadence types correctly with validation
- ✓ MatchesDate returns correct boolean for all types
- ✓ Monthly clamping works (Feb 28/29, Apr 30 edge cases)
- ✓ String() round-trips through ParseRule
- ✓ All 27 tests pass (parse, match, errors, round-trip)

**Plan 21-02 (Schema & CRUD):**
- ✓ Schedule struct with all fields (ID, TemplateID, CadenceType, CadenceValue, PlaceholderDefaults, CreatedAt)
- ✓ Todo fields ScheduleID and ScheduleDate added
- ✓ Migration v4 creates schedules table with FK CASCADE to templates
- ✓ Migration v5 adds schedule columns to todos with FK SET NULL and UNIQUE dedup index
- ✓ TodoStore interface has all 7 schedule methods
- ✓ SQLite implements all 7 methods (AddSchedule, ListSchedules, ListSchedulesForTemplate, DeleteSchedule, UpdateSchedule, TodoExistsForSchedule, AddScheduledTodo)
- ✓ JSON store stubs all 7 methods
- ✓ FK CASCADE verified: deleting template deletes schedules
- ✓ FK SET NULL verified: deleting schedule nullifies todo.schedule_id
- ✓ Dedup index verified: UNIQUE(schedule_id, schedule_date) enforced
- ✓ All 5 integration tests pass
- ✓ All existing tests continue to pass (no regressions)

**Test Results:**
```
go test ./internal/recurring/ -v
=== All 27 tests PASS ===

go test ./internal/store/ -v
=== All 9 tests PASS (including 5 new schedule tests) ===
```

**Requirements:** REQ-26, REQ-27, REQ-28 all satisfied.

**Data layer ready** for Phase 22 to build auto-creation and UI on top.

---

_Verified: 2026-02-07T12:50:49Z_
_Verifier: Claude (gsd-verifier)_
