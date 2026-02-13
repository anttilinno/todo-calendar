# Phase 31: Priority Data Layer - Research

**Researched:** 2026-02-13
**Domain:** SQLite schema migration, Go struct extension, store interface evolution
**Confidence:** HIGH

## Summary

Phase 31 adds a `priority` INTEGER column to the SQLite todos table, extends the `Todo` struct, and updates the `TodoStore` interface so that `Add()` and `Update()` accept a priority parameter. This is a pure data layer change with no UI modifications. The phase follows the exact same pattern used by migration v6 (which added `date_precision`) and migration v5 (which added `schedule_id` and `schedule_date`).

The codebase already has a well-established migration pattern using `PRAGMA user_version`. The current version is 6; this phase bumps it to 7. The `todoColumns` constant and `scanTodo()` function provide a single place to add the new column to all SELECT queries. The `Add()` and `Update()` interface methods need an additional `priority int` parameter, which is a breaking change that affects 3 callers in the todolist model and 1 caller in the recurring package's `AddScheduledTodo`.

**Primary recommendation:** Follow the v6 migration pattern exactly -- `ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0`, extend `todoColumns`/`scanTodo`, add `priority` parameter to `Add()`/`Update()` signatures, update all callers to pass `0` (no priority) where priority is not yet user-specified (scheduled todos, existing callers during the transition).

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `modernc.org/sqlite` | v1.44.3 | Pure-Go SQLite driver | Already in use, no CGO dependency |
| `database/sql` | stdlib | SQL interface | Standard Go database access |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `testing` | stdlib | Unit tests | All test files |

No new dependencies are needed for this phase. The priority data layer is entirely within the existing `internal/store` package.

## Architecture Patterns

### Existing Migration Pattern (v1-v6)

The codebase uses `PRAGMA user_version` for schema versioning. Each migration is guarded by `if version < N` and ends with `PRAGMA user_version = N`. This is the single source of truth for schema state.

**Pattern from v6 (date_precision -- most recent, most relevant):**

```go
// Source: internal/store/sqlite.go lines 143-155
if version < 6 {
    if _, err := s.db.Exec(`ALTER TABLE todos ADD COLUMN date_precision TEXT NOT NULL DEFAULT 'day'`); err != nil {
        return fmt.Errorf("add date_precision column: %w", err)
    }
    // Floating todos (date IS NULL) should have empty date_precision, not 'day'.
    if _, err := s.db.Exec(`UPDATE todos SET date_precision = '' WHERE date IS NULL`); err != nil {
        return fmt.Errorf("fix floating date_precision: %w", err)
    }
    if _, err := s.db.Exec(`PRAGMA user_version = 6`); err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

**Key observations:**
- `ALTER TABLE ... ADD COLUMN` with `NOT NULL DEFAULT` is the standard approach for new columns
- Post-migration fixup updates (v6 fixed floating todos) are applied when needed
- Priority migration does NOT need a fixup: `DEFAULT 0` means "no priority" for all existing rows, which is the correct semantic

### Column Extension Pattern (todoColumns + scanTodo)

All SELECT queries use the `todoColumns` constant. All row scanning goes through `scanTodo()`. This means adding a column requires changes in exactly 2 places for reads, plus each INSERT/UPDATE statement for writes.

```go
// Source: internal/store/sqlite.go lines 164-189
const todoColumns = "id, text, body, date, done, created_at, sort_order, schedule_id, schedule_date, date_precision"

func scanTodo(scanner interface{ Scan(...any) error }) (Todo, error) {
    var t Todo
    var date sql.NullString
    var done int
    var scheduleID sql.NullInt64
    var scheduleDate sql.NullString
    err := scanner.Scan(&t.ID, &t.Text, &t.Body, &date, &done, &t.CreatedAt, &t.SortOrder, &scheduleID, &scheduleDate, &t.DatePrecision)
    // ...
}
```

### Interface Method Signature Pattern

The `TodoStore` interface method `Add()` currently takes `(text, date, datePrecision)` and `Update()` takes `(id, text, date, datePrecision)`. When `date_precision` was added in v1.9, the signatures were extended by appending the new parameter. The same approach applies for priority.

```go
// Source: internal/store/iface.go lines 8-9
Add(text string, date string, datePrecision string) Todo
Update(id int, text string, date string, datePrecision string)
```

### Todo Struct Helper Method Pattern

The `Todo` struct has helper methods like `HasBody()`, `IsMonthPrecision()`, `HasDate()`, `IsFuzzy()`. Priority should follow this pattern with `HasPriority()` and `PriorityLabel()`.

### Anti-Patterns to Avoid

- **Storing priority as TEXT ("p1", "p2"):** Cannot sort numerically, wastes storage, requires mapping on every read. Use INTEGER.
- **Using 0 as P1:** Zero is the default for `INTEGER NOT NULL DEFAULT 0`. It MUST mean "no priority" so existing rows are semantically correct without a backfill UPDATE.
- **Adding a priority index:** Not needed. Priority is not used in WHERE clauses or ORDER BY in SQL. Sorting by priority (if ever needed) would be an in-app concern. An unnecessary index wastes disk and slows writes.
- **Changing AddScheduledTodo signature:** Scheduled todos always get priority 0 (no priority). Do not pollute the recurring engine with priority awareness -- priority is a user-set attribute, not an auto-generated one.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Schema versioning | Custom version tracking | `PRAGMA user_version` | Already established in codebase, SQLite native |
| Column addition | Manual CREATE TABLE with all columns | `ALTER TABLE ADD COLUMN` | SQLite handles existing rows with DEFAULT |
| Priority validation | Complex enum/type system | Simple range check `p >= 0 && p <= 4` | Only 5 valid values, integer comparison is sufficient |

**Key insight:** This entire phase is mechanical application of established patterns. Every decision follows a precedent already in the codebase. The risk is low because the pattern has been validated 6 times already (migrations v1 through v6).

## Common Pitfalls

### Pitfall 1: Forgetting to Update AddScheduledTodo's INSERT
**What goes wrong:** The `AddScheduledTodo` INSERT statement does not include `priority`, so it uses the column DEFAULT (0). This is actually correct behavior, but the returned `Todo` struct should have `Priority: 0` explicitly set for clarity.
**Why it happens:** `AddScheduledTodo` has its own INSERT that bypasses `Add()`.
**How to avoid:** Ensure the `AddScheduledTodo` INSERT includes `priority` in its column list (value 0), and the returned Todo struct sets `Priority: 0`.
**Warning signs:** `AddScheduledTodo` test returns a Todo with Priority not matching expectations.

### Pitfall 2: scanTodo Column Order Mismatch
**What goes wrong:** `todoColumns` lists columns in one order but `scanTodo` scans in a different order, causing data corruption (priority value ends up in wrong field).
**Why it happens:** Manual string column lists must exactly match the Scan() call order.
**How to avoid:** Always append the new column to the END of both `todoColumns` and the `Scan()` call. Verify the test roundtrip catches any mismatch.
**Warning signs:** Tests pass but data values are wrong (e.g., priority shows up as date_precision).

### Pitfall 3: Interface Signature Change Breaks Callers
**What goes wrong:** Changing `Add(text, date, datePrecision)` to `Add(text, date, datePrecision, priority)` breaks all callers at compile time. If any caller is missed, the build fails.
**Why it happens:** Go interfaces are strict -- all implementors and all callers must be updated.
**How to avoid:** This is actually a safety feature. Let the compiler find all sites. Known callers:
  - `todolist.Model.saveAdd()` at line 868 -- pass `0` initially (UI not yet wired)
  - `todolist.Model.saveEdit()` at line 831 -- pass `0` initially
  - `recurring/generate.go` line 57 via `AddScheduledTodo` -- this method is separate, not affected by Add signature change
  - `recurring/generate_test.go` line 33 `fakeStore.AddScheduledTodo` -- separate interface
**Warning signs:** Compile errors (which is the desired behavior).

### Pitfall 4: Test Database Already at Version 7
**What goes wrong:** Running tests twice against the same database file means the migration does not execute on the second run, hiding migration bugs.
**Why it happens:** Tests use `t.TempDir()` which provides a fresh directory, so this should not occur in practice.
**How to avoid:** Continue using `t.TempDir()` for test databases. Also write a test that opens an existing v6 database and verifies migration to v7.
**Warning signs:** Migration test passes on first run but fails on second run against same DB.

### Pitfall 5: Priority Value Validation at Wrong Layer
**What goes wrong:** Invalid priority values (e.g., -1, 5, 99) are stored in the database, causing display bugs later.
**Why it happens:** No validation at the store layer.
**How to avoid:** The store layer should clamp or reject invalid values. However, the current codebase does NOT validate other fields at the store level (e.g., `datePrecision` is not validated in `Add()`). For consistency, do NOT add store-level validation now. Validation will happen at the UI layer (Phase 32) where the edit form restricts input to 0-4. The store trusts its callers, matching the existing pattern.
**Warning signs:** None in this phase -- validation is deferred to the UI phase.

## Code Examples

### Migration v7

```go
// Add to migrate() in internal/store/sqlite.go, after the version < 6 block
if version < 7 {
    if _, err := s.db.Exec(`ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0`); err != nil {
        return fmt.Errorf("add priority column: %w", err)
    }
    if _, err := s.db.Exec(`PRAGMA user_version = 7`); err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

No post-migration fixup needed. `DEFAULT 0` = no priority, which is correct for all existing todos.

### Todo Struct Extension

```go
// In internal/store/todo.go
type Todo struct {
    ID            int    `json:"id"`
    Text          string `json:"text"`
    Body          string `json:"body,omitempty"`
    Date          string `json:"date,omitempty"`
    Done          bool   `json:"done"`
    CreatedAt     string `json:"created_at"`
    SortOrder     int    `json:"sort_order,omitempty"`
    ScheduleID    int    `json:"schedule_id,omitempty"`
    ScheduleDate  string `json:"schedule_date,omitempty"`
    DatePrecision string `json:"date_precision"`
    Priority      int    `json:"priority"`  // 0=none, 1=P1, 2=P2, 3=P3, 4=P4
}

// HasPriority reports whether the todo has a priority set (1-4).
func (t Todo) HasPriority() bool {
    return t.Priority >= 1 && t.Priority <= 4
}

// PriorityLabel returns "P1"-"P4" for prioritized todos, or "" for no priority.
func (t Todo) PriorityLabel() string {
    if t.Priority >= 1 && t.Priority <= 4 {
        return fmt.Sprintf("P%d", t.Priority)
    }
    return ""
}
```

### Updated todoColumns and scanTodo

```go
// In internal/store/sqlite.go
const todoColumns = "id, text, body, date, done, created_at, sort_order, schedule_id, schedule_date, date_precision, priority"

func scanTodo(scanner interface{ Scan(...any) error }) (Todo, error) {
    var t Todo
    var date sql.NullString
    var done int
    var scheduleID sql.NullInt64
    var scheduleDate sql.NullString
    err := scanner.Scan(&t.ID, &t.Text, &t.Body, &date, &done, &t.CreatedAt, &t.SortOrder, &scheduleID, &scheduleDate, &t.DatePrecision, &t.Priority)
    if err != nil {
        return Todo{}, err
    }
    t.Done = done != 0
    if date.Valid {
        t.Date = date.String
    }
    if scheduleID.Valid {
        t.ScheduleID = int(scheduleID.Int64)
    }
    if scheduleDate.Valid {
        t.ScheduleDate = scheduleDate.String
    }
    return t, nil
}
```

### Updated Interface

```go
// In internal/store/iface.go
type TodoStore interface {
    Add(text string, date string, datePrecision string, priority int) Todo
    // ...
    Update(id int, text string, date string, datePrecision string, priority int)
    // ...
    // AddScheduledTodo signature unchanged -- scheduled todos always get priority 0
    AddScheduledTodo(text, date, body string, scheduleID int) Todo
}
```

### Updated Add Implementation

```go
// In internal/store/sqlite.go
func (s *SQLiteStore) Add(text string, date string, datePrecision string, priority int) Todo {
    createdAt := time.Now().Format(dateFormat)

    var maxOrder int
    _ = s.db.QueryRow("SELECT COALESCE(MAX(sort_order), 0) FROM todos").Scan(&maxOrder)
    sortOrder := maxOrder + 10

    var dateVal any
    if date != "" {
        dateVal = date
    }
    if date == "" {
        datePrecision = ""
    }

    result, err := s.db.Exec(
        "INSERT INTO todos (text, body, date, done, created_at, sort_order, date_precision, priority) VALUES (?, '', ?, 0, ?, ?, ?, ?)",
        text, dateVal, createdAt, sortOrder, datePrecision, priority,
    )
    if err != nil {
        return Todo{}
    }

    id, _ := result.LastInsertId()
    return Todo{
        ID:            int(id),
        Text:          text,
        Date:          date,
        Done:          false,
        CreatedAt:     createdAt,
        SortOrder:     sortOrder,
        DatePrecision: datePrecision,
        Priority:      priority,
    }
}
```

### Updated Update Implementation

```go
func (s *SQLiteStore) Update(id int, text string, date string, datePrecision string, priority int) {
    var dateVal any
    if date != "" {
        dateVal = date
    }
    if date == "" {
        datePrecision = ""
    }
    s.db.Exec("UPDATE todos SET text = ?, date = ?, date_precision = ?, priority = ? WHERE id = ?",
        text, dateVal, datePrecision, priority, id)
}
```

### Updated AddScheduledTodo (include priority in INSERT)

```go
func (s *SQLiteStore) AddScheduledTodo(text, date, body string, scheduleID int) Todo {
    // ... same as before but with priority in the INSERT
    result, err := s.db.Exec(
        "INSERT INTO todos (text, body, date, done, created_at, sort_order, schedule_id, schedule_date, date_precision, priority) VALUES (?, ?, ?, 0, ?, ?, ?, ?, 'day', 0)",
        text, body, dateVal, createdAt, sortOrder, scheduleID, date,
    )
    // ... returned Todo includes Priority: 0
}
```

### Caller Updates (pass 0 for now)

```go
// In internal/todolist/model.go saveAdd() -- line 868
// Before: todo := m.store.Add(text, isoDate, precision)
// After:
todo := m.store.Add(text, isoDate, precision, 0)

// In internal/todolist/model.go saveEdit() -- line 831
// Before: m.store.Update(m.editingID, text, isoDate, precision)
// After:
m.store.Update(m.editingID, text, isoDate, precision, 0)
```

Note: The `0` value is a placeholder. Phase 32 (Priority UI) will wire the actual edit form value (`m.editPriority`) into these calls. For Phase 31, all user-created todos get priority 0, preserving existing behavior exactly.

### Priority Roundtrip Test

```go
func TestPriorityRoundtrip(t *testing.T) {
    s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
    if err != nil {
        t.Fatalf("create store: %v", err)
    }
    defer s.Close()

    // Add with priority
    todo := s.Add("Urgent task", "2026-03-15", "day", 1)
    if todo.Priority != 1 {
        t.Errorf("Add: want priority 1, got %d", todo.Priority)
    }

    // Find roundtrip
    found := s.Find(todo.ID)
    if found == nil {
        t.Fatal("Find returned nil")
    }
    if found.Priority != 1 {
        t.Errorf("Find: want priority 1, got %d", found.Priority)
    }

    // Update priority
    s.Update(todo.ID, "Urgent task", "2026-03-15", "day", 3)
    updated := s.Find(todo.ID)
    if updated.Priority != 3 {
        t.Errorf("Update: want priority 3, got %d", updated.Priority)
    }

    // Verify in Todos() listing
    all := s.Todos()
    var match *Todo
    for i := range all {
        if all[i].ID == todo.ID {
            match = &all[i]
            break
        }
    }
    if match == nil {
        t.Fatal("todo not in Todos()")
    }
    if match.Priority != 3 {
        t.Errorf("Todos: want priority 3, got %d", match.Priority)
    }
}

func TestPriorityDefaultZero(t *testing.T) {
    s, err := NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
    if err != nil {
        t.Fatalf("create store: %v", err)
    }
    defer s.Close()

    // Add without priority (0 = no priority)
    todo := s.Add("Normal task", "2026-03-15", "day", 0)
    if todo.Priority != 0 {
        t.Errorf("want priority 0, got %d", todo.Priority)
    }

    found := s.Find(todo.ID)
    if found.Priority != 0 {
        t.Errorf("Find: want priority 0, got %d", found.Priority)
    }
}

func TestPriorityHelpers(t *testing.T) {
    tests := []struct {
        priority  int
        hasPrio   bool
        label     string
    }{
        {0, false, ""},
        {1, true, "P1"},
        {2, true, "P2"},
        {3, true, "P3"},
        {4, true, "P4"},
        {-1, false, ""},
        {5, false, ""},
    }
    for _, tt := range tests {
        todo := Todo{Priority: tt.priority}
        if got := todo.HasPriority(); got != tt.hasPrio {
            t.Errorf("Priority %d: HasPriority() = %v, want %v", tt.priority, got, tt.hasPrio)
        }
        if got := todo.PriorityLabel(); got != tt.label {
            t.Errorf("Priority %d: PriorityLabel() = %q, want %q", tt.priority, got, tt.label)
        }
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No priority field | Adding priority INTEGER column | This phase (v7) | Enables priority display in Phase 32 |
| Add(text, date, precision) | Add(text, date, precision, priority) | This phase | All callers must pass priority value |

**Deprecated/outdated:**
- Nothing deprecated. This is purely additive to the existing schema and interface.

## Impact Analysis

### Files Modified

| File | Change | Scope |
|------|--------|-------|
| `internal/store/todo.go` | Add `Priority int` field, `HasPriority()`, `PriorityLabel()` | Small -- 2 methods + 1 field |
| `internal/store/iface.go` | Extend `Add()` and `Update()` signatures | Small -- 2 method signatures |
| `internal/store/sqlite.go` | Migration v7, update `todoColumns`, `scanTodo`, `Add`, `Update`, `AddScheduledTodo` | Medium -- 6 touchpoints |
| `internal/store/sqlite_test.go` | Add priority roundtrip tests, update existing test Add/Update calls | Medium -- new tests + call updates |
| `internal/todolist/model.go` | Pass `0` to `store.Add()` and `store.Update()` calls | Small -- 2 call sites |
| `internal/recurring/generate_test.go` | If fakeStore interface changes (it should not -- AddScheduledTodo is unchanged) | None expected |

### Files NOT Modified

| File | Why Not |
|------|---------|
| `internal/theme/theme.go` | No priority colors yet -- that is Phase 32 |
| `internal/todolist/styles.go` | No priority styles yet -- Phase 32 |
| `internal/search/model.go` | No priority display yet -- Phase 32 |
| `internal/calendar/*.go` | No calendar indicator changes -- Phase 32 |
| `internal/recurring/generate.go` | `AddScheduledTodo` interface unchanged |

## Open Questions

1. **Should `saveEdit()` preserve existing priority or reset to 0?**
   - What we know: Phase 31 passes `0` to `Update()` since the UI has no priority field yet
   - What's unclear: When a user edits a todo that has priority (set in a future phase), the Phase 31 code would reset it to 0
   - Recommendation: In Phase 31, pass `0` to `Update()` -- this is correct because no todo can have a non-zero priority until Phase 32 wires the UI. Phase 32 will change `saveEdit()` to pass `m.editPriority` which is populated from the existing todo when entering edit mode. No issue in practice.

2. **Should we add a `UpdatePriority(id, priority)` method?**
   - What we know: The current pattern is to include all mutable fields in `Update()`
   - What's unclear: Phase 32 might benefit from a separate method for inline priority changes
   - Recommendation: Do NOT add `UpdatePriority()` now. The requirements specify "priority set via edit form only" and "no inline priority cycling" (out of scope). If Phase 32 or a future phase needs it, it can be added then. Keep the interface minimal.

## Sources

### Primary (HIGH confidence)
- Codebase: `internal/store/sqlite.go` -- migration pattern v1-v6, `todoColumns`, `scanTodo`, `Add()`, `Update()`, `AddScheduledTodo()` implementations
- Codebase: `internal/store/todo.go` -- current Todo struct (10 fields), helper method pattern
- Codebase: `internal/store/iface.go` -- TodoStore interface (27 methods), Add/Update signatures
- Codebase: `internal/store/sqlite_test.go` -- existing test patterns, TempDir usage
- Codebase: `internal/todolist/model.go` -- `saveAdd()` line 868, `saveEdit()` line 831, current callers of Add/Update
- Codebase: `internal/recurring/generate.go` -- `AddScheduledTodo` call at line 57
- Project: `.planning/research/ARCHITECTURE.md` -- prior architecture research for v2.1 milestone
- Project: `.planning/REQUIREMENTS.md` -- PRIO-08, PRIO-09 requirements

### Secondary (MEDIUM confidence)
- Project: `.planning/research/SUMMARY.md` -- phase breakdown rationale, anti-patterns list

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all changes within existing `internal/store` package
- Architecture: HIGH -- follows exact precedent from migration v6 (date_precision), pattern validated 6 times
- Pitfalls: HIGH -- derived from direct codebase analysis, all integration points identified with line numbers

**Research date:** 2026-02-13
**Valid until:** 2026-03-15 (stable -- data layer patterns do not change rapidly)
