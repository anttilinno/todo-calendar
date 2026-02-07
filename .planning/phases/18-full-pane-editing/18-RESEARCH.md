# Phase 18: Full-Pane Editing - Research

**Researched:** 2026-02-07
**Domain:** Bubble Tea TUI layout -- replacing inline input with full-pane edit views
**Confidence:** HIGH

## Summary

Phase 18 replaces the current inline text input (appended at the bottom of the todo list) with a full-pane layout that takes over the entire right pane when adding or editing todos. This is purely a View/layout change within the existing todolist package, reusing the existing `textinput.Model` and mode state machine.

The current implementation already has the correct mode transitions (inputMode, dateInputMode, editTextMode, editDateMode) and all Update logic. The change is: when in any editing mode, `View()` should render a centered form layout instead of appending the textinput below the todo list. For dated todos, the two-step flow (title then date) should display both fields simultaneously with Tab to switch between them.

**Primary recommendation:** Modify `todolist.View()` to branch on mode and render a full-pane edit form. Add a `SetSize(w, h)` method for accurate pane dimensions. For dated todo add/edit, introduce a second `textinput.Model` so both fields display simultaneously with Tab switching. No new packages needed.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `charmbracelet/bubbles/textinput` | v0.21.1 | Single-line text input with cursor, placeholder, prompt | Already in use, handles all input mechanics |
| `charmbracelet/lipgloss` | v1.1.1 | Layout (centering, padding, styling) | Already in use for all styling |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `charmbracelet/bubbles/key` | v0.21.1 | Key binding definitions | Already in use, needed for Tab field switching |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Two `textinput.Model` for title+date | Single input with mode switching (current approach) | Current sequential flow works but violates EDIT-02 requirement of showing both fields simultaneously. Two inputs allow Tab switching and visual context. |
| `huh` form library | Two textinputs + manual layout | `huh` adds a dependency and has its own Bubble Tea model lifecycle. The scope is just 2 fields -- manual layout is simpler and consistent with codebase patterns. |

**Installation:** No new dependencies needed. Everything required is already in go.mod.

## Architecture Patterns

### Current Rendering Flow

```
app.View()
  -> computes todoInnerWidth, contentHeight
  -> todoStyle.Width(todoInnerWidth).Height(contentHeight)
  -> todoStyle.Render(m.todoList.View())
```

The todolist.View() currently returns content of arbitrary height. The lipgloss pane style in app.View() constrains and clips it. The todolist model receives `tea.WindowSizeMsg` with **full terminal dimensions** but never uses them. This is fine for a scrollable list but problematic for centered full-pane layouts.

### Recommended: Pane-Aware Dimensions

```
// In todolist package:
func (m *Model) SetSize(w, h int) {
    m.width = w
    m.height = h
}
```

The app model already computes `todoInnerWidth` and `contentHeight`. Call `m.todoList.SetSize(todoInnerWidth, contentHeight)` from app.Update (on WindowSizeMsg) and anywhere dimensions change.

This follows the established pattern from `settings.SetSize()` and `search.SetSize()`.

### Pattern 1: Mode-Branched View

**What:** The `View()` method checks the current mode and renders either the normal todo list or the full-pane edit form.
**When to use:** When the same component has two fundamentally different visual states.
**Example:**

```go
func (m Model) View() string {
    switch m.mode {
    case inputMode, dateInputMode, editTextMode, editDateMode:
        return m.editView()
    default:
        return m.normalView()
    }
}

func (m Model) editView() string {
    var b strings.Builder

    // Title
    title := "Add Todo"
    if m.mode == editTextMode || m.mode == editDateMode {
        title = "Edit Todo"
    }
    b.WriteString(m.styles.EditTitle.Render(title))
    b.WriteString("\n\n")

    // Title field
    b.WriteString(m.styles.FieldLabel.Render("Title"))
    b.WriteString("\n")
    b.WriteString(m.titleInput.View())
    b.WriteString("\n\n")

    // Date field (if applicable)
    if m.addingDated || m.mode == editDateMode || m.mode == dateInputMode {
        b.WriteString(m.styles.FieldLabel.Render("Date"))
        b.WriteString("\n")
        b.WriteString(m.dateInput.View())
        b.WriteString("\n\n")
    }

    // Minimal help
    b.WriteString(m.styles.EditHint.Render("Enter confirm | Esc cancel"))

    // Center vertically
    content := b.String()
    if m.height > 0 {
        lines := strings.Count(content, "\n") + 1
        topPad := (m.height - lines) / 2
        if topPad > 0 {
            content = strings.Repeat("\n", topPad) + content
        }
    }

    return content
}
```

### Pattern 2: Dual TextInput with Tab Switching

**What:** Two `textinput.Model` fields for title and date, with Tab cycling focus between them.
**When to use:** When the full-pane form needs to show both fields simultaneously for dated todo add/edit.
**Example:**

```go
// In Model struct, add:
//   dateInput   textinput.Model   // separate input for date in full-pane mode
//   editField   int               // 0=title, 1=date (which field is active)

// Tab handling in edit modes:
case key.Matches(msg, m.keys.SwitchField):
    if m.editField == 0 {
        m.editField = 1
        m.input.Blur()
        return m, m.dateInput.Focus()
    } else {
        m.editField = 0
        m.dateInput.Blur()
        return m, m.input.Focus()
    }
```

### Pattern 3: Vertical Centering (Established)

**What:** Content is vertically centered using `strings.Repeat("\n", topPad)` prefix.
**When to use:** For any full-pane overlay or form view.
**Example:** Already used by `settings.View()` and `search.View()`.

### Anti-Patterns to Avoid

- **Don't create a separate edit overlay component:** The edit state belongs to the todolist model. Creating a separate package like `editform` would split the state machine across packages and complicate message routing. The mode already exists in todolist -- just change what View() renders.
- **Don't modify the app-level routing for edit modes:** Unlike settings/search/preview which are truly separate overlays, editing is an intrinsic todolist concern. The app model should NOT get `showEdit` state or intercept edit messages.
- **Don't remove the existing mode constants:** The mode state machine (inputMode, dateInputMode, etc.) is correct. The change is purely in how View() renders when in those modes.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Text input with cursor | Custom rune handling | `textinput.Model` | Already used, handles unicode, cursor movement, clipboard |
| Vertical centering | Custom layout math | `strings.Repeat("\n", topPad)` pattern | Already established in settings.View() and search.View() |
| Key binding matching | Raw string comparison | `key.Matches()` with `key.Binding` | Consistent with entire codebase |

**Key insight:** This phase requires zero new libraries. It is a View() refactor with a second textinput and a Tab key binding.

## Common Pitfalls

### Pitfall 1: Pane Dimension Mismatch

**What goes wrong:** Using `m.width`/`m.height` from `tea.WindowSizeMsg` for centering produces incorrect layout because those are terminal dimensions, not pane dimensions.
**Why it happens:** The todolist receives raw terminal WindowSizeMsg. The actual pane is smaller (terminal width minus calendar pane minus borders minus padding).
**How to avoid:** Add `SetSize(w, h)` method. Call it from app.Update with computed `todoInnerWidth` and `contentHeight`. Use these stored values for centering math.
**Warning signs:** Edit form appears off-center or extends beyond pane boundaries.

### Pitfall 2: Tab Key Conflict with App-Level Pane Switching

**What goes wrong:** Tab is used at the app level for pane switching. If the todolist captures Tab for field switching during edit modes, the app's Tab handler must be suppressed.
**Why it happens:** The app checks `isInputting` before processing Tab: `key.Matches(msg, m.keys.Tab) && !isInputting`. Since `IsInputting()` already returns true for all non-normal modes, Tab is already NOT processed at the app level during input modes. So there is NO conflict -- Tab will naturally reach the todolist Update handler.
**How to avoid:** Verify that `IsInputting()` returns true for all edit modes (it does, since it checks `m.mode != normalMode`). Tab for field switching just works because the app already suppresses Tab during input.
**Warning signs:** Tab switches panes instead of fields during editing.

### Pitfall 3: Two-Step Dated Flow vs Simultaneous Display

**What goes wrong:** The current dated-todo flow is sequential: inputMode (title) -> press Enter -> dateInputMode (date). EDIT-02 requires showing BOTH fields simultaneously.
**Why it happens:** Current design was optimized for inline display where only one field can be shown at a time.
**How to avoid:** For the full-pane view, when adding a dated todo (`m.addingDated == true`), show both title and date fields immediately. Title field is focused first, Tab switches to date, Enter confirms from either field (validates both). This changes the flow: instead of two modes in sequence, use a single combined mode with field tracking.
**Warning signs:** Form shows only one field at a time, or Enter behavior is inconsistent.

### Pitfall 4: Forgetting to Set Width on textinput

**What goes wrong:** The textinput displays at default width which may be too narrow or stretch oddly.
**Why it happens:** `textinput.Model.Width` defaults to 0 (unlimited) which works for inline use but looks bad in a centered form.
**How to avoid:** Set `m.input.Width` and `m.dateInput.Width` based on `m.width` when rendering the edit form or when SetSize is called.
**Warning signs:** Input field looks oddly sized relative to the pane.

### Pitfall 5: Template Modes Not Updated

**What goes wrong:** templateSelectMode, placeholderInputMode, templateNameMode, and templateContentMode still render inline and look inconsistent.
**Why it happens:** Focus on EDIT-01 through EDIT-04 misses the template modes.
**How to avoid:** Phase 18 scope is EDIT-01 through EDIT-05 (add/edit title/date). Template modes are NOT in scope. They can continue rendering inline until a future phase addresses them. The View branching should explicitly handle only `inputMode`, `dateInputMode`, `editTextMode`, `editDateMode`.
**Warning signs:** Trying to refactor template modes leads to scope creep.

### Pitfall 6: Confirm Logic Needs Both Fields When Dated

**What goes wrong:** When both title and date fields are shown simultaneously, pressing Enter should collect values from BOTH fields, not just the focused one.
**Why it happens:** Current flow uses two separate modes where each Confirm only reads one input.
**How to avoid:** In the combined dated flow, the Confirm handler should read `m.input.Value()` for title and `m.dateInput.Value()` for date regardless of which field is focused. Validate both: title must be non-empty, date must parse correctly.
**Warning signs:** Pressing Enter on the date field loses the title, or vice versa.

## Code Examples

### Example 1: Adding SetSize to Todolist

```go
// In todolist/model.go
func (m *Model) SetSize(w, h int) {
    m.width = w
    m.height = h
}

// In app/model.go, in the View() method, BEFORE rendering:
// Also call from WindowSizeMsg handler after computing dimensions
func (m *Model) syncTodoSize() {
    frameH, frameV := m.styles.Pane(true).GetFrameSize()
    helpHeight := 1 // approximate; exact needs help render
    contentHeight := m.height - helpHeight - frameV
    calendarInnerWidth := 38
    todoInnerWidth := m.width - calendarInnerWidth - (frameH * 2)
    m.todoList.SetSize(todoInnerWidth, contentHeight)
}
```

### Example 2: Second TextInput for Date

```go
// In todolist New():
dateInput := textinput.New()
dateInput.Placeholder = "YYYY-MM-DD"
dateInput.Prompt = "Date: "
dateInput.CharLimit = 10

// In Model struct:
dateInput   textinput.Model
editField   int  // 0 = title, 1 = date
```

### Example 3: Edit Styles

```go
// In todolist/styles.go, add to Styles struct:
EditTitle  lipgloss.Style  // "Add Todo" / "Edit Todo" heading
FieldLabel lipgloss.Style  // "Title", "Date" labels
EditHint   lipgloss.Style  // "Enter confirm | Esc cancel | Tab switch"

// In NewStyles():
EditTitle:  lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
FieldLabel: lipgloss.NewStyle().Bold(true).Foreground(t.NormalFg),
EditHint:   lipgloss.NewStyle().Foreground(t.MutedFg),
```

### Example 4: HelpBindings Update for Edit Modes

```go
// In HelpBindings(), for modes with two fields:
func (m Model) HelpBindings() []key.Binding {
    if m.mode == inputMode && !m.addingDated {
        // Single field: no Tab
        return []key.Binding{m.keys.Confirm, m.keys.Cancel}
    }
    if m.mode == inputMode && m.addingDated {
        // Two fields: include Tab
        return []key.Binding{m.keys.Confirm, m.keys.Cancel, m.keys.SwitchField}
    }
    if m.mode == editDateMode {
        return []key.Binding{m.keys.Confirm, m.keys.Cancel}
    }
    // ... etc
    return []key.Binding{m.keys.Confirm, m.keys.Cancel}
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Inline input appended below list | (this phase) Full-pane centered form | Phase 18 | Cleaner, more focused editing experience |
| Sequential title then date entry | (this phase) Simultaneous two-field form | Phase 18 | User sees both fields at once, Tab switches |

**No deprecated libraries or APIs involved.** All libraries used are current versions already in go.mod.

## Open Questions

1. **Exact vertical positioning**
   - What we know: Settings and Search both center vertically using `strings.Repeat("\n", topPad)`.
   - What's unclear: Should the edit form be vertically centered in the pane, or anchored to the top third? Centering is the established pattern.
   - Recommendation: Use vertical centering consistent with settings/search overlays.

2. **Dated-add flow restructuring depth**
   - What we know: Currently uses inputMode -> dateInputMode as two sequential modes. EDIT-02 requires both fields visible.
   - What's unclear: Should we merge these into a single mode with field tracking, or keep both modes but render both fields regardless?
   - Recommendation: Keep both modes (`inputMode` for title-focused, `dateInputMode` for date-focused) but in View, when `addingDated` is true, always render both fields. The mode determines which field is focused. This minimizes changes to Update logic. Tab switches mode between inputMode and dateInputMode (or a new combined mode).

3. **Edit text + edit date as single form**
   - What we know: EDIT-03 and EDIT-04 are separate (edit title, edit date). Currently they are separate modes triggered by `e` and `E`.
   - What's unclear: Should editing show a single form with both title and date (like the add-dated flow), or keep them as separate single-field views?
   - Recommendation: Keep them as separate single-field full-pane views for now. The `e` key edits title only (one field), `E` edits date only (one field). This matches current UX expectations. If desired, a future enhancement could unify them.

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `internal/todolist/model.go` (939 lines) -- complete mode state machine, View, Update logic
- Codebase analysis: `internal/app/model.go` (453 lines) -- pane layout, dimension computation, message routing
- Codebase analysis: `internal/settings/model.go` -- full-pane overlay pattern with vertical centering and SetSize
- Codebase analysis: `internal/search/model.go` -- full-pane overlay pattern with textinput and vertical centering
- Codebase analysis: `go.mod` -- all dependencies verified, no new packages needed
- charmbracelet/bubbles textinput source -- Width field confirmed for display sizing

### Secondary (MEDIUM confidence)
- None needed -- this is an internal refactor with no external library questions

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all existing libraries
- Architecture: HIGH -- follows established patterns (SetSize, vertical centering, mode-branched View)
- Pitfalls: HIGH -- identified from direct codebase analysis (dimension mismatch, Tab conflict analysis, flow restructuring)

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (stable -- internal refactor, no external API dependencies)
