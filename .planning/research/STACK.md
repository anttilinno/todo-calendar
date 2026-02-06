# Stack Research: v1.4 Data & Editing

**Domain:** SQLite backend, markdown todo bodies with templates, external editor integration
**Researched:** 2026-02-06
**Confidence:** HIGH

## Executive Summary

v1.4 requires three new capabilities not present in the existing stack: (1) SQLite database access, (2) markdown rendering in the terminal, and (3) external process execution for $EDITOR. The external editor capability already exists in Bubble Tea v1.3.10 via `tea.ExecProcess` -- no new dependency needed. SQLite and markdown rendering each require one new library. The template system uses Go's standard library `text/template` with zero new dependencies.

Total new direct dependencies: **2** (SQLite driver + markdown renderer).

---

## Recommended Stack Changes

### 1. SQLite Driver: `modernc.org/sqlite` v1.44.3

| Attribute | Value |
|-----------|-------|
| **Import** | `modernc.org/sqlite` |
| **Version** | v1.44.3 (released Jan 20, 2026) |
| **Bundled SQLite** | 3.51.2 |
| **License** | BSD-3-Clause |
| **CGO required** | No (pure Go) |

**Why this library:**
- **Pure Go, no CGO.** The project currently has zero CGO dependencies. Keeping it CGO-free means `go build` just works on any platform without a C compiler. This is the single most important criterion for a personal-use local TUI app.
- **Standard `database/sql` interface.** Register the driver and use `sql.Open("sqlite", path)`. All Go developers know this interface. No proprietary API to learn.
- **Actively maintained.** Four releases in January 2026 alone. Tracks upstream SQLite closely (currently at 3.51.2).
- **Battle-tested.** Used by Gogs, River queue, and many production Go projects. The modernc.org ecosystem (libc, cc, etc.) is mature.
- **Good enough performance.** Benchmarks show ncruces/go-sqlite3 is faster for bulk inserts, but for a personal todo app doing single-row CRUD, the difference is immeasurable. Both complete simple operations in microseconds.

**Driver registration pattern:**
```go
import (
    "database/sql"
    _ "modernc.org/sqlite"
)

db, err := sql.Open("sqlite", filepath.Join(configDir, "todos.db"))
```

**Alternatives considered and rejected:**

| Driver | Why Not |
|--------|---------|
| `mattn/go-sqlite3` | Requires CGO. Would break the zero-CGO build. Cross-compilation becomes painful. Not worth the marginal performance gain for a todo app. |
| `ncruces/go-sqlite3` v0.30.5 | Pure Go via WASM+wazero. Slightly faster in benchmarks. However: (a) still pre-v1 (v0.30.x), (b) adds wazero as a transitive dependency which is a large WASM runtime, (c) modernc is more widely adopted and has been stable longer. For a personal todo app, modernc's stability and simplicity wins. |
| `crawshaw.dev/sqlite` | Abandoned. No updates since 2023. |

### 2. Markdown Rendering: `github.com/charmbracelet/glamour` v0.10.0

| Attribute | Value |
|-----------|-------|
| **Import** | `github.com/charmbracelet/glamour` |
| **Version** | v0.10.0 (released Apr 16, 2025) |
| **License** | MIT |
| **Key deps** | goldmark (parser), chroma (syntax highlighting), lipgloss (styling) |

**Why this library:**
- **Same ecosystem.** Already using Bubble Tea and Lipgloss from Charmbracelet. Glamour shares transitive dependencies (lipgloss, termenv, x/ansi) which are already in go.sum. The marginal new dependency weight is lower than it appears.
- **Purpose-built for terminal markdown.** Converts markdown to styled ANSI output with configurable width, word wrapping, and theme support. Exactly what's needed to render todo body previews in the TUI.
- **Simple API.** `glamour.Render(markdown, "dark")` returns styled string. For more control: `glamour.NewTermRenderer(glamour.WithWordWrap(width))`.
- **Used in production.** GitHub CLI (`gh`) uses glamour for rendering PR/issue bodies. Charmbracelet's own `glow` tool is built on it.

**Usage pattern for todo body preview:**
```go
import "github.com/charmbracelet/glamour"

rendered, err := glamour.NewTermRenderer(
    glamour.WithWordWrap(panelWidth),
    glamour.WithStyles(glamour.DarkStyleConfig), // or match app theme
)
out, _ := rendered.Render(todo.Body)
```

**Note on version:** Glamour has not reached v1.0 (current is v0.10.0). A v2 module path exists but has no stable tagged release. Use v0.10.0 -- it is the latest tagged release and is what production tools depend on.

**Alternatives considered and rejected:**

| Library | Why Not |
|---------|---------|
| `github.com/MichaelMure/go-term-markdown` | Less maintained, smaller community, doesn't integrate with Charmbracelet styling ecosystem. |
| Raw `goldmark` + manual ANSI | Glamour already wraps goldmark with proper terminal rendering. Reimplementing terminal-aware rendering is significant effort for no benefit. |
| No markdown rendering (plain text only) | The whole point of Phase 15 is rich markdown bodies. Without rendering, users would see raw markdown syntax in the TUI preview, defeating the purpose. |

### 3. Template System: Go stdlib `text/template` (no new dependency)

| Attribute | Value |
|-----------|-------|
| **Import** | `text/template` |
| **Version** | Go stdlib (Go 1.25.6) |
| **New dependency** | None |

**Why stdlib:**
- **Zero new dependencies.** The template needs are simple: replace `{{.Date}}`, `{{.Title}}`, `{{.DayOfWeek}}` placeholders in markdown text. Go's `text/template` does exactly this.
- **Well understood.** Every Go developer knows `text/template`. No learning curve.
- **Powerful enough.** Supports conditionals (`{{if}}`) and range loops if users want them later. But the initial implementation only needs `{{.Variable}}` substitution.
- **No escaping issues.** Using `text/template` (not `html/template`) means markdown syntax passes through unmodified. Curly braces in markdown code blocks work fine since they won't match `{{.FieldName}}` patterns.

**Template storage:** Templates are markdown files stored in `~/.config/todo-calendar/templates/` directory. Loaded via `template.ParseFiles()` or `template.ParseFS()` with `os.DirFS()`.

**Usage pattern:**
```go
import "text/template"

tmpl, err := template.ParseFiles(templatePath)
var buf bytes.Buffer
err = tmpl.Execute(&buf, map[string]string{
    "Date":      "2026-02-06",
    "Title":     "Weekly review",
    "DayOfWeek": "Friday",
})
todoBody := buf.String()
```

**Alternatives considered and rejected:**

| Library | Why Not |
|---------|---------|
| `valyala/fasttemplate` | Faster than text/template for high-throughput scenarios. But we're rendering one template at a time for a personal todo app. The performance difference is literally zero. Adds a dependency for no benefit. |
| Custom `strings.ReplaceAll` | Brittle. No error handling for missing placeholders. No escape mechanism. text/template handles all edge cases. |
| `html/template` | Would HTML-escape markdown syntax characters, breaking the output. Must use `text/template`. |

### 4. External Editor: `tea.ExecProcess` (already in Bubble Tea v1.3.10)

| Attribute | Value |
|-----------|-------|
| **Import** | `github.com/charmbracelet/bubbletea` (already imported) |
| **Function** | `tea.ExecProcess(c *exec.Cmd, fn ExecCallback) Cmd` |
| **New dependency** | None |

**Why no new library:**
- **Built into Bubble Tea.** The `tea.ExecProcess` function exists in v1.3.10 (already the project's version). It suspends the TUI, gives stdin/stdout/stderr to the child process, and resumes the program when the child exits.
- **Proven pattern.** Bubble Tea's own examples include an `exec` example that launches `$EDITOR`. This is the canonical way to do it.
- **Clean lifecycle.** The function returns a `tea.Cmd`. When the editor exits, it delivers a message to `Update()`. The TUI redraws automatically. No manual terminal state management needed.

**Implementation pattern:**
```go
func openInEditor(filePath string) tea.Cmd {
    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "vim"
    }
    c := exec.Command(editor, filePath)
    return tea.ExecProcess(c, func(err error) tea.Msg {
        return editorFinishedMsg{err: err}
    })
}
```

**Workflow:** Write todo body to temp file -> launch editor on temp file -> read temp file back on return -> update todo body -> delete temp file.

### 5. Schema Migration: Manual with `PRAGMA user_version` (no new dependency)

| Attribute | Value |
|-----------|-------|
| **Approach** | Embedded SQL + `PRAGMA user_version` |
| **New dependency** | None |

**Why no migration library:**
- **Single-table schema.** The app has one table (`todos`) with a handful of columns. There will be at most 2-3 schema versions ever. A migration framework is overkill.
- **SQLite's `PRAGMA user_version`** is a built-in integer that persists in the database file. Use it to track schema version. Check on startup, run `ALTER TABLE` statements as needed.
- **Keeps dependency count minimal.** Libraries like goose or golang-migrate are designed for teams with dozens of migrations across multiple environments. This is a personal todo app.

**Implementation pattern:**
```go
func migrate(db *sql.DB) error {
    var version int
    db.QueryRow("PRAGMA user_version").Scan(&version)

    if version < 1 {
        // Initial schema
        db.Exec(`CREATE TABLE IF NOT EXISTS todos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            text TEXT NOT NULL,
            body TEXT NOT NULL DEFAULT '',
            date TEXT NOT NULL DEFAULT '',
            done INTEGER NOT NULL DEFAULT 0,
            created_at TEXT NOT NULL,
            sort_order INTEGER NOT NULL DEFAULT 0
        )`)
        db.Exec("PRAGMA user_version = 1")
    }
    // Future: if version < 2 { ALTER TABLE ... }
    return nil
}
```

**Alternatives considered and rejected:**

| Library | Why Not |
|---------|---------|
| `pressly/goose` | Full migration framework with CLI, versioned files, up/down migrations. Massive overkill for 1-2 tables. Adds large dependency tree. |
| `golang-migrate/migrate` | Same over-engineering problem. Also drags in multiple database driver dependencies. |
| `adlio/schema` | Lighter than goose but still an external dependency for what amounts to 20 lines of migration code. |

---

## Unchanged Core Stack

These technologies continue as-is. Listed for completeness.

| Technology | Version | Role in v1.4 |
|------------|---------|--------------|
| Go | 1.25.6 | Language runtime. `os/exec`, `text/template`, `database/sql`, `io/os` all from stdlib. |
| Bubble Tea | v1.3.10 | TUI framework. `tea.ExecProcess` for editor. Message-passing for async DB operations. |
| Lipgloss | v1.1.0 | Styling. Glamour shares this dependency. |
| Bubbles | v0.21.1 | Input components. Text input for template placeholder prompts. |
| BurntSushi/toml | v1.6.0 | Config file. New fields: `editor` (optional), `default_template`. |
| rickar/cal/v2 | v2.1.27 | Holiday provider. Unchanged. |

---

## Integration Points

### SQLite Replacing JSON Store

The current `store.Store` struct uses in-memory `[]Todo` with JSON file persistence. The migration path:

1. **Interface extraction.** The existing `store.Store` methods (`Add`, `Toggle`, `Delete`, `Update`, `Find`, `Todos`, `TodosForMonth`, etc.) become an interface. The new SQLite implementation satisfies the same interface.
2. **JSON-to-SQLite migration.** On first launch with no `todos.db`, check for `todos.json`. If it exists, read it, insert all todos into SQLite, rename the JSON file to `todos.json.bak`.
3. **Path change.** `~/.config/todo-calendar/todos.json` becomes `~/.config/todo-calendar/todos.db`. The `config/paths.go` file gains a `DBPath()` function alongside the existing `TodosPath()`.
4. **New `body` column.** The `Todo` struct gains a `Body string` field. Stored in SQLite as `TEXT NOT NULL DEFAULT ''`. Existing migrated todos get empty body.

### Glamour Rendering in Todo View

The existing `todolist/model.go` renders each todo as a single line. With v1.4:

1. **Collapsed view** (default): Shows single-line title as before. If body is non-empty, show a small indicator (e.g., `[...]` or a body-present icon).
2. **Expanded view** (toggle): When a todo is selected and user presses a key, render the markdown body below the title using glamour. The renderer needs the panel width (already available via `tea.WindowSizeMsg`).
3. **Glamour theme integration.** Use `glamour.DarkStyleConfig` or `glamour.LightStyleConfig` based on the app's current theme. The theme system already distinguishes dark/light.

### Editor Workflow Integration

The editor launch integrates with the Bubble Tea message loop:

1. **Trigger:** User presses `e` (or configured key) on a selected todo.
2. **Prepare:** Write `todo.Body` to a temp file (e.g., `os.CreateTemp("", "todo-*.md")`).
3. **Launch:** Return `tea.ExecProcess(exec.Command(editor, tempFile), callback)` from `Update()`.
4. **Suspend:** Bubble Tea automatically suspends the TUI and yields the terminal to the editor.
5. **Resume:** Editor exits. Callback fires `editorFinishedMsg`. Read temp file. Update todo body in SQLite. Delete temp file. TUI redraws.

### Template System Integration

Templates connect the config system, file system, and todo creation:

1. **Storage:** `~/.config/todo-calendar/templates/*.md` files. Each file is a `text/template` with `{{.Placeholders}}`.
2. **Config integration.** `config.Config` gains `DefaultTemplate string` field. Settings overlay can list available templates.
3. **Creation flow.** When user creates a todo and selects a template: parse template, prompt for each placeholder value (using existing `bubbles/textinput`), execute template, set result as todo body.
4. **Built-in placeholders.** `{{.Date}}`, `{{.Title}}`, `{{.DayOfWeek}}`, `{{.Month}}`, `{{.Year}}` are auto-populated from context. Custom placeholders prompt the user.

---

## Deliberately NOT Adding

| Consideration | Decision | Rationale |
|---------------|----------|-----------|
| **GORM or any ORM** | Not adding | The app has one table with simple CRUD. Raw `database/sql` with hand-written queries is clearer, faster, and adds zero dependencies. ORMs add complexity, reflection overhead, and a learning curve for contributors -- all for no benefit at this scale. |
| **Migration framework (goose, migrate)** | Not adding | 1-2 schema versions. `PRAGMA user_version` + inline SQL is sufficient and zero-dependency. See rationale above. |
| **Markdown editor component** | Not adding | The project uses `$EDITOR` for editing (vim, neovim, etc.). Building an in-TUI markdown editor would be a massive effort and inevitably worse than the user's preferred editor. The right architectural boundary is: TUI shows read-only preview, external editor handles editing. |
| **Glamour v2** | Not adding | v2 has no stable tagged release (only a pre-release hash from Nov 2025). v0.10.0 is the latest stable and is what production tools use. |
| **ncruces/go-sqlite3** | Not adding | Pre-v1, adds wazero runtime as transitive dep. modernc.org/sqlite is more conservative choice. See driver comparison above. |
| **File-watching (fsnotify)** | Not adding | Considered for detecting external template file changes. Overkill. Reload templates on app startup or when user explicitly refreshes. Personal app restarts are cheap. |
| **Embedded template files (embed.FS)** | Not adding for user templates | User templates must live on disk so users can create/edit them with their text editor. Built-in default templates could use `embed.FS` but a simple `os.ReadFile` is clearer for 1-2 default templates. |

---

## Installation

```bash
# New dependencies for v1.4
go get modernc.org/sqlite@v1.44.3
go get github.com/charmbracelet/glamour@v0.10.0

# Verify
go mod tidy
```

Expected `go.mod` additions:
```
require (
    // ... existing ...
    github.com/charmbracelet/glamour v0.10.0
    modernc.org/sqlite v1.44.3
)
```

Note: `modernc.org/sqlite` brings in `modernc.org/libc`, `modernc.org/mathutil`, and several other `modernc.org/*` transitive dependencies. These are all pure Go and required for the C-to-Go transpilation layer. They are well-maintained by the same author (Jan Mercl).

Glamour brings in `github.com/yuin/goldmark` (markdown parser), `github.com/alecthomas/chroma/v2` (syntax highlighting), and `github.com/microcosm-cc/bluemonday` (HTML sanitizer). Several of its other deps (`lipgloss`, `termenv`, `x/ansi`) are already in the dependency tree.

---

## Binary Size Impact

Rough estimates based on community reports:

| Addition | Approximate Impact |
|----------|--------------------|
| modernc.org/sqlite | +15-20 MB (the transpiled SQLite C library is large) |
| glamour + goldmark + chroma | +5-8 MB (chroma's lexer registry is the bulk) |
| **Total estimated increase** | **+20-28 MB** |

Current binary is likely ~10-12 MB. Final binary will be ~30-40 MB. For a personal desktop TUI app, this is perfectly acceptable. If binary size ever becomes a concern, `go build -ldflags="-s -w"` strips debug info and reduces size by ~30%.

---

## Sources

- [modernc.org/sqlite on pkg.go.dev](https://pkg.go.dev/modernc.org/sqlite) - v1.44.3, Jan 20, 2026. HIGH confidence.
- [ncruces/go-sqlite3 on pkg.go.dev](https://pkg.go.dev/github.com/ncruces/go-sqlite3) - v0.30.5, Jan 24, 2026. HIGH confidence.
- [Go SQLite driver benchmarks](https://github.com/cvilsmeier/go-sqlite-bench) - Aug 2025 benchmarks. MEDIUM confidence (benchmarks are workload-dependent).
- [Charmbracelet glamour on GitHub](https://github.com/charmbracelet/glamour) - v0.10.0. HIGH confidence.
- [glamour v0.10.0 release notes](https://github.com/charmbracelet/glamour/releases/tag/v0.10.0) - Apr 2025. HIGH confidence.
- [Bubble Tea exec example](https://github.com/charmbracelet/bubbletea/blob/main/examples/exec/main.go) - canonical $EDITOR pattern. HIGH confidence.
- [tea.ExecProcess documentation](https://pkg.go.dev/github.com/charmbracelet/bubbletea@v1.3.10) - confirmed in v1.3.10. HIGH confidence.
- [Go text/template documentation](https://pkg.go.dev/text/template) - stdlib, always current. HIGH confidence.
