# Architecture Research: v1.6 Template Management & Recurring Todos

**Domain:** Template management overlay, recurring schedule definitions, auto-creation of scheduled todos
**Researched:** 2026-02-07
**Confidence:** HIGH for integration patterns (verified against existing codebase); HIGH for schema design (established migration pattern); MEDIUM for rolling window algorithm (pattern-based, straightforward logic)

## Current Architecture Summary (Post v1.5)

```
main.go
  |
  app.Model (root orchestrator)
  |-- calendar.Model    (left pane: grid + overview)
  |-- todolist.Model    (right pane: todo list + input modes)
  |-- settings.Model    (full-screen overlay, showSettings bool)
  |-- search.Model      (full-screen overlay, showSearch bool)
  |-- preview.Model     (full-screen overlay, showPreview bool)
  |-- editor            (external process, editing bool)
  |
  store.SQLiteStore     (TodoStore interface, SQLite with WAL)
  config.Config         (TOML config, Save/Load)
  theme.Theme           (14 semantic color roles)
  holidays.Provider     (rickar/cal wrapper)
  tmpl                  (ExtractPlaceholders, ExecuteTemplate)
```

**Critical facts for v1.6 planning:**

1. **TodoStore interface** has 17 methods including template operations: `AddTemplate`, `ListTemplates`, `FindTemplate`, `DeleteTemplate`. No `UpdateTemplate` or `RenameTemplate` method exists.
2. **Template struct** (store/todo.go) has 4 fields: `ID int`, `Name string`, `Content string`, `CreatedAt string`. No schedule-related fields.
3. **Templates table** (SQLite, version 2 migration) has schema: `id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, content TEXT NOT NULL, created_at TEXT NOT NULL`.
4. **PRAGMA user_version** is currently at 3. Next migration is version 4.
5. **Overlay pattern** is well-established: `showX bool` in app.Model, dedicated `updateX(msg)` routing method, `SetSize(w,h)`, `SetTheme(t)`, `HelpBindings()`, custom Msg types for close/actions.
6. **Template workflow** in todolist.Model is comprehensive: `templateSelectMode`, `placeholderInputMode`, `templateNameMode`, `templateContentMode` -- these handle selection, placeholder prompting, and inline creation.
7. **app.Init()** currently returns `nil` -- no startup commands run.
8. **Todo.Body** stores template-rendered markdown. The link between a todo and the template it was created from is NOT tracked -- once created, the todo is independent.

---

## Integration Architecture: Three New Capabilities

### Capability 1: Template Management Overlay

**What:** Full-screen overlay for listing, viewing, editing, renaming, and deleting templates. Currently, template management is buried inside todolist's inline modes (templateSelectMode, templateNameMode, templateContentMode). A dedicated overlay provides a proper management experience.

### Capability 2: Recurring Schedule Definitions

**What:** Attach scheduling rules to templates. A schedule defines when a template should automatically create todos (e.g., "every weekday", "every Monday and Friday", "monthly on the 15th").

### Capability 3: Auto-Creation Engine

**What:** On app launch, examine all schedules and create any missing todos for a rolling window (today + 7 days). Includes placeholder prompting for templates that have `{{.Variable}}` fields.

---

## Component Analysis: New vs Modified

### New Package: `internal/tmplmgr/` (Template Manager Overlay)

**Why a new package, not extending `internal/tmpl/`:** The existing `tmpl` package is a pure utility package (ExtractPlaceholders, ExecuteTemplate) with no TUI concerns. A management overlay is a full Bubble Tea component with Model/Update/View, keys, styles -- the same structure as `settings`, `search`, and `preview`. Mixing TUI model logic into a utility package violates the project's clean separation.

**Why not extending `todolist/`:** The template management modes already in todolist (templateSelectMode, templateNameMode, etc.) are tightly coupled to the "create todo from template" workflow. A management overlay is a distinct user flow: browse all templates, view their content, edit content, rename, delete, manage schedules. Cramming this into todolist would bloat an already-large model (1104 lines) and mix management concerns with todo CRUD concerns.

**Structure follows the established overlay pattern:**

```go
package tmplmgr

type Model struct {
    templates    []store.Template
    cursor       int
    viewMode     viewMode  // list, view, edit, rename
    store        store.TodoStore
    width, height int
    keys         KeyMap
    styles       Styles
    input        textinput.Model
    textarea     textarea.Model
}

// Messages emitted to app.Model
type CloseMsg struct{}
type TemplateUpdatedMsg struct{}  // signal to refresh if todolist caches templates

func New(s store.TodoStore, t theme.Theme) Model { ... }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { ... }
func (m Model) View() string { ... }
func (m *Model) SetSize(w, h int) { ... }
func (m *Model) SetTheme(t theme.Theme) { ... }
func (m Model) HelpBindings() []key.Binding { ... }
```

**View modes within the overlay:**

| Mode | View | Keys |
|------|------|------|
| `listMode` | Template list with cursor, name + preview | j/k navigate, enter view, e edit, r rename, d delete, n new, esc close |
| `viewMode` | Full template content rendered as text | esc back to list |
| `editMode` | Textarea with template content | Ctrl+D save, esc cancel |
| `renameMode` | Text input for new name | enter confirm, esc cancel |

### Modified: `app/model.go`

Add template management overlay routing, following the exact pattern of settings/search/preview:

```go
type Model struct {
    // ... existing fields
    showTmplMgr  bool
    tmplMgr      tmplmgr.Model
}
```

**Integration points in app.Model:**

1. **Key binding:** New key (recommend `m` for "manage templates") in app.KeyMap, handled when `!isInputting` and no overlay is active.
2. **Message routing:** `tmplmgr.CloseMsg` handler sets `showTmplMgr = false`. Place it alongside `preview.CloseMsg` and `search.CloseMsg`.
3. **Update routing:** `if m.showTmplMgr { return m.updateTmplMgr(msg) }` -- placed after showSettings, showPreview, showSearch checks.
4. **View routing:** `if m.showTmplMgr { ... }` -- same pattern as other overlays.
5. **Theme propagation:** Add `m.tmplMgr.SetTheme(t)` to `applyTheme()`.

### Modified: `store/store.go` (TodoStore Interface)

Add methods needed by the management overlay and scheduling:

```go
// Template management additions
UpdateTemplate(id int, name, content string) error
// Note: name is separate from content to support rename-only operations

// Schedule operations (new)
AddSchedule(templateID int, rule string) (Schedule, error)
ListSchedules() []Schedule
ListSchedulesForTemplate(templateID int) []Schedule
DeleteSchedule(id int)
UpdateSchedule(id int, rule string) error
TodoExistsForSchedule(scheduleID int, date string) bool
```

**Why TodoExistsForSchedule:** The auto-creation engine needs to check whether a todo for a given schedule+date combination already exists before creating it. Without this, restarting the app would create duplicate todos.

### Modified: `store/sqlite.go`

Implement the new interface methods. Add migration to version 4 (and possibly 5).

### Modified: `store/todo.go`

Add new struct types:

```go
// Schedule represents a recurring rule attached to a template.
type Schedule struct {
    ID         int
    TemplateID int
    Rule       string  // serialized schedule rule (see below)
    CreatedAt  string
}
```

### New: Auto-creation logic location

**Where should auto-creation run?**

Three options analyzed:

| Option | Location | Mechanism | Pros | Cons |
|--------|----------|-----------|------|------|
| A | `app.Init()` | Return `tea.Cmd` that runs auto-creation | Runs once at startup, idiomatic Bubble Tea | Blocks first render until complete |
| B | `main.go` before `app.New()` | Direct function call | Simple, synchronous, clear ordering | Not in the Bubble Tea lifecycle |
| C | Background goroutine via `tea.Cmd` | Async command in `app.Init()` | Non-blocking | Complexity, race conditions with store |

**Recommendation: Option B -- run in `main.go` before `app.New()`.**

Rationale:
- Auto-creation is a data preparation step, not a UI interaction. It belongs in the startup sequence alongside config loading and store initialization.
- The store is single-connection (`MaxOpenConns(1)`), so running auto-creation before the TUI starts avoids any concurrency concerns.
- Templates with placeholders require user prompting, which is handled separately (see "Placeholder prompting for recurring todos" below).
- The rolling window calculation is pure logic on a small dataset -- it completes in under 1ms. No need for async execution.

```go
// In main.go, after store initialization:
s, err := store.NewSQLiteStore(dbPath)
if err != nil { ... }
defer s.Close()

// Auto-create scheduled todos for rolling window
pending := recurring.GeneratePending(s, time.Now(), 7)
// pending contains todos that need creation (some may need placeholder prompting)
recurring.CreateSimpleTodos(s, pending.NoPlaceholders)
// pending.NeedPrompting is passed to app.New() for UI prompting
```

Wait -- this creates a problem. Templates with placeholders need TUI interaction for prompting, but we are running before the TUI starts. Two approaches:

**Approach A (recommended): Split auto-creation into two steps.**
1. **Pre-TUI (main.go):** Create todos from templates with NO placeholders. These can be fully auto-created without user interaction.
2. **Post-TUI-init (app.Init or first Update):** For templates WITH placeholders, show a prompting flow. Or simply skip placeholder prompting for auto-created recurring todos (use empty placeholder values).

**Approach B: Skip placeholder prompting entirely for recurring todos.**
Recurring todos created from templates execute with empty placeholder values (which text/template handles via `missingkey=zero`). The user can fill in details via the external editor. This is simpler and arguably better UX -- the recurring todo is created as a reminder, and the user fills in specifics when they get to it.

**Recommendation: Approach B.** Auto-created recurring todos use templates with empty placeholder values. The todo title comes from the template name or a configured title, and the body is the template content with blanks. User can edit via the external editor. This keeps auto-creation fully synchronous and pre-TUI.

Revised startup:

```go
// In main.go, after store initialization:
recurring.AutoCreate(s, time.Now(), 7)  // creates any missing scheduled todos
```

---

## Schema Design

### Migration Version 4: Add schedules table

```sql
CREATE TABLE IF NOT EXISTS schedules (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    rule        TEXT    NOT NULL,
    created_at  TEXT    NOT NULL
);

CREATE INDEX idx_schedules_template ON schedules(template_id);
```

**Key decisions:**

- **Foreign key to templates with CASCADE delete:** When a template is deleted, its schedules are automatically removed. This prevents orphaned schedules. `foreign_keys(ON)` is already set in the SQLite DSN.
- **`rule` is a TEXT column** storing a serialized schedule rule (see "Schedule Rule Format" below). A structured TEXT format is simpler than multiple columns (`frequency`, `day_of_week`, `day_of_month`, etc.) and more extensible.

### Migration Version 5: Add schedule tracking to todos

```sql
ALTER TABLE todos ADD COLUMN schedule_id INTEGER REFERENCES schedules(id) ON DELETE SET NULL;
ALTER TABLE todos ADD COLUMN schedule_date TEXT;

CREATE INDEX idx_todos_schedule ON todos(schedule_id, schedule_date);
```

**Key decisions:**

- **`schedule_id` nullable FK with SET NULL on delete:** When a schedule is deleted, existing todos created by it remain (they are valid todos) but lose the schedule link. This is the right behavior -- deleting a recurring rule should not delete already-created todos.
- **`schedule_date` stores the logical date** the todo was created for (the date from the rolling window), not the creation timestamp. This is the deduplication key: `TodoExistsForSchedule(scheduleID, date)` checks `WHERE schedule_id = ? AND schedule_date = ?`.
- **Why not a junction table:** A todo is created by at most one schedule on a specific date. The 1:1 relationship is cleanly modeled as columns on the todos table. A junction table adds complexity for no benefit.

### Schedule Rule Format

**Use a structured string format rather than cron expressions.**

Cron is overkill and unfamiliar to most users. The supported recurrence patterns are simple enough for a custom format:

```
daily                       -- every day
weekdays                    -- Mon-Fri
weekly:mon,wed,fri          -- specific days of week
monthly:15                  -- 15th of each month
monthly:last                -- last day of each month
```

Serialization/deserialization is trivial:

```go
type ScheduleRule struct {
    Type      string   // "daily", "weekdays", "weekly", "monthly"
    Days      []string // for weekly: ["mon", "wed", "fri"]
    DayOfMonth int     // for monthly: 1-31 or -1 for last
}

func ParseRule(s string) (ScheduleRule, error) { ... }
func (r ScheduleRule) String() string { ... }
func (r ScheduleRule) MatchesDate(d time.Time) bool { ... }
```

**Why not cron:** Cron expressions (`0 9 * * 1-5`) are powerful but opaque. The target user is someone managing personal todos in a terminal, not scheduling infrastructure jobs. "weekdays" and "weekly:mon,fri" are self-documenting. Cron is a v2 consideration if users demand complex schedules.

---

## Data Flow: Recurring Todo Auto-Creation

### Startup Flow

```
main.go
  |
  1. config.Load()
  2. store.NewSQLiteStore(dbPath)
     |-- migrate() runs v4+v5 if needed
  3. recurring.AutoCreate(store, time.Now(), 7)
     |
     |-- store.ListSchedules()
     |   returns all schedules with their template_id and rule
     |
     |-- For each schedule:
     |   |-- ParseRule(schedule.Rule)
     |   |-- For each date in [today, today+7]:
     |       |-- rule.MatchesDate(date)?
     |       |   NO  -> skip
     |       |   YES -> store.TodoExistsForSchedule(schedule.ID, date)?
     |              |   YES -> skip (already created)
     |              |   NO  -> Create todo:
     |                    |-- store.FindTemplate(schedule.TemplateID)
     |                    |-- tmpl.ExecuteTemplate(content, {})  // empty placeholders
     |                    |-- store.AddWithSchedule(title, date, body, scheduleID, scheduleDate)
     |
  4. app.New(provider, mondayStart, store, theme, cfg)
  5. tea.NewProgram(model).Run()
```

### Template Management Flow

```
User presses 'm' (manage templates)
  |
  app.Model:
    m.tmplMgr = tmplmgr.New(m.store, theme.ForName(m.cfg.Theme))
    m.tmplMgr.SetSize(m.width, m.height)
    m.showTmplMgr = true
  |
  tmplmgr.Model (overlay active):
    |-- List view: shows all templates with names
    |-- View mode: shows template content + attached schedules
    |-- Edit mode: textarea for content editing
    |-- Rename mode: text input for name change
    |-- Schedule mode: add/remove recurring rules
    |
    On close:
      emits tmplmgr.CloseMsg
  |
  app.Model:
    m.showTmplMgr = false
```

### Schedule Attachment Flow (within tmplmgr overlay)

```
User on template list view, presses 'r' (recurring/schedule):
  |
  tmplmgr.Model enters scheduleMode:
    |-- Shows existing schedules for this template
    |-- 'a' to add new schedule:
    |     Enters scheduleInputMode
    |     Presents options: daily / weekdays / weekly / monthly
    |     For weekly: prompts for day selection (toggle Mon-Sun)
    |     For monthly: prompts for day number
    |     Confirm creates schedule via store.AddSchedule()
    |-- 'd' to delete selected schedule
    |-- 'esc' back to template detail
```

---

## Integration Points Detailed

### 1. app.Model Overlay Routing

The routing priority for overlays must be defined. Current order in `Update()`:

```
1. Settings-specific messages (ThemeChangedMsg, SaveMsg, CancelMsg) -- always
2. search.JumpMsg, search.CloseMsg -- always
3. preview.CloseMsg -- always
4. todolist.PreviewMsg, todolist.OpenEditorMsg -- always
5. editor.EditorFinishedMsg -- always
6. if showSettings -> updateSettings(msg)
7. if showPreview -> updatePreview(msg)
8. if showSearch -> updateSearch(msg)
9. Normal routing (key handling, pane routing)
```

**Template manager fits at position 6.5:**

```
6. if showSettings -> updateSettings(msg)
7. if showTmplMgr -> updateTmplMgr(msg)    // NEW
8. if showPreview -> updatePreview(msg)
9. if showSearch -> updateSearch(msg)
10. Normal routing
```

**Add tmplmgr.CloseMsg handling alongside other close messages:**

```go
case tmplmgr.CloseMsg:
    m.showTmplMgr = false
    return m, nil
```

### 2. TodoStore Interface Extension

New methods needed (all added to interface, implemented in SQLiteStore, stubbed in JSON Store):

```go
// Template management
UpdateTemplate(id int, name, content string) error

// Schedule CRUD
AddSchedule(templateID int, rule string) (Schedule, error)
ListSchedules() []Schedule
ListSchedulesForTemplate(templateID int) []Schedule
DeleteSchedule(id int)
UpdateSchedule(id int, rule string) error

// Auto-creation support
TodoExistsForSchedule(scheduleID int, date string) bool
AddScheduledTodo(text, date, body string, scheduleID int, scheduleDate string) Todo
```

**Why `AddScheduledTodo` separate from `Add`:** The existing `Add(text, date)` method creates a plain todo. Scheduled todos need additional fields (`schedule_id`, `schedule_date`). Rather than changing the signature of `Add` (which would break all existing callers), a separate method is cleaner.

### 3. Theme Propagation

Add to `app.Model.applyTheme()`:

```go
m.tmplMgr.SetTheme(t)
```

This follows the established pattern used by all other overlay components.

### 4. Help Bar Integration

Add to `app.Model.currentHelpKeys()`:

```go
if m.showTmplMgr {
    return helpKeyMap{bindings: m.tmplMgr.HelpBindings()}
}
```

Place after `showPreview` check, before `showSearch` check.

### 5. Window Resize Propagation

Add `updateTmplMgr()` method to app.Model following the established pattern:

```go
func (m Model) updateTmplMgr(msg tea.Msg) (tea.Model, tea.Cmd) {
    if wsm, ok := msg.(tea.WindowSizeMsg); ok {
        m.width = wsm.Width
        m.height = wsm.Height
        m.ready = true
        m.help.Width = wsm.Width
        m.tmplMgr.SetSize(wsm.Width, wsm.Height)

        var calCmd tea.Cmd
        m.calendar, calCmd = m.calendar.Update(msg)
        m.syncTodoSize()
        return m, calCmd
    }

    var cmd tea.Cmd
    m.tmplMgr, cmd = m.tmplMgr.Update(msg)
    return m, cmd
}
```

---

## New Package: `internal/recurring/`

**Purpose:** Schedule rule parsing, date matching logic, and auto-creation orchestration. This is pure business logic with no TUI concerns.

```
internal/recurring/
    rule.go       -- ScheduleRule type, ParseRule, String, MatchesDate
    generate.go   -- AutoCreate function, rolling window logic
    rule_test.go  -- Unit tests for rule parsing and date matching
    generate_test.go -- Unit tests for auto-creation logic
```

**Why separate from tmplmgr:** The auto-creation logic runs in `main.go` before the TUI starts. It has no TUI dependency. The `tmplmgr` package is a Bubble Tea component. Mixing them would create an unnecessary dependency chain.

**Why separate from store:** The recurring logic uses the store but is not the store. It orchestrates reads (ListSchedules, FindTemplate, TodoExistsForSchedule) and writes (AddScheduledTodo). This is application logic, not persistence logic.

### Rolling Window Algorithm

```go
func AutoCreate(s store.TodoStore, now time.Time, windowDays int) {
    schedules := s.ListSchedules()
    for _, sched := range schedules {
        rule, err := ParseRule(sched.Rule)
        if err != nil {
            continue // skip malformed rules
        }
        tpl := s.FindTemplate(sched.TemplateID)
        if tpl == nil {
            continue // template was deleted but CASCADE didn't fire (shouldn't happen)
        }

        for d := 0; d < windowDays; d++ {
            date := now.AddDate(0, 0, d)
            dateStr := date.Format("2006-01-02")

            if !rule.MatchesDate(date) {
                continue
            }
            if s.TodoExistsForSchedule(sched.ID, dateStr) {
                continue
            }

            // Execute template with empty placeholders
            body, _ := tmpl.ExecuteTemplate(tpl.Content, map[string]string{})
            title := tpl.Name // use template name as todo title
            s.AddScheduledTodo(title, dateStr, body, sched.ID, dateStr)
        }
    }
}
```

**Window size of 7 days:** Creates todos for today through 6 days from now. This means:
- User sees upcoming scheduled todos when they navigate to those dates.
- If the user does not open the app for a week, they get 7 days of catch-up on next launch.
- Missing days beyond 7 are not retroactively created (this is intentional -- stale recurring todos for missed days are noise, not value).

**Edge cases:**
- Template deleted between schedule check and template lookup: `FindTemplate` returns nil, skip.
- Schedule rule malformed: `ParseRule` returns error, skip. Log to stderr if desired.
- App opened multiple times per day: `TodoExistsForSchedule` deduplicates -- no duplicates created.

---

## Relationship Between Templates and Recurring Todos

```
templates (1) --- (0..N) schedules (1) --- (0..N) todos
    |                        |                      |
    | has content +          | has rule string       | has schedule_id
    | placeholders           | + template_id FK      | + schedule_date
    |                        |                      | + body (rendered from template)
```

**Key design principle: Todos are independent after creation.**

Once a recurring todo is created from a template, it is a regular todo. If the template content changes later, existing todos are NOT retroactively updated. This is the correct behavior because:
1. The user may have already edited the todo body via the external editor.
2. Retroactive updates would destroy user customizations.
3. The template is a starting point, not a live binding.

**Tracking lineage:** `schedule_id` on the todo tracks which schedule created it, not which template. This is for deduplication only. The user never sees this linkage in the UI.

---

## Suggested Build Order

The features have clear dependencies. Build order minimizes risk and ensures each step is independently testable.

### Phase A: Template Management Overlay (Foundation)

**What:** Create `tmplmgr` package with full overlay UI. Add `UpdateTemplate` to store interface. Wire into app.Model.

**Why first:** This is a standalone feature with no dependency on recurring/scheduling. It provides value immediately (users can manage templates in a proper UI instead of the cramped inline modes). It also establishes the UI infrastructure that the schedule management UI will be built into.

**New/Modified:**
- NEW: `internal/tmplmgr/model.go`, `keys.go`, `styles.go`
- MOD: `store/store.go` (add `UpdateTemplate` to interface)
- MOD: `store/sqlite.go` (implement `UpdateTemplate`)
- MOD: `app/model.go` (add overlay routing, key binding)
- MOD: `app/keys.go` (add template management key)

**Risk:** Low. Follows established overlay pattern exactly.

### Phase B: Schedule Schema + CRUD

**What:** Add schedules table (migration v4), schedule tracking columns on todos (migration v5). Implement schedule CRUD in store. Create `internal/recurring/rule.go` with rule parsing and date matching.

**Why second:** The schema and data layer must exist before either the UI for schedule management or the auto-creation engine can work. This phase is all backend -- no UI changes.

**New/Modified:**
- NEW: `internal/recurring/rule.go`, `rule_test.go`
- MOD: `store/todo.go` (add Schedule struct, update Todo with schedule fields)
- MOD: `store/store.go` (add schedule methods to interface)
- MOD: `store/sqlite.go` (implement schedule methods, migrations v4+v5)

**Risk:** Medium. Schema migrations need careful testing. Rule parsing is straightforward but edge cases exist (e.g., monthly:31 on February).

### Phase C: Auto-Creation + Schedule UI

**What:** Implement `recurring.AutoCreate()` in `main.go` startup. Add schedule management UI to the `tmplmgr` overlay (new modes for viewing/adding/deleting schedules on a template).

**Why third:** Depends on both the overlay (Phase A) and the schema/CRUD (Phase B). This is where everything comes together.

**New/Modified:**
- NEW: `internal/recurring/generate.go`, `generate_test.go`
- MOD: `main.go` (add auto-creation call before app.New)
- MOD: `internal/tmplmgr/model.go` (add schedule management modes)

**Risk:** Medium. The auto-creation algorithm is simple but the schedule UI adds more modes to the overlay. Integration testing (does a scheduled todo actually appear on the calendar?) requires end-to-end verification.

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Storing Schedules as Cron Strings

**What:** Using full cron syntax (`0 9 * * 1-5`) for schedule rules.

**Why bad:** Cron is powerful but opaque for this use case. The supported recurrence types (daily, weekdays, specific weekdays, monthly) are simple enough to warrant a custom format that is human-readable. Cron also has no standard Go library that doesn't pull in significant dependencies.

**Instead:** Use the custom rule format (`daily`, `weekdays`, `weekly:mon,fri`, `monthly:15`). It is self-documenting, trivial to parse, and covers all v1.6 use cases. Complex cadences like "every 2nd Tuesday" are explicitly out of scope (v2 candidate per PROJECT.md).

### Anti-Pattern 2: Live-Binding Todos to Templates

**What:** Making todos maintain a live link to their source template, automatically updating when the template changes.

**Why bad:** Destroys user edits. If a user creates a recurring todo from "Daily Plan" template and then customizes the body via the external editor, a template change would overwrite their work. Also introduces synchronization complexity (when to sync? on view? on edit?).

**Instead:** Todos are independent after creation. The template is a stamp, not a link. `schedule_id` exists for deduplication only.

### Anti-Pattern 3: Retroactive Todo Creation for Missed Days

**What:** When the app is opened after being closed for 2 weeks, creating recurring todos for all 14 missed days.

**Why bad:** Creates a flood of stale todos that are no longer relevant. A "Daily Plan" todo for 10 days ago is noise, not value. The user did not open the app; they presumably had a different workflow those days.

**Instead:** Use a rolling window from today forward only (7 days). Do NOT create todos for past days. If the user needs to catch up, they can manually create todos. The rolling window is a prospective tool, not a retroactive one.

### Anti-Pattern 4: Putting Auto-Creation in app.Init()

**What:** Running auto-creation as a `tea.Cmd` from `app.Init()`.

**Why bad:** Mixes data preparation with TUI lifecycle. Creates timing complexity (what if the user navigates before auto-creation completes?). The store is single-connection, so concurrent access from a background command and the render loop could cause "database is locked" errors.

**Instead:** Run auto-creation synchronously in `main.go` before the TUI starts. It completes in under 1ms for reasonable schedule counts. The TUI sees a fully prepared dataset from frame one.

### Anti-Pattern 5: Adding Schedule Fields to the Template Struct

**What:** Putting `ScheduleRule`, `ScheduleDays`, etc. directly on the Template struct.

**Why bad:** Conflates two concerns (template content and recurrence rules). A template can have zero or multiple schedules (e.g., "Daily Plan" might recur on weekdays AND monthly on the 1st as a "Monthly Plan" variant). Embedding schedule in template limits to 1:1.

**Instead:** Separate `schedules` table with `template_id` FK. Clean 1:N relationship. Each schedule has its own rule, ID, and lifecycle.

### Anti-Pattern 6: Rebuilding Template Select in the Overlay

**What:** Duplicating the template selection UI that already exists in todolist.Model.

**Why bad:** The todolist template workflow (select template -> fill placeholders -> create todo) is a distinct user flow from template management. Combining them in one place creates confusion about intent (am I creating a todo or managing templates?).

**Instead:** Keep both entry points: `t` in todolist for "use template to create todo" and `m` in app for "manage templates and schedules". They serve different purposes.

---

## Component Dependency Map (After v1.6)

```
main.go
  |
  +-- config.Load() -> config.Config
  |
  +-- store.NewSQLiteStore(dbPath) -> store.TodoStore
  |     +-- migrate() runs v4 (schedules table), v5 (todo schedule columns)
  |
  +-- recurring.AutoCreate(store, time.Now(), 7)              [NEW]
  |     +-- store.ListSchedules()
  |     +-- recurring.ParseRule()
  |     +-- store.FindTemplate()
  |     +-- tmpl.ExecuteTemplate()
  |     +-- store.TodoExistsForSchedule()
  |     +-- store.AddScheduledTodo()
  |
  +-- app.New(provider, mondayStart, store, theme, cfg) -> app.Model
        |
        +-- calendar.Model     (unchanged)
        +-- todolist.Model     (unchanged -- existing template workflow stays)
        +-- settings.Model     (unchanged)
        +-- search.Model       (unchanged)
        +-- preview.Model      (unchanged)
        +-- tmplmgr.Model      [NEW -- template management overlay]
        |     +-- store.TodoStore (ListTemplates, UpdateTemplate, DeleteTemplate)
        |     +-- store.TodoStore (ListSchedulesForTemplate, AddSchedule, DeleteSchedule)
        |
        +-- recurring package  [NEW -- used only in main.go, not in app.Model]
```

---

## Sources

- Codebase analysis: `internal/app/model.go` (overlay routing pattern) -- HIGH confidence, direct inspection
- Codebase analysis: `internal/settings/model.go` (overlay component pattern) -- HIGH confidence, direct inspection
- Codebase analysis: `internal/store/sqlite.go` (migration pattern, PRAGMA user_version) -- HIGH confidence, direct inspection
- Codebase analysis: `internal/store/store.go` (TodoStore interface, current method set) -- HIGH confidence, direct inspection
- Codebase analysis: `internal/todolist/model.go` (template workflow modes) -- HIGH confidence, direct inspection
- Codebase analysis: `internal/tmpl/tmpl.go` (ExtractPlaceholders, ExecuteTemplate) -- HIGH confidence, direct inspection
- Codebase analysis: `internal/store/todo.go` (Template struct, Todo struct) -- HIGH confidence, direct inspection
- Codebase analysis: `main.go` (startup sequence, Init returns nil) -- HIGH confidence, direct inspection
- PROJECT.md: Active requirements and v2 candidates -- HIGH confidence, project specification
- SQLite foreign key documentation (sqlite.org/foreignkeys.html) -- HIGH confidence for CASCADE behavior
