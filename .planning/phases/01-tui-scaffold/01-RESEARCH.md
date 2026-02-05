# Phase 1: TUI Scaffold - Research

**Researched:** 2026-02-05
**Domain:** Go TUI split-pane scaffold with Bubble Tea v1
**Confidence:** HIGH

## Summary

Phase 1 delivers the foundational Bubble Tea application: a runnable binary that shows a two-pane terminal layout with keyboard navigation and responsive resize handling. This phase implements requirements UI-01 (split-pane layout), UI-02 (keyboard navigation with Tab), and UI-04 (terminal resize). No calendar rendering, no todo management, no persistence -- just the structural skeleton.

The standard approach is well-established: a root Bubble Tea model composing two child models (left pane, right pane) using Lip Gloss `JoinHorizontal` for layout. Focus routing via an `activePane` enum determines which child receives keyboard messages. All children receive `tea.WindowSizeMsg`. The `ready` guard pattern prevents rendering before terminal dimensions are known.

The stack is locked: Bubble Tea v1.3.10, Lip Gloss v1.1.0, Bubbles v0.21.1. The project structure follows `internal/` package conventions with separate packages for the root app, calendar component, and todolist component. Phase 1 creates placeholder child models that display static text -- later phases flesh them out.

**Primary recommendation:** Build the thinnest possible scaffold -- root model with two placeholder panes, Tab switching, focus-aware borders, WindowSizeMsg propagation, and quit handling. Resist adding any feature logic.

## Standard Stack

### Core (Phase 1 only)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.25.x | Language runtime | Current stable release line |
| Bubble Tea | v1.3.10 | TUI framework (Elm Architecture) | De facto Go TUI standard; locked decision from project research |
| Lip Gloss | v1.1.0 | Terminal styling and layout | Official Charm companion; provides JoinHorizontal, borders, Width/Height |
| Bubbles | v0.21.1 | Pre-built TUI components | Provides `key.Binding`, `key.Matches`, `help.Model` for key management |

### Supporting (Phase 1)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `bubbles/key` | v0.21.1 | Key binding definitions | All keyboard handling -- use `key.NewBinding` + `key.Matches` pattern |
| `bubbles/help` | v0.21.1 | Help bar component | Optional in Phase 1; needed in Phase 3. Can stub a minimal help bar now |

### Not Needed in Phase 1

| Library | Why Not Yet |
|---------|-------------|
| `bubbles/list` | Todo list is Phase 3 |
| `bubbles/textinput` | Todo input is Phase 3 |
| `bubbles/viewport` | Scrolling not needed until content overflows |
| `rickar/cal` | Holidays are Phase 2 |
| `BurntSushi/toml` | Config is Phase 2 |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled split pane | `john-marinelli/panes` library | panes has 14 stars, no releases, 5 commits -- too immature; lipgloss JoinHorizontal is trivial for a 2-pane layout |
| lipgloss JoinHorizontal | lipgloss.Place | Place is for centering content in a box; JoinHorizontal is the correct tool for side-by-side panes |

**Installation (Phase 1):**

```bash
# Initialize module
go mod init github.com/antti/todo-calendar

# Core framework (all three needed even for scaffold)
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/lipgloss@v1.1.0
go get github.com/charmbracelet/bubbles@v0.21.1
```

## Architecture Patterns

### Recommended Project Structure (Phase 1 creates this skeleton)

```
todo-calendar/
├── main.go                  # Entry point: tea.NewProgram + Run
├── go.mod
├── go.sum
├── internal/
│   ├── app/
│   │   ├── model.go         # Root model: Init, Update, View
│   │   ├── keys.go          # Global key bindings (quit, tab)
│   │   └── styles.go        # Lip Gloss styles (borders, focus colors)
│   ├── calendar/
│   │   └── model.go         # Placeholder: "Calendar pane" text
│   └── todolist/
│       └── model.go         # Placeholder: "Todo list pane" text
└── .planning/               # Not shipped
```

**Phase 1 rationale:** Create all three packages now even though calendar and todolist are placeholders. This establishes the package boundaries, import paths, and message-routing patterns from the start. Phase 2 and 3 fill in the real implementations without restructuring.

### Pattern 1: Root Model with Focus Routing

**What:** Root model owns an `activePane` enum and two child models. Global keys (quit, Tab) are handled by the root. All other KeyMsg is forwarded only to the focused child. WindowSizeMsg is broadcast to all children.

**When to use:** Always -- this is the standard Bubble Tea multi-component pattern.

**Confidence:** HIGH (verified from official composable-views example and pkg.go.dev docs)

**Example:**

```go
// internal/app/model.go
// Source: Pattern verified against charmbracelet/bubbletea composable-views example

type pane int

const (
    calendarPane pane = iota
    todoPane
)

type Model struct {
    calendar   calendar.Model
    todoList   todolist.Model
    activePane pane
    width      int
    height     int
    ready      bool
    keys       KeyMap
}

func New() Model {
    return Model{
        calendar:   calendar.New(),
        todoList:   todolist.New(),
        activePane: calendarPane,
        keys:       DefaultKeyMap(),
    }
}

func (m Model) Init() tea.Cmd {
    return nil // No async init needed for scaffold
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Global keys first
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keys.Tab):
            if m.activePane == calendarPane {
                m.activePane = todoPane
            } else {
                m.activePane = calendarPane
            }
            m.calendar.SetFocused(m.activePane == calendarPane)
            m.todoList.SetFocused(m.activePane == todoPane)
            return m, nil
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.ready = true
        // Broadcast to ALL children
        var cmd tea.Cmd
        m.calendar, cmd = m.calendar.Update(msg)
        cmds = append(cmds, cmd)
        m.todoList, cmd = m.todoList.Update(msg)
        cmds = append(cmds, cmd)
        return m, tea.Batch(cmds...)
    }

    // Route to focused child only
    var cmd tea.Cmd
    switch m.activePane {
    case calendarPane:
        m.calendar, cmd = m.calendar.Update(msg)
    case todoPane:
        m.todoList, cmd = m.todoList.Update(msg)
    }
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}
```

### Pattern 2: Side-by-Side View with Frame-Aware Layout

**What:** Root View joins two styled pane strings with `lipgloss.JoinHorizontal`. The focused pane gets a bright border, the unfocused gets a dim border. Width calculation subtracts frame overhead using `Style.GetFrameSize()`.

**When to use:** Always for the split-pane layout.

**Confidence:** HIGH (verified JoinHorizontal, GetFrameSize, Width, Height from pkg.go.dev lipgloss v1.1.0 docs)

**Example:**

```go
// internal/app/model.go - View method
// Source: Verified against lipgloss v1.1.0 pkg.go.dev documentation

func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }

    // Measure frame overhead from the pane style
    frameH, frameV := m.focusedStyle().GetFrameSize()

    // Reserve height for help bar (even if minimal in Phase 1)
    helpHeight := 1
    contentHeight := m.height - helpHeight - frameV

    // Calendar pane gets a fixed width; todo pane gets the rest
    calendarInnerWidth := 24 // enough for 7-column month grid
    todoInnerWidth := m.width - calendarInnerWidth - (frameH * 2) // both panes have frames

    // Apply styles based on focus
    calStyle := m.paneStyle(m.activePane == calendarPane).
        Width(calendarInnerWidth).
        Height(contentHeight)
    todoStyle := m.paneStyle(m.activePane == todoPane).
        Width(todoInnerWidth).
        Height(contentHeight)

    // Compose layout
    top := lipgloss.JoinHorizontal(lipgloss.Top,
        calStyle.Render(m.calendar.View()),
        todoStyle.Render(m.todoList.View()),
    )

    statusBar := "Press q to quit | Tab to switch panes"

    return lipgloss.JoinVertical(lipgloss.Left, top, statusBar)
}
```

### Pattern 3: Lazy Initialization (ready guard)

**What:** Components check a `ready` bool before rendering layout-dependent content. The `ready` flag is set to `true` on the first `tea.WindowSizeMsg`.

**When to use:** Every component that depends on terminal dimensions. In Phase 1, the root model uses this pattern.

**Confidence:** HIGH (documented in Bubble Tea issue #282 and multiple official examples)

**Example:**

```go
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }
    // Layout code that uses m.width and m.height
}
```

### Pattern 4: Child Model Interface (not tea.Model)

**What:** Child models (calendar, todolist) do NOT implement `tea.Model` directly. They return their own concrete type from Update, not `tea.Model`. This avoids type assertions when the root assigns the updated child back.

**When to use:** All child components in a composed Bubble Tea app.

**Confidence:** HIGH (standard pattern in composable-views example and community guides)

**Example:**

```go
// internal/calendar/model.go
type Model struct {
    focused bool
    width   int
    height  int
}

func New() Model {
    return Model{}
}

// Returns (calendar.Model, tea.Cmd), NOT (tea.Model, tea.Cmd)
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}

func (m Model) View() string {
    if m.focused {
        return "[ Calendar Pane - FOCUSED ]"
    }
    return "  Calendar Pane"
}

func (m *Model) SetFocused(f bool) {
    m.focused = f
}
```

### Pattern 5: main.go Minimal Entry Point

**What:** main.go does nothing except create the root model, wrap it in `tea.NewProgram` with `tea.WithAltScreen()`, call `Run()`, and exit on error.

**Confidence:** HIGH (verified from Bubble Tea pkg.go.dev docs)

**Example:**

```go
// main.go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/antti/todo-calendar/internal/app"
)

func main() {
    m := app.New()
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Anti-Patterns to Avoid

- **Monolithic model:** Do NOT put calendar and todo state in the root model. Each pane gets its own package with its own Model type, even if it is a placeholder in Phase 1.
- **Pointer receivers on Update:** Bubble Tea's Elm Architecture expects value semantics. `Update` returns a new model value, not mutating via pointer. Use `func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)`.
- **Hardcoded dimensions:** Never use `width := 80`. Always derive from `tea.WindowSizeMsg`.
- **Skipping the ready guard:** Always check `m.ready` or `m.width == 0` before layout computation.
- **Implementing tea.Model on children:** Child `Update` should return their concrete type, not `tea.Model`. Only the root model satisfies the `tea.Model` interface.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Side-by-side pane layout | Custom ANSI cursor positioning | `lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)` | JoinHorizontal handles line-by-line alignment, padding differences, and ANSI-aware width |
| Measuring styled string width | `len(str)` or `utf8.RuneCountInString` | `lipgloss.Width(str)` | Width() ignores ANSI escape sequences and handles wide characters (CJK, emoji) |
| Border rendering | Manual box-drawing characters | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` | Lip Gloss handles corner joining, side rendering, and padding consistently |
| Key binding matching | `msg.String() == "q"` string comparison | `key.Matches(msg, m.keys.Quit)` from `bubbles/key` | key.Matches supports multiple key alternatives per binding, integrates with help.Model |
| Alt screen management | Manual ANSI escape sequences for alt screen | `tea.WithAltScreen()` program option | Bubble Tea handles enter/exit alt screen, restoring terminal state on quit/crash |
| Frame size calculation | Manual counting of border/padding chars | `style.GetFrameSize()` returns `(horizontalOverhead, verticalOverhead)` | Accounts for borders, padding, and margins in one call; no manual arithmetic |

**Key insight:** Lip Gloss and Bubbles already solve every layout and keyboard problem Phase 1 needs. The only custom code is the root model compositor and placeholder child models.

## Common Pitfalls

### Pitfall 1: View() Called Before WindowSizeMsg Arrives

**What goes wrong:** Bubble Tea calls View() immediately after Init(). If View() computes layout using `m.width` and `m.height`, these are zero, causing division-by-zero or negative widths.

**Why it happens:** WindowSizeMsg is delivered asynchronously after program start. The first View() call happens before it arrives.

**How to avoid:** Use the `ready` guard pattern. Return "Initializing..." until the first WindowSizeMsg sets `m.ready = true`.

**Warning signs:** Panic on startup, flash of garbage layout, zero-width panes.

**Confidence:** HIGH (documented in bubbletea issue #282)

### Pitfall 2: Frame Size Accounting Errors

**What goes wrong:** Pane widths overflow the terminal because borders and padding are not subtracted from the available width. Content wraps to the next line, destroying the side-by-side layout.

**Why it happens:** A rounded border adds 2 characters horizontally (left + right). Padding adds more. Two panes with borders add 4+ characters total. Developers forget to subtract this overhead.

**How to avoid:** Always call `style.GetFrameSize()` to measure overhead. Calculate: `availableContentWidth = terminalWidth - leftFrameH - rightFrameH`. Then divide content width between panes.

**Warning signs:** Layout breaks at certain terminal widths, line wrapping in the terminal, manual `+1`/`-1` fudge constants.

**Confidence:** HIGH (verified GetFrameSize exists in lipgloss v1.1.0 docs)

### Pitfall 3: Not Broadcasting WindowSizeMsg to All Children

**What goes wrong:** Only the focused child receives the resize message. The unfocused child renders with stale dimensions. When the user tabs to it, the layout breaks.

**Why it happens:** The default message routing pattern sends messages to the focused child only. WindowSizeMsg is special -- it must go to ALL children.

**How to avoid:** In the root Update, handle `tea.WindowSizeMsg` as a special case that broadcasts to every child model before returning. Do not fall through to the focus-based routing.

**Warning signs:** Switching focus after a resize causes layout breakage.

**Confidence:** HIGH (documented in bubbletea discussion #943)

### Pitfall 4: Using tea.Model Interface for Child Components

**What goes wrong:** If calendar.Update returns `(tea.Model, tea.Cmd)` instead of `(calendar.Model, tea.Cmd)`, the root model needs a type assertion `m.calendar = updated.(calendar.Model)` which is fragile and panics if wrong.

**Why it happens:** Developers assume all models must implement tea.Model. Only the root model needs to satisfy the interface.

**How to avoid:** Child Update methods return their concrete type. The root model calls `m.calendar, cmd = m.calendar.Update(msg)` with no type assertion needed.

**Warning signs:** Type assertions in the root Update method.

**Confidence:** HIGH (standard pattern from composable-views example)

### Pitfall 5: Forgetting tea.WithAltScreen

**What goes wrong:** Without alt screen, the TUI renders inline in the terminal. Scrollback history mixes with the app output. Quitting leaves the UI remnants in the terminal.

**Why it happens:** `tea.NewProgram(model)` works without options but renders inline by default.

**How to avoid:** Always use `tea.NewProgram(model, tea.WithAltScreen())` for full-screen TUI applications.

**Warning signs:** App appears to render "within" the terminal output rather than taking over the full screen.

**Confidence:** HIGH (verified in pkg.go.dev bubbletea v1.3.10 docs)

### Pitfall 6: Negative or Zero Width Passed to Child

**What goes wrong:** When the terminal is very narrow, the width calculation for the todo pane can become zero or negative: `todoWidth = termWidth - calendarWidth - frameOverhead` might be negative if `termWidth < calendarWidth + frameOverhead`. Lip Gloss Width() with a negative value causes unexpected behavior.

**Why it happens:** No minimum width guard on the calculation.

**How to avoid:** Clamp all calculated widths to a minimum of 1. If the terminal is too narrow to display both panes, either hide one pane or display a "terminal too small" message.

**Warning signs:** Panics or garbled output when the terminal is resized very narrow.

**Confidence:** MEDIUM (logical deduction from API behavior; not directly documented as a pitfall)

## Code Examples

### Complete Key Map for Phase 1

```go
// internal/app/keys.go
// Source: bubbles/key v0.21.1 pkg.go.dev docs

package app

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
    Quit key.Binding
    Tab  key.Binding
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        Quit: key.NewBinding(
            key.WithKeys("q", "ctrl+c"),
            key.WithHelp("q", "quit"),
        ),
        Tab: key.NewBinding(
            key.WithKeys("tab"),
            key.WithHelp("tab", "switch pane"),
        ),
    }
}

// ShortHelp implements help.KeyMap (for future help bar)
func (k KeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Tab, k.Quit}
}

// FullHelp implements help.KeyMap
func (k KeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Tab, k.Quit},
    }
}
```

### Complete Styles for Phase 1

```go
// internal/app/styles.go
// Source: lipgloss v1.1.0 pkg.go.dev docs

package app

import "github.com/charmbracelet/lipgloss"

var (
    focusedBorderColor   = lipgloss.Color("62")  // purple-ish
    unfocusedBorderColor = lipgloss.Color("240")  // gray

    focusedStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(focusedBorderColor).
        Padding(0, 1)

    unfocusedStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(unfocusedBorderColor).
        Padding(0, 1)
)

func paneStyle(focused bool) lipgloss.Style {
    if focused {
        return focusedStyle
    }
    return unfocusedStyle
}
```

### Minimum Viable Placeholder Child Model

```go
// internal/calendar/model.go
package calendar

import tea "github.com/charmbracelet/bubbletea"

type Model struct {
    focused bool
    width   int
    height  int
}

func New() Model {
    return Model{}
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}

func (m Model) View() string {
    label := "Calendar"
    if m.focused {
        label += " (focused)"
    }
    return label
}

func (m *Model) SetFocused(f bool) {
    m.focused = f
}
```

### Program Entry Point

```go
// main.go
// Source: bubbletea v1.3.10 pkg.go.dev docs
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/antti/todo-calendar/internal/app"
)

func main() {
    p := tea.NewProgram(app.New(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
        os.Exit(1)
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `msg.String() == "q"` | `key.Matches(msg, binding)` | Bubbles key package | Type-safe, supports multiple keys per binding, integrates with help |
| `lipgloss.Width(n)` as method | `style.Width(n)` on Style, `lipgloss.Width(str)` as function | lipgloss v1.x | Two different things: style method sets min width, function measures rendered width |
| `tea.WithAltScreen` as `Cmd` in Init | `tea.WithAltScreen()` as `ProgramOption` | bubbletea v0.23+ | Use as program option, not as command; program option is the correct approach for v1 |
| Inline rendering | Alt screen by default for full-screen apps | Always | Full-screen TUIs should always use alt screen |

**Deprecated/outdated:**
- `ioutil.WriteFile`: Removed in Go 1.16+; use `os.WriteFile` instead (not relevant for Phase 1 but worth noting)
- `tea.ClearScreen` as initial command: Not needed with alt screen mode
- Manual terminal size detection via `term.GetSize`: Use `tea.WindowSizeMsg` instead -- Bubble Tea sends it automatically

## Open Questions

1. **Calendar pane width**
   - What we know: A `cal`-like grid is 7 columns x 3 characters = 21, plus borders and padding
   - What's unclear: Exact inner width depends on whether we show week numbers (adds ~3 chars) and month/year header alignment
   - Recommendation: Use 24 as the initial fixed inner width for the calendar pane. Adjust in Phase 2 when the actual calendar grid is implemented. The architecture supports changing this constant easily.

2. **Minimum terminal size**
   - What we know: The split-pane layout needs at least ~50 columns to display both panes meaningfully
   - What's unclear: Should we enforce a minimum or degrade gracefully?
   - Recommendation: Add a simple check in View() -- if width < 50, show a "terminal too small" message instead of the layout. This prevents garbled output on very narrow terminals.

3. **Help bar in Phase 1**
   - What we know: UI-03 (help bar) is mapped to Phase 3. But having a minimal status line ("q: quit | Tab: switch") in Phase 1 aids development.
   - What's unclear: Whether to use `help.Model` from Bubbles now or just a plain string
   - Recommendation: Use a plain string for Phase 1. Switch to `help.Model` in Phase 3 when context-sensitive bindings are needed. This avoids premature complexity.

## Sources

### Primary (HIGH confidence)
- [bubbletea v1.3.10 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea@v1.3.10) -- tea.Model interface, WindowSizeMsg, KeyMsg, ProgramOption types verified
- [lipgloss v1.1.0 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/lipgloss@v1.1.0) -- JoinHorizontal, JoinVertical, Width(), GetFrameSize(), Border types verified
- [bubbles/key v0.21.1 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/key) -- key.Binding, key.NewBinding, key.Matches, key.WithKeys, key.WithHelp verified
- [bubbles/help v0.21.1 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/help) -- help.Model, help.KeyMap interface verified
- [composable-views example](https://github.com/charmbracelet/bubbletea/blob/main/examples/composable-views/main.go) -- Official pattern for multi-component Bubble Tea apps
- [lipgloss layout example](https://github.com/charmbracelet/lipgloss/blob/master/examples/layout/main.go) -- Full-screen layout composition pattern
- [bubbletea issue #282](https://github.com/charmbracelet/bubbletea/issues/282) -- View() called before WindowSizeMsg documented behavior

### Secondary (MEDIUM confidence)
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) -- Component tree, message routing patterns
- [bubbletea discussion #943](https://github.com/charmbracelet/bubbletea/discussions/943) -- WindowSizeMsg propagation pitfall documented by community
- [Managing nested models](https://donderom.com/posts/managing-nested-models-with-bubble-tea/) -- Parent-child composition patterns

### Tertiary (LOW confidence)
- [john-marinelli/panes](https://github.com/john-marinelli/panes) -- Evaluated and rejected: 14 stars, no releases, too immature for production use

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- versions locked by project-level research; APIs verified against pkg.go.dev
- Architecture: HIGH -- patterns verified from official examples and documentation
- Pitfalls: HIGH -- sourced from official GitHub issues/discussions and verified API docs
- Code examples: HIGH -- all API calls verified against v1.3.10/v1.1.0/v0.21.1 documentation

**Research date:** 2026-02-05
**Valid until:** 2026-03-07 (30 days -- stable v1 libraries, no imminent breaking changes)
