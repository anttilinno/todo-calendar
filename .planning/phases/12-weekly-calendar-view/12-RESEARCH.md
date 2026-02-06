# Phase 12: Weekly Calendar View - Research

**Researched:** 2026-02-06
**Domain:** Go TUI calendar rendering, week-based date arithmetic, Bubble Tea view mode state management
**Confidence:** HIGH

## Summary

This phase adds a weekly calendar view mode that users can toggle into from the existing monthly view. The implementation touches the calendar model (new view mode state, week navigation, weekly grid renderer), the app model (toggle keybinding, syncing todolist view month when in weekly mode), and the calendar keys/help. No new libraries are needed -- Go's `time` package provides `Weekday()` and `AddDate()` for all week arithmetic, and the existing `holidays.Provider.HolidaysInMonth()` already returns the data needed for holiday display. The store already exposes `IncompleteTodosPerDay()` for indicators.

The core architectural decision is how to model the "current week" in the calendar. The cleanest approach is to track a `weekStart time.Time` field representing the Monday (or Sunday, depending on `mondayStart`) of the currently viewed week. When toggling from monthly to weekly view, this is computed from "today". Navigation adds/subtracts 7 days to `weekStart`. The weekly grid renderer is a new pure function (`RenderWeekGrid`) analogous to the existing `RenderGrid`, producing a single-row calendar showing 7 days with day numbers, holiday markers, and todo indicators.

The todolist must sync to the month containing the displayed week. Since a week can span two months (e.g., Jan 28 - Feb 3), the todolist should sync to the month containing the majority of the week's days, or more simply, the month of the Wednesday of the week (ISO convention). This avoids jarring month switches when navigating and keeps the UX predictable.

**Primary recommendation:** Add a `ViewMode` enum (`MonthView`/`WeekView`) and a `weekStart time.Time` field to the calendar model. Add a `RenderWeekGrid` pure function in `grid.go`. Toggle with `w` key. Navigate with existing left/right arrow keys (which get contextual behavior based on view mode). Sync todolist view month to `weekStart`'s month.

## Standard Stack

No new libraries needed. This feature uses only Go standard library and existing project dependencies.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go `time` package | Go 1.25.6 | `Weekday()` for week-start calculation, `AddDate(0,0,7)` for week navigation | Built-in, handles DST and calendar edge cases correctly |
| `charmbracelet/bubbletea` v1.3.10 | Already in go.mod | Model/Update/View architecture for view mode state | Already used for all UI |
| `charmbracelet/lipgloss` v1.1.0 | Already in go.mod | Styled rendering of weekly grid cells | Already used for all styling |
| `charmbracelet/bubbles` v0.21.1 | Already in go.mod | `key` package for keybinding | Already used for all keys |
| `rickar/cal/v2` v2.1.27 | Already in go.mod | Holiday lookup (via existing Provider) | Already used for holiday detection |

### Supporting
No additional libraries needed.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `time.Time` for weekStart | `snabb/isoweek` third-party package | Provides ISO week utilities but adds an unnecessary dependency; Go stdlib is sufficient for our needs |
| View mode enum in calendar model | Separate weekly model component | Separate component would duplicate holiday/indicator/navigation logic; a mode flag in the existing model is simpler |
| Reusing `RenderGrid` with week filter | New `RenderWeekGrid` function | Trying to force a 7-day view through the month grid renderer would add complexity to the existing function; a dedicated function is cleaner |

**Installation:** No new dependencies.

## Architecture Patterns

### View Mode State in Calendar Model

Add a `ViewMode` type and a `weekStart` field to the existing calendar model. The `weekStart` tracks the first day (Monday or Sunday, per `mondayStart`) of the currently displayed week.

```go
// ViewMode controls whether the calendar shows a full month or a single week.
type ViewMode int

const (
    MonthView ViewMode = iota
    WeekView
)

// Model additions:
type Model struct {
    // ... existing fields ...
    viewMode  ViewMode
    weekStart time.Time // first day of the currently viewed week
}
```

**Confidence:** HIGH -- follows the existing pattern of the `mode` enum in `todolist/model.go`.

### Week Start Calculation

When toggling from monthly to weekly view, compute the week start from today's date:

```go
// weekStartFor returns the date of the first day of the week containing t.
// If mondayStart is true, weeks start on Monday; otherwise Sunday.
func weekStartFor(t time.Time, mondayStart bool) time.Time {
    wd := int(t.Weekday()) // Sunday=0, Monday=1, ..., Saturday=6
    if mondayStart {
        // Monday=0, Tuesday=1, ..., Sunday=6
        offset := (wd + 6) % 7
        return t.AddDate(0, 0, -offset)
    }
    // Sunday=0 already
    return t.AddDate(0, 0, -wd)
}
```

**Confidence:** HIGH -- verified against Go `time.Weekday()` documentation and the existing `startCol` calculation in `grid.go`.

### Weekly Grid Renderer (Pure Function)

A new `RenderWeekGrid` function in `grid.go`, analogous to `RenderGrid`. It renders a single header row and a single data row of 7 days.

```go
// RenderWeekGrid produces a 34-character wide single-week grid with 4-char cells.
// It is a pure function with no side effects.
//
// Parameters:
//   - weekStart: the first day of the week to render
//   - today: today's date (zero value for none)
//   - holidays: provider for holiday lookup
//   - mondayStart: if true, weeks start on Monday
//   - indicators: map of "YYYY-MM-DD" -> count of incomplete todos
//   - s: calendar styles
func RenderWeekGrid(weekStart time.Time, today time.Time, holidays *holidays.Provider, mondayStart bool, store *store.Store, s Styles) string {
    // ... renders header line (week date range), weekday headers, 7 day cells
}
```

The key difference from `RenderGrid`: the weekly view needs to look up holidays and indicators across potentially two months (the week may straddle a month boundary). This means:
- Holidays: call `provider.HolidaysInMonth()` for both months if the week spans a boundary, or use a per-day check approach.
- Indicators: call `store.IncompleteTodosPerDay()` for both months if needed.

A simpler approach: compute holidays and indicators per-day inline during rendering, since we only have 7 days. This avoids managing two month-maps.

**Confidence:** HIGH -- the rendering logic is a direct simplification of the existing `RenderGrid`.

### Navigation Context Switching

The existing `PrevMonth`/`NextMonth` keybindings (left/right arrows, h/l) change behavior based on view mode:
- In `MonthView`: navigate by month (existing behavior)
- In `WeekView`: navigate by week (`weekStart = weekStart.AddDate(0, 0, -7)` or `+7`)

This avoids introducing new keybindings for week navigation. The help text updates to reflect the current mode.

```go
// In calendar Update():
case key.Matches(msg, m.keys.PrevMonth):
    if m.viewMode == WeekView {
        m.weekStart = m.weekStart.AddDate(0, 0, -7)
        m.year = m.weekStart.Year()
        m.month = m.weekStart.Month()
    } else {
        // existing month navigation
    }
```

When in weekly view, `m.year` and `m.month` should track the "dominant" month of the week (the month containing Wednesday, or simply the month of `weekStart`). This ensures the todolist syncs to a sensible month.

**Confidence:** HIGH -- `AddDate(0,0,7)` is the documented approach for week navigation per Go `time` package.

### Toggle Keybinding

Use `w` to toggle between monthly and weekly view. This key is currently unused (verified: app keys use `q`, `tab`, `s`; calendar keys use `left`/`right`/`h`/`l`; todo keys use `j`/`k`/`J`/`K`/`a`/`A`/`x`/`d`/`e`/`E`/`enter`/`esc`).

```go
// In calendar KeyMap:
ToggleWeek key.Binding

// Default:
ToggleWeek: key.NewBinding(
    key.WithKeys("w"),
    key.WithHelp("w", "weekly view"),
),
```

The toggle should be handled at the **calendar model level** (not app level) since it's a calendar-specific view concern. When toggling TO weekly view, compute `weekStart` from today. When toggling BACK to monthly view, restore `m.year` and `m.month` to sensible values (the month of the week that was being viewed).

**Confidence:** HIGH -- `w` is free in all modes, and the calendar model already handles its own navigation keys.

### Todolist Month Sync for Weekly View

The app model currently syncs the todolist with `m.todoList.SetViewMonth(m.calendar.Year(), m.calendar.Month())`. In weekly mode, the calendar's Year/Month should reflect the week being displayed. The simplest approach: when navigating weeks, update `m.year` and `m.month` to match `weekStart`'s year/month. The existing `Year()` and `Month()` methods then return the correct values automatically.

**Confidence:** HIGH -- the existing sync mechanism works without changes if Year/Month track weekStart.

### Weekly View Header

The weekly view header should show the date range of the week, e.g., "Feb 2 - Feb 8, 2026" or "Jan 28 - Feb 3, 2026" (when spanning months). This replaces the monthly header ("February 2026") while in weekly mode.

```go
// Header examples:
// Same month: "Feb 2 - 8, 2026"
// Cross month: "Jan 26 - Feb 1, 2026"
// Cross year: "Dec 29, 2025 - Jan 4, 2026"
```

**Confidence:** HIGH -- straightforward string formatting with Go `time.Format`.

### Recommended File Modifications

| File | Changes |
|------|---------|
| `internal/calendar/model.go` | Add `ViewMode` type, `viewMode` and `weekStart` fields, toggle logic in `Update()`, conditional rendering in `View()`, `weekStartFor()` helper, updated `Year()`/`Month()` behavior |
| `internal/calendar/grid.go` | Add `RenderWeekGrid()` function |
| `internal/calendar/keys.go` | Add `ToggleWeek` binding to `KeyMap`, update `ShortHelp()`/`FullHelp()`, make `PrevMonth`/`NextMonth` help text dynamic |
| `internal/app/model.go` | Sync todolist view month after toggle and week navigation (existing sync mechanism suffices), update help bar to show mode-appropriate bindings |

**Files NOT modified:** `store/`, `todolist/`, `theme/`, `holidays/`, `settings/`, `config/`, `main.go`.

### Anti-Patterns to Avoid
- **Creating a separate WeeklyCalendar component:** This would duplicate holiday/indicator/style/navigation logic. Use a mode flag in the existing calendar model instead.
- **Modifying the store for weekly queries:** The store already has `IncompleteTodosPerDay(year, month)`. For a 7-day view, call this for the relevant month(s). Don't add a per-week query.
- **Changing the monthly grid renderer:** `RenderGrid` works perfectly for monthly view. Don't add weekly-mode conditionals to it. Create a separate `RenderWeekGrid` pure function.
- **Using ISO week numbers for navigation:** ISO weeks start on Monday unconditionally. Our app has configurable `mondayStart`. Track `weekStart` as a `time.Time` and calculate from the configured week start day.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Week start date calculation | Manual day arithmetic with magic numbers | `time.Weekday()` + `time.AddDate()` | Handles DST transitions, leap years, month boundaries correctly |
| Week navigation (prev/next) | `weekStart - 7*24*time.Hour` duration math | `weekStart.AddDate(0, 0, -7)` | Duration math breaks on DST transitions; `AddDate` is calendar-aware |
| Holiday lookup per day | Custom holiday map merging for cross-month weeks | `provider.HolidaysInMonth()` for each relevant month | Already exists and is tested |
| Indicator lookup per day | Custom per-week indicator query | `store.IncompleteTodosPerDay()` for each relevant month | Already exists and handles all filtering |

**Key insight:** All data access methods already exist. The weekly view is purely a rendering change with minimal model state additions.

## Common Pitfalls

### Pitfall 1: Week Straddling Month Boundaries
**What goes wrong:** A week like Jan 28 - Feb 3 spans two months. If you only fetch holidays and indicators for one month, days from the other month show no holidays or indicators.
**Why it happens:** `HolidaysInMonth()` and `IncompleteTodosPerDay()` both take a single `(year, month)` pair.
**How to avoid:** For a 7-day week, check which month each day belongs to. If the week straddles two months, fetch data for both months and merge. Since we only have 7 days, this is trivially efficient. Alternatively, check holidays per-day by calling `provider.cal.IsHoliday()` directly, but this would require exposing the `cal` field -- fetching two month-maps is simpler and matches the existing API.
**Warning signs:** Holiday markers disappear for the last few or first few days of a straddling week.

### Pitfall 2: MondayStart Inconsistency Between Weekly and Monthly View
**What goes wrong:** The monthly grid already handles `mondayStart` for column ordering. If the weekly view calculates week start differently (e.g., always using ISO Monday-start), the two views show inconsistent week boundaries.
**Why it happens:** ISO weeks always start on Monday, but the app's `mondayStart` config might be `false` (Sunday start). Using `time.ISOWeek()` naively forces Monday start.
**How to avoid:** Do NOT use `ISOWeek()` for week start calculation. Use the `weekStartFor()` helper that respects `mondayStart`. The weekly grid's day order must match the monthly grid's column order.
**Warning signs:** User sets "Sunday" as first day of week, toggles to weekly view, and the week starts on Monday.

### Pitfall 3: Toggle to Weekly View Shows Wrong Week
**What goes wrong:** When toggling from monthly view, the weekly view shows the first week of the month instead of the week containing today.
**Why it happens:** Developer computes `weekStart` from the first day of the current calendar month instead of from `time.Now()`.
**How to avoid:** WKVIEW-04 requires: "Current week is auto-selected when switching from monthly to weekly view." Always compute `weekStart = weekStartFor(time.Now(), m.mondayStart)` when toggling to weekly mode.
**Warning signs:** User toggles to weekly view on Feb 6 and sees the Jan 26 - Feb 1 week instead of Feb 2 - Feb 8.

### Pitfall 4: Year/Month Desync After Week Navigation
**What goes wrong:** After navigating weeks in weekly mode, switching back to monthly mode shows the wrong month because `m.year`/`m.month` weren't updated during week navigation.
**Why it happens:** Week navigation changes `weekStart` but forgets to update `m.year`/`m.month`.
**How to avoid:** After every `weekStart` change (navigation or toggle), update `m.year` and `m.month` to match `weekStart`. This keeps the todolist in sync and ensures toggling back to monthly shows the right month.
**Warning signs:** User navigates forward a few weeks, toggles back to monthly, and the calendar still shows the old month.

### Pitfall 5: Navigation Key Help Text Not Updating
**What goes wrong:** Help bar says "prev month / next month" while in weekly view, confusing the user.
**Why it happens:** The `KeyMap` help text is set once at construction time.
**How to avoid:** Make the `Keys()` method or help bindings context-aware. Return different help text based on `viewMode`. Or create separate key bindings with different help text and swap them based on mode.
**Warning signs:** Help bar says "prev month" but pressing left navigates by week.

### Pitfall 6: Week Grid Width Mismatch
**What goes wrong:** The weekly grid renders at a different width than the monthly grid, causing visual jumping when toggling.
**Why it happens:** Different spacing/formatting logic in `RenderWeekGrid` vs `RenderGrid`.
**How to avoid:** Use the same `gridWidth` constant (34 chars) and the same 4-char cell layout in `RenderWeekGrid`. The weekly grid should be visually identical to a single row of the monthly grid, just with a different header.
**Warning signs:** Calendar pane border shifts or content jumps when toggling between views.

## Code Examples

### Week Start Calculation (Respecting mondayStart)

```go
// Source: Go time package (Weekday), verified against existing grid.go startCol logic
func weekStartFor(t time.Time, mondayStart bool) time.Time {
    wd := int(t.Weekday()) // Sunday=0 .. Saturday=6
    if mondayStart {
        offset := (wd + 6) % 7 // Monday=0 .. Sunday=6
        return time.Date(t.Year(), t.Month(), t.Day()-offset, 0, 0, 0, 0, time.Local)
    }
    return time.Date(t.Year(), t.Month(), t.Day()-wd, 0, 0, 0, 0, time.Local)
}
```

### Week Navigation

```go
// Source: Go time package AddDate documentation
// Navigate to previous week:
m.weekStart = m.weekStart.AddDate(0, 0, -7)

// Navigate to next week:
m.weekStart = m.weekStart.AddDate(0, 0, 7)

// Update tracking fields:
m.year = m.weekStart.Year()
m.month = m.weekStart.Month()
```

### Toggle Logic

```go
// In calendar Update(), handling toggle key:
case key.Matches(msg, m.keys.ToggleWeek):
    if m.viewMode == MonthView {
        m.viewMode = WeekView
        m.weekStart = weekStartFor(time.Now(), m.mondayStart)
        m.year = m.weekStart.Year()
        m.month = m.weekStart.Month()
    } else {
        m.viewMode = MonthView
        // m.year and m.month already track the week's month
    }
    m.refreshData()
```

### Weekly Grid Header (Date Range)

```go
// Source: Go time.Format documentation
func weekRangeHeader(weekStart time.Time) string {
    weekEnd := weekStart.AddDate(0, 0, 6)
    if weekStart.Month() == weekEnd.Month() {
        // Same month: "Feb 2 - 8, 2026"
        return fmt.Sprintf("%s %d - %d, %d",
            weekStart.Month().String()[:3], weekStart.Day(),
            weekEnd.Day(), weekEnd.Year())
    }
    if weekStart.Year() == weekEnd.Year() {
        // Cross month: "Jan 26 - Feb 1, 2026"
        return fmt.Sprintf("%s %d - %s %d, %d",
            weekStart.Month().String()[:3], weekStart.Day(),
            weekEnd.Month().String()[:3], weekEnd.Day(),
            weekEnd.Year())
    }
    // Cross year: "Dec 29, 2025 - Jan 4, 2026"
    return fmt.Sprintf("%s %d, %d - %s %d, %d",
        weekStart.Month().String()[:3], weekStart.Day(), weekStart.Year(),
        weekEnd.Month().String()[:3], weekEnd.Day(), weekEnd.Year())
}
```

### Holiday/Indicator Lookup for Cross-Month Week

```go
// For a 7-day week that may straddle two months:
func holidaysForWeek(weekStart time.Time, provider *holidays.Provider) map[int]bool {
    // Collect holidays for both months the week might touch
    m1 := provider.HolidaysInMonth(weekStart.Year(), weekStart.Month())
    weekEnd := weekStart.AddDate(0, 0, 6)
    m2 := provider.HolidaysInMonth(weekEnd.Year(), weekEnd.Month())

    // Build per-day holiday map keyed by day offset (0-6)
    result := make(map[int]bool)
    for i := 0; i < 7; i++ {
        d := weekStart.AddDate(0, 0, i)
        if d.Month() == weekStart.Month() {
            if m1[d.Day()] {
                result[i] = true
            }
        } else {
            if m2[d.Day()] {
                result[i] = true
            }
        }
    }
    return result
}
```

### Contextual Help Text

```go
// Dynamic key help based on view mode:
func (m Model) Keys() KeyMap {
    k := m.keys
    if m.viewMode == WeekView {
        k.PrevMonth = key.NewBinding(
            key.WithKeys("left", "h"),
            key.WithHelp("<-/h", "prev week"),
        )
        k.NextMonth = key.NewBinding(
            key.WithKeys("right", "l"),
            key.WithHelp("->/l", "next week"),
        )
    }
    return k
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Monthly view only | Monthly + weekly toggle | This phase | Users can focus on a single week |
| Fixed month navigation (left/right) | Context-aware navigation (month or week) | This phase | Same keys, different behavior per mode |
| No view mode concept | `ViewMode` enum in calendar model | This phase | Foundation for potential future daily view |

## Open Questions

1. **What should happen to the overview section in weekly view?**
   - What we know: The overview panel shows per-month todo counts below the calendar grid. In weekly view, the grid is much shorter (2 rows vs 6-8 rows).
   - Options: (a) Keep overview unchanged, (b) Hide overview in weekly view, (c) Show a week-specific summary
   - Recommendation: Keep overview unchanged. The extra vertical space from the shorter grid naturally accommodates the overview. The overview shows cross-month data which is still useful context. Simplest implementation.

2. **Should toggling back to monthly view return to the "current month" or the month the user was viewing in weekly mode?**
   - What we know: When toggling TO weekly, WKVIEW-04 says use "current week" (today). When toggling back, the requirement is silent.
   - Recommendation: Return to the month containing the weekStart. This preserves context -- if the user navigated to a different week, they probably want to see that month in the monthly view, not snap back to the original month.

3. **How should the weekly grid handle the visual space below the single row of days?**
   - What we know: Monthly view uses 6-8 lines for the grid. Weekly view uses 2-3 lines (header + weekday labels + one row of days).
   - Recommendation: Let the overview section fill the extra space naturally. The calendar pane has a fixed width but the height is set by the app model. No special handling needed -- lipgloss will render the shorter content and the overview will appear closer to the top.

## Sources

### Primary (HIGH confidence)
- **Codebase audit** -- All 22 Go source files read and analyzed for calendar rendering, navigation, key binding, and data access patterns
- **Go `time` package documentation** -- [pkg.go.dev/time](https://pkg.go.dev/time) -- verified `ISOWeek()`, `Weekday()`, `AddDate()` methods
- **Existing grid.go** -- `RenderGrid` function provides the exact template for `RenderWeekGrid`
- **Existing model.go patterns** -- `ViewMode` enum follows `todolist.mode` pattern; `weekStartFor()` follows `startCol` calculation in grid.go

### Secondary (MEDIUM confidence)
- **WebSearch: Go week calculation** -- confirmed `AddDate(0,0,7)` is preferred over duration arithmetic for calendar-aware week navigation
- **ISO 8601 week standard** -- confirmed week numbering conventions (weeks 1-53, Monday start for ISO)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new libraries, all Go stdlib and existing dependencies
- Architecture: HIGH -- follows established patterns in codebase, pure functions for rendering, mode enum for state
- Pitfalls: HIGH -- cross-month week straddling and mondayStart consistency identified from codebase analysis
- Code examples: HIGH -- based on actual codebase patterns and verified Go time package API

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (stable domain, no external dependencies to change)
