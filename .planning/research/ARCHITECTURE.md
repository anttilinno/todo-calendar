# Architecture Research: v1.4 Data & Editing

**Domain:** SQLite backend, markdown todo bodies, external editor integration
**Researched:** 2026-02-06
**Confidence:** HIGH for SQLite + editor integration (verified against Bubble Tea API docs and Go SQLite ecosystem); MEDIUM for markdown template design (pattern-based, not library-verified)

## Current Architecture Summary

The app follows Bubble Tea's Elm Architecture with clean component boundaries:

```
main.go
  |
  app.Model (root orchestrator)
  |-- calendar.Model   (left pane: grid + overview)
  |-- todolist.Model   (right pane: todo list + input)
  |-- settings.Model   (full-screen overlay when active)
  |-- search.Model     (full-screen overlay when active)
  |
  store.Store          (pure data layer, JSON persistence)
  config.Config        (TOML config, Save/Load)
  theme.Theme          (14 semantic color roles)
  holidays.Provider    (rickar/cal wrapper)
```

**Critical architectural facts for v1.4 planning:**

1. **store.Store is a concrete struct**, not an interface. All components (calendar, todolist, search) hold `*store.Store` directly. Every mutation calls `s.Save()` which writes the entire JSON file atomically.
2. **store.Todo struct** has 6 fields: `ID int`, `Text string`, `Date string`, `Done bool`, `CreatedAt string`, `SortOrder int`. No Body/Description field exists.
3. **store.Data envelope** holds `NextID int` and `Todos []Todo` -- the entire dataset lives in memory.
4. **All store queries are in-memory loops** over `s.data.Todos` (TodosForMonth, FloatingTodos, SearchTodos, etc.).
5. **app.Model uses `tea.WithAltScreen()`** in main.go when creating the Program.
6. **todolist.Model owns text input** via a single `textinput.Model` field, reused across modes (add, edit, filter, date).

---

## SQLite Integration

### Why SQLite

The current JSON store loads the entire dataset into memory and rewrites the full file on every mutation. This works for <1000 todos but does not scale, does not support full-text search efficiently, and makes schema evolution painful. SQLite provides:
- ACID transactions with WAL mode for crash safety
- Indexed queries (by date, completion status, full-text search via FTS5)
- Schema migrations for evolving the data model (adding Body field, tags, etc.)
- The same single-file, no-server deployment model as JSON

### Library Choice: modernc.org/sqlite

**Recommendation: Use `modernc.org/sqlite` (pure Go, CGo-free).**

Rationale:
- **Cross-compilation:** No C compiler needed. The project currently has zero CGo dependencies (go.mod shows pure Go deps). Introducing CGo via `mattn/go-sqlite3` would complicate builds for all platforms.
- **Performance is sufficient:** For a single-user desktop TUI with at most a few thousand todos, the 10-50% overhead of the pure Go transpilation vs native C SQLite is irrelevant. Benchmarks show sub-millisecond query times for datasets of this size.
- **database/sql compatible:** Registers as driver name `"sqlite"`. Standard `sql.Open("sqlite", path)` works. All Go database tooling (migrations, testing) works unchanged.
- **Actively maintained:** Published on pkg.go.dev, used by production projects (River queue, Watermill, etc.).

Import: `_ "modernc.org/sqlite"` for driver registration, then use `database/sql` throughout.

### Components Modified

| Component | Change | Reason |
|-----------|--------|--------|
| `store/store.go` | Replace JSON read/write with `*sql.DB` operations | Core persistence swap |
| `store/todo.go` | Add `Body string` field to Todo struct | Support markdown bodies |
| `main.go` | Open SQLite DB, pass to store, handle migration | Initialization change |
| `config/paths.go` | Add `DBPath()` function | New file path for `todos.db` |

### Data Model Changes

**Current JSON schema:**
```json
{
  "next_id": 42,
  "todos": [
    {"id": 1, "text": "...", "date": "2026-01-15", "done": false, "created_at": "...", "sort_order": 10}
  ]
}
```

**Target SQLite schema (v1):**
```sql
CREATE TABLE IF NOT EXISTS todos (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    text       TEXT    NOT NULL,
    body       TEXT    NOT NULL DEFAULT '',
    date       TEXT,        -- YYYY-MM-DD or NULL for floating
    done       INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_todos_date ON todos(date);
CREATE INDEX idx_todos_done ON todos(done);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL
);
```

**Key design decisions:**
- `AUTOINCREMENT` replaces the manual `NextID` counter. SQLite handles ID generation.
- `body` is a plain TEXT column (not stored as a separate file). Markdown is just text; storing it in-DB keeps the single-file deployment model and enables FTS5 search across bodies later.
- `date` is nullable TEXT rather than empty string. SQL NULL semantics are cleaner for "no date" than empty string, and enable proper `WHERE date IS NULL` queries for floating todos.
- `done` is INTEGER (0/1) since SQLite has no native boolean.
- No separate migration library is needed initially. A simple version-check table (`schema_version`) with hand-written `ALTER TABLE` statements is sufficient for a single-user desktop app. Introduce goose/golang-migrate only if migration complexity grows beyond 3-4 schema versions.

### Migration Strategy: JSON to SQLite

**One-time migration at startup, keeping JSON as backup:**

```
1. Check if todos.db exists
2. If not, check if todos.json exists
3. If JSON exists:
   a. Create todos.db with schema v1
   b. Read JSON, insert all todos into SQLite (in a transaction)
   c. Rename todos.json -> todos.json.backup
4. If neither exists, create fresh todos.db with schema v1
5. If todos.db exists, check schema_version and run any pending migrations
```

**Rationale for rename-not-delete:** Users who encounter issues can restore the backup. The migration is one-way; the app never writes JSON again after migration.

**Transaction safety:** The entire JSON import must happen in a single transaction. If it fails partway, the transaction rolls back and the JSON file is untouched.

### Store Refactoring Strategy

**Phase 1: Extract interface, keep JSON implementation.**

Define a `store.TodoStore` interface matching the current method set:

```go
type TodoStore interface {
    Add(text string, date string) Todo
    Toggle(id int)
    Delete(id int)
    Find(id int) *Todo
    Update(id int, text string, date string)
    Todos() []Todo
    TodosForMonth(year int, month time.Month) []Todo
    FloatingTodos() []Todo
    IncompleteTodosPerDay(year int, month time.Month) map[int]int
    TodoCountsByMonth() []MonthCount
    FloatingTodoCounts() FloatingCount
    SwapOrder(id1, id2 int)
    SearchTodos(query string) []Todo
    EnsureSortOrder()
    Save() error
}
```

All consumers (`calendar.Model`, `todolist.Model`, `search.Model`, `app.Model`) change from `*store.Store` to `store.TodoStore`. This is a mechanical refactor -- change the field type, verify it compiles. No behavioral change.

**Phase 2: Implement SQLite backend.**

Create `store.SQLiteStore` implementing `TodoStore`. The JSON `Store` remains as a fallback/reference. The constructor in `main.go` decides which to instantiate.

**Why interface-first:** Extracting the interface before writing SQLite code means every consumer is already decoupled. The SQLite implementation can be developed and tested independently. It also enables a `store.MemoryStore` for testing.

### SQLite Pragmas for Desktop TUI

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
```

- **WAL mode** prevents readers from blocking writers. Even though this is single-user, the TUI might query while a write transaction is in-flight (e.g., RefreshIndicators called during Add).
- **synchronous = NORMAL** provides good crash safety without the performance cost of FULL. For a todo app, losing the last few milliseconds of data on power loss is acceptable.
- **busy_timeout** prevents "database is locked" errors if operations overlap.

### Save() Semantics Change

Currently every mutation (Add, Toggle, Delete, etc.) calls `s.Save()` which rewrites the entire JSON file. With SQLite, each mutation is its own transaction -- there is no separate Save() step.

**Options:**
1. **Remove Save() from the interface.** Each mutation method commits its own transaction. Simplest, correct for SQLite. But callers that currently call Save() explicitly (like EnsureSortOrder) need adjustment.
2. **Keep Save() as a no-op on SQLiteStore.** Maintains backward compatibility during transition. JSON store keeps its behavior. SQLiteStore treats Save() as a no-op since mutations auto-commit.

**Recommendation:** Option 2 during transition, then remove Save() from the interface once JSON backend is dropped. This avoids a disruptive API change during the migration phase.

---

## Markdown Todo Bodies

### Concept

Each todo gets an optional `body` field containing markdown text. The one-line `text` field remains as the title/summary displayed in the list. The `body` is viewed/edited through the external editor and optionally previewed in a detail pane.

### Data Flow

```
User presses 'e' on a todo (or new keybinding like 'b' for body)
  |
  v
todolist.Model emits OpenEditorMsg{TodoID: id}
  |
  v
app.Model handles OpenEditorMsg:
  1. Reads todo from store (gets current body)
  2. Writes body to a temp file with markdown template
  3. Returns tea.ExecProcess(exec.Command(editor, tempfile), callback)
  |
  v
Bubble Tea suspends TUI, launches editor
  |
  v
User edits markdown in external editor, saves, exits
  |
  v
Callback fires -> EditorFinishedMsg{TodoID, TempPath, Err}
  |
  v
app.Model handles EditorFinishedMsg:
  1. Reads temp file content
  2. Parses: extract title line + body (or just body if title unchanged)
  3. Calls store.UpdateBody(id, body)
  4. Removes temp file
  5. TUI resumes rendering
```

### Markdown Template Format

When opening a todo in the editor, write a temp file with this structure:

```markdown
# Buy groceries

- [ ] Milk
- [ ] Bread
- [ ] Eggs

Notes: Check the sale at Lidl this week.
```

The first `# heading` line is the todo title (`Text` field). Everything after is the body. This gives the user a natural editing experience -- they see and can edit the title in context.

**Parsing on save:**
1. Read the file content.
2. Find the first line matching `^# (.+)$`. That becomes the new `Text`.
3. Everything after the first heading line (trimmed) becomes the `Body`.
4. If no heading is found, keep the original `Text` and treat the entire content as `Body`.

**Why not YAML frontmatter:** Frontmatter (`---` delimiters) is a developer-oriented pattern. For a todo app used by non-technical users, a plain markdown file with a heading is more intuitive. The heading IS the title -- no parsing ceremony needed.

### Components Modified

| Component | Change | Reason |
|-----------|--------|--------|
| `store/todo.go` | Add `Body string` field | Store markdown body |
| `store/store.go` or `store/sqlite.go` | Add `UpdateBody(id int, body string)` | Persist body changes |
| `todolist/model.go` | Add keybinding to open body editor | Trigger editor flow |
| `todolist/model.go` | Show body indicator (e.g., `[+]` icon) | Visual cue that body exists |
| `app/model.go` | Handle `OpenEditorMsg`, `EditorFinishedMsg` | Orchestrate editor lifecycle |

### New Components

| Component | Purpose |
|-----------|---------|
| `internal/editor/editor.go` | `OpenEditor(todoID int, title string, body string) tea.Cmd` -- writes temp file, constructs exec.Command, returns `tea.ExecProcess` |
| `internal/editor/parse.go` | `ParseMarkdown(content string) (title, body string)` -- extracts title from `# heading` and body from remainder |
| `internal/editor/template.go` | `RenderTemplate(title, body string) string` -- builds the markdown file content |

**Why a separate `editor` package:** The editor logic (temp file management, markdown parsing, $EDITOR resolution) is distinct from both the store and the TUI models. Keeping it in its own package prevents the todolist or app package from growing file-management concerns.

### Body Display in TUI

For v1.4, the body is primarily edited externally. In-TUI display options (in order of complexity):

1. **Indicator only (simplest, recommended for v1.4):** Show `[+]` or a note icon next to todos that have a non-empty body. No inline rendering.
2. **Preview pane:** When cursor is on a todo with a body, show a rendered preview below or beside the list. Use `charmbracelet/glamour` for markdown-to-ANSI rendering. This is a v1.5+ feature.
3. **Inline expansion:** Toggle body visibility inline in the list. Most complex, layout-heavy. Defer to v2+.

**Recommendation:** Start with indicator-only in v1.4. The external editor is the primary body interface. Glamour-based preview is a natural v1.5 differentiator.

---

## External Editor Integration

### Bubble Tea ExecProcess API

The integration uses `tea.ExecProcess` which is stable in Bubble Tea v1.x (currently v1.3.10 per go.mod).

```go
func tea.ExecProcess(c *exec.Cmd, fn ExecCallback) Cmd
type ExecCallback = func(error) Msg
```

`ExecProcess` suspends the TUI Program, yields the terminal to the external process, and resumes when the process exits. The callback converts the exit error into a Bubble Tea message.

### Known Issue: Alt Screen Output Leak

**This app uses `tea.WithAltScreen()`.** When `tea.ExecProcess` is called, Bubble Tea exits the alt screen, which triggers a final `View()` render that briefly appears on the normal screen before the editor opens.

**Workaround (verified from Bubble Tea issue #424 and #431):**

Add an `editing` state to `app.Model`. When `editing == true`, `View()` returns an empty string. This prevents the flash:

```go
// In app.Model:
type Model struct {
    // ... existing fields
    editing bool  // true while external editor is running
}

// In View():
func (m Model) View() string {
    if m.editing {
        return ""  // suppress render during editor handoff
    }
    // ... normal rendering
}

// In Update(), handling the editor trigger:
case OpenEditorMsg:
    m.editing = true
    return m, editor.Open(msg.TodoID, msg.Title, msg.Body)

// In Update(), handling editor completion:
case EditorFinishedMsg:
    m.editing = false
    // ... process result
```

### $EDITOR Resolution

```go
func resolveEditor() string {
    if e := os.Getenv("EDITOR"); e != "" {
        return e
    }
    if e := os.Getenv("VISUAL"); e != "" {
        return e
    }
    return "vi"  // POSIX fallback
}
```

Check `$EDITOR` first, then `$VISUAL`, then fall back to `vi` (not `vim` -- `vi` is POSIX-mandated and more universally available).

### Editor Integration Architecture

```
todolist.Model                 app.Model                     editor package
     |                              |                              |
     |--OpenEditorMsg{id}---------->|                              |
     |                              |--editor.Open(id,title,body)->|
     |                              |   1. Write temp file         |
     |                              |   2. Resolve $EDITOR         |
     |                              |   3. exec.Command(editor,f)  |
     |                              |<-tea.ExecProcess(cmd,cb)-----|
     |                              |                              |
     |                     [TUI suspended, editor runs]            |
     |                              |                              |
     |                              |<-EditorFinishedMsg{id,path}--|
     |                              |   1. Read temp file          |
     |                              |   2. Parse title + body      |
     |                              |   3. store.Update + UpdateBody|
     |                              |   4. Remove temp file        |
     |                              |   5. m.editing = false       |
```

### Message Types

```go
// In todolist or a shared messages package:
type OpenEditorMsg struct {
    TodoID int
}

// In editor package:
type EditorFinishedMsg struct {
    TodoID   int
    TempPath string
    Err      error
}
```

**Message routing:** `OpenEditorMsg` is emitted by `todolist.Model` via a `tea.Cmd` (not directly -- it returns a command that produces the message). `app.Model` intercepts it in Update(), similar to how it handles `settings.SaveMsg` and `search.JumpMsg` today.

### Temp File Management

```go
// Write temp file in the OS temp directory with a descriptive name
func writeTempFile(title, body string) (string, error) {
    f, err := os.CreateTemp("", "todo-calendar-*.md")
    if err != nil {
        return "", err
    }
    defer f.Close()

    content := RenderTemplate(title, body)
    if _, err := f.WriteString(content); err != nil {
        os.Remove(f.Name())
        return "", err
    }
    return f.Name(), nil
}
```

- Use `.md` extension so editors apply markdown syntax highlighting.
- Use `os.CreateTemp` for safe temp file creation.
- Clean up in the EditorFinishedMsg handler (not in a defer -- the file must persist while the editor has it open).

### Components Modified

| Component | Change | Reason |
|-----------|--------|--------|
| `app/model.go` | Add `editing bool` field; handle `OpenEditorMsg` + `EditorFinishedMsg` | Orchestrate editor lifecycle |
| `app/model.go` | `View()` returns empty string when `editing == true` | Alt screen output leak workaround |
| `todolist/model.go` | Add keybinding (e.g., `b` for body) that emits `OpenEditorMsg` | User trigger |
| `todolist/keys.go` | Add `EditBody` key binding | Key definition |

### New Components

| Component | Purpose |
|-----------|---------|
| `internal/editor/editor.go` | `Open(todoID int, title, body string) tea.Cmd` -- orchestrates temp file + exec |
| `internal/editor/parse.go` | `ParseMarkdown(content string) (title, body string)` |
| `internal/editor/template.go` | `RenderTemplate(title, body string) string` |

---

## Component Dependency Map (After v1.4)

```
main.go
  |
  +-- config.Load() -> config.Config
  |     +-- config.DBPath() [NEW]
  |
  +-- store.NewSQLiteStore(dbPath) -> store.TodoStore [INTERFACE, NEW]
  |     +-- JSON migration if needed
  |     +-- Schema migration
  |
  +-- app.New(provider, mondayStart, store, theme, cfg) -> app.Model
        |
        +-- calendar.New(provider, mondayStart, store, theme)
        |     store field type: store.TodoStore [CHANGED from *store.Store]
        |
        +-- todolist.New(store, theme)
        |     store field type: store.TodoStore [CHANGED from *store.Store]
        |     New keybinding: 'b' -> OpenEditorMsg
        |
        +-- search.New(store, theme, cfg)
        |     store field type: store.TodoStore [CHANGED from *store.Store]
        |
        +-- settings.Model (unchanged)
        |
        +-- editor.Open(id, title, body) [NEW]
              Returns tea.ExecProcess -> EditorFinishedMsg
```

---

## Suggested Build Order

The three features have distinct dependency relationships. Build order should minimize risk and maximize testability at each step.

### Step 1: Extract Store Interface (no behavioral change)

**What:** Define `store.TodoStore` interface. Change all consumers from `*store.Store` to `store.TodoStore`. Verify compilation. Run existing tests.

**Rationale:** This is a pure mechanical refactor. Zero behavioral change. But it is a prerequisite for Steps 2 and 3. Do it first so the rest of v1.4 builds cleanly on the abstraction.

**Risk:** Low. Interface extraction in Go is straightforward. The existing Store struct already satisfies the interface implicitly.

**Files changed:** `store/store.go` (add interface), `app/model.go`, `calendar/model.go`, `todolist/model.go`, `search/model.go` (change field types).

### Step 2: SQLite Backend + Migration

**What:** Implement `store.SQLiteStore` behind the `TodoStore` interface. Include JSON-to-SQLite migration. Wire into `main.go`.

**Rationale:** This is the highest-risk item (new dependency, data migration, persistence change). Do it early so issues surface before building features on top.

**Dependencies:** Step 1 (interface must exist).

**Substeps:**
1. Add `modernc.org/sqlite` dependency
2. Create `store/sqlite.go` with `NewSQLiteStore` constructor
3. Implement schema creation (todos table, schema_version table)
4. Implement all `TodoStore` methods via SQL queries
5. Create `store/migrate.go` with JSON-to-SQLite migration logic
6. Add `config.DBPath()` to `config/paths.go`
7. Update `main.go` to use SQLite store
8. Test: create fresh DB, migrate from JSON, verify all operations

**Risk:** Medium. The SQL queries are straightforward but the migration needs careful testing. Key concern: ensuring `AUTOINCREMENT` IDs do not conflict with migrated IDs (use `INSERT` with explicit IDs during migration, then let autoincrement take over).

### Step 3: Add Body Field to Todo + Store

**What:** Add `Body string` to `store.Todo`. Add `UpdateBody(id int, body string)` to the interface and both implementations. Update SQLite schema (body column already in initial schema from Step 2; if doing incrementally, add `ALTER TABLE todos ADD COLUMN body TEXT NOT NULL DEFAULT ''`).

**Rationale:** The body field is needed before the editor can be useful. But it is a small change.

**Dependencies:** Step 2 (SQLite backend should be working).

**Files changed:** `store/todo.go`, `store/sqlite.go`, interface definition.

### Step 4: External Editor Integration

**What:** Create `internal/editor/` package. Add `OpenEditorMsg` handling to `app.Model`. Add body-edit keybinding to `todolist.Model`. Implement the `editing` state and alt-screen workaround.

**Rationale:** This is the user-facing feature that ties everything together. It depends on the body field (Step 3) and the store interface (Step 1). Build it last so the underlying data infrastructure is stable.

**Dependencies:** Step 1 (interface), Step 3 (body field).

**Substeps:**
1. Create `editor/template.go` -- template rendering
2. Create `editor/parse.go` -- markdown parsing (title extraction)
3. Create `editor/editor.go` -- Open() function with ExecProcess
4. Add `editing bool` to `app.Model`, update `View()` for alt-screen workaround
5. Add `EditorFinishedMsg` handling to `app.Model.Update()`
6. Add keybinding to `todolist.Model` for opening body editor
7. Add body indicator (`[+]`) to `todolist.Model.renderTodo()`
8. Test: manual testing required (editor integration cannot be unit tested easily)

**Risk:** Medium. The ExecProcess API is well-documented and the pattern is established. The alt-screen workaround is the main gotcha, but it is well-understood. Parsing markdown is simple for the heading-extraction use case.

### Step 5 (Optional, v1.4 stretch): Search Body Text

**What:** Extend `SearchTodos` to also search the body field. In SQL: `WHERE text LIKE ? OR body LIKE ?`. For FTS5, create a virtual table.

**Rationale:** Once bodies exist, users will expect search to find content in them. But basic `LIKE` search works for v1.4; FTS5 is a v1.5 optimization.

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Storing Bodies as Separate Files

**What:** Storing each todo's body as a separate `.md` file in a directory.

**Why bad:** Breaks the single-file deployment model. Introduces file-system sync issues (what if the DB references a body file that was deleted?). Complicates backup/migration. The database IS the single source of truth.

**Instead:** Store body as a TEXT column in SQLite. Markdown is just text. Even 10,000 todos with 1KB bodies each is only 10MB -- trivial for SQLite.

### Anti-Pattern 2: Using an ORM

**What:** Introducing GORM or ent for the SQLite backend.

**Why bad:** Massive dependency for 1 table with 7 columns. ORMs add complexity, hide SQL, and make debugging harder. The queries for this app are simple SELECTs, INSERTs, and UPDATEs.

**Instead:** Use `database/sql` directly with hand-written SQL. It is clear, debuggable, and has zero dependencies beyond the driver.

### Anti-Pattern 3: Running Editor in a Goroutine

**What:** Launching the editor in a background goroutine while the TUI continues running.

**Why bad:** The editor and the TUI both need stdin/stdout. They cannot share the terminal. This will cause garbled output and input corruption.

**Instead:** Use `tea.ExecProcess` which properly suspends the TUI, yields the terminal, and resumes when the editor exits.

### Anti-Pattern 4: Complex Migration Framework for 1 Table

**What:** Pulling in pressly/goose or golang-migrate for schema migrations.

**Why bad:** Overkill for a single-table desktop app. These tools are designed for multi-service, multi-developer, multi-environment setups. They add dependencies and complexity.

**Instead:** Use a `schema_version` table and hand-written migration functions. When the app opens the DB, it checks the version and applies any needed migrations sequentially. This is simple, transparent, and sufficient for a single-user app that will have at most 5-10 schema versions over its lifetime.

### Anti-Pattern 5: Breaking the Store Interface with SQLite-Specific Methods

**What:** Adding methods like `Query()` or `Exec()` to the store interface that expose SQLite internals.

**Why bad:** Leaks implementation details. Makes it impossible to swap backends (e.g., for testing with an in-memory store).

**Instead:** Keep the interface in terms of domain operations (Add, Toggle, Search, etc.). If a consumer needs a new capability, add a domain-level method to the interface, not a raw SQL escape hatch.

---

## Scalability Considerations

| Concern | Current (JSON) | After v1.4 (SQLite) |
|---------|----------------|----------------------|
| Load time (1000 todos) | ~5ms (parse JSON) | ~1ms (DB open, no full load) |
| Single mutation | ~10ms (rewrite full file) | <1ms (single row INSERT/UPDATE) |
| Search (1000 todos) | ~1ms (in-memory loop) | ~1ms (LIKE query; <0.5ms with FTS5) |
| File size (1000 todos) | ~200KB JSON | ~150KB SQLite |
| Concurrent safety | None (single writer assumed) | WAL mode handles read/write overlap |
| Schema evolution | Manual JSON migration | ALTER TABLE + version tracking |
| Crash recovery | Atomic rename (good) | WAL + journal (better) |

---

## Sources

- [Bubble Tea ExecProcess API](https://pkg.go.dev/github.com/charmbracelet/bubbletea) -- HIGH confidence, official Go package docs
- [Bubble Tea exec example](https://github.com/charmbracelet/bubbletea/blob/main/examples/exec/main.go) -- HIGH confidence, official example
- [Bubble Tea alt-screen + ExecProcess issue #424](https://github.com/charmbracelet/bubbletea/discussions/424) -- HIGH confidence, maintainer-confirmed workaround
- [Bubble Tea ExecProcess output leak issue #431](https://github.com/charmbracelet/bubbletea/issues/431) -- HIGH confidence, confirmed bug
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) -- HIGH confidence, official package docs
- [Go SQLite benchmarks](https://github.com/cvilsmeier/go-sqlite-bench) -- MEDIUM confidence, community benchmarks
- [SQLite WAL mode](https://sqlite.org/wal.html) -- HIGH confidence, official SQLite docs
- [Go Repository Pattern](https://threedots.tech/post/repository-pattern-in-go/) -- MEDIUM confidence, well-regarded blog
- [charmbracelet/glamour](https://github.com/charmbracelet/glamour) -- HIGH confidence, official repo (for future body preview)
