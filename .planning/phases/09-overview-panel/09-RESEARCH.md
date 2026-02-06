# Phase 9: Overview Panel - Research

**Researched:** 2026-02-06
**Domain:** Bubble Tea sub-component rendering, store aggregation queries, calendar pane layout
**Confidence:** HIGH

## Summary

Phase 9 adds an "overview panel" below the calendar grid in the left pane. This panel shows todo counts per month (e.g., `January [7]`) for all months that have todos, plus a count of undated floating todos (e.g., `Unknown [12]`). Counts must update live as todos are added, completed, or deleted.

The existing codebase already has all the building blocks. The store has `TodosForMonth(year, month)` and `FloatingTodos()` which can be used to compute counts. The calendar model's `View()` calls `RenderGrid()` and returns its output -- the overview can be appended below the grid in the same `View()` method. The live-update mechanism already exists: `app.Model.Update()` calls `calendar.RefreshIndicators()` after every update cycle. The same pattern extends to overview data -- either by adding a `RefreshOverview()` method or by computing overview data at render time directly from the store.

The main design decisions are: (1) how to efficiently query counts across all months, (2) how to format the overview to fit within the 38-char calendar pane, and (3) whether to cache counts or compute them on every render. Given the small data volume (personal todo list), computing counts on every `View()` call from the store is the simplest and most correct approach -- it guarantees live updates without any cache invalidation complexity.

**Primary recommendation:** Add a `TodoCountsByMonth()` method to the store that returns a sorted list of (year, month, count) tuples for all months that have todos. Render the overview below the calendar grid inside `calendar.Model.View()`, using the same Styles system for consistent theming.

## Standard Stack

No new dependencies required. Everything needed is already in the project.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `internal/store` | project | New `TodoCountsByMonth()` method for aggregation queries | Follows existing query pattern (TodosForMonth, FloatingTodos) |
| `internal/calendar` | project | Extend View() to render overview below grid | Natural home -- overview is part of the calendar pane |
| lipgloss | v1.1.0 | Style overview text consistently with theme | Already used in calendar styles |

### Existing (unchanged)
| Library | Version | Purpose |
|---------|---------|---------|
| Bubble Tea | v1.3.10 | TUI framework |
| Bubbles | v0.21.1 | key.Binding, help.Model |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Computing counts in View() | Caching counts in model fields | Cache requires invalidation on every mutation. For a personal todo list (<100 items), iterating on every render is negligible. Simplicity wins. |
| Rendering overview inside calendar.Model.View() | Separate overview component composed in app.Model.View() | Adds unnecessary complexity. The overview is visually and logically part of the calendar pane. Keeping it inside calendar.Model maintains cohesion. |
| Store method returning counts | Looping through Todos() in the calendar model | Breaks encapsulation -- the calendar model should not be parsing todo dates directly. Store methods keep the query logic centralized. |

## Architecture Patterns

### Files Modified
```
internal/
  store/
    store.go      # Add TodoCountsByMonth() and FloatingTodoCount() methods
  calendar/
    model.go      # Extend View() to render overview below grid
    styles.go     # Add OverviewHeader and OverviewCount styles
```

### Pattern 1: Store Aggregation Query

**What:** Add a method `TodoCountsByMonth()` that iterates all todos, groups by year-month, and returns a sorted slice of structs with year, month, and count. Add `FloatingTodoCount()` that returns the count of floating (undated) todos. These are read-only queries with no side effects.

**When to use:** Called from `calendar.View()` on every render.

**Confidence:** HIGH (follows exact pattern of existing store queries)

**Example:**
```go
// internal/store/store.go

// MonthCount represents the todo count for a specific year-month.
type MonthCount struct {
    Year  int
    Month time.Month
    Count int
}

// TodoCountsByMonth returns the number of todos in each year-month
// that has at least one todo, sorted chronologically.
func (s *Store) TodoCountsByMonth() []MonthCount {
    counts := make(map[string]*MonthCount)
    for _, t := range s.data.Todos {
        if t.Date == "" {
            continue
        }
        d, err := time.Parse(dateFormat, t.Date)
        if err != nil {
            continue
        }
        key := fmt.Sprintf("%d-%02d", d.Year(), d.Month())
        if counts[key] == nil {
            counts[key] = &MonthCount{Year: d.Year(), Month: d.Month()}
        }
        counts[key].Count++
    }
    result := make([]MonthCount, 0, len(counts))
    for _, mc := range counts {
        result = append(result, *mc)
    }
    sort.Slice(result, func(i, j int) bool {
        if result[i].Year != result[j].Year {
            return result[i].Year < result[j].Year
        }
        return result[i].Month < result[j].Month
    })
    return result
}

// FloatingTodoCount returns the number of todos with no date assigned.
func (s *Store) FloatingTodoCount() int {
    count := 0
    for _, t := range s.data.Todos {
        if !t.HasDate() {
            count++
        }
    }
    return count
}
```

### Pattern 2: Overview Rendered Below Calendar Grid

**What:** The calendar `View()` method currently returns just `RenderGrid(...)`. Extend it to append the overview section below. The overview lists each month with todos and the floating count. The overview fits within the same 34-char grid width.

**When to use:** Always -- the overview is a permanent part of the calendar pane view.

**Confidence:** HIGH (straightforward string concatenation in View())

**Example:**
```go
// calendar/model.go

func (m Model) View() string {
    todayDay := 0
    now := time.Now()
    if now.Year() == m.year && now.Month() == m.month {
        todayDay = now.Day()
    }

    grid := RenderGrid(m.year, m.month, todayDay, m.holidays, m.mondayStart, m.indicators, m.styles)

    // Append overview below grid
    overview := m.renderOverview()
    if overview != "" {
        return grid + "\n" + overview
    }
    return grid
}
```

### Pattern 3: Overview Rendering Format

**What:** The overview renders as a compact list of month-count rows below a section header. Format matches the requirement examples: `January [7]`, `Unknown [12]`. The currently viewed month is highlighted to show which month the user is looking at.

**When to use:** In the `renderOverview()` method.

**Confidence:** HIGH (simple string building following existing patterns)

**Example:**
```go
// calendar/model.go

func (m Model) renderOverview() string {
    var b strings.Builder

    b.WriteString("\n")
    b.WriteString(m.styles.OverviewHeader.Render("Overview"))
    b.WriteString("\n")

    monthCounts := m.store.TodoCountsByMonth()
    for _, mc := range monthCounts {
        label := fmt.Sprintf("%s %d", mc.Month.String(), mc.Year)
        count := fmt.Sprintf("[%d]", mc.Count)
        line := fmt.Sprintf(" %-20s %s", label, count)

        if mc.Year == m.year && mc.Month == m.month {
            b.WriteString(m.styles.OverviewActive.Render(line))
        } else {
            b.WriteString(m.styles.OverviewCount.Render(line))
        }
        b.WriteString("\n")
    }

    // Floating count
    floatingCount := m.store.FloatingTodoCount()
    line := fmt.Sprintf(" %-20s [%d]", "Unknown", floatingCount)
    b.WriteString(m.styles.OverviewCount.Render(line))
    b.WriteString("\n")

    return b.String()
}
```

### Pattern 4: Theme Integration for Overview Styles

**What:** Add `OverviewHeader`, `OverviewCount`, and `OverviewActive` styles to the calendar `Styles` struct, following the exact same pattern as existing styles. `OverviewHeader` uses the same AccentFg as section headers elsewhere. `OverviewCount` uses MutedFg for non-active months. `OverviewActive` uses the normal foreground or a bold style to highlight the currently viewed month.

**When to use:** In `calendar/styles.go` `NewStyles()`.

**Confidence:** HIGH (exact pattern from existing styles)

**Example:**
```go
// calendar/styles.go

type Styles struct {
    Header          lipgloss.Style
    WeekdayHdr      lipgloss.Style
    Normal          lipgloss.Style
    Today           lipgloss.Style
    Holiday         lipgloss.Style
    Indicator       lipgloss.Style
    OverviewHeader  lipgloss.Style  // NEW
    OverviewCount   lipgloss.Style  // NEW
    OverviewActive  lipgloss.Style  // NEW
}

func NewStyles(t theme.Theme) Styles {
    return Styles{
        // ... existing styles ...
        OverviewHeader: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
        OverviewCount:  lipgloss.NewStyle().Foreground(t.MutedFg),
        OverviewActive: lipgloss.NewStyle().Bold(true).Foreground(t.NormalFg),
    }
}
```

### Anti-Patterns to Avoid

- **Caching overview counts in model fields with manual invalidation:** The data set is tiny (personal todo list). Computing from the store on every View() is O(n) where n is typically <100. Caching adds complexity and a whole class of "stale data" bugs. Just query the store directly.
- **Creating a separate overview component with its own Model/Update/View:** The overview has no interactive behavior (no key handling, no cursor). It is a read-only display derived from store data. Making it a full Bubble Tea component is over-engineering.
- **Rendering the overview outside the calendar pane:** The requirement says "Calendar panel shows todo counts per month below the calendar grid." The overview belongs inside the calendar pane's View output, not as a third pane or app-level element.
- **Filtering overview to only show the current year:** Users may have todos in past or future years. Show all months that have todos regardless of year.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Month name formatting | Custom month name strings | `time.Month.String()` | Go's standard library already provides "January", "February", etc. |
| Chronological sorting of year-month pairs | Custom sort logic | `sort.Slice` with year then month comparison | Same pattern used in all existing store sort functions |
| Date parsing for grouping | Custom string splitting | `time.Parse(dateFormat, t.Date)` | Already used in `IncompleteTodosPerDay()` and `InMonth()` |

**Key insight:** The entire overview feature is pure rendering logic -- no new user interactions, no state mutations, no new keybindings. It reads data from the store and renders strings. This makes it the simplest phase in the project.

## Common Pitfalls

### Pitfall 1: Overview Not Updating After Todo Mutations
**What goes wrong:** User adds/deletes/completes a todo but the overview counts do not change.
**Why it happens:** If overview counts are cached in model fields and only refreshed on explicit events rather than computed fresh on each render.
**How to avoid:** Compute overview counts directly from the store in `View()` or `renderOverview()`. Since `View()` is called after every `Update()` cycle, the counts are always fresh. The store is the single source of truth.
**Warning signs:** Adding a todo does not change the count until navigating away and back.

### Pitfall 2: Overview Overflows the Calendar Pane Height
**What goes wrong:** If a user has todos in many different months (e.g., 20+ months), the overview list is longer than the pane height, causing content to be cut off or the layout to break.
**Why it happens:** The calendar pane has a fixed height determined by `contentHeight` in `app.View()`. The calendar grid takes ~8 lines. The overview must fit in the remaining space.
**How to avoid:** Calculate available height in the overview rendering. If there are more month entries than available lines, truncate with an indicator (e.g., "... and N more"). For a personal todo app, this is unlikely to be an issue, but defensive code is good practice.
**Warning signs:** Overview text extends beyond the pane border or gets clipped.

### Pitfall 3: No Month Entries Displayed When All Todos Are Floating
**What goes wrong:** User only has floating todos (no dates). The overview shows only "Unknown [12]" with no month rows. This is correct behavior but looks sparse.
**Why it happens:** `TodoCountsByMonth()` returns an empty slice when no todos have dates.
**How to avoid:** This is actually correct behavior. The overview should show only what exists. An empty month list with just the Unknown row is fine. No special handling needed.
**Warning signs:** None -- this is expected behavior.

### Pitfall 4: Calendar Pane Width Assumption
**What goes wrong:** Overview text is wider than the calendar pane inner width, causing wrapping or misalignment.
**Why it happens:** The calendar grid is 34 chars wide, and the pane inner width is 38 (including padding). If month labels plus counts exceed this width, text may wrap.
**How to avoid:** Use abbreviated month names or left-align within a fixed format width. `time.Month.String()` returns full names (max "September" = 9 chars). Format `" September 2026    [7]"` is ~24 chars, well within 34 chars. Safe.
**Warning signs:** Overview rows wrapping to the next line.

### Pitfall 5: Store Method Uses fmt Without Import
**What goes wrong:** The new `TodoCountsByMonth()` method uses `fmt.Sprintf` for the map key but `store.go` may not import `fmt`.
**Why it happens:** Easy to overlook imports when adding new methods.
**How to avoid:** Use a struct key or concatenated integer key instead of string formatting. For example, `year*100 + int(month)` as an integer key. This avoids the fmt import entirely.
**Warning signs:** Compilation error on missing import.

## Code Examples

### Store TodoCountsByMonth (Using Integer Key)
```go
// internal/store/store.go

// MonthCount represents the todo count for a specific year-month.
type MonthCount struct {
    Year  int
    Month time.Month
    Count int
}

// TodoCountsByMonth returns the number of todos in each year-month
// that has at least one todo, sorted chronologically.
func (s *Store) TodoCountsByMonth() []MonthCount {
    type ym struct{ y int; m time.Month }
    counts := make(map[ym]int)
    for _, t := range s.data.Todos {
        if t.Date == "" {
            continue
        }
        d, err := time.Parse(dateFormat, t.Date)
        if err != nil {
            continue
        }
        counts[ym{d.Year(), d.Month()}]++
    }
    result := make([]MonthCount, 0, len(counts))
    for k, c := range counts {
        result = append(result, MonthCount{Year: k.y, Month: k.m, Count: c})
    }
    sort.Slice(result, func(i, j int) bool {
        if result[i].Year != result[j].Year {
            return result[i].Year < result[j].Year
        }
        return result[i].Month < result[j].Month
    })
    return result
}
```

### Store FloatingTodoCount
```go
// internal/store/store.go

// FloatingTodoCount returns the number of todos with no date assigned.
func (s *Store) FloatingTodoCount() int {
    count := 0
    for _, t := range s.data.Todos {
        if !t.HasDate() {
            count++
        }
    }
    return count
}
```

### Calendar View with Overview
```go
// internal/calendar/model.go

func (m Model) View() string {
    todayDay := 0
    now := time.Now()
    if now.Year() == m.year && now.Month() == m.month {
        todayDay = now.Day()
    }

    grid := RenderGrid(m.year, m.month, todayDay, m.holidays, m.mondayStart, m.indicators, m.styles)
    overview := m.renderOverview()

    return grid + overview
}

func (m Model) renderOverview() string {
    var b strings.Builder

    b.WriteString("\n")
    b.WriteString(m.styles.OverviewHeader.Render("Overview"))
    b.WriteString("\n")

    monthCounts := m.store.TodoCountsByMonth()
    for _, mc := range monthCounts {
        label := mc.Month.String()
        if mc.Year != m.year {
            label = fmt.Sprintf("%s %d", label, mc.Year)
        }
        line := fmt.Sprintf(" %-16s[%d]", label, mc.Count)

        if mc.Year == m.year && mc.Month == m.month {
            b.WriteString(m.styles.OverviewActive.Render(line))
        } else {
            b.WriteString(m.styles.OverviewCount.Render(line))
        }
        b.WriteString("\n")
    }

    // Floating (undated) count
    floatingCount := m.store.FloatingTodoCount()
    line := fmt.Sprintf(" %-16s[%d]", "Unknown", floatingCount)
    b.WriteString(m.styles.OverviewCount.Render(line))
    b.WriteString("\n")

    return b.String()
}
```

### Calendar Styles Extension
```go
// internal/calendar/styles.go

type Styles struct {
    Header          lipgloss.Style
    WeekdayHdr      lipgloss.Style
    Normal          lipgloss.Style
    Today           lipgloss.Style
    Holiday         lipgloss.Style
    Indicator       lipgloss.Style
    OverviewHeader  lipgloss.Style
    OverviewCount   lipgloss.Style
    OverviewActive  lipgloss.Style
}

func NewStyles(t theme.Theme) Styles {
    return Styles{
        Header:         lipgloss.NewStyle().Bold(true).Foreground(t.HeaderFg),
        WeekdayHdr:     lipgloss.NewStyle().Foreground(t.WeekdayFg),
        Normal:         lipgloss.NewStyle().Foreground(t.NormalFg),
        Today:          lipgloss.NewStyle().Bold(true).Foreground(t.TodayFg).Background(t.TodayBg),
        Holiday:        lipgloss.NewStyle().Foreground(t.HolidayFg),
        Indicator:      lipgloss.NewStyle().Bold(true).Foreground(t.IndicatorFg),
        OverviewHeader: lipgloss.NewStyle().Bold(true).Foreground(t.AccentFg),
        OverviewCount:  lipgloss.NewStyle().Foreground(t.MutedFg),
        OverviewActive: lipgloss.NewStyle().Bold(true).Foreground(t.NormalFg),
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Calendar pane shows only grid | Calendar pane shows grid + overview panel | Phase 9 | At-a-glance view of where work is concentrated |
| No count aggregation queries | Store provides TodoCountsByMonth(), FloatingTodoCount() | Phase 9 | Centralized query logic, reusable |

**Deprecated/outdated:**
- Nothing deprecated. Phase 9 adds new query methods and extends the calendar view. All existing functionality is preserved.

## Open Questions

1. **Should the overview show only months for the currently viewed year, or all years?**
   - What we know: The requirement says "todo count per month." Users may have todos spanning multiple years.
   - Recommendation: Show all months that have todos, across all years. For months in the currently viewed year, show just the month name (e.g., "January"). For months in other years, include the year (e.g., "January 2025"). This keeps the common case compact.

2. **Should the currently viewed month be visually distinct in the overview?**
   - What we know: The overview lists all months. The user is currently viewing one specific month.
   - Recommendation: Yes, highlight the current month row with bold/normal foreground (OverviewActive style) while other months use muted foreground (OverviewCount style). This provides a visual anchor.

3. **Should the "Unknown" row always be shown even when there are zero floating todos?**
   - What we know: The requirement says "Overview shows count of undated (floating) todos."
   - Recommendation: Always show the Unknown row. Even "Unknown [0]" is informative -- it confirms there are no floating todos. Consistent presence is easier to scan than conditional presence.

4. **Height overflow with many months?**
   - What we know: The pane height is typically 20-30 lines. Grid takes ~8-9 lines. That leaves ~12-20 lines for overview.
   - Recommendation: For v1, render all months without truncation. A personal todo app is unlikely to have 20+ distinct months with todos. If this becomes an issue, add truncation in a future iteration.

## Sources

### Primary (HIGH confidence)
- Project source: `internal/store/store.go` -- existing query patterns (TodosForMonth, FloatingTodos, IncompleteTodosPerDay), sort patterns, Todo iteration patterns
- Project source: `internal/store/todo.go` -- Todo struct, HasDate(), InMonth(), dateFormat constant
- Project source: `internal/calendar/model.go` -- View() calling RenderGrid(), RefreshIndicators() pattern, store field access
- Project source: `internal/calendar/grid.go` -- RenderGrid() pure function, gridWidth constant (34 chars)
- Project source: `internal/calendar/styles.go` -- Styles struct pattern, NewStyles(theme.Theme) constructor
- Project source: `internal/app/model.go` -- calendarInnerWidth=38, contentHeight calculation, RefreshIndicators() call after every update
- Project source: `internal/theme/theme.go` -- AccentFg, MutedFg, NormalFg color roles available for overview styling

### Secondary (MEDIUM confidence)
- None needed -- all patterns verified in existing codebase

### Tertiary (LOW confidence)
- None -- all findings verified with primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- zero new dependencies; all changes are extensions of existing patterns
- Architecture: HIGH -- overview rendering follows exact same patterns as existing grid rendering and store queries
- Pitfalls: HIGH -- derived from direct code audit (pane width, height constraints, live update mechanism, store iteration)
- Code examples: HIGH -- all code follows exact conventions from existing codebase (Styles struct, NewStyles constructor, View() string building, store query methods)

**Research date:** 2026-02-06
**Valid until:** 2026-03-08 (30 days -- stable domain, all libraries at current versions, no external dependencies)
