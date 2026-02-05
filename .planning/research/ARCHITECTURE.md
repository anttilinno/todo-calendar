# Architecture Research

**Domain:** TUI calendar + todo application (Go / Bubble Tea)
**Researched:** 2026-02-05
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Terminal (Alt Screen)                        │
├─────────────────────────────────────────────────────────────────┤
│                        Root Model                               │
│  ┌───────────────────────┐   ┌───────────────────────────────┐  │
│  │   Calendar Component  │   │     Todo List Component       │  │
│  │                       │   │                               │  │
│  │  ┌─────────────────┐  │   │  ┌─────────────────────────┐  │  │
│  │  │  Month Grid      │  │   │  │  Date-Bound Todos       │  │  │
│  │  │  (custom render) │  │   │  │  (list/viewport)        │  │  │
│  │  └─────────────────┘  │   │  ├─────────────────────────┤  │  │
│  │  ┌─────────────────┐  │   │  │  Floating Todos          │  │  │
│  │  │  Nav Controls    │  │   │  │  (list/viewport)        │  │  │
│  │  │  (prev/next mo.) │  │   │  └─────────────────────────┘  │  │
│  │  └─────────────────┘  │   │                               │  │
│  └───────────────────────┘   └───────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                     Status / Help Bar                           │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  help.Model (key bindings display)                          ││
│  └─────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                      Data Layer                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │  Todo Store   │  │  Holiday     │  │  Config              │  │
│  │  (JSON file)  │  │  Provider    │  │  (YAML/TOML file)    │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| Root Model | Owns all state, routes messages to focused child, composes layout with lipgloss.JoinHorizontal | Single struct embedding calendar + todo models, tracks focus/active pane |
| Calendar Component | Renders month grid, handles date navigation (prev/next month, cursor movement), highlights today and holidays | Custom model with month/year state, no off-the-shelf Bubbles calendar exists |
| Todo List Component | Displays todos for selected month, separates date-bound from floating, handles add/check/delete | Could use bubbles/list or a custom viewport-based list |
| Help Bar | Shows context-sensitive key bindings for currently focused pane | bubbles/help.Model auto-generates from key.Binding definitions |
| Todo Store | Reads/writes todos to local JSON file, CRUD operations | Plain Go struct with file I/O methods, called via tea.Cmd |
| Holiday Provider | Returns holidays for a given country + year | Go package wrapping a holiday data source or embedded data |
| Config | Loads user preferences (country, file paths, theme) | Parsed at startup from YAML/TOML, passed into model |

## Recommended Project Structure

```
todo-calendar/
├── main.go                  # Entry point: creates tea.Program, runs it
├── go.mod
├── go.sum
├── internal/
│   ├── app/
│   │   ├── model.go         # Root model: state, Init, Update, View
│   │   ├── keys.go          # Global key bindings (quit, tab, help toggle)
│   │   ├── styles.go        # Lipgloss styles (borders, colors, focus indicators)
│   │   └── messages.go      # App-level custom messages (TodoSavedMsg, etc.)
│   ├── calendar/
│   │   ├── model.go         # Calendar model: month state, cursor, Init/Update/View
│   │   ├── grid.go          # Month grid rendering logic (pure View helper)
│   │   ├── keys.go          # Calendar-specific key bindings
│   │   └── messages.go      # Calendar messages (DateSelectedMsg, MonthChangedMsg)
│   ├── todolist/
│   │   ├── model.go         # Todo list model: items, cursor, Init/Update/View
│   │   ├── keys.go          # Todo-specific key bindings (add, check, delete)
│   │   └── messages.go      # Todo messages (TodoAddedMsg, TodoToggledMsg)
│   ├── store/
│   │   ├── store.go         # Todo persistence: Load, Save, Add, Toggle, Delete
│   │   ├── store_test.go    # Unit tests for store operations
│   │   └── types.go         # Todo struct, serialization tags
│   ├── holiday/
│   │   ├── provider.go      # Holiday lookup by country + date
│   │   ├── provider_test.go # Tests for holiday data
│   │   └── data.go          # Embedded or generated holiday data
│   └── config/
│       ├── config.go        # Config loading, defaults, validation
│       └── types.go         # Config struct definition
└── .planning/               # Project planning (not shipped)
```

### Structure Rationale

- **internal/app/:** The root Bubble Tea model lives here. It is the only model that `main.go` knows about. Separating `keys.go`, `styles.go`, and `messages.go` from `model.go` keeps the root model file focused on Init/Update/View logic rather than growing into a monolith.
- **internal/calendar/:** Isolated component with its own Model, Update, View. The calendar is the most complex rendering piece (month grid with day numbers, week alignment, holiday coloring). Keeping `grid.go` separate from `model.go` separates pure rendering from state management.
- **internal/todolist/:** Isolated component. Could wrap a bubbles/list or be fully custom. Has its own key bindings that only apply when this pane is focused.
- **internal/store/:** Pure data layer with no Bubble Tea dependency. This can be tested independently with standard Go unit tests. All file I/O happens here, invoked through tea.Cmd wrappers in the app or todolist packages.
- **internal/holiday/:** Pure data layer. No UI dependency. Returns `[]Holiday` for a given month+year. Can be swapped between embedded data, an API call, or a generated data file.
- **internal/config/:** Loaded once at startup. Passed as a value into the root model constructor. Not a Bubble Tea model itself.

## Architectural Patterns

### Pattern 1: Elm Architecture (Model-Update-View)

**What:** Every component implements the same three-method interface: `Init() tea.Cmd`, `Update(msg tea.Msg) (tea.Model, tea.Cmd)`, `View() string`. State is immutable between Update calls. Side effects happen only through returned tea.Cmd functions.

**When to use:** Always. This is not optional in Bubble Tea -- it is the framework's core pattern.

**Trade-offs:** Forces unidirectional data flow (good for reasoning about state), but requires message types for all communication (verbose for simple state changes).

**Example:**
```go
// internal/calendar/model.go
type Model struct {
    year      int
    month     time.Month
    cursor    int        // selected day (1-based)
    holidays  []Holiday
    width     int
    height    int
    focused   bool
    keys      KeyMap
}

func (m Model) Init() tea.Cmd {
    return nil // Calendar has no async init
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if !m.focused {
            return m, nil
        }
        switch {
        case key.Matches(msg, m.keys.NextMonth):
            m = m.advanceMonth(1)
            return m, m.emitMonthChanged()
        case key.Matches(msg, m.keys.PrevMonth):
            m = m.advanceMonth(-1)
            return m, m.emitMonthChanged()
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}

func (m Model) View() string {
    return m.renderGrid() // Delegates to grid.go
}
```

### Pattern 2: Parent-Child Composition with Focus Routing

**What:** The root model embeds child models as struct fields (not pointers). The root Update method routes messages to the currently focused child. Only the focused child receives key messages; all children receive WindowSizeMsg.

**When to use:** Any multi-pane Bubble Tea application. This is the standard pattern shown in the official composable-views example.

**Trade-offs:** Simple and explicit. The root model grows linearly with the number of children (manageable for 2-3 panes). Focus state lives in the root, not the children.

**Example:**
```go
// internal/app/model.go
type pane int

const (
    calendarPane pane = iota
    todoPane
)

type Model struct {
    calendar   calendar.Model
    todoList   todolist.Model
    help       help.Model
    activePane pane
    width      int
    height     int
    keys       KeyMap
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Global keys handled first (quit, tab, help)
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keys.Tab):
            m.activePane = (m.activePane + 1) % 2
            m.calendar.SetFocused(m.activePane == calendarPane)
            m.todoList.SetFocused(m.activePane == todoPane)
            return m, nil
        }

    case tea.WindowSizeMsg:
        // Broadcast to ALL children
        m.width = msg.Width
        m.height = msg.Height
        m.calendar, _ = m.calendar.Update(msg)
        m.todoList, _ = m.todoList.Update(msg)
        return m, nil

    case calendar.MonthChangedMsg:
        // Cross-component communication: calendar tells todo list
        m.todoList, cmd = m.todoList.Update(msg)
        cmds = append(cmds, cmd)
    }

    // Route remaining messages to focused child only
    switch m.activePane {
    case calendarPane:
        var cmd tea.Cmd
        m.calendar, cmd = m.calendar.Update(msg)
        cmds = append(cmds, cmd)
    case todoPane:
        var cmd tea.Cmd
        m.todoList, cmd = m.todoList.Update(msg)
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}
```

### Pattern 3: Side-by-Side Layout with Lipgloss

**What:** The root View method renders each child's View independently, then joins them horizontally using `lipgloss.JoinHorizontal`. Each child is wrapped in a styled container (border, padding). The focused pane gets a distinct border color.

**When to use:** Split-pane layouts where panes sit side by side.

**Trade-offs:** Simple and declarative. Width calculation requires subtracting borders/padding from total terminal width and dividing between panes. Height must account for the help bar.

**Example:**
```go
// internal/app/model.go
func (m Model) View() string {
    // Calculate available space
    helpView := m.help.View(m.keys)
    helpHeight := lipgloss.Height(helpView)
    contentHeight := m.height - helpHeight - 2 // 2 for borders

    // Calculate pane widths (calendar is fixed-ish, todo gets the rest)
    calendarWidth := 24  // 7 cols x 3 chars + padding = ~24
    todoWidth := m.width - calendarWidth - 4 // 4 for borders/gaps

    // Style panes based on focus
    calStyle := m.paneStyle(m.activePane == calendarPane).
        Width(calendarWidth).
        Height(contentHeight)
    todoStyle := m.paneStyle(m.activePane == todoPane).
        Width(todoWidth).
        Height(contentHeight)

    // Compose
    top := lipgloss.JoinHorizontal(lipgloss.Top,
        calStyle.Render(m.calendar.View()),
        todoStyle.Render(m.todoList.View()),
    )

    return lipgloss.JoinVertical(lipgloss.Left, top, helpView)
}

func (m Model) paneStyle(focused bool) lipgloss.Style {
    if focused {
        return focusedStyle // bright border
    }
    return unfocusedStyle  // dim border
}
```

### Pattern 4: Commands for All I/O

**What:** All file reads, writes, and external data fetching happen inside `tea.Cmd` functions, never directly in `Update`. The command runs asynchronously and returns a message with the result. `Update` then processes the result message.

**When to use:** Always. This is not optional -- blocking in Update freezes the entire UI.

**Trade-offs:** More verbose than direct function calls, but keeps the UI responsive. Even fast disk reads should use commands because the pattern is consistent and future-proofs against slow operations.

**Example:**
```go
// internal/store/store.go (pure data layer, no tea dependency)
type Store struct {
    path string
}

func (s *Store) Load() ([]Todo, error) {
    data, err := os.ReadFile(s.path)
    if err != nil {
        if os.IsNotExist(err) {
            return []Todo{}, nil
        }
        return nil, err
    }
    var todos []Todo
    err = json.Unmarshal(data, &todos)
    return todos, err
}

func (s *Store) Save(todos []Todo) error {
    data, err := json.MarshalIndent(todos, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(s.path, data, 0644)
}

// internal/app/messages.go (tea.Cmd wrappers)
type todosLoadedMsg struct {
    todos []Todo
    err   error
}

func loadTodos(store *store.Store) tea.Cmd {
    return func() tea.Msg {
        todos, err := store.Load()
        return todosLoadedMsg{todos: todos, err: err}
    }
}

type todoSavedMsg struct {
    err error
}

func saveTodos(store *store.Store, todos []Todo) tea.Cmd {
    return func() tea.Msg {
        err := store.Save(todos)
        return todoSavedMsg{err: err}
    }
}
```

### Pattern 5: Lazy Initialization for Window Size

**What:** Components that need terminal dimensions do not initialize their layout in `Init()`. Instead, they wait for the first `tea.WindowSizeMsg` before calculating sizes and setting a `ready` flag. The `View` method returns a loading placeholder until `ready == true`.

**When to use:** Any component whose rendering depends on terminal width/height. Essentially all components in a responsive layout.

**Trade-offs:** Adds a `ready` bool and a brief "loading" flash on startup (typically one frame, imperceptible). Prevents panics from zero-width/height rendering.

**Example:**
```go
func (m Model) View() string {
    if !m.ready {
        return "Loading..."
    }
    return m.renderGrid()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        if !m.ready {
            m.ready = true
            // First-time initialization that needs dimensions
        }
    }
    return m, nil
}
```

## Data Flow

### Message Flow (Elm Architecture Loop)

```
User Input (keypress, mouse, resize)
    │
    ▼
tea.Program event loop
    │
    ▼
Root Model.Update(msg)
    │
    ├─── Global key? (quit/tab/help) ──▶ Handle directly, return
    │
    ├─── WindowSizeMsg? ──▶ Broadcast to ALL children
    │
    ├─── Cross-component msg? ──▶ Route to relevant child
    │    (e.g., MonthChangedMsg ──▶ todoList.Update)
    │
    └─── Other msg? ──▶ Route to FOCUSED child only
              │
              ▼
         Child.Update(msg)
              │
              ├─── Returns updated child model
              └─── Returns tea.Cmd (I/O) or nil
                        │
                        ▼
                   Async execution ──▶ Returns result tea.Msg
                                            │
                                            ▼
                                    Back to Root.Update(msg)
```

### Cross-Component Communication

```
Calendar: user navigates to March 2026
    │
    ▼
calendar.Update(KeyMsg) returns MonthChangedMsg{Year: 2026, Month: 3}
    │
    ▼
Root.Update receives MonthChangedMsg
    │
    ├──▶ Root issues loadTodos(store, 2026, 3) tea.Cmd
    │
    └──▶ todoList.Update(MonthChangedMsg) -- todo list updates its header
              │
              ▼
         todosLoadedMsg{todos: [...]} arrives
              │
              ▼
         Root.Update routes to todoList.Update
              │
              ▼
         Todo list re-renders with March todos
```

### Key Data Flows

1. **Startup flow:** `main.go` loads config, creates store, creates root Model with all children initialized, starts `tea.NewProgram(model, tea.WithAltScreen())`. Root's `Init()` returns `loadTodos(store, currentMonth)`. On `todosLoadedMsg`, todo list populates. First `WindowSizeMsg` triggers layout calculation.

2. **Month navigation flow:** Calendar emits `MonthChangedMsg`. Root intercepts, fires `loadTodos` command for new month. Todo list receives loaded data and re-renders.

3. **Todo CRUD flow:** User adds/toggles/deletes a todo in the todo list. Todo list emits a `TodoChangedMsg`. Root intercepts, fires `saveTodos` command. On `todoSavedMsg`, root can show a brief status message or silently succeed.

4. **Date selection flow:** User moves cursor in calendar to a specific date. Calendar emits `DateSelectedMsg{Date}`. Root forwards to todo list, which scrolls to or highlights that date's todos.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| Single user, single file | Current architecture: JSON file on disk, in-process store. No changes needed. |
| Multiple todo files / projects | Add a project selector. Store becomes directory-aware. Root model gains a project switcher. |
| Large todo lists (1000+ items) | Switch from loading all todos to paginated/streamed loading. Use bubbles/viewport for virtual scrolling. |
| Multi-user / sync | Out of scope for a local TUI. Would require a server, database, and sync protocol -- fundamentally different architecture. |

### Scaling Priorities

1. **First bottleneck:** Rendering performance with very long todo lists. A month with 100+ items could slow View. Mitigation: use viewport with fixed visible window, render only visible items.
2. **Second bottleneck:** File I/O on every change. For frequent toggling, batch writes with a debounce (save after 500ms of no changes, not on every toggle). Implement as a tea.Tick-based command.

## Anti-Patterns

### Anti-Pattern 1: Goroutines Instead of Commands

**What people do:** Launch `go func()` in Update to perform async work, then try to mutate the model from the goroutine.

**Why it's wrong:** Bubble Tea's event loop is single-threaded. Mutating the model from a goroutine causes data races. The framework cannot re-render when the goroutine finishes because it does not know about the change.

**Do this instead:** Return a `tea.Cmd` from Update. The command runs in a managed goroutine and returns a `tea.Msg` that flows back through Update.

### Anti-Pattern 2: Monolithic Root Model

**What people do:** Put all state (calendar state, todo state, config, styles) in a single large struct with a single Update function full of switch cases.

**Why it's wrong:** The Update function becomes hundreds of lines long. Key binding conflicts become hard to track. Testing requires constructing the entire app state.

**Do this instead:** Each pane is its own model with its own Update/View. The root model is a thin compositor that routes messages and joins views.

### Anti-Pattern 3: Hardcoded Dimensions

**What people do:** Set fixed column widths and row heights (e.g., `width := 80`) without responding to `tea.WindowSizeMsg`.

**Why it's wrong:** Breaks on any terminal size other than the one tested. Renders incorrectly on resize. Text wraps awkwardly or gets truncated.

**Do this instead:** Store width/height from `WindowSizeMsg`. Calculate pane sizes dynamically. Use lipgloss `Width()`, `Height()`, `MaxWidth()` to constrain rendering. Give the calendar a minimum width and let the todo list flex.

### Anti-Pattern 4: Not Propagating WindowSizeMsg

**What people do:** Handle `WindowSizeMsg` in the root model but forget to forward it to child models.

**Why it's wrong:** Children render with stale dimensions (often zero). This causes layout breakage, missing content, or panics. This is the single most common layout bug in multi-component Bubble Tea apps, as documented in GitHub discussion #943.

**Do this instead:** In root's Update, when receiving `WindowSizeMsg`, always forward to ALL children, not just the focused one. Calculate each child's allocated width/height and send adjusted messages.

### Anti-Pattern 5: Blocking I/O in Update

**What people do:** Call `os.ReadFile()` or `json.Marshal()` directly inside the Update function.

**Why it's wrong:** Update runs on the event loop. Any blocking call freezes the entire UI. Even fast disk reads should not block because the pattern normalizes blocking code in the hot path.

**Do this instead:** Wrap all I/O in `tea.Cmd` functions. Return the command from Update. Handle the result message when it arrives.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Local filesystem (todo JSON) | tea.Cmd wrapping os.ReadFile/WriteFile | Use atomic writes (write temp, rename) to prevent corruption |
| Holiday data source | Embedded Go data or fetched at build time | Avoid runtime API calls for a local-first app |
| Terminal | tea.WindowSizeMsg, tea.WithAltScreen | Alt screen provides clean enter/exit; always use it for full-screen TUIs |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| Root <-> Calendar | Message delegation: Root forwards KeyMsg to calendar when focused. Calendar returns MonthChangedMsg, DateSelectedMsg. | Calendar never touches todo data. |
| Root <-> TodoList | Message delegation: Root forwards KeyMsg when focused. Root sends todosLoadedMsg after loading. TodoList returns TodoChangedMsg. | TodoList never touches file system directly. |
| Root <-> Store | tea.Cmd only: Root creates commands that call store methods. Store returns data via messages. | Store has zero Bubble Tea dependency -- it is a plain Go package. |
| Calendar <-> Holiday Provider | Direct function call at construction or via tea.Cmd. Calendar holds holiday data for current month. | Holiday provider has zero Bubble Tea dependency. |
| Root <-> Help | help.Model is a Bubbles component. Root passes current key map to help.View(). | Help model has no custom messages -- it is view-only. |

## Build Order Implications

The component dependency graph dictates the natural build order:

```
Phase 1: Foundation (no UI)
    store/types.go      ── Todo struct definition
    store/store.go      ── File persistence (Load/Save)
    config/types.go     ── Config struct
    config/config.go    ── Config loading

Phase 2: Calendar Component (standalone)
    calendar/model.go   ── Month state, navigation
    calendar/grid.go    ── Grid rendering
    calendar/keys.go    ── Key bindings
    (Can be developed and tested independently)

Phase 3: Todo List Component (standalone)
    todolist/model.go   ── List state, CRUD operations
    todolist/keys.go    ── Key bindings
    (Can be developed and tested independently)

Phase 4: Root Composition (wires everything together)
    app/model.go        ── Embeds calendar + todolist
    app/keys.go         ── Global keys (tab, quit, help)
    app/styles.go       ── Pane styles, focus indicators
    app/messages.go     ── Cross-component messages
    main.go             ── Entry point

Phase 5: Holidays
    holiday/provider.go ── Holiday data integration
    calendar/model.go   ── Updated to highlight holidays
```

**Why this order:**
- Store and config have zero UI dependencies; build and test them first.
- Calendar and todolist can be developed in parallel once store types exist.
- Root composition is the riskiest part (layout, focus, message routing) -- tackle it only after children are solid.
- Holidays are cosmetic enhancement layered on last; the app is fully functional without them.

## Sources

- [Bubble Tea GitHub Repository](https://github.com/charmbracelet/bubbletea) -- official README, interface documentation (HIGH confidence)
- [Bubbles Component Library](https://github.com/charmbracelet/bubbles) -- available components (HIGH confidence)
- [Lipgloss Repository](https://github.com/charmbracelet/lipgloss) -- layout functions (HIGH confidence)
- [Lipgloss v2 API Documentation](https://pkg.go.dev/github.com/charmbracelet/lipgloss/v2) -- JoinHorizontal, JoinVertical, Place, Width, Height (HIGH confidence)
- [Tips for Building Bubble Tea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) -- component tree, message routing, debugging (MEDIUM confidence)
- [Managing Nested Models with Bubble Tea](https://donderom.com/posts/managing-nested-models-with-bubble-tea/) -- parent-child composition pattern (MEDIUM confidence)
- [Bubbletea State Machine Pattern](https://zackproser.com/blog/bubbletea-state-machine) -- stage-based architecture (MEDIUM confidence)
- [Component Integration DeepWiki](https://deepwiki.com/charmbracelet/bubbletea/6.5-component-integration) -- message delegation, lazy init, layout management (MEDIUM confidence)
- [Commands in Bubble Tea Blog Post](https://charm.land/blog/commands-in-bubbletea/) -- tea.Cmd patterns for I/O (HIGH confidence, official Charm blog)
- [GitHub Discussion #307: Layout Handling](https://github.com/charmbracelet/bubbletea/discussions/307) -- lipgloss join for multi-pane (MEDIUM confidence)
- [GitHub Discussion #943: View Layout Issues](https://github.com/charmbracelet/bubbletea/discussions/943) -- WindowSizeMsg propagation bug (HIGH confidence, demonstrates real pitfall)
- [GitHub Discussion #900: Overlapping Key Mappings](https://github.com/charmbracelet/bubbletea/discussions/900) -- focus-based key routing (MEDIUM confidence)
- [Composable Views Example](https://github.com/charmbracelet/bubbletea/blob/main/examples/composable-views/main.go) -- official side-by-side pattern (HIGH confidence)
- [Lipgloss Layout Example](https://github.com/charmbracelet/lipgloss/blob/master/examples/layout/main.go) -- full-screen layout pattern (HIGH confidence)
- [Glow Project Structure](https://github.com/charmbracelet/glow) -- real-world Charm app with ui/ directory (MEDIUM confidence)

---
*Architecture research for: TUI calendar + todo application*
*Researched: 2026-02-05*
