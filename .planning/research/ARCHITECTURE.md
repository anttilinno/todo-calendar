# Architecture Research: Priority Levels & Natural Language Date Input

**Domain:** Priority levels (P1-P4) and natural language date input for existing Go/Bubble Tea TUI todo-calendar app
**Researched:** 2026-02-12
**Confidence:** HIGH for priority integration (codebase patterns well-established, change is additive); MEDIUM for NL date parsing (library selection verified, integration design is novel for this codebase)

## Current Architecture Summary (Post v2.0)

```
main.go
  |
  app.Model (root orchestrator)
  |-- calendar.Model    (left pane: grid + overview)
  |-- todolist.Model    (right pane: 4-section todo list + 4-field edit form)
  |-- settings.Model    (full-screen overlay, showSettings bool)
  |-- search.Model      (full-screen overlay, showSearch bool)
  |-- preview.Model     (full-screen overlay, showPreview bool)
  |-- tmplmgr.Model     (full-screen overlay, showTmplMgr bool)
  |-- editor            (external process, editing bool)
  |
  store.SQLiteStore     (TodoStore interface, SQLite with WAL, PRAGMA user_version=6)
  config.Config         (TOML config, 6 settings, Save/Load)
  theme.Theme           (16 semantic color roles across 4 themes)
  holidays.Provider     (rickar/cal wrapper)
  recurring             (AutoCreate engine, ScheduleRule parsing)
  tmpl                  (ExtractPlaceholders, ExecuteTemplate)
  fuzzy                 (fuzzy text matching for search/filter)
```

**Critical facts for this milestone:**

1. **Todo struct** (store/todo.go) has 10 fields: ID, Text, Body, Date, Done, CreatedAt, SortOrder, ScheduleID, ScheduleDate, DatePrecision. No priority field.
2. **TodoStore interface** has 27 methods. `Add(text, date, datePrecision)` and `Update(id, text, date, datePrecision)` are the creation/mutation entry points.
3. **PRAGMA user_version** is at 6. Next migration is version 7.
4. **todoColumns** constant lists all SELECT columns. `scanTodo` scans them. Both must be extended.
5. **Edit form** has 4 fields: Title(0), Date segments(1), Body textarea(2), Template picker(3). `editField int` tracks focus. `SwitchField` (Tab) cycles through them.
6. **Date input** uses 3 segmented textinputs (day/month/year) with format-aware ordering and auto-advance. `deriveDateFromSegments()` produces ISO date + precision.
7. **renderTodo()** renders each line as: cursor + checkbox + text + [+] body indicator + [R] recurring indicator + date.
8. **visibleItems()** builds display list from 4 sections (sectionDated, sectionMonth, sectionYear, sectionFloating). Items are sorted by sort_order within each section.
9. **Styles struct** in todolist has 14 styles. Theme struct has 16 color roles.
10. **Settings overlay** uses `option` structs with cycling values (h/l arrows). SettingChangedMsg propagates changes immediately.

---

## Feature 1: Priority Levels (P1-P4)

### Design Decision: Integer Priority with Display Mapping

Use an integer field `priority` with values 0-4:
- **0** = no priority (default, backward compatible)
- **1** = P1 (urgent/critical)
- **2** = P2 (high)
- **3** = P3 (medium)
- **4** = P4 (low)

**Why integer, not string:** Integers enable `ORDER BY priority` for sorting, simple comparison for filtering, and zero-value means "no priority" (backward compatible with existing todos). String enums ("p1", "p2") require mapping and are harder to sort.

**Why 0 = no priority, not 5:** Zero is the default for SQLite `INTEGER NOT NULL DEFAULT 0`. Existing todos automatically get "no priority" without a data backfill. If we used 5 for "none", we would need `UPDATE todos SET priority = 5` in the migration.

### Schema Migration (Version 7)

```sql
ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
```

Single statement. No backfill needed (DEFAULT 0 = no priority). No index needed for priority alone -- the display order within sections already uses sort_order, and priority filtering is in-app (not SQL).

**Migration code follows the established pattern:**

```go
if version < 7 {
    if _, err := s.db.Exec(`ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0`); err != nil {
        return fmt.Errorf("add priority column: %w", err)
    }
    if _, err := s.db.Exec(`PRAGMA user_version = 7`); err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

### Todo Struct Extension

```go
type Todo struct {
    ID            int    `json:"id"`
    Text          string `json:"text"`
    Body          string `json:"body,omitempty"`
    Date          string `json:"date,omitempty"`
    Done          bool   `json:"done"`
    CreatedAt     string `json:"created_at"`
    SortOrder     int    `json:"sort_order,omitempty"`
    ScheduleID    int    `json:"schedule_id,omitempty"`
    ScheduleDate  string `json:"schedule_date,omitempty"`
    DatePrecision string `json:"date_precision"`
    Priority      int    `json:"priority"`              // NEW: 0=none, 1=P1, 2=P2, 3=P3, 4=P4
}

// PriorityLabel returns "P1"-"P4" or "" for no priority.
func (t Todo) PriorityLabel() string {
    if t.Priority >= 1 && t.Priority <= 4 {
        return fmt.Sprintf("P%d", t.Priority)
    }
    return ""
}

// HasPriority reports whether the todo has a priority set.
func (t Todo) HasPriority() bool {
    return t.Priority >= 1 && t.Priority <= 4
}
```

### TodoStore Interface Changes

The `Add` and `Update` signatures must include priority:

```go
Add(text string, date string, datePrecision string, priority int) Todo
Update(id int, text string, date string, datePrecision string, priority int)
```

**Alternatively**, add a dedicated `UpdatePriority(id, priority int)` method and keep Add/Update unchanged, adding priority as a separate step. However, this approach means two SQL roundtrips for every add/edit, and the form already saves all fields atomically.

**Recommendation: Extend Add/Update signatures.** This is a breaking interface change but the interface has only 2 implementations (SQLiteStore and the compile-time check). All callers are in this codebase. The change is mechanical.

Affected callers:
- `todolist.Model.saveAdd()` -- calls `store.Add()`
- `todolist.Model.saveEdit()` -- calls `store.Update()`
- `store.SQLiteStore.Add()` -- implements Add
- `store.SQLiteStore.Update()` -- implements Update
- `store.SQLiteStore.AddScheduledTodo()` -- scheduled todos get priority 0

### todoColumns and scanTodo Extension

```go
const todoColumns = "id, text, body, date, done, created_at, sort_order, schedule_id, schedule_date, date_precision, priority"

func scanTodo(scanner interface{ Scan(...any) error }) (Todo, error) {
    var t Todo
    var date sql.NullString
    var done int
    var scheduleID sql.NullInt64
    var scheduleDate sql.NullString
    err := scanner.Scan(&t.ID, &t.Text, &t.Body, &date, &done, &t.CreatedAt, &t.SortOrder, &scheduleID, &scheduleDate, &t.DatePrecision, &t.Priority)
    // ... rest unchanged
}
```

### Edit Form Integration

Add priority as a new field in the edit form. Current field cycle: Title(0) -> Date(1) -> Body(2) -> Template(3) -> Title(0).

**New cycle:** Title(0) -> Date(1) -> Priority(2) -> Body(3) -> Template(4) -> Title(0).

**Why priority before body:** Priority is a quick single-keystroke selection. Body is a large textarea. Placing priority before body keeps the "quick fields first, large fields last" ordering.

**Priority input widget:** A simple cycling selector, not a text input. Display the current priority and cycle with left/right arrows or typed digits 0-4:

```
Priority: < P2 - High >
```

Or more simply, since the edit form uses Tab to advance fields:

```
Priority: [none]   (type 1-4 to set, 0 or backspace to clear)
```

**Implementation:** A single textinput with 1-character limit accepting only digits 0-4. Or simpler: a stateful field on the Model that cycles on keypress (like settings options).

**Recommendation: Numeric keypress.** In the priority field, pressing 1/2/3/4 sets priority, pressing 0 or backspace clears it. No separate widget needed -- just intercept keystrokes when `editField == 2` (the new priority position).

```go
// In Model:
editPriority int  // 0=none, 1-4=priority level

// In editField routing for priority:
case 2: // priority field
    switch msg.String() {
    case "1": m.editPriority = 1
    case "2": m.editPriority = 2
    case "3": m.editPriority = 3
    case "4": m.editPriority = 4
    case "0", "backspace": m.editPriority = 0
    }
```

### Display Integration

Render priority as a colored badge before the todo text:

```
> [x] [P1] Fix critical bug                    2026-02-12
  [ ] [P2] Review pull request            [+]  2026-02-13
  [ ]      Normal todo without priority         2026-02-14
```

**In renderTodo():**

```go
func (m Model) renderTodo(b *strings.Builder, t *store.Todo, selected bool) {
    // Cursor indicator
    if selected {
        b.WriteString(m.styles.Cursor.Render("> "))
    } else {
        b.WriteString("  ")
    }

    // Checkbox
    if t.Done {
        b.WriteString(m.styles.CheckboxDone.Render("[x]"))
    } else {
        b.WriteString(m.styles.Checkbox.Render("[ ]"))
    }
    b.WriteString(" ")

    // Priority badge (NEW)
    if t.HasPriority() {
        style := m.priorityStyle(t.Priority)
        b.WriteString(style.Render("[" + t.PriorityLabel() + "]"))
        b.WriteString(" ")
    }

    // Text content
    text := t.Text
    if t.Done {
        text = m.styles.Completed.Render(text)
    }
    b.WriteString(text)

    // ... body indicator, recurring indicator, date (unchanged)
}
```

### Theme Extension for Priority Colors

Add 4 new color roles to the Theme struct:

```go
type Theme struct {
    // ... existing 16 fields ...

    // Priority levels
    PriorityP1 lipgloss.Color // P1 urgent (red/warm)
    PriorityP2 lipgloss.Color // P2 high (orange/yellow)
    PriorityP3 lipgloss.Color // P3 medium (blue/cyan)
    PriorityP4 lipgloss.Color // P4 low (muted/grey)
}
```

**Color assignments per theme:**

| Theme | P1 | P2 | P3 | P4 |
|-------|----|----|----|----|
| Dark | `#D75F5F` (rose) | `#D7AF5F` (gold) | `#5F87D7` (blue) | `#585858` (muted) |
| Light | `#D70000` (red) | `#AF8700` (amber) | `#005FAF` (blue) | `#8A8A8A` (grey) |
| Nord | `#BF616A` (aurora red) | `#EBCB8B` (aurora yellow) | `#81A1C1` (frost) | `#4C566A` (polar night) |
| Solarized | `#DC322F` (red) | `#B58900` (yellow) | `#268BD2` (blue) | `#586E75` (base01) |

**Design rationale:** P1 uses the theme's "danger/alert" color (same family as HolidayFg/PendingFg). P2 uses "warning" (same family as IndicatorFg). P3 uses "info" (same family as AccentFg/BorderFocused). P4 uses "muted" (same as MutedFg). This creates a natural urgency gradient using colors already in the theme's palette.

### Todolist Styles Extension

Add 4 new styles:

```go
type Styles struct {
    // ... existing 14 fields ...
    PriorityP1 lipgloss.Style
    PriorityP2 lipgloss.Style
    PriorityP3 lipgloss.Style
    PriorityP4 lipgloss.Style
}

// Helper method on Model:
func (m Model) priorityStyle(p int) lipgloss.Style {
    switch p {
    case 1: return m.styles.PriorityP1
    case 2: return m.styles.PriorityP2
    case 3: return m.styles.PriorityP3
    case 4: return m.styles.PriorityP4
    default: return lipgloss.NewStyle()
    }
}
```

### Sort Order Consideration

**Should priority affect sort order within sections?**

Two approaches:

**A. Visual priority only (no sort change):** Priority badges are displayed but todos remain sorted by sort_order. User manually reorders via J/K.

**B. Priority-then-sort ordering:** Within each section, sort by priority (1 first, then 2, 3, 4, 0 last), then by sort_order within same priority.

**Recommendation: Approach A for now (visual only).** Manual reordering via J/K is an established UX pattern in this app. Automatic priority-based sorting would fight against manual reordering -- a user who carefully arranged their todos would see them jump around when priorities change. If auto-sort is desired later, it can be added as a setting ("Sort by: manual / priority").

**Search integration:** The search overlay renders todos with date. It should also show priority badges. This is a display-only change in `search/model.go`'s `View()`.

---

## Feature 2: Natural Language Date Input

### Library Selection

**Use `github.com/tj/go-naturaldate` because:**

1. **Pure Go, zero transitive dependencies.** The go.mod stays clean. Compare: `olebedev/when` has rule-based architecture requiring explicit language registration; `sho0pi/naturaltime` wraps a JavaScript engine (goja) -- adding a JS runtime to a Go TUI is absurd.

2. **API is minimal and correct.** One function: `Parse(s string, ref time.Time, options ...Option) (time.Time, error)`. Returns a standard `time.Time`. No custom result structs to unwrap.

3. **Handles the expressions that matter for a todo app:**
   - "today", "tomorrow", "yesterday"
   - "next monday", "next friday"
   - "next week", "next month"
   - "december 25", "jan 15"
   - "5 days from now", "in 2 weeks"
   - "last sunday" (with configurable Past/Future direction)

4. **313 stars, MIT license, stable (15 commits, no churn).** It does one thing well.

5. **Direction option maps to our UX needs.** `WithDirection(naturaldate.Future)` makes "monday" mean "next monday" which is the right default for a todo app (you are scheduling future work).

**Why not olebedev/when (1.5k stars):** More popular but heavier. Requires registering language-specific rule sets (`w.Add(en.All...)`). Returns a `Result` struct with `Index`, `Text`, `Source`, `Time` -- more than we need. The pluggable rule system is designed for extracting dates from prose, which is overkill for a single-purpose date input field.

**Why not araddon/dateparse:** Parses date *formats* (e.g., "2006-01-02", "Jan 2 2006"), not natural language expressions. "tomorrow" and "next monday" would not parse. Wrong tool for this job.

### Integration Architecture

Natural language date input replaces the existing segmented date input in the edit form. The segmented input (dd/mm/yyyy fields) remains as the structured fallback. The NL input is the primary input method.

**Design: Single text input with dual parsing.**

The date field becomes a single textinput that accepts either:
1. Natural language: "tomorrow", "next friday", "jan 15"
2. Formatted date: "2026-02-15", "15.02.2026", "02/15/2026" (per user's date format setting)

Parsing order:
1. Try exact date format parse (user's configured format)
2. Try natural language parse via go-naturaldate
3. If both fail, show error hint

```go
func (m Model) parseDate(input string) (isoDate string, precision string, err error) {
    input = strings.TrimSpace(input)
    if input == "" {
        return "", "", nil  // floating todo
    }

    // 1. Try exact format parse
    if t, err := time.Parse(m.dateLayout, input); err == nil {
        return t.Format("2006-01-02"), "day", nil
    }

    // 2. Try natural language parse
    ref := time.Now()
    t, err := naturaldate.Parse(input, ref, naturaldate.WithDirection(naturaldate.Future))
    if err == nil {
        // Determine precision from the input
        precision := derivePrecisionFromNL(input, t)
        return t.Format("2006-01-02"), precision, nil
    }

    return "", "", fmt.Errorf("unrecognized date: %q", input)
}
```

### Precision Derivation from NL Input

The existing segmented input derives precision from which segments are filled (year only = year precision, year+month = month precision, all three = day precision). With NL input, we need a different approach.

**Strategy: All NL dates default to day precision.**

Rationale: Natural language expressions like "tomorrow", "next friday", "jan 15" all resolve to specific days. Month-level ("this month") and year-level ("this year") expressions are unusual enough that they do not warrant special parsing. Users who want month/year precision can use the structured input mode.

**Exception handling:**
- "next month" -> could mean month precision. But `go-naturaldate` resolves it to a specific day (same day next month). Day precision is correct for this interpretation.
- "january" -> resolves to January 1st of the appropriate year. Day precision is technically wrong (user may mean "sometime in January"), but there is no reliable way to distinguish "January the month" from "January 1st" in NL parsing. Default to day precision.

**If month/year precision is needed:** The user can still enter a partial date in the structured format. The parser tries exact format first, so "2026-02" or "2026" could be caught before NL parsing. But this is a minor edge case -- the primary UX improvement is "tomorrow" and "next monday" working.

### Replacing Segmented Input with NL Input

**The big UX question:** Do we replace the 3-segment date input entirely, or add NL as an alternative mode?

**Recommendation: Replace with a single textinput.** The segmented input was a reasonable approach before NL parsing existed. But it requires 6 Tab presses to fill (day->month->year, 2 chars each), format-aware ordering logic, backspace-to-previous-segment, and auto-advance. A single textinput where the user types "tomorrow" or "2026-02-15" is strictly better UX.

**What we lose:** The segmented input had implicit precision derivation (leave day blank = month precision). The NL input cannot do this as elegantly. However, precision can be set explicitly via the priority-style approach: default to day, and add a "precision" field or allow special syntax ("~feb 2026" for month, "~2026" for year).

**Practical approach:** Keep it simple. The NL textinput handles day-precision dates. For month/year precision, accept patterns like "feb 2026" (month precision) and "2026" (year precision) by checking the input format before NL parsing:

```go
func (m Model) parseDate(input string) (string, string, error) {
    input = strings.TrimSpace(input)
    if input == "" {
        return "", "", nil // floating
    }

    // Check for year-only (4 digits)
    if matched, _ := regexp.MatchString(`^\d{4}$`, input); matched {
        return input + "-01-01", "year", nil
    }

    // Check for month+year patterns: "feb 2026", "2026-02", "02.2026"
    if isoDate, ok := parseMonthYear(input); ok {
        return isoDate, "month", nil
    }

    // Try exact date format (user's configured format)
    if t, err := time.Parse(m.dateLayout, input); err == nil {
        return t.Format("2006-01-02"), "day", nil
    }

    // Try ISO format explicitly
    if t, err := time.Parse("2006-01-02", input); err == nil {
        return t.Format("2006-01-02"), "day", nil
    }

    // Try natural language
    ref := time.Now()
    t, err := naturaldate.Parse(input, ref, naturaldate.WithDirection(naturaldate.Future))
    if err == nil {
        return t.Format("2006-01-02"), "day", nil
    }

    return "", "", fmt.Errorf("unrecognized date")
}
```

### Form Field Changes

**Before (current):**
```
Title:    [_______________]
Date:     [yyyy] - [mm] - [dd]     (3 segmented inputs)
Body:     [textarea]
Template: [picker trigger]
```

**After:**
```
Title:    [_______________]
Date:     [_______________________________]     (single textinput)
          (try: tomorrow, next fri, jan 15, or 2026-02-15)
Priority: [none]     (press 1-4 to set)
Body:     [textarea]
Template: [picker trigger]
```

The hint text below the date field guides users to the NL capability. It replaces the existing "(leave day blank for month todo...)" hint.

### Model Changes for NL Date

**Remove from todolist.Model:**
```go
// REMOVE these fields:
dateSegDay   textinput.Model
dateSegMonth textinput.Model
dateSegYear  textinput.Model
dateSegFocus int
dateSegOrder [3]int
dateFormat   string
```

**Add to todolist.Model:**
```go
// ADD:
dateInput textinput.Model  // single NL date input
```

**Remove helper methods:**
```go
// REMOVE: renderDateSegments, deriveDateFromSegments, updateDateSegment,
//         dateSegmentOrder, dateSegmentByPos, dateSegSeparator,
//         dateSegPlaceholderByPos, focusDateSegment, blurAllDateSegments,
//         clearAllDateSegments, dateSegCharLimit
```

**Add helper methods:**
```go
// ADD: parseDate (see above), parseMonthYear
```

This is a net reduction in code complexity. The segmented input system is ~200 lines of logic (dateSegmentOrder, auto-advance, backspace navigation, format-aware ordering). The NL input replaces it with ~50 lines of parse logic.

### New Package: `internal/nldate/`

**Purpose:** Isolate NL date parsing from the todolist model. Pure function, easily testable.

```go
package nldate

import (
    "time"
    "github.com/tj/go-naturaldate"
)

// Parse attempts to parse a date string using multiple strategies:
// 1. Year-only (4 digits) -> year precision
// 2. Month+year patterns -> month precision
// 3. Exact date format (user's layout) -> day precision
// 4. ISO date format -> day precision
// 5. Natural language -> day precision
//
// Returns (isoDate, precision, error). Empty input returns ("", "", nil) for floating.
func Parse(input string, ref time.Time, userLayout string) (string, string, error) {
    // ... implementation
}
```

**Why a separate package:** The parsing logic is pure (no TUI state) and needs thorough unit testing with many input variations. Putting it in todolist would make it harder to test independently.

### Test Strategy for NL Date Parsing

```go
func TestParse(t *testing.T) {
    ref := time.Date(2026, 2, 12, 10, 0, 0, 0, time.Local)

    tests := []struct {
        input     string
        wantDate  string
        wantPrec  string
        wantErr   bool
    }{
        // Empty -> floating
        {"", "", "", false},
        {"  ", "", "", false},

        // Year-only -> year precision
        {"2026", "2026-01-01", "year", false},
        {"2027", "2027-01-01", "year", false},

        // Month+year -> month precision
        {"feb 2026", "2026-02-01", "month", false},
        {"march 2027", "2027-03-01", "month", false},

        // Exact ISO date -> day precision
        {"2026-02-15", "2026-02-15", "day", false},
        {"2026-12-25", "2026-12-25", "day", false},

        // Natural language -> day precision
        {"today", "2026-02-12", "day", false},
        {"tomorrow", "2026-02-13", "day", false},
        {"next friday", "2026-02-13", "day", false},  // Feb 12 2026 is Thursday
        {"next monday", "2026-02-16", "day", false},

        // Invalid
        {"not a date at all", "", "", true},
    }
    // ...
}
```

---

## Component Analysis: New vs Modified

### New Components

| Component | Type | Purpose |
|-----------|------|---------|
| `internal/nldate/` | New package | NL date parsing with multi-strategy fallback |
| `nldate/nldate.go` | New file | `Parse()` function |
| `nldate/nldate_test.go` | New file | Comprehensive parse tests |

### Modified Components

| Component | Change | Impact |
|-----------|--------|--------|
| `store/todo.go` | Add `Priority int` field, helper methods | Low -- additive |
| `store/iface.go` | Extend `Add()` and `Update()` signatures with priority param | Medium -- interface change, affects all callers |
| `store/sqlite.go` | Migration v7, extend Add/Update/scanTodo/todoColumns, priority in INSERT/UPDATE | Medium -- mechanical changes |
| `theme/theme.go` | Add 4 priority color fields to Theme, values for all 4 themes | Low -- additive |
| `todolist/styles.go` | Add 4 PriorityP1-P4 styles | Low -- additive |
| `todolist/model.go` | Replace segmented date input with single NL textinput, add priority field, update edit form cycle, update renderTodo, update save methods | HIGH -- most changed file |
| `todolist/keys.go` | No changes needed (Tab/Enter/Esc handle the new fields) | None |
| `search/model.go` | Show priority badge in search results | Low -- display only |
| `app/model.go` | No changes needed (todo display is in todolist) | None |
| `go.mod` | Add `github.com/tj/go-naturaldate` dependency | Low |

### Removed Code

| Code | Reason |
|------|--------|
| `dateSegDay`, `dateSegMonth`, `dateSegYear` fields | Replaced by single `dateInput` |
| `dateSegFocus`, `dateSegOrder`, `dateFormat` fields | No longer needed |
| `renderDateSegments()` | Replaced by `dateInput.View()` |
| `deriveDateFromSegments()` | Replaced by `nldate.Parse()` |
| `updateDateSegment()` | Standard textinput handles everything |
| `dateSegmentOrder()`, `dateSegmentByPos()`, etc. | All segmented input helpers removed |
| ~200 lines of segmented input logic | Net simplification |

---

## Data Flow Changes

### Add Todo Flow (Before)

```
User types title -> Tab -> types dd -> auto-advance -> types mm -> auto-advance -> types yyyy -> Tab -> types body -> Enter
  |
  deriveDateFromSegments() -> isoDate, precision
  |
  store.Add(text, isoDate, precision)
```

### Add Todo Flow (After)

```
User types title -> Tab -> types "tomorrow" or "2026-02-15" -> Tab -> types 1-4 for priority -> Tab -> types body -> Enter
  |
  nldate.Parse(dateInput.Value(), time.Now(), userLayout) -> isoDate, precision
  |
  store.Add(text, isoDate, precision, priority)
```

### Edit Todo Flow Changes

When entering edit mode, populate the date field with a human-readable date string instead of splitting into segments:

```go
// Before (populating segments):
parts := strings.SplitN(fresh.Date, "-", 3)
if len(parts) == 3 {
    m.dateSegYear.SetValue(parts[0])
    m.dateSegMonth.SetValue(parts[1])
    m.dateSegDay.SetValue(parts[2])
}

// After (populating NL input):
if fresh.Date != "" {
    m.dateInput.SetValue(config.FormatDate(fresh.Date, m.dateLayout))
}
m.editPriority = fresh.Priority
```

### SetDateFormat Impact

The `SetDateFormat(format, layout, placeholder)` method currently updates segment ordering. After the change, it only needs to update the `dateLayout` for the parser and the display format:

```go
func (m *Model) SetDateFormat(format, layout, placeholder string) {
    m.dateLayout = layout
    // dateSegOrder no longer needed
}
```

---

## Integration Points Detailed

### 1. Store Interface Change (Priority Parameter)

This is the highest-impact change. Every caller of `Add()` and `Update()` must pass a priority.

**Callers of Add():**
- `todolist.Model.saveAdd()` -> pass `m.editPriority`
- `store.SQLiteStore.AddScheduledTodo()` -> pass `0` (scheduled todos get no priority by default)

**Callers of Update():**
- `todolist.Model.saveEdit()` -> pass `m.editPriority`

### 2. todoColumns / scanTodo (Column Addition)

Add `priority` to the end of the column list. This is the same pattern used when `date_precision` was added in v1.9.

### 3. Theme Propagation (4 New Colors)

Every theme constructor (`Dark()`, `Light()`, `Nord()`, `Solarized()`) must return the 4 new color fields. `NewStyles()` in todolist must create 4 new style objects.

### 4. Edit Form Field Count

`editField` range changes from 0-3 to 0-4. The `SwitchField` handler in both `updateInputMode` and `updateEditMode` must be updated to include the new priority field at position 2.

### 5. NL Date Dependency

`go.mod` gains one new direct dependency. Install: `go get github.com/tj/go-naturaldate@latest`

### 6. Search Results Display

`search/model.go` View() needs to show priority badges. This means the search model needs access to priority styles, which come from the theme. The search model already has a Styles struct and SetTheme() -- add priority styles there.

---

## Suggested Build Order

Dependencies flow: Schema -> Store methods -> NL parse package -> UI integration. Priority and NL date are independent features that share the edit form. Build order minimizes merge conflicts by doing schema first, then the two features in parallel or sequence.

### Phase 1: Priority Data Layer

**What:** Migration v7 (add priority column), extend Todo struct, extend Add/Update/scan, add Priority helper methods.

**Why first:** Schema migration must exist before any UI can read/write priority. This is backend-only, no UI changes, easily testable in isolation.

**New/Modified:**
- MOD: `store/todo.go` (Priority field, helper methods)
- MOD: `store/iface.go` (Add/Update signatures)
- MOD: `store/sqlite.go` (migration v7, todoColumns, scanTodo, Add, Update implementations)
- MOD: `store/sqlite_test.go` (priority roundtrip tests)

**Risk:** Low. Follows the established migration pattern (v6 added date_precision the same way).

### Phase 2: Priority UI + Theme

**What:** Add 4 priority colors to Theme, 4 styles to Styles, priority field in edit form, priority badge in renderTodo, priority in search results.

**Why second:** Depends on Phase 1 (store must read/write priority). Self-contained UI feature.

**New/Modified:**
- MOD: `theme/theme.go` (4 new color fields in Theme struct + all 4 theme constructors)
- MOD: `todolist/styles.go` (4 new style fields in Styles struct + NewStyles)
- MOD: `todolist/model.go` (editPriority field, priority in edit form cycle, priority input handling, priority in renderTodo, priority in saveAdd/saveEdit)
- MOD: `search/model.go` + `search/styles.go` (priority badge in search results)

**Risk:** Low-Medium. The edit form field cycling logic is the most complex change but follows the existing pattern.

### Phase 3: Natural Language Date Input

**What:** Create nldate package, replace segmented date input with single NL textinput, update form field handling, remove segmented input code.

**Why third:** Independent of priority (could be parallel), but putting it last means the priority form changes are stable before refactoring the date input. The date input replacement is a larger refactor that touches more of the edit form logic.

**New/Modified:**
- NEW: `internal/nldate/nldate.go` (Parse function)
- NEW: `internal/nldate/nldate_test.go` (comprehensive tests)
- MOD: `go.mod` (add go-naturaldate dependency)
- MOD: `todolist/model.go` (replace dateSegDay/Month/Year with dateInput, replace deriveDateFromSegments with nldate.Parse, update form rendering, remove all segmented input helpers)

**Risk:** Medium. The segmented input replacement is a significant refactor that touches many methods. The NL parsing library is third-party and needs edge case testing. The edit form hint text needs care to guide users.

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Storing Priority as a String

**What:** Using `TEXT` column with values "p1", "p2", "p3", "p4", "none".
**Why bad:** Can not sort by priority with `ORDER BY`. Wastes storage (5 bytes vs 1). Requires mapping on every read.
**Instead:** `INTEGER NOT NULL DEFAULT 0`. Zero = no priority, 1-4 = priority levels.

### Anti-Pattern 2: Adding NL Date Parsing Inline in todolist.Model

**What:** Putting the `naturaldate.Parse()` call directly in the save methods with all the format detection logic.
**Why bad:** Untestable without spinning up the full TUI model. Many edge cases in date parsing need isolated unit tests.
**Instead:** `internal/nldate/` package with a pure `Parse()` function. todolist calls `nldate.Parse()`.

### Anti-Pattern 3: Keeping Segmented Input Alongside NL Input

**What:** Offering both segmented (dd/mm/yyyy) and NL input as separate modes or fields.
**Why bad:** Two ways to do the same thing creates confusion ("which should I use?"). The segmented input is strictly inferior once NL parsing exists. Maintaining both doubles the code surface.
**Instead:** Single textinput that accepts both NL and formatted dates. The parser handles format detection transparently.

### Anti-Pattern 4: Auto-Sorting by Priority

**What:** Automatically reordering todos within a section by priority level (P1 first, P4 last).
**Why bad:** Conflicts with manual J/K reordering, which is the established pattern. A user who carefully arranged todos by context (e.g., "morning tasks first") would see their order destroyed when they set a priority.
**Instead:** Priority is visual-only (colored badge). Sort order remains manual (sort_order column). Auto-sort by priority can be a settings option in a future milestone if users request it.

### Anti-Pattern 5: Complex Priority Cycling in Normal Mode

**What:** Adding a keybinding in normal mode to cycle priority on the selected todo (e.g., pressing "p" cycles through P1->P2->P3->P4->none).
**Why bad:** Conflicts with the existing "p" key (preview). Adds accidental priority changes from a single keypress. Priority is not something you change as frequently as toggling done/not-done.
**Instead:** Priority is set in the edit form only. Press "e" to edit, Tab to priority field, type 1-4. Intentional, not accidental.

---

## Scalability Considerations

| Concern | At 100 todos | At 1K todos | At 10K todos |
|---------|-------------|-------------|-------------- |
| Priority column impact | Negligible | Negligible | Add index if filtering by priority in SQL |
| NL date parsing | < 1ms | N/A (one parse per input, not per todo) | N/A |
| Priority badge rendering | Negligible | Minimal overhead (1 style.Render per todo) | May need viewport (already an issue without priority) |
| Migration v7 | Instant | < 100ms | < 1s |

---

## Sources

- Codebase analysis: `store/todo.go` (Todo struct, 10 fields, helper methods) -- HIGH confidence
- Codebase analysis: `store/iface.go` (TodoStore interface, 27 methods, Add/Update signatures) -- HIGH confidence
- Codebase analysis: `store/sqlite.go` (migration pattern v1-v6, todoColumns, scanTodo, Add/Update implementations) -- HIGH confidence
- Codebase analysis: `todolist/model.go` (edit form fields 0-3, segmented date input system, renderTodo, save methods) -- HIGH confidence
- Codebase analysis: `theme/theme.go` (16 color roles, 4 theme constructors) -- HIGH confidence
- Codebase analysis: `todolist/styles.go` (14 styles, NewStyles pattern) -- HIGH confidence
- Codebase analysis: `search/model.go` (search result rendering, SetTheme pattern) -- HIGH confidence
- Library: [tj/go-naturaldate](https://github.com/tj/go-naturaldate) -- 313 stars, MIT, Parse() API, WithDirection option -- MEDIUM confidence (not tested locally)
- Library: [olebedev/when](https://github.com/olebedev/when) -- 1.5k stars, evaluated and rejected for complexity -- MEDIUM confidence
- Library: [sho0pi/naturaltime](https://github.com/sho0pi/naturaltime) -- wraps JS runtime, rejected -- MEDIUM confidence
- Library: [araddon/dateparse](https://github.com/araddon/dateparse) -- format parsing not NL, rejected -- MEDIUM confidence
- [go-naturaldate pkg.go.dev](https://pkg.go.dev/github.com/tj/go-naturaldate) -- API documentation verified -- MEDIUM confidence
