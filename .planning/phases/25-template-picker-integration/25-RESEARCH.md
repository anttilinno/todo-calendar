# Phase 25: Template Picker Integration - Research

**Researched:** 2026-02-07
**Domain:** Bubble Tea inline template picker within existing multi-field add form
**Confidence:** HIGH

## Summary

Phase 25 wires the currently inert Template field in the add form (inputMode, editField=3) to a functional template picker that lists available templates, lets the user select one, handles placeholder prompting, and then pre-fills the Title and Body fields with the rendered template. After pre-fill, the user can navigate back to Title/Body to edit before saving.

The codebase has all the building blocks. The old template use flow (removed in Phase 23 commit 87ba1d4) implemented exactly this pattern: `templateSelectMode` showed a list of templates with cursor navigation, `placeholderInputMode` prompted for each placeholder variable, and then the rendered body was attached to the new todo. The tmplmgr overlay also implements a similar list-with-cursor pattern. The key difference for Phase 25 is that the picker lives _within_ the add form's full-pane view rather than as a separate overlay or top-level mode.

The implementation approach is: when the user is in inputMode with editField=3 (Template field) and presses Enter, transition to a new sub-mode (e.g., `templatePickMode`) that shows a scrollable template list within the add form's editView. Selecting a template either immediately pre-fills (if no placeholders) or enters a placeholder prompting sub-mode. After completion, the form returns to inputMode with Title and Body pre-filled, and the Template field shows the selected template name.

**Primary recommendation:** Add a `templatePickMode` sub-state within inputMode that renders a template list in the editView, handles selection and placeholder prompting inline, then returns to inputMode with pre-filled fields. Reuse `tmpl.ExtractPlaceholders()` and `tmpl.ExecuteTemplate()` from the tmpl package.

## Standard Stack

Already in place -- no new dependencies needed.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charmbracelet/bubbles | v0.21.1 | textinput for placeholder prompting | Already used throughout |
| charmbracelet/bubbletea | v1.3.10 | Elm architecture framework | Core framework |
| internal/tmpl | n/a | ExtractPlaceholders() and ExecuteTemplate() | Purpose-built for this exact use case |
| internal/store | n/a | ListTemplates(), TodoStore interface | Template data access |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| internal/theme | n/a | Themed styles | For picker list styling |
| charmbracelet/lipgloss | v1.1.0 | Style rendering | Cursor, selected item, muted text |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Inline picker sub-mode | Separate overlay (like tmplmgr) | Overlay breaks the add form flow; inline keeps context, user stays in the form |
| New mode constants | Reuse editField with sub-state flags | Mode constants are cleaner and match existing patterns (templateSelectMode existed before) |
| Prompting placeholders inline | Skip placeholder prompting entirely | Templates with {{.Variable}} would render with empty strings; must prompt |

**Installation:** No new packages needed.

## Architecture Patterns

### Current State (after Phase 24)
```
inputMode with editField cycling:
  0 = Title (textinput)
  1 = Date (textinput)
  2 = Body (textarea)
  3 = Template (textinput, CharLimit=0, inert placeholder)

Enter on editField 3 does nothing (placeholder for Phase 25)
```

### Target Architecture After Phase 25
```
inputMode with editField cycling: same 0-3 cycle
  Enter on editField 3 opens template picker inline

New state tracking on Model:
  pickingTemplate    bool            // true when showing template list
  pickerTemplates    []store.Template // cached template list
  pickerCursor       int             // cursor position in list
  promptingPlaceholders bool         // true when prompting for placeholder values
  pickerPlaceholderNames []string    // extracted placeholder names
  pickerPlaceholderIndex int         // current placeholder being prompted
  pickerPlaceholderValues map[string]string // collected values
  pickerSelectedTemplate *store.Template   // template being filled

When pickingTemplate=true:
  - editView renders template list instead of the 4-field form
  - j/k navigates list, Enter selects, Esc cancels back to editField=3

When promptingPlaceholders=true:
  - editView renders placeholder prompt (name + text input)
  - Enter advances to next placeholder or completes
  - Esc cancels back to template list

After selection completes:
  - Title field set to template name
  - Body field set to rendered template content
  - Template field shows selected template name
  - editField set to 0 (title), user can edit
  - pickingTemplate and promptingPlaceholders reset to false
```

### Pattern 1: Template Picker as Sub-State of inputMode
**What:** Instead of adding new top-level mode constants, use boolean flags on the Model to track picker state within inputMode.
**When to use:** When the picker is a transient sub-interaction within the add form, not a standalone mode.
**Rationale:** The old templateSelectMode was a top-level mode because it was a standalone entry point. Now it is embedded within the add form flow. Using booleans avoids proliferating mode constants and keeps the mode enum clean.
**Implementation approach:**
```go
// New fields on Model:
pickingTemplate         bool
pickerTemplates         []store.Template
pickerCursor            int
promptingPlaceholders   bool
pickerPlaceholderNames  []string
pickerPlaceholderIndex  int
pickerPlaceholderValues map[string]string
pickerSelectedTemplate  *store.Template

// In updateInputMode, when editField==3 and Enter pressed:
case key.Matches(msg, m.keys.Confirm):
    if m.editField == 3 {
        templates := m.store.ListTemplates()
        if len(templates) == 0 {
            return m, nil // no templates available
        }
        m.pickingTemplate = true
        m.pickerTemplates = templates
        m.pickerCursor = 0
        return m, nil
    }
```

### Pattern 2: Pre-fill Title and Body After Selection
**What:** After template selection (with or without placeholders), set the input values directly.
**When to use:** After the template selection flow completes.
**Implementation:**
```go
// After template is selected and rendered:
m.input.SetValue(selectedTemplate.Name)    // Pre-fill Title
m.bodyTextarea.SetValue(renderedBody)       // Pre-fill Body
m.templateInput.SetValue(selectedTemplate.Name) // Show selected template name
m.pickingTemplate = false
m.promptingPlaceholders = false
m.editField = 0  // Return focus to Title so user can edit
return m, m.input.Focus()
```

### Pattern 3: Placeholder Prompting Inline
**What:** When a template has {{.Variable}} placeholders, prompt for each value before rendering.
**When to use:** After selecting a template that has placeholders.
**Implementation follows the old removed flow:**
```go
// After template selected:
names, err := tmpl.ExtractPlaceholders(selected.Content)
if err != nil || len(names) == 0 {
    // No placeholders - render immediately
    body, _ := tmpl.ExecuteTemplate(selected.Content, map[string]string{})
    // Pre-fill and return to form
} else {
    // Has placeholders - enter prompting sub-state
    m.promptingPlaceholders = true
    m.pickerPlaceholderNames = names
    m.pickerPlaceholderIndex = 0
    m.pickerPlaceholderValues = make(map[string]string)
    m.input.SetValue("")
    m.input.Placeholder = names[0]
    m.input.Prompt = names[0] + ": "
    return m, m.input.Focus()
}
```

### Pattern 4: Rendering Picker in editView
**What:** editView() checks pickingTemplate/promptingPlaceholders flags and renders appropriate UI.
**When to use:** In the View rendering path.
**Implementation:**
```go
case inputMode:
    if m.pickingTemplate {
        // Render "Select Template" heading + template list with cursor
        b.WriteString(m.styles.EditTitle.Render("Select Template"))
        b.WriteString("\n\n")
        for i, t := range m.pickerTemplates {
            if i == m.pickerCursor {
                b.WriteString(m.styles.Cursor.Render("> "))
            } else {
                b.WriteString("  ")
            }
            b.WriteString(t.Name)
            // Optional: show content preview
            preview := t.Content
            if len(preview) > 40 {
                preview = preview[:40] + "..."
            }
            preview = strings.ReplaceAll(preview, "\n", " ")
            b.WriteString("  " + m.styles.Empty.Render(preview))
            b.WriteString("\n")
        }
    } else if m.promptingPlaceholders {
        // Render placeholder prompt heading + input
        title := fmt.Sprintf("Fill Placeholder (%d/%d)",
            m.pickerPlaceholderIndex+1, len(m.pickerPlaceholderNames))
        b.WriteString(m.styles.EditTitle.Render(title))
        b.WriteString("\n\n")
        b.WriteString(m.styles.FieldLabel.Render(m.pickerPlaceholderNames[m.pickerPlaceholderIndex]))
        b.WriteString("\n")
        b.WriteString(m.input.View())
    } else {
        // Normal 4-field form (existing code)
    }
```

### Anti-Patterns to Avoid
- **Adding new top-level mode constants:** The picker is a sub-interaction of inputMode, not a standalone mode. Using mode constants would require updating all mode switches (View, Update, IsInputting, HelpBindings, AllHelpBindings, blink forwarding) and adds complexity.
- **Opening tmplmgr overlay from within the add form:** The tmplmgr overlay is for template management (rename, delete, schedule), not selection. Building a picker within the add form keeps the UX focused.
- **Skipping placeholder prompting:** Templates with `{{.Variable}}` will render with empty strings if not prompted. The old flow handled this correctly; replicate it.
- **Not allowing editing after pre-fill:** ADD-04 explicitly requires users can edit Title and Body after template selection. Pre-fill then return to editField=0 with cursor in Title field.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Template placeholder extraction | Regex parsing | tmpl.ExtractPlaceholders() | Handles If/Range/With nodes via AST walk |
| Template rendering with values | String replacement | tmpl.ExecuteTemplate() | Proper Go template execution with missingkey=zero |
| Template list from database | In-memory cache | store.ListTemplates() | Always fresh, ordered by name |
| Date parsing in save | Custom parser | config.ParseUserDate() | Already handles ISO/EU/US formats |
| Cursor blink in inputs | Manual timers | textinput.Focus()/Blur() | Bubble Tea handles cursor lifecycle |

**Key insight:** The old template use flow (removed in Phase 23) was a complete working implementation of this exact feature. The tmpl package functions and store methods are all still present. This phase just needs to wire them into the add form's existing structure.

## Common Pitfalls

### Pitfall 1: Blink Messages Not Forwarded During Picker/Prompting
**What goes wrong:** Cursor stops blinking in the placeholder input because blink tick messages are not forwarded when pickingTemplate or promptingPlaceholders is true.
**Why it happens:** The blink forwarding block in Update() (lines 290-306) forwards based on editField. During placeholder prompting, the input is m.input but editField might still be 3.
**How to avoid:** When promptingPlaceholders is true, always forward non-key messages to m.input regardless of editField value. Add a check before the editField switch:
```go
if m.promptingPlaceholders {
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```
**Warning signs:** Frozen cursor during placeholder prompting.

### Pitfall 2: Input State Leaking Between Picker and Form
**What goes wrong:** After selecting a template and returning to the form, the input field retains placeholder prompt text or the template name overwrites user's custom title.
**Why it happens:** m.input is shared between the Title field (editField=0) and the placeholder prompting flow. If not properly reset, state leaks.
**How to avoid:** After template selection completes, explicitly set all field values:
```go
m.input.SetValue(selectedTemplate.Name)  // Title = template name
m.input.Placeholder = "What needs doing?" // Restore original placeholder
m.input.Prompt = "> "                     // Restore original prompt
m.bodyTextarea.SetValue(renderedBody)
m.templateInput.SetValue(selectedTemplate.Name)
```

### Pitfall 3: Esc in Picker Cancels Entire Add Form
**What goes wrong:** User presses Esc in template list expecting to go back to the Template field, but the form closes entirely.
**Why it happens:** The existing inputMode Cancel handler closes the form from fields 0/1 and goes to field 0 from fields 2/3. If the picker doesn't intercept Esc first, it falls through.
**How to avoid:** In updateInputMode, check pickingTemplate and promptingPlaceholders BEFORE the existing Cancel handler:
```go
// In updateInputMode:
if m.pickingTemplate {
    return m.updateTemplatePicker(msg)
}
if m.promptingPlaceholders {
    return m.updatePlaceholderPrompting(msg)
}
// ... existing switch cases
```

### Pitfall 4: Empty Template List
**What goes wrong:** User presses Enter on Template field but there are no templates. App crashes or does nothing with no feedback.
**Why it happens:** store.ListTemplates() returns empty slice when no templates exist.
**How to avoid:** Check for empty list before entering picker mode. Either show a brief message or simply do nothing (return early). The existing templateInput placeholder already says "Press Enter to select template" so doing nothing is acceptable. Optionally update the placeholder to say "(no templates)" temporarily.

### Pitfall 5: Template Field Not Cleared on Cancel
**What goes wrong:** After selecting a template and then pressing Esc to cancel the entire add form, the next time the add form opens, the Template field still shows the previously selected template name.
**Why it happens:** The cancel handler in inputMode clears input, dateInput, and bodyTextarea but may not clear templateInput.
**How to avoid:** In the cancel handler (and in saveAdd), also clear all picker state:
```go
m.templateInput.SetValue("")
m.pickingTemplate = false
m.promptingPlaceholders = false
m.pickerSelectedTemplate = nil
```

### Pitfall 6: Help Bindings Not Updated for Picker Sub-States
**What goes wrong:** Help bar shows "tab: switch field / enter: confirm" when the user is in the template picker, where j/k/enter/esc are the relevant keys.
**Why it happens:** HelpBindings() checks mode and editField but not pickingTemplate/promptingPlaceholders flags.
**How to avoid:** Add checks at the top of HelpBindings() and AllHelpBindings():
```go
case inputMode:
    if m.pickingTemplate {
        return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Confirm, m.keys.Cancel}
    }
    if m.promptingPlaceholders {
        return []key.Binding{m.keys.Confirm, m.keys.Cancel}
    }
    // ... existing field-based logic
```

## Code Examples

### Template Picker Key Handling (based on old removed flow)
```go
// Source: git show 87ba1d4 (removed code, pattern to re-implement)
// updateTemplatePicker handles key events in the template picker sub-state
func (m Model) updateTemplatePicker(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, m.keys.Up):
        if m.pickerCursor > 0 {
            m.pickerCursor--
        }
        return m, nil

    case key.Matches(msg, m.keys.Down):
        if m.pickerCursor < len(m.pickerTemplates)-1 {
            m.pickerCursor++
        }
        return m, nil

    case key.Matches(msg, m.keys.Confirm):
        selected := m.pickerTemplates[m.pickerCursor]
        m.pickerSelectedTemplate = &selected
        names, err := tmpl.ExtractPlaceholders(selected.Content)
        if err != nil || len(names) == 0 {
            // No placeholders -- render and pre-fill immediately
            body, _ := tmpl.ExecuteTemplate(selected.Content, map[string]string{})
            return m.prefillFromTemplate(&selected, body), m.input.Focus()
        }
        // Has placeholders -- enter prompting sub-state
        m.promptingPlaceholders = true
        m.pickingTemplate = false
        m.pickerPlaceholderNames = names
        m.pickerPlaceholderIndex = 0
        m.pickerPlaceholderValues = make(map[string]string)
        m.input.SetValue("")
        m.input.Placeholder = names[0]
        m.input.Prompt = names[0] + ": "
        return m, m.input.Focus()

    case key.Matches(msg, m.keys.Cancel):
        m.pickingTemplate = false
        m.pickerTemplates = nil
        m.pickerCursor = 0
        // Return to Template field (editField=3)
        return m, m.templateInput.Focus()
    }
    return m, nil
}
```

### Placeholder Prompting Key Handling (based on old removed flow)
```go
// Source: git show 87ba1d4 (removed code, pattern to re-implement)
func (m Model) updatePlaceholderPrompting(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, m.keys.Confirm):
        value := strings.TrimSpace(m.input.Value())
        m.pickerPlaceholderValues[m.pickerPlaceholderNames[m.pickerPlaceholderIndex]] = value
        m.pickerPlaceholderIndex++
        if m.pickerPlaceholderIndex < len(m.pickerPlaceholderNames) {
            // More placeholders remain
            name := m.pickerPlaceholderNames[m.pickerPlaceholderIndex]
            m.input.Placeholder = name
            m.input.Prompt = name + ": "
            m.input.SetValue("")
            return m, nil
        }
        // All placeholders filled -- render and pre-fill
        body, _ := tmpl.ExecuteTemplate(
            m.pickerSelectedTemplate.Content,
            m.pickerPlaceholderValues,
        )
        return m.prefillFromTemplate(m.pickerSelectedTemplate, body), m.input.Focus()

    case key.Matches(msg, m.keys.Cancel):
        // Go back to template picker
        m.promptingPlaceholders = false
        m.pickingTemplate = true
        m.input.Blur()
        return m, nil
    }

    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```

### Pre-fill Helper
```go
// prefillFromTemplate sets form fields from a selected template.
func (m Model) prefillFromTemplate(t *store.Template, renderedBody string) Model {
    m.pickingTemplate = false
    m.promptingPlaceholders = false
    m.input.SetValue(t.Name)
    m.input.Placeholder = "What needs doing?"
    m.input.Prompt = "> "
    m.input.CursorEnd()
    m.bodyTextarea.SetValue(renderedBody)
    m.templateInput.SetValue(t.Name)
    m.editField = 0
    // Clear picker state
    m.pickerTemplates = nil
    m.pickerCursor = 0
    m.pickerSelectedTemplate = nil
    m.pickerPlaceholderNames = nil
    m.pickerPlaceholderIndex = 0
    m.pickerPlaceholderValues = nil
    return m
}
```

### Existing tmpl Package Functions
```go
// Source: internal/tmpl/tmpl.go
// ExtractPlaceholders parses template content and returns unique {{.Field}} names
func ExtractPlaceholders(content string) ([]string, error)

// ExecuteTemplate fills placeholders with provided values (missing keys = empty string)
func ExecuteTemplate(content string, values map[string]string) (string, error)
```

### Existing Store Methods
```go
// Source: internal/store/iface.go + sqlite.go
ListTemplates() []Template          // Returns all templates ordered by name
FindTemplate(id int) *Template      // Returns template by ID or nil
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Standalone `t` key -> templateSelectMode top-level mode | Embedded picker within add form's Template field | Phase 23 removed old, Phase 25 rebuilds inline | Picker is contextual, not standalone |
| Template selection creates todo immediately | Template selection pre-fills form for editing | Phase 25 (this phase) | User can review/edit before saving (ADD-04) |
| Separate placeholder prompting mode | Placeholder prompting as sub-state of inputMode | Phase 25 (this phase) | Cleaner state management |

**Deprecated/outdated:**
- `templateSelectMode`: Removed in Phase 23 (was: top-level mode for template selection)
- `placeholderInputMode`: Removed in Phase 23 (was: top-level mode for placeholder prompting)
- `t` keybinding: Removed in Phase 23 (was: enter template selection from normal mode)
- `TemplateUse` key binding: Removed in Phase 23

## Open Questions

1. **Should the picker allow template content preview?**
   - What we know: The old templateSelectMode showed a 40-char inline preview next to each template name. The tmplmgr overlay shows full content preview below the list.
   - Recommendation: Show a brief inline preview (40 chars) next to each template name in the picker list. This matches the old removed flow and gives the user enough context to pick the right template. Full preview is available via the template management overlay (`M` key).

2. **Should title pre-fill be the template name or empty?**
   - What we know: ADD-03 says "selecting a template pre-fills the Title field with the template name." The old flow used m.input as the title field after template selection (user typed their own title). The requirement explicitly calls for template name.
   - Recommendation: Pre-fill Title with the template name. User can then modify it per ADD-04.

3. **What happens if user selects a second template after already selecting one?**
   - What we know: The form has one Template field. If user tabs back to it and presses Enter again, they should be able to pick a different template.
   - Recommendation: Allow re-selection. Opening the picker again should work. The new selection overwrites the previously pre-filled Title and Body.

4. **Should the templateInput field be made read-only or text-editable?**
   - What we know: Currently CharLimit=0 makes it effectively read-only. The field's purpose is to show the selected template name and trigger the picker on Enter.
   - Recommendation: Keep it functionally read-only (CharLimit=0 or ignore typed characters). Its only interactive behavior is Enter to open picker and showing the selected template name. Tab cycles through it. Type events should be ignored.

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `internal/todolist/model.go` -- full read, all modes, field cycling, editView rendering (984 lines)
- Codebase analysis: `internal/tmpl/tmpl.go` -- ExtractPlaceholders and ExecuteTemplate functions (92 lines)
- Codebase analysis: `internal/tmplmgr/model.go` -- template list rendering pattern, cursor navigation (718 lines)
- Codebase analysis: `internal/store/iface.go` -- TodoStore interface, ListTemplates/FindTemplate signatures
- Codebase analysis: `internal/store/sqlite.go` -- ListTemplates implementation (returns []Template ordered by name)
- Codebase analysis: `internal/app/model.go` -- app-level message routing, overlay patterns (611 lines)
- Git history: `git show 87ba1d4` -- removed templateSelectMode, placeholderInputMode code (322 lines removed)
- Phase 24 research: `.planning/phases/24-unified-add-form/24-RESEARCH.md` -- editField architecture decisions

### Secondary (MEDIUM confidence)
- `.planning/REQUIREMENTS.md` -- ADD-03 and ADD-04 requirement definitions
- `.planning/PROJECT.md` -- Architecture decisions and tech stack versions

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all libraries already in go.mod, tmpl package already built for this
- Architecture: HIGH - old flow was working before removal; pattern is proven, just needs re-embedding
- Pitfalls: HIGH - identified from direct code reading and analysis of the old removed flow's edge cases

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (stable -- all patterns are internal codebase patterns, not external library concerns)
