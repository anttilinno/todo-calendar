# Stack Research

**Domain:** TUI Calendar v1.3 features (weekly view, search/filter, overview colors, date format)
**Researched:** 2026-02-06
**Confidence:** HIGH

## Verdict: No New Dependencies Required

All four v1.3 features can be implemented using the existing stack with zero new library additions. The existing Go stdlib, Bubble Tea v1.3.10, Lipgloss v1.1.0, and Bubbles v0.21.1 provide everything needed.

## Core Technologies (No Changes)

| Technology | Current Version | Status for v1.3 |
|------------|----------------|------------------|
| Go | 1.25.6 | Sufficient. `strings`, `time`, `fmt` stdlib packages cover all new needs |
| Bubble Tea | v1.3.10 | Sufficient. Elm Architecture handles new view modes and overlays |
| Lipgloss | v1.1.0 | Sufficient. `Foreground()`, `Background()`, `Bold()` cover all new color coding |
| Bubbles | v0.21.1 | Sufficient. `textinput.Model` already used; reuse for search filter input |
| BurntSushi/toml | v1.6.0 | Sufficient. New config field `date_format` is just a string |
| rickar/cal/v2 | v2.1.27 | Unchanged. No new holiday needs |
| encoding/json | stdlib | Unchanged. No storage schema changes |

## Supporting Libraries

### Evaluated and Rejected

| Library | Purpose | Why NOT to Add |
|---------|---------|----------------|
| sahilm/fuzzy | Fuzzy string matching for search | Overkill. Todo list is small (tens to low hundreds of items). Simple `strings.Contains(strings.ToLower(...))` is sufficient and adds zero dependencies. Fuzzy matching would confuse users searching short todo text strings. |
| bubbles/list | List component with built-in filtering | Already deliberately not used. The project has a custom todo list rendering that integrates tightly with the section-header/todo/empty item model. Adopting `list.Model` would require a rewrite of the todo pane for no benefit. |
| bubbles/viewport | Scrollable content container | Already not needed for the main views. For the full-screen search overlay, results fit in rendered string lines with manual scroll tracking -- same pattern as the existing todo list. Only consider if search results routinely exceed terminal height, which is unlikely with a personal todo list. |
| charmbracelet/huh | Form/dialog components | No forms needed. All four features use existing patterns: cycling options in settings, textinput for search, pure rendering for weekly view and colors. |

### Confirmed Existing Components to Reuse

| Component | Import | Current Use | v1.3 Reuse |
|-----------|--------|-------------|------------|
| `bubbles/textinput` | `textinput.Model` | Todo add/edit text input | Search filter input (inline and overlay) |
| `bubbles/key` | `key.Binding` | Keybinding definitions | New bindings for search toggle, view toggle, filter activation |
| `bubbles/help` | `help.Model` | Context-sensitive help bar | Updated help for new modes |
| `lipgloss.Style` | `lipgloss.NewStyle()` | All styled rendering | Overview color coding (red/green foreground on month counts) |

## Feature-Specific Stack Notes

### 1. Weekly Calendar View

**Stack impact: None. Pure rendering logic change.**

The weekly view is a variant of `RenderGrid()` in `calendar/grid.go`. The existing pure function takes `(year, month, today, holidays, mondayStart, indicators, styles)` and returns a string. A weekly variant would take an additional parameter (the "focus week" or current date) and render a single 7-day row instead of the full month grid.

Key considerations:
- The `gridWidth` constant (34 chars = 7 cells x 4 chars + 6 separators) stays the same for weekly view -- same 7 columns, just 1 row instead of 4-6
- The `Model` needs a `viewMode` field (monthly/weekly) to toggle between `RenderGrid` and a new `RenderWeek` function
- Week calculation uses Go stdlib: `time.Date(year, month, day, ...).Weekday()` to find week boundaries, same as existing grid logic
- The overview panel rendering is independent of grid mode -- no changes needed there
- Toggle keybinding (e.g., `w`) switches `viewMode` in the calendar model

**No new imports required.** All date math uses `time` stdlib already imported.

### 2. Search/Filter Todos

**Stack impact: None. Reuses existing `textinput.Model` pattern.**

The project already has a `textinput.Model` instance in `todolist/model.go` for adding and editing todos. Search/filter uses the same component with a different prompt.

**Inline filter (todo panel):**
- Activates with a keybinding (e.g., `/`)
- Reuses or creates a second `textinput.Model` with `Placeholder: "Filter..."` and `Prompt: "/ "`
- Filters `visibleItems()` by checking `strings.Contains(strings.ToLower(todo.Text), strings.ToLower(query))`
- Renders filtered results using the same `renderTodo()` method
- No new Bubbles components needed

**Full-screen search overlay:**
- Same pattern as the existing `settings.Model` overlay (full-screen, shown/hidden via `showSettings` flag in app model)
- Contains its own `textinput.Model` for the search query
- Iterates `store.Todos()` (all todos) and filters with case-insensitive substring matching
- Renders results grouped by month, similar to existing `visibleItems()` but across all months
- User can select a result and jump to that month

**Why NOT fuzzy matching:**
- Todo text is short (5-30 chars typically). Fuzzy matching excels with long filenames/paths, not short phrases.
- Substring search is instantly understandable: type "doc" to find "Doctor appointment"
- `strings.Contains(strings.ToLower(s), strings.ToLower(q))` is O(n*m) per item but with <1000 items and <120 char strings, this is sub-millisecond
- Zero dependency cost

**Stdlib functions used:**
```go
import "strings"

// Case-insensitive substring match -- sufficient for personal todo list
func matchesTodo(todo store.Todo, query string) bool {
    return strings.Contains(
        strings.ToLower(todo.Text),
        strings.ToLower(query),
    )
}
```

### 3. Overview Color Coding

**Stack impact: None. Extends existing theme system.**

The theme system already has 14 semantic color roles in `theme.Theme`. Overview color coding needs 2 new roles:

```go
// Add to theme.Theme struct:
OverviewDoneFg    lipgloss.Color // months where all todos are completed (green)
OverviewPendingFg lipgloss.Color // months with uncompleted todos (red)
```

These integrate into the existing `calendar/styles.go` Styles struct:

```go
// Add to calendar.Styles struct:
OverviewDone    lipgloss.Style // green text for completed months
OverviewPending lipgloss.Style // red text for months with pending work
```

The `renderOverview()` method in `calendar/model.go` already iterates `MonthCount` entries. It just needs to check completion status (new store method: `CompletedTodoCountForMonth(year, month)`) and apply the appropriate style.

**All 4 themes need the 2 new color values.** Example assignments:
- Dark: done=`#00AF00` (green), pending=`#AF0000` (red) -- matches existing `HolidayFg` red
- Light: done=`#008700` (dark green), pending=`#D70000` (dark red)
- Nord: done=`#A3BE8C` (nord14 aurora green, already used for IndicatorFg), pending=`#BF616A` (nord11 aurora red, already used for HolidayFg)
- Solarized: done=`#859900` (green, already used for IndicatorFg), pending=`#DC322F` (red, already used for HolidayFg)

**No new libraries.** Lipgloss `Foreground()` handles all color application.

### 4. Date Format Setting

**Stack impact: None, but requires careful handling of Go's reference time.**

Go's `time.Format()` uses a reference time layout, NOT format specifiers like `strftime`. This is the single most error-prone aspect of this feature.

**Go Reference Time (memorize this):**
```
Mon Jan 2 15:04:05 MST 2006
 01  02            01    06
```

The reference time is: **January 2, 2006, 3:04:05 PM, MST (UTC-0700)**

**Three preset date formats and their Go layout strings:**

| Display Name | Example | Go Layout String | Notes |
|-------------|---------|------------------|-------|
| YYYY-MM-DD (ISO) | 2026-02-06 | `"2006-01-02"` | Same as `time.DateOnly` constant (Go 1.20+). Current internal format. |
| DD.MM.YYYY (European) | 06.02.2026 | `"02.01.2006"` | `02` = zero-padded day, `01` = zero-padded month |
| MM/DD/YYYY (US) | 02/06/2026 | `"01/02/2006"` | Looks like a date itself -- this IS the Go reference time date! |

**Critical Go `time.Format` quirks to handle:**

1. **`01/02/2006` IS the reference date.** The US format `MM/DD/YYYY` maps to `01/02/2006` which looks like January 2, 2006 -- because it IS. This will confuse developers reading the code. Add a comment.

2. **Day padding matters.** `2` = no padding (single digit: `6`), `02` = zero-padded (`06`), `_2` = space-padded (` 6`). All three presets should use `02` (zero-padded) for consistency.

3. **Month `1` vs `01`.** `1` = no zero padding, `01` = zero-padded. Use `01` in all presets.

4. **Internal storage stays YYYY-MM-DD.** The date format setting is DISPLAY ONLY. The JSON store (`todo.Date`) and all `time.Parse()` calls continue using `"2006-01-02"`. Only rendering (`renderTodo`, `renderOverview`, date input placeholder, calendar header) changes.

5. **Custom format option.** If offering custom format strings, the user types a Go layout string. This is unintuitive for most users. Recommendation: provide the 3 presets in settings and defer custom format to config file editing. In the TOML config, `date_format = "02.01.2006"` is a raw Go layout string. Document the reference time in a comment in the default config.

**Config integration:**

```go
// In config.Config struct:
DateFormat string `toml:"date_format"`

// In config.DefaultConfig():
DateFormat: "2006-01-02"
```

**Settings integration:**

Add a 4th option row to the existing settings overlay, cycling through the 3 presets:

```go
// Preset mapping
var dateFormatPresets = []struct {
    Layout  string // Go layout string
    Display string // Human-readable label
    Example string // Example output
}{
    {"2006-01-02", "ISO (YYYY-MM-DD)", "2026-02-06"},
    {"02.01.2006", "European (DD.MM.YYYY)", "06.02.2026"},
    {"01/02/2006", "US (MM/DD/YYYY)", "02/06/2026"},
}
```

The settings `option` type already supports `values []string` and `display []string` -- just add a 4th `option` entry. The existing cycling mechanism (left/right arrows) works as-is.

**Where date format applies (all in View/render functions):**

| Location | Current Code | Change Needed |
|----------|-------------|---------------|
| `todolist/model.go` renderTodo | `t.Date` displayed raw | Format with `time.Parse(dateFormat, t.Date)` then `d.Format(displayLayout)` |
| `todolist/model.go` date input placeholder | Hardcoded `"YYYY-MM-DD"` | Show placeholder matching current format |
| `todolist/model.go` date validation | `time.Parse("2006-01-02", date)` | Parse input using display format, store as `"2006-01-02"` |
| `calendar/model.go` overview month labels | `mc.Month.String()` | No change -- month names are not date-formatted |

**Date INPUT is the tricky part.** When the user types a date while adding/editing a todo, they should type in their configured display format (e.g., `06.02.2026` for European). The parse logic must:
1. Parse the input string using the display format layout
2. Store the result formatted as `"2006-01-02"` (internal canonical format)

```go
// Parse user input in their configured format, return canonical YYYY-MM-DD
func parseUserDate(input string, displayLayout string) (string, error) {
    t, err := time.Parse(displayLayout, input)
    if err != nil {
        return "", err
    }
    return t.Format("2006-01-02"), nil
}
```

## What NOT to Add

| Avoid | Why | What to Do Instead |
|-------|-----|-------------------|
| `sahilm/fuzzy` | Overkill for small todo lists; substring search is clearer for short text; adds a dependency for zero user benefit | `strings.Contains(strings.ToLower(...), strings.ToLower(...))` |
| `bubbles/list` | Would require rewriting the entire todo pane to adopt its model; existing custom rendering is simpler and more flexible | Keep custom `visibleItems()` + `renderTodo()` pattern |
| `bubbles/viewport` | Search overlay results are small enough for simple line-based rendering with manual cursor tracking | Manual scroll offset like the existing todo list |
| `bubbles/paginator` | No pagination needed; search results are a small subset of a small dataset | Simple scroll if results exceed height |
| `charmbracelet/huh` | No form dialogs needed; all inputs are single textinput fields or cycling selectors | Existing `textinput.Model` and `settings.option` cycling |
| Any date picker library | Go stdlib `time` package handles all date math; no complex date selection UI needed in TUI | Manual date entry with format-aware parsing |
| `regexp` for search | Regular expressions are overkill and confusing for users searching their todo list | `strings.Contains` substring matching |

## Version Compatibility

No changes to dependency versions. The `go.mod` stays exactly as-is:

```
go 1.25.6

require (
    github.com/BurntSushi/toml v1.6.0
    github.com/charmbracelet/bubbles v0.21.1
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/rickar/cal/v2 v2.1.27
)
```

## New Store Methods Needed

These use only existing imports (`time`, `sort`) already in `store/store.go`:

| Method | Purpose | Used By |
|--------|---------|---------|
| `SearchTodos(query string) []Todo` | Return all todos whose text contains query (case-insensitive) | Search overlay |
| `CompletionStatusByMonth() []MonthStatus` | Return per-month done/total counts for overview coloring | Overview color coding |
| `TodosForWeek(year int, month time.Month, day int) []Todo` | Return todos for the 7-day week containing the given date | Weekly view todo filtering |

All implementable with existing `time.Parse`, `strings.Contains`, `strings.ToLower` from Go stdlib.

## New Theme Fields Needed

Two new `lipgloss.Color` fields added to `theme.Theme`:

```go
OverviewDoneFg    lipgloss.Color // completed months in overview
OverviewPendingFg lipgloss.Color // pending months in overview
```

This brings the theme from 14 to 16 semantic color roles. The existing pattern (define in each of the 4 theme functions, consume via `NewStyles()` constructor) scales cleanly.

## New Config Field Needed

One new string field in `config.Config`:

```go
DateFormat string `toml:"date_format"` // Go time layout string
```

Default: `"2006-01-02"`. Validated against the 3 known presets on load (unknown values fall back to default, same pattern as theme/country).

## Architecture Impact Summary

| Feature | New Files | Modified Files | New Types/Structs |
|---------|-----------|----------------|-------------------|
| Weekly view | 0 (add `RenderWeek` to existing `grid.go`) | `calendar/grid.go`, `calendar/model.go`, `calendar/keys.go` | `viewMode` enum in calendar |
| Search/filter | 1 (`internal/search/model.go` for overlay) or 0 (embed in app) | `todolist/model.go`, `app/model.go`, `store/store.go` | `searchMode` in todolist or new `search.Model` |
| Overview colors | 0 | `theme/theme.go`, `calendar/styles.go`, `calendar/model.go`, `store/store.go` | 2 new theme fields, 2 new styles |
| Date format | 0 | `config/config.go`, `settings/model.go`, `todolist/model.go` | 1 new config field, 1 new settings option |

**Total new Go files: 0-1. Total modified files: ~10. Zero new dependencies.**

## Sources

- [Go time package documentation](https://pkg.go.dev/time) -- DateOnly constant, Format reference time layout, all format verbs (verified 2026-02-06) -- HIGH confidence
- [Go time format cheatsheet](https://gosamples.dev/date-time-format-cheatsheet/) -- Reference time component mapping -- HIGH confidence
- [Bubbles textinput documentation](https://pkg.go.dev/github.com/charmbracelet/bubbles/textinput) -- CharLimit, Placeholder, Prompt, Focus/Blur API -- HIGH confidence
- [Bubbles list filtering](https://pkg.go.dev/github.com/charmbracelet/bubbles/list) -- DefaultFilter uses sahilm/fuzzy; confirmed NOT a transitive dep of this project -- HIGH confidence
- [sahilm/fuzzy GitHub](https://github.com/sahilm/fuzzy) -- Fuzzy matching library; evaluated and rejected for this use case -- HIGH confidence
- [Case-insensitive string search in Go](https://programming-idioms.org/idiom/133/case-insensitive-string-contains/1723/go) -- `strings.ToLower` + `strings.Contains` is the standard approach -- HIGH confidence
- [Bubbles viewport](https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport) -- Scrollable container; evaluated and not needed -- MEDIUM confidence
- Existing codebase analysis: `go.mod`, `go.sum`, all 21 `.go` source files reviewed -- HIGH confidence

---
*Stack research for: Todo Calendar v1.3 features*
*Researched: 2026-02-06*
