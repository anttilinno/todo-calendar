# Stack Research: v1.6 Recurring Todos & Template Management

**Domain:** Recurring schedule storage, auto-creation logic, template management overlay
**Researched:** 2026-02-07
**Confidence:** HIGH

## Executive Summary

v1.6 requires **zero new external dependencies**. Every capability needed -- schedule storage, cadence representation, auto-creation logic, template management UI -- can be built with the existing stack (Go stdlib, SQLite, Bubble Tea, existing overlay patterns). The key decisions are about **data modeling**, not library selection.

The three decisions that matter:
1. How to represent recurring cadences in SQLite (custom TEXT columns, not cron, not RRULE)
2. Where to hook auto-creation logic (store layer at startup, not Bubble Tea Init)
3. What new methods the TodoStore interface needs (6 new methods for schedules + 2 for templates)

Total new direct dependencies: **0**.

---

## Decision 1: Cadence Representation -- Custom Columns

### Recommendation: Structured TEXT columns in a `schedules` table

Store cadence as a combination of `cadence_type TEXT` + `cadence_value TEXT`:

```sql
CREATE TABLE schedules (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id   INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    cadence_type  TEXT    NOT NULL,  -- 'daily' | 'weekdays' | 'monthly'
    cadence_value TEXT    NOT NULL DEFAULT '',  -- '' | 'mon,wed,fri' | '15'
    last_created  TEXT,              -- ISO date of most recent auto-created todo
    created_at    TEXT    NOT NULL
);
```

**How each cadence maps:**

| User Intent | `cadence_type` | `cadence_value` | Expansion Logic |
|-------------|----------------|-----------------|-----------------|
| Every day | `daily` | `""` (empty) | Every date in window |
| Every Monday and Friday | `weekdays` | `"mon,fri"` | Filter window dates by `time.Weekday()` |
| Every weekday (Mon-Fri) | `weekdays` | `"mon,tue,wed,thu,fri"` | Same filter, all 5 days |
| 15th of every month | `monthly` | `"15"` | Check if day 15 falls in window |

### Why NOT cron expressions

Considered `robfig/cron` and raw cron strings (e.g., `0 0 * * 1,3,5`). Rejected because:

1. **Overkill.** Cron encodes hours, minutes, seconds, and month -- none of which matter for a daily/weekly/monthly todo app. The v1.6 requirements are "daily", "weekday(s)", "monthly on Nth". Cron's expressiveness is wasted complexity.
2. **Parser dependency.** Would need `robfig/cron` or similar just to parse strings. That is a new dependency for string parsing that custom code handles in 20 lines.
3. **Opaque to users.** If we ever expose the cadence in UI (e.g., "Recurs: every Mon, Wed, Fri"), parsing cron back to human-readable text is harder than parsing `"mon,wed,fri"`.
4. **No time-of-day component.** Cron's primary purpose -- scheduling at specific times -- is irrelevant. Todos are date-level, not time-level.

**Confidence:** HIGH. The project already stores dates as plain TEXT strings (YYYY-MM-DD). Using TEXT for cadence_type and cadence_value is consistent with the existing design philosophy of simple, debuggable storage.

### Why NOT RFC 5545 RRULE

Considered `teambition/rrule-go` (369 stars, last updated Oct 2023, maintenance status: inactive) and `stephens2424/rrule-go` (17 stars, self-described as not production-tested). Rejected because:

1. **Massive over-engineering.** RRULE supports "every 2nd Tuesday", "last Friday of month", "every 3rd week until Dec 2027", "yearly on Feb 29 with Feb 28 fallback". The v1.6 requirements are three cadence types. RRULE's complexity buys nothing here.
2. **Both Go libraries are undermaintained.** teambition/rrule-go is inactive (Snyk advisory). stephens2424/rrule has 17 stars and no production usage. Neither inspires confidence for a long-lived project.
3. **v2 escape hatch exists.** The requirements explicitly defer complex cadences ("every 2nd Tuesday", "last Friday of month") to v2. If v2 needs RRULE-level power, the `cadence_type`/`cadence_value` columns can be extended or a migration can convert existing rows. The simple v1.6 schema does not paint us into a corner.
4. **Debugging difficulty.** An RRULE string like `FREQ=WEEKLY;BYDAY=MO,WE,FR;INTERVAL=1` is harder to inspect in `sqlite3` CLI than `cadence_type='weekdays', cadence_value='mon,wed,fri'`.

**Confidence:** HIGH. Both RRULE libraries were reviewed. Neither is well-maintained. Custom columns are simpler, more debuggable, and sufficient for the stated requirements.

### Why NOT JSON blob in a single column

Considered storing cadence as `{"type": "weekdays", "days": ["mon", "wed", "fri"]}` in a single TEXT column.

1. **Not queryable.** SQLite's JSON functions exist but are verbose for simple lookups. Separate columns allow `WHERE cadence_type = 'daily'` directly.
2. **Schema enforcement.** Separate columns make the data model explicit. A JSON blob hides structure.
3. **Consistency.** The project uses structured columns everywhere else (todos, templates). No JSON blobs in the schema.

---

## Decision 2: Auto-Creation Logic -- Store Layer at Startup

### Recommendation: `SQLiteStore.CreateScheduledTodos(windowDays int)` called from `main.go`

The auto-creation hook belongs in the store layer, called once at startup before the TUI launches:

```go
// In main.go, after NewSQLiteStore:
s, err := store.NewSQLiteStore(dbPath)
// ...
s.CreateScheduledTodos(7) // rolling 7-day window
```

**Why the store layer, not Bubble Tea Init:**

1. **Deterministic timing.** `main.go` runs synchronously before `tea.NewProgram`. Todos are created before the UI renders. No race conditions, no loading states, no flickering.
2. **No Bubble Tea coupling.** The schedule logic is pure data: read schedules, compute dates, insert todos. It does not need terminal size, key events, or message passing. Putting it in `Init()` or `Update()` would couple it to the TUI lifecycle unnecessarily.
3. **Testable.** A `CreateScheduledTodos(windowDays int)` method on SQLiteStore is trivially testable: create a store in `t.TempDir()`, add a schedule, call the method, assert todos exist. No Bubble Tea test harness needed.
4. **Idempotent via `last_created` tracking.** The `last_created` column in the schedules table stores the ISO date of the most recently created todo for that schedule. On each call, the method only creates todos for dates after `last_created` and within the window. Running the app multiple times per day is safe.

**Auto-creation algorithm:**

```
for each schedule:
    start_date = max(last_created + 1 day, today)
    end_date = today + windowDays
    for each date in [start_date, end_date]:
        if schedule matches date:
            create todo from template (with placeholder prompting deferred)
            update schedule.last_created = date
```

**Placeholder prompting for auto-created todos:** Auto-created recurring todos should use the template content directly with empty placeholders (or auto-fill date-like placeholders from context). Interactive placeholder prompting at startup would block the app launch and annoy users. Instead:

- Auto-fill `{{.Date}}` with the target date
- Leave other placeholders as literal `{{.VarName}}` text in the body
- User can edit the body later via the existing `$EDITOR` integration

The milestone requirement says "Placeholder prompting for auto-created recurring todos on first launch." This should be interpreted as: on the first app launch after a schedule is created, if the template has unfilled placeholders, prompt the user via a one-time overlay (similar to how templateSelectMode works). Subsequent auto-creations for the same schedule skip prompting and reuse the previously-provided values stored alongside the schedule.

### Deduplication Strategy

The `last_created` column is the dedup mechanism. No need for a separate `schedule_instances` table or unique constraints on (schedule_id, date). The linear scan from `last_created + 1` to `today + 7` is at most 7 iterations per schedule -- negligible cost.

Edge case: If the user deletes an auto-created todo, it should NOT be re-created. The `last_created` marker has already advanced past that date. This is the correct behavior -- the user explicitly deleted it.

---

## Decision 3: TodoStore Interface Extensions

### New methods needed for schedules (6 methods):

```go
// Schedule CRUD
AddSchedule(templateID int, cadenceType string, cadenceValue string) (Schedule, error)
ListSchedules() []Schedule
FindSchedule(id int) *Schedule
UpdateSchedule(id int, cadenceType string, cadenceValue string)
DeleteSchedule(id int)

// Auto-creation
CreateScheduledTodos(windowDays int) []Todo
```

### New methods needed for template management (2 methods):

The existing interface has `AddTemplate`, `ListTemplates`, `FindTemplate`, `DeleteTemplate`. Missing for the template management overlay:

```go
UpdateTemplate(id int, name string, content string) error  // rename + edit content
```

One method suffices: `UpdateTemplate` handles both rename and content edit. The overlay calls it with new name and/or new content. If only renaming, pass the existing content. If only editing content, pass the existing name. This avoids splitting into `RenameTemplate` + `UpdateTemplateContent` for a trivial operation.

Also needed for schedule placeholder storage:

```go
// Store placeholder defaults alongside a schedule so auto-creation
// can reuse previously-provided values without re-prompting.
// This is stored as a JSON map in the schedules table itself
// (placeholder_defaults TEXT NOT NULL DEFAULT '{}')
```

This means the schedules table gains one more column:

```sql
CREATE TABLE schedules (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id           INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    cadence_type          TEXT    NOT NULL,
    cadence_value         TEXT    NOT NULL DEFAULT '',
    placeholder_defaults  TEXT    NOT NULL DEFAULT '{}',  -- JSON: {"Date": "auto", "Topic": "Weekly Standup"}
    last_created          TEXT,
    created_at            TEXT    NOT NULL
);
```

The `placeholder_defaults` column is the one place where a JSON blob is justified: placeholder names and counts vary per template, and this is write-once-read-many data that does not need to be queried by individual fields.

### Schedule struct:

```go
type Schedule struct {
    ID                  int
    TemplateID          int
    CadenceType         string            // "daily" | "weekdays" | "monthly"
    CadenceValue        string            // "" | "mon,wed,fri" | "15"
    PlaceholderDefaults map[string]string  // deserialized from JSON
    LastCreated         string            // ISO date or ""
    CreatedAt           string
}
```

### Full updated interface:

```go
type TodoStore interface {
    // Existing todo methods (unchanged)
    Add(text string, date string) Todo
    Toggle(id int)
    Delete(id int)
    Find(id int) *Todo
    Update(id int, text string, date string)
    Todos() []Todo
    TodosForMonth(year int, month time.Month) []Todo
    FloatingTodos() []Todo
    IncompleteTodosPerDay(year int, month time.Month) map[int]int
    TotalTodosPerDay(year int, month time.Month) map[int]int
    TodoCountsByMonth() []MonthCount
    FloatingTodoCounts() FloatingCount
    UpdateBody(id int, body string)
    SwapOrder(id1, id2 int)
    SearchTodos(query string) []Todo
    EnsureSortOrder()
    Save() error

    // Existing template methods (unchanged)
    AddTemplate(name, content string) (Template, error)
    ListTemplates() []Template
    FindTemplate(id int) *Template
    DeleteTemplate(id int)

    // NEW: template management
    UpdateTemplate(id int, name string, content string) error

    // NEW: schedule CRUD
    AddSchedule(templateID int, cadenceType string, cadenceValue string) (Schedule, error)
    ListSchedules() []Schedule
    FindSchedule(id int) *Schedule
    UpdateSchedule(id int, cadenceType string, cadenceValue string)
    DeleteSchedule(id int)

    // NEW: auto-creation
    CreateScheduledTodos(windowDays int) []Todo
}
```

The JSON store stubs will return errors/nil/no-ops for the new methods, consistent with how template methods are already stubbed.

---

## Decision 4: Schema Migration -- Version 4

### Recommendation: `PRAGMA user_version = 4` adding schedules table

The project is currently at `user_version = 3` (v1-v2 created todos + templates tables, v3 seeded default templates). The new migration:

```go
if version < 4 {
    _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schedules (
        id                    INTEGER PRIMARY KEY AUTOINCREMENT,
        template_id           INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
        cadence_type          TEXT    NOT NULL,
        cadence_value         TEXT    NOT NULL DEFAULT '',
        placeholder_defaults  TEXT    NOT NULL DEFAULT '{}',
        last_created          TEXT,
        created_at            TEXT    NOT NULL
    )`)
    if err != nil {
        return fmt.Errorf("create schedules table: %w", err)
    }
    _, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_schedules_template ON schedules(template_id)`)
    if err != nil {
        return fmt.Errorf("create schedule template index: %w", err)
    }
    _, err = s.db.Exec("PRAGMA user_version = 4")
    if err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

**Foreign key ON DELETE CASCADE:** When a template is deleted, all its schedules are automatically removed. This is correct behavior -- you cannot have a recurring schedule without the template it is based on. The DSN already enables foreign keys: `_pragma=foreign_keys(ON)`.

**No new migration tool needed.** The existing `PRAGMA user_version` pattern continues. One new version step, one new table. Consistent with the established approach.

---

## Decision 5: Template Management Overlay -- Existing Pattern

### Recommendation: New `internal/tmplmgr` package following settings/search/preview overlay pattern

The template management overlay follows the identical pattern used by settings, search, and preview:

1. **Own package:** `internal/tmplmgr/` with `model.go`, `keys.go`, `styles.go`
2. **Own Bubble Tea model:** `tmplmgr.Model` with `Update()` and `View()`
3. **Message-based lifecycle:** `tmplmgr.CloseMsg` to dismiss overlay
4. **Parent integration:** `app.Model` gains `showTmplMgr bool` and `tmplMgr tmplmgr.Model`
5. **SetSize/SetTheme:** Same interface as settings/search/preview

**No new dependencies.** The overlay uses `charmbracelet/bubbles` textinput and textarea (already imported), lipgloss for styling (already imported), and the existing store interface for CRUD.

**Overlay modes (internal state machine):**
- **List mode:** Show templates with cursor, navigate with j/k, select with Enter
- **View mode:** Show selected template content (rendered with glamour)
- **Edit mode:** Textarea for editing content (reuses existing `bubbles/textarea`)
- **Rename mode:** Textinput for new name

This is the same pattern as `todolist.Model`'s mode enum (`normalMode`, `inputMode`, etc.) already used in the codebase. No new architectural pattern needed.

**Schedule management** can live in the same overlay (templates and their schedules are tightly coupled) or as a sub-view within the template detail view. Recommended: keep it in the same overlay. When viewing a template, show its schedule (if any) and allow creating/editing/deleting the schedule. This avoids yet another overlay.

---

## Unchanged Stack Components

| Technology | Version | Role in v1.6 |
|------------|---------|--------------|
| Go | 1.25.6 | `time.Weekday()` for day matching, `encoding/json` for placeholder_defaults, `strconv` for monthly day parsing |
| Bubble Tea | v1.3.10 | Overlay message passing, same lifecycle as settings/search/preview |
| Lipgloss | v1.1.1-0.x | Overlay styling |
| Bubbles | v0.21.1 | textinput for rename/placeholder input, textarea for template content editing |
| Glamour | v0.10.0 | Rendering template content preview in the overlay |
| modernc.org/sqlite | v1.44.3 | New schedules table, foreign key enforcement |
| BurntSushi/toml | v1.6.0 | Unchanged |
| rickar/cal/v2 | v2.1.27 | Unchanged |
| text/template | stdlib | Template execution for auto-created todos (already used by tmpl package) |

---

## Deliberately NOT Adding

| Consideration | Decision | Rationale |
|---------------|----------|-----------|
| **robfig/cron** | Not adding | Cron expressions encode time-of-day precision irrelevant to a date-level todo app. Three cadence types (daily, weekdays, monthly) are trivially implemented with `time.Weekday()` and day-of-month comparison. See Decision 1. |
| **teambition/rrule-go** | Not adding | RRULE library is inactive (Snyk advisory), 369 stars. RFC 5545 is massive overkill for 3 cadence types. Complex cadences deferred to v2. See Decision 1. |
| **stephens2424/rrule** | Not adding | 17 stars, author states no production usage. See Decision 1. |
| **Any ORM** | Not adding | Project uses hand-written SQL. One new table with simple CRUD. Consistent with established pattern. |
| **Background goroutine / ticker** | Not adding | Auto-creation runs once at startup. The app does not need to create todos while running. If the user leaves it open for days, they restart when they want fresh scheduled todos. A background ticker adds concurrency complexity for zero user benefit. |
| **fsnotify (template file watching)** | Not adding | Templates are in SQLite, not the filesystem. No files to watch. |
| **Separate schedule instances table** | Not adding | A `schedule_instances` table tracking each generated (schedule_id, date) pair would enable finer dedup but is unnecessary. The `last_created` column provides sufficient dedup for a linear forward-only schedule. |

---

## Integration Summary

### Where new code lives:

| Location | What | New/Modified |
|----------|------|--------------|
| `internal/store/todo.go` | `Schedule` struct, updated `TodoStore` interface | Modified |
| `internal/store/sqlite.go` | Migration v4, schedule CRUD, `CreateScheduledTodos()` | Modified |
| `internal/store/store.go` | JSON store stubs for new interface methods | Modified |
| `internal/tmplmgr/model.go` | Template management overlay model | New |
| `internal/tmplmgr/keys.go` | Overlay keybindings | New |
| `internal/tmplmgr/styles.go` | Overlay styles | New |
| `internal/app/model.go` | `showTmplMgr` bool, tmplmgr integration | Modified |
| `main.go` | `s.CreateScheduledTodos(7)` call before TUI launch | Modified |

### Data flow for auto-creation:

```
App Launch
    |
    v
main.go: store.NewSQLiteStore(dbPath)
    |
    v
sqlite.go: migrate() -- creates schedules table if needed
    |
    v
main.go: s.CreateScheduledTodos(7)
    |
    v
sqlite.go: for each schedule:
    1. Read last_created
    2. Compute target dates (today..today+7, filtered by cadence)
    3. For each target date > last_created:
       a. Execute template with placeholder_defaults + auto-fill date
       b. Insert todo (text = template.Name, date = target date, body = rendered template)
       c. Update schedule.last_created
    4. Return created todos (for logging/future UI notification)
    |
    v
main.go: app.New(...) -- TUI sees all todos including newly created ones
```

### Data flow for template management overlay:

```
User presses 'T' (or configured key)
    |
    v
app.Model: showTmplMgr = true, tmplMgr = tmplmgr.New(store, theme)
    |
    v
tmplmgr.Model (list mode): shows templates with cursor
    |
    +-- Enter --> view mode (show template content + schedule if any)
    +-- e --> edit mode (textarea with template content)
    +-- r --> rename mode (textinput with template name)
    +-- d --> delete (with confirmation)
    +-- s --> schedule sub-view (create/edit/delete schedule for this template)
    +-- Esc --> close overlay (tmplmgr.CloseMsg --> app.Model)
```

---

## Sources

- [Recurring Events Database Design (Redgate)](https://www.red-gate.com/blog/again-and-again-managing-recurring-events-in-a-data-model) - Schema patterns for recurring events. MEDIUM confidence (general patterns, not Go-specific).
- [RFC 5545 RRULE specification](https://icalendar.org/iCalendar-RFC-5545/3-8-5-3-recurrence-rule.html) - Full recurrence rule format. HIGH confidence (official spec, reviewed to confirm it is overkill for v1.6).
- [teambition/rrule-go on GitHub](https://github.com/teambition/rrule-go) - 369 stars, inactive maintenance. HIGH confidence (verified via Snyk and GitHub).
- [stephens2424/rrule on GitHub](https://github.com/stephens2424/rrule) - 17 stars, no production usage. HIGH confidence (author's own README states this).
- [robfig/cron on pkg.go.dev](https://pkg.go.dev/github.com/robfig/cron) - Cron library for Go. HIGH confidence (reviewed API to confirm time-of-day focus).
- [Database design for recurring tasks (SitePoint)](https://www.sitepoint.com/community/t/database-design-for-daily-weekly-monthly-and-adhoc-tasks/26163) - Community patterns. LOW confidence (forum discussion, not authoritative).
- Existing codebase: `internal/store/sqlite.go`, `internal/store/store.go`, `internal/store/todo.go`, `internal/tmpl/tmpl.go`, `internal/settings/model.go`, `internal/app/model.go`, `internal/todolist/model.go` -- HIGH confidence (direct code inspection).
