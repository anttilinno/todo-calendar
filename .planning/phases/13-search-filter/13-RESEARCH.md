# Phase 13: Search & Filter - Research

**Researched:** 2026-02-06
**Domain:** Bubble Tea TUI inline filtering and full-screen search overlay
**Confidence:** HIGH

## Summary

This phase adds two distinct search/filter capabilities to the todo calendar app: (1) an inline text filter on the todolist pane that narrows visible todos in the current month when the user presses `/`, and (2) a full-screen search overlay that searches across ALL months and allows jumping to a result's month.

The existing codebase already has all the building blocks needed. The `textinput.Model` from charmbracelet/bubbles is already used extensively in the todolist for adding/editing todos. The settings overlay (`internal/settings`) establishes the pattern for full-screen overlays routed through `app.Model`. The store's `Todos() []Todo` method returns all todos for cross-month search, and `TodosForMonth()` returns month-scoped todos for inline filtering. No new dependencies are required.

The implementation requires: (a) adding a `filterMode` to the todolist's mode enum and filtering `visibleItems()` based on a query string using case-insensitive substring matching, (b) creating a new `internal/search` package following the settings overlay pattern with its own Model/KeyMap/Styles, and (c) wiring both into the app layer with appropriate key bindings and message routing.

**Primary recommendation:** Use `strings.Contains(strings.ToLower(todo.Text), strings.ToLower(query))` for matching (explicit out-of-scope: no fuzzy matching). Follow existing overlay/mode patterns exactly.

## Standard Stack

No new dependencies needed. Everything uses the existing stack.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charmbracelet/bubbletea | v1.3.10 | TUI framework (Elm Architecture) | Already in use, the app's foundation |
| charmbracelet/bubbles | v0.21.1 | textinput.Model for filter/search input | Already in use for todo text/date input |
| charmbracelet/lipgloss | v1.1.0 | Styled rendering | Already in use for all UI components |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| strings (stdlib) | go1.25 | Case-insensitive substring matching | `strings.Contains` + `strings.ToLower` for search |
| time (stdlib) | go1.25 | Date parsing for search result display | Already used throughout for date handling |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Manual substring match | bubbles/list built-in filter | bubbles/list is a full component replacement; overkill since we have a custom todolist |
| strings.Contains | sahilm/fuzzy or similar | Explicitly out of scope per REQUIREMENTS.md |

**Installation:**
```bash
# No new packages needed
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── search/          # NEW: full-screen search overlay
│   ├── model.go     # Search model with textinput, results list, cursor
│   ├── keys.go      # KeyMap for search-specific bindings
│   └── styles.go    # Theme-aware styles for search overlay
├── todolist/
│   └── model.go     # MODIFIED: add filterMode, filter query, filtered visibleItems
├── app/
│   ├── model.go     # MODIFIED: add showSearch bool, search Model, wire routing
│   └── keys.go      # MODIFIED: add Search key binding (Ctrl+F or similar)
└── store/
    └── store.go     # MODIFIED: add SearchTodos(query) method
```

### Pattern 1: Inline Filter Mode (in todolist)
**What:** Add a `filterMode` to the existing todolist mode enum. When active, show a textinput at the top of the todo list and filter `visibleItems()` to only show todos whose text contains the query substring.
**When to use:** SRCH-01 and SRCH-02 (current-month filtering)
**Example:**
```go
// In todolist/model.go - extend mode enum
const (
    normalMode    mode = iota
    inputMode
    dateInputMode
    editTextMode
    editDateMode
    filterMode         // NEW: inline filtering
)

// Add fields to Model
type Model struct {
    // ... existing fields ...
    filterQuery string // current filter text (empty = no filter)
}

// In visibleItems(), apply filter when filterQuery is set
func (m Model) visibleItems() []visibleItem {
    // ... existing logic to build items ...
    // After building items, if filterQuery != "", filter todoItems:
    if m.filterQuery != "" {
        query := strings.ToLower(m.filterQuery)
        var filtered []visibleItem
        for _, item := range items {
            if item.kind != todoItem {
                filtered = append(filtered, item) // keep headers
                continue
            }
            if strings.Contains(strings.ToLower(item.todo.Text), query) {
                filtered = append(filtered, item)
            }
        }
        items = filtered
    }
    // ... rest of method ...
}
```

### Pattern 2: Full-Screen Search Overlay (new package)
**What:** A new `internal/search` package modeled exactly after `internal/settings`. The app layer holds a `showSearch bool` and a `search.Model`, routing messages when active.
**When to use:** SRCH-03, SRCH-04, SRCH-05 (cross-month search)
**Example:**
```go
// search/model.go
package search

// SearchResult represents a matched todo with its date context.
type SearchResult struct {
    Todo      store.Todo
    DateLabel string // formatted date for display
}

// JumpMsg is emitted when user selects a result to jump to its month.
type JumpMsg struct {
    Year  int
    Month time.Month
}

// CloseMsg is emitted when user presses Esc to close search.
type CloseMsg struct{}

type Model struct {
    input      textinput.Model
    results    []SearchResult
    cursor     int
    store      *store.Store
    dateLayout string
    width      int
    height     int
    keys       KeyMap
    styles     Styles
}
```

### Pattern 3: App-Level Overlay Routing (follows settings pattern)
**What:** The `app.Model` gains `showSearch bool` and `search search.Model`. In `Update()`, when `showSearch` is true, messages route to the search model. `JumpMsg` causes the app to navigate the calendar to the target month, close search, and sync the todolist.
**When to use:** Wiring the search overlay into the app
**Example:**
```go
// In app/model.go Update(), handle search messages
case search.JumpMsg:
    m.showSearch = false
    // Navigate calendar to target month
    m.calendar.SetYearMonth(msg.Year, msg.Month)
    m.todoList.SetViewMonth(msg.Year, msg.Month)
    m.calendar.RefreshIndicators()
    return m, nil

case search.CloseMsg:
    m.showSearch = false
    return m, nil
```

### Anti-Patterns to Avoid
- **Modifying the store for filtering:** The inline filter should NOT modify store data. It should only filter the `visibleItems()` display slice. The store remains the source of truth.
- **Sharing filter state between todolist and search:** The inline filter (in todolist) and the full-screen search (in search package) are independent features. Don't couple them.
- **Blocking the main loop during search:** All matching must be synchronous and fast. The store holds all todos in memory, so iterating with substring match is trivially fast even with thousands of items.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Text input widget | Custom key-by-key reader | `bubbles/textinput.Model` | Already used, handles unicode, paste, cursor, scrolling |
| Case-insensitive match | Custom rune-by-rune comparison | `strings.Contains(strings.ToLower(a), strings.ToLower(b))` | Standard, correct for all Unicode |
| Overlay message routing | Custom event bus | Bubble Tea's message/command pattern (see settings overlay) | Already proven in this codebase |

**Key insight:** The project already has all the patterns needed. The settings overlay is the template for the search overlay. The todolist's mode system is the template for the filter mode. Follow existing patterns exactly.

## Common Pitfalls

### Pitfall 1: Cursor Index Invalidation After Filter
**What goes wrong:** When the user types a filter character, the visible items list shrinks. If the cursor index was 5 and only 3 items remain visible, the app panics with an index-out-of-bounds.
**Why it happens:** The cursor is an index into the selectable items. Filtering changes the selectable count.
**How to avoid:** After updating `filterQuery`, always clamp the cursor: `m.cursor = min(m.cursor, max(0, len(newSelectable)-1))`. The todolist already does this in `updateEditDateMode` after date edits move items between sections.
**Warning signs:** Panic on typing in the filter input.

### Pitfall 2: Key Binding Conflicts with `/`
**What goes wrong:** The `/` key needs to activate the filter when the todolist pane is focused in normal mode, but it should NOT activate when the user is in input/edit mode (typing a todo or date).
**Why it happens:** The `isInputting()` check in app.Model already gates certain keys, but `/` activation needs to be within the todolist's own key handling or gated properly.
**How to avoid:** Add the filter activation key (`/`) to the todolist's `updateNormalMode()` handler, NOT to the app-level key map. This way it only fires when the todolist is focused and in normal mode. Alternatively, handle it at the app level with the same `!isInputting` guard used for other keys.
**Warning signs:** Filter activates while user is typing a todo.

### Pitfall 3: Forgetting to Clear Filter on Month Change
**What goes wrong:** User filters todos, then switches to calendar pane, navigates to a different month, comes back -- the filter is still active but the todo list shows a different month's items. This is confusing.
**Why it happens:** Filter state persists across month changes.
**How to avoid:** Clear `filterQuery` and reset to `normalMode` whenever `SetViewMonth()` is called, or when the user switches panes via Tab.
**Warning signs:** Stale filter text shown after navigating away and back.

### Pitfall 4: Search Results Not Showing Floating (Undated) Todos
**What goes wrong:** The full-screen search only shows dated todos, missing floating ones.
**Why it happens:** The search method only iterates `TodosForMonth` calls instead of `Todos()`.
**How to avoid:** Use `store.Todos()` which returns ALL todos (dated and floating). For floating todos, display "No date" or "Floating" instead of a date.
**Warning signs:** User searches for a todo they know exists but it doesn't appear.

### Pitfall 5: Calendar Lacks SetYearMonth Method
**What goes wrong:** When a user selects a search result to jump to its month, there is no method on `calendar.Model` to navigate directly to a specific year/month.
**Why it happens:** The calendar only exposes `Year()`, `Month()`, and navigation via PrevMonth/NextMonth keys. No direct setter exists.
**How to avoid:** Add a `SetYearMonth(year int, month time.Month)` method to `calendar.Model` that sets the year/month and refreshes holidays/indicators. This is a small addition.
**Warning signs:** Compile error when trying to set calendar position from search result.

## Code Examples

### Inline Filter: Activating with `/`
```go
// In todolist/model.go updateNormalMode()
case key.Matches(msg, m.keys.Filter):
    m.mode = filterMode
    m.filterQuery = ""
    m.input.Placeholder = "Filter todos..."
    m.input.Prompt = "/ "
    m.input.SetValue("")
    return m, m.input.Focus()
```

### Inline Filter: Handling Keystrokes
```go
// In todolist/model.go - new updateFilterMode()
func (m Model) updateFilterMode(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch {
    case key.Matches(msg, m.keys.Cancel):
        // Esc clears filter (SRCH-02)
        m.mode = normalMode
        m.filterQuery = ""
        m.input.Blur()
        m.input.SetValue("")
        // Clamp cursor after filter removal
        selectable := selectableIndices(m.visibleItems())
        if m.cursor >= len(selectable) {
            m.cursor = max(0, len(selectable)-1)
        }
        return m, nil
    }
    // Forward to textinput, then update filterQuery
    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    m.filterQuery = m.input.Value()
    // Clamp cursor after filter change
    selectable := selectableIndices(m.visibleItems())
    if m.cursor >= len(selectable) {
        m.cursor = max(0, len(selectable)-1)
    }
    return m, cmd
}
```

### Store: Search Method
```go
// In store/store.go
// SearchTodos returns all todos whose text contains the query (case-insensitive).
// Results are sorted by date (dated first, then floating).
func (s *Store) SearchTodos(query string) []Todo {
    q := strings.ToLower(query)
    var results []Todo
    for _, t := range s.data.Todos {
        if strings.Contains(strings.ToLower(t.Text), q) {
            results = append(results, t)
        }
    }
    sort.Slice(results, func(i, j int) bool {
        // Dated before floating
        if results[i].HasDate() != results[j].HasDate() {
            return results[i].HasDate()
        }
        // By date ascending
        if results[i].Date != results[j].Date {
            return results[i].Date < results[j].Date
        }
        return results[i].ID < results[j].ID
    })
    return results
}
```

### Search Overlay: Result Selection and Jump
```go
// In search/model.go
case key.Matches(msg, m.keys.Select):
    if len(m.results) > 0 && m.cursor < len(m.results) {
        result := m.results[m.cursor]
        if result.Todo.HasDate() {
            d, _ := time.Parse("2006-01-02", result.Todo.Date)
            return m, func() tea.Msg {
                return JumpMsg{Year: d.Year(), Month: d.Month()}
            }
        }
        // For floating todos, just close search (no month to jump to)
        return m, func() tea.Msg { return CloseMsg{} }
    }
```

### App: Opening Search Overlay
```go
// In app/model.go Update() - key handling section
case key.Matches(msg, m.keys.Search) && !isInputting:
    m.search = search.New(m.todoList.Store(), theme.ForName(m.cfg.Theme), m.cfg)
    m.search.SetSize(m.width, m.height)
    m.showSearch = true
    return m, m.search.Init()
```

### Calendar: Direct Navigation Method
```go
// In calendar/model.go
// SetYearMonth navigates directly to the specified year and month.
func (m *Model) SetYearMonth(year int, month time.Month) {
    m.year = year
    m.month = month
    m.holidays = m.provider.HolidaysInMonth(year, month)
    m.indicators = m.store.IncompleteTodosPerDay(year, month)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| bubbles v1 list with built-in filter | Custom filter on existing todolist | N/A | We use a custom todolist, not bubbles/list, so we implement filtering ourselves |

**Notes:**
- The bubbles/list component has built-in filtering, but this project uses a custom todolist. Adopting bubbles/list would be a rewrite, not a search feature.
- The bubbletea-overlay package exists but is unnecessary; the existing settings overlay pattern using a bool flag and message routing is simpler and already proven in this codebase.

## Open Questions

1. **Key binding for full-screen search**
   - What we know: `/` is for inline filter (specified in requirements). The full-screen search needs a separate key.
   - What's unclear: Should it be `Ctrl+F`, `?`, or `F` (capital)? The settings overlay uses `s`.
   - Recommendation: Use `Ctrl+F` since it is the universal "find" shortcut and doesn't conflict with any existing bindings. It works in both panes (app-level binding). Alternative: use `?` which is common in vim-style apps for reverse search.

2. **Should inline filter also filter floating todos?**
   - What we know: The inline filter operates on the current month view, which shows dated todos AND floating todos.
   - What's unclear: Should the filter apply to both sections or only the dated section?
   - Recommendation: Filter both sections. The user sees both in the todolist pane and expects the filter to apply to all visible items.

3. **Search result highlight/matching indicator**
   - What we know: Requirements say results show "matching todos with their dates." They don't mention highlighting the matched substring.
   - What's unclear: Should we highlight the matching portion of the text?
   - Recommendation: Keep it simple for v1 -- just show the todo text and date. Highlighting is a nice-to-have for a future phase.

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `internal/settings/model.go` - overlay pattern with messages
- Codebase analysis: `internal/todolist/model.go` - mode enum, textinput usage, visibleItems pattern
- Codebase analysis: `internal/store/store.go` - `Todos()` method for all todos, `TodosForMonth()` for month-scoped
- Codebase analysis: `internal/app/model.go` - overlay routing with `showSettings` bool
- go.mod: bubbletea v1.3.10, bubbles v0.21.1, lipgloss v1.1.0

### Secondary (MEDIUM confidence)
- [charmbracelet/bubbles textinput docs](https://pkg.go.dev/github.com/charmbracelet/bubbles/textinput) - v0.21.1 API confirmed
- [charmbracelet/bubbles list docs](https://pkg.go.dev/github.com/charmbracelet/bubbles/list) - confirmed list has built-in filtering (but not used here)

### Tertiary (LOW confidence)
- None needed; this phase relies entirely on existing codebase patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - no new deps, all patterns exist in codebase
- Architecture: HIGH - follows established settings overlay and todolist mode patterns exactly
- Pitfalls: HIGH - identified from direct codebase analysis (cursor clamping, key conflicts, month navigation)

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (stable; no external dependencies changing)
