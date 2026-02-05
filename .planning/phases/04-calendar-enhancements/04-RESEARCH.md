# Phase 4: Calendar Enhancements - Research

**Researched:** 2026-02-05
**Domain:** TUI calendar rendering, config parsing, grid alignment
**Confidence:** HIGH

## Summary

Phase 4 adds two features to the existing calendar: (1) bracket indicators `[N]` on dates with incomplete todos, and (2) a configurable first day of week via `first_day_of_week` in config.toml. Both features modify the pure `RenderGrid` function in `internal/calendar/grid.go` and the `Config` struct in `internal/config/config.go`.

The primary challenge is grid alignment (INDI-03). Currently, each date cell is 2 characters (`%2d`), with 1-character spaces between columns, totaling a 20-character-wide grid. Adding brackets around dates changes cells to 4 characters (`[ 5]`, `[15]`), requiring ALL cells to become 4 characters wide to maintain alignment. This expands the grid to 34 characters (7 x 4 + 6 x 1). The `calendarInnerWidth` constant in `app/model.go` (currently 24) must be updated accordingly.

The config change from `monday_start` (bool) to `first_day_of_week` (string) is a backward-compatibility consideration. The current config field is `monday_start bool` with TOML tag `monday_start`. The requirements specify `first_day_of_week = "monday"` or `"sunday"`. This is a field rename and type change.

**Primary recommendation:** Widen all calendar cells to a uniform 4-character width, use `[ N]`/`[NN]` for indicated dates and ` N `/`NN ` for non-indicated dates, and add a `IncompleteTodosPerDay` store method that the calendar model consumes.

## Standard Stack

No new libraries are needed. Phase 4 uses only the existing stack.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `time` | 1.25.6 | `time.Weekday` for day-of-week math, `time.Date` for month boundaries | Already used in grid.go |
| Go stdlib `fmt` | 1.25.6 | `fmt.Sprintf` for cell formatting with `%2d`, `%4s` patterns | Already used in grid.go |
| BurntSushi/toml | v1.6.0 | Config parsing with struct tags | Already used in config.go |
| Lipgloss | v1.1.0 | Style rendering for bracket indicator style | Already used in calendar/styles.go |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib `strings` | 1.25.6 | `strings.Builder` for grid assembly | Already used in grid.go |
| Go stdlib `strconv` | 1.25.6 | Number formatting if needed | Prefer `fmt.Sprintf` which is already in use |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Uniform 4-char cells | Variable-width cells with lipgloss padding | Variable width breaks column alignment with ANSI escape codes; uniform width is simpler and guaranteed correct |
| `first_day_of_week` string | Keep `monday_start` bool | Requirements explicitly specify string config `first_day_of_week = "monday"` or `"sunday"` |
| Store method for counts | Inline counting in calendar View() | Violates separation of concerns; store should own data querying |

**Installation:** No new dependencies needed.

## Architecture Patterns

### Current Architecture (Relevant Files)

```
internal/
  calendar/
    grid.go         # RenderGrid() pure function - PRIMARY CHANGE TARGET
    model.go        # Calendar Model - needs store access for indicator data
    styles.go       # Lipgloss styles - add indicator style
    keys.go         # No changes needed
  config/
    config.go       # Config struct - change field type
    paths.go        # No changes needed
  store/
    store.go        # Store methods - add IncompleteTodosPerDay()
    todo.go         # Todo struct - no changes needed
  app/
    model.go        # Root model - update calendarInnerWidth, pass store to calendar
    styles.go       # No changes needed
  todolist/
    model.go        # No changes needed
```

### Pattern 1: Pure Rendering with Data Injection
**What:** RenderGrid stays pure -- callers pass in all data including todo indicator counts
**When to use:** Always for rendering functions in this codebase
**Example:**
```go
// BEFORE: RenderGrid(year, month, today, holidays, mondayStart)
// AFTER:  RenderGrid(year, month, today, holidays, mondayStart, indicators)

// indicators is map[int]int: day number -> count of incomplete todos
// Zero or missing means no indicator for that day
func RenderGrid(year int, month time.Month, today int, holidays map[int]bool,
    mondayStart bool, indicators map[int]int) string {
    // ...
}
```

### Pattern 2: Store Query Method
**What:** Add a method to Store that returns incomplete todo counts grouped by day
**When to use:** When the calendar needs per-day incomplete counts
**Example:**
```go
// IncompleteTodosPerDay returns a map of day-number to count of incomplete todos
// for the given month. Days with zero incomplete todos are omitted.
func (s *Store) IncompleteTodosPerDay(year int, month time.Month) map[int]int {
    counts := make(map[int]int)
    for _, t := range s.data.Todos {
        if t.Done || !t.InMonth(year, month) {
            continue
        }
        d, err := time.Parse(dateFormat, t.Date)
        if err != nil {
            continue
        }
        counts[d.Day()]++
    }
    return counts
}
```

### Pattern 3: Config Field Migration
**What:** Change `monday_start` bool to `first_day_of_week` string with backward compatibility
**When to use:** When config format changes
**Example:**
```go
type Config struct {
    Country       string `toml:"country"`
    FirstDayOfWeek string `toml:"first_day_of_week"`
}

func DefaultConfig() Config {
    return Config{
        Country:        "us",
        FirstDayOfWeek: "sunday",
    }
}

// MondayStart is a convenience method for use by calendar code
func (c Config) MondayStart() bool {
    return c.FirstDayOfWeek == "monday"
}
```

### Pattern 4: Calendar Model Gets Store Access
**What:** Calendar model needs access to the store to query indicator data
**When to use:** When the calendar needs todo data for rendering
**Example:**
```go
// Calendar model gains a store reference
type Model struct {
    // ... existing fields ...
    store       *store.Store
    indicators  map[int]int // cached per-day incomplete counts
}

// Constructor updated
func New(provider *holidays.Provider, mondayStart bool, s *store.Store) Model {
    // ...
}

// View passes indicators to RenderGrid
func (m Model) View() string {
    return RenderGrid(m.year, m.month, todayDay, m.holidays, m.mondayStart, m.indicators)
}
```

### Anti-Patterns to Avoid
- **Querying store in RenderGrid:** RenderGrid must remain a pure function. Pass indicators as a parameter.
- **Computing indicators only in View():** Indicators should be cached/updated when the month changes (same pattern as holidays), not recomputed every render.
- **Using lipgloss Width() for alignment:** Lipgloss has known Unicode width issues. Use plain `fmt.Sprintf` with fixed-width format strings for cell content, then apply styles afterward.
- **Variable-width cells:** Do not make indicator cells wider than non-indicator cells. All cells must be the same width.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Day-of-week offset calculation | Custom weekday math | `(int(firstDay) + 6) % 7` pattern already in grid.go | The modular arithmetic is already correct and tested |
| Days in month | Manual day counting | `time.Date(year, month+1, 0, ...).Day()` | Already used in grid.go, handles leap years correctly |
| Config default values | Checking for empty strings everywhere | `DefaultConfig()` function with sensible defaults | Pattern already established |
| TOML parsing | Manual file parsing | `toml.DecodeFile` with struct tags | BurntSushi/toml already in use, handles missing fields gracefully |

**Key insight:** The existing codebase patterns are well-established. This phase extends existing patterns rather than introducing new ones.

## Common Pitfalls

### Pitfall 1: ANSI Escape Codes Break Column Alignment
**What goes wrong:** When Lipgloss styles are applied to a cell BEFORE it is placed in the grid, the ANSI escape codes add invisible characters. If you use string length (instead of visual width) for padding, columns misalign.
**Why it happens:** `lipgloss.NewStyle().Render("5")` produces something like `\x1b[0m5\x1b[0m` which is ~10+ bytes but renders as 1 visible character.
**How to avoid:** Format the cell to its final visual width FIRST using `fmt.Sprintf`, THEN apply the Lipgloss style. This is exactly what the current code does: `cell := fmt.Sprintf("%2d", day)` then `cell = todayStyle.Render(cell)`. Maintain this pattern for the new 4-char cells.
**Warning signs:** Columns visibly shift when styled cells appear next to unstyled cells.

### Pitfall 2: Grid Width Constant Not Updated
**What goes wrong:** The `calendarInnerWidth` in app/model.go is hardcoded to 24. If grid cells change width without updating this constant, the calendar panel will clip or have excess space.
**Why it happens:** The width change propagates from grid.go through to the layout in app/model.go.
**How to avoid:** Calculate the new grid width: 7 cells x 4 chars + 6 spaces = 34 chars. Add padding (currently 2 from `Padding(0, 1)`) plus 2 for border = 38. Update `calendarInnerWidth` accordingly.
**Warning signs:** Calendar text wraps or is clipped within its panel; right panel is too narrow.

### Pitfall 3: Indicator Counts Not Refreshed on Toggle
**What goes wrong:** User toggles a todo complete/incomplete in the todo panel, but the calendar indicators do not update because they were cached when the month was loaded.
**Why it happens:** Indicators are computed from store data, and the calendar model does not know when the store changes.
**How to avoid:** Recompute indicators whenever the app routes back to the calendar pane or when the todo list performs a mutation. The simplest approach: recompute indicators in the calendar's View() method by calling the store method each time, OR refresh indicators after every Update cycle in the root model.
**Warning signs:** Toggling a todo does not change the bracket indicator on the calendar until navigating away and back.

### Pitfall 4: Off-by-One in First Day of Week
**What goes wrong:** The header says "Mo Tu We..." but the days are shifted by one column, or vice versa.
**Why it happens:** The `startCol` calculation uses modular arithmetic that must match the header. If the header order changes but the startCol math does not (or vice versa), dates land in wrong columns.
**How to avoid:** The current code already handles both cases correctly with `(int(firstDay) + 6) % 7` for Monday-start. The `first_day_of_week` config change is a data flow change, not a math change. Ensure the bool `mondayStart` derived from the config reaches both the header rendering and the startCol calculation in the same function.
**Warning signs:** First day of month appears under the wrong weekday header.

### Pitfall 5: Backward-Incompatible Config Change
**What goes wrong:** Users who have `monday_start = true` in their config.toml get a parse error or silent behavior change after the update.
**Why it happens:** Renaming the TOML field from `monday_start` to `first_day_of_week` breaks existing configs.
**How to avoid:** Either (a) keep `monday_start` as a deprecated fallback and check for it, or (b) since this is a personal-use app with a small user base (the developer), just document the breaking change and update the config. Option (b) is simplest and appropriate here.
**Warning signs:** App errors on startup for users with old config format.

### Pitfall 6: Double-Digit Dates With Brackets Overflow Cell Width
**What goes wrong:** `[15]` is 4 characters but the cell is sized for 3 or 2, breaking the grid.
**Why it happens:** Not accounting for the maximum cell width including brackets and two-digit dates.
**How to avoid:** All cells must be 4 characters wide. Single-digit non-indicated: ` 5  ` (with trailing space). Single-digit indicated: `[ 5]`. Double-digit non-indicated: ` 15` (with leading space). Double-digit indicated: `[15]`. Wait -- this does not work evenly. Better approach: 4-char cells where non-indicated dates are right-aligned with trailing space: `  5 ` and ` 15 `, while indicated dates use brackets: `[ 5]` and `[15]`.
**Warning signs:** Grid columns shift when reaching day 10+.

## Code Examples

### Cell Formatting for 4-Character Uniform Width

```go
// Source: derived from current grid.go pattern

// Format a day cell with optional indicator brackets.
// All cells are exactly 4 visible characters wide.
func formatCell(day int, indicatorCount int) string {
    if indicatorCount > 0 {
        // Bracketed: [5] or [15] -- always 4 chars
        return fmt.Sprintf("[%2d]", day)
    }
    // Non-bracketed: " 5 " or " 15" -- always 4 chars (centered-ish)
    return fmt.Sprintf(" %2d ", day)
}
```

### Updated RenderGrid Signature

```go
// Source: extends current internal/calendar/grid.go

func RenderGrid(year int, month time.Month, today int, holidays map[int]bool,
    mondayStart bool, indicators map[int]int) string {
    var b strings.Builder

    // Title line: month and year, centered in new grid width
    gridWidth := 34 // 7 * 4 + 6 * 1
    title := fmt.Sprintf("%s %d", month.String(), year)
    pad := (gridWidth - len(title)) / 2
    if pad < 0 {
        pad = 0
    }
    b.WriteString(strings.Repeat(" ", pad))
    b.WriteString(headerStyle.Render(title))
    b.WriteString("\n")

    // Weekday header (4 chars per day with 1-char separator)
    if mondayStart {
        b.WriteString(weekdayHdrStyle.Render(" Mo  Tu  We  Th  Fr  Sa  Su"))
    } else {
        b.WriteString(weekdayHdrStyle.Render(" Su  Mo  Tu  We  Th  Fr  Sa"))
    }
    b.WriteString("\n")

    // ... rest of grid logic with 4-char cells and 1-char separator
}
```

Note: The weekday header strings above are 34 characters (matching 7 x 4 + 6 x 1). Each day label is 4 chars: ` Mo `, ` Tu `, etc. separated by no extra spaces since the labels already include spacing.

Actually, let me reconsider the header. With 4-char cells and 1-char separators:
- Cell widths: `[ 5]` = 4 chars, separator = 1 char
- 7 cells + 6 separators = 34 chars
- Header: `Mo   Tu   We   Th   Fr   Sa   Su` -- each label should center in 4 chars

A cleaner header approach: ` Mo  Tu  We  Th  Fr  Sa  Su ` with 4-char per label.

### Weekday Header Alignment

```go
// Each header label padded to 4 chars to match cell width
var mondayHeaders = [7]string{" Mo ", " Tu ", " We ", " Th ", " Fr ", " Sa ", " Su "}
var sundayHeaders = [7]string{" Su ", " Mo ", " Tu ", " We ", " Th ", " Fr ", " Sa "}

func weekdayHeader(mondayStart bool) string {
    headers := sundayHeaders
    if mondayStart {
        headers = mondayHeaders
    }
    return strings.Join(headers[:], " ")
}
```

### Config With String first_day_of_week

```go
// Source: extends current internal/config/config.go

type Config struct {
    Country        string `toml:"country"`
    FirstDayOfWeek string `toml:"first_day_of_week"`
}

func DefaultConfig() Config {
    return Config{
        Country:        "us",
        FirstDayOfWeek: "sunday",
    }
}

// MondayStart returns true if the configured first day of week is Monday.
func (c Config) MondayStart() bool {
    return c.FirstDayOfWeek == "monday"
}
```

### Store Method for Incomplete Counts

```go
// Source: extends current internal/store/store.go

// IncompleteTodosPerDay returns a map from day-of-month to count of
// incomplete (not done) todos for the specified year and month.
func (s *Store) IncompleteTodosPerDay(year int, month time.Month) map[int]int {
    counts := make(map[int]int)
    for _, t := range s.data.Todos {
        if t.Done || !t.InMonth(year, month) {
            continue
        }
        d, err := time.Parse(dateFormat, t.Date)
        if err != nil {
            continue
        }
        counts[d.Day()]++
    }
    return counts
}
```

### Calendar Model With Store Access

```go
// Source: extends current internal/calendar/model.go

type Model struct {
    focused     bool
    width       int
    height      int
    year        int
    month       time.Month
    today       time.Time
    holidays    map[int]bool
    indicators  map[int]int        // day -> incomplete todo count
    provider    *holidays.Provider
    store       *store.Store       // NEW: store reference
    keys        KeyMap
    mondayStart bool
}

func New(provider *holidays.Provider, mondayStart bool, s *store.Store) Model {
    now := time.Now()
    y, m, _ := now.Date()

    return Model{
        year:        y,
        month:       m,
        today:       now,
        holidays:    provider.HolidaysInMonth(y, m),
        indicators:  s.IncompleteTodosPerDay(y, m),
        provider:    provider,
        store:       s,
        keys:        DefaultKeyMap(),
        mondayStart: mondayStart,
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `monday_start` bool config field | `first_day_of_week` string field | Phase 4 | Config is more readable, extensible |
| 2-char date cells (20-char grid) | 4-char date cells (34-char grid) | Phase 4 | Wider calendar panel, indicators fit cleanly |
| Calendar has no store access | Calendar reads store for indicators | Phase 4 | Calendar model gains `*store.Store` dependency |

**Deprecated/outdated:**
- `monday_start` bool config field: replaced by `first_day_of_week` string
- 20-char grid width assumption: replaced by 34-char grid width
- `calendarInnerWidth = 24`: must be recalculated for new grid width

## Open Questions

1. **Indicator refresh strategy**
   - What we know: Indicators must update when todos are toggled/added/deleted
   - What's unclear: Whether to recompute on every View() call (simple but possibly wasteful) or cache and invalidate (complex but efficient)
   - Recommendation: Recompute in View() for now. The store iterates ~tens of todos, which is negligible. Premature optimization is not needed for a personal app. Alternatively, refresh indicators in the root model's Update() after routing to child components, similar to how `SetViewMonth` syncs after calendar navigation.

2. **Exact cell format for dates 1-9 without indicators**
   - What we know: Cells must be 4 chars. Indicated: `[ 5]`, `[15]`. Non-indicated need matching width.
   - What's unclear: Whether ` 5 ` (space-padded center) or `  5 ` (right-aligned with trailing space) looks better visually.
   - Recommendation: Use `fmt.Sprintf(" %2d ", day)` which produces ` 5 ` for single-digit and ` 15` for double-digit (note: this is only 4 chars for single digit ` 5 ` but ` 15` is also 4 chars). Actually `" %2d "` always produces 4 chars: ` _5_` and `_15_` where _ is space. This is the cleanest format.

3. **What happens when calendarInnerWidth grows to ~38?**
   - What we know: The todo panel width is calculated as `m.width - calendarInnerWidth - (frameH * 2)`
   - What's unclear: Whether very narrow terminals (< 60 cols) still work
   - Recommendation: The existing guard `if todoInnerWidth < 1` handles this. The wider calendar is ~14 chars more, so terminals under ~52 cols would show "Terminal too small". This is acceptable for a TUI app.

## Sources

### Primary (HIGH confidence)
- Project source code: `internal/calendar/grid.go` - current RenderGrid implementation
- Project source code: `internal/config/config.go` - current Config struct
- Project source code: `internal/store/store.go` - current Store methods
- Project source code: `internal/app/model.go` - current layout constants and wiring
- Go stdlib `time` package - Weekday constants (Sunday=0 through Saturday=6)

### Secondary (MEDIUM confidence)
- [Lipgloss GitHub](https://github.com/charmbracelet/lipgloss) - ANSI width handling caveats
- [Go time package docs](https://pkg.go.dev/time) - Weekday enumeration

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - no new libraries, all existing code patterns
- Architecture: HIGH - patterns extend existing calendar/config/store code, all verifiable in source
- Pitfalls: HIGH - derived from reading actual source code and understanding ANSI rendering
- Code examples: MEDIUM - alignment math needs validation during implementation

**Research date:** 2026-02-05
**Valid until:** 2026-03-05 (stable domain, no external dependencies changing)
