# Architecture Research

**Domain:** TUI Calendar v1.3 integration (weekly view, search/filter, overview colors, date format)
**Researched:** 2026-02-06
**Confidence:** HIGH -- based on full codebase review of all 20 source files (2,492 LOC)

## Current Architecture (Reference)

The app follows Bubble Tea's Elm Architecture with clean component boundaries:

```
main.go
  |
  app.Model (root orchestrator)
  |-- calendar.Model   (left pane: grid + overview)
  |-- todolist.Model   (right pane: todo list + input)
  |-- settings.Model   (full-screen overlay when active)
  |
  store.Store          (pure data layer, JSON persistence)
  config.Config        (TOML config, Save/Load)
  theme.Theme          (14 semantic color roles)
  holidays.Provider    (rickar/cal wrapper)
```

**Key architectural patterns already established:**

1. **Message routing:** Root model routes messages to focused child only. Settings overlay intercepts all messages when `showSettings == true`.
2. **State sync:** Calendar month changes trigger `todoList.SetViewMonth()`. Every Update cycle calls `calendar.RefreshIndicators()`.
3. **Theme propagation:** `applyTheme()` on root calls `SetTheme()` on calendar, todolist, and settings. Each component rebuilds its Styles struct.
4. **Settings lifecycle:** `showSettings` bool gates overlay. SaveMsg/CancelMsg bubble up from settings. Live preview via ThemeChangedMsg.
5. **Overlay pattern:** Settings is the only overlay. It replaces the main view entirely in `View()`.
6. **Pure rendering:** `RenderGrid()` is a pure function. `renderOverview()` reads fresh from store.
7. **Input mode state machine:** todolist.Model has 5 modes (normalMode, inputMode, dateInputMode, editTextMode, editDateMode).

**Fixed layout constraints:**
- Calendar grid is exactly 34 chars wide (`gridWidth` constant in grid.go)
- App gives calendar pane 38 chars inner width (34 grid + padding)
- Todo pane gets the remainder: `m.width - 38 - (frameH * 2)`

## Integration Plan

### Feature 1: Weekly Calendar View

**Approach: Add view mode enum to calendar.Model, new `RenderWeekGrid()` pure function.**

The calendar model currently only supports month view. A weekly view needs:

**State changes to calendar.Model:**
```go
type viewMode int
const (
    monthView viewMode = iota
    weekView
)

// Add to Model struct:
viewMode    viewMode
weekOffset  int  // which week within the current month (0-based)
```

**Toggle mechanism:**
- New key binding `w` in calendar KeyMap to toggle `viewMode` between monthView and weekView.
- Root model does NOT need changes for the toggle -- it already routes key messages to calendar when focused.
- The `View()` method dispatches: month view calls `RenderGrid()`, week view calls a new `RenderWeekGrid()`.

**New pure function `RenderWeekGrid()`:**
- Takes same general parameters as RenderGrid (year, month, weekOffset, today, holidays, mondayStart, indicators, styles).
- Renders 7 columns like the month grid but only 1 row of days.
- Can reuse the same Styles struct -- same cell styling logic (today/holiday/indicator/normal).
- Width: Same 34 chars as month grid. No layout changes needed in app.Model.

**Navigation changes:**
- In weekView: left/right arrow moves by 1 week instead of 1 month.
- The calendar's `Update()` needs a branch on `m.viewMode`.
- When weekOffset goes below 0, move to previous month's last week. When weekOffset exceeds the month's week count, move to next month's first week.

**Week offset approach (recommended over time.Time):**
Using a weekOffset integer relative to the current year/month keeps the existing year/month tracking intact. The month grid already computes the first weekday and number of days. The week grid just picks one row from that same data. When navigating past month boundaries, the existing month-advance logic (with year rollover) already works -- just reset weekOffset to 0 or to the last week.

**Overview impact:** The overview panel renders below the grid. In weekView, the grid is shorter (1 row vs 5-6 rows), leaving more vertical space for the overview. No structural change needed.

**Todo list sync:** `SetViewMonth()` still works -- the todolist shows the month containing the displayed week. No change needed to todolist.

**Data flow:**
```
User presses 'w' -> KeyMsg to calendar.Update()
  -> calendar toggles viewMode, sets weekOffset to week containing today (or 0)
  -> calendar.View() calls RenderWeekGrid() instead of RenderGrid()
  -> calendar.renderOverview() unchanged (still reads from store)

User navigates (left/right) in weekView -> calendar.Update()
  -> updates weekOffset by +/-1
  -> if past boundary: advance/retreat month, reset weekOffset
  -> recomputes holidays and indicators for the new month
  -> root syncs todolist via SetViewMonth()
```

**Files modified:**
- `internal/calendar/model.go` -- Add viewMode, weekOffset, toggle logic, navigation branching
- `internal/calendar/keys.go` -- Add ToggleView key binding
- `internal/calendar/grid.go` -- Add `RenderWeekGrid()` pure function (new function, same file)
- `internal/app/model.go` -- Update `currentHelpKeys()` to include the new toggle binding

**Files NOT modified:** store, todolist, config, theme, settings -- weekly view is entirely contained within calendar.

### Feature 2: Search/Filter

**Approach: Two mechanisms -- inline filter in todolist (simple), full-screen search overlay (new component).**

#### 2a: Inline Filter in Todo List

Add a filter mode to the todolist's existing input state machine.

**State changes to todolist.Model:**
```go
// New mode value:
filterMode  // typing filter query

// New fields:
filterQuery string  // current active filter text
```

**Mechanism:**
- New key binding `/` enters filterMode (vim convention for search).
- Reuses existing `textinput.Model` with a different prompt ("Filter: ").
- While `filterQuery != ""`, `visibleItems()` filters todos by case-insensitive substring match on `todo.Text`.
- Esc clears filter and returns to normalMode.
- Enter confirms filter and returns to normalMode (filter stays active).
- A separate key (e.g., Esc when filter is active and mode is normal) clears the filter.

**Integration:**
- The filter happens entirely within `visibleItems()` -- the single function that builds the display list.
- No store changes needed. Filtering is a view-layer concern.
- Cursor needs clamping after filter changes (already handled for delete).

**Data flow:**
```
User presses '/' -> todolist enters filterMode
  -> textinput shown with "Filter: " prompt
  -> visibleItems() applies filterQuery to store results
  -> View() renders filtered list

User presses Esc -> clears filterQuery, back to normalMode
```

#### 2b: Full-Screen Search Overlay

For cross-month search, add a new component following the settings overlay pattern.

**New component: `internal/search/model.go`**

```go
type Model struct {
    input    textinput.Model
    results  []SearchResult
    cursor   int
    width    int
    height   int
    store    *store.Store
    keys     KeyMap
    styles   Styles
}

type SearchResult struct {
    Todo  store.Todo
    Match string  // the matched portion for highlighting
}

// Message types (same pattern as settings):
type SelectMsg struct { TodoID int; Year int; Month time.Month }
type CloseMsg struct{}
```

**Root model integration (same pattern as settings):**
```go
// Add to app.Model:
showSearch bool
search     search.Model

// In Update():
case search.SelectMsg:
    m.showSearch = false
    // Navigate calendar to the result's month
    // Position todolist cursor on the result
case search.CloseMsg:
    m.showSearch = false

// In View():
if m.showSearch {
    return search view + help bar
}
```

**Store addition needed:**
```go
// New method on Store:
func (s *Store) Search(query string) []Todo {
    // Case-insensitive substring match across ALL todos
    // Returns matches sorted by date (most recent first), floating last
}
```

**Key binding:** `S` (shift-s) or `/` at app level opens search overlay. Using `S` avoids conflict with the todolist's `/` for inline filter.

**Data flow:**
```
User presses 'S' -> app.Model sets showSearch=true
  -> creates search.Model with store reference
  -> search.Update() handles typing, queries store.Search()
  -> search.View() shows results with highlighted matches

User selects result (enter) -> search.SelectMsg bubbles up
  -> app.Model navigates calendar to result's month
  -> app.Model syncs todolist to that month
  -> app.Model sets showSearch=false
```

**Files created:**
- `internal/search/model.go` -- Search overlay component
- `internal/search/keys.go` -- Key bindings
- `internal/search/styles.go` -- Themed styles

**Files modified:**
- `internal/store/store.go` -- Add `Search()` method
- `internal/todolist/model.go` -- Add filterMode, filterQuery, filter logic in `visibleItems()`
- `internal/todolist/keys.go` -- Add Filter key binding
- `internal/app/model.go` -- Add showSearch, search field, message routing, overlay logic
- `internal/app/keys.go` -- Add Search key binding

### Feature 3: Overview Color Coding

**Approach: Modify existing `renderOverview()` and add 2 new color roles to Theme.**

The overview panel currently renders all month counts in `OverviewCount` style (muted) with `OverviewActive` style (bold) for the current month. There is no distinction between completed and incomplete.

**Theme additions:**
```go
// Add to Theme struct:
OverviewDoneFg    lipgloss.Color  // completed count color (green family)
OverviewPendingFg lipgloss.Color  // uncompleted count color (red family)
```

All 4 themes (Dark, Light, Nord, Solarized) need these 2 new fields populated with appropriate colors.

**Store changes:**
The current `TodoCountsByMonth()` returns total count per month. We need completed vs incomplete breakdown.

```go
type MonthCount struct {
    Year    int
    Month   time.Month
    Count   int
    Done    int   // NEW: completed count
    Pending int   // NEW: incomplete count
}
```

Modify `TodoCountsByMonth()` to populate Done and Pending alongside Count.

**Calendar styles additions:**
```go
// Add to calendar.Styles:
OverviewDone    lipgloss.Style
OverviewPending lipgloss.Style
```

**renderOverview() changes:**
Instead of rendering `[count]`, render something like `[done/total]` where the color indicates status. The simplest effective approach: color the count number red if there are pending items, green if all done.

**Recommended format:** Keep `[count]` format for simplicity but color it: green (`OverviewDone`) when all todos in that month are completed, red/warm (`OverviewPending`) when any are incomplete. The active month line keeps its bold styling on top.

**Data flow:**
```
renderOverview() calls store.TodoCountsByMonth()
  -> each MonthCount now has Done and Pending
  -> format string uses conditional styling:
     if Pending > 0: count rendered with OverviewPending style
     if Pending == 0: count rendered with OverviewDone style
```

**Files modified:**
- `internal/theme/theme.go` -- Add OverviewDoneFg, OverviewPendingFg to Theme struct; add values to all 4 themes
- `internal/calendar/styles.go` -- Add OverviewDone, OverviewPending to Styles struct; wire in NewStyles()
- `internal/calendar/model.go` -- Modify `renderOverview()` to use new styles based on done/pending
- `internal/store/store.go` -- Modify `MonthCount` struct and `TodoCountsByMonth()` to include done/pending

**Files NOT modified:** todolist, app, config, settings (overview colors are contained within calendar+store+theme).

### Feature 4: Date Format

**Approach: Config field + format string propagated through the app.**

This is the most cross-cutting feature. Dates currently display as raw `YYYY-MM-DD` strings everywhere.

**Where dates are displayed (audit of all display points):**

1. **todolist.renderTodo()** (line 478): `t.Date` rendered raw with `m.styles.Date` style
2. **todolist.updateDateInputMode()** (line 332): Validates input against `"2006-01-02"` format
3. **todolist.updateEditDateMode()** (line 394): Validates input against `"2006-01-02"` format
4. **todolist input placeholder** (line 275): Shows `"YYYY-MM-DD (empty = floating)"`
5. **calendar.RenderGrid()**: Days shown as numbers only (no full dates) -- unaffected
6. **calendar.renderOverview()**: Shows month names only -- unaffected
7. **store/todo.go**: `dateFormat = "2006-01-02"` is the STORAGE format -- must NOT change

**Critical distinction: storage format vs display format.**

The storage format (`"2006-01-02"` / YYYY-MM-DD) in store/todo.go is the internal representation. It must remain fixed for data integrity. The display format is a presentation concern.

**Config addition:**
```go
type Config struct {
    Country        string `toml:"country"`
    FirstDayOfWeek string `toml:"first_day_of_week"`
    Theme          string `toml:"theme"`
    DateFormat     string `toml:"date_format"`  // NEW
}
```

**Preset values:**
- `"iso"` -> `"2006-01-02"` (YYYY-MM-DD) -- default, matches storage
- `"eu"` -> `"02.01.2006"` (DD.MM.YYYY)
- `"us"` -> `"01/02/2006"` (MM/DD/YYYY)

Custom format support can be deferred to later if desired.

**Format propagation path:**

The date format needs to reach every component that displays dates. Use the same setter pattern as `SetTheme()`:

Add `SetDateFormat(string)` to todolist.Model. The root model calls it on init and after settings save.

**Format conversion helper in config package:**
```go
func (c Config) DateDisplayFormat() string {
    switch c.DateFormat {
    case "eu":  return "02.01.2006"
    case "us":  return "01/02/2006"
    default:    return "2006-01-02"  // iso
}

func (c Config) DatePlaceholder() string {
    switch c.DateFormat {
    case "eu":  return "DD.MM.YYYY"
    case "us":  return "MM/DD/YYYY"
    default:    return "YYYY-MM-DD"
    }
}
```

**Settings integration:** Add a "Date Format" option row to the settings overlay, same cycling pattern as theme/country/first-day-of-week. No live preview needed (unlike theme) -- date format takes effect on save.

**Input validation update:** When the user types a date in todolist (dateInputMode, editDateMode), the input prompt and validation must adapt to the configured display format. Parse the user's input using the display format, then convert to storage format (YYYY-MM-DD) before calling `store.Add()` or `store.Update()`.

Bidirectional conversion:
```
Display -> Storage:  time.Parse(displayFormat, input) -> time.Format("2006-01-02")
Storage -> Display:  time.Parse("2006-01-02", stored) -> time.Format(displayFormat)
```

**Data flow:**
```
Config.DateFormat = "eu"
  -> config.DateDisplayFormat() returns "02.01.2006"
  -> app.New() calls todolist.SetDateFormat("02.01.2006")
  -> todolist.renderTodo() converts stored "2025-03-15" -> "15.03.2025"
  -> todolist date input prompt shows "DD.MM.YYYY" instead of "YYYY-MM-DD"
  -> todolist date input validation parses with display format
  -> on confirm, converts back to "2006-01-02" for storage

settings.SaveMsg with new DateFormat
  -> app.Model calls todolist.SetDateFormat(cfg.DateDisplayFormat())
  -> all displayed dates update on next render
```

**Files modified:**
- `internal/config/config.go` -- Add DateFormat field, DefaultConfig, DateDisplayFormat(), DatePlaceholder()
- `internal/todolist/model.go` -- Add dateFormat field, SetDateFormat(), modify renderTodo() and date input/validation
- `internal/settings/model.go` -- Add "Date Format" option row
- `internal/app/model.go` -- Call todolist.SetDateFormat() on init and settings save

**Files NOT modified:** store (storage format stays YYYY-MM-DD), calendar (no full dates displayed), theme.

## New Components

| Component | Package | Files | Purpose |
|-----------|---------|-------|---------|
| Search overlay | `internal/search/` | model.go, keys.go, styles.go | Full-screen cross-month todo search |

This is the only new package. All other features integrate into existing components.

## Modified Components

| Component | Feature | Changes |
|-----------|---------|---------|
| **calendar.Model** | Weekly view | Add viewMode, weekOffset, toggle key, navigation branching |
| **calendar.Model** | Overview colors | Modify renderOverview() to use done/pending styles |
| **calendar/grid.go** | Weekly view | Add `RenderWeekGrid()` pure function |
| **calendar.Styles** | Overview colors | Add OverviewDone, OverviewPending styles |
| **calendar.KeyMap** | Weekly view | Add ToggleView binding |
| **todolist.Model** | Search/filter | Add filterMode, filterQuery, filter logic in visibleItems() |
| **todolist.Model** | Date format | Add dateFormat field, SetDateFormat(), format conversion in renderTodo() and input validation |
| **todolist.KeyMap** | Search/filter | Add Filter binding |
| **store.Store** | Search | Add Search() method |
| **store.Store** | Overview colors | Modify MonthCount struct and TodoCountsByMonth() to include done/pending |
| **config.Config** | Date format | Add DateFormat field, DateDisplayFormat(), DatePlaceholder() |
| **theme.Theme** | Overview colors | Add OverviewDoneFg, OverviewPendingFg (2 new color roles -> 16 total) |
| **settings.Model** | Date format | Add "Date Format" option row |
| **app.Model** | Search | Add showSearch, search field, message routing |
| **app.Model** | Date format | Call SetDateFormat() on init and settings save |
| **app.KeyMap** | Search | Add Search binding |

## Data Flow Changes

### Current Data Flow (Reference)
```
store.TodosForMonth() -> todolist.visibleItems() -> todolist.View()
store.IncompleteTodosPerDay() -> calendar.indicators -> RenderGrid()
store.TodoCountsByMonth() -> calendar.renderOverview()
config.Load() -> app.New() -> child constructors
settings.SaveMsg -> config.Save() -> app applies to children
```

### New Data Flows

**Weekly View:**
```
calendar.viewMode toggle -> calendar.View() dispatches to RenderGrid or RenderWeekGrid
calendar weekOffset navigation -> if month boundary crossed: advance month, reset offset
  -> recompute holidays/indicators -> root syncs todolist month
```

**Inline Filter:**
```
todolist.filterQuery -> todolist.visibleItems() applies substring match -> filtered display
(store unchanged, filtering is pure view logic)
```

**Search Overlay:**
```
app.showSearch -> search.Model gets input -> store.Search(query) -> results displayed
search.SelectMsg -> app navigates calendar to result's month -> syncs todolist
```

**Overview Colors:**
```
store.TodoCountsByMonth() now returns Done/Pending per month
calendar.renderOverview() applies OverviewDone/OverviewPending styles conditionally
```

**Date Format:**
```
config.DateFormat -> config.DateDisplayFormat() -> todolist.SetDateFormat()
todolist.renderTodo(): store date -> parse "2006-01-02" -> format displayFormat -> display
todolist date input: user input -> parse displayFormat -> format "2006-01-02" -> store
settings DateFormat option -> SaveMsg -> app calls SetDateFormat()
```

## Suggested Build Order

Based on dependency analysis and risk assessment:

### Phase 10: Overview Color Coding
**Why first:**
- Smallest scope (4 files modified, no new packages)
- Zero risk of breaking existing functionality -- purely additive
- Contained within calendar + store + theme -- no app.Model routing changes
- Builds confidence with a quick win
- Dependencies: None

### Phase 11: Date Format
**Why second:**
- Cross-cutting but mechanically straightforward (config field + format propagation)
- Should be done before search overlay (search results display dates in configured format)
- Settings integration follows the exact same cycling pattern already established
- The bidirectional format conversion is the main complexity; better to solve it early
- Dependencies: None, but should precede search

### Phase 12: Weekly Calendar View
**Why third:**
- Entirely self-contained within calendar package (no new packages needed)
- Calendar's existing pure-function pattern (RenderGrid) extends naturally to RenderWeekGrid
- No store changes, no todolist changes beyond existing month sync
- Dependencies: None, but benefits from date format being available for week header

### Phase 13: Search/Filter
**Why last:**
- Most complex: new package (search overlay) + todolist filter mode + store method + app routing
- The search overlay follows the settings overlay pattern, so it benefits from all prior phases being stable
- Search results should display dates in configured format (date format dependency)
- Adding a second overlay to app.Model is the most architecturally impactful change
- Dependencies: Benefits from date format (Phase 11)

### Alternative: Split search into two sub-phases
If Phase 13 feels too large:
- **13a:** Inline filter in todolist (contained, no new packages, ~2 files changed)
- **13b:** Full-screen search overlay (new package, app routing, 5+ files)

## Anti-Patterns to Avoid

### 1. Do NOT merge view mode state into app.Model
The weekly/monthly toggle belongs in calendar.Model. The root model should not know or care which calendar view mode is active. It already delegates rendering to `m.calendar.View()`.

**Wrong:** `app.Model.calendarViewMode` with conditional rendering in `app.View()`.
**Right:** `calendar.Model.viewMode` with dispatch in `calendar.View()`.

### 2. Do NOT add a third overlay state enum
With settings and search both being overlays, keep two independent bools (`showSettings`, `showSearch`). They are mutually exclusive by construction (key bindings are suppressed when either is open). An enum adds complexity for no benefit with exactly 2 overlays.

### 3. Do NOT change the storage date format
The `store/todo.go` `dateFormat = "2006-01-02"` is the internal representation and must stay fixed. Date format configuration is purely a display concern. All store methods, `InMonth()`, and `HasDate()` must continue using ISO format.

### 4. Do NOT filter in the Store layer
Inline filtering is a view concern. `visibleItems()` already builds the display list from store data. Adding filter logic there keeps the store as a pure data layer. A `store.Search()` method is appropriate for the search overlay (cross-month query), but the inline filter should NOT add a method to Store.

### 5. Do NOT make RenderWeekGrid depend on RenderGrid
They should be two independent pure functions. Extracting shared helpers (cell styling, weekday headers) is fine, but do not try to make RenderGrid "configurable" to handle both modes. The month grid has complex multi-row layout; the week grid is a single row.

### 6. Do NOT break the fixed 34-char grid width
Both RenderGrid and RenderWeekGrid must produce 34-char wide output. The app layout depends on `calendarInnerWidth := 38`. Changing this would cascade into app.View() layout math.

### 7. Do NOT propagate date format through the Store
Store methods return `[]Todo` with `.Date` in storage format. Conversion to display format happens at the View layer (todolist.renderTodo). Do not add format parameters to Store methods.

## Search Result Navigation Detail

When a user selects a search result, the app needs to:
1. Set calendar to the result's year/month
2. Sync todolist to that month
3. Position the todolist cursor on the matching todo

Step 3 requires `SelectMsg` to carry the todo ID. The todolist needs a `SetCursorToTodoID(id int)` method that finds the todo in `visibleItems()` and sets `m.cursor` accordingly.

```go
type SelectMsg struct {
    TodoID int
    Year   int
    Month  time.Month
}

// In app.Model.Update():
case search.SelectMsg:
    m.showSearch = false
    m.calendar.SetYearMonth(msg.Year, msg.Month)  // new setter needed on calendar
    m.todoList.SetViewMonth(msg.Year, msg.Month)
    m.todoList.SetCursorToTodo(msg.TodoID)         // new method on todolist
    m.activePane = todoPane
    m.calendar.SetFocused(false)
    m.todoList.SetFocused(true)
```

This also requires a `SetYearMonth(int, time.Month)` setter on calendar.Model to navigate without user key input.

## Sources

- Full codebase review: all 20 source files in `internal/` (app, calendar, todolist, store, config, theme, settings, holidays)
- Bubble Tea Elm Architecture patterns observed in existing code
- Go `time` package format strings (`"2006-01-02"` layout convention)
- Existing settings overlay pattern (SaveMsg/CancelMsg/ThemeChangedMsg)
- Existing theme propagation pattern (SetTheme + Styles struct + NewStyles constructor)
