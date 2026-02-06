# Features Research: v1.4 Data & Editing

**Domain:** SQLite backend, markdown todo bodies with templates, external editor integration
**Researched:** 2026-02-06
**Confidence:** HIGH for database migration and editor integration (well-documented patterns), MEDIUM for markdown templates (less established precedent in TUI todo apps)

## Database Backend (Phase 14)

### Table Stakes

Features that are absolutely required when migrating from JSON to SQLite. Missing any of these and the migration feels broken or untrustworthy.

| Feature | Why Expected | Complexity | Depends On | Notes |
|---------|--------------|------------|------------|-------|
| Automatic one-time JSON-to-SQLite migration | Users must not lose existing data. The app has been shipping with JSON since v1.0. First launch after upgrade must silently migrate. | MEDIUM | Existing `store.Data` JSON structure | Read existing `todos.json`, insert all rows into SQLite, preserve all fields (ID, Text, Date, Done, CreatedAt, SortOrder). Migrate `NextID` into a SQLite sequence or tracked value. |
| Identical CRUD behavior post-migration | All 7 store methods (Add, Toggle, Delete, Update, Find, SwapOrder, SearchTodos) plus all query methods (Todos, TodosForMonth, FloatingTodos, IncompleteTodosPerDay, TodoCountsByMonth, FloatingTodoCounts) must behave identically. | MEDIUM | Current `Store` interface | The SQLite store must be a drop-in replacement. Tests should pass against both backends. |
| Backup of JSON file before migration | Users need a safety net. If migration goes wrong, the original JSON file should still be intact. | LOW | File system access | Rename `todos.json` to `todos.json.bak` after successful migration, or copy it before starting. |
| Schema versioning with migration support | Future schema changes (Phase 15 adds `body` column, templates table) need a clean upgrade path. | LOW | SQLite `PRAGMA user_version` or a `schema_version` table | Use `PRAGMA user_version` for simplicity in a single-user app. Check version on open, apply pending migrations. |
| Atomic writes / crash safety | The existing JSON store uses atomic temp-file-then-rename pattern. SQLite must provide equivalent or better crash safety. | LOW | SQLite WAL mode | SQLite with WAL mode is inherently more crash-safe than the current atomic JSON pattern. Enable `PRAGMA journal_mode=WAL` on first open. |
| Same file location convention | Data should live in the same XDG config directory (`~/.config/todo-calendar/`). | LOW | `config.paths.go` pattern | Place database at `~/.config/todo-calendar/todos.db` alongside existing `config.toml`. |
| Store interface abstraction | The rest of the app should not know or care whether the backend is JSON or SQLite. | MEDIUM | Current `*store.Store` usage throughout codebase | Define a `Store` interface that both JSON and SQLite implementations satisfy. The app layer depends on the interface, not the concrete type. |

### Differentiators

Features that go beyond basic migration and add real value from the database upgrade.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Query-based filtering instead of in-memory scan | Current store scans all todos in memory for every query (TodosForMonth, SearchTodos, etc.). SQLite enables indexed queries that scale to thousands of todos. | MEDIUM | Add indexes on `date`, `done`, and a full-text search index for text. For a personal app this is marginal, but it is "free" with SQLite. |
| Transaction support for batch operations | Currently each mutation calls `Save()` which rewrites the entire JSON file. SQLite transactions make batch operations (e.g., bulk delete, reordering multiple items) atomic and fast. | LOW | Wrap multi-step operations in `BEGIN/COMMIT`. |
| Foundation for future features (tags, projects, recurring todos) | SQLite makes adding columns, new tables, and relational data trivial. JSON requires rewriting the entire file for any structural change. | N/A | This is the strategic reason for the migration. It enables Phase 15 (body/templates) and future milestones. |

### Anti-Features

Things to deliberately NOT build during the database migration.

| Anti-Feature | Why It Seems Related | Why Avoid | What to Do Instead |
|--------------|---------------------|-----------|-------------------|
| ORM or heavy query builder | Go has GORM, ent, sqlc. Seems like the "proper" way to do database access. | This is a 3,200 LOC personal app with 6 tables max. An ORM adds dependency weight, code generation steps, and abstraction overhead that is not justified. The notepad-tui project used GORM and it adds significant dependency bloat. | Use `database/sql` with raw SQL strings. The queries are simple (CRUD + a few aggregates). Hand-written SQL is more readable and debuggable at this scale. |
| Concurrent multi-process access | SQLite supports multiple readers. Could design for external tools accessing the same DB. | Single-user TUI app. One process at a time. Designing for concurrency adds complexity (connection pooling, retry logic, lock handling) for zero benefit. | Use `SetMaxOpenConns(1)`. Single connection is simplest and correct for this use case. |
| Encryption at rest | SQLite can be encrypted with SQLCipher. Seems security-conscious. | Personal local app. The JSON file was not encrypted either. Encryption adds a C dependency (SQLCipher) or forces CGo. | Store the DB file with standard filesystem permissions (0600). If users want encryption, they can use full-disk encryption. |
| Export/import commands | Migration feels like it should come with data portability tools. | Over-scoping Phase 14. The migration is one-time JSON-to-SQLite. Export can be added later if needed. | Focus on the one-way migration. Keep the `.bak` JSON file as the "export." |
| Dual backend (keep JSON as option) | Could make the backend configurable: JSON or SQLite. Gives users a choice. | Maintaining two backends doubles testing and bug surface. The whole point of migrating is to move to a better foundation. Having two options means neither gets full attention. | Migrate to SQLite unconditionally. Remove JSON store code after migration is validated. |

---

## Markdown Templates (Phase 15)

### Table Stakes

Features users expect when "markdown body" and "templates" are advertised.

| Feature | Why Expected | Complexity | Depends On | Notes |
|---------|--------------|------------|------------|-------|
| Multi-line body field on todos | The core promise. Todos currently have only a single-line `Text` field. Adding a `Body` field (markdown string) allows notes, checklists, details. | LOW | Phase 14 (SQLite -- add `body TEXT` column) | This is an `ALTER TABLE todos ADD COLUMN body TEXT DEFAULT ''` migration. Existing todos get empty bodies. |
| In-app preview of markdown body | When viewing a todo, users should see the body rendered (or at minimum, displayed as raw text in a scrollable area). | MEDIUM | Glamour library for rendering, bubbles/viewport for scrolling | Use `charmbracelet/glamour` to render markdown to ANSI. Display in a viewport component. If body is empty, show nothing extra (preserve current compact view). |
| Template CRUD (create, list, delete) | If templates exist, users need to manage them. At minimum: create a template, see available templates, delete one. | MEDIUM | New `templates` table in SQLite | Templates are stored as rows: `id`, `name`, `body` (markdown with placeholders). |
| Create todo from template | The primary workflow. User picks a template, gets prompted for placeholder values, and a new todo is created with the filled-in body. | MEDIUM | Template storage, placeholder parsing | Use Go `text/template` or simpler `strings.ReplaceAll` with `{{.VariableName}}` syntax. The simpler approach is better for non-programmers. |
| Placeholder prompting during creation | When creating a todo from a template with `{{.ProjectName}}` and `{{.DueDate}}`, the app must prompt for each value interactively. | MEDIUM | TUI input flow, template parsing | Parse template for `{{.VarName}}` patterns, present each as a textinput prompt in sequence. |
| Template body supports standard markdown | Headers, bullet lists, checkboxes (`- [ ]`), bold, italic, code blocks. This is what users expect from "markdown." | LOW | Glamour rendering handles this natively | Do not build a custom markdown parser. Glamour supports GFM (GitHub Flavored Markdown). |

### Differentiators

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Inline body preview in todo list | Show first 1-2 lines of the body below the todo title in the list view (truncated). Gives at-a-glance context without opening the full view. | MEDIUM | Requires adjusting todo list item height calculation. Multi-line list items need careful layout math. |
| Checkbox toggling in preview | If the body contains `- [ ] subtask`, allow toggling checkboxes directly from the preview without opening the editor. dstask uses this pattern -- "checklists are useful here." | HIGH | Would require parsing markdown checkboxes, tracking line positions, updating body on toggle. Likely too complex for v1.4. |
| Default template for new todos | A setting that automatically applies a specific template when creating a new todo, so every todo starts with a consistent structure. | LOW | Config field `default_template` referencing a template name/ID. |
| Template variables with defaults | Placeholders like `{{.Date | default "today"}}` that auto-fill if the user skips the prompt. Reduces friction for repetitive templates. | MEDIUM | Requires extending the placeholder parser beyond simple substitution. |

### Anti-Features

| Anti-Feature | Why It Seems Related | Why Avoid | What to Do Instead |
|--------------|---------------------|-----------|-------------------|
| Built-in markdown editor | If todos have markdown bodies, building an in-app editor seems natural. TUI-Journal has a built-in editor with Emacs keybindings. | Building a good text editor is an enormous undertaking. Even TUI-Journal's built-in editor is limited and they still offer external editor fallback. The external editor (Phase 16) is the correct solution. | Use external editor for editing bodies. The TUI shows a read-only rendered preview. |
| WYSIWYG markdown editing | Rich text editing in the terminal. Bold text appears bold as you type. | This does not exist in a practical form for terminal apps. Even the nhn/tui.editor is a web component, not a terminal one. Terminal WYSIWYG markdown is not a solved problem. | Render markdown for display. Edit as raw markdown in external editor. |
| Template inheritance / composition | Templates that extend other templates, or templates that include partials. | Go `text/template` supports this, but it is over-engineering for a todo app. Users want "meeting notes template" and "weekly review template," not a template programming language. | Flat templates only. Each template is a standalone markdown document with simple `{{.Variable}}` placeholders. |
| Template sharing / import-export | Import templates from files, export to share with others. | This is a personal single-user app. Template sharing adds UI for file picking, format validation, conflict handling. | Templates are managed in-app only. Users can create templates by typing or pasting markdown. If they want to share, they can copy the body text. |
| Markdown rendering of the title field | Could render the todo title as markdown (bold, links, etc.). | The title is a single line shown in a list. Markdown rendering would break alignment, introduce variable widths, and make the list visually inconsistent. | Title remains plain text. Only the body field is markdown. |
| Complex placeholder types (date pickers, dropdowns) | Templates could have typed placeholders: `{{.Date:date}}` for a date picker, `{{.Priority:select:high,medium,low}}`. | Massive UI complexity for minimal gain. Text input covers all cases. | All placeholders are text inputs. Users type whatever they want. The template provides structure, not validation. |

---

## External Editor (Phase 16)

### Table Stakes

Features users expect when an app offers external editor integration. This is a well-established pattern in CLI/TUI apps (git commit, crontab -e, kubectl edit, lazygit).

| Feature | Why Expected | Complexity | Depends On | Notes |
|---------|--------------|------------|------------|-------|
| `$VISUAL` / `$EDITOR` environment variable support | The Unix standard. Check `$VISUAL` first (screen-based editor), then `$EDITOR` (line editor), then fall back to a sensible default. | LOW | `os.Getenv` | Fallback chain: `$VISUAL` -> `$EDITOR` -> `vi`. Not `vim` -- `vi` is more universally available. TUI-Journal uses the same chain with `vi` as final fallback. |
| Clean TUI suspend and resume | The TUI must fully release the terminal before the editor launches, and fully restore it after the editor exits. No visual artifacts, no lost state. | LOW | Bubble Tea `tea.ExecProcess` handles this natively | Bubble Tea's `ExecProcess` calls `ReleaseTerminal()`, runs the process, then `RestoreTerminal()`. This is battle-tested. |
| Write todo body to temp file, open in editor, read back | The standard temp-file workflow: write current body to a temp file, open editor pointing at that file, read file contents after editor exits, update the todo body. | LOW | `os.CreateTemp`, `os.ReadFile` | Use `.md` extension on the temp file so editors enable markdown syntax highlighting. TUI-Journal configures this with `temp_file_extension`. |
| Handle editor errors gracefully | If the editor crashes, exits with non-zero status, or the temp file is deleted, the app must not crash or corrupt data. | LOW | Error handling in `ExecCallback` | If `err != nil` from `ExecProcess`, show an error message and keep the original body unchanged. |
| Single keybinding to open editor | A clear, discoverable key (e.g., `e` for edit) that opens the currently selected todo in the editor. | LOW | Key binding system, help bar | Only active when a todo is selected. Show in help bar: `e: edit in editor`. |
| Preserve original body if editor exits without saving | If the user opens vim, types `:q!` (quit without saving), the body should remain unchanged. | LOW | Compare file modification time or content hash before/after | Simplest approach: read file after editor exits. If content is the same as what was written, no update. Alternatively, always read and update -- if user quit without saving, the file content is unchanged anyway. |
| App state preserved during editor session | When the user returns from the editor, the app should be in the same state: same todo selected, same calendar month, same view mode. | LOW | Bubble Tea handles this -- the `Model` is preserved across `ExecProcess` | The Elm Architecture means model state persists. The `ExecProcess` only suspends rendering, not state. |

### Differentiators

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Editor setting in config/settings overlay | Allow overriding the editor in `config.toml` (e.g., `editor = "nvim"`) without setting environment variables. Settings overlay gets a new "Editor" row. | LOW | New config field. If set, takes priority over `$VISUAL`/`$EDITOR`. Useful for users who want a different editor for todos vs. git commits. |
| Pre-populated body from template on first edit | If a todo has no body yet, opening the editor could pre-populate the temp file with a default template or structured markdown (e.g., `# Notes\n\n## Checklist\n- [ ] `). | LOW | Check if body is empty, write template instead of empty file. |
| Edit title + body together | The temp file could contain the title as the first line and the body below a separator (like git commit messages: first line = subject, blank line, then body). | MEDIUM | Requires parsing the file back: split on first blank line, extract title and body. Edge cases with empty titles, missing separators. |
| Visual indicator that a todo has a body | In the todo list, show a small icon or marker (e.g., a trailing `[+]` or different color) for todos that have a non-empty body. | LOW | Check `len(todo.Body) > 0` in the render function. Simple visual cue. |

### Anti-Features

| Anti-Feature | Why It Seems Related | Why Avoid | What to Do Instead |
|--------------|---------------------|-----------|-------------------|
| Built-in editor as fallback | If `$EDITOR` is not set and `vi` is not available, could provide a basic built-in editor. | Building even a minimal editor (multi-line text input with save/cancel) is surprisingly complex. The bubbles library has `textarea` but it is not a file editor. If the user has no editor installed, that is a system configuration issue, not our problem. | Show a clear error message: "No editor found. Set $EDITOR or install vi." |
| Auto-save / live sync while editor is open | Could watch the temp file for changes and update the todo body in real-time while the editor is still open. | The TUI is suspended during `ExecProcess`. There is no event loop running to watch files. This would require a fundamentally different architecture (background process, file watcher). | Read the file once when the editor exits. Single-shot update. |
| Multiple file editing | Open multiple todos in separate editor buffers/tabs. | This requires a much more complex temp file management system, and editors handle multi-file workflows differently. | One todo at a time. User edits one, saves, returns, then can open another. |
| Diff view of changes | After editor exits, show a diff of what changed before saving. | Over-engineering. The user just edited the content -- they know what they changed. A diff view adds UI complexity for no practical benefit in a personal app. | Save changes immediately on editor exit. The user can undo by reopening the editor. |
| Remote editor / SSH editor support | The Charmbracelet `wish` library enables running Bubble Tea over SSH. Supporting remote editors would extend this. | The `wish` library's `ExecProcess` does not work for remote editor spawning (confirmed by GitHub issue #196). This is a local-only app. | Local editor only. If running over SSH, the user's `$EDITOR` on the remote machine works naturally. |

---

## Feature Dependencies

```
Phase 14: Database Backend
    |
    +-- No dependencies on other v1.4 features
    +-- Replaces: internal/store/store.go (JSON implementation)
    +-- Preserves: Store interface used by app, calendar, todolist, search
    +-- Enables: Phase 15 (body column), Phase 16 (body editing)

Phase 15: Markdown Templates
    |
    +-- REQUIRES Phase 14 (SQLite for body column + templates table)
    +-- REQUIRES charmbracelet/glamour (new dependency for rendering)
    +-- REQUIRES bubbles/viewport (already available, for scrollable body preview)
    +-- Modifies: Todo struct (add Body field)
    +-- Modifies: Todo list rendering (show body preview)
    +-- Modifies: Store interface (add body to CRUD, add template methods)
    +-- Enables: Phase 16 (body content exists to edit)

Phase 16: External Editor
    |
    +-- REQUIRES Phase 15 (body field must exist to edit)
    +-- REQUIRES tea.ExecProcess (already available in bubbletea v1.3.10)
    +-- Modifies: Todo list keys (add 'e' for edit)
    +-- Modifies: App model (handle editor launch/return messages)
    +-- Modifies: Help bar (show editor keybinding)
    +-- Independent of: Glamour rendering (editor works with raw markdown)
```

**Strict dependency chain: Phase 14 -> Phase 15 -> Phase 16.** These cannot be reordered. Each phase builds on the previous one's data model changes.

## MVP Recommendation

For each phase, the minimum viable implementation:

### Phase 14 MVP (Database Backend)
1. SQLite store implementing the existing Store interface
2. Automatic JSON-to-SQLite migration on first launch
3. JSON file backed up as `.bak`
4. Schema versioning with `PRAGMA user_version`
5. WAL mode enabled

**Defer:** Query optimization, indexes, full-text search. The dataset is small enough that table scans are fine.

### Phase 15 MVP
1. `body TEXT` column added to todos table
2. Templates table with `id`, `name`, `body`
3. Body preview (rendered markdown) visible when todo is selected (could be a detail pane, overlay, or inline)
4. At least one way to create/manage templates
5. Create-todo-from-template flow with placeholder prompting

**Defer:** Inline body preview in the list (multi-line items are hard), checkbox toggling, default template setting.

### Phase 16 MVP
1. `e` key opens selected todo body in `$VISUAL`/`$EDITOR`/`vi`
2. Temp file with `.md` extension
3. Body saved back on editor exit
4. Error handling for missing editor / editor failure
5. Help bar shows editor keybinding

**Defer:** Config-based editor override, title+body combined editing, pre-populated template for empty bodies.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Risk | Priority |
|---------|-----------|-------------------|------|----------|
| JSON-to-SQLite migration | CRITICAL | MEDIUM | MEDIUM (data loss risk) | P0 |
| Store interface abstraction | HIGH | MEDIUM | LOW | P0 (part of Phase 14) |
| Schema versioning | HIGH | LOW | LOW | P0 (part of Phase 14) |
| Body field on todos | HIGH | LOW | LOW | P1 (Phase 15 start) |
| External editor for body | HIGH | LOW | LOW | P1 (Phase 16 core) |
| Markdown preview in-app | MEDIUM | MEDIUM | LOW | P2 (Phase 15) |
| Template CRUD | MEDIUM | MEDIUM | LOW | P2 (Phase 15) |
| Create from template | MEDIUM | MEDIUM | MEDIUM | P2 (Phase 15) |
| Body indicator in list | MEDIUM | LOW | LOW | P3 (nice-to-have) |
| Editor config in settings | LOW | LOW | LOW | P3 (nice-to-have) |

## Competitor Feature Comparison (v1.4 Scope)

| Feature | dstask | tui-journal | notepad-tui | taskwarrior | Our v1.4 Approach |
|---------|--------|-------------|-------------|-------------|-------------------|
| Storage backend | JSON files (git-synced) | JSON or SQLite (configurable) | SQLite (GORM) | Custom binary format | SQLite only (migrated from JSON) |
| Note/body per task | Markdown note per task (`note` command) | Full markdown body with title | Markdown in BLOB column | Annotations (short text) | Markdown body field per todo |
| Templates | None | None | None | Recurrence only | Markdown templates with placeholders |
| External editor | `$EDITOR` for notes | Built-in + external fallback | `$EDITOR` for .md files | `$EDITOR` for annotations | `$VISUAL`/`$EDITOR` via tea.ExecProcess |
| Markdown rendering | Raw text display | Built-in rendering | Not in TUI | N/A | Glamour-rendered preview in viewport |
| Body in list view | Not shown in list | Not shown in list | Not in TUI | Not shown | Body indicator icon, optional preview |

**Key insight:** Markdown templates with placeholder prompting is genuinely novel among TUI todo apps. No competitor found in research offers this. This is the primary differentiator of v1.4.

## Sources

### HIGH Confidence (official docs, authoritative)

- [Bubble Tea `tea.ExecProcess` API and exec.go source](https://github.com/charmbracelet/bubbletea/blob/main/exec.go) -- Confirmed API for suspending TUI and launching external process
- [Bubble Tea exec example](https://github.com/charmbracelet/bubbletea/blob/main/examples/exec/main.go) -- Working example of editor integration pattern
- [Charmbracelet Glamour markdown rendering](https://github.com/charmbracelet/glamour) -- Stylesheet-based markdown to ANSI rendering, used by glow, gh CLI, and GitLab CLI
- [SQLite PRAGMA user_version](https://www.sqlite.org/pragma.html) -- Official SQLite pragma documentation for schema versioning
- [SQLite WAL mode documentation](https://sqlite.org/wal.html) -- Official write-ahead logging documentation
- [Go `text/template` package](https://pkg.go.dev/text/template) -- Standard library template system with `{{.Variable}}` syntax
- [modernc.org/sqlite CGo-free driver](https://pkg.go.dev/modernc.org/sqlite) -- Pure Go SQLite implementation, no C compiler needed

### MEDIUM Confidence (multiple sources agree, verified patterns)

- [SQLite migrations with PRAGMA user_version](https://levlaz.org/sqlite-db-migrations-with-pragma-user_version/) -- Practical pattern for lightweight schema versioning
- [modernc.org/sqlite vs mattn/go-sqlite3 benchmarks](https://datastation.multiprocess.io/blog/2022-05-12-sqlite-in-go-with-and-without-cgo.html) -- Performance comparison showing modernc is 2x slower but CGo-free
- [tui-journal architecture](https://github.com/AmmarAbouZor/tui-journal) -- Reference implementation: Rust TUI with SQLite + markdown + external editor
- [dstask markdown note per task](https://github.com/naggie/dstask) -- Go CLI todo with markdown notes, `$EDITOR` integration
- [notepad-tui SQLite + markdown](https://github.com/Tempost/notepad-tui) -- Go TUI with SQLite (GORM) storing markdown in BLOB column
- [$VISUAL vs $EDITOR convention](https://bash.cyberciti.biz/guide/$VISUAL_vs._$EDITOR_variable_%E2%80%93_what_is_the_difference%3F) -- Standard fallback chain: $VISUAL -> $EDITOR -> vi
- [SQLite performance tuning for local apps](https://phiresky.github.io/blog/2020/sqlite-performance-tuning/) -- WAL mode, synchronous=NORMAL, cache_size pragmas

### LOW Confidence (single source, needs validation during implementation)

- Glamour v2 may have different API than v1 -- verify import path and API during Phase 15 implementation
- `tea.ExecProcess` behavior with alt-screen mode -- the app uses `tea.WithAltScreen()`, verify clean transition when editor launches

---
*Feature research for: TUI Calendar v1.4 Data & Editing*
*Researched: 2026-02-06*
