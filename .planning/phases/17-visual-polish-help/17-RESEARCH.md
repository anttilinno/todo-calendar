# Phase 17: Visual Polish & Help - Research

**Researched:** 2026-02-07
**Domain:** Bubble Tea TUI rendering -- lipgloss styling, help bar, todolist view
**Confidence:** HIGH

## Summary

Phase 17 modifies two areas: (1) the todo list visual rendering in `internal/todolist/model.go` and `internal/todolist/styles.go`, and (2) the help bar system across `internal/todolist` (bindings source) and `internal/app/model.go` (help bar assembly and rendering). Both areas are pure view-layer changes -- no store, config, or data model changes are needed.

The todo list currently renders items with zero vertical spacing (single `\n` between items), bold accent-colored section headers with no separators, and dates/status as inline text with only muted-color differentiation. The help bar currently shows all 15 todo-pane bindings + 4 app bindings in normal mode, using `bubbles/help` ShortHelp rendering. No `?` toggle or `ShowAll` mechanism exists.

**Primary recommendation:** Split work into two plans -- (1) VIS-01/VIS-02/VIS-03 visual polish changes in todolist package, (2) HELP-01/HELP-02/HELP-03 help bar rework across todolist and app packages.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charmbracelet/lipgloss | v1.1.1 (pre-release) | Terminal style rendering | Already used throughout; provides Foreground, Bold, Italic, Underline, MarginBottom, PaddingBottom, Border methods |
| charmbracelet/bubbles/help | v0.21.1 | Help bar rendering | Already used in app; has `ShowAll` toggle, `ShortHelp`/`FullHelp` dual-mode, width-aware truncation with ellipsis |
| charmbracelet/bubbles/key | v0.21.1 | Key binding definitions | Already used; `SetEnabled()` hides disabled bindings from help, `WithHelp()` sets display text |

### Supporting

No new dependencies needed. All changes use existing libraries.

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| lipgloss MarginBottom for spacing | Manual `\n` insertion | MarginBottom is cleaner but operates on rendered strings; extra `\n` is simpler and more predictable in a line-by-line builder pattern. **Use extra `\n` for spacing between todo items.** |
| bubbles/help FullHelp columns | Custom multi-line help rendering | FullHelp renders columns side-by-side; for the `?` expanded view a single-column list is clearer. **Use FullHelp with a single group (current pattern already does this).** |
| Unicode box-drawing separator lines | Simple `---` or empty line for headers | Box drawing requires width calculation. **Use a thin line of `---` chars or styled horizontal rule under headers.** |

## Architecture Patterns

### Current Todo Rendering Flow

```
app.View()
  -> todolist.View()
    -> visibleItems() returns []visibleItem (headerItem, todoItem, emptyItem)
    -> for each item:
       headerItem:  styles.SectionHeader.Render(label) + "\n"
       emptyItem:   "  " + styles.Empty.Render(label) + "\n"
       todoItem:    renderTodo(&b, todo, selected) -> cursor + checkbox + text + "\n"
```

### Current Help Bar Flow

```
app.View()
  -> currentHelpKeys() returns helpKeyMap
     todoPane normal mode: 15 todo bindings + Tab + Settings + Search + Quit = 19 bindings
     todoPane input mode:  Confirm + Cancel = 2 bindings
     calendarPane:         PrevMonth + NextMonth + ToggleWeek + Tab + Settings + Search + Quit = 7 bindings
  -> help.View(keyMap) calls ShortHelp() -> renders inline with " . " separators
     (ShowAll is never set, so FullHelp is never called)
```

### Pattern 1: Vertical Spacing Between Todo Items (VIS-01)

**What:** Add an empty line after each todo item to create breathing room.
**When to use:** In `renderTodo()` -- append `"\n"` twice instead of once, or add a blank line between items in the View loop.
**Implementation note:** Only add spacing between todo items, not after the last item before mode-specific UI. Track whether we just rendered a todo item and conditionally add the extra newline before the next item.

```go
// In View(), after renderTodo:
case todoItem:
    isSelected := selectableIdx < len(selectable) && selectableIdx == m.cursor && m.focused
    m.renderTodo(&b, item.todo, isSelected)
    selectableIdx++
    // VIS-01: breathing room between items
    b.WriteString("\n")
```

**Concern:** Extra blank lines consume vertical space. With many todos, items may overflow the pane. This is acceptable -- the pane is already fixed-height and has no scrolling, so the visual trade-off is worthwhile for readability.

### Pattern 2: Section Header Distinction (VIS-02)

**What:** Make section headers (month name, "Floating") stand out more from todo items.
**When to use:** In View() headerItem rendering.
**Options (in preference order):**

1. **Separator line below header + empty line above (except first):** Add a thin `---` line rendered in muted color after each header label. Add empty line before headers that aren't the first item. This creates clear visual sections.

```go
case headerItem:
    if idx > 0 {
        b.WriteString("\n") // spacing before non-first headers
    }
    b.WriteString(m.styles.SectionHeader.Render(item.label))
    b.WriteString("\n")
    b.WriteString(m.styles.Separator.Render("──────────"))
    b.WriteString("\n")
```

2. **Underline style on header text:** Use `lipgloss.NewStyle().Bold(true).Underline(true).Foreground(t.AccentFg)`. Simpler but less visually striking.

**Recommendation:** Use option 1 (separator line). It creates the strongest visual grouping. Requires adding a `Separator` style to the Styles struct.

### Pattern 3: Date and Status Visual Differentiation (VIS-03)

**What:** Make checkbox status and dates visually distinct from todo text.
**Current state:** Checkbox is inline plaintext `[ ] ` or `[x] `. Date is rendered with `styles.Date` (muted foreground). When done, entire line including checkbox gets `styles.Completed` (strikethrough + muted).
**Improvement options:**

1. **Styled checkbox:** Render `[x]` with a distinct color (e.g., CompletedCountFg green for done, AccentFg for pending). Keep checkbox separate from completed strikethrough.
2. **Right-aligned date:** This is complex because lipgloss needs to know the available width. Simpler: keep date at end but wrap in parentheses or brackets for visual separation.
3. **Distinct date styling:** Make dates more visible with italic or a dedicated color that's distinct from both NormalFg and MutedFg.

**Recommendation:** Style the checkbox separately from the text. Use accent color for unchecked `[ ]`, completed color for `[x]`. Render date with slightly more visible styling (italic or different color). This requires changes to `renderTodo()` and possibly a new style (`Checkbox` style) in the Styles struct.

```go
// Styled checkbox rendering
var checkStyle lipgloss.Style
if t.Done {
    checkStyle = m.styles.CheckboxDone // green/completed color
} else {
    checkStyle = m.styles.Checkbox // accent color
}
check := checkStyle.Render("[ ] ")
if t.Done {
    check = checkStyle.Render("[x] ")
}
```

### Pattern 4: Help Bar with ? Toggle (HELP-01, HELP-02, HELP-03)

**What:** Reduce normal mode help to max 5 keys, add `?` to expand full help.
**Mechanism:** The `bubbles/help` library already supports this via `help.Model.ShowAll`. When `ShowAll` is false, `help.View()` calls `ShortHelp()`. When true, it calls `FullHelp()`.

**Implementation approach:**

1. **Add `?` keybinding** to both todolist.KeyMap and app.KeyMap (or just app since it's a global toggle).
2. **Track expanded state** in the app Model: `helpExpanded bool`.
3. **Toggle on `?`:** Set `m.help.ShowAll = !m.help.ShowAll` or track with `helpExpanded`.
4. **ShortHelp returns 5 keys** for normal mode: `a/add`, `x/done`, `d/delete`, `e/edit`, `?/more`.
5. **FullHelp returns all bindings** organized in groups.
6. **Input modes:** Always return only `Enter/Esc` regardless of expanded state (already done in `todolist.HelpBindings()` for non-normal modes).

**Key architectural decision:** Where does `?` live?
- Option A: App-level keybinding. `?` toggles `help.ShowAll` in the app model. The `currentHelpKeys()` method returns different bindings based on `help.ShowAll`. This is cleanest since the help bar belongs to the app.
- Option B: Todolist-level keybinding. More complex, would require the todolist to know about help state.

**Recommendation:** Option A. Add `?` as `app.KeyMap.Help` binding. Handle in `app.Update()`. The todolist's `HelpBindings()` stays as is for the full list; the app slices it to 5 items for short help.

**Revised currentHelpKeys() approach:**

```go
func (m Model) currentHelpKeys() helpKeyMap {
    // ... overlay checks unchanged ...

    var bindings []key.Binding
    switch m.activePane {
    case calendarPane:
        calKeys := m.calendar.Keys()
        bindings = append(bindings, calKeys.PrevMonth, calKeys.NextMonth, calKeys.ToggleWeek)
    case todoPane:
        todoBindings := m.todoList.HelpBindings()
        if !m.help.ShowAll && !m.todoList.IsInputting() {
            // Short mode: pick top 5 most useful + ?
            bindings = shortTodoBindings(todoBindings)
        } else {
            bindings = todoBindings
        }
    }
    bindings = append(bindings, m.keys.Tab, m.keys.Settings, m.keys.Search, m.keys.Quit)
    if !m.todoList.IsInputting() {
        bindings = append(bindings, m.keys.Help) // ? more/less
    }
    return helpKeyMap{bindings: bindings}
}
```

**Alternative approach (simpler):** Don't use `help.ShowAll` at all. Instead, have `todolist.HelpBindings()` return the short list (5 keys) always, and introduce a new method `todolist.FullHelpBindings()` for the expanded view. The app toggles `helpExpanded` and calls the appropriate method.

**Recommendation:** Use the simpler approach. `HelpBindings()` returns the short list (max 5). New `AllHelpBindings()` returns the full list. App tracks `helpExpanded` bool and chooses which to call. This avoids coupling to `help.ShowAll` and keeps the help bar always using `ShortHelpView` (single line) for the compact form, while for expanded form we can render multi-line help manually or use `FullHelpView`.

### Pattern 5: Expanded Help Rendering

**What:** When `?` is pressed, show all keybindings.
**Options:**

1. **Use `help.ShowAll = true`:** Renders via `FullHelpView` which creates side-by-side columns. Works if bindings are grouped into multiple `[]key.Binding` slices.
2. **Custom multi-line rendering above the help bar:** Render a "help overlay" or expanded section. More control but more code.

**Recommendation:** Use `help.ShowAll = true`. It's built-in and works with the existing help infrastructure. Group bindings logically: navigation group, CRUD group, advanced group. The `FullHelp()` method in helpKeyMap would return `[][]key.Binding` with 2-3 columns.

**Expanded help sizing concern:** The `FullHelpView` renders multi-line columns. This changes `helpHeight` from 1 to potentially 5-8 lines. The app `View()` currently hardcodes `helpHeight := 1`. This must become dynamic based on `help.ShowAll`.

```go
helpBar := m.help.View(m.currentHelpKeys())
helpHeight := lipgloss.Height(helpBar)
contentHeight := m.height - helpHeight - frameV
```

**Implementation:** Calculate `helpBar` first, then measure its height, then use remaining height for content panes. This means moving `helpBar` calculation before pane sizing in `View()`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Help bar layout | Custom help string formatting | `bubbles/help` ShortHelpView + FullHelpView | Width-aware truncation, ellipsis, consistent styling |
| Key enabled/disabled toggling | Manual if/else for help filtering | `key.Binding.SetEnabled()` | Help renderer auto-skips disabled bindings |
| Style management | Inline lipgloss calls | Styles struct with constructor DI (existing pattern) | Consistent theming, theme-swappable |

**Key insight:** The existing `bubbles/help` library already handles the hard problems (width truncation, ellipsis, dual-mode rendering). The phase just needs to wire the toggle and adjust which bindings are passed.

## Common Pitfalls

### Pitfall 1: Vertical Space Overflow

**What goes wrong:** Adding blank lines between todo items doubles vertical consumption. With 10+ todos in a section, items may get cut off since the pane has fixed height and no scrolling.
**Why it happens:** The todo pane height is `contentHeight` calculated from terminal size minus help bar. There's no scroll mechanism.
**How to avoid:** Accept the trade-off for now. The visual polish is worth it. Phase 18 (full-pane editing) will redesign the pane layout anyway. If needed, add a simple vertical truncation guard that stops rendering todos when the pane height is reached.
**Warning signs:** Long todo lists get cut off at the bottom of the pane.

### Pitfall 2: Help Height Changes Break Layout

**What goes wrong:** When expanded help is shown via `ShowAll = true`, the help bar grows from 1 line to 5+ lines, pushing content panes off-screen or causing flicker.
**Why it happens:** Current `helpHeight := 1` is hardcoded in `app.View()`.
**How to avoid:** Calculate help bar content first, measure height, then allocate remaining space to content panes. This requires reordering the View() function.
**Warning signs:** Content pane shrinks when help expands; layout breaks on small terminals.

### Pitfall 3: ? Keybinding Conflicts with Input Modes

**What goes wrong:** Pressing `?` while in input mode (typing todo text) inserts a literal `?` instead of toggling help.
**Why it happens:** Input modes forward all key events to the textinput component.
**How to avoid:** Only handle `?` in normal mode. The existing `isInputting` guard in `app.Update()` already prevents non-quit keys from being handled at app level during input. The `?` toggle should be inside the same guard.
**Warning signs:** `?` doesn't work in input mode (which is correct), or `?` triggers help toggle instead of inserting the character (which would be a bug).

### Pitfall 4: Separator Width Hardcoding

**What goes wrong:** A hardcoded separator width (e.g., `"──────────"`) looks wrong on narrow or wide terminals.
**Why it happens:** The todolist doesn't currently track its inner width precisely.
**How to avoid:** Use a short fixed-width separator (e.g., 10-20 chars) that works at any width. Avoid trying to fill the full pane width since the todolist doesn't receive its rendered width reliably. The pane border and padding are managed by the app's lipgloss styles.
**Warning signs:** Separator extends beyond pane or looks too short.

### Pitfall 5: Completed Todos Lose Styled Checkbox

**What goes wrong:** When a todo is done, `m.styles.Completed.Render(check + text)` applies strikethrough to everything including the checkbox. If the checkbox is styled separately (VIS-03), the completed style would override the checkbox styling.
**Why it happens:** Lipgloss styles applied to a string that already contains ANSI escape codes can interact unpredictably.
**How to avoid:** When rendering completed todos, apply the checkbox style first, then concatenate with the completed-styled text. Don't wrap both in a single Completed.Render(). Test visually.
**Warning signs:** Checkbox color disappears on completed items, or strikethrough doesn't apply to text.

## Code Examples

### Example 1: Current renderTodo (what exists)

```go
// Source: internal/todolist/model.go:850-880
func (m Model) renderTodo(b *strings.Builder, t *store.Todo, selected bool) {
    if selected {
        b.WriteString(m.styles.Cursor.Render("> "))
    } else {
        b.WriteString("  ")
    }
    check := "[ ] "
    if t.Done {
        check = "[x] "
    }
    text := t.Text
    if t.HasBody() {
        text += " " + m.styles.BodyIndicator.Render("[+]")
    }
    if t.HasDate() {
        text += " " + m.styles.Date.Render(config.FormatDate(t.Date, m.dateLayout))
    }
    if t.Done {
        b.WriteString(m.styles.Completed.Render(check + text))
    } else {
        b.WriteString(check + text)
    }
    b.WriteString("\n")
}
```

### Example 2: Proposed renderTodo with VIS-03 (styled checkbox)

```go
func (m Model) renderTodo(b *strings.Builder, t *store.Todo, selected bool) {
    // Cursor
    if selected {
        b.WriteString(m.styles.Cursor.Render("> "))
    } else {
        b.WriteString("  ")
    }

    // Styled checkbox (VIS-03)
    if t.Done {
        b.WriteString(m.styles.CheckboxDone.Render("[x]"))
    } else {
        b.WriteString(m.styles.Checkbox.Render("[ ]"))
    }
    b.WriteString(" ")

    // Text (with completed styling applied only to text)
    text := t.Text
    if t.Done {
        b.WriteString(m.styles.Completed.Render(text))
    } else {
        b.WriteString(text)
    }

    // Body indicator
    if t.HasBody() {
        b.WriteString(" " + m.styles.BodyIndicator.Render("[+]"))
    }

    // Date (VIS-03: visually distinct)
    if t.HasDate() {
        b.WriteString(" " + m.styles.Date.Render(config.FormatDate(t.Date, m.dateLayout)))
    }

    b.WriteString("\n")
}
```

### Example 3: Short Help Bindings (HELP-01)

```go
// In todolist.HelpBindings() - return short list in normal mode
func (m Model) HelpBindings() []key.Binding {
    if m.mode != normalMode {
        return []key.Binding{m.keys.Confirm, m.keys.Cancel}
    }
    // HELP-01: max 5 most-used keys
    return []key.Binding{m.keys.Add, m.keys.Toggle, m.keys.Delete, m.keys.Edit, m.keys.Filter}
}

// New method for full help
func (m Model) AllHelpBindings() []key.Binding {
    if m.mode != normalMode {
        return []key.Binding{m.keys.Confirm, m.keys.Cancel}
    }
    return []key.Binding{m.keys.Up, m.keys.Down, m.keys.MoveUp, m.keys.MoveDown,
        m.keys.Add, m.keys.AddDated, m.keys.Edit, m.keys.EditDate,
        m.keys.Toggle, m.keys.Delete, m.keys.Filter,
        m.keys.Preview, m.keys.OpenEditor, m.keys.TemplateUse, m.keys.TemplateCreate}
}
```

### Example 4: App-level ? Toggle (HELP-03)

```go
// In app.KeyMap:
Help: key.NewBinding(
    key.WithKeys("?"),
    key.WithHelp("?", "help"),
),

// In app.Update() normal-mode key handling:
case key.Matches(msg, m.keys.Help) && !isInputting:
    m.help.ShowAll = !m.help.ShowAll
    return m, nil

// In app.currentHelpKeys():
case todoPane:
    if m.help.ShowAll {
        bindings = append(bindings, m.todoList.AllHelpBindings()...)
    } else {
        bindings = append(bindings, m.todoList.HelpBindings()...)
    }
```

### Example 5: Dynamic Help Height in View()

```go
// In app.View() - calculate help first, then size content
m.help.Width = m.width
helpBar := m.help.View(m.currentHelpKeys())
helpHeight := lipgloss.Height(helpBar)
if helpHeight < 1 {
    helpHeight = 1
}

frameH, frameV := m.styles.Pane(true).GetFrameSize()
contentHeight := m.height - helpHeight - frameV
```

### Example 6: Styles Struct Additions

```go
type Styles struct {
    SectionHeader lipgloss.Style
    Separator     lipgloss.Style  // NEW: thin line under section headers
    Completed     lipgloss.Style
    Cursor        lipgloss.Style
    Checkbox      lipgloss.Style  // NEW: unchecked checkbox styling
    CheckboxDone  lipgloss.Style  // NEW: checked checkbox styling
    Date          lipgloss.Style
    Empty         lipgloss.Style
    BodyIndicator lipgloss.Style
}

func NewStyles(t theme.Theme) Styles {
    return Styles{
        SectionHeader: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
        Separator:     lipgloss.NewStyle().Foreground(t.MutedFg),
        Completed:     lipgloss.NewStyle().Strikethrough(true).Foreground(t.CompletedFg),
        Cursor:        lipgloss.NewStyle().Foreground(t.AccentFg),
        Checkbox:      lipgloss.NewStyle().Foreground(t.AccentFg),
        CheckboxDone:  lipgloss.NewStyle().Foreground(t.CompletedCountFg),
        Date:          lipgloss.NewStyle().Faint(true).Foreground(t.MutedFg),
        Empty:         lipgloss.NewStyle().Foreground(t.EmptyFg),
        BodyIndicator: lipgloss.NewStyle().Foreground(t.MutedFg),
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single help mode | `help.ShowAll` toggle between ShortHelp/FullHelp | bubbles v0.14+ | Built-in dual-mode help; no custom code needed |
| Manual help string building | `help.Model.View(keyMap)` with KeyMap interface | bubbles v0.14+ | Width-aware, auto-truncating help rendering |

**Deprecated/outdated:**
- `help.NewModel()` is deprecated in favor of `help.New()` (already using `help.New()` in this project).

## Open Questions

1. **Exact 5 keys for HELP-01**
   - What we know: Requirements suggest `a/add, x/done, d/delete, e/edit, ?/more` as the 5 keys.
   - What's unclear: Should `?/more` count as one of the 5, or should there be 5 action keys plus `?`? The success criteria says "at most 5 key bindings" -- so `?` would make it 6 total in the help bar (plus app-level Tab/Settings/Search/Quit).
   - Recommendation: Show 5 todo action keys + `?` in the todo section. The app adds Tab/Settings/Search/Quit. Total visible: ~10 bindings. This is much better than current 19. If the requirement strictly means 5 total in the todo section, drop one action key (filter or preview) and include `?` as one of the 5.

2. **Calendar pane help when ? is pressed**
   - What we know: The `?` toggle is global (app level).
   - What's unclear: When calendar pane is focused and `?` is pressed, should it show all calendar bindings too? Calendar only has 3 bindings, so there's no reduction needed.
   - Recommendation: `?` toggle is primarily for todo pane. Calendar pane can show all its bindings always. `?` still works as a toggle but has no visible effect when calendar is focused.

3. **FullHelp column layout**
   - What we know: `bubbles/help` FullHelpView renders `[][]key.Binding` as side-by-side columns.
   - What's unclear: How many groups/columns look best for 15+ bindings?
   - Recommendation: 3 columns: Navigation (Up/Down/MoveUp/MoveDown), CRUD (Add/AddDated/Edit/EditDate/Toggle/Delete), Advanced (Filter/Preview/OpenEditor/TemplateUse/TemplateCreate). Plus app-level bindings in a 4th column.

## Sources

### Primary (HIGH confidence)
- `internal/todolist/model.go` lines 142-148 -- current HelpBindings implementation
- `internal/todolist/model.go` lines 768-880 -- current View and renderTodo implementation
- `internal/todolist/styles.go` -- current Styles struct (6 styles)
- `internal/app/model.go` lines 329-352 -- current currentHelpKeys and View
- `internal/app/keys.go` -- current app KeyMap (4 bindings: Quit, Tab, Settings, Search)
- `internal/theme/theme.go` -- Theme struct with 16 color roles
- `charmbracelet/bubbles@v0.21.1/help/help.go` -- help.Model with ShowAll, ShortHelpView, FullHelpView

### Secondary (MEDIUM confidence)
- None needed -- all findings are from source code inspection

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all libraries already in use, no new dependencies
- Architecture: HIGH -- direct source code inspection of all affected files
- Pitfalls: HIGH -- identified from concrete code analysis (hardcoded helpHeight, input mode conflicts, style nesting)

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (stable; no external dependencies changing)
