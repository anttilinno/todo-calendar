# Phase 5: Todo Editing - Research

**Researched:** 2026-02-05
**Domain:** In-place todo text editing, date mutation (add/change/remove), Bubble Tea mode-based input, store update methods
**Confidence:** HIGH

## Summary

Phase 5 adds editing capabilities to the existing todo list: text editing (EDIT-01), date mutation (EDIT-02), and immediate persistence (EDIT-03). This builds directly on the Phase 3 todo CRUD system, reusing the existing `textinput.Model`, mode-based input handling, and store persistence patterns.

The existing codebase already has all the primitives needed. The `textinput.Model` supports `SetValue()` to pre-fill existing text and `CursorEnd()` to position the cursor at the end -- these are the building blocks for "edit mode." The store already has `Save()` with atomic writes. The todolist model already has `inputMode` and `dateInputMode` with Enter/Escape handling. The editing feature is fundamentally an extension of the existing add workflow: instead of starting with an empty input, start with the selected todo's text pre-filled, and instead of calling `store.Add()` on confirm, call a new `store.UpdateText()` or `store.UpdateDate()` method.

The main design decisions for this phase are: (1) what key triggers edit (`e` is the natural choice, matching vim conventions used throughout the app), (2) how to handle date editing UX (a separate `E` key for date editing, mirroring the `a`/`A` add pattern), and (3) whether text and date editing share modes or get new modes. The recommendation is to add two new modes (`editTextMode` and `editDateMode`) and two new store methods (`UpdateText` and `UpdateDate`), keeping the implementation clean and parallel to the existing add flow.

**Primary recommendation:** Add `e` key for text editing (pre-fills existing text) and `E` key for date editing (pre-fills existing date or shows empty for floating todos). Reuse the existing `textinput.Model` and mode-switching patterns. Add `UpdateText(id, text)` and `UpdateDate(id, date)` methods to the store. No new dependencies needed.

## Standard Stack

No new libraries are needed. Phase 5 uses only the existing stack.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `bubbles/textinput` | v0.21.1 | Pre-fill existing text via `SetValue()`, cursor at end via `CursorEnd()` | Already used for todo add; `SetValue` verified in official docs |
| `bubbles/key` | v0.21.1 | New `Edit` and `EditDate` key bindings | Already used for all other keybindings |
| `internal/store` | project | New `UpdateText()` and `UpdateDate()` methods | Follows existing `Add`/`Toggle`/`Delete` patterns |

### Existing (unchanged)
| Library | Version | Purpose |
|---------|---------|---------|
| Bubble Tea | v1.3.10 | TUI framework |
| Lip Gloss | v1.1.0 | Terminal styling |
| Bubbles | v0.21.1 | `key.Binding`, `textinput.Model`, `help.Model` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Separate `e`/`E` keys for text/date | Single `e` key with multi-step edit flow | Two separate keys is simpler, mirrors `a`/`A` pattern already established for add. A single key with "what do you want to edit?" step adds a mode and slows the user down. |
| New `editTextMode`/`editDateMode` modes | Reuse `inputMode`/`dateInputMode` with an `editing` flag | New modes are clearer in the code. An `editing` flag means every branch in `updateInputMode` checks `if m.editing` -- messier than separate handler functions. |
| `UpdateText` + `UpdateDate` store methods | Single `Update(id int, text string, date string)` method | Separate methods match the separate UX flows. A combined method requires the caller to always pass both fields, even when only one changed. However, a combined method is also viable and simpler if you prefer fewer store methods -- this is a judgment call. Recommendation: use a single `Update(id, text, date)` for simplicity since both fields are always known at edit time. |

**Installation:** No new `go get` needed. All functionality comes from existing dependencies and the Go standard library.

## Architecture Patterns

### Files Modified
```
internal/
  todolist/
    model.go      # Add editTextMode, editDateMode, edit key handlers, editingID field
    keys.go       # Add Edit and EditDate key bindings
  store/
    store.go      # Add Update(id, text, date) method
  app/
    model.go      # No changes needed (isInputting already covers new modes)
```

### Pattern 1: Edit Modes Parallel to Add Modes
**What:** Add `editTextMode` and `editDateMode` constants to the existing `mode` type. These work identically to `inputMode` and `dateInputMode` but on confirm they call `store.Update()` instead of `store.Add()`. The model gains an `editingID int` field to track which todo is being edited.

**When to use:** Whenever the user presses `e` or `E` on a selected todo.

**Confidence:** HIGH (follows exact same pattern as existing add flow)

**Example:**
```go
const (
    normalMode    mode = iota
    inputMode          // typing NEW todo text
    dateInputMode      // typing date for a NEW dated todo
    editTextMode       // editing EXISTING todo text
    editDateMode       // editing EXISTING todo date
)

type Model struct {
    // ... existing fields ...
    editingID   int    // ID of the todo being edited (valid in edit modes)
}
```

### Pattern 2: Pre-fill textinput with SetValue + CursorEnd
**What:** When entering edit mode, call `m.input.SetValue(existingText)` to pre-fill the input with the todo's current text, then `m.input.CursorEnd()` to place the cursor at the end. This gives the user their existing text ready to modify.

**When to use:** When entering `editTextMode` or `editDateMode`.

**Confidence:** HIGH (verified `SetValue` and `CursorEnd` in official bubbles v0.21.1 docs)

**Example:**
```go
case key.Matches(msg, m.keys.Edit):
    if len(selectable) > 0 && m.cursor < len(selectable) {
        idx := selectable[m.cursor]
        if items[idx].todo != nil {
            todo := items[idx].todo
            m.editingID = todo.ID
            m.mode = editTextMode
            m.input.Placeholder = "Edit todo text"
            m.input.Prompt = "> "
            m.input.SetValue(todo.Text)
            m.input.CursorEnd()
            return m, m.input.Focus()
        }
    }
```

### Pattern 3: Store Update Method
**What:** A single `Update(id int, text string, date string)` method that finds the todo by ID and overwrites its Text and Date fields, then persists. This follows the exact same pattern as `Toggle` and `Delete`.

**When to use:** When edit is confirmed.

**Confidence:** HIGH (trivial extension of existing store patterns)

**Example:**
```go
// Update modifies the text and date of the todo with the given ID and persists.
// Date should be "YYYY-MM-DD" or "" for floating.
func (s *Store) Update(id int, text string, date string) {
    for i := range s.data.Todos {
        if s.data.Todos[i].ID == id {
            s.data.Todos[i].Text = text
            s.data.Todos[i].Date = date
            s.Save()
            return
        }
    }
}
```

### Pattern 4: IsInputting Covers Edit Modes
**What:** The existing `IsInputting()` method returns `m.mode != normalMode`. Since `editTextMode` and `editDateMode` are not `normalMode`, the root model's input guard (`isInputting := m.activePane == todoPane && m.todoList.IsInputting()`) automatically works for edit modes too. No changes needed in `app/model.go`.

**When to use:** Always -- this is why the existing pattern works.

**Confidence:** HIGH (verified by reading current code)

### Pattern 5: Date Edit UX -- E Key with Pre-filled Date
**What:** Pressing `E` on a selected todo enters date edit mode. The input pre-fills with the current date (e.g., `2026-02-15`) if the todo has one, or is empty if the todo is floating. The user can:
- Type a new date and press Enter to set/change the date
- Clear the input and press Enter to remove the date (making it floating)
- Press Escape to cancel

This mirrors the existing `dateInputMode` flow but with pre-fill and the ability to clear.

**When to use:** When the user wants to add a date to a floating todo, change an existing date, or remove a date.

**Confidence:** HIGH (reuses existing date input validation pattern)

**Example:**
```go
case key.Matches(msg, m.keys.EditDate):
    if len(selectable) > 0 && m.cursor < len(selectable) {
        idx := selectable[m.cursor]
        if items[idx].todo != nil {
            todo := items[idx].todo
            m.editingID = todo.ID
            m.mode = editDateMode
            m.input.Placeholder = "YYYY-MM-DD (empty = floating)"
            m.input.Prompt = "Date: "
            m.input.SetValue(todo.Date) // pre-fill existing date or ""
            m.input.CursorEnd()
            return m, m.input.Focus()
        }
    }
```

### Pattern 6: Edit Date Confirmation with Empty-Means-Floating
**What:** In `editDateMode`, the confirmation logic differs from `dateInputMode` for new todos:
- Empty input = remove date (set to "", making todo floating). This is NOT an error -- it is a valid action.
- Non-empty input = validate as YYYY-MM-DD, reject invalid dates (stay in edit mode).

This is the key difference from `dateInputMode` for new todos, where empty input was rejected.

**Confidence:** HIGH (straightforward logic)

**Example:**
```go
func (m Model) updateEditDateMode(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, m.keys.Confirm):
        date := strings.TrimSpace(m.input.Value())
        if date != "" {
            // Validate date format
            if _, err := time.Parse("2006-01-02", date); err != nil {
                return m, nil // stay in edit mode
            }
        }
        // date is either valid "YYYY-MM-DD" or "" (floating)
        m.store.Update(m.editingID, m.editingText(m.editingID), date)
        m.mode = normalMode
        m.input.Blur()
        m.input.SetValue("")
        return m, nil

    case key.Matches(msg, m.keys.Cancel):
        m.mode = normalMode
        m.input.Blur()
        m.input.SetValue("")
        return m, nil
    }

    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```

### Anti-Patterns to Avoid
- **Editing by delete-and-re-add:** Loses the original ID, creation timestamp, and done status. Always update in place.
- **Sharing mode handlers between add and edit:** The confirm logic differs (Add vs Update, empty-date handling). Keep separate handler functions like `updateEditTextMode` and `updateEditDateMode`.
- **Forgetting to clear editingID on cancel:** After Escape, `editingID` should be reset to 0 (or left as-is since it is only read when in edit mode -- but resetting is cleaner).
- **Not preserving the non-edited field:** When editing text, the date must not change. When editing date, the text must not change. The `store.Update(id, text, date)` method takes both, so the caller must pass the current value of the non-edited field. This can be done by reading the todo from the store at confirmation time, or by storing the non-edited field when entering edit mode.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Text input pre-fill with cursor | Manual rune-by-rune pre-population | `textinput.SetValue()` + `CursorEnd()` | SetValue handles all internal state (cursor, viewport scrolling) correctly |
| Date validation | Custom regex or string parsing | `time.Parse("2006-01-02", date)` | Already used in `dateInputMode`; handles leap years, month boundaries, etc. |
| Atomic persistence after edit | Custom write logic | Existing `store.Save()` via atomic temp+rename | Already proven in add/toggle/delete flows |
| Mode-based input routing | Complex if/else chains | Mode enum + switch statement | Already established pattern in todolist/model.go |

**Key insight:** This phase requires almost no new patterns. Every building block exists in the codebase from Phase 3. The work is wiring existing patterns to new key bindings and a new store method.

## Common Pitfalls

### Pitfall 1: Stale Todo Data When Confirming Edit
**What goes wrong:** User presses `e`, gets text pre-filled. But between entering edit mode and pressing Enter, something else could theoretically modify the todo (not in current app, but defensive coding). More realistically: if `store.Update()` takes separate text and date params, the caller must know the *current* value of the field they are NOT editing.
**Why it happens:** The edit flow captures the todo's state when entering edit mode but applies changes later.
**How to avoid:** When entering edit mode, store both the todo's current text AND date. When confirming text edit, pass the stored date (not re-queried). When confirming date edit, pass the stored text. Alternatively, read the todo fresh from the store at confirmation time. The simplest approach: store `editingID` and look up the todo in the store when confirming to get the non-edited field's current value. Since this is single-threaded Bubble Tea, no race conditions exist.
**Warning signs:** Editing a todo's text accidentally clears its date, or vice versa.

### Pitfall 2: Cursor Position After SetValue
**What goes wrong:** After `SetValue()`, the cursor position might be at position 0 (start of text). User expects to append to existing text but instead types at the beginning.
**Why it happens:** `SetValue()` resets internal state. The cursor position after `SetValue` may default to the end or to the start depending on the implementation.
**How to avoid:** Always call `CursorEnd()` immediately after `SetValue()` to ensure the cursor is at the end of the pre-filled text. This is the expected UX for editing.
**Warning signs:** Cursor appears at the beginning of the text when entering edit mode.

### Pitfall 3: Empty Text on Edit Confirm
**What goes wrong:** User clears all text and presses Enter. An empty todo gets saved.
**Why it happens:** The add flow already guards against empty text, but the edit flow might forget this check.
**How to avoid:** Apply the same `strings.TrimSpace(text) == ""` guard in edit text confirmation. If empty, either stay in edit mode (same as add) or cancel the edit.
**Warning signs:** Blank-text todos appearing in the list.

### Pitfall 4: Edit Key Pressed With No Todos Selected
**What goes wrong:** If the list is empty (no todos), pressing `e` or `E` should do nothing. But if the selectable check is missing, it could index into an empty slice.
**Why it happens:** Same guard needed as toggle/delete but easy to forget for new keys.
**How to avoid:** Copy the exact same guard pattern from toggle/delete: `if len(selectable) > 0 && m.cursor < len(selectable)`.
**Warning signs:** Panic on empty list when pressing `e`.

### Pitfall 5: HelpBindings Not Updated for Edit Keys
**What goes wrong:** New `e`/`E` keys work but do not appear in the help bar. Users do not discover the feature.
**Why it happens:** `HelpBindings()` returns a hardcoded slice; new bindings must be added.
**How to avoid:** Add `m.keys.Edit` and `m.keys.EditDate` to the `HelpBindings()` return in normal mode.
**Warning signs:** Help bar does not show edit bindings.

### Pitfall 6: Todo Moves Between Sections After Date Edit
**What goes wrong:** User edits a dated todo to remove its date (making it floating). The todo now appears in the "Floating" section. The cursor index still points to the old position in the dated section, which may now be out of bounds or pointing at a different todo.
**Why it happens:** Editing a date can move a todo between sections, changing the selectable indices.
**How to avoid:** After confirming a date edit, rebuild the visible items and clamp the cursor, just like after delete. Use the same pattern: `newSelectable := selectableIndices(m.visibleItems()); if m.cursor >= len(newSelectable) { m.cursor = max(0, len(newSelectable)-1) }`.
**Warning signs:** Cursor jumps to wrong todo or panics after changing a todo from dated to floating (or vice versa).

## Code Examples

### Store Update Method
```go
// internal/store/store.go
// Source: follows existing Add/Toggle/Delete pattern

// Update modifies the text and date of the todo with the given ID and persists.
// An empty date makes the todo floating.
func (s *Store) Update(id int, text string, date string) {
    for i := range s.data.Todos {
        if s.data.Todos[i].ID == id {
            s.data.Todos[i].Text = text
            s.data.Todos[i].Date = date
            s.Save()
            return
        }
    }
}
```

### New Key Bindings
```go
// internal/todolist/keys.go
// Source: follows existing KeyMap pattern

type KeyMap struct {
    Up       key.Binding
    Down     key.Binding
    Add      key.Binding
    AddDated key.Binding
    Toggle   key.Binding
    Delete   key.Binding
    Edit     key.Binding  // NEW
    EditDate key.Binding  // NEW
    Confirm  key.Binding
    Cancel   key.Binding
}

// In DefaultKeyMap():
Edit: key.NewBinding(
    key.WithKeys("e"),
    key.WithHelp("e", "edit text"),
),
EditDate: key.NewBinding(
    key.WithKeys("E"),
    key.WithHelp("E", "edit date"),
),
```

### Entering Edit Text Mode
```go
// internal/todolist/model.go
// Source: mirrors existing Add flow with SetValue pre-fill

case key.Matches(msg, m.keys.Edit):
    if len(selectable) > 0 && m.cursor < len(selectable) {
        idx := selectable[m.cursor]
        if items[idx].todo != nil {
            todo := items[idx].todo
            m.editingID = todo.ID
            m.mode = editTextMode
            m.input.Placeholder = "Edit todo text"
            m.input.Prompt = "> "
            m.input.SetValue(todo.Text)
            m.input.CursorEnd()
            return m, m.input.Focus()
        }
    }
```

### Confirming Edit Text
```go
// internal/todolist/model.go

func (m Model) updateEditTextMode(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, m.keys.Confirm):
        text := strings.TrimSpace(m.input.Value())
        if text == "" {
            return m, nil // don't save empty text
        }
        // Look up the todo to get its current date (non-edited field)
        todo := m.store.Find(m.editingID)
        if todo != nil {
            m.store.Update(m.editingID, text, todo.Date)
        }
        m.mode = normalMode
        m.input.Blur()
        m.input.SetValue("")
        return m, nil

    case key.Matches(msg, m.keys.Cancel):
        m.mode = normalMode
        m.input.Blur()
        m.input.SetValue("")
        return m, nil
    }

    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```

### Store Find Method (Helper)
```go
// internal/store/store.go
// Needed to look up a todo's current state when confirming a partial edit

// Find returns a pointer to the todo with the given ID, or nil if not found.
func (s *Store) Find(id int) *Todo {
    for i := range s.data.Todos {
        if s.data.Todos[i].ID == id {
            return &s.data.Todos[i]
        }
    }
    return nil
}
```

### Confirming Edit Date (with empty = floating)
```go
// internal/todolist/model.go

func (m Model) updateEditDateMode(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, m.keys.Confirm):
        date := strings.TrimSpace(m.input.Value())
        if date != "" {
            if _, err := time.Parse("2006-01-02", date); err != nil {
                return m, nil // invalid date, stay in edit mode
            }
        }
        // date is "" (floating) or valid "YYYY-MM-DD"
        todo := m.store.Find(m.editingID)
        if todo != nil {
            m.store.Update(m.editingID, todo.Text, date)
        }
        m.mode = normalMode
        m.input.Blur()
        m.input.SetValue("")
        // Clamp cursor since todo may have moved between sections
        newSelectable := selectableIndices(m.visibleItems())
        if m.cursor >= len(newSelectable) {
            m.cursor = max(0, len(newSelectable)-1)
        }
        return m, nil

    case key.Matches(msg, m.keys.Cancel):
        m.mode = normalMode
        m.input.Blur()
        m.input.SetValue("")
        return m, nil
    }

    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```

### Updated HelpBindings
```go
// internal/todolist/model.go

func (m Model) HelpBindings() []key.Binding {
    if m.mode != normalMode {
        return []key.Binding{m.keys.Confirm, m.keys.Cancel}
    }
    return []key.Binding{
        m.keys.Up, m.keys.Down,
        m.keys.Add, m.keys.AddDated,
        m.keys.Edit, m.keys.EditDate,
        m.keys.Toggle, m.keys.Delete,
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Delete and re-add to edit | In-place editing via `store.Update()` | Phase 5 | Preserves ID, creation timestamp, done status |
| No date mutation | Date add/change/remove via `editDateMode` | Phase 5 | Floating todos can become dated and vice versa |

**Deprecated/outdated:**
- Nothing deprecated. Phase 5 extends existing patterns without replacing any.

## Open Questions

1. **Should `e` edit both text and date in sequence, or only text?**
   - What we know: The requirements say "press `e` to edit selected todo's text" (EDIT-01) and "change a todo's date" (EDIT-02) as separate requirements.
   - What's unclear: Whether they must be separate keys or could be a single flow.
   - Recommendation: Use `e` for text only and `E` for date only. This mirrors `a`/`A` for add. It is simpler, faster for the common case (editing text), and each flow is self-contained. The user can press `e` then `E` if they want to change both.

2. **Should there be visual feedback that a todo is being edited?**
   - What we know: In add mode, the input shows at the bottom of the list. In edit mode, should the selected todo be highlighted differently?
   - What's unclear: Whether special highlighting is needed or the input being visible is sufficient.
   - Recommendation: Keep it simple -- the input appearing with pre-filled text is sufficient feedback. The prompt text changes to indicate editing (e.g., `"> "` with existing text). No special highlighting needed for v1.1. The help bar already shows "enter confirm / esc cancel" during input.

3. **Cursor clamping after date edit moves todo between sections**
   - What we know: Changing a dated todo to floating (or vice versa) moves it between sections, potentially invalidating the cursor.
   - What's unclear: Should the cursor follow the edited todo to its new section?
   - Recommendation: Clamp the cursor to valid bounds (same as delete). Following the todo to its new section would require scanning for it by ID, which adds complexity for minimal UX benefit. The user can see where it moved and navigate to it.

## Sources

### Primary (HIGH confidence)
- [bubbles/textinput v0.21.1 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/textinput) -- `SetValue()`, `CursorEnd()`, `Focus()`, `Blur()`, `Value()` API verified
- Project source code: `internal/todolist/model.go` -- existing mode-based input handling, `inputMode`/`dateInputMode` patterns
- Project source code: `internal/store/store.go` -- existing `Add()`/`Toggle()`/`Delete()` patterns, `Save()` atomic write
- Project source code: `internal/todolist/keys.go` -- existing `KeyMap` struct and `DefaultKeyMap()` pattern
- Project source code: `internal/app/model.go` -- existing `isInputting` guard that will cover new edit modes

### Secondary (MEDIUM confidence)
- [bubbles textinput source](https://github.com/charmbracelet/bubbles/blob/master/textinput/textinput.go) -- SetValue implementation details

### Tertiary (LOW confidence)
- None -- all findings verified with primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- zero new dependencies; all APIs verified against pkg.go.dev and existing codebase
- Architecture: HIGH -- patterns are direct extensions of existing Phase 3 code, all verifiable in source
- Pitfalls: HIGH -- derived from reading actual source code and understanding mode-based input flow
- Code examples: HIGH -- all API calls verified against existing usage and official documentation

**Research date:** 2026-02-05
**Valid until:** 2026-03-07 (30 days -- stable libraries, no external dependencies changing)
