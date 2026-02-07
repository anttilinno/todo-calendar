# Phase 24: Unified Add Form - Research

**Researched:** 2026-02-07
**Domain:** Bubble Tea TUI form with multi-field input (textinput + textarea)
**Confidence:** HIGH

## Summary

Phase 24 transforms the existing single-field add flow (`inputMode`) into a full-pane multi-field form matching the established `editMode` pattern. The codebase already has all the building blocks: the `editMode` in `todolist/model.go` implements a 3-field form (Title, Date, Body) with Tab cycling, Enter/Ctrl+D save semantics, and full-pane rendering via `editView()`. Phase 24 extends `inputMode` to become functionally equivalent to `editMode` but for creating new todos instead of editing existing ones, and adds a 4th field (Template) for Phase 25 integration.

The key architectural insight is that the current `inputMode` is a simplified single-field mode while `editMode` already does exactly what the unified add form needs. The implementation strategy should either: (a) repurpose `inputMode` to use the same multi-field pattern as `editMode`, or (b) introduce a new `addMode` that shares the edit mode's field cycling and rendering logic. Option (a) is simpler -- modify `inputMode` to show all 4 fields and reuse the existing `editView()` rendering with minimal changes.

The store API already supports the operations needed: `Add(text, date)` creates a todo with optional date, and `UpdateBody(id, body)` sets the body. The Template field in Phase 24 is a UI placeholder for Tab cycling only -- actual template picker behavior is Phase 25 (ADD-03/ADD-04).

**Primary recommendation:** Extend `inputMode` to a 4-field form (Title, Date, Body, Template) reusing the existing `editView()` rendering and `editMode`'s field-cycling pattern, rather than creating a new mode from scratch.

## Standard Stack

Already in place -- no new dependencies needed.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charmbracelet/bubbles | v0.21.1 | textinput, textarea components | Already used for edit mode fields |
| charmbracelet/bubbletea | v1.3.10 | Elm architecture TUI framework | Core framework |
| charmbracelet/lipgloss | v1.1.1 | Styling and layout | Core styling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| internal/config | n/a | Date format parsing/display | For date field validation (ParseUserDate, FormatDate) |
| internal/store | n/a | TodoStore interface | Add(text, date), UpdateBody(id, body) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Extending inputMode | New addMode constant | New mode adds complexity; extending is simpler since inputMode already exists and editMode proves the pattern works |
| textinput for Template | textarea for Template | Template field just needs a single-line display showing selected template name; textinput is sufficient. In Phase 25, it becomes a trigger field (Enter opens picker) |

**Installation:** No new packages needed.

## Architecture Patterns

### Current Mode Architecture (todolist/model.go)
```
Modes: normalMode | inputMode | editMode | filterMode | templateNameMode | templateContentMode

inputMode: Single textinput field -> store.Add(text, "")  [floating only]
editMode:  3 fields (title+date+body) with Tab cycling -> store.Update() + UpdateBody()
```

### Target Architecture After Phase 24
```
inputMode: 4 fields (title+date+body+template) with Tab cycling -> store.Add(text, date) + UpdateBody()
editMode:  unchanged (3 fields: title+date+body)
```

### Pattern 1: Extend inputMode to Multi-Field Form
**What:** Reuse the field cycling pattern from `editMode` for `inputMode`, adding a 4th field
**When to use:** When the add form needs the same UX as edit mode
**Implementation approach:**
```go
// In inputMode, use editField to track active field (same as editMode):
// editField 0 = title (textinput m.input)
// editField 1 = date (textinput m.dateInput)
// editField 2 = body (textarea m.bodyTextarea)
// editField 3 = template (textinput - new field, placeholder for Phase 25)

// Tab cycles: 0 -> 1 -> 2 -> 3 -> 0
// Enter saves from fields 0,1 (title, date)
// Ctrl+D saves from fields 2,3 (body, template)
// Esc cancels from fields 0,1; goes to field 0 from fields 2,3
```

### Pattern 2: Save Logic for New Todos
**What:** The add form saves using store.Add() + store.UpdateBody(), not store.Update()
**Implementation:**
```go
func (m Model) saveAdd() (Model, tea.Cmd) {
    text := strings.TrimSpace(m.input.Value())
    if text == "" {
        return m, nil
    }

    // Parse date (empty = floating, filled = dated)
    date := strings.TrimSpace(m.dateInput.Value())
    isoDate := ""
    if date != "" {
        var err error
        isoDate, err = config.ParseUserDate(date, m.dateLayout)
        if err != nil {
            // Invalid date - focus date field
            m.editField = 1
            m.input.Blur()
            return m, m.dateInput.Focus()
        }
    }

    // Create the todo
    todo := m.store.Add(text, isoDate)

    // Set body if non-empty
    body := m.bodyTextarea.Value()
    if strings.TrimSpace(body) != "" {
        m.store.UpdateBody(todo.ID, body)
    }

    // Reset all fields and return to normal mode
    m.mode = normalMode
    m.input.Blur()
    m.dateInput.Blur()
    m.bodyTextarea.Blur()
    m.input.SetValue("")
    m.dateInput.SetValue("")
    m.bodyTextarea.SetValue("")
    m.editField = 0
    return m, nil
}
```

### Pattern 3: Shared editView() Rendering
**What:** The editView() method already handles rendering for inputMode and editMode; extend it for 4 fields
**Current behavior:**
```go
// editView() switches on m.mode:
// - editMode: renders Title + Date + Body labels and fields
// - inputMode: renders only Title label and field
// After Phase 24, inputMode should render Title + Date + Body + Template
```

### Pattern 4: Template Field as Placeholder
**What:** In Phase 24, the Template field is a read-only textinput showing "(no template)" or similar
**Rationale:** ADD-03 and ADD-04 are Phase 25. Phase 24 needs the field to exist for Tab cycling (ADD-02).
**Implementation:**
```go
// New field on Model:
templateInput textinput.Model  // Template selector field (Phase 25 wires picker)

// Initialize in New():
tmplInput := textinput.New()
tmplInput.Placeholder = "Press Enter to select template"
tmplInput.Prompt = "> "
tmplInput.CharLimit = 0  // Read-only in Phase 24

// In editView() for inputMode, render after Body:
// Template
// > (Press Enter to select template)
```

### Anti-Patterns to Avoid
- **Creating a separate addMode:** The existing inputMode already represents "adding a todo." Adding a new mode constant creates confusion about when to use addMode vs inputMode. Instead, extend inputMode to support multi-field.
- **Duplicating editMode logic:** The save validation (date parsing, empty title check) is identical between add and edit. Extract shared logic or pattern it consistently rather than copy-pasting.
- **Wiring template picker in Phase 24:** Phase 25 handles ADD-03/ADD-04. The Template field in Phase 24 should be inert (Tab-cyclable but not functional). Don't pre-build picker integration.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Date parsing/validation | Custom parser | config.ParseUserDate(input, layout) | Already handles ISO, EU, US formats with proper error returns |
| Multi-line body editing | Custom text handling | bubbles/textarea | Already used in editMode, handles cursor, scrolling, line wrapping |
| Field focus management | Manual focus tracking | textinput.Focus()/Blur(), textarea.Focus()/Blur() | Bubble Tea components handle cursor blink, prompt styling |
| Date format display | Format string logic | config.FormatDate() and config.DatePlaceholder() | Already wired through SetDateFormat() |

**Key insight:** Every component needed for the unified add form already exists in the codebase. The editMode already implements 90% of the required behavior. This is a wiring/restructuring task, not a build-from-scratch task.

## Common Pitfalls

### Pitfall 1: Blink Messages Not Forwarded in inputMode
**What goes wrong:** Cursor stops blinking in date/body/template fields because blink tick messages aren't forwarded to the correct component.
**Why it happens:** The existing Update() method forwards non-key messages to specific fields based on mode and editField. Currently inputMode only forwards to m.input. After adding fields, it needs to forward to whichever field is active.
**How to avoid:** Extend the blink forwarding block (lines 277-293 of model.go) to handle inputMode's multi-field case identically to editMode. The switch should be:
```go
case inputMode, editMode:
    // Forward blink/tick to active field based on editField
```
**Warning signs:** Cursor appears frozen after Tabbing to a new field.

### Pitfall 2: editField State Not Reset on Mode Entry
**What goes wrong:** Opening the add form after a previous add left editField at 2 (body), causing the cursor to start in the body field instead of title.
**Why it happens:** editField is shared between inputMode and editMode. If not explicitly reset to 0 when entering inputMode, stale state leaks.
**How to avoid:** Always set `m.editField = 0` when transitioning to inputMode (in the `key.Matches(msg, m.keys.Add)` handler).

### Pitfall 3: Esc Behavior Inconsistency Between Fields
**What goes wrong:** User expects Esc to cancel the form from any field, but the edit mode pattern has Esc from body field going to title field first.
**Why it happens:** editMode's Esc behavior is: from title/date -> cancel entirely; from body -> go to title. This is intentional to prevent accidental loss of body text.
**How to avoid:** Mirror the exact same Esc behavior from editMode: fields 0/1 -> cancel, fields 2/3 -> go to field 0. Document this in help bindings.

### Pitfall 4: Textarea Height Not Sized
**What goes wrong:** Body textarea takes up too little or too much vertical space.
**Why it happens:** The textarea's default height may not match the available pane height minus other fields.
**How to avoid:** The current editMode doesn't explicitly size the textarea height and it works. Follow the same pattern -- let the textarea use its default height. If needed, the height can be set in SetSize().

### Pitfall 5: Calendar Indicators Not Refreshed After Add
**What goes wrong:** Adding a dated todo doesn't update the calendar indicators until switching panes.
**Why it happens:** The app model's Update() calls RefreshIndicators() after every update cycle, but only when activePane is todoPane. This should work correctly since the add form runs in todoPane.
**How to avoid:** Verify that after save, the normal model.Update() flow triggers RefreshIndicators(). The existing code at line 325 of app/model.go already handles this.

### Pitfall 6: Help Bindings Not Updated for New Mode Shape
**What goes wrong:** Help bar shows wrong bindings (e.g., shows "enter: confirm" when in body field where Ctrl+D is needed).
**Why it happens:** HelpBindings() checks editField == 2 for body in editMode. It needs to also handle the template field (editField == 3) and the inputMode multi-field case.
**How to avoid:** Update HelpBindings() to handle inputMode identically to editMode -- when editField is 2 or 3, show "ctrl+d: save" and "tab: switch field"; when 0 or 1, show "enter: confirm" and "tab: switch field".

## Code Examples

### Existing inputMode Entry (current - will be modified)
```go
// Source: internal/todolist/model.go lines 360-365
case key.Matches(msg, m.keys.Add):
    m.mode = inputMode
    m.input.Placeholder = "What needs doing?"
    m.input.Prompt = "> "
    m.input.SetValue("")
    return m, m.input.Focus()
```

### Existing editMode Tab Cycling (pattern to replicate)
```go
// Source: internal/todolist/model.go lines 528-543
case key.Matches(msg, m.keys.SwitchField):
    // Cycle: title(0) -> date(1) -> body(2) -> title(0)
    switch m.editField {
    case 0:
        m.editField = 1
        m.input.Blur()
        return m, m.dateInput.Focus()
    case 1:
        m.editField = 2
        m.dateInput.Blur()
        return m, m.bodyTextarea.Focus()
    case 2:
        m.editField = 0
        m.bodyTextarea.Blur()
        return m, m.input.Focus()
    }
```

### Existing editView() Rendering (pattern to extend)
```go
// Source: internal/todolist/model.go lines 688-753
// editView() already handles both inputMode and editMode with a switch on m.mode.
// editMode renders: Title label + input, Date label + dateInput, Body label + bodyTextarea
// inputMode renders: Title label + input only
// Phase 24 changes inputMode to render all 4 fields.
```

### Existing saveEdit() Pattern (pattern for saveAdd)
```go
// Source: internal/todolist/model.go lines 578-617
// saveEdit: validates title, parses date, gets body, calls Update+UpdateBody
// saveAdd will: validate title, parse date, get body, call Add+UpdateBody
// Nearly identical except Add() vs Update() and handling new todo.ID for UpdateBody
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Separate a/A/t keybindings for float/dated/template add | Single `a` opens unified form | Phase 23 removed old keys, Phase 24 builds new form | Simpler UX, single entry point |
| inputMode = single title field | inputMode = full-pane 4-field form | Phase 24 (this phase) | Matches editMode consistency |
| Add always creates floating todo | Add form date field determines floating vs dated | Phase 24 (this phase) | No more separate flows |

**Deprecated/outdated:**
- `A` keybinding: Removed in Phase 23 (was: add dated todo)
- `t` keybinding: Removed in Phase 23 (was: template use)
- Single-field inputMode: Being replaced by multi-field form in this phase

## Open Questions

1. **Template field editField index**
   - What we know: Tab cycles Title(0) -> Date(1) -> Body(2) -> Template(3) -> Title(0). Phase 25 will wire the template picker.
   - What's unclear: Should the Template field be a textinput or a static label? If textinput, it needs CharLimit=0 or similar to be read-only.
   - Recommendation: Use a textinput with a descriptive placeholder. In Phase 25, pressing Enter on it will open the template picker. For Phase 24, it's inert but Tab-cyclable.

2. **Should inputMode's editView share code with editMode?**
   - What we know: Both modes render nearly identical forms (Title + Date + Body, with inputMode adding Template).
   - What's unclear: Whether to refactor editView() to share rendering or keep separate branches.
   - Recommendation: Keep the switch/case structure in editView() but add the Template field rendering for inputMode. The rendering code is simple string building and doesn't benefit from over-abstraction.

3. **Help bindings for 4 fields**
   - What we know: editMode already changes help based on editField. inputMode currently shows only Confirm/Cancel.
   - Recommendation: inputMode should show the same field-aware help as editMode: "tab: next field", "enter: save" or "ctrl+d: save" depending on which field is focused.

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `internal/todolist/model.go` -- full read of 864 lines, all modes and rendering
- Codebase analysis: `internal/todolist/keys.go` -- all key bindings
- Codebase analysis: `internal/todolist/styles.go` -- all styles including EditTitle, FieldLabel
- Codebase analysis: `internal/app/model.go` -- app-level message routing, overlay pattern
- Codebase analysis: `internal/store/iface.go` -- TodoStore interface (Add, UpdateBody signatures)
- Codebase analysis: `internal/store/sqlite.go` -- Add() implementation (line 192)
- Codebase analysis: `internal/config/config.go` -- ParseUserDate, FormatDate, DatePlaceholder
- Codebase analysis: `go.mod` -- bubbles v0.21.1, bubbletea v1.3.10

### Secondary (MEDIUM confidence)
- `.planning/REQUIREMENTS.md` -- ADD-01 through ADD-07 requirement definitions
- `.planning/ROADMAP.md` -- Phase 24/25 scoping and dependencies

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all libraries already in go.mod, no new dependencies
- Architecture: HIGH - editMode already implements 90% of the pattern; this is extension, not invention
- Pitfalls: HIGH - identified from direct code reading; blink forwarding and state reset are verifiable

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (stable -- all patterns are internal codebase patterns, not external library concerns)
