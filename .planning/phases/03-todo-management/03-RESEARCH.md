# Phase 3: Todo Management - Research

**Researched:** 2026-02-05
**Domain:** Todo CRUD, JSON persistence with atomic writes, Bubble Tea text input/list components, context-sensitive help bar
**Confidence:** HIGH

## Summary

Phase 3 transforms the placeholder todo list pane into a fully functional todo management system with add/complete/delete operations, optional date assignment, per-month and floating-section display, JSON persistence to disk, and a context-sensitive help bar. This is the final phase and delivers the core value proposition of the app.

The implementation requires four new concerns layered onto the existing Bubble Tea scaffold: (1) a todo data model with JSON serialization and atomic file persistence, (2) a text input mode for adding todos using `bubbles/textinput`, (3) a list rendering and navigation UI (custom, not `bubbles/list` -- see rationale below), and (4) a `bubbles/help` model replacing the current plain-string status bar with context-sensitive keybindings. The biggest architectural challenge is coordinating the calendar pane's viewed month with the todo pane's filter -- this requires the root model to propagate the calendar's current year/month to the todo model.

The project already made key decisions: use atomic writes (write-temp-rename pattern), store data as JSON in XDG-compliant paths (`~/.config/todo-calendar/`), and use `help.Model` from Bubbles for the help bar. No new external dependencies beyond the existing stack are needed -- `encoding/json` and `os` from the stdlib plus the existing Bubbles components cover everything. For atomic writes, a simple stdlib implementation (write to temp file, rename) is sufficient and avoids adding a new dependency.

**Primary recommendation:** Build a custom todo list renderer (not `bubbles/list`) for simplicity, use `bubbles/textinput` for the add-todo input, use `bubbles/help` for the context-sensitive help bar, persist todos as indented JSON via stdlib atomic write-temp-rename, and synchronize calendar month with todo filtering via a message from root model.

## Standard Stack

### Core (Phase 3 additions -- no new Go modules needed)

| Library/Package | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `encoding/json` | stdlib | Todo serialization/deserialization | Built-in, stable, sufficient for simple JSON; no need for v2 experimental |
| `bubbles/textinput` | v0.21.1 (already imported) | Inline text input for adding todos | Official Charm component; Focus/Blur/Value API; handles cursor, paste, key bindings |
| `bubbles/help` | v0.21.1 (already imported) | Context-sensitive help bar | Official Charm component; auto-truncates to terminal width; uses KeyMap interface already on our key structs |
| `bubbles/key` | v0.21.1 (already imported) | Dynamic enable/disable of keybindings | `SetEnabled(bool)` hides bindings from help and disables matching; already used throughout |
| `os` | stdlib | File operations, `os.UserConfigDir()`, `os.MkdirAll`, `os.CreateTemp`, `os.Rename` | Stdlib atomic write pattern; XDG path resolution already in use |

### Existing (from Phase 1 & 2, unchanged)

| Library | Version | Purpose |
|---------|---------|---------|
| Bubble Tea | v1.3.10 | TUI framework |
| Lip Gloss | v1.1.0 | Terminal styling |
| Bubbles | v0.21.1 | `key.Binding`, `textinput.Model`, `help.Model` |
| `BurntSushi/toml` | v1.6.0 | Config file |
| `rickar/cal/v2` | v2.1.27 | Holiday data |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom list rendering | `bubbles/list` | `bubbles/list` is a full-featured component with filtering, pagination, spinner, title bar, status bar -- massive overkill for a simple todo list with ~20 items. It also fights with our existing pane layout (its own help, its own sizing). Custom rendering is 50 lines of View() code. |
| Stdlib atomic write | `google/renameio/v2` | renameio adds a dependency for what is 10 lines of stdlib code (`os.CreateTemp` + `os.Rename`). For a single JSON file, stdlib is sufficient. |
| Stdlib atomic write | `natefinch/atomic` | Same reasoning; unnecessary dependency for simple use case |
| `encoding/json` | `encoding/json/v2` | v2 is experimental (requires `GOEXPERIMENT=jsonv2`), not subject to Go 1 compat promise |
| Integer auto-increment IDs | UUID library (`google/uuid`) | Auto-increment counter stored in the JSON file is simpler and sufficient for a single-user local app. No need for UUIDs. |
| `~/.config/` for data | `~/.local/share/` via `adrg/xdg` | The requirement explicitly says `~/.config/todo-calendar/`. Go stdlib has no `os.UserDataDir()`. Adding `adrg/xdg` for strict XDG_DATA_HOME compliance adds a dependency when the requirement already specifies the path. Store both config and data under `~/.config/todo-calendar/`. |

**No new `go get` needed.** All required functionality comes from existing dependencies and the Go standard library.

## Architecture Patterns

### Recommended Project Structure (Phase 3 additions)

```
internal/
├── app/
│   ├── model.go         # Add: month sync between calendar and todo, help.Model
│   ├── keys.go          # Add: new bindings (add, delete, complete, help toggle)
│   └── styles.go        # (minimal changes)
├── calendar/
│   ├── model.go         # Add: exported Year()/Month() or ViewedMonth message
│   └── ...
├── todolist/
│   ├── model.go         # REWRITE: full todo list UI with CRUD, modes, rendering
│   ├── keys.go          # NEW: todo-specific key bindings
│   └── styles.go        # NEW: todo-specific styles (complete, section headers)
└── store/
    ├── store.go         # NEW: JSON persistence with atomic writes
    └── todo.go          # NEW: Todo data model (struct + serialization)
```

### Pattern 1: Todo Data Model with JSON Serialization

**What:** A `Todo` struct with ID, text, optional date, and completion status. A `Store` manages the slice of todos, loading from and saving to a JSON file. IDs are auto-incrementing integers tracked by a `NextID` counter in the JSON.

**When to use:** All todo data management.

**Confidence:** HIGH (standard Go JSON patterns, verified `encoding/json` API)

**Example:**

```go
// internal/store/todo.go

package store

import "time"

// Todo represents a single todo item.
type Todo struct {
    ID        int        `json:"id"`
    Text      string     `json:"text"`
    Date      *time.Time `json:"date,omitempty"` // nil = floating (undated)
    Done      bool       `json:"done"`
    CreatedAt time.Time  `json:"created_at"`
}

// Data is the top-level JSON structure persisted to disk.
type Data struct {
    NextID int    `json:"next_id"`
    Todos  []Todo `json:"todos"`
}
```

### Pattern 2: Atomic JSON Persistence (Stdlib)

**What:** Write JSON to a temp file in the same directory, then atomically rename to the target path. This prevents data loss if the process is interrupted during a write. Use `json.MarshalIndent` for human-readable output.

**When to use:** Every mutation (add, complete, delete) triggers a save.

**Confidence:** HIGH (verified `os.CreateTemp`, `os.Rename` in Go stdlib; prior decision mandates this pattern)

**Example:**

```go
// internal/store/store.go

package store

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// Store manages todo persistence.
type Store struct {
    path string
    data Data
}

// NewStore creates a store that reads/writes from the given path.
func NewStore(path string) (*Store, error) {
    s := &Store{path: path}
    if err := s.load(); err != nil {
        return nil, err
    }
    return s, nil
}

// Save writes the current data to disk atomically.
func (s *Store) Save() error {
    b, err := json.MarshalIndent(s.data, "", "  ")
    if err != nil {
        return err
    }

    dir := filepath.Dir(s.path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Write to temp file in same directory (same filesystem for atomic rename)
    tmp, err := os.CreateTemp(dir, "todos-*.json.tmp")
    if err != nil {
        return err
    }
    tmpName := tmp.Name()

    if _, err := tmp.Write(b); err != nil {
        tmp.Close()
        os.Remove(tmpName)
        return err
    }
    if err := tmp.Sync(); err != nil {
        tmp.Close()
        os.Remove(tmpName)
        return err
    }
    if err := tmp.Close(); err != nil {
        os.Remove(tmpName)
        return err
    }

    return os.Rename(tmpName, s.path)
}

// load reads the JSON file, creating empty data if file doesn't exist.
func (s *Store) load() error {
    b, err := os.ReadFile(s.path)
    if os.IsNotExist(err) {
        s.data = Data{NextID: 1, Todos: []Todo{}}
        return nil
    }
    if err != nil {
        return err
    }
    return json.Unmarshal(b, &s.data)
}
```

### Pattern 3: Mode-Based Input Handling

**What:** The todo list model has two modes: `normalMode` (navigating, completing, deleting) and `inputMode` (typing a new todo). In `normalMode`, j/k navigate the cursor, x toggles completion, d deletes, and a/A opens the text input. In `inputMode`, the `textinput.Model` receives all keystrokes except Escape (cancel) and Enter (confirm add). The help bar shows different bindings per mode.

**When to use:** Always for the todo pane.

**Confidence:** HIGH (standard Bubble Tea pattern; verified textinput Focus/Blur API; verified key.SetEnabled for dynamic help)

**Example:**

```go
// internal/todolist/model.go

type mode int

const (
    normalMode mode = iota
    inputMode
)

type Model struct {
    focused  bool
    width    int
    height   int
    mode     mode
    cursor   int           // selected todo index
    input    textinput.Model
    store    *store.Store
    viewYear  int
    viewMonth time.Month
    keys     KeyMap
}
```

### Pattern 4: Calendar-Todo Month Synchronization

**What:** When the calendar navigates to a new month, the root model must inform the todo pane so it filters todos for the new month. Two approaches: (A) the root model reads `calendar.Year()` and `calendar.Month()` after each calendar update and calls `todoList.SetViewMonth(year, month)`, or (B) the calendar emits a custom `MonthChangedMsg` that the root forwards to the todo pane.

**Recommended:** Approach A (direct setter) is simpler and consistent with the existing `SetFocused` pattern. The calendar model needs two new exported methods: `Year() int` and `Month() time.Month`.

**When to use:** After every calendar Update in the root model.

**Confidence:** HIGH (consistent with existing patterns in the codebase)

**Example:**

```go
// internal/calendar/model.go -- add:
func (m Model) Year() int        { return m.year }
func (m Model) Month() time.Month { return m.month }

// internal/app/model.go -- in Update, after calendar processes a key:
case calendarPane:
    m.calendar, cmd = m.calendar.Update(msg)
    // Sync viewed month to todo list
    m.todoList.SetViewMonth(m.calendar.Year(), m.calendar.Month())
```

### Pattern 5: Context-Sensitive Help Bar with help.Model

**What:** Replace the current plain string status bar with `help.Model` from Bubbles. The help bar renders different bindings based on which pane is focused and which mode is active. Use `key.Binding.SetEnabled(bool)` to dynamically show/hide bindings.

**When to use:** Always for the bottom status bar.

**Confidence:** HIGH (verified help.Model API: `View(KeyMap)`, `Width` field, `ShortHelpView`; verified `key.SetEnabled`)

**Example:**

```go
// internal/app/model.go

type Model struct {
    // ...existing fields...
    help help.Model
}

func New(...) Model {
    h := help.New()
    // ...
    return Model{
        // ...
        help: h,
    }
}

// Compose a dynamic KeyMap that aggregates bindings from active context
func (m Model) helpKeys() help.KeyMap {
    // Return different key sets based on activePane and todoList mode
    // Disabled bindings are automatically hidden by help.Model
}

func (m Model) View() string {
    // ...
    m.help.Width = m.width
    helpBar := m.help.View(m.helpKeys())
    return lipgloss.JoinVertical(lipgloss.Left, top, helpBar)
}
```

### Pattern 6: Two-Section Todo Rendering

**What:** The todo pane View renders two sections: (1) dated todos for the currently viewed month, with a header like "February 2026", and (2) floating (undated) todos under a "Floating" or "No Date" header. Each section shows a cursor indicator for the selected item. If either section is empty, show a subtle placeholder message.

**When to use:** Always for the todo pane View.

**Confidence:** HIGH (pure rendering logic, follows established format-before-style pattern)

**Example structure:**

```
  February 2026
  ──────────────────
  [x] Pay rent (Feb 1)
> [ ] Submit report (Feb 15)
  [ ] Dentist appointment (Feb 20)

  Floating
  ──────────────────
  [ ] Read that book
  [ ] Fix the leaky faucet
```

### Anti-Patterns to Avoid

- **Using `bubbles/list` for the todo display:** It manages its own help, title, pagination, filtering, status bar, and size -- all things we already handle at the app level. It would fight with the pane layout and double-render help.
- **Storing `time.Time` with timezone in JSON:** JSON has no timezone type. Use `time.Time` with a custom date-only format (`"2006-01-02"`) for the optional date field to avoid timezone deserialization issues.
- **Saving on every keystroke during input:** Only save when a mutation completes (Enter confirms add, x toggles complete, d deletes). Never save during typing.
- **Blocking the UI on file I/O:** For a single small JSON file, synchronous writes are fine (sub-millisecond). Do NOT use `tea.Cmd` for async I/O here -- it adds complexity for no benefit.
- **Monolithic KeyMap:** Keep separate KeyMaps per component (app, calendar, todolist) and compose them in the help bar. Do not merge all bindings into one giant struct.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Text input with cursor, paste, editing | Custom rune-by-rune key handling | `bubbles/textinput` | Handles cursor positioning, word-boundary movement, paste, char limit, placeholder, focus/blur -- hundreds of edge cases |
| Help bar that truncates to terminal width | Manual string truncation | `bubbles/help` `ShortHelpView` | Gracefully truncates bindings that don't fit, shows ellipsis, auto-hides disabled bindings |
| Dynamic keybinding visibility | Manual if/else in View() for help text | `key.Binding.SetEnabled(bool)` + `help.Model` | Disabled bindings automatically excluded from help.View output |
| XDG config directory resolution | Hardcoded `$HOME/.config` | `os.UserConfigDir()` | Already used in `config/paths.go`; respects `$XDG_CONFIG_HOME` |
| JSON indentation | Manual string formatting | `json.MarshalIndent(v, "", "  ")` | Handles all nesting, escaping, and formatting automatically |

**Key insight:** Phase 3 needs zero new Go module dependencies. Everything is covered by `encoding/json` (stdlib), `os` (stdlib), and the existing Bubbles components (`textinput`, `help`, `key`).

## Common Pitfalls

### Pitfall 1: Time Zone Corruption in JSON Date Serialization

**What goes wrong:** `json.Marshal` serializes `time.Time` as RFC 3339 with timezone offset (e.g., `"2026-02-15T00:00:00+02:00"`). When deserialized on a machine with a different timezone, the date can shift by a day.
**Why it happens:** Todos have an optional *date*, not a datetime. But `time.Time` always carries timezone info.
**How to avoid:** Use a custom JSON marshaler that serializes dates as `"2006-01-02"` strings, or store the date as a string field and parse it when needed. A simple approach: store `Date string` (format `"YYYY-MM-DD"`) in the JSON struct and convert to `time.Time` only for filtering.
**Warning signs:** Todos appear on wrong days after moving between timezones or DST changes.
**Confidence:** HIGH (fundamental JSON/timezone behavior)

### Pitfall 2: Text Input Swallowing Global Keys

**What goes wrong:** When `textinput.Model` is focused, it consumes *all* key messages including `q` (quit), `Tab` (switch pane), and arrow keys. The app appears frozen or ignores global commands.
**Why it happens:** The textinput model processes any printable character as input. If you route messages to it before checking global keys, those keys get eaten.
**How to avoid:** In `inputMode`, intercept Escape and Enter at the todo model level BEFORE forwarding to textinput. At the root level, do NOT forward `q` to the todo model when it is in input mode -- or better, disable the `q`-to-quit binding when inputting text (only `Ctrl+C` and `Escape` exit). The root model should check `todoList.IsInputting()` before processing global key bindings.
**Warning signs:** Pressing `q` types "q" into the input instead of quitting.
**Confidence:** HIGH (fundamental Bubble Tea message routing behavior)

### Pitfall 3: Cursor Index Out of Bounds After Delete

**What goes wrong:** User deletes the last item in the list. The cursor index still points to the old position, which is now past the end. Next render panics with index out of range.
**Why it happens:** Delete removes an item but cursor position is not adjusted.
**How to avoid:** After any delete, clamp the cursor: `if m.cursor >= len(items) { m.cursor = len(items) - 1 }`. Also guard against `m.cursor < 0` (empty list).
**Warning signs:** Panic after deleting the last todo in a section.
**Confidence:** HIGH (classic off-by-one in list management)

### Pitfall 4: Empty Data File vs Missing File

**What goes wrong:** An empty file (`""`) fails `json.Unmarshal` with a syntax error. A missing file should return defaults. A corrupted file should return an error. These three cases need distinct handling.
**Why it happens:** `os.ReadFile` on an empty file returns `[]byte("")`, and `json.Unmarshal([]byte(""), &v)` returns `unexpected end of JSON input`.
**How to avoid:** Check `os.IsNotExist(err)` for missing file (return defaults). Check `len(b) == 0` for empty file (treat as missing, return defaults). Only attempt unmarshal for non-empty content.
**Warning signs:** App crashes on first run or after manual file editing leaves it empty.
**Confidence:** HIGH (verified json.Unmarshal behavior with empty input)

### Pitfall 5: Losing Data When Two Instances Run

**What goes wrong:** If two instances of the app run simultaneously, they each load the JSON, make changes, and overwrite each other's changes. Last write wins.
**Why it happens:** No file locking. Atomic writes prevent corruption but not lost updates.
**How to avoid:** For a single-user personal tool, this is acceptable. Document that concurrent instances are not supported. Optionally, add a warning if another instance is detected (e.g., check for `.lock` file). Do NOT implement complex file locking -- it's out of scope per project constraints.
**Warning signs:** Todos disappear after closing the app when another instance was open.
**Confidence:** MEDIUM (design tradeoff, not a bug per se)

### Pitfall 6: help.Model Width Not Set

**What goes wrong:** `help.Model` renders at zero width, truncating all bindings to nothing. The help bar appears empty.
**Why it happens:** `help.New()` initializes with `Width: 0`. You must set `m.help.Width` in the WindowSizeMsg handler.
**How to avoid:** In the root model's `WindowSizeMsg` handler, always set `m.help.Width = msg.Width`.
**Warning signs:** Help bar is blank or shows only "...".
**Confidence:** HIGH (verified help.Model struct -- Width field defaults to 0, ShortHelpView truncates based on Width)

## Code Examples

### Todo Data Model with Date-Only Serialization

```go
// internal/store/todo.go
// Source: encoding/json stdlib + custom marshaler pattern

package store

import (
    "encoding/json"
    "fmt"
    "time"
)

const dateFormat = "2006-01-02"

// Todo represents a single todo item.
type Todo struct {
    ID        int    `json:"id"`
    Text      string `json:"text"`
    Date      string `json:"date,omitempty"` // "YYYY-MM-DD" or "" for floating
    Done      bool   `json:"done"`
    CreatedAt string `json:"created_at"`
}

// HasDate returns true if the todo has a date assigned.
func (t Todo) HasDate() bool {
    return t.Date != ""
}

// ParseDate returns the parsed date. Returns zero time if no date.
func (t Todo) ParseDate() (time.Time, error) {
    if t.Date == "" {
        return time.Time{}, nil
    }
    return time.Parse(dateFormat, t.Date)
}

// InMonth returns true if the todo's date falls in the given year/month.
func (t Todo) InMonth(year int, month time.Month) bool {
    if t.Date == "" {
        return false
    }
    d, err := time.Parse(dateFormat, t.Date)
    if err != nil {
        return false
    }
    return d.Year() == year && d.Month() == month
}

// Data is the top-level JSON structure.
type Data struct {
    NextID int    `json:"next_id"`
    Todos  []Todo `json:"todos"`
}
```

### Atomic File Write (Stdlib)

```go
// internal/store/store.go
// Source: os.CreateTemp + os.Rename pattern (Go stdlib)

func (s *Store) Save() error {
    b, err := json.MarshalIndent(s.data, "", "  ")
    if err != nil {
        return err
    }
    // Append newline for POSIX compliance
    b = append(b, '\n')

    dir := filepath.Dir(s.path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    tmp, err := os.CreateTemp(dir, ".todos-*.tmp")
    if err != nil {
        return err
    }
    tmpName := tmp.Name()

    // Write, sync, close, rename -- any failure cleans up
    if _, err := tmp.Write(b); err != nil {
        tmp.Close()
        os.Remove(tmpName)
        return err
    }
    if err := tmp.Sync(); err != nil {
        tmp.Close()
        os.Remove(tmpName)
        return err
    }
    if err := tmp.Close(); err != nil {
        os.Remove(tmpName)
        return err
    }
    return os.Rename(tmpName, s.path)
}
```

### Text Input Integration for Adding Todos

```go
// internal/todolist/model.go (relevant excerpt)
// Source: bubbles/textinput v0.21.1 pkg.go.dev

import "github.com/charmbracelet/bubbles/textinput"

func New(store *store.Store) Model {
    ti := textinput.New()
    ti.Placeholder = "Buy groceries"
    ti.CharLimit = 120
    ti.Prompt = "> "
    // Do NOT call ti.Focus() here -- focus on demand when entering input mode

    return Model{
        input: ti,
        store: store,
        mode:  normalMode,
        keys:  DefaultKeyMap(),
    }
}

// Entering input mode:
func (m *Model) enterInputMode() tea.Cmd {
    m.mode = inputMode
    m.input.Reset()
    return m.input.Focus() // returns blink command
}

// Confirming input:
func (m *Model) confirmInput() {
    text := strings.TrimSpace(m.input.Value())
    if text == "" {
        return
    }
    m.store.Add(text, m.pendingDate) // pendingDate may be "" for floating
    m.input.Blur()
    m.mode = normalMode
}

// In Update:
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if m.mode == inputMode {
        switch msg := msg.(type) {
        case tea.KeyMsg:
            switch msg.String() {
            case "enter":
                m.confirmInput()
                return m, nil
            case "esc":
                m.input.Blur()
                m.mode = normalMode
                return m, nil
            }
        }
        // Forward everything else to textinput
        var cmd tea.Cmd
        m.input, cmd = m.input.Update(msg)
        return m, cmd
    }

    // Normal mode key handling...
}
```

### Context-Sensitive Help Bar

```go
// internal/app/model.go (help integration)
// Source: bubbles/help v0.21.1 pkg.go.dev

import "github.com/charmbracelet/bubbles/help"

type Model struct {
    calendar   calendar.Model
    todoList   todolist.Model
    activePane pane
    width      int
    height     int
    ready      bool
    keys       KeyMap
    help       help.Model
}

// helpKeyMap aggregates bindings for the current context.
// This satisfies the help.KeyMap interface.
type helpKeyMap struct {
    bindings []key.Binding
}

func (h helpKeyMap) ShortHelp() []key.Binding { return h.bindings }
func (h helpKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{h.bindings} }

func (m Model) currentHelpKeys() helpKeyMap {
    var bindings []key.Binding

    switch m.activePane {
    case calendarPane:
        bindings = append(bindings,
            m.calendar.Keys().PrevMonth,
            m.calendar.Keys().NextMonth,
        )
    case todoPane:
        bindings = append(bindings, m.todoList.HelpBindings()...)
    }
    bindings = append(bindings, m.keys.Tab, m.keys.Quit)
    return helpKeyMap{bindings: bindings}
}

func (m Model) View() string {
    // ... pane rendering ...
    m.help.Width = m.width
    helpBar := m.help.View(m.currentHelpKeys())
    return lipgloss.JoinVertical(lipgloss.Left, top, helpBar)
}
```

### Data File Path Resolution

```go
// internal/store/paths.go (or extend config/paths.go)
// Source: os.UserConfigDir() -- already used for config

import (
    "os"
    "path/filepath"
)

// TodosPath returns the path to the todos JSON file.
func TodosPath() (string, error) {
    dir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, "todo-calendar", "todos.json"), nil
}
```

### JSON File Format

```json
{
  "next_id": 4,
  "todos": [
    {
      "id": 1,
      "text": "Pay rent",
      "date": "2026-02-01",
      "done": true,
      "created_at": "2026-01-28"
    },
    {
      "id": 2,
      "text": "Submit report",
      "date": "2026-02-15",
      "done": false,
      "created_at": "2026-02-01"
    },
    {
      "id": 3,
      "text": "Read that book",
      "done": false,
      "created_at": "2026-02-03"
    }
  ]
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `ioutil.WriteFile` | `os.WriteFile` or atomic write pattern | Go 1.16 | `ioutil` deprecated; use `os` package |
| `encoding/json` v1 | `encoding/json` v1 (v2 is experimental) | Go 1.25+ | v2 exists behind `GOEXPERIMENT=jsonv2` but not stable; use v1 |
| `time.Time` in JSON | Date-only string `"2006-01-02"` for dates | N/A | Avoids timezone corruption in todo dates |
| Plain string help bar | `bubbles/help` with `help.Model` | Bubbles v0.20+ | Auto-truncation, style support, KeyMap interface |

**Deprecated/outdated:**
- `ioutil.ReadFile` / `ioutil.WriteFile`: Use `os.ReadFile` / `os.WriteFile` (Go 1.16+)
- `bubbles/textinput` `CursorStyle` field: Deprecated in v0.21.1; use `Cursor cursor.Model` instead
- `bubbles/textinput` `BlinkSpeed` field: Deprecated in v0.21.1; configure via `Cursor` model instead

## Open Questions

1. **Date input format for new todos**
   - What we know: Users need to optionally attach a date when adding a todo. The simplest UX is a single text input line.
   - What's unclear: Should the date be entered in the same input as the text (e.g., "Buy groceries @2026-02-15"), parsed separately after text, or via a separate prompt?
   - Recommendation: Use a two-step approach: (1) press `a` to add a floating todo (text only), or (2) press `A` to add a dated todo -- first enter text, then a date prompt with the current viewed month's first day as placeholder (e.g., "2026-02-01"). This keeps the input simple and avoids complex parsing. Alternatively, a single-line `@YYYY-MM-DD` suffix parser is also viable for power users.

2. **Cursor navigation across sections**
   - What we know: The todo pane has two sections (dated, floating). The cursor must move across both.
   - What's unclear: Should the cursor be a single unified index across both sections, or two independent cursors (one per section)?
   - Recommendation: Single unified cursor over a combined "visible items" slice. Sections are visual headers, not separate navigable entities. When rendering, insert non-selectable header rows. The cursor skips headers automatically.

3. **Sorting of dated todos**
   - What we know: Dated todos should show for the viewed month.
   - What's unclear: Sort order within the month section.
   - Recommendation: Sort dated todos by date ascending, then by ID ascending for same-date items. Floating todos sorted by creation order (ID ascending). This provides a natural chronological view.

4. **What happens to completed todos?**
   - What we know: Users can mark todos as complete. No editing requirement.
   - What's unclear: Do completed todos stay visible forever? Are they hidden? Is there a purge mechanism?
   - Recommendation: Keep completed todos visible (with strikethrough/checkmark styling) in the same list position. They serve as a visual record of accomplishment. A future enhancement could add a "clear completed" action, but this is not required for v1.

## Sources

### Primary (HIGH confidence)
- [bubbles/textinput v0.21.1 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/textinput) -- Model struct, Focus/Blur/Value API, KeyMap, Placeholder, CharLimit verified
- [bubbles/help v0.21.1 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/help) -- Model struct, View(KeyMap), ShortHelpView, Width, Styles, KeyMap interface verified
- [bubbles/key v0.21.1 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/key) -- SetEnabled, Enabled, NewBinding, Matches verified
- [bubbles/list v0.21.1 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/list) -- Full API reviewed; evaluated and rejected for this use case
- [encoding/json pkg.go.dev](https://pkg.go.dev/encoding/json) -- Marshal, MarshalIndent, Unmarshal, struct tags, omitempty
- [encoding/json/v2 pkg.go.dev](https://pkg.go.dev/encoding/json/v2) -- Confirmed experimental, requires GOEXPERIMENT flag, not for production
- [os pkg.go.dev](https://pkg.go.dev/os) -- CreateTemp, Rename, UserConfigDir, MkdirAll, ReadFile verified
- [google/renameio/v2 pkg.go.dev](https://pkg.go.dev/github.com/google/renameio/v2) -- WriteFile, TempFile, PendingFile API reviewed; evaluated as unnecessary dependency
- [bubbletea help example](https://github.com/charmbracelet/bubbletea/blob/main/examples/help/main.go) -- KeyMap integration pattern, help.Width setup, ShowAll toggle
- [bubbletea textinput example](https://github.com/charmbracelet/bubbletea/blob/main/examples/textinput/main.go) -- textinput.Model integration, Focus in init, Update delegation
- [bubbletea composable-views example](https://github.com/charmbracelet/bubbletea/blob/main/examples/composable-views/main.go) -- Mode switching, message routing between components

### Secondary (MEDIUM confidence)
- [adrg/xdg GitHub](https://github.com/adrg/xdg) -- XDG_DATA_HOME support reviewed; decided not to use (requirement specifies ~/.config/ path)
- [Go issue #62382](https://github.com/golang/go/issues/62382) -- os.UserDataDir proposal; confirmed not yet in stdlib
- [Atomic file writing in Go](https://michael.stapelberg.ch/posts/2017-01-28-golang_atomically_writing/) -- Write-temp-rename pattern analysis

### Tertiary (LOW confidence)
- None -- all findings verified with primary or secondary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- zero new dependencies; all APIs verified against pkg.go.dev for exact versions
- Architecture: HIGH -- patterns follow established codebase conventions (Phase 1/2); textinput and help APIs verified
- Pitfalls: HIGH (5/6), MEDIUM (1/6) -- concurrent instance pitfall is a design tradeoff, not a verified bug
- Code examples: HIGH -- all API calls verified against pkg.go.dev documentation for specific library versions

**Research date:** 2026-02-05
**Valid until:** 2026-03-07 (30 days -- stable libraries, no imminent breaking changes)
