# Research Summary: v1.4 Data & Editing

**Project:** TUI Calendar - v1.4 Data & Editing Milestone
**Domain:** Desktop TUI todo application with SQLite persistence, markdown bodies, and external editor integration
**Researched:** 2026-02-06
**Confidence:** HIGH

## Executive Summary

v1.4 transforms the TUI Calendar from a JSON-backed single-line todo app into a database-backed system supporting rich markdown bodies. The migration requires three tightly coupled changes: (1) replacing the JSON file store with SQLite for ACID persistence and schema evolution, (2) adding a markdown `body` field to support detailed notes and checklists beyond the title, and (3) integrating external editor support for editing these markdown bodies. The existing Bubble Tea stack already provides the external editor capability via `tea.ExecProcess` with zero new dependencies. SQLite and markdown rendering require exactly 2 new libraries: `modernc.org/sqlite` (pure Go driver, no CGo) and `charmbracelet/glamour` (markdown-to-ANSI renderer from the same ecosystem).

The recommended approach prioritizes data safety and minimizes dependency bloat. Use `modernc.org/sqlite` over the CGo-based `mattn/go-sqlite3` to preserve the project's zero-CGo build. Implement the SQLite store behind an interface extracted from the existing concrete `Store` struct to enable incremental migration and testing. The one-time JSON-to-SQLite migration must be transactional and preserve the original JSON as backup. Markdown templates use Go's stdlib `text/template` for placeholder substitution (zero new dependencies). The editor integration follows the established Bubble Tea `ExecProcess` pattern with a critical gotcha: the `editing` flag must suppress `View()` output to prevent alt-screen buffer leakage.

The key risks are data migration integrity (addressed through transactional migration with backup), SQLite connection pool misconfiguration (solved with `MaxOpenConns(1)` for single-user desktop use), and the subtle alt-screen output leak during editor launch (mitigated with the `editing` flag pattern from Bubble Tea's official examples). The strict dependency chain (SQLite backend → body field → editor integration) means phases cannot be reordered without breaking functionality.

## Key Findings

### Recommended Stack

The existing Bubble Tea (v1.3.10) stack requires minimal additions. External editor support already exists via `tea.ExecProcess`, and templates use stdlib `text/template`. Only 2 new direct dependencies are needed.

**Core technologies:**
- **modernc.org/sqlite v1.44.3**: Pure Go SQLite driver (no CGo required) — preserves zero-CGo build, provides standard `database/sql` interface, actively maintained with 4 releases in January 2026 alone. Trades ~10-50% performance overhead for build simplicity, which is irrelevant for a single-user desktop todo app with sub-1000 todos.
- **charmbracelet/glamour v0.10.0**: Markdown-to-ANSI renderer — integrates with existing Charmbracelet ecosystem (shares lipgloss/termenv deps), purpose-built for terminal markdown rendering with word wrapping and theme support. Used by GitHub CLI (`gh`) and Charmbracelet's own `glow` tool.
- **text/template (stdlib)**: Template system for markdown placeholders — zero new dependencies, handles `{{.Variable}}` substitution with conditionals/loops if needed later. Text (not HTML) template avoids escaping markdown syntax.
- **tea.ExecProcess (already in bubbletea v1.3.10)**: External process execution — built-in TUI suspension, clean terminal handoff, message-passing on editor exit. No new library needed.

**Deliberately rejected:**
- ORMs (GORM, ent): Massive overkill for 1 table with 7 columns. Hand-written SQL is clearer and adds zero dependencies.
- Migration frameworks (goose, migrate): Use `PRAGMA user_version` for schema versioning instead. Single-user desktop app needs simplicity, not multi-environment migration tooling.
- ncruces/go-sqlite3: Pre-v1.0, adds wazero WASM runtime as transitive dependency. modernc.org/sqlite is more conservative and widely adopted.
- Built-in markdown editor: Building even a basic text editor is an enormous undertaking. External `$EDITOR` is the correct architectural boundary.

### Expected Features

**Must have (table stakes):**
- **Automatic JSON-to-SQLite migration**: Users must not lose existing data. First launch after upgrade silently migrates `todos.json` to `todos.db`, preserves all fields, and backs up the JSON file.
- **Identical CRUD behavior post-migration**: All 7 store methods plus 6 query methods must behave identically. SQLite store is a drop-in replacement behind the same interface.
- **Multi-line markdown body field**: Core promise of Phase 15. Todos gain a `Body` field (TEXT column in SQLite) for notes, checklists, and details beyond the single-line title.
- **$EDITOR integration**: Standard Unix workflow (check `$VISUAL` → `$EDITOR` → fallback to `vi`). TUI suspends, editor opens temp file, TUI resumes on editor exit and saves body content.
- **Markdown preview in-app**: When viewing a todo, render the body (or display raw text in scrollable area). Use Glamour for markdown-to-ANSI rendering, displayed in a viewport component.
- **Template CRUD**: Users create/list/delete templates. Templates are stored in a `templates` table with `id`, `name`, `body` (markdown with `{{.Placeholder}}` syntax).
- **Create todo from template**: Pick template, get prompted for placeholder values, create todo with filled-in body. Uses `text/template` or simple string substitution.

**Should have (competitive):**
- **Inline body preview in list**: Show first 1-2 lines of body below todo title in list view (truncated). Gives context without opening full view. Requires multi-line list item height calculation — deferred to v1.5.
- **Default template setting**: Config field `default_template` automatically applies a template when creating new todos. Reduces friction for repetitive workflows.
- **Visual body indicator**: In todo list, show `[+]` icon or marker for todos with non-empty bodies. Simple `len(todo.Body) > 0` check in render function.
- **Search includes body text**: Extend `SearchTodos` to query `WHERE text LIKE ? OR body LIKE ?`. For v1.4, basic LIKE search works; FTS5 virtual table is a v1.5 optimization.

**Defer (v2+):**
- **Checkbox toggling in preview**: Parse markdown checkboxes (`- [ ] subtask`), toggle directly from preview without opening editor. High complexity, requires line position tracking and body updates.
- **Template inheritance/composition**: Templates extending other templates or including partials. Over-engineering for a personal app. Users want standalone templates, not a programming language.
- **WYSIWYG markdown editing**: Terminal WYSIWYG markdown is not a solved problem. Render markdown for display, edit as raw markdown in external editor.

### Architecture Approach

The migration from JSON to SQLite requires interface extraction to decouple consumers (calendar, todolist, search models) from the concrete store implementation. Extract a `TodoStore` interface matching the existing `Store` API, then implement `SQLiteStore` behind this interface. The SQLite backend uses query-on-read semantics (no in-memory cache) — the dataset is small enough (<1000 todos) that querying the database on every `View()` cycle has zero perceptible latency and eliminates cache invalidation bugs.

**Major components:**
1. **Store Interface (`store.TodoStore`)**: Defines domain operations (Add, Toggle, Delete, Update, Find, Search, TodosForMonth, etc.). Both JSON and SQLite stores implement this interface. Consumers depend on the interface, not concrete types.
2. **SQLite Store (`store.SQLiteStore`)**: Implements `TodoStore` via `database/sql` with hand-written SQL queries. Uses `modernc.org/sqlite` driver. Schema includes `todos` table with `body TEXT DEFAULT ''` column from the start (Phase 14 schema anticipates Phase 15 body field).
3. **JSON-to-SQLite Migration (`store/migrate.go`)**: One-time migration triggered on first launch when `todos.db` doesn't exist but `todos.json` does. Transaction-wrapped INSERT of all todos with explicit IDs (not autoincrement), then rename JSON to `.bak`.
4. **External Editor Package (`internal/editor/`)**: Encapsulates temp file management, markdown parsing (extract title from `# heading`), and `$EDITOR` resolution. Returns `tea.ExecProcess` command that suspends TUI and delivers `EditorFinishedMsg` on completion.
5. **Markdown Template System**: Templates stored in `templates` table. Creation flow: parse template for `{{.Variable}}` placeholders, prompt user for each value, execute `text/template`, store expanded result as todo body. Expansion happens once at creation — bodies are plain markdown thereafter.

**Critical pattern: Editing flag for alt-screen workaround**: Bubble Tea's `ExecProcess` with `tea.WithAltScreen()` leaks the final `View()` output to the normal terminal buffer during alt-screen teardown. Solution: Add `editing bool` to `app.Model`. When `editing == true`, `View()` returns empty string to suppress the leak. Set flag before returning `tea.ExecProcess` command, clear it in `editorFinishedMsg` handler.

### Critical Pitfalls

1. **Store Interface Extraction Breaks Existing Callers**: Replacing JSON with SQLite while simultaneously changing method signatures (adding `error` returns) creates a massive, error-prone changeover. Extract the `TodoStore` interface BEFORE changing the backend. Keep the interface matching current API (no new error returns during transition). SQLite failures are rare on local desktop — log errors internally rather than propagating them to TUI layer.

2. **JSON-to-SQLite Migration Loses Data or Runs Repeatedly**: Migration must be idempotent (check if SQLite already has data before migrating). Wrap migration in a transaction (BEGIN → insert all todos → COMMIT). Preserve `todos.json` as `.bak` after successful migration, never delete it. Handle `NextID` explicitly: use existing JSON IDs as explicit values in INSERT (not autoincrement), then let autoincrement take over for new todos.

3. **SQLite "Database Is Locked" Errors on Single-User Desktop App**: Go's `database/sql` uses connection pooling by default. Multiple connections trigger file-level lock contention in SQLite. Solution: `db.SetMaxOpenConns(1)` for single-connection mode. Enable WAL mode (`PRAGMA journal_mode=WAL`), set `PRAGMA busy_timeout=5000`, and `PRAGMA foreign_keys=ON`. These pragmas must be set via DSN string or per-connection hook (they're not all persistent across reconnections).

4. **tea.ExecProcess Leaks View Content to Terminal**: During alt-screen exit for editor launch, Bubble Tea renders one final frame to the normal buffer. Without mitigation, the entire TUI layout appears as garbled text in terminal scrollback. Workaround: `editing bool` flag that makes `View()` return empty string during editor execution (official pattern from Bubble Tea exec example).

5. **Temp File Extension Breaks Editor Syntax Highlighting**: If temp file uses `.tmp` extension, editors don't detect markdown filetype. Users lose syntax highlighting, spell checking, and markdown keybindings. Solution: Use `.md` extension in `os.CreateTemp("", "todo-*.md")` pattern. Clean up temp file in `editorFinishedMsg` handler with `os.Remove()`.

## Implications for Roadmap

Based on research, the milestone has a strict 3-phase dependency chain that cannot be reordered:

### Phase 14: Database Backend
**Rationale:** Foundation for all v1.4 features. The SQLite backend must be working before adding body field or editor integration. Highest-risk item (data migration, new dependency, persistence change) should be addressed first so issues surface early.
**Delivers:** SQLite store behind `TodoStore` interface, automatic JSON-to-SQLite migration with backup, schema versioning with `PRAGMA user_version`, WAL mode enabled, single-connection pool configuration.
**Addresses:** Must-have table-stakes features (migration, CRUD parity, crash safety, schema evolution).
**Avoids:** Pitfalls #1 (interface extraction first), #2 (transactional migration), #3 (connection config), and Patterns 1-2 (directory creation, ID preservation).

### Phase 15: Markdown Templates
**Rationale:** Requires Phase 14 (SQLite) because the `body TEXT` column must exist in the database schema. Adding the body field is a prerequisite for the editor integration (Phase 16) to have content to edit. Templates provide structured content creation but are secondary to the core body field feature.
**Delivers:** `body TEXT DEFAULT ''` column on todos table (if not already in Phase 14 schema), templates table (`id`, `name`, `body`), template CRUD operations, create-todo-from-template flow with placeholder prompting, body indicator in todo list rendering.
**Uses:** Glamour (markdown rendering), bubbles/viewport (scrollable preview), `text/template` (stdlib placeholder substitution).
**Implements:** Markdown template system component, body preview rendering in TUI.
**Avoids:** Pitfalls #6 (body field in schema from Phase 14), Integration #3 (pass pane width to renderer), Pattern 4 (expand templates once, store result), Pattern 7 (title-only in list view).

### Phase 16: External Editor
**Rationale:** The culmination of v1.4. Requires Phase 15 (body field must exist to edit). Uses `tea.ExecProcess` which is already available in current Bubble Tea version — no new dependency. This is the user-facing feature tying everything together and should be built last once the data infrastructure is stable.
**Delivers:** `e` key binding to open selected todo in `$VISUAL`/`$EDITOR`/`vi`, temp file with `.md` extension, body content saved back on editor exit, error handling for missing/failed editor, help bar shows keybinding.
**Uses:** `tea.ExecProcess` (bubbletea v1.3.10), `internal/editor/` package (template rendering, markdown parsing, editor resolution).
**Implements:** External editor integration component, editing state management in app model.
**Avoids:** Pitfall #4 (editing flag + View guard), #5 (.md extension), Integration #4 (RefreshIndicators after save), Integration #5 (split $EDITOR on spaces, fallback chain), Patterns 3, 5 (exit code handling, resize during edit).

### Phase Ordering Rationale

- **Strict dependency chain enforced**: Phase 14 enables 15 (SQLite provides body column), Phase 15 enables 16 (body content exists to edit). These cannot be reordered without breaking functionality. The research confirms no parallelization is possible.
- **Risk-first strategy**: Highest-risk item (database migration) comes first. If migration fails or SQLite integration has issues, the entire milestone is blocked. Building the risky foundation first means failures surface early when recovery cost is lower.
- **Incremental testability**: Each phase delivers a working state. Phase 14: app runs with SQLite backend, all existing features work. Phase 15: todos have markdown bodies visible in preview. Phase 16: bodies are editable via external editor. Each checkpoint is independently testable and shippable.
- **Interface abstraction prevents cascade failures**: Extracting the `TodoStore` interface in Phase 14 means Phases 15 and 16 build on a stable abstraction. If SQLite implementation has bugs, they're isolated behind the interface. If editor integration has issues, the store layer is unaffected.

### Research Flags

**Phases with well-documented patterns (no additional research needed):**
- **Phase 14**: SQLite with `database/sql` is a well-trodden path in Go. The `modernc.org/sqlite` driver documentation is comprehensive, migration patterns are established, and community resources are abundant. No research-phase needed — proceed directly to planning.
- **Phase 16**: Bubble Tea's `tea.ExecProcess` API is documented with official examples. External editor integration is a standard Unix pattern (git commit, crontab -e). No research-phase needed — follow established patterns.

**Phases that MAY benefit from targeted research during planning:**
- **Phase 15**: Markdown template placeholder prompting is a novel feature (no TUI todo app competitor offers this). The basic pattern (parse placeholders → prompt user → execute template) is straightforward, but the UX details (prompt sequencing, validation, error handling) may need a quick research spike during planning to survey text input best practices in TUI apps. This is a 30-minute "how do other Bubble Tea apps handle multi-step user input flows?" question, not a full research-phase.

**No deep research needed for any phase**: All 3 phases use well-understood technologies with high-confidence documentation. The PITFALLS.md research already identified the gotchas (alt-screen leak, connection pooling, migration integrity). Proceed to planning with existing research.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All recommendations based on official pkg.go.dev docs, verified library versions, and active community benchmarks. The modernc.org/sqlite driver is battle-tested in production (River queue, Watermill). Glamour is used by GitHub CLI. |
| Features | HIGH | Table-stakes features derived from codebase analysis (existing Store API must be preserved) and established TUI/CLI patterns (external editor is standard). Differentiators validated against competitor research (dstask, tui-journal, notepad-tui). |
| Architecture | HIGH | Store interface extraction is a standard Go refactoring pattern. SQLite query-on-read for small datasets is well-documented. Bubble Tea `ExecProcess` API is confirmed in v1.3.10 with official examples. Alt-screen workaround verified from maintainer-confirmed GitHub issues. |
| Pitfalls | HIGH | Critical pitfalls derived from: (1) codebase analysis (current Store API, alt-screen usage), (2) verified Bubble Tea GitHub issues (#424, #431), (3) multiple independent sources on Go+SQLite connection pooling, and (4) SQLite PRAGMA documentation. |

**Overall confidence:** HIGH

### Gaps to Address

**Markdown rendering width in narrow terminals:** The research recommends passing pane width to Glamour's renderer, but testing is needed at various terminal sizes (80, 100, 120+ columns) to ensure the todo list remains usable. The fallback of showing only title-in-list with body-in-preview sidesteps this issue, but if inline preview is desired, responsive layout testing is required. Handle during Phase 15 planning by deciding the display strategy (indicator-only vs. preview pane vs. inline expansion) upfront.

**Template placeholder parsing complexity:** The research assumes simple `{{.Variable}}` patterns. If users want advanced features (placeholders with defaults like `{{.Date | default "today"}}`, multi-select placeholders, date pickers), the parsing becomes significantly more complex. For v1.4, strictly limit to simple text input prompts for each placeholder. Document this decision in Phase 15 plan to avoid scope creep. Advanced placeholder types are explicitly deferred to v2+.

**Editor exit code semantics:** The research recommends comparing file content before/after editor execution instead of trusting exit codes (vim `:cq` returns 1 but is intentional discard, some plugin errors return non-zero even on successful save). This pattern needs validation during Phase 16 implementation — does `os.Stat` modification time check or content hash comparison perform better? Test with vim, neovim, nano, and emacs to ensure the pattern works across editors.

**SQLite PRAGMA persistence across reconnections:** The research identifies that some pragmas (busy_timeout, foreign_keys, synchronous) are per-connection settings. With `MaxOpenConns(1)` and infinite `ConnMaxLifetime`, the connection should stay alive forever, but the DSN pragma syntax (`?_pragma=busy_timeout(5000)`) is the more robust approach. Verify during Phase 14 implementation that the `modernc.org/sqlite` driver supports DSN pragmas (documentation suggests it does, but confirm via testing).

## Sources

### Primary (HIGH confidence)
- **Bubble Tea pkg.go.dev v1.3.10** — Confirmed `tea.ExecProcess` API, message-passing patterns
- **Bubble Tea exec example (official GitHub repo)** — Editor integration pattern, editing flag workaround
- **Bubble Tea GitHub Issue #431, Discussion #424** — Alt-screen output leak behavior, maintainer-confirmed workaround
- **modernc.org/sqlite pkg.go.dev v1.44.3** — Pure Go SQLite driver, `database/sql` compatibility, release notes
- **SQLite official documentation** — PRAGMA user_version, WAL mode, journal modes, connection pragmas
- **Go text/template stdlib docs** — Template syntax, execution model, text vs HTML templates
- **Charmbracelet Glamour GitHub v0.10.0** — Markdown-to-ANSI rendering, word wrap configuration
- **Current codebase analysis** — Store API (store.go), Todo struct (todo.go), app model (app/model.go), alt-screen usage (main.go)

### Secondary (MEDIUM confidence)
- **Go SQLite benchmarks (github.com/cvilsmeier/go-sqlite-bench)** — Performance comparison modernc.org vs mattn vs ncruces (Aug 2025)
- **SQLite performance tuning blog (phiresky.github.io)** — WAL mode, synchronous=NORMAL, cache_size pragmas for local apps
- **Go + SQLite best practices (jacob.gold)** — Connection pooling, MaxOpenConns(1) for single-user apps, pragma configuration
- **dstask, tui-journal, notepad-tui GitHub repositories** — Competitor feature analysis, external editor patterns, SQLite usage in TUI apps
- **$VISUAL vs $EDITOR convention (bash.cyberciti.biz)** — Fallback chain standard ($VISUAL → $EDITOR → vi)

### Tertiary (LOW confidence, needs validation)
- **Glamour v2 module path** — v2 exists but has no stable tagged release, use v0.10.0 instead
- **ncruces/go-sqlite3 performance claims** — Benchmarks show it's faster than modernc, but it's pre-v1.0 and adds wazero runtime

---
*Research completed: 2026-02-06*
*Ready for roadmap: yes*
