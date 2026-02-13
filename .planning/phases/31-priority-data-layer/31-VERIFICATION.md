---
phase: 31-priority-data-layer
verified: 2026-02-13T19:15:00Z
status: passed
score: 5/5 truths verified
re_verification: false
---

# Phase 31: Priority Data Layer Verification Report

**Phase Goal:** Todos have a priority field that persists through the full store roundtrip
**Verified:** 2026-02-13T19:15:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #   | Truth                                                          | Status     | Evidence                                                                    |
| --- | -------------------------------------------------------------- | ---------- | --------------------------------------------------------------------------- |
| 1   | SQLite schema is at version 7 with a priority INTEGER column   | ✓ VERIFIED | Migration v7 exists (line 156), adds priority INTEGER NOT NULL DEFAULT 0    |
| 2   | Existing todos have priority 0 (no priority) after migration   | ✓ VERIFIED | DEFAULT 0 in ALTER TABLE ensures all existing rows get priority 0           |
| 3   | Store Add() and Update() accept priority and persist it        | ✓ VERIFIED | Add/Update signatures include priority int, INSERT/UPDATE include priority  |
| 4   | Store queries return todos with their priority value populated | ✓ VERIFIED | todoColumns includes priority, scanTodo populates &t.Priority (line 183)    |
| 5   | All tests pass including new priority roundtrip tests          | ✓ VERIFIED | TestPriorityRoundtrip, TestPriorityDefaultZero, TestPriorityHelpers all pass |

**Score:** 5/5 truths verified

### Required Artifacts

| Artifact                             | Expected                                                                          | Status     | Details                                                                                 |
| ------------------------------------ | --------------------------------------------------------------------------------- | ---------- | --------------------------------------------------------------------------------------- |
| `internal/store/todo.go`             | Priority field on Todo struct plus HasPriority() and PriorityLabel() helpers      | ✓ VERIFIED | Priority int field (line 25), HasPriority() (line 29), PriorityLabel() (line 34)       |
| `internal/store/iface.go`            | Extended Add and Update signatures with priority int parameter                    | ✓ VERIFIED | Add(text, date, datePrecision, priority int) (line 8), Update(..., priority int) (12)  |
| `internal/store/sqlite.go`           | Migration v7, updated todoColumns/scanTodo/Add/Update/AddScheduledTodo            | ✓ VERIFIED | Migration v7 (line 156), todoColumns includes priority (174), scanTodo scans Priority (183), Add/Update/AddScheduledTodo all handle priority |
| `internal/store/sqlite_test.go`      | Priority roundtrip and helper tests                                               | ✓ VERIFIED | TestPriorityRoundtrip (line 528), TestPriorityDefaultZero (577), TestPriorityHelpers (598) |
| `internal/todolist/model.go`         | Updated callers to pass priority 0                                                | ✓ VERIFIED | saveEdit() passes 0 (line 831), saveAdd() passes 0 (line 868)                          |
| `internal/recurring/generate_test.go` | fakeStore stubs match updated interface signatures                                | ✓ VERIFIED | Add() signature includes priority int (line 48), Update() signature includes priority int (52) |

### Key Link Verification

| From                                 | To                        | Via                                                  | Status  | Details                                                                                       |
| ------------------------------------ | ------------------------- | ---------------------------------------------------- | ------- | --------------------------------------------------------------------------------------------- |
| `internal/store/iface.go`            | `internal/store/sqlite.go` | SQLiteStore implements TodoStore with new priority parameter | ✓ WIRED | Add signature matches (line 216), Update signature matches (line 278)                        |
| `internal/store/sqlite.go`           | `internal/store/todo.go`   | scanTodo populates Priority field from DB column    | ✓ WIRED | scanTodo includes &t.Priority in Scan (line 183), todoColumns includes priority (line 174)   |
| `internal/todolist/model.go`         | `internal/store/iface.go`  | saveAdd and saveEdit call Add/Update with priority 0 | ✓ WIRED | saveEdit calls Update(..., 0) (line 831), saveAdd calls Add(..., 0) (line 868)               |
| `internal/recurring/generate_test.go` | `internal/store/iface.go`  | fakeStore stubs match updated interface signatures   | ✓ WIRED | fakeStore.Add includes priority int (line 48), fakeStore.Update includes priority int (line 52) |

### Requirements Coverage

| Requirement | Description                                                            | Status      | Supporting Evidence                                                                      |
| ----------- | ---------------------------------------------------------------------- | ----------- | ---------------------------------------------------------------------------------------- |
| PRIO-08     | Priority stored as INTEGER (0=none, 1-4) in SQLite with migration v7   | ✓ SATISFIED | Migration v7 adds priority INTEGER NOT NULL DEFAULT 0, helper methods validate 1-4 range |
| PRIO-09     | Existing todos default to priority 0 (no priority) after migration     | ✓ SATISFIED | DEFAULT 0 in ALTER TABLE ensures all existing rows get priority 0 without data loss      |

### Anti-Patterns Found

None.

All files show substantive implementations with no TODO/FIXME markers, no stub implementations, and no orphaned code.

### Human Verification Required

None.

All verification can be performed programmatically through tests and code inspection. No visual UI, user flows, or external services are involved in this data layer phase.

### Gaps Summary

No gaps found. All must-haves verified. Phase goal fully achieved.

---

## Detailed Verification

### 1. Schema Migration v7

**Location:** `internal/store/sqlite.go` lines 156-163

**Verification:**
- Migration block exists at correct version number (7)
- Adds `priority INTEGER NOT NULL DEFAULT 0` column
- Sets PRAGMA user_version = 7
- DEFAULT 0 ensures all existing todos get priority 0 (no data loss)

**Evidence:**
```go
if version < 7 {
    if _, err := s.db.Exec(`ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0`); err != nil {
        return fmt.Errorf("add priority column: %w", err)
    }
    if _, err := s.db.Exec(`PRAGMA user_version = 7`); err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

### 2. Todo Struct Extension

**Location:** `internal/store/todo.go` lines 25, 28-39

**Verification:**
- Priority field added to Todo struct (line 25): `Priority int \`json:"priority"\``
- HasPriority() method validates range 1-4 (lines 28-31)
- PriorityLabel() returns P1-P4 or empty string (lines 33-39)
- fmt package imported for Sprintf (line 4)

**Edge case testing:** TestPriorityHelpers covers 0, 1-4, -1, 5

### 3. Store Interface Extension

**Location:** `internal/store/iface.go` lines 8, 12

**Verification:**
- Add() signature extended: `Add(text string, date string, datePrecision string, priority int) Todo`
- Update() signature extended: `Update(id int, text string, date string, datePrecision string, priority int)`
- All other methods unchanged (AddScheduledTodo signature stays the same per plan)

### 4. SQLite Implementation

**Location:** `internal/store/sqlite.go` lines 174, 183, 216-252, 278-288, 699-733

**Verification:**
- todoColumns includes priority (line 174)
- scanTodo populates &t.Priority (line 183)
- Add() includes priority parameter and persists it (lines 216-252)
- Update() includes priority parameter and persists it (lines 278-288)
- AddScheduledTodo() hardcodes priority 0 without changing signature (lines 699-733)

**Wiring check:**
- Add() INSERT includes priority column and value ✓
- Update() SET clause includes priority = ? ✓
- scanTodo Scan includes &t.Priority as last argument ✓
- Returned Todo structs include Priority field ✓

### 5. Caller Updates

**Location:** `internal/todolist/model.go` lines 831, 868

**Verification:**
- saveEdit() calls Update(..., 0) at line 831
- saveAdd() calls Add(..., 0) at line 868
- Priority 0 is a placeholder until Phase 32 wires actual priority from edit form

**Location:** `internal/recurring/generate_test.go` lines 48, 52

**Verification:**
- fakeStore.Add() signature includes priority int (line 48)
- fakeStore.Update() signature includes priority int (line 52)
- Stubs return store.Todo{} and do nothing (test doubles, not production code)

### 6. Test Coverage

**Location:** `internal/store/sqlite_test.go` lines 528-622

**Verification:**
- TestPriorityRoundtrip: Add → Find → Update → Find → Todos roundtrip (lines 528-575)
- TestPriorityDefaultZero: Add with priority 0 persists correctly (lines 577-596)
- TestPriorityHelpers: Table-driven test for HasPriority() and PriorityLabel() covering edge cases 0, 1-4, -1, 5 (lines 598-622)
- All existing tests updated for new Add/Update signatures (lines 378, 384, 393, 402, 428, 430, 456, 458, 484, 486, 488)

**Test execution:**
```
$ go test ./internal/store/ -v -run TestPriority
=== RUN   TestPriorityRoundtrip
--- PASS: TestPriorityRoundtrip (0.00s)
=== RUN   TestPriorityDefaultZero
--- PASS: TestPriorityDefaultZero (0.00s)
=== RUN   TestPriorityHelpers
--- PASS: TestPriorityHelpers (0.00s)
PASS
```

**Full test suite:**
```
$ go test ./...
ok  	github.com/antti/todo-calendar/internal/fuzzy	(cached)
ok  	github.com/antti/todo-calendar/internal/holidays	(cached)
ok  	github.com/antti/todo-calendar/internal/recurring	(cached)
ok  	github.com/antti/todo-calendar/internal/store	(cached)
```

### 7. Build Verification

**Compilation:**
```
$ go build ./...
(clean build, no errors)
```

**Static analysis:**
```
$ go vet ./...
(no issues)
```

---

## Commits

Task 1: `9dc7855` - feat(31-01): add priority data layer - schema v7, struct, interface, callers
Task 2: `c766bea` - test(31-01): add priority roundtrip and helper tests

Both commits verified in git log.

---

## Success Criteria Met

- [x] SQLite schema version is 7 with `priority INTEGER NOT NULL DEFAULT 0` column
- [x] Todo struct has Priority int field with HasPriority() and PriorityLabel() helpers
- [x] TodoStore interface Add() and Update() accept priority int parameter
- [x] SQLiteStore Add/Update/AddScheduledTodo persist priority correctly
- [x] All callers (todolist model, recurring fakeStore) updated for new signatures
- [x] All tests pass including 3 new priority-specific tests
- [x] No regressions in existing functionality

---

_Verified: 2026-02-13T19:15:00Z_
_Verifier: Claude (gsd-verifier)_
