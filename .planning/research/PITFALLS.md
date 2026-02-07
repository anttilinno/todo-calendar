# Pitfalls Research: v1.6 Template Management & Recurring Todos

**Domain:** TUI Calendar v1.6 -- Template management overlay, recurring schedules, auto-creation on launch
**Researched:** 2026-02-07
**Confidence:** HIGH (pitfalls derived from codebase analysis, SQLite date function documentation, and recurring task design patterns)

This document covers pitfalls specific to ADDING template management (edit/delete/rename) and recurring todo scheduling to the existing Go Bubble Tea TUI app (5,209 LOC, 16 completed phases, SQLite backend at PRAGMA user_version 3).

---

## Critical Pitfalls

These cause data corruption, duplicate creation, or require schema redesigns if not addressed upfront.

### Pitfall 1: Duplicate Todo Creation on Repeated App Launch

**What goes wrong:** The "auto-create scheduled todos on app launch for rolling 7-day window" feature creates duplicate todos every time the user launches the app. If the user opens the app three times on Monday, they get three copies of every recurring todo for that day.

**Why it happens:** The naive implementation is: "On startup, iterate over schedules, check if today/tomorrow/.../+6 matches a schedule, call `store.Add()` for each match." There is no mechanism to know whether a todo has already been created for a given schedule+date pair.

**Consequences:** The user's todo list fills with duplicates. Once created, there is no programmatic way to distinguish "real" manually-created todos from auto-created ones, making cleanup impossible without manual deletion.

**Prevention:**

1. **Track creation with a ledger table.** Add a `schedule_instances` (or `recurring_log`) table:
   ```sql
   CREATE TABLE schedule_instances (
       schedule_id INTEGER NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
       date        TEXT NOT NULL,
       todo_id     INTEGER NOT NULL REFERENCES todos(id) ON DELETE SET NULL,
       created_at  TEXT NOT NULL,
       PRIMARY KEY (schedule_id, date)
   );
   ```
   Before creating a todo for schedule S on date D, check `SELECT 1 FROM schedule_instances WHERE schedule_id = ? AND date = ?`. If a row exists, skip. The PRIMARY KEY constraint makes this naturally idempotent.

2. **Alternatively, stamp created todos.** Add a `schedule_id` and `scheduled_date` column to the `todos` table, then check for existing rows before inserting. This is simpler (one table instead of two) but mixes recurring metadata into the todo model, which every consumer then has to understand.

**Recommendation:** Use the ledger table approach (option 1). It keeps the `todos` table clean and unchanged, allows ON DELETE CASCADE to clean up when a schedule is removed, and the PRIMARY KEY constraint prevents duplicates at the database level.

**Warning signs:**
- Auto-creation code uses `store.Add()` without checking for prior creation
- No new table or column to track which schedule+date pairs have been materialized
- Tests only run the creation logic once instead of testing idempotency

**Detection:** Run the app twice in the same terminal session. If any todo appears twice, the deduplication is broken.

**Phase to address:** The phase that implements auto-creation logic (likely the scheduling/auto-create phase). Must be designed before the first line of creation code.

---

### Pitfall 2: "Monthly on the 31st" and Other Calendar Day Edge Cases

**What goes wrong:** A schedule configured as "monthly on the 31st" produces no todos in February (28/29 days), April (30 days), June (30 days), September (30 days), and November (30 days). The user sets up a recurring todo expecting it monthly and it silently skips 5 months per year.

**Why it happens:** The code does a simple check like `if today.Day() == schedule.DayOfMonth` which fails for months shorter than the scheduled day. Similarly, "monthly on the 29th" fails in non-leap-year February.

**Consequences:** Recurring todos silently fail to appear. The user does not notice until they miss a task. There is no error, no warning -- just a missing todo.

**Prevention:**

1. **Clamp to last day of month.** When evaluating whether a schedule should fire for a given date, if the schedule's day-of-month exceeds the number of days in the target month, treat it as the last day. So "monthly on the 31st" fires on Feb 28/29, Apr 30, etc. This matches how most calendar applications (Google Calendar, Apple Calendar) handle this case.

2. **Document the clamping behavior.** When the user sets up a "monthly on 31st" schedule, the UI should indicate "fires on the last day of months with fewer than 31 days" or similar.

3. **Implementation pattern:**
   ```go
   func shouldFireOnDate(schedule Schedule, date time.Time) bool {
       if schedule.Type != Monthly {
           // ... handle other types
       }
       lastDay := daysInMonth(date.Year(), date.Month())
       targetDay := schedule.DayOfMonth
       if targetDay > lastDay {
           targetDay = lastDay
       }
       return date.Day() == targetDay
   }
   ```

4. **Do NOT use SQLite's date arithmetic for this.** SQLite's `+N months` modifier has a "ceiling" default that wraps overflow into the next month (Jan 31 + 1 month = Mar 03 in non-leap years), which is the opposite of what you want. Use Go's own date logic.

**Warning signs:**
- Schedule evaluation uses `==` on day-of-month without bounds checking
- No test cases for February, April, June, September, November with day 29/30/31
- Using SQLite `date()` function with `+1 month` for schedule generation

**Detection:** Create a "monthly on 31st" schedule and check what happens in February.

**Phase to address:** The phase that implements schedule matching/evaluation logic. Must include explicit test cases for short months.

---

### Pitfall 3: Template Deletion Orphans Active Schedules

**What goes wrong:** User creates a template "Daily Standup", attaches a recurring schedule, then deletes the template from the template management overlay. The schedule row still references the deleted template ID. Next app launch, the auto-creation code tries to find the template to materialize it, gets a nil/error, and either crashes or silently creates empty todos.

**Why it happens:** The existing `DeleteTemplate` method does a simple `DELETE FROM templates WHERE id = ?` with no awareness of related data. The current codebase has `foreign_keys(ON)` in the DSN (verified in `sqlite.go` line 29), so foreign key constraints will enforce referential integrity -- but only if the schema actually declares the foreign key relationship.

**Consequences:** If FK constraints are properly declared with ON DELETE CASCADE, deleting a template auto-removes its schedules (safe but potentially surprising). If FK constraints are NOT declared (just an integer column), orphaned schedule rows cause runtime errors or ghost todos. If the code catches the nil template and skips, the user sees a schedule in the schedule list that does nothing -- confusing.

**Prevention:**

1. **Declare the FK relationship with ON DELETE CASCADE in the schedules table schema:**
   ```sql
   CREATE TABLE schedules (
       id          INTEGER PRIMARY KEY AUTOINCREMENT,
       template_id INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
       ...
   );
   ```
   This ensures deleting a template automatically removes its schedules. The app already has `_pragma=foreign_keys(ON)` in the DSN, so this will work.

2. **Warn the user on delete.** In the template management overlay, when deleting a template that has associated schedules, show a confirmation or at minimum a visual indicator that schedules will also be removed.

3. **Also cascade to the schedule_instances ledger.** If a schedule is deleted (via cascade from template deletion), the ledger entries should also cascade-delete so that re-creating the same template+schedule does not think instances already exist.

**Warning signs:**
- The `schedules` table uses `template_id INTEGER NOT NULL` without a REFERENCES clause
- Template deletion does not check for or mention schedules
- No test for: create template, add schedule, delete template, verify schedule gone

**Detection:** Delete a template that has a schedule. Check `SELECT * FROM schedules` directly -- orphaned rows mean the FK is not working.

**Phase to address:** The schema migration phase. The FK must be in the CREATE TABLE DDL, not added as an afterthought.

---

### Pitfall 4: Placeholder Prompting Blocks Batch Auto-Creation

**What goes wrong:** The "auto-create on launch for rolling 7-day window" feature creates multiple todos at once (potentially 7+ days times N schedules). If any of those templates have unfilled placeholders, the system needs to prompt the user for values. But the TUI has not even rendered yet -- the user is seeing "Initializing..." while the startup code tries to do interactive prompting.

**Why it happens:** The existing placeholder workflow (in `todolist/model.go`) is deeply embedded in the TUI's mode state machine (`placeholderInputMode`). It prompts one placeholder at a time using the text input. This works for one-shot manual template usage but breaks completely when batch-creating N todos on startup before the TUI event loop is running.

**Consequences:** Either (a) the startup hangs waiting for user input that cannot be shown, (b) todos are created with empty placeholder values making the body useless, or (c) placeholder prompting is skipped entirely, defeating the purpose.

**Prevention:**

1. **Separate the two creation paths.** Templates without placeholders can be auto-created silently on startup (the body content is fully determined). Templates WITH placeholders should either:
   - **Queue for prompting after TUI launch.** Store a list of "pending placeholder prompts" and present them to the user after the main view renders, one at a time.
   - **Pre-fill with date-based defaults.** For auto-created recurring todos, substitute `{{.Date}}` with the scheduled date automatically. Only prompt for genuinely user-specific values.

2. **Design the schedule schema to support pre-filled values.** Allow the user to fill in placeholder values when setting up the schedule (not at creation time). Store these as JSON in the schedule row:
   ```sql
   CREATE TABLE schedules (
       ...
       placeholder_values TEXT NOT NULL DEFAULT '{}',  -- JSON: {"Topic": "Standup", "Date": "auto"}
       ...
   );
   ```
   When auto-creating, use these stored values. The "Daily Plan" template has zero placeholders and needs no values. The "Meeting Notes" template could have `{"Topic": "Standup"}` pre-filled and `{"Date": "auto"}` meaning substitute the scheduled date.

3. **"auto" date placeholder.** Since recurring todos are date-bound by nature, the `{{.Date}}` placeholder can be auto-filled with the target date. This eliminates the most common placeholder case without user interaction.

**Recommendation:** Option 2 (pre-fill at schedule setup time) is the cleanest. It means auto-creation is always non-interactive. The user fills in placeholders once when creating the schedule, not on every occurrence.

**Warning signs:**
- Auto-creation code calls into the TUI's placeholder prompting flow
- No distinction between "templates with placeholders" and "templates without placeholders" in the scheduling logic
- Startup code blocks on user input

**Detection:** Create a schedule using a template with placeholders (e.g., "Meeting Notes" with `{{.Topic}}` and `{{.Date}}`). Restart the app and check what the created todo's body contains.

**Phase to address:** Must be designed in the schedule setup phase. The placeholder values must be captured when the user creates the schedule, not when the todo is materialized.

---

## Moderate Pitfalls

These cause technical debt, confusing UX, or require rework but are recoverable.

### Pitfall 5: TodoStore Interface Grows Unwieldy

**What goes wrong:** Adding schedule CRUD (CreateSchedule, ListSchedules, UpdateSchedule, DeleteSchedule, ListSchedulesForTemplate, MaterializeScheduledTodos, etc.) to the existing `TodoStore` interface bloats it from 17 methods to 25+. The JSON Store (which still exists in the codebase as a compile-time interface check) needs stub implementations for every new method, even though it will never support schedules.

**Why it happens:** The existing `TodoStore` interface already includes template methods (`AddTemplate`, `ListTemplates`, `FindTemplate`, `DeleteTemplate`). The natural impulse is to keep adding schedule methods to the same interface.

**Prevention:**

1. **Consider a separate ScheduleStore interface** for schedule-specific operations. The auto-creation logic only needs ScheduleStore, not the full TodoStore. This follows the Interface Segregation Principle.

2. **Alternatively, accept the growth.** For a personal-use app with one real implementation (SQLite), a single interface with 25 methods is not a maintainability crisis. Just add the JSON Store stubs and move on. Do not over-engineer.

3. **If keeping one interface:** Group the new methods logically in the interface definition with comments separating todo, template, and schedule method groups.

**Recommendation:** Keep the single `TodoStore` interface and add stubs to the JSON Store. This app has one real backend. The interface segregation principle matters less when there is only one consumer of each method group and one implementation.

**Warning signs:**
- Spending more time on interface design than on the actual schedule logic
- Creating multiple interfaces that only one type implements

**Phase to address:** The schema/store phase. Decide the interface strategy before implementing.

---

### Pitfall 6: Schedule Table Schema Missing Key Fields

**What goes wrong:** The schedule schema is designed without enough fields, requiring a migration in a later phase. Common missing fields: `enabled` (to pause without deleting), `last_created_date` (to track what has been materialized), `created_at` (for audit), or `name` (to identify the schedule independently of the template name).

**Why it happens:** The initial design focuses on the happy path (create schedule, it fires, todos appear) and does not consider the management lifecycle (pause, resume, view history, debug why something did or did not fire).

**Prevention:** Include these fields from the start in the schema:

```sql
CREATE TABLE schedules (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id        INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    schedule_type      TEXT NOT NULL,          -- 'daily', 'weekdays', 'weekly', 'monthly'
    schedule_value     TEXT NOT NULL DEFAULT '', -- e.g., '1,3,5' for Mon/Wed/Fri, '15' for monthly
    placeholder_values TEXT NOT NULL DEFAULT '{}',
    enabled            INTEGER NOT NULL DEFAULT 1,
    created_at         TEXT NOT NULL
);
```

The `schedule_instances` ledger (from Pitfall 1) replaces the need for `last_created_date` since you can query `MAX(date) FROM schedule_instances WHERE schedule_id = ?`.

**Warning signs:**
- Schema has only id, template_id, and a type field
- No way to pause a schedule without deleting and recreating it
- Adding columns via ALTER TABLE in later phases

**Phase to address:** The schema migration phase. Design the full schema upfront.

---

### Pitfall 7: Template Rename Breaks UNIQUE Constraint Silently

**What goes wrong:** The templates table has `name TEXT NOT NULL UNIQUE` (from the v1.4 migration). If the user renames a template to a name that already exists, the SQLite UPDATE fails with a UNIQUE constraint violation. If the error is swallowed (as the existing store code tends to do -- see `Toggle`, `Delete`, etc. which ignore errors), the rename silently fails and the user thinks it worked.

**Why it happens:** The existing store pattern is to silently ignore errors from `db.Exec()`. This works fine for most operations but is dangerous for operations with constraints.

**Prevention:**

1. **Return an error from the rename method.** Unlike Toggle/Delete, rename can fail in a user-actionable way (pick a different name). The new `UpdateTemplate(id int, name, content string) error` method should return the error.

2. **Check for duplicates in the UI layer before calling the store.** Query `ListTemplates()`, check if any other template has the target name. This gives a better user experience (inline error message) than a database error.

3. **Show an error message in the overlay** when the rename fails due to duplicate name. The existing overlay pattern (settings, search) supports rendering error states.

**Warning signs:**
- `UpdateTemplate` does not return an error
- No test for renaming to a duplicate name
- The overlay has no error display mechanism

**Phase to address:** The template management overlay phase.

---

### Pitfall 8: Overlay State Leaks When Switching Between Template Management and Schedule Setup

**What goes wrong:** The template management overlay lets the user view/edit/delete templates. Adding a schedule to a template requires a sub-flow (select recurrence type, configure days, optionally fill placeholders). If this sub-flow is implemented as a mode within the overlay, switching between "viewing template list" and "editing schedule for template X" corrupts state (wrong template selected, schedule partially configured, cursor position lost).

**Why it happens:** The existing `todolist/model.go` already has 10 modes in its mode enum. Adding template management as another overlay (like settings/search/preview) with its own internal modes creates nested state machines. The Bubble Tea model is a value type -- every Update returns a new copy. If the overlay's internal state is not properly initialized and cleared, stale data persists.

**Prevention:**

1. **Follow the established overlay pattern.** Settings, search, and preview are all separate packages with their own Model, Update, View, and a closing message type. Template management should follow this exact pattern: `internal/templates/model.go` with its own state machine.

2. **Use explicit sub-modes within the template overlay.** The overlay should have clear modes: `listMode`, `editMode`, `scheduleSetupMode`, `confirmDeleteMode`. Each mode transition should fully initialize the target mode's state.

3. **Create fresh overlay state on open.** Follow the search overlay pattern (`search.New()` called on Ctrl+F) where a fresh model is created each time the overlay opens. Never reuse stale overlay state.

**Warning signs:**
- Template management shares state with the todolist model instead of being its own package
- Mode transitions that only set `m.mode = newMode` without initializing mode-specific fields
- Stale cursor positions or selected templates after closing and reopening the overlay

**Phase to address:** The template management overlay phase. Architectural decision about overlay vs. inline must be made first.

---

### Pitfall 9: Rolling Window Creates Todos Too Far in the Future

**What goes wrong:** The "rolling 7-day window" creates todos for days the user has not navigated to yet. The user opens the app on Monday and suddenly sees 7 new recurring todos spanning Monday through Sunday. This clutters the current view and creates cognitive overload, especially if multiple schedules are active.

**Why it happens:** A 7-day window seemed reasonable in theory, but in practice it means the user is always looking at pre-created tasks they cannot act on yet (Wednesday's standup on Monday morning).

**Prevention:**

1. **Start with today + 1 day (or just today).** Only create todos for today and tomorrow. The user opens the app, sees today's recurring todos, and tomorrow's as a heads-up. This is sufficient for a personal todo app.

2. **Make the window configurable.** Default to 1 day, allow the user to increase via config/settings if they want weekly pre-creation.

3. **Create on-demand as the user navigates months.** Instead of pre-creating for a fixed window, materialize recurring todos when the user navigates to a month that has schedules. This avoids the "startup burst" entirely. However, this means todos only exist after the user views the month, which is fine for a personal app.

**Recommendation:** Start with "today only" creation on app launch. Expand to a configurable window only if the user requests it. Simpler is better for a personal tool.

**Warning signs:**
- User opens app and sees a wall of new recurring todos
- Todos for future days cannot be acted on meaningfully
- The creation burst takes noticeable time on startup

**Phase to address:** The auto-creation phase. The window size is a UX decision that should be validated early.

---

### Pitfall 10: Weekday Schedule Mishandles time.Weekday Numbering

**What goes wrong:** Go's `time.Weekday` uses Sunday=0, Monday=1, ..., Saturday=6. The user selects "Monday, Wednesday, Friday" in the UI and the code stores `1,3,5`. But the UI display maps these to different days depending on whether the developer assumed 0-indexed or 1-indexed, or ISO weekday numbering (Monday=1, ..., Sunday=7).

**Why it happens:** Three common weekday numbering schemes exist:
- Go's `time.Weekday`: Sunday=0, Monday=1, ..., Saturday=6
- ISO 8601: Monday=1, Tuesday=2, ..., Sunday=7
- Cron-style: Sunday=0, Monday=1, ..., Saturday=6 (same as Go but sometimes Sunday=7 too)

If the UI presents days in Monday-first order (matching the app's `first_day_of_week` config) but stores Go weekday numbers, the mapping gets confused. Especially since this app supports both Sunday-start and Monday-start weeks.

**Prevention:**

1. **Store Go `time.Weekday` integer values directly** (0-6). This is what the schedule evaluation code will use (`time.Now().Weekday()`), so the storage format matches the runtime check exactly.

2. **In the UI, map display order to storage values explicitly.** If the user's week starts on Monday, display Mon/Tue/Wed/Thu/Fri/Sat/Sun but store 1/2/3/4/5/6/0. Do not assume display order equals storage order.

3. **Write explicit tests** for: Sunday=0 schedule when week starts on Monday, Saturday=6 schedule when week starts on Sunday, "weekdays" (Mon-Fri = 1,2,3,4,5) preset.

**Warning signs:**
- Schedule stores day names as strings ("Monday") instead of integers
- No mapping layer between display order and storage order
- Off-by-one in weekday selection when `first_day_of_week` changes

**Phase to address:** The schedule setup UI phase. The weekday mapping must account for the existing `first_day_of_week` configuration.

---

## Minor Pitfalls

These are annoyances or polish issues, not architectural problems.

### Pitfall 11: Template Edit Loses Unsaved Changes on Accidental Escape

**What goes wrong:** User is editing a template's content in the template management overlay, accidentally presses Escape, and all changes are lost with no confirmation.

**Prevention:** Follow the settings overlay pattern which has save/cancel semantics. Store original values and offer a "discard changes?" confirmation if the content has been modified.

**Phase to address:** Template management overlay phase.

---

### Pitfall 12: Schedule Display in Template List Obscures Template Content

**What goes wrong:** The template management overlay shows template names and content previews. Adding schedule information (recurrence type, enabled/disabled) to each row makes the list too dense to read in a terminal.

**Prevention:** Show schedules as a detail view when a template is selected (cursor on it), not inline in the list. Use the existing pattern from search results where the selected item shows expanded details.

**Phase to address:** Template management overlay phase.

---

### Pitfall 13: Migration Version Gap Creates Confusion

**What goes wrong:** The current schema is at PRAGMA user_version = 3. If this milestone adds multiple migrations (e.g., 4 for schedules table, 5 for schedule_instances table), and a future developer (or the user's pre-existing DB) misses one, the sequential `if version < N` pattern might apply later migrations without earlier ones.

**Prevention:** This is already handled correctly by the existing migration pattern -- the `if version < N` blocks in `migrate()` run sequentially and each increments the version. As long as new migrations follow the same pattern, this is safe. Just ensure all DDL for a single migration is in one `if version < N` block, not split across multiple blocks.

**Warning signs:**
- Two separate `if version < N` blocks that both need to succeed for the schema to be consistent
- DDL for related tables (schedules + schedule_instances) split across different migration versions

**Recommendation:** Put the entire schedule schema (schedules table, schedule_instances table, indexes) in a single migration block (`if version < 4`).

**Phase to address:** The schema migration phase.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation | Severity |
|-------------|---------------|------------|----------|
| Schema migration | FK not declared, orphaned schedules | Always use REFERENCES with ON DELETE CASCADE | Critical |
| Schema migration | Incomplete schema, missing enabled/placeholder_values fields | Design full schema upfront per Pitfall 6 | Moderate |
| Schema migration | Split migrations for related tables | Single migration block for all schedule DDL | Minor |
| Template management overlay | State leaks between modes | Separate package following settings/search pattern | Moderate |
| Template management overlay | Rename to duplicate name silently fails | Return error, show in UI | Moderate |
| Template management overlay | Unsaved edit loss on Escape | Confirm discard if modified | Minor |
| Schedule setup UI | Weekday numbering mismatch | Store Go time.Weekday values, explicit display mapping | Moderate |
| Schedule setup UI | Placeholder values not captured upfront | Pre-fill at schedule creation time | Critical |
| Schedule evaluation | Monthly day > days-in-month | Clamp to last day of month | Critical |
| Schedule evaluation | Weekday off-by-one | Test all 7 days with both week-start settings | Moderate |
| Auto-creation logic | Duplicate todos on re-launch | Ledger table with PRIMARY KEY(schedule_id, date) | Critical |
| Auto-creation logic | Too many future todos created | Start with today-only, expand if needed | Moderate |
| TodoStore interface | Interface bloat | Accept it for single-backend app, add stubs | Moderate |

---

## Integration Risks with Existing System

### Risk 1: Existing Template Deletion Path Needs Schedule Awareness

The current `DeleteTemplate` in `sqlite.go` (line 455-457) does a bare `DELETE FROM templates WHERE id = ?`. Once schedules reference templates via FK, this delete will either cascade (if FK is declared) or fail (if FK constraint is violated). The existing template deletion in `templateSelectMode` (todolist/model.go line 759-773) calls `m.store.DeleteTemplate()` without any schedule awareness.

**Mitigation:** If ON DELETE CASCADE is used, the existing code works correctly without changes -- deleting a template auto-removes its schedules. This is the path of least disruption. Just make sure the FK is declared in the schema.

### Risk 2: TodoStore Interface Must Add Methods for Both Implementations

Every new method added to `TodoStore` must also be stubbed in the JSON `Store` (store.go). Currently there are 4 template stubs (lines 202-217). Schedule methods will need similar stubs. This is mechanical but easy to forget, causing a compile error.

**Mitigation:** Run `go build ./...` after every interface change. The compile-time check `var _ TodoStore = (*Store)(nil)` on line 40 of store.go will catch missing implementations immediately.

### Risk 3: App Startup Sequence Needs Schedule Processing Hook

The current `main.go` startup is: config -> provider -> DB path -> SQLiteStore -> theme -> app.New -> tea.Run. There is no hook between "store is ready" and "TUI starts" where schedule materialization can run. The auto-creation logic needs to execute after the store is open but before (or immediately after) the TUI event loop starts.

**Mitigation:** Add the auto-creation call in `main.go` between `NewSQLiteStore()` and `app.New()`. This keeps it synchronous and simple. If the creation takes noticeable time (unlikely for a personal app), it can be moved to `Init()` as a Cmd.

---

## Sources

- [SQLite Date and Time Functions](https://sqlite.org/lang_datefunc.html) -- verified month arithmetic edge cases with floor/ceiling modifiers (HIGH confidence)
- Codebase analysis of `/home/antti/Repos/Misc/todo-calendar/internal/store/sqlite.go` -- FK pragma, migration pattern, existing template methods (HIGH confidence)
- Codebase analysis of `/home/antti/Repos/Misc/todo-calendar/internal/todolist/model.go` -- mode state machine, placeholder workflow, template selection (HIGH confidence)
- Codebase analysis of `/home/antti/Repos/Misc/todo-calendar/internal/app/model.go` -- overlay routing pattern, startup sequence (HIGH confidence)
- [Idempotency patterns](https://temporal.io/blog/idempotency-and-durable-execution) -- deduplication strategies (MEDIUM confidence, pattern-level)
- [Recurring event database storage patterns](https://www.codegenes.net/blog/calendar-recurring-repeating-events-best-storage-method/) -- hybrid table design for recurrence rules (MEDIUM confidence)
- [Bubble Tea overlay patterns](https://leg100.github.io/en/posts/building-bubbletea-programs/) -- state management for TUI modals (MEDIUM confidence)
