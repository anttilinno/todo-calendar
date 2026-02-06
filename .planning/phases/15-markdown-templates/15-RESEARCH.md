# Phase 15: Markdown Templates - Research

**Researched:** 2026-02-06
**Domain:** Markdown body field, template system, Glamour rendering, multi-step TUI input
**Confidence:** HIGH

## Summary

Phase 15 adds markdown body support to todos, reusable templates with `{{.Variable}}` placeholders, interactive placeholder prompting, and a styled markdown preview pane. The codebase already has the `body` column in the SQLite schema (TEXT NOT NULL DEFAULT '') but it is excluded from SELECTs and the `Todo` struct lacks a `Body` field.

The standard approach uses: (1) `text/template` from Go's stdlib for template parsing and execution, (2) `text/template/parse` for extracting placeholder field names from template content, (3) `github.com/charmbracelet/glamour` v0.10.0 for terminal markdown rendering with `ansi.StyleConfig` for theme integration. Multi-step placeholder input follows the existing sequential-mode pattern already used in the todolist (inputMode -> dateInputMode) rather than pulling in the `huh` forms library.

**Primary recommendation:** Add Body to Todo struct and SQL queries, create a `templates` table with migration v2, build a preview overlay following the search/settings overlay pattern, and implement template creation and placeholder prompting as additional modes in the todolist component.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| text/template | stdlib | Template parsing and execution with `{{.Variable}}` placeholders | Go stdlib; exactly matches requirement spec |
| text/template/parse | stdlib | Walk parse tree to extract placeholder field names for prompting | Go stdlib; direct access to template AST |
| github.com/charmbracelet/glamour | v0.10.0 | Render markdown to styled ANSI terminal output | Charm ecosystem; integrates with lipgloss/bubbletea; active maintenance |
| github.com/charmbracelet/glamour/ansi | (part of glamour) | `StyleConfig` struct for custom theme colors | Required for matching app theme colors to glamour rendering |
| github.com/charmbracelet/glamour/styles | (part of glamour) | `DarkStyleConfig`, `LightStyleConfig` base styles to copy/modify | Starting point for theme-matched rendering |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/charmbracelet/bubbles/viewport | (already in bubbles dep) | Scrollable viewport for long markdown bodies | Preview pane for bodies longer than terminal height |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| glamour v0.10.0 | glamour v2 | v2 is untagged pre-release (v2.0.0-2025...); not stable. Use v1 |
| text/template | regexp for {{.Var}} extraction | Fragile; template/parse gives exact AST, handles edge cases |
| hand-rolled multi-step input | charmbracelet/huh forms library | huh is a large new dependency; existing sequential mode pattern (inputMode -> dateInputMode) already works perfectly |
| Custom markdown renderer | goldmark + lipgloss | Reinventing glamour; no benefit |

**Installation:**
```bash
go get github.com/charmbracelet/glamour@v0.10.0
```
(No other new dependencies needed -- text/template is stdlib, viewport is already available through bubbles.)

## Architecture Patterns

### Recommended Changes by Package

```
internal/
├── store/
│   ├── todo.go        # Add Body field to Todo struct
│   ├── sqlite.go      # Update todoColumns, scanTodo, Add, Update, Find queries
│   │                  # Add templates table in migration v2
│   │                  # Add template CRUD methods
│   └── store.go       # Extend TodoStore interface with body-aware methods + template methods
├── todolist/
│   ├── model.go       # Add preview mode, template selection mode, placeholder input mode
│   ├── keys.go        # Add Preview, TemplateAdd, BodyEdit keybindings
│   └── styles.go      # Add BodyIndicator style
├── preview/           # NEW package: markdown preview overlay
│   ├── model.go       # Bubble Tea model wrapping glamour + viewport
│   ├── styles.go      # Theme-integrated glamour StyleConfig builder
│   └── keys.go        # Scroll, close keybindings
└── app/
    ├── model.go       # Wire preview overlay (like search/settings pattern)
    └── keys.go        # Add preview toggle keybinding
```

### Pattern 1: Body Field in Todo Struct and SQL
**What:** Add `Body string` to `Todo` struct, include `body` column in all SELECTs, update `Add` and `Update` to write body.
**When to use:** MDTPL-01 implementation.
**Example:**
```go
// store/todo.go
type Todo struct {
    ID        int    `json:"id"`
    Text      string `json:"text"`
    Body      string `json:"body"`      // NEW
    Date      string `json:"date,omitempty"`
    Done      bool   `json:"done"`
    CreatedAt string `json:"created_at"`
    SortOrder int    `json:"sort_order,omitempty"`
}

// HasBody reports whether the todo has a non-empty markdown body.
func (t Todo) HasBody() bool {
    return t.Body != ""
}
```

```go
// store/sqlite.go -- update column list and scanner
const todoColumns = "id, text, body, date, done, created_at, sort_order"

func scanTodo(scanner interface{ Scan(...any) error }) (Todo, error) {
    var t Todo
    var date sql.NullString
    var done int
    err := scanner.Scan(&t.ID, &t.Text, &t.Body, &date, &done, &t.CreatedAt, &t.SortOrder)
    // ...
}
```

### Pattern 2: Templates Table (Migration v2)
**What:** Create a `templates` table for storing reusable markdown templates with placeholders.
**When to use:** MDTPL-02 implementation.
**Example:**
```go
// store/sqlite.go -- migration v2
if version < 2 {
    if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS templates (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        name       TEXT    NOT NULL UNIQUE,
        content    TEXT    NOT NULL,
        created_at TEXT    NOT NULL
    )`); err != nil {
        return fmt.Errorf("create templates table: %w", err)
    }
    if _, err := s.db.Exec(`PRAGMA user_version = 2`); err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

### Pattern 3: Template struct and store methods
**What:** Define a Template type and CRUD methods.
**Example:**
```go
// store/todo.go (or store/template.go)
type Template struct {
    ID        int
    Name      string
    Content   string
    CreatedAt string
}

// store/store.go -- extend TodoStore (or create TemplateStore)
type TodoStore interface {
    // ... existing methods ...

    // Template methods
    AddTemplate(name, content string) (Template, error)
    ListTemplates() []Template
    FindTemplate(id int) *Template
    DeleteTemplate(id int)
}
```

### Pattern 4: Placeholder Extraction via parse Tree Walk
**What:** Parse template content, walk the AST to find all `FieldNode` instances, extract unique placeholder names.
**When to use:** MDTPL-03 -- before prompting user for values.
**Example:**
```go
// Can live in a utility function, e.g., store/template.go or a new internal/tmpl package

import "text/template/parse"

// ExtractPlaceholders returns the unique {{.Field}} names from a template string.
func ExtractPlaceholders(content string) ([]string, error) {
    trees, err := parse.Parse("tpl", content, "{{", "}}")
    if err != nil {
        return nil, err
    }
    tree := trees["tpl"]
    seen := make(map[string]bool)
    var names []string
    walkFields(tree.Root, seen, &names)
    return names, nil
}

func walkFields(node parse.Node, seen map[string]bool, names *[]string) {
    if node == nil {
        return
    }
    switch n := node.(type) {
    case *parse.ListNode:
        for _, child := range n.Nodes {
            walkFields(child, seen, names)
        }
    case *parse.ActionNode:
        if n.Pipe != nil {
            walkFields(n.Pipe, seen, names)
        }
    case *parse.PipeNode:
        for _, cmd := range n.Cmds {
            walkFields(cmd, seen, names)
        }
    case *parse.CommandNode:
        for _, arg := range n.Args {
            walkFields(arg, seen, names)
        }
    case *parse.FieldNode:
        // FieldNode.Ident is ["FieldName"] for {{.FieldName}}
        if len(n.Ident) > 0 {
            name := n.Ident[0]
            if !seen[name] {
                seen[name] = true
                *names = append(*names, name)
            }
        }
    case *parse.IfNode:
        walkFields(n.List, seen, names)
        walkFields(n.ElseList, seen, names)
        if n.Pipe != nil {
            walkFields(n.Pipe, seen, names)
        }
    case *parse.RangeNode:
        walkFields(n.List, seen, names)
        walkFields(n.ElseList, seen, names)
        if n.Pipe != nil {
            walkFields(n.Pipe, seen, names)
        }
    case *parse.WithNode:
        walkFields(n.List, seen, names)
        walkFields(n.ElseList, seen, names)
        if n.Pipe != nil {
            walkFields(n.Pipe, seen, names)
        }
    }
}
```

### Pattern 5: Template Execution (Filling Placeholders)
**What:** Execute a `text/template` with user-provided values to produce the final markdown body.
**Example:**
```go
import "text/template"

func ExecuteTemplate(content string, values map[string]string) (string, error) {
    tmpl, err := template.New("tpl").Parse(content)
    if err != nil {
        return "", err
    }
    var buf strings.Builder
    if err := tmpl.Execute(&buf, values); err != nil {
        return "", err
    }
    return buf.String(), nil
}
```
Note: `text/template` accepts `map[string]string` as the data argument since `{{.Key}}` accesses map keys directly.

### Pattern 6: Glamour Renderer with Theme Integration
**What:** Create a glamour `TermRenderer` that matches the app's current theme.
**When to use:** MDTPL-04 -- rendering markdown body in preview pane.
**Example:**
```go
import (
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/glamour/ansi"
    "github.com/charmbracelet/glamour/styles"
)

// NewMarkdownRenderer creates a glamour renderer matching the app theme.
func NewMarkdownRenderer(themeName string, width int) (*glamour.TermRenderer, error) {
    // Start from a base style matching the theme brightness
    var baseStyle ansi.StyleConfig
    switch themeName {
    case "light":
        baseStyle = styles.LightStyleConfig
    default:
        baseStyle = styles.DarkStyleConfig
    }
    // Optionally customize specific colors from our theme here
    // e.g., override heading colors to match AccentFg

    return glamour.NewTermRenderer(
        glamour.WithStyles(baseStyle),
        glamour.WithWordWrap(width),
    )
}
```

### Pattern 7: Preview Overlay (Following Search/Settings Pattern)
**What:** Create a full-screen overlay for viewing a todo's rendered markdown body.
**When to use:** MDTPL-04 -- showing the preview pane.
**Example architecture:**
```go
// preview/model.go
type Model struct {
    viewport viewport.Model  // scrollable content area
    content  string          // rendered markdown output
    width    int
    height   int
    keys     KeyMap
    styles   Styles
}

// Close message to return to normal view
type CloseMsg struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if key.Matches(msg, m.keys.Close) {
            return m, func() tea.Msg { return CloseMsg{} }
        }
    }
    // Forward to viewport for scrolling
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}
```

The app/model.go wires this identically to how search and settings overlays work:
- `showPreview bool` + `preview preview.Model` fields
- `preview.CloseMsg` handler
- Route messages to preview when open
- Render preview view when showPreview is true

### Pattern 8: Multi-Step Placeholder Input (Extending Existing Mode Pattern)
**What:** Add new modes to todolist for sequential placeholder prompting, following the existing inputMode -> dateInputMode pattern.
**When to use:** MDTPL-03 -- filling template placeholders.
**Example:**
```go
// todolist/model.go -- extend mode enum
const (
    normalMode         mode = iota
    inputMode
    dateInputMode
    editTextMode
    editDateMode
    filterMode
    templateSelectMode    // NEW: choosing a template
    placeholderInputMode  // NEW: filling in each placeholder
)

// New Model fields:
// pendingTemplate   *store.Template  // selected template
// placeholderNames  []string         // extracted field names
// placeholderIndex  int              // which placeholder we're prompting for
// placeholderValues map[string]string // collected values so far
```

The flow:
1. User presses template-add keybinding -> enter templateSelectMode
2. Show list of templates, user selects one with enter
3. Extract placeholders from template content using ExtractPlaceholders()
4. If no placeholders, execute template directly and create todo
5. If placeholders exist, enter placeholderInputMode
6. Prompt for each placeholder (input.Placeholder = "Enter value for: ProjectName")
7. On enter, store value, advance to next placeholder
8. After last placeholder, execute template, create todo with body
9. On esc at any point, cancel and return to normalMode

### Pattern 9: Body Indicator in Todo List
**What:** Show a small indicator (e.g., "[+]" or a special character) next to todos that have a non-empty body.
**When to use:** MDTPL-01 -- visual affordance that body exists.
**Example:**
```go
// todolist/model.go -- in renderTodo()
func (m Model) renderTodo(b *strings.Builder, t *store.Todo, selected bool) {
    // ... existing cursor + checkbox logic ...

    text := t.Text
    if t.HasBody() {
        text += " " + m.styles.BodyIndicator.Render("[+]")
    }
    if t.HasDate() {
        text += " " + m.styles.Date.Render(config.FormatDate(t.Date, m.dateLayout))
    }
    // ... rest of rendering ...
}
```

### Anti-Patterns to Avoid
- **Embedding the full markdown viewer inline in the todo list**: The todo list renders single lines per item. Preview must be a separate overlay/pane.
- **Using glamour WithAutoStyle()**: This detects terminal background at runtime. Since we have an explicit theme system, use WithStyles() with a style config derived from the current theme.
- **Storing templates as files**: Templates belong in SQLite alongside todos for portability and consistency.
- **Using huh forms library**: Adds a heavy new dependency for something the existing sequential-mode pattern handles naturally.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Markdown to ANSI rendering | Custom markdown parser with lipgloss | glamour | Handles headings, lists, code blocks, tables, links, emphasis, colors |
| Template variable parsing | Regex for `{{.Var}}` | text/template/parse | Handles nested templates, pipelines, conditionals correctly |
| Template execution | String replacement | text/template.Execute | Type-safe, handles escaping, missing fields, conditionals |
| Scrollable content pane | Manual offset tracking | bubbles/viewport | Handles page up/down, mouse scroll, content overflow |

**Key insight:** Markdown rendering is deceptively complex. Even "simple" markdown has dozens of element types with nesting rules. Glamour handles all of this correctly with terminal-aware line wrapping.

## Common Pitfalls

### Pitfall 1: Glamour Adds Trailing Newlines and Padding
**What goes wrong:** Glamour's rendered output includes leading/trailing whitespace and margin padding that conflicts with the app's own layout.
**Why it happens:** Glamour's StyleConfig has `Document.Margin` set to non-zero by default.
**How to avoid:** When creating the custom StyleConfig, set `Document.Margin` to 0. Also trim trailing whitespace from rendered output if needed.
**Warning signs:** Double padding, unexpected blank lines in preview.

### Pitfall 2: Width Mismatch Between Glamour and Viewport
**What goes wrong:** Glamour wraps at one width, but the viewport/pane is a different width, causing ugly line breaks or cut-off text.
**Why it happens:** Glamour's WithWordWrap width must match the available content width minus any padding/borders.
**How to avoid:** Compute the actual inner content width (accounting for pane borders and padding) and pass that to glamour. Re-create the renderer when the terminal resizes.
**Warning signs:** Lines wrapping mid-word, horizontal scrolling not working.

### Pitfall 3: Template Execution Panics on Missing Fields
**What goes wrong:** If a template references `{{.Foo}}` but the values map does not have key "Foo", `text/template` produces an error (or panics in strict mode).
**Why it happens:** `text/template` has `Option("missingkey=error")` by default for struct fields, but maps return zero value.
**How to avoid:** Since we use `map[string]string`, missing keys produce empty string (not error). But validate that all extracted placeholders have values before executing. Consider using `template.New("").Option("missingkey=zero")` explicitly.
**Warning signs:** Empty spots in generated body where placeholder values should be.

### Pitfall 4: JSON Store Compatibility
**What goes wrong:** The JSON-based `Store` in store.go also implements `TodoStore` but its `Todo` struct needs `Body` too.
**Why it happens:** The interface change affects both backends.
**How to avoid:** Update `Store.Add()`, `Store.Update()` signatures to include body, and the JSON `Todo` struct. The JSON store may be vestigial (main.go uses SQLite), but the interface must be consistent. Either update the JSON store or remove it.
**Warning signs:** Compile errors on interface satisfaction check `var _ TodoStore = (*Store)(nil)`.

### Pitfall 5: Glamour Import Pulls in Many Dependencies
**What goes wrong:** Glamour has a large dependency tree (goldmark, chroma for syntax highlighting, etc.).
**Why it happens:** Full markdown rendering requires a complete parser and syntax highlighter.
**How to avoid:** This is acceptable for the feature. Just be aware of the binary size increase. No mitigation needed.
**Warning signs:** `go.sum` grows significantly. This is expected.

### Pitfall 6: Re-creating Glamour Renderer Per Render Cycle
**What goes wrong:** Creating a new glamour.TermRenderer on every View() call is expensive.
**Why it happens:** Temptation to inline renderer creation in the view function.
**How to avoid:** Create the renderer once (in the preview model constructor or on theme/size change) and cache it. Only recreate on width change or theme change.
**Warning signs:** Sluggish preview rendering, visible latency.

## Code Examples

### Creating a glamour renderer with custom theme colors
```go
// Source: glamour pkg.go.dev docs + styles package docs
import (
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/glamour/ansi"
    "github.com/charmbracelet/glamour/styles"
    "github.com/antti/todo-calendar/internal/theme"
)

func rendererForTheme(t theme.Theme, width int) (*glamour.TermRenderer, error) {
    // Pick base style
    base := styles.DarkStyleConfig
    // Could pick LightStyleConfig for "light" theme

    // Override heading colors to match app accent
    accentStr := string(t.AccentFg)
    if accentStr != "" {
        base.H1.StylePrimitive.Color = &accentStr
        base.H2.StylePrimitive.Color = &accentStr
        base.H3.StylePrimitive.Color = &accentStr
    }

    // Zero out document margin (we handle padding in lipgloss)
    zero := uint(0)
    base.Document.Margin = &zero

    return glamour.NewTermRenderer(
        glamour.WithStyles(base),
        glamour.WithWordWrap(width),
    )
}
```

### Extracting and prompting for template placeholders
```go
// Source: text/template/parse stdlib docs

// 1. User selects template with content:
//    "# {{.ProjectName}}\n\nDue: {{.DueDate}}\n\n## Tasks\n- [ ] {{.FirstTask}}"

// 2. Extract placeholders:
names, _ := ExtractPlaceholders(template.Content)
// names = ["ProjectName", "DueDate", "FirstTask"]

// 3. Prompt user for each (in placeholderInputMode):
// input.Placeholder = "ProjectName"  -> user types "My Project"
// input.Placeholder = "DueDate"      -> user types "2026-03-01"
// input.Placeholder = "FirstTask"    -> user types "Set up repo"

// 4. Execute template:
values := map[string]string{
    "ProjectName": "My Project",
    "DueDate":     "2026-03-01",
    "FirstTask":   "Set up repo",
}
body, _ := ExecuteTemplate(template.Content, values)
// body = "# My Project\n\nDue: 2026-03-01\n\n## Tasks\n- [ ] Set up repo"

// 5. Create todo with body:
store.AddWithBody(todoText, date, body)
```

### Viewport-based preview overlay
```go
// Source: charmbracelet/bubbles viewport pattern
import "github.com/charmbracelet/bubbles/viewport"

type PreviewModel struct {
    viewport viewport.Model
    title    string
    width    int
    height   int
}

func NewPreview(title, renderedMarkdown string, width, height int) PreviewModel {
    vp := viewport.New(width, height-2) // -2 for title + hint line
    vp.SetContent(renderedMarkdown)
    return PreviewModel{
        viewport: vp,
        title:    title,
        width:    width,
        height:   height,
    }
}
```

### SQLite migration v2 for templates table
```go
// Source: existing migrate() pattern in store/sqlite.go
if version < 2 {
    _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS templates (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        name       TEXT    NOT NULL UNIQUE,
        content    TEXT    NOT NULL,
        created_at TEXT    NOT NULL
    )`)
    if err != nil {
        return fmt.Errorf("create templates table: %w", err)
    }
    _, err = s.db.Exec(`PRAGMA user_version = 2`)
    if err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

### TodoStore interface extension
```go
// Minimal additions to support body and templates
type TodoStore interface {
    // Existing methods unchanged...
    Add(text string, date string) Todo
    // ...

    // New body-aware methods
    UpdateBody(id int, body string)

    // Template methods
    AddTemplate(name, content string) (Template, error)
    ListTemplates() []Template
    FindTemplate(id int) *Template
    DeleteTemplate(id int)
}
```
Note: `Add()` could gain a `body` parameter, or we could keep `Add()` as-is and have a separate `AddWithBody(text, date, body string) Todo` method to avoid breaking the existing interface for consumers that don't need body. The decision depends on whether we want a single method or avoid signature changes. Given this is internal code with 2 implementations, changing the signature is fine.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| glamour v1 (v0.x) | glamour v2 pre-release | Oct 2025 | v2 is NOT stable; stick with v0.10.0 for now |
| Manual ANSI rendering | glamour | 2020+ | glamour is the standard for Go terminal markdown |
| Custom template systems | text/template stdlib | Always | Stdlib is correct for {{.Var}} placeholder patterns |

**Deprecated/outdated:**
- glamour v2: Pre-release, not tagged, not stable. Use v0.10.0.

## Open Questions

1. **Should `Add()` signature change or add a new `AddWithBody()` method?**
   - What we know: Both JSON and SQLite stores implement `Add(text, date)`. Adding body parameter changes the interface.
   - What's unclear: Whether the JSON store is still used or can be removed.
   - Recommendation: Change `Add()` to `Add(text, date, body string)` since the JSON store exists only as a compile-time interface check and main.go uses SQLite exclusively. Update both implementations.

2. **Where should template management UI live?**
   - What we know: Templates are a store-level concern. Creating/listing/deleting templates needs UI.
   - What's unclear: Should template management be in settings, a new overlay, or inline in todolist?
   - Recommendation: Template management (create/delete) can be a new overlay accessed via a keybinding. Template *selection* during todo creation happens inline in the todolist component's modal flow.

3. **How should body editing work before Phase 16 (External Editor)?**
   - What we know: Phase 16 adds $EDITOR integration specifically for body editing.
   - What's unclear: Whether Phase 15 needs any body editing at all, or just viewing.
   - Recommendation: Phase 15 only needs body *viewing* (preview pane) and body *creation* (via template). Direct editing of existing bodies is deferred to Phase 16's external editor.

4. **Preview pane: overlay or split pane?**
   - What we know: Settings and search use full-screen overlays. The main view is already split (calendar + todolist).
   - What's unclear: Whether preview should be a third pane, replace the calendar pane, or be a full-screen overlay.
   - Recommendation: Full-screen overlay (like search). It is simpler, allows full width for markdown rendering, and follows existing patterns.

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/github.com/charmbracelet/glamour` - Full API surface, TermRendererOption types
- `pkg.go.dev/github.com/charmbracelet/glamour/ansi` - StyleConfig, StylePrimitive, StyleBlock struct definitions
- `pkg.go.dev/github.com/charmbracelet/glamour/styles` - DarkStyleConfig, LightStyleConfig, all exported style configs
- `pkg.go.dev/text/template/parse` - Node types, FieldNode.Ident for placeholder extraction
- `pkg.go.dev/text/template` - Template.Execute with map data, Must helper

### Secondary (MEDIUM confidence)
- GitHub charmbracelet/glamour README - Version v0.10.0 (Apr 2025), basic usage examples
- GitHub charmbracelet/wizard-tutorial - Multi-step input pattern with Input interface
- GitHub charmbracelet/bubbletea credit-card-form example - Sequential text input navigation pattern

### Tertiary (LOW confidence)
- glamour v2 status - Pre-release observation from pkg.go.dev timestamp (Oct 2025); may have changed

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - glamour and text/template are well-documented, stable, and widely used
- Architecture: HIGH - Patterns directly follow existing codebase conventions (overlays, modes, TodoStore interface)
- Pitfalls: HIGH - Based on direct API documentation review (StyleConfig margins, TermRenderer lifecycle)
- Template extraction: HIGH - text/template/parse API verified via official pkg.go.dev docs

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (30 days -- stable libraries)
