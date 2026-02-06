# Pitfalls Research: v1.4 Data & Editing

**Domain:** TUI Calendar v1.4 -- SQLite backend, markdown todo bodies, external editor integration
**Researched:** 2026-02-06
**Confidence:** HIGH (pitfalls derived from codebase analysis + verified library documentation + community issue reports)

This document covers pitfalls specific to ADDING SQLite persistence, markdown template bodies, and external editor integration to the existing Go Bubble Tea TUI app (3,263 LOC, 13 completed phases).

---

## Critical Pitfalls

### Pitfall 1: Store Interface Extraction Breaks Existing Callers

**What goes wrong:** The current `Store` is a concrete struct pointer (`*store.Store`) threaded through `app.New()`, `todolist.New()`, `search.New()`, and `calendar.Model`. Replacing JSON with SQLite requires changing the Store internals, but every consumer depends directly on the concrete type. If the SQLite store has a different method signature (e.g., methods now return `error` where before they silently called `Save()`), every call site must be updated simultaneously. This creates a massive, error-prone changeover with no incremental testing.

**Why it happens:** The existing store methods (`Add`, `Toggle`, `Delete`, `Update`, `SwapOrder`) call `s.Save()` internally and swallow the error. With SQLite, each operation becomes a database call that can fail (locked, full disk, schema error). The natural impulse is to make methods return `error`, but this changes the interface for all 6+ callers across 4 packages.

**How to avoid:**
- Extract a `Store` interface BEFORE changing the backend. The interface should match the current concrete API exactly -- methods that don't return errors continue not returning errors. The SQLite implementation can log or handle errors internally during the transition.
- Alternatively, keep the current API contract (no returned errors) for the SQLite store too. A personal desktop SQLite database on a local filesystem should virtually never fail. Log errors rather than propagating them up to the TUI layer, which has no meaningful way to display them anyway.
- Do NOT attempt to add error returns to the store methods and simultaneously swap to SQLite. These are two separate changes.

**Warning signs:**
- Store method signatures change (adding `error` returns) in the same commit that adds SQLite
- Compile errors in 4+ packages during the migration
- Tests for existing features break because of interface changes, not SQLite bugs

**Recovery cost:** MEDIUM -- requires touching every call site. If done after SQLite is already integrated, you cannot tell whether bugs are from the interface change or the backend change.

**Phase to address:** Phase 14 (Database Backend). Extract interface first, then swap implementation.

---

### Pitfall 2: JSON-to-SQLite Migration Loses Data or Runs Repeatedly

**What goes wrong:** The auto-migration (read existing `todos.json`, write to SQLite, optionally rename/delete JSON) either: (a) silently loses data if the JSON has fields the SQLite schema doesn't expect, (b) runs on every startup instead of once because the "already migrated" flag is unreliable, or (c) creates duplicate todos if the migration is interrupted mid-way and re-run.

**Why it happens:** The current JSON envelope (`Data{NextID, Todos}`) is simple, but edge cases exist: what if `todos.json` contains a todo with `SortOrder: 0` (legacy data -- the `EnsureSortOrder()` method already handles this)? What if `CreatedAt` is empty (older entries before that field was added)? The migration must handle ALL historical variations of the JSON format, not just the current schema.

**How to avoid:**
- Migration must be idempotent. Check if the SQLite database already has data (a schema version table, or simply `SELECT COUNT(*) FROM todos`). If data exists, skip migration.
- Preserve the original `todos.json` file as `todos.json.bak` after successful migration. Never delete it during migration. Let users manually clean up.
- Use a transaction for the entire migration: BEGIN, insert all todos, COMMIT. If anything fails, ROLLBACK and fall back to JSON mode. Never leave the user with neither working backend.
- Handle the `NextID` value explicitly. SQLite autoincrement and the JSON `NextID` are different ID schemes. Either seed the SQLite autoincrement to `NextID` or use the existing integer IDs as explicit values (not autoincrement).
- Run `EnsureSortOrder()` equivalent during migration, not after, so migrated data is clean.

**Warning signs:**
- Migration code does not check whether SQLite already has data
- No transaction wrapping the migration inserts
- `todos.json` deleted before SQLite write is confirmed
- `NextID` from JSON not carried over, causing ID collisions with existing references

**Recovery cost:** HIGH -- data loss is irrecoverable if the JSON backup was deleted. Duplicate data requires manual dedup.

**Phase to address:** Phase 14 (Database Backend). Migration logic is the highest-risk code in the entire milestone.

---

### Pitfall 3: SQLite Connection Misconfiguration Causes "Database Is Locked" on Single-User Desktop App

**What goes wrong:** Go's `database/sql` package uses a connection pool by default. With SQLite (a file-level lock database), multiple connections from the pool attempt concurrent access and produce `SQLITE_BUSY` / "database is locked" errors. This manifests as random, non-reproducible save failures -- a todo toggle works 99% of the time but occasionally silently fails.

**Why it happens:** The default `database/sql` pool has unlimited `MaxOpenConns`. For PostgreSQL or MySQL this is fine (server handles concurrency). For SQLite, the database file itself IS the server. Multiple goroutines opening connections trigger file-level lock contention. Even with WAL mode, write contention from multiple connections in the same process is problematic.

**How to avoid:**
- Set `db.SetMaxOpenConns(1)` for a single-user desktop app. This eliminates lock contention entirely. SQLite operations are fast enough that serialized access through one connection has no perceptible latency for a personal todo app.
- Set `db.SetMaxIdleConns(1)` and `db.SetConnMaxLifetime(0)` (infinite) to keep that one connection alive.
- Enable WAL mode via pragma: `PRAGMA journal_mode=WAL` -- this allows concurrent reads even with a single writer, useful if future features add background queries.
- Set `PRAGMA busy_timeout=5000` as a safety net, even with single-connection config.
- Set `PRAGMA foreign_keys=ON` explicitly (SQLite defaults to OFF for backward compatibility).
- These pragmas must be run on EVERY new connection. Use a connection-init hook or run them immediately after `sql.Open()`.

**Warning signs:**
- `sql.Open()` called without `SetMaxOpenConns(1)`
- Pragmas set only once at startup instead of per-connection
- Intermittent "database is locked" errors during rapid todo toggling
- `_journal_mode` or `_busy_timeout` not in the DSN or connection setup

**Recovery cost:** LOW -- configuration fix, no data impact. But the intermittent nature makes it hard to diagnose.

**Phase to address:** Phase 14 (Database Backend). Set connection config as the very first thing after `sql.Open()`.

---

### Pitfall 4: tea.ExecProcess Leaks View Content to Terminal

**What goes wrong:** When using `tea.ExecProcess` to launch `$EDITOR`, Bubble Tea exits the alternate screen buffer before the editor starts. During this transition, the `View()` function renders one final frame to the NORMAL terminal buffer (not the alternate screen). The entire TUI layout (calendar, todos, help bar) appears as garbled text in the regular terminal, persisting even after the editor closes and the TUI resumes.

**Why it happens:** This is a known Bubble Tea behavior ([GitHub Issue #431](https://github.com/charmbracelet/bubbletea/issues/431), [Discussion #424](https://github.com/charmbracelet/bubbletea/discussions/424)). When `ExecProcess` is called, the framework does a final render during alternate screen teardown. Since the app uses `tea.WithAltScreen()` (line 42 of `main.go`), this transition leaks the current `View()` output to stdout.

**How to avoid:**
- Set an `editing` boolean flag on the model BEFORE returning the `tea.ExecProcess` command. In `View()`, check this flag and return an empty string (or minimal "Editing..." text):
  ```
  if m.editing {
      return ""
  }
  ```
- After the editor exits (in the `editorFinishedMsg` handler), set `m.editing = false` to restore normal rendering.
- This is the same pattern documented in the official exec example (`m.quitting` flag). It is not optional -- without it, every editor launch produces visual garbage.

**Warning signs:**
- No state flag checked in `View()` before `ExecProcess` returns
- TUI content appears in terminal scrollback after editor closes
- Testing only with non-altscreen mode (where the bug doesn't manifest)

**Recovery cost:** LOW -- adding a flag and View guard is a small change. But if discovered late, users have already seen the broken behavior.

**Phase to address:** Phase 16 (External Editor). Must be in the initial implementation, not a follow-up fix.

---

### Pitfall 5: Temp File for Editor Has Wrong Extension, Breaking Syntax Highlighting

**What goes wrong:** The external editor workflow writes the todo's markdown body to a temp file, opens the editor, then reads it back. If the temp file is created with `os.CreateTemp("", "todo-*.tmp")`, editors like vim/neovim detect filetype from the extension. A `.tmp` extension means no markdown syntax highlighting, no spell checking, and no markdown-specific keybindings. Users editing a markdown body see plain text with no formatting aids.

**Why it happens:** `os.CreateTemp` generates random filenames with the provided pattern. The `*` is replaced with random characters, so `"todo-*.tmp"` produces `todo-abc123.tmp`. Vim/neovim use the extension for filetype detection. The `.tmp` extension maps to no known filetype.

**How to avoid:**
- Use `.md` as the extension in the temp file pattern: `os.CreateTemp("", "todo-*.md")`. This gives vim/neovim the `markdown` filetype automatically.
- Alternatively, pass `+set ft=markdown` as a vim/neovim argument, but this is editor-specific and breaks for other editors (nano, micro, emacs). The `.md` extension approach works universally.
- Clean up the temp file after reading it back. Use `defer os.Remove(tmpFile.Name())` -- but be aware this runs even if the editor fails to launch. That's fine; the file should be cleaned up regardless.

**Warning signs:**
- Temp file pattern uses `.tmp` or no extension
- Editor opens with plain text mode when editing markdown content
- No `defer os.Remove()` for the temp file

**Recovery cost:** LOW -- one-line fix to the temp file pattern. But poor UX if users don't get syntax highlighting for their markdown.

**Phase to address:** Phase 16 (External Editor). Use `.md` extension from the start.

---

### Pitfall 6: Adding "Body" Field to Todo Struct Breaks JSON Backward Compatibility

**What goes wrong:** Phase 15 adds a `Body string` field to the `Todo` struct for markdown content. If the SQLite migration (Phase 14) has already replaced JSON, this isn't a JSON problem. But if Phase 14 and 15 are developed concurrently or Phase 14 is delayed, the new field in the struct changes JSON serialization. Existing `todos.json` files without a `body` field deserialize with `Body: ""` (correct), but if `omitempty` is NOT on the json tag, re-serialized JSON includes `"body": ""` for every todo, bloating the file. If `omitempty` IS used, round-trip works but is cosmetically different.

**Why it happens:** Phase ordering matters. Phase 15 depends on Phase 14 (the roadmap says so). But developers sometimes work on features in parallel or re-order phases. If `Body` is added to the struct while JSON persistence still exists, the JSON format changes.

**How to avoid:**
- Enforce phase ordering strictly. Phase 14 (SQLite) MUST ship before Phase 15 (markdown bodies) begins. The SQLite schema can include the `body TEXT DEFAULT ''` column from the start, making the later addition seamless.
- If phases DO overlap: add `Body string \`json:"body,omitempty"\`` with `omitempty` to maintain JSON backward compatibility during the transition window.
- The SQLite schema should define the `body` column in Phase 14 even though Phase 15 populates it. This avoids a schema migration between phases.

**Warning signs:**
- `Body` field added to Todo struct while JSON store is still in use
- No `omitempty` json tag on the Body field
- SQLite schema created without a `body` column, requiring ALTER TABLE later

**Recovery cost:** LOW if caught during development. MEDIUM if users have already serialized todos with the broken format.

**Phase to address:** Phase 14 design (include body column in schema) and Phase 15 implementation.

---

## Integration Pitfalls

### Integration Pitfall 1: Store.Save() Calls Embedded in Every Mutation Don't Map to SQLite

**What goes wrong:** The current store calls `s.Save()` (full JSON write) inside every mutation: `Add()`, `Toggle()`, `Delete()`, `Update()`, `SwapOrder()`. With SQLite, each mutation is an individual SQL statement (INSERT, UPDATE, DELETE). There's no equivalent of "write the whole file." If a developer naively wraps each SQL call in a transaction, every mutation is its own transaction -- correct but slow for batch operations. Conversely, if mutations are grouped in a transaction that gets interrupted (user hits Ctrl+C mid-batch), partial writes occur.

**Why it happens:** The JSON store has a simple mental model: mutate in-memory, then flush to disk. SQLite has a different model: each SQL statement is immediately durable (with WAL). The `Save()` pattern has no equivalent.

**How to avoid:**
- Remove the `Save()` method from the SQLite store. Each mutation method (Add, Toggle, Delete, Update, SwapOrder) directly executes its SQL statement and returns. There is no "flush" step.
- For the migration insert (bulk operation), use an explicit transaction. For normal CRUD, individual statements are fine -- SQLite autocommit handles durability.
- Do NOT carry over the `s.data.Todos` in-memory slice pattern. The SQLite store should query the database for every read, not maintain an in-memory cache. For a personal todo app (hundreds of todos at most), this eliminates cache-invalidation bugs with zero perceptible performance cost.

**Warning signs:**
- SQLite store maintains an in-memory `[]Todo` alongside the database
- A `Save()` method exists on the SQLite store that writes the in-memory slice to SQLite
- Reads come from memory instead of the database
- Data inconsistencies between in-memory state and database

**Recovery cost:** MEDIUM -- requires rethinking the store's read pattern if the in-memory cache is already integrated.

**Phase to address:** Phase 14 (Database Backend). Design the store as query-on-read from the start.

---

### Integration Pitfall 2: EnsureSortOrder Migration Semantics in SQLite

**What goes wrong:** The current JSON store has `EnsureSortOrder()` which runs at load time to backfill `SortOrder` for legacy todos. In a SQLite backend, this logic needs to be a one-time schema migration, not a per-startup operation. If `EnsureSortOrder()` runs on every startup against SQLite, it performs an unnecessary UPDATE on every todo every time the app launches.

**Why it happens:** The current code calls `EnsureSortOrder()` in `NewStore()` (line 36 of `store.go`). It's harmless for JSON (cheap in-memory loop) but wasteful for SQLite (N UPDATE statements).

**How to avoid:**
- Perform the SortOrder backfill as part of the JSON-to-SQLite migration. During migration, if a todo has `SortOrder == 0`, assign `(row_index + 1) * 10` before inserting into SQLite.
- Do NOT call `EnsureSortOrder()` on the SQLite store. Instead, make `SortOrder` NOT NULL DEFAULT in the schema with a reasonable value. New todos get `MAX(sort_order) + 10` via SQL, matching the current Go logic.
- If you choose to keep `EnsureSortOrder()` for safety, gate it behind a version check: only run if the database is at schema version 1 (initial migration).

**Warning signs:**
- `EnsureSortOrder()` still called in the SQLite store constructor
- EXPLAIN shows full table scans on every app startup
- SortOrder column is nullable in the schema

**Recovery cost:** LOW -- removing the call is trivial. But unnecessary database writes on startup are wasteful.

**Phase to address:** Phase 14 (Database Backend).

---

### Integration Pitfall 3: Markdown Body Rendering Conflicts With Fixed-Width Pane Layout

**What goes wrong:** The todo list pane has a calculated width: `todoInnerWidth := m.width - calendarInnerWidth - (frameH * 2)`. Rendering markdown in this space using a library like Glamour produces output that may exceed this width (long lines, wide code blocks, table rendering). If the rendered markdown is wider than the pane, lipgloss layout breaks -- text wraps awkwardly, columns misalign, or content overflows into the calendar pane.

**Why it happens:** Glamour renders markdown to a specified width, but certain elements (code blocks, long URLs, pre-formatted text) may resist wrapping. The todo list pane is already the flexible-width pane (it absorbs whatever space the calendar doesn't use), but on narrow terminals (80 columns), the todo pane is only ~34 characters wide after the calendar takes its 38.

**How to avoid:**
- Pass the pane width to the markdown renderer: `glamour.RenderWithEnvironmentConfig(body, glamour.WithWordWrap(todoInnerWidth))`.
- For Phase 15 (markdown templates), the initial display in the todo LIST can remain single-line (title only). Only show the full rendered markdown body in a detail/preview view or when editing. This sidesteps the rendering-in-narrow-pane problem entirely for the list view.
- If a full markdown preview IS shown inline, truncate to a fixed number of lines (e.g., 3-line preview) with a "..." indicator.

**Warning signs:**
- Markdown rendered at full terminal width instead of pane width
- Long markdown lines cause the todo pane to overflow
- No width parameter passed to the markdown renderer
- Narrow terminal (80 cols) test shows broken layout

**Recovery cost:** LOW -- passing width to renderer is a one-parameter fix. But discovering it requires testing at various terminal sizes.

**Phase to address:** Phase 15 (Markdown Templates). Decide the display strategy during planning.

---

### Integration Pitfall 4: External Editor Workflow Does Not Refresh Store State

**What goes wrong:** The editor writes the modified markdown body to a temp file. After the editor exits, the app reads the temp file and calls `store.Update()`. But if the store is SQLite with query-on-read semantics, other parts of the app may have stale cached data. More subtly, if the todolist model has a `cursor` pointing at index 3 and the body update changes the todo's sort-relevant properties (it shouldn't, but might), the cursor points at the wrong item after refresh.

**Why it happens:** The `editorFinishedMsg` handler reads the temp file and updates the store, but the todo list's `visibleItems()` re-queries the store on every `View()` call (assuming SQLite query-on-read), so it should be fine. The real danger is with in-memory cache designs (Integration Pitfall 1) where the cache is stale after a direct database update.

**How to avoid:**
- If using query-on-read (recommended), the `View()` cycle after `editorFinishedMsg` automatically shows fresh data. No explicit refresh needed.
- Also call `m.calendar.RefreshIndicators()` after the editor update, as the existing pattern does for all store mutations (line 181 of `app/model.go`).
- The `editorFinishedMsg` handler should: (1) read temp file, (2) call store.UpdateBody(), (3) remove temp file, (4) set `m.editing = false`, (5) return nil cmd. Keep this sequence simple and synchronous.

**Warning signs:**
- Calendar indicators not refreshing after editor save
- Todo list showing old body content after editor save (cache not invalidated)
- Temp file not cleaned up on editor error

**Recovery cost:** LOW -- adding a refresh call is trivial. But stale data display is confusing to users.

**Phase to address:** Phase 16 (External Editor).

---

### Integration Pitfall 5: $EDITOR Environment Variable Handling Edge Cases

**What goes wrong:** The app reads `$EDITOR` to determine which editor to launch. But: (a) `$EDITOR` is empty on many systems (especially macOS default), (b) `$EDITOR` can contain arguments (e.g., `EDITOR="code --wait"`), (c) `$EDITOR` can point to a GUI editor that doesn't block the terminal (VS Code without `--wait`, Sublime Text). If the editor returns immediately (GUI editor without wait flag), the app reads the temp file before the user has finished editing -- saving empty or stale content.

**Why it happens:** `exec.Command(os.Getenv("EDITOR"), tmpFile)` works only if `$EDITOR` is a single executable name with no arguments. If it's `"vim"`, fine. If it's `"code --wait"`, `exec.Command` treats the entire string as the executable name and fails. If `$EDITOR` is unset and the app defaults to `"vi"`, it fails on systems without `vi` installed.

**How to avoid:**
- Check `$EDITOR`, then `$VISUAL`, then fall back to a sensible default (`"vi"` on Unix).
- Split `$EDITOR` on spaces to separate the command from arguments. Use the first token as the executable and append the rest as arguments before the filename:
  ```
  parts := strings.Fields(os.Getenv("EDITOR"))
  args := append(parts[1:], tmpFile)
  cmd := exec.Command(parts[0], args...)
  ```
- This handles `EDITOR="vim"`, `EDITOR="code --wait"`, and `EDITOR="nvim -u NONE"`.
- For the non-blocking GUI editor case: document that `$EDITOR` should be a terminal editor or a GUI editor with a wait flag. This is a user configuration issue, not a bug to solve. The same limitation applies to `git commit`, `crontab -e`, and every other tool that uses `$EDITOR`.

**Warning signs:**
- `exec.Command(editor, file)` where `editor` is the raw `$EDITOR` string
- No fallback when `$EDITOR` is empty
- Editor launches but app reads the file back immediately (before user saves)
- `$EDITOR` with spaces in the path (Windows-style paths) not handled

**Recovery cost:** LOW -- string splitting fix. But empty-$EDITOR crashes are bad first-run experience.

**Phase to address:** Phase 16 (External Editor).

---

## "Looks Done But Isn't" Patterns

These are bugs that pass initial manual testing but fail under real-world conditions.

### Pattern 1: SQLite Database Path Not Created

**What looks done:** SQLite store opens successfully in tests using `:memory:` or a path in the current directory.

**What's actually wrong:** In production, the database path is in `~/.config/todo-calendar/todos.db`. If the directory doesn't exist (first-time user), `sql.Open()` with `modernc.org/sqlite` does NOT create intermediate directories -- it creates the file but only if the parent directory exists. The current JSON store has `os.MkdirAll(dir, 0755)` in `Save()` (line 63 of `store.go`). The SQLite equivalent must also ensure the directory exists before opening.

**How to detect:** Test with a fresh user profile (empty `~/.config/`). The app should create `~/.config/todo-calendar/` and then `todos.db` inside it.

**Phase to address:** Phase 14.

---

### Pattern 2: Migration Succeeds But IDs Don't Match

**What looks done:** All todos appear in SQLite after migration. Counts match.

**What's actually wrong:** SQLite autoincrement starts at 1. The JSON store has `NextID` which may be 47 (after 46 todos, some deleted). If the migration uses autoincrement, new SQLite IDs are 1-N, not the original IDs. Any external references to todo IDs (none currently, but future features like markdown links `[see todo #12]`) would break. More immediately, the `editingID` field in the todolist model (line 58) stores the todo ID during editing -- if editing starts before migration and finishes after, the ID is wrong.

**How to detect:** After migration, verify that `SELECT id FROM todos` matches the original JSON IDs exactly. Verify that `Find(originalID)` returns the correct todo.

**Phase to address:** Phase 14. Use explicit ID values in INSERT, not autoincrement.

---

### Pattern 3: Editor Returns Error But File Was Actually Saved

**What looks done:** Error handling checks `editorFinishedMsg.err` and shows an error if non-nil.

**What's actually wrong:** Some editors return non-zero exit codes for valid reasons (vim returns 1 if the user quit with `:cq`, which means "quit without saving" -- but neovim may return 1 for certain plugin errors even if the file was saved). If the error handler skips reading the file on ANY error, legitimate saves are lost. Conversely, if the error handler always reads the file regardless of error, a `:cq` quit (intentional discard) saves content the user wanted to discard.

**How to detect:** Test with `:wq` (save and quit), `:q!` (quit without saving, but file was already written), and `:cq` (quit with error, file unchanged). The app should save the file content in all cases EXCEPT when the file content is unchanged (no user edits).

**Recommended approach:** Compare file content before and after editor. If content changed, save it -- regardless of exit code. If content is identical to what was written, skip the save. This handles both "normal save" and "intentional discard" correctly.

**Phase to address:** Phase 16.

---

### Pattern 4: Markdown Template Placeholders in User Content

**What looks done:** Templates with `{{date}}` and `{{title}}` placeholders work correctly.

**What's actually wrong:** If a user creates a todo with text containing `{{` or `}}` (literal curly braces), and that text is later processed through the template engine, Go's `text/template` attempts to parse it as a template action. This produces either parse errors or unexpected output. For example, a todo titled "Configure {{nginx}}" would fail template parsing.

**How to detect:** Create a todo with `{{` in its title or body, then apply any template operation that processes the text.

**How to avoid:** Template expansion should only happen at todo CREATION time (when filling in a template). Once a todo body is populated, it is plain markdown -- never re-processed through the template engine. Store the expanded result, not the template + data.

**Phase to address:** Phase 15.

---

### Pattern 5: Concurrent TUI State During Editor Execution

**What looks done:** Editor launches, user edits, editor closes, TUI resumes.

**What's actually wrong:** While the editor is running, the Bubble Tea program is "paused" but the terminal can still receive signals. If the user resizes the terminal while the editor is open, the `WindowSizeMsg` is received when the TUI resumes. If the model's `editing` flag prevents normal `View()` rendering but NOT normal `Update()` processing, the window size message updates `m.width` and `m.height` correctly. This is actually fine. BUT: if the editing flag also prevents `Update()` from processing, the resize message is lost and the TUI renders at the wrong size after resuming.

**How to detect:** Open editor, resize terminal while editor is open, close editor. The TUI should render at the new size.

**How to avoid:** The `editing` flag should only affect `View()` (return empty string). Do NOT block `Update()` processing -- let `WindowSizeMsg` and other messages update the model state normally.

**Phase to address:** Phase 16.

---

### Pattern 6: SQLite PRAGMA Settings Lost on Connection Reconnect

**What looks done:** Pragmas set after `sql.Open()` work correctly during the session.

**What's actually wrong:** Go's `database/sql` pool may close and reopen connections transparently. PRAGMAs like `journal_mode=WAL` are persistent (they survive reconnection) but `busy_timeout`, `foreign_keys`, and `synchronous` are per-connection settings. If the pool closes the idle connection and opens a new one, these pragmas revert to defaults. With `MaxOpenConns(1)` and `MaxIdleConns(1)` and `ConnMaxLifetime(0)`, the connection should stay alive forever -- but this depends on correct pool configuration.

**How to detect:** Set a long `ConnMaxLifetime` (e.g., 1 hour), wait past it, then perform an operation. Check if `foreign_keys` is still ON.

**How to avoid:**
- Use `db.Conn(ctx)` to get a dedicated connection and keep it for the app's lifetime. Or:
- Use a connection-init hook (`sql.Register` a driver wrapper that runs PRAGMAs on new connections). For `modernc.org/sqlite`, you can append pragmas to the DSN: `file:path/todos.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)`.
- The DSN approach is the simplest and most reliable.

**Phase to address:** Phase 14.

---

### Pattern 7: Body Field Displayed in Single-Line Todo List View

**What looks done:** Todo body field exists and can be edited via external editor.

**What's actually wrong:** The `renderTodo()` function in `todolist/model.go` (line 546) renders each todo as a single line: `[x] Todo text 2026-02-06`. After adding the `Body` field, if the body is naively included in this rendering, a multi-line markdown body explodes the single-line layout. Each todo could be 20+ lines, making the list unusable for navigation.

**How to detect:** Add a todo with a multi-line body. The todo list should remain navigable with single-line items.

**How to avoid:** The todo list renders ONLY the `Text` (title) field, never the `Body`. The body is shown in a detail view, a preview pane, or only when editing in the external editor. The list view should show a body indicator (e.g., a small icon or `[...]` suffix) to signal that a body exists, without rendering the body itself.

**Phase to address:** Phase 15 design. Decide the body display strategy before implementation.

---

## Phase-Specific Pitfall Summary

| Phase | Pitfall | Severity | Prevention |
|-------|---------|----------|------------|
| 14: Database | Store interface extraction (#1) | MEDIUM | Extract interface first, then swap backend |
| 14: Database | Migration data loss (#2) | HIGH | Transaction-wrap migration, keep JSON backup |
| 14: Database | Connection pool misconfiguration (#3) | MEDIUM | `MaxOpenConns(1)`, WAL mode, busy_timeout |
| 14: Database | Database path not created (Pattern 1) | LOW | `os.MkdirAll` before `sql.Open` |
| 14: Database | ID mismatch after migration (Pattern 2) | MEDIUM | Use explicit IDs, not autoincrement |
| 14: Database | Save() pattern doesn't map to SQL (Integration #1) | MEDIUM | Query-on-read, no in-memory cache |
| 14: Database | EnsureSortOrder on every startup (Integration #2) | LOW | One-time migration, not per-startup |
| 14: Database | Pragmas lost on reconnect (Pattern 6) | LOW | Set pragmas via DSN string |
| 14: Database | Body column in schema (Pitfall #6) | LOW | Include `body TEXT DEFAULT ''` in initial schema |
| 15: Templates | Template placeholders in user content (Pattern 4) | MEDIUM | Expand templates once at creation, store result |
| 15: Templates | Markdown rendering width (Integration #3) | LOW | Pass pane width to renderer; title-only in list view |
| 15: Templates | Body in list view (Pattern 7) | LOW | Render title only in list; body in detail/editor |
| 15: Templates | JSON backward compat (#6) | LOW | Enforce phase ordering; SQLite before bodies |
| 16: Editor | View content leak (#4) | MEDIUM | `editing` flag, return empty View |
| 16: Editor | Wrong temp file extension (#5) | LOW | Use `.md` extension pattern |
| 16: Editor | $EDITOR edge cases (Integration #5) | LOW | Split on spaces, check $VISUAL, fallback |
| 16: Editor | Store refresh after edit (Integration #4) | LOW | RefreshIndicators() in editorFinishedMsg handler |
| 16: Editor | Editor exit code ambiguity (Pattern 3) | LOW | Compare content before/after, ignore exit code |
| 16: Editor | Terminal resize during edit (Pattern 5) | LOW | Only guard View(), not Update() |

---

## Sources

**Codebase analysis (HIGH confidence -- primary source):**
- `internal/store/store.go` -- Save() pattern, mutation methods, EnsureSortOrder
- `internal/store/todo.go` -- Todo struct, Data envelope, JSON tags
- `internal/app/model.go` -- WithAltScreen usage, overlay pattern, message routing
- `internal/todolist/model.go` -- input state machine, renderTodo, cursor management
- `internal/config/config.go` -- XDG paths, atomic write pattern
- `main.go` -- program initialization, tea.NewProgram options

**Bubble Tea ExecProcess issues (HIGH confidence -- official repo):**
- [tea.ExecProcess writes View output to stdout (Issue #431)](https://github.com/charmbracelet/bubbletea/issues/431)
- [ExecProcess with WithAltScreen prints outside altscreen (Discussion #424)](https://github.com/charmbracelet/bubbletea/discussions/424)
- [Official exec example](https://github.com/charmbracelet/bubbletea/blob/main/examples/exec/main.go)
- [Bubble Tea pkg.go.dev - tea.Exec, tea.ExecProcess docs](https://pkg.go.dev/github.com/charmbracelet/bubbletea)

**Go SQLite best practices (HIGH confidence -- verified with multiple sources):**
- [Go + SQLite Best Practices (Jake Gold)](https://jacob.gold/posts/go-sqlite-best-practices/) -- connection pooling, WAL, pragmas
- [Resolve "database is locked" (Ben Boyter)](https://boyter.org/posts/go-sqlite-database-is-locked/) -- MaxOpenConns(1) recommendation
- [SQLite concurrent writes and "database is locked"](https://tenthousandmeters.com/blog/sqlite-concurrent-writes-and-database-is-locked-errors/) -- busy_timeout behavior, BEGIN IMMEDIATE
- [SQLITE_BUSY despite timeout (Bert Hubert)](https://berthub.eu/articles/posts/a-brief-post-on-sqlite3-database-locked-despite-timeout/) -- transaction upgrade pitfall
- [Go and SQLite: when database/sql chafes (David Crawshaw)](https://crawshaw.io/blog/go-and-sqlite) -- connection pool limitations

**SQLite driver selection (MEDIUM confidence -- benchmarks may be outdated):**
- [SQLite in Go, with and without cgo (multiprocess.io)](https://datastation.multiprocess.io/blog/2022-05-12-sqlite-in-go-with-and-without-cgo.html) -- modernc.org/sqlite vs mattn/go-sqlite3 performance
- [go-sqlite-bench (GitHub)](https://github.com/cvilsmeier/go-sqlite-bench) -- comparative benchmarks
- [modernc.org/sqlite with Go](https://theitsolutions.io/blog/modernc.org-sqlite-with-go) -- DSN pragma syntax

**Glamour markdown rendering (MEDIUM confidence):**
- [Glamour GitHub](https://github.com/charmbracelet/glamour) -- WordWrap option, width configuration

**Vim/Neovim filetype detection (HIGH confidence -- official docs):**
- [Neovim filetype detection docs](https://neovim.io/doc/user/filetype.html) -- extension-based detection for `.md`

---

*Pitfalls research for: TUI Calendar v1.4 Data & Editing*
*Researched: 2026-02-06*
