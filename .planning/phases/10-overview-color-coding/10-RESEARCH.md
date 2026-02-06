# Phase 10: Overview Color Coding - Research

**Researched:** 2026-02-06
**Domain:** Bubble Tea theme-aware rendering, store aggregation queries, lipgloss styling
**Confidence:** HIGH

## Summary

Phase 10 enhances the existing overview panel (built in Phase 9) to split the single per-month todo count into two separate counts -- pending (incomplete) and completed -- and display them with distinct, theme-aware colors. The current overview renders a single `[N]` count per month using `OverviewCount` and `OverviewActive` styles. This phase modifies that to show two counts (e.g., pending and completed) with separate color roles that follow the active theme.

The requirements are OVCLR-01 (split count with pending=red, completed=green) and OVCLR-02 (colors follow the active theme, not hardcoded). The "red" and "green" from OVCLR-01 describe the semantic intent, while OVCLR-02 explicitly says these colors must come from the theme. This means adding two new color roles to the `Theme` struct (`OverviewPendingFg` and `OverviewCompletedFg`) and defining appropriate colors for each of the 4 themes (Dark, Light, Nord, Solarized).

The implementation touches 3 files: `internal/store/store.go` (modify `MonthCount` to include pending/completed split and update `TodoCountsByMonth()`), `internal/theme/theme.go` (add 2 new color roles), and `internal/calendar/styles.go` + `internal/calendar/model.go` (add 2 new styles and update `renderOverview()` formatting). The theme propagation path already works (`app.applyTheme()` -> `calendar.SetTheme()` -> `calendar.NewStyles()`) so the new colors will automatically update when themes change. No new dependencies are needed.

**Primary recommendation:** Add `PendingFg` and `CompletedCountFg` fields to `Theme`, add `OverviewPending` and `OverviewCompleted` styles to calendar `Styles`, modify `MonthCount` to carry separate pending/completed counts, and update `renderOverview()` to display both counts with their respective styles.

## Standard Stack

No new dependencies. All changes extend existing project code.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `internal/theme` | project | Add `PendingFg` and `CompletedCountFg` color roles | Follows semantic-role-per-purpose pattern established in Phase 6 |
| `internal/store` | project | Modify `MonthCount` and `TodoCountsByMonth()` to split pending/completed | Extends existing aggregation query pattern from Phase 9 |
| `internal/calendar` | project | Add overview styles and update rendering format | Extends overview rendering from Phase 9 |
| lipgloss | v1.1.0 | Style pending/completed counts with theme colors | Already used throughout all styles |

### Existing (unchanged)
| Library | Version | Purpose |
|---------|---------|---------|
| Bubble Tea | v1.3.10 | TUI framework |
| Bubbles | v0.21.1 | key.Binding, help.Model |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Two new theme roles (`PendingFg`, `CompletedCountFg`) | Reuse existing `HolidayFg` for red and `IndicatorFg` for green | Reusing couples unrelated UI elements; changing holiday color would unintentionally change overview pending color. Dedicated roles are cleaner and consistent with the semantic theme pattern. |
| Modifying `MonthCount` struct to carry pending+completed | Creating a separate `MonthStatusCount` return type | Extra type for minimal benefit. The existing `MonthCount` can simply gain two fields instead of one `Count`. Simpler, less API surface. |
| Showing `P:3 C:2` format | Showing `[3/2]` or `3+2` format | `P:N C:N` is explicit but verbose. A compact format like `3/2` (pending/completed) or colored count pairs is better for the narrow 34-char grid. Recommendation: use two adjacent styled numbers like `3 2` where 3 is pending-colored and 2 is completed-colored. |

## Architecture Patterns

### Files Modified
```
internal/
  theme/
    theme.go       # Add PendingFg, CompletedCountFg to Theme struct + all 4 themes
  store/
    store.go       # Modify MonthCount to have Pending+Completed, update TodoCountsByMonth()
  calendar/
    styles.go      # Add OverviewPending and OverviewCompleted styles
    model.go       # Update renderOverview() to display split counts with colors
```

### Pattern 1: New Semantic Theme Color Roles

**What:** Add two new color fields to the `Theme` struct: `PendingFg` (for uncompleted/pending todo counts) and `CompletedCountFg` (for completed todo counts in the overview). These are separate from the existing `CompletedFg` (which styles completed todo text with strikethrough in the todo list) because the overview uses a different visual treatment -- counts need to be clearly visible, not muted/strikethrough.

**When to use:** In theme definitions and calendar overview styles.

**Confidence:** HIGH (follows exact pattern of existing 14 theme color roles)

**Color choices per theme:**
- **Dark**: PendingFg = warm red (`#D75F5F`), CompletedCountFg = soft green (`#87AF87`) -- visible on dark backgrounds without being harsh
- **Light**: PendingFg = medium red (`#D70000`), CompletedCountFg = forest green (`#008700`) -- readable on light backgrounds
- **Nord**: PendingFg = nord11 aurora red (`#BF616A`), CompletedCountFg = nord14 aurora green (`#A3BE8C`) -- canonical Nord palette colors
- **Solarized**: PendingFg = solarized red (`#DC322F`), CompletedCountFg = solarized green (`#859900`) -- canonical Solarized palette colors

**Example:**
```go
// internal/theme/theme.go

type Theme struct {
    // ... existing 14 fields ...

    // Overview
    PendingFg        lipgloss.Color // uncompleted todo count in overview
    CompletedCountFg lipgloss.Color // completed todo count in overview
}
```

### Pattern 2: Split MonthCount in Store

**What:** Modify the existing `MonthCount` struct to carry `Pending` and `Completed` counts instead of a single `Count` field. Update `TodoCountsByMonth()` to populate both by checking `Todo.Done`. The `FloatingTodoCount()` method should similarly be split into pending and completed.

**When to use:** Called from `calendar.renderOverview()` on every render.

**Confidence:** HIGH (extends existing pattern, `Todo.Done` field already exists)

**Example:**
```go
// internal/store/store.go

type MonthCount struct {
    Year      int
    Month     time.Month
    Pending   int
    Completed int
}

func (s *Store) TodoCountsByMonth() []MonthCount {
    type ym struct{ y int; m time.Month }
    pending := make(map[ym]int)
    completed := make(map[ym]int)
    for _, t := range s.data.Todos {
        if t.Date == "" {
            continue
        }
        d, err := time.Parse(dateFormat, t.Date)
        if err != nil {
            continue
        }
        k := ym{d.Year(), d.Month()}
        if t.Done {
            completed[k]++
        } else {
            pending[k]++
        }
    }
    // Collect all keys from both maps
    allKeys := make(map[ym]bool)
    for k := range pending {
        allKeys[k] = true
    }
    for k := range completed {
        allKeys[k] = true
    }
    result := make([]MonthCount, 0, len(allKeys))
    for k := range allKeys {
        result = append(result, MonthCount{
            Year:      k.y,
            Month:     k.m,
            Pending:   pending[k],
            Completed: completed[k],
        })
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

Similarly for floating todos:
```go
type FloatingCount struct {
    Pending   int
    Completed int
}

func (s *Store) FloatingTodoCounts() FloatingCount {
    var fc FloatingCount
    for _, t := range s.data.Todos {
        if !t.HasDate() {
            if t.Done {
                fc.Completed++
            } else {
                fc.Pending++
            }
        }
    }
    return fc
}
```

### Pattern 3: Updated Overview Rendering with Color-Coded Counts

**What:** Update `renderOverview()` to display pending and completed counts side by side with different styles. The format should be compact to fit within the 34-char grid width. Each line shows: month label + pending count (styled with `OverviewPending`) + completed count (styled with `OverviewCompleted`).

**When to use:** In `calendar.Model.renderOverview()`.

**Confidence:** HIGH (straightforward extension of existing rendering)

**Format design:** The current format is `" January         [7]"` (24 chars). The new format needs to show two counts. Options:
- `" January         3  2"` -- pending then completed, colored (compact)
- `" January      P:3 C:2"` -- explicit labels (wider)

Recommendation: Use the compact colored format since the colors themselves distinguish meaning, and add a small legend in the overview header (e.g., render the "Overview" header with an inline hint).

**Example:**
```go
func (m Model) renderOverview() string {
    var b strings.Builder

    b.WriteString("\n")
    b.WriteString(m.styles.OverviewHeader.Render("Overview"))
    b.WriteString("\n")

    months := m.store.TodoCountsByMonth()
    for _, mc := range months {
        label := mc.Month.String()
        if mc.Year != m.year {
            label = fmt.Sprintf("%s %d", mc.Month.String(), mc.Year)
        }

        pending := m.styles.OverviewPending.Render(fmt.Sprintf("%d", mc.Pending))
        completed := m.styles.OverviewCompleted.Render(fmt.Sprintf("%d", mc.Completed))
        line := fmt.Sprintf(" %-16s%s  %s", label, pending, completed)

        if mc.Year == m.year && mc.Month == m.month {
            // For the active month, we still want the label to be bold/prominent
            styledLabel := m.styles.OverviewActive.Render(fmt.Sprintf(" %-16s", label))
            line = styledLabel + pending + "  " + completed
        }
        b.WriteString(line)
        b.WriteString("\n")
    }

    // Floating (undated) count
    fc := m.store.FloatingTodoCounts()
    pending := m.styles.OverviewPending.Render(fmt.Sprintf("%d", fc.Pending))
    completed := m.styles.OverviewCompleted.Render(fmt.Sprintf("%d", fc.Completed))
    label := m.styles.OverviewCount.Render(fmt.Sprintf(" %-16s", "Unknown"))
    b.WriteString(label + pending + "  " + completed)
    b.WriteString("\n")

    return b.String()
}
```

### Pattern 4: Calendar Styles Extension

**What:** Add `OverviewPending` and `OverviewCompleted` styles to the calendar `Styles` struct, wired from the new theme color roles.

**Confidence:** HIGH (exact pattern from existing styles)

**Example:**
```go
// internal/calendar/styles.go

type Styles struct {
    // ... existing 9 fields ...
    OverviewPending   lipgloss.Style
    OverviewCompleted lipgloss.Style
}

func NewStyles(t theme.Theme) Styles {
    return Styles{
        // ... existing styles ...
        OverviewPending:   lipgloss.NewStyle().Foreground(t.PendingFg),
        OverviewCompleted: lipgloss.NewStyle().Foreground(t.CompletedCountFg),
    }
}
```

### Anti-Patterns to Avoid

- **Hardcoding red/green ANSI colors in the overview rendering:** Violates OVCLR-02. Colors must come from the theme via style objects, never from inline color literals in the rendering code.
- **Reusing `CompletedFg` for overview completed count:** The existing `CompletedFg` is designed for strikethrough muted text in the todo list. Overview counts need to be clearly visible (green-ish), not muted. Use a distinct theme role.
- **Reusing `HolidayFg` for pending count:** Couples unrelated concerns. Changing the holiday color would break the overview. Use dedicated roles.
- **Adding a legend/key that takes up vertical space:** The calendar pane has limited height. A multi-line legend wastes space. If a legend is needed, keep it on the same line as the header (e.g., "Overview (P C)").
- **Changing the `MonthCount` struct name or creating a parallel struct:** Keep the existing `MonthCount` name and extend it. The `calendar.Model.renderOverview()` already uses it. Changing the name or creating a parallel type causes unnecessary churn.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Theme-aware color propagation | Manual color passing through function args | Theme struct -> Styles struct -> lipgloss.Style | The Styles constructor DI pattern already handles this for all 4 themes. Just add fields. |
| Color values for Nord/Solarized palettes | Guessing hex values | Official palette specifications | Nord: nordtheme.com, Solarized: ethanschoonover.com/solarized. The existing theme.go already uses canonical values. |
| Pending/completed classification | Custom status tracking | `Todo.Done` boolean already in store | The Done field is already the single source of truth for completion status. |

**Key insight:** This phase is primarily a theme and rendering change. The only logic change is splitting the count in `TodoCountsByMonth()` by `Done` status. Everything else is adding color roles and updating format strings.

## Common Pitfalls

### Pitfall 1: Styled Text Breaks Column Alignment
**What goes wrong:** When you apply lipgloss styles to individual numbers within a `fmt.Sprintf` format string, the ANSI escape codes add invisible characters that break `%-16s` alignment.
**Why it happens:** `fmt.Sprintf("%-16s%s", label, styledNumber)` counts ANSI escapes as visible characters when the label is pre-styled, or the styled number appears shorter/longer than expected.
**How to avoid:** Apply styles AFTER formatting, or format the label padding separately from the styled counts. Build the line in parts: (1) format and style the label to a fixed width, (2) append styled pending count, (3) append styled completed count. Do not mix `fmt.Sprintf` width specifiers with pre-styled strings.
**Warning signs:** Overview rows are misaligned -- labels shift left/right depending on count digits.

### Pitfall 2: Overview Colors Do Not Update on Theme Switch
**What goes wrong:** User switches theme in settings, but overview pending/completed colors remain the old theme's colors.
**Why it happens:** Forgetting to add the new theme roles to one of the 4 theme definitions, or forgetting to wire them through `NewStyles()`.
**How to avoid:** Verify all 4 theme functions (Dark, Light, Nord, Solarized) define the new fields. Verify `calendar.NewStyles()` wires the new theme fields to style objects. The existing `calendar.SetTheme()` calls `NewStyles()` so propagation is automatic -- but only if the fields are wired.
**Warning signs:** Colors work in one theme but are invisible (empty string default) in another.

### Pitfall 3: Zero Counts Display Awkwardly
**What goes wrong:** A month with 5 pending and 0 completed shows `5  0` with a green-colored "0". This is visually noisy.
**Why it happens:** Always rendering both counts regardless of value.
**How to avoid:** Two approaches: (1) always show both -- simpler, consistent, the "0" confirms no items. (2) only show non-zero counts -- cleaner but inconsistent column width. Recommendation: always show both for consistency and scannability. A "0" in green next to a "3" in red tells the user "3 pending, none completed" at a glance.
**Warning signs:** User confusion about what the numbers mean when one is 0.

### Pitfall 4: Dark Theme Colors Invisible on Dark Backgrounds
**What goes wrong:** Chosen red/green colors are too dark to see on a dark terminal background.
**Why it happens:** Using pure red (#FF0000) or dark red (#AF0000) that blends into dark backgrounds.
**How to avoid:** Use lighter/warmer variants for the dark theme: `#D75F5F` (warm rose) for pending, `#87AF87` (soft sage) for completed. Test visually. The existing Dark theme uses `#AF0000` for holidays which is already somewhat dark -- the overview pending should be lighter for better visibility of small count numbers.
**Warning signs:** Count numbers are barely visible in the Dark theme.

### Pitfall 5: FloatingTodoCount Return Type Change Breaks Existing Callers
**What goes wrong:** `FloatingTodoCount()` currently returns `int`. If renamed/changed to return a struct, callers break.
**Why it happens:** Changing the method signature without updating all call sites.
**How to avoid:** Add a NEW method `FloatingTodoCounts()` (plural) that returns the split struct. Keep `FloatingTodoCount()` as-is if it is used elsewhere, or rename/replace if the only caller is `renderOverview()`. Check all callers first.
**Warning signs:** Compilation errors after changing the method.

## Code Examples

### Theme with New Color Roles (All 4 Themes)
```go
// internal/theme/theme.go

type Theme struct {
    // Panel borders
    BorderFocused   lipgloss.Color
    BorderUnfocused lipgloss.Color

    // Calendar
    HeaderFg    lipgloss.Color
    WeekdayFg   lipgloss.Color
    TodayFg     lipgloss.Color
    TodayBg     lipgloss.Color
    HolidayFg   lipgloss.Color
    IndicatorFg lipgloss.Color

    // Todo list
    AccentFg    lipgloss.Color
    MutedFg     lipgloss.Color
    CompletedFg lipgloss.Color
    EmptyFg     lipgloss.Color

    // General
    NormalFg lipgloss.Color
    NormalBg lipgloss.Color

    // Overview counts (NEW)
    PendingFg        lipgloss.Color // pending todo count in overview
    CompletedCountFg lipgloss.Color // completed todo count in overview
}

func Dark() Theme {
    return Theme{
        // ... existing fields unchanged ...
        PendingFg:        lipgloss.Color("#D75F5F"), // warm rose - visible on dark bg
        CompletedCountFg: lipgloss.Color("#87AF87"), // soft sage green
    }
}

func Light() Theme {
    return Theme{
        // ... existing fields unchanged ...
        PendingFg:        lipgloss.Color("#D70000"), // medium red
        CompletedCountFg: lipgloss.Color("#008700"), // forest green
    }
}

func Nord() Theme {
    return Theme{
        // ... existing fields unchanged ...
        PendingFg:        lipgloss.Color("#BF616A"), // nord11 aurora red
        CompletedCountFg: lipgloss.Color("#A3BE8C"), // nord14 aurora green
    }
}

func Solarized() Theme {
    return Theme{
        // ... existing fields unchanged ...
        PendingFg:        lipgloss.Color("#DC322F"), // solarized red
        CompletedCountFg: lipgloss.Color("#859900"), // solarized green
    }
}
```

### Modified Store MonthCount with Pending/Completed Split
```go
// internal/store/store.go

type MonthCount struct {
    Year      int
    Month     time.Month
    Pending   int
    Completed int
}

func (s *Store) TodoCountsByMonth() []MonthCount {
    type ym struct {
        y int
        m time.Month
    }
    counts := make(map[ym]*MonthCount)
    for _, t := range s.data.Todos {
        if t.Date == "" {
            continue
        }
        d, err := time.Parse(dateFormat, t.Date)
        if err != nil {
            continue
        }
        k := ym{d.Year(), d.Month()}
        if counts[k] == nil {
            counts[k] = &MonthCount{Year: k.y, Month: k.m}
        }
        if t.Done {
            counts[k].Completed++
        } else {
            counts[k].Pending++
        }
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
```

### Modified FloatingTodoCounts
```go
// internal/store/store.go

type FloatingCount struct {
    Pending   int
    Completed int
}

func (s *Store) FloatingTodoCounts() FloatingCount {
    var fc FloatingCount
    for _, t := range s.data.Todos {
        if !t.HasDate() {
            if t.Done {
                fc.Completed++
            } else {
                fc.Pending++
            }
        }
    }
    return fc
}
```

### Updated Calendar Styles
```go
// internal/calendar/styles.go

type Styles struct {
    Header            lipgloss.Style
    WeekdayHdr        lipgloss.Style
    Normal            lipgloss.Style
    Today             lipgloss.Style
    Holiday           lipgloss.Style
    Indicator         lipgloss.Style
    OverviewHeader    lipgloss.Style
    OverviewCount     lipgloss.Style
    OverviewActive    lipgloss.Style
    OverviewPending   lipgloss.Style  // NEW
    OverviewCompleted lipgloss.Style  // NEW
}

func NewStyles(t theme.Theme) Styles {
    return Styles{
        // ... existing 9 styles unchanged ...
        OverviewPending:   lipgloss.NewStyle().Foreground(t.PendingFg),
        OverviewCompleted: lipgloss.NewStyle().Foreground(t.CompletedCountFg),
    }
}
```

### Updated renderOverview
```go
// internal/calendar/model.go

func (m Model) renderOverview() string {
    var b strings.Builder

    b.WriteString("\n")
    b.WriteString(m.styles.OverviewHeader.Render("Overview"))
    b.WriteString("\n")

    months := m.store.TodoCountsByMonth()
    for _, mc := range months {
        label := mc.Month.String()
        if mc.Year != m.year {
            label = fmt.Sprintf("%s %d", mc.Month.String(), mc.Year)
        }

        // Format: " Label           P  C"
        // Label is styled (active or muted), P is pending-colored, C is completed-colored
        paddedLabel := fmt.Sprintf(" %-16s", label)
        pending := m.styles.OverviewPending.Render(fmt.Sprintf("%d", mc.Pending))
        completed := m.styles.OverviewCompleted.Render(fmt.Sprintf("%d", mc.Completed))

        if mc.Year == m.year && mc.Month == m.month {
            b.WriteString(m.styles.OverviewActive.Render(paddedLabel))
        } else {
            b.WriteString(m.styles.OverviewCount.Render(paddedLabel))
        }
        b.WriteString(pending)
        b.WriteString("  ")
        b.WriteString(completed)
        b.WriteString("\n")
    }

    // Floating (undated) count
    fc := m.store.FloatingTodoCounts()
    paddedLabel := fmt.Sprintf(" %-16s", "Unknown")
    b.WriteString(m.styles.OverviewCount.Render(paddedLabel))
    b.WriteString(m.styles.OverviewPending.Render(fmt.Sprintf("%d", fc.Pending)))
    b.WriteString("  ")
    b.WriteString(m.styles.OverviewCompleted.Render(fmt.Sprintf("%d", fc.Completed)))
    b.WriteString("\n")

    return b.String()
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single total count `[N]` per month in overview | Split pending + completed counts with theme-aware colors | Phase 10 | Users see progress at a glance -- how much is done vs. remaining |
| 14 semantic color roles in Theme | 16 semantic color roles (added PendingFg, CompletedCountFg) | Phase 10 | Overview has dedicated colors independent of other UI elements |
| `MonthCount.Count` single field | `MonthCount.Pending` + `MonthCount.Completed` fields | Phase 10 | Store provides richer aggregation data |
| `FloatingTodoCount()` returns int | `FloatingTodoCounts()` returns `FloatingCount{Pending, Completed}` | Phase 10 | Floating todos also show completion breakdown |

**Deprecated/outdated:**
- `MonthCount.Count` field: Replaced by `Pending` + `Completed`. No backward compatibility concern since this struct is internal only and `renderOverview()` is the sole consumer.
- `FloatingTodoCount()`: Replaced by `FloatingTodoCounts()`. Remove the old method to avoid dead code.

## Open Questions

1. **Should the overview show a column header for pending/completed?**
   - What we know: The grid is 34 chars wide. Adding column headers like "P  C" above the numbers would clarify meaning but cost one vertical line.
   - Recommendation: Add a subtle inline hint on the Overview header line, e.g., render "Overview" followed by a muted right-aligned hint. Or simply rely on color convention (red=pending, green=done) which is universally understood. Start without a legend; add one if user feedback suggests confusion.

2. **Display format when a month has zero pending or zero completed?**
   - What we know: `0` in red or green is technically correct but could look odd.
   - Recommendation: Always show both numbers for consistency. `0` is informative -- it means "all done" or "nothing started." Hiding zeros would create inconsistent row widths.

3. **Should the old `FloatingTodoCount()` be removed or kept alongside `FloatingTodoCounts()`?**
   - What we know: The only caller of `FloatingTodoCount()` is `calendar.renderOverview()`. No other code references it.
   - Recommendation: Replace it entirely with `FloatingTodoCounts()`. If the old signature is needed elsewhere in the future, it can be trivially re-derived from the struct.

## Sources

### Primary (HIGH confidence)
- Project source: `internal/theme/theme.go` -- 14 existing color roles, 4 theme definitions with canonical palette colors
- Project source: `internal/store/store.go` -- `MonthCount` struct, `TodoCountsByMonth()`, `FloatingTodoCount()`, `Todo.Done` field usage
- Project source: `internal/store/todo.go` -- `Todo.Done` boolean field, `HasDate()` method
- Project source: `internal/calendar/model.go` -- `renderOverview()` current implementation, `SetTheme()` propagation
- Project source: `internal/calendar/styles.go` -- `Styles` struct with 9 fields, `NewStyles(theme.Theme)` constructor pattern
- Project source: `internal/app/model.go` -- `applyTheme()` method that propagates to all sub-models
- Nord theme palette: https://www.nordtheme.com (canonical colors already used in theme.go)
- Solarized palette: https://ethanschoonover.com/solarized (canonical colors already used in theme.go)

### Secondary (MEDIUM confidence)
- None needed -- all patterns verified in existing codebase

### Tertiary (LOW confidence)
- Dark theme color choices (`#D75F5F`, `#87AF87`): Selected for visibility on dark backgrounds based on ANSI 256-color conventions. Should be visually verified during implementation.
- Light theme color choices (`#D70000`, `#008700`): Standard terminal red/green. Should be visually verified during implementation.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- zero new dependencies; all changes are field additions to existing structs
- Architecture: HIGH -- extends exact patterns from Phase 9 (theme roles, Styles struct, store queries, renderOverview)
- Pitfalls: HIGH -- derived from direct code audit (ANSI alignment, theme propagation path, method signature changes, color visibility)
- Code examples: HIGH -- all code follows exact conventions from existing codebase

**Research date:** 2026-02-06
**Valid until:** 2026-03-08 (30 days -- stable domain, no external dependencies, all libraries at current versions)
