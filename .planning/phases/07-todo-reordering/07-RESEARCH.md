# Phase 7: Todo Reordering - Research

**Researched:** 2026-02-06
**Domain:** List item reordering in Bubble Tea, sort order persistence in JSON, backwards-compatible data model changes
**Confidence:** HIGH

## Summary

Phase 7 adds manual reordering to the todo list: users press a key to move the selected todo up or down in the list (REORD-01, REORD-02), the custom order persists across restarts (REORD-03), and reorder keybindings appear in the help bar when a todo is selected.

The existing codebase determines todo order via sort functions: `TodosForMonth()` sorts by date ascending then ID, and `FloatingTodos()` sorts by ID ascending. There is no explicit order field. To support reordering, a `sort_order` integer field must be added to the `Todo` struct. This field acts as the primary sort key within each section (dated month, floating), with the existing date/ID sort as tiebreaker for todos with the same `sort_order` (which ensures backwards compatibility -- existing todos with `sort_order: 0` retain their current ordering by date/ID).

The reordering operation itself is simple slice manipulation in the store: swap the `sort_order` values of two adjacent todos and persist. The todolist model handles the key events, identifies the currently selected todo and its neighbor, calls a store method, and adjusts the cursor. No new dependencies are needed. All patterns follow the established codebase conventions from Phases 3-6.

The main subtlety is that reordering must respect section boundaries -- moving a dated todo "up" at the top of the dated section must not swap it with a floating todo, and vice versa. The reorder operation is scoped to the visible list within a section.

**Primary recommendation:** Add a `sort_order` integer field to `Todo` (JSON `"sort_order,omitempty"` for backwards compatibility). Add `MoveUp(id)` and `MoveDown(id)` methods to the store that swap `sort_order` values between adjacent todos within the same section. Add `K`/`J` (shift+k/shift+j) keybindings for move-up/move-down. Update `TodosForMonth()` and `FloatingTodos()` to sort by `sort_order` first.

## Standard Stack

No new libraries are needed. Phase 7 uses only the existing stack.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `internal/store` | project | New `sort_order` field, `MoveUp()`/`MoveDown()` methods | Follows existing `Toggle`/`Delete`/`Update` pattern |
| `bubbles/key` | v0.21.1 | New `MoveUp` and `MoveDown` key bindings | Already used for all other keybindings |

### Existing (unchanged)
| Library | Version | Purpose |
|---------|---------|---------|
| Bubble Tea | v1.3.10 | TUI framework |
| Lip Gloss | v1.1.0 | Terminal styling |
| Bubbles | v0.21.1 | `key.Binding`, `textinput.Model`, `help.Model` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `sort_order` integer field | Array position as implicit order | Array position breaks when filtering by month or floating -- the same `Todos` slice contains all todos, but they are displayed in different sections. An explicit `sort_order` lets each section sort independently. |
| `sort_order` integer field | Linked list (next_id field) | Over-engineered for this use case. Swapping two items in a linked list requires updating up to 4 pointers vs simply swapping two integers. |
| `K`/`J` (Shift+k/j) for move | `ctrl+up`/`ctrl+down` | Shift+letter is consistent with the existing `a`/`A` (add/add-dated) and `e`/`E` (edit-text/edit-date) conventions where shift = variant of same key. Also, terminal ctrl+arrow support is inconsistent across terminals. |
| Swapping `sort_order` values | Renumbering all items | Swap is O(1) and minimal; renumbering is O(n) and creates unnecessary writes. Swap is sufficient since we only move one position at a time. |

**Installation:** No new `go get` needed.

## Architecture Patterns

### Files Modified
```
internal/
  store/
    todo.go       # Add SortOrder field to Todo struct
    store.go      # Update sort functions, add MoveUp/MoveDown methods, add EnsureSortOrder()
  todolist/
    model.go      # Add move key handlers in updateNormalMode, update HelpBindings
    keys.go       # Add MoveUp and MoveDown key bindings
```

### Pattern 1: SortOrder Field with Backwards Compatibility
**What:** Add a `SortOrder int` field to the `Todo` struct with JSON tag `"sort_order,omitempty"`. When loading an existing `todos.json` that lacks `sort_order` fields, all todos get `sort_order: 0`. The sort functions use `sort_order` as the primary key, then fall back to the existing date/ID ordering for ties. Since all legacy todos have `sort_order: 0`, they tie and fall through to existing logic -- preserving the exact same order as before.

**When to use:** Always -- this is the data model change.

**Confidence:** HIGH (Go's `json.Unmarshal` defaults missing int fields to 0; `omitempty` omits 0-valued ints from JSON output, keeping the file clean for legacy todos)

**Example:**
```go
// internal/store/todo.go
type Todo struct {
    ID        int    `json:"id"`
    Text      string `json:"text"`
    Date      string `json:"date,omitempty"`
    Done      bool   `json:"done"`
    CreatedAt string `json:"created_at"`
    SortOrder int    `json:"sort_order,omitempty"`
}
```

### Pattern 2: Sort Order Initialization on Load
**What:** After loading from disk, call `EnsureSortOrder()` to assign unique `sort_order` values to any todos that have `sort_order: 0` (i.e., legacy todos that were created before this feature). This is done once at load time, and only writes if changes were made. The initialization assigns incrementing values based on the current position in the slice (which is the legacy order by insertion time).

**When to use:** In `NewStore()` after `load()` succeeds.

**Confidence:** HIGH (one-time migration ensures all todos have meaningful sort orders going forward)

**Example:**
```go
// internal/store/store.go
func (s *Store) EnsureSortOrder() {
    needsSave := false
    for i := range s.data.Todos {
        if s.data.Todos[i].SortOrder == 0 {
            s.data.Todos[i].SortOrder = (i + 1) * 10
            needsSave = true
        }
    }
    if needsSave {
        s.Save()
    }
}
```

Using increments of 10 leaves gaps for future insertion-at-position without renumbering (not needed now, but free to include).

### Pattern 3: Updated Sort Functions
**What:** Modify `TodosForMonth()` and `FloatingTodos()` to sort by `SortOrder` as primary key. For `TodosForMonth()`, the full sort is: `SortOrder` ascending, then `Date` ascending, then `ID` ascending. For `FloatingTodos()`: `SortOrder` ascending, then `ID` ascending.

**When to use:** These replace the existing sort logic in store.go.

**Confidence:** HIGH (trivial extension of existing `sort.Slice` calls)

**Example:**
```go
// TodosForMonth
sort.Slice(result, func(i, j int) bool {
    if result[i].SortOrder != result[j].SortOrder {
        return result[i].SortOrder < result[j].SortOrder
    }
    if result[i].Date != result[j].Date {
        return result[i].Date < result[j].Date
    }
    return result[i].ID < result[j].ID
})

// FloatingTodos
sort.Slice(result, func(i, j int) bool {
    if result[i].SortOrder != result[j].SortOrder {
        return result[i].SortOrder < result[j].SortOrder
    }
    return result[i].ID < result[j].ID
})
```

### Pattern 4: Store MoveUp / MoveDown Methods
**What:** Store methods that take a todo ID and a "neighbor ID" and swap their `SortOrder` values, then persist. The todolist model is responsible for determining which todo is the neighbor (the one visually above or below in the current section). The store just does the swap.

Alternatively, the store can expose a `SwapOrder(id1, id2 int)` method and the todolist model calls it with the correct pair.

**When to use:** When the user triggers move-up or move-down.

**Confidence:** HIGH (trivial swap + save)

**Example:**
```go
// internal/store/store.go

// SwapOrder swaps the SortOrder values of two todos and persists.
func (s *Store) SwapOrder(id1, id2 int) {
    var t1, t2 *Todo
    for i := range s.data.Todos {
        switch s.data.Todos[i].ID {
        case id1:
            t1 = &s.data.Todos[i]
        case id2:
            t2 = &s.data.Todos[i]
        }
    }
    if t1 != nil && t2 != nil {
        t1.SortOrder, t2.SortOrder = t2.SortOrder, t1.SortOrder
        s.Save()
    }
}
```

### Pattern 5: Todolist Model Move Handlers
**What:** In `updateNormalMode`, handle `MoveUp` and `MoveDown` keys. The handler:
1. Gets the visible items and selectable indices
2. Finds the currently selected todo
3. Determines the neighbor (the todo one position up/down in the **same section**)
4. Calls `store.SwapOrder(selectedID, neighborID)`
5. Moves the cursor to follow the moved todo

**When to use:** When user presses the move keybinding.

**Confidence:** HIGH (follows existing toggle/delete pattern for cursor management)

**Section boundary detection:** The `visibleItems()` list interleaves headers, todos, and empty placeholders. To determine section boundaries, we need to check if the todo above/below the cursor is in the same section. The simplest approach: build the selectable items list, check if the neighbor selectable index maps to a todo in the same section (both dated, or both floating).

**Example:**
```go
case key.Matches(msg, m.keys.MoveUp):
    if len(selectable) > 0 && m.cursor > 0 && m.cursor < len(selectable) {
        curIdx := selectable[m.cursor]
        prevIdx := selectable[m.cursor-1]
        curTodo := items[curIdx].todo
        prevTodo := items[prevIdx].todo
        // Only swap within same section (both dated or both floating)
        if curTodo != nil && prevTodo != nil &&
            curTodo.HasDate() == prevTodo.HasDate() {
            m.store.SwapOrder(curTodo.ID, prevTodo.ID)
            m.cursor--
        }
    }

case key.Matches(msg, m.keys.MoveDown):
    if len(selectable) > 0 && m.cursor < len(selectable)-1 {
        curIdx := selectable[m.cursor]
        nextIdx := selectable[m.cursor+1]
        curTodo := items[curIdx].todo
        nextTodo := items[nextIdx].todo
        // Only swap within same section (both dated or both floating)
        if curTodo != nil && nextTodo != nil &&
            curTodo.HasDate() == nextTodo.HasDate() {
            m.store.SwapOrder(curTodo.ID, nextTodo.ID)
            m.cursor++
        }
    }
```

### Pattern 6: Section Boundary Check via HasDate
**What:** Two todos are in the same section if they have the same "has date" status. Dated todos with dates in the viewed month appear in the month section; floating todos (no date) appear in the floating section. The `HasDate()` method already exists on `Todo`. Comparing `curTodo.HasDate() == neighborTodo.HasDate()` determines if they are in the same section.

**Why this works:** The `visibleItems()` function builds the list with month todos first, then floating todos. Within `selectableIndices`, the dated todos come before floating todos. So if the current todo `HasDate()` and the neighbor does not (or vice versa), they are in different sections.

**Edge case:** Dated todos from different months would not appear in the same visible list (the view filters by month), so cross-month swaps cannot happen.

**Confidence:** HIGH (verified by reading `visibleItems()` and `TodosForMonth()`)

### Pattern 7: New Todo Sort Order Assignment
**What:** When `Add()` creates a new todo, assign it a `SortOrder` that places it at the end of its section. The simplest approach: find the maximum `SortOrder` of all existing todos and add 10.

**When to use:** In `store.Add()`.

**Confidence:** HIGH (straightforward extension)

**Example:**
```go
func (s *Store) Add(text string, date string) Todo {
    maxOrder := 0
    for _, t := range s.data.Todos {
        if t.SortOrder > maxOrder {
            maxOrder = t.SortOrder
        }
    }
    t := Todo{
        ID:        s.data.NextID,
        Text:      text,
        Date:      date,
        Done:      false,
        CreatedAt: time.Now().Format(dateFormat),
        SortOrder: maxOrder + 10,
    }
    s.data.NextID++
    s.data.Todos = append(s.data.Todos, t)
    s.Save()
    return t
}
```

### Anti-Patterns to Avoid
- **Using array index as implicit order:** The `Todos` slice contains ALL todos. Dated and floating todos are filtered separately. Array position is meaningless across sections.
- **Moving todos across sections with reorder:** Moving a floating todo into the dated section (or vice versa) requires date assignment, not reordering. Reorder stays within a section.
- **Modifying cursor without following the moved item:** After a swap, the cursor must move with the todo (decrement on move-up, increment on move-down) so the user can keep pressing the key to move the item multiple positions.
- **Allowing reorder in non-normal mode:** Reorder keys should only work in `normalMode`, just like all other keybindings.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Sort order persistence | Custom ordering file | `sort_order` field in existing `Todo` struct/JSON | Single source of truth, atomic saves already work |
| Sort stability | Custom stable sort | `sort.Slice` with multi-key comparison | Go's `sort.Slice` is stable enough with explicit tiebreakers (date, ID) |
| Backwards compatibility | Migration script | `json:"sort_order,omitempty"` + `EnsureSortOrder()` on load | Go's zero-value semantics handle missing fields automatically |
| Section boundary detection | Complex section tracking | `HasDate()` comparison between current and neighbor | Already exists, perfectly indicates section membership |

**Key insight:** The entire reordering feature is just: (1) one new int field on Todo, (2) updated sort comparators, (3) a swap method on store, (4) two key handlers in the todolist model. No new patterns, no new dependencies.

## Common Pitfalls

### Pitfall 1: Cross-Section Swap
**What goes wrong:** User is at the bottom of the dated section and presses move-down. The next selectable item is in the floating section. Swapping their sort orders makes no sense -- it would change relative order within different sections unpredictably.
**Why it happens:** The selectable indices list combines dated and floating todos. The cursor can be at a boundary.
**How to avoid:** Before swapping, check that both todos are in the same section using `HasDate()` comparison. If they are not, the move is a no-op.
**Warning signs:** A dated todo appearing in unexpected position after moving down, or a floating todo appearing in the wrong position after moving up.

### Pitfall 2: Sort Order Collision After Legacy Migration
**What goes wrong:** All legacy todos get `sort_order: 0`. If `EnsureSortOrder()` is not called, all todos tie on `sort_order` and fall through to date/ID ordering. This works for display but breaks after the first swap -- swapping two todos that both had `sort_order: 0` would give one `0` and the other `0` (no change).
**Why it happens:** Without initialization, all todos share the same `sort_order` value.
**How to avoid:** Call `EnsureSortOrder()` in `NewStore()` to assign unique values to all legacy todos before any reorder operations.
**Warning signs:** Move-up/move-down appears to have no effect on legacy data.

### Pitfall 3: Cursor Not Following the Moved Item
**What goes wrong:** User presses move-down. The todo moves down, but the cursor stays in place. Now the cursor is on a different todo. The user presses move-down again expecting to keep moving the same todo, but instead moves a different one.
**Why it happens:** Forgetting to update `m.cursor` after the swap.
**How to avoid:** After move-down, increment `m.cursor`. After move-up, decrement `m.cursor`. This keeps the cursor on the same todo.
**Warning signs:** User has to manually re-navigate to the moved todo to continue moving it.

### Pitfall 4: visibleItems Returns Copies, Not Pointers to Store Data
**What goes wrong:** The `visibleItems()` method creates `visibleItem` structs with pointers to local copies of `Todo` values (from `TodosForMonth` and `FloatingTodos` which return `[]Todo` by value). The `todo.ID` is still valid for identifying the todo in the store, but the `SortOrder` value read from `items[idx].todo.SortOrder` is a snapshot, not a live reference.
**Why it happens:** `TodosForMonth()` and `FloatingTodos()` return `[]Todo` (value slices), not `[]*Todo`.
**How to avoid:** Use `todo.ID` to identify todos, not pointers. Pass IDs to `SwapOrder()`. The store method looks up todos by ID internally. This is already the pattern used by `Toggle`, `Delete`, and `Update`.
**Warning signs:** Sort order appears to not persist after swap.

### Pitfall 5: HelpBindings Overflow
**What goes wrong:** Adding two more keybindings (`K` move up, `J` move down) to the help bar makes it too wide for narrow terminals. The help bar already shows 8 bindings in normal mode.
**Why it happens:** The `bubbles/help` component truncates at `help.Width`.
**How to avoid:** The `help.Model` already handles truncation gracefully. But consider that 10 bindings may be a lot. Options: (a) just add them and let the help model truncate if needed, (b) only show move bindings when a todo is selected (they are always shown in normal mode since a todo is always selected if any exist). The simplest approach is (a) -- add them to the list. The help component handles overflow.
**Warning signs:** Move bindings not visible in the help bar on narrow terminals.

### Pitfall 6: New Todos Added Without SortOrder
**What goes wrong:** If `Add()` is not updated to assign a `SortOrder`, new todos get `SortOrder: 0`. This places them before all other todos (which have positive sort orders from migration). The new todo appears at the top instead of the bottom.
**Why it happens:** Go zero-value for int is 0.
**How to avoid:** Update `Add()` to assign `SortOrder = maxOrder + 10` for new todos.
**Warning signs:** Newly added todos appear at the top of their section instead of the bottom.

## Code Examples

### Updated Todo Struct
```go
// internal/store/todo.go
type Todo struct {
    ID        int    `json:"id"`
    Text      string `json:"text"`
    Date      string `json:"date,omitempty"`
    Done      bool   `json:"done"`
    CreatedAt string `json:"created_at"`
    SortOrder int    `json:"sort_order,omitempty"`
}
```

### EnsureSortOrder Migration
```go
// internal/store/store.go

// EnsureSortOrder assigns unique SortOrder values to any todos that
// have the zero value (legacy data). Called once at load time.
func (s *Store) EnsureSortOrder() {
    needsSave := false
    for i := range s.data.Todos {
        if s.data.Todos[i].SortOrder == 0 {
            s.data.Todos[i].SortOrder = (i + 1) * 10
            needsSave = true
        }
    }
    if needsSave {
        s.Save()
    }
}
```

### SwapOrder Store Method
```go
// internal/store/store.go

// SwapOrder swaps the SortOrder values of two todos identified by ID
// and persists the change.
func (s *Store) SwapOrder(id1, id2 int) {
    var t1, t2 *Todo
    for i := range s.data.Todos {
        switch s.data.Todos[i].ID {
        case id1:
            t1 = &s.data.Todos[i]
        case id2:
            t2 = &s.data.Todos[i]
        }
    }
    if t1 != nil && t2 != nil {
        t1.SortOrder, t2.SortOrder = t2.SortOrder, t1.SortOrder
        s.Save()
    }
}
```

### Updated Add Method
```go
// internal/store/store.go

func (s *Store) Add(text string, date string) Todo {
    maxOrder := 0
    for _, t := range s.data.Todos {
        if t.SortOrder > maxOrder {
            maxOrder = t.SortOrder
        }
    }
    t := Todo{
        ID:        s.data.NextID,
        Text:      text,
        Date:      date,
        Done:      false,
        CreatedAt: time.Now().Format(dateFormat),
        SortOrder: maxOrder + 10,
    }
    s.data.NextID++
    s.data.Todos = append(s.data.Todos, t)
    s.Save()
    return t
}
```

### Updated Sort in TodosForMonth
```go
// internal/store/store.go

sort.Slice(result, func(i, j int) bool {
    if result[i].SortOrder != result[j].SortOrder {
        return result[i].SortOrder < result[j].SortOrder
    }
    if result[i].Date != result[j].Date {
        return result[i].Date < result[j].Date
    }
    return result[i].ID < result[j].ID
})
```

### Updated Sort in FloatingTodos
```go
// internal/store/store.go

sort.Slice(result, func(i, j int) bool {
    if result[i].SortOrder != result[j].SortOrder {
        return result[i].SortOrder < result[j].SortOrder
    }
    return result[i].ID < result[j].ID
})
```

### New Key Bindings
```go
// internal/todolist/keys.go

type KeyMap struct {
    Up       key.Binding
    Down     key.Binding
    MoveUp   key.Binding  // NEW
    MoveDown key.Binding  // NEW
    Add      key.Binding
    AddDated key.Binding
    Toggle   key.Binding
    Delete   key.Binding
    Edit     key.Binding
    EditDate key.Binding
    Confirm  key.Binding
    Cancel   key.Binding
}

// In DefaultKeyMap():
MoveUp: key.NewBinding(
    key.WithKeys("K"),
    key.WithHelp("K", "move up"),
),
MoveDown: key.NewBinding(
    key.WithKeys("J"),
    key.WithHelp("J", "move down"),
),
```

### Move Handlers in updateNormalMode
```go
// internal/todolist/model.go

case key.Matches(msg, m.keys.MoveUp):
    if len(selectable) > 0 && m.cursor > 0 && m.cursor < len(selectable) {
        curIdx := selectable[m.cursor]
        prevIdx := selectable[m.cursor-1]
        curTodo := items[curIdx].todo
        prevTodo := items[prevIdx].todo
        if curTodo != nil && prevTodo != nil &&
            curTodo.HasDate() == prevTodo.HasDate() {
            m.store.SwapOrder(curTodo.ID, prevTodo.ID)
            m.cursor--
        }
    }

case key.Matches(msg, m.keys.MoveDown):
    if len(selectable) > 0 && m.cursor < len(selectable)-1 {
        curIdx := selectable[m.cursor]
        nextIdx := selectable[m.cursor+1]
        curTodo := items[curIdx].todo
        nextTodo := items[nextIdx].todo
        if curTodo != nil && nextTodo != nil &&
            curTodo.HasDate() == nextTodo.HasDate() {
            m.store.SwapOrder(curTodo.ID, nextTodo.ID)
            m.cursor++
        }
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
        m.keys.MoveUp, m.keys.MoveDown,
        m.keys.Add, m.keys.AddDated,
        m.keys.Edit, m.keys.EditDate,
        m.keys.Toggle, m.keys.Delete,
    }
}
```

### Updated ShortHelp and FullHelp in KeyMap
```go
// internal/todolist/keys.go

func (k KeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Up, k.Down, k.MoveUp, k.MoveDown, k.Add, k.AddDated, k.Toggle, k.Delete, k.Edit, k.EditDate}
}

func (k KeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down, k.MoveUp, k.MoveDown, k.Add, k.AddDated, k.Toggle, k.Delete, k.Edit, k.EditDate},
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Implicit order by date/ID | Explicit `sort_order` field as primary sort key | Phase 7 | Users control todo priority |
| No reorder capability | `K`/`J` keys for move up/down within section | Phase 7 | Vim-like muscle memory |

**Deprecated/outdated:**
- Nothing deprecated. Phase 7 extends existing sort logic by adding `sort_order` as a higher-priority key in the existing comparators.

## Open Questions

1. **Should move-up at the very top of a section wrap to the bottom?**
   - What we know: Current navigation (j/k) does not wrap. It stops at boundaries.
   - Recommendation: Do not wrap. Consistency with existing navigation. A no-op at boundary is expected behavior.

2. **Should completed (done) todos be reorderable?**
   - What we know: There is no restriction on toggling or editing completed todos. They appear in the same list.
   - Recommendation: Yes, completed todos should be reorderable like any other todo. No special treatment needed. The `HasDate()` section check is sufficient.

3. **What happens to sort order when a todo's date is edited (changing its section)?**
   - What we know: Phase 5 already handles section changes via cursor clamping. The todo's `sort_order` moves with it.
   - Recommendation: When a todo moves from dated to floating (or vice versa) via date edit, its existing `sort_order` travels with it. Since `sort_order` values are globally unique (not per-section), the todo will appear at whatever position its `sort_order` dictates in the new section. This is acceptable behavior -- the user can re-sort it in the new section with `K`/`J`.

## Sources

### Primary (HIGH confidence)
- Project source: `internal/store/store.go` -- existing sort logic in `TodosForMonth()`, `FloatingTodos()`, `Add()`, `Toggle()`, `Delete()`, `Update()` patterns
- Project source: `internal/store/todo.go` -- existing `Todo` struct, `HasDate()` method, JSON tags with `omitempty`
- Project source: `internal/todolist/model.go` -- existing `visibleItems()`, `selectableIndices()`, `updateNormalMode()` key handling, cursor management
- Project source: `internal/todolist/keys.go` -- existing `KeyMap` struct, `DefaultKeyMap()`, `ShortHelp()`/`FullHelp()` pattern
- Project source: `internal/app/model.go` -- existing help bar aggregation via `currentHelpKeys()`
- Go standard library `encoding/json` -- zero-value behavior for missing fields (int defaults to 0), `omitempty` for int omits 0 values
- Phase 5 research: established patterns for adding new keys, modes, and store methods

### Secondary (MEDIUM confidence)
- None needed -- all patterns verified in existing codebase

### Tertiary (LOW confidence)
- None -- all findings verified with primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- zero new dependencies; all changes are extensions of existing patterns
- Architecture: HIGH -- patterns directly parallel existing Toggle/Delete/Update patterns, verified by reading source
- Pitfalls: HIGH -- derived from careful analysis of visibleItems(), sort logic, and section boundaries
- Code examples: HIGH -- all code follows exact conventions from existing codebase

**Research date:** 2026-02-06
**Valid until:** 2026-03-08 (30 days -- stable libraries, no external dependencies changing)
