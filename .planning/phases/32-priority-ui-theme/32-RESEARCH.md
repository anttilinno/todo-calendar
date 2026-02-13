# Phase 32: Priority UI + Theme - Research

**Researched:** 2026-02-13
**Domain:** Bubble Tea TUI rendering, Lip Gloss styling, theme color extension, form field wiring
**Confidence:** HIGH

## Summary

Phase 32 implements the user-facing priority system on top of Phase 31's data layer. It spans seven requirements (PRIO-01 through PRIO-07) across five packages: `theme` (priority color definitions), `todolist` (edit form wiring + badge rendering), `search` (badge rendering), `calendar` (priority-aware day indicators), and `store` (new query method for highest-priority-per-day).

The codebase has a mature, consistent architecture. The `Theme` struct in `internal/theme/theme.go` defines semantic color roles, and each component's `Styles` struct maps those to `lipgloss.Style` values. Adding priority colors means extending `Theme` with 4 new color fields (one per priority level), defining appropriate hex values for all 4 themes (Dark, Light, Nord, Solarized), and creating new `lipgloss.Style` entries in each component's `Styles` struct. The edit form in `todolist/model.go` follows a field-cycling pattern (Tab cycles through numbered editFields) that needs one additional field for priority selection. The calendar grid needs a new store method (`HighestPriorityPerDay`) to know which priority color to use for each day's indicator bracket.

The most complex requirement is PRIO-06 (calendar indicators reflecting priority color). Currently, the calendar uses `IncompleteTodosPerDay` (returns count per day) and a fixed `Indicator` style. This must change to use priority-aware styling: each day cell chooses its bracket color based on the highest-priority incomplete todo on that day. This requires both a new store query and modifications to the grid rendering logic.

**Primary recommendation:** Extend `Theme` with `PriorityP1Fg` through `PriorityP4Fg` color fields, add a `PriorityStyle(level int)` helper method to each Styles struct (or a shared utility), wire the priority field into the edit form as `editField=3` (shifting template to 4 in inputMode), and add `HighestPriorityPerDay` to the store interface.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/charmbracelet/lipgloss` | v1.1.1-0.20250404 | Terminal styling (colors, bold, strikethrough) | Already in use for all rendering |
| `github.com/charmbracelet/bubbles` | v0.21.1 | TUI components (textinput, textarea) | Already in use for form fields |
| `github.com/charmbracelet/bubbletea` | v1.3.10 | TUI framework (Model/Update/View) | Already in use as app framework |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `modernc.org/sqlite` | v1.44.3 | Database backend | New query for HighestPriorityPerDay |
| `database/sql` | stdlib | SQL interface | New query implementation |

No new dependencies are needed. All work uses existing libraries.

## Architecture Patterns

### Recommended Change Structure
```
internal/
  theme/theme.go          # Add PriorityP1Fg..PriorityP4Fg to Theme struct + all 4 themes
  store/iface.go          # Add HighestPriorityPerDay method
  store/sqlite.go         # Implement HighestPriorityPerDay query
  store/sqlite_test.go    # Test HighestPriorityPerDay
  todolist/styles.go      # Add PriorityBadge[1-4] styles, CompletedBadge[1-4] styles
  todolist/model.go       # Add priority editField, wire saveAdd/saveEdit, render badges
  search/styles.go        # Add PriorityBadge[1-4] styles
  search/model.go         # Render priority badges in search results
  calendar/styles.go      # Add IndicatorP[1-4] and TodayIndicatorP[1-4] styles
  calendar/grid.go        # Use HighestPriorityPerDay for indicator coloring
  calendar/model.go       # Store/refresh priority-per-day data
```

### Pattern 1: Theme Color Extension
**What:** Add semantic color roles to the Theme struct, define values for all 4 themes.
**When to use:** Any time a new visual element needs theme-aware coloring.
**Example:**
```go
// In internal/theme/theme.go
type Theme struct {
    // ... existing fields ...

    // Priority level colors
    PriorityP1Fg lipgloss.Color // P1 (urgent/critical) -- red family
    PriorityP2Fg lipgloss.Color // P2 (high) -- orange family
    PriorityP3Fg lipgloss.Color // P3 (medium) -- blue family
    PriorityP4Fg lipgloss.Color // P4 (low) -- grey/muted family
}
```

### Pattern 2: Fixed-Width Badge Slot (PRIO-04)
**What:** Every todo line reserves a fixed-width slot for the priority badge, whether or not the todo has a priority. This ensures text after the badge aligns into columns.
**When to use:** Rendering todo lines in both todolist and search.
**Example:**
```go
// Badge rendering with fixed-width slot
const badgeWidth = 5 // "[P1] " = 4 chars + 1 space, "     " = 5 chars for no priority

func renderBadge(t *store.Todo, styles Styles) string {
    if !t.HasPriority() {
        return strings.Repeat(" ", badgeWidth) // fixed-width empty slot
    }
    label := fmt.Sprintf("[%s]", t.PriorityLabel()) // "[P1]" = 4 chars
    style := styles.priorityStyle(t.Priority)
    return style.Render(label) + " "
}
```

### Pattern 3: Edit Form Field Cycling
**What:** The edit form uses `editField` int to track which field is focused. Tab cycles through fields sequentially.
**Current cycle (editMode):** title(0) -> date(1) -> body(2) -> title(0)
**Current cycle (inputMode):** title(0) -> date(1) -> body(2) -> template(3) -> title(0)
**New cycle (editMode):** title(0) -> date(1) -> priority(2) -> body(3) -> title(0)
**New cycle (inputMode):** title(0) -> date(1) -> priority(2) -> body(3) -> template(4) -> title(0)

Note: Priority field is inserted between date and body because it is a quick single-value selection (not a freeform text area). The priority "field" can be rendered as a simple selector: `Priority: [none] P1 P2 P3 P4` where left/right arrows change the selection. This avoids a textinput entirely.

### Pattern 4: Calendar Priority-Aware Indicators
**What:** Calendar day cells `[dd]` currently use a single Indicator style (for pending). PRIO-06 requires the bracket color to reflect the highest-priority incomplete todo.
**How:** Replace the single `Indicator` style with priority-graded styles, falling back to the existing `Indicator` style when no prioritized todo exists on that day.

```go
// Priority cascade for day cell styling:
// 1. If day has incomplete P1 todo -> IndicatorP1 style
// 2. If day has incomplete P2 todo -> IndicatorP2 style
// 3. If day has incomplete P3 todo -> IndicatorP3 style
// 4. If day has incomplete P4 todo -> IndicatorP4 style
// 5. If day has incomplete todos (no priority) -> existing Indicator style (default)
// 6. If day has all-done todos -> existing IndicatorDone style
// 7. No todos -> Normal style
```

### Pattern 5: Completed Priority Rendering (PRIO-03)
**What:** Completed todos with priority show the colored badge but grey strikethrough text.
**How:** The badge and the text use DIFFERENT styles. Badge keeps its priority color. Text uses the existing `Completed` style (strikethrough + CompletedFg).

```go
// Completed todo with priority:
//   [P2] ~~Buy groceries~~
//   ^^^^              ^^^^
//   orange badge      grey strikethrough text
```

### Anti-Patterns to Avoid
- **Dynamic style creation per render:** Do NOT create `lipgloss.NewStyle()` in the View() method. Create styles once in `NewStyles()` and store them in the Styles struct. The existing codebase correctly follows this pattern.
- **Hardcoded hex colors in rendering code:** All colors must flow through `theme.Theme`. Never use `lipgloss.Color("#FF0000")` directly in rendering code.
- **Auto-sorting by priority:** The requirements explicitly state "priority is visual-only, no auto-sort". Do NOT change any SQL ORDER BY clauses.
- **Variable-width badge slots:** If P1 gets `[P1] ` but no-priority gets `""`, the text column will misalign. Always use fixed-width slots.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Terminal color output | ANSI escape sequences | `lipgloss.NewStyle().Foreground()` | Already used everywhere, handles terminal compatibility |
| Fixed-width text alignment | Manual space counting | `fmt.Sprintf("%-5s", badge)` or constant-width string | Simpler, less error-prone |
| Priority color lookup | Switch statement in every render function | Priority style array/method on Styles struct | DRY, single place to change |
| Priority-per-day SQL | Fetching all todos then computing in Go | SQL `MIN(priority)` with WHERE filter | More efficient, less data transfer |

**Key insight:** The priority badge is just a styled string prefix. The complexity is in ensuring consistent width and correct color selection, not in any novel rendering technique.

## Common Pitfalls

### Pitfall 1: Badge Width Inconsistency Between Priority and No-Priority
**What goes wrong:** Todos with priority get a 5-char prefix (`[P1] `) but no-priority todos get 0 chars, causing the checkbox column to misalign.
**Why it happens:** Easy to forget the empty slot when a todo has no priority.
**How to avoid:** Define a constant `badgeWidth = 5` and always emit exactly that many characters. For no-priority todos, emit 5 spaces.
**Warning signs:** Visual: todo text column is jagged when mixing prioritized and non-prioritized todos.

### Pitfall 2: saveEdit Resets Priority to 0
**What goes wrong:** When editing a P2 todo, the priority field is not populated from the existing todo, so saving resets it to 0.
**Why it happens:** Phase 31 hardcoded `0` in `saveEdit()`. Phase 32 must populate and pass the current priority.
**How to avoid:** When entering editMode (in `updateNormalMode`, case `key.Matches(msg, m.keys.Edit)`), read `fresh.Priority` and store it in a new `editPriority int` field on the Model. Pass `m.editPriority` to `store.Update()`.
**Warning signs:** Editing any field of a prioritized todo silently removes its priority.

### Pitfall 3: Calendar Indicator Priority Query Performance
**What goes wrong:** Fetching highest priority per day requires a new SQL query. If done naively (one query per day), it is 31 queries per month.
**Why it happens:** The existing `IncompleteTodosPerDay` uses a GROUP BY query that returns all days in one shot. The new priority query must follow the same pattern.
**How to avoid:** Use a single SQL query: `SELECT day, MIN(CASE WHEN priority > 0 THEN priority ELSE 999 END) FROM todos WHERE done = 0 AND ... GROUP BY day`. MIN works because P1 (1) is higher priority than P4 (4) -- lower number = higher priority.
**Warning signs:** Slow calendar rendering, especially when navigating months.

### Pitfall 4: Theme Colors Not Accessible on All Backgrounds
**What goes wrong:** Priority colors that look great on dark backgrounds are unreadable on light backgrounds (and vice versa).
**Why it happens:** Using the same hex values for all themes.
**How to avoid:** Define priority colors per-theme. The Dark and Nord themes need lighter/brighter colors; Light theme needs darker colors. Test by visually inspecting each theme.
**Warning signs:** Colors appear invisible or unreadable when switching themes.

### Pitfall 5: editField Numbering Conflicts After Insertion
**What goes wrong:** Inserting priority as editField=2 shifts body to 3 and template to 4, but existing code has many hardcoded checks like `if m.editField == 2` (body textarea handling).
**Why it happens:** The editField values are used as magic numbers throughout the Update logic.
**How to avoid:** Systematically update ALL editField references. Consider defining named constants (e.g., `fieldTitle = 0`, `fieldDate = 1`, `fieldPriority = 2`, `fieldBody = 3`, `fieldTemplate = 4`) to make the code self-documenting and reduce errors during renumbering.
**Warning signs:** Pressing Tab skips a field, or the body textarea captures Enter when priority field is focused.

### Pitfall 6: Search Results Not Using Same Badge Rendering
**What goes wrong:** Todo list shows `[P1] Buy groceries` but search shows `Buy groceries` without badge, violating PRIO-07.
**Why it happens:** Search has its own View() method in `search/model.go` that builds result lines independently.
**How to avoid:** Ensure the search View() method renders priority badges using the same fixed-width slot logic as the todo list.
**Warning signs:** Visual: priority badges appear in todo list but not in search results.

## Code Examples

### Theme Color Definitions (All 4 Themes)

```go
// In internal/theme/theme.go -- extend Theme struct
type Theme struct {
    // ... existing fields ...

    // Priority colors (foreground for badge text)
    PriorityP1Fg lipgloss.Color // P1 = red/critical
    PriorityP2Fg lipgloss.Color // P2 = orange/high
    PriorityP3Fg lipgloss.Color // P3 = blue/medium
    PriorityP4Fg lipgloss.Color // P4 = grey/low
}

// Dark theme priority colors
func Dark() Theme {
    return Theme{
        // ... existing ...
        PriorityP1Fg: lipgloss.Color("#FF5F5F"), // bright red
        PriorityP2Fg: lipgloss.Color("#FFAF5F"), // orange
        PriorityP3Fg: lipgloss.Color("#5F87FF"), // blue
        PriorityP4Fg: lipgloss.Color("#808080"), // grey
    }
}

// Light theme priority colors
func Light() Theme {
    return Theme{
        // ... existing ...
        PriorityP1Fg: lipgloss.Color("#D70000"), // dark red
        PriorityP2Fg: lipgloss.Color("#AF5F00"), // dark orange
        PriorityP3Fg: lipgloss.Color("#005FAF"), // dark blue
        PriorityP4Fg: lipgloss.Color("#8A8A8A"), // medium grey
    }
}

// Nord theme priority colors
func Nord() Theme {
    return Theme{
        // ... existing ...
        PriorityP1Fg: lipgloss.Color("#BF616A"), // nord11 aurora red
        PriorityP2Fg: lipgloss.Color("#D08770"), // nord12 aurora orange
        PriorityP3Fg: lipgloss.Color("#5E81AC"), // nord10 frost blue
        PriorityP4Fg: lipgloss.Color("#4C566A"), // nord3 polar night
    }
}

// Solarized theme priority colors
func Solarized() Theme {
    return Theme{
        // ... existing ...
        PriorityP1Fg: lipgloss.Color("#DC322F"), // solarized red
        PriorityP2Fg: lipgloss.Color("#CB4B16"), // solarized orange
        PriorityP3Fg: lipgloss.Color("#268BD2"), // solarized blue
        PriorityP4Fg: lipgloss.Color("#586E75"), // solarized base01
    }
}
```

### Store Method: HighestPriorityPerDay

```go
// In internal/store/iface.go -- add to TodoStore interface
HighestPriorityPerDay(year int, month time.Month) map[int]int

// In internal/store/sqlite.go
func (s *SQLiteStore) HighestPriorityPerDay(year int, month time.Month) map[int]int {
    start := fmt.Sprintf("%04d-%02d-01", year, month)
    end := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Format(dateFormat)

    rows, err := s.db.Query(
        `SELECT CAST(substr(date, 9, 2) AS INTEGER) AS day,
                MIN(CASE WHEN priority BETWEEN 1 AND 4 THEN priority ELSE NULL END)
         FROM todos
         WHERE done = 0 AND date >= ? AND date <= ? AND date_precision = 'day' AND priority BETWEEN 1 AND 4
         GROUP BY day`,
        start, end,
    )
    if err != nil {
        return nil
    }
    defer rows.Close()

    priorities := make(map[int]int)
    for rows.Next() {
        var day, prio int
        if err := rows.Scan(&day, &prio); err == nil {
            priorities[day] = prio
        }
    }
    return priorities
}
```

### Priority Badge Rendering in Todo List

```go
// In internal/todolist/model.go -- updated renderTodo
func (m Model) renderTodo(b *strings.Builder, t *store.Todo, selected bool) {
    // Cursor indicator
    if selected {
        b.WriteString(m.styles.Cursor.Render("> "))
    } else {
        b.WriteString("  ")
    }

    // Priority badge -- fixed width slot (PRIO-04)
    if t.HasPriority() {
        label := fmt.Sprintf("[%s]", t.PriorityLabel())
        if t.Done {
            // PRIO-03: completed todos keep colored badge
            style := m.styles.priorityBadgeStyle(t.Priority)
            b.WriteString(style.Render(label))
        } else {
            style := m.styles.priorityBadgeStyle(t.Priority)
            b.WriteString(style.Render(label))
        }
        b.WriteString(" ")
    } else {
        b.WriteString("     ") // 5 chars: same width as "[P1] "
    }

    // Styled checkbox (VIS-03)
    if t.Done {
        b.WriteString(m.styles.CheckboxDone.Render("[x]"))
    } else {
        b.WriteString(m.styles.Checkbox.Render("[ ]"))
    }
    b.WriteString(" ")

    // Text content
    text := t.Text
    if t.Done {
        text = m.styles.Completed.Render(text) // PRIO-03: grey strikethrough
    }
    b.WriteString(text)

    // ... body indicator, recurring indicator, date as before ...
    b.WriteString("\n")
}
```

### Priority Field in Edit Form

```go
// Model field additions
type Model struct {
    // ... existing fields ...
    editPriority int // 0-4, current priority being edited
}

// Priority selector rendering (no textinput needed)
func (m Model) renderPrioritySelector() string {
    options := []string{"none", "P1", "P2", "P3", "P4"}
    var parts []string
    for i, opt := range options {
        if i == m.editPriority {
            parts = append(parts, m.styles.Cursor.Render("["+opt+"]"))
        } else {
            parts = append(parts, " "+opt+" ")
        }
    }
    return strings.Join(parts, " ")
}

// In updateEditMode/updateInputMode, handle priority field:
// Left/Right arrows change m.editPriority (clamped to 0-4)
// Tab advances to next field
```

### Calendar Priority-Aware Grid Cell

```go
// In calendar/grid.go RenderGrid -- replace indicator style selection
priorities := st.HighestPriorityPerDay(year, month)

// In the day cell styling switch:
case hasPending:
    prio := priorities[day]
    switch prio {
    case 1:
        cell = s.IndicatorP1.Render(cell)
    case 2:
        cell = s.IndicatorP2.Render(cell)
    case 3:
        cell = s.IndicatorP3.Render(cell)
    case 4:
        cell = s.IndicatorP4.Render(cell)
    default:
        cell = s.Indicator.Render(cell) // no-priority pending: use default
    }
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `store.Add(..., 0)` hardcoded | `store.Add(..., m.editPriority)` from form | This phase | Users can set priority |
| `store.Update(..., 0)` hardcoded | `store.Update(..., m.editPriority)` from form | This phase | Editing preserves priority |
| Single `Indicator` style for pending | Priority-graded indicator styles (P1-P4 + default) | This phase | Calendar reflects priority |
| No badge prefix on todo lines | Fixed-width `[P1]`-`[P4]` badge slot | This phase | Visual priority identification |

**Deprecated/outdated:**
- The hardcoded `0` priority values in `saveAdd()` and `saveEdit()` from Phase 31 will be replaced with the actual form value.

## Impact Analysis

### Files Modified

| File | Change | Scope |
|------|--------|-------|
| `internal/theme/theme.go` | Add 4 `PriorityP*Fg` fields to Theme struct, define in all 4 themes | Small -- 4 fields x 4 themes |
| `internal/store/iface.go` | Add `HighestPriorityPerDay` method | Small -- 1 method |
| `internal/store/sqlite.go` | Implement `HighestPriorityPerDay` | Small -- ~20 lines |
| `internal/store/sqlite_test.go` | Test `HighestPriorityPerDay` | Small -- ~30 lines |
| `internal/todolist/styles.go` | Add priority badge styles (4 levels) | Small -- 4 styles |
| `internal/todolist/model.go` | Priority editField, editPriority state, renderTodo badge, saveAdd/saveEdit wiring | Large -- multiple touchpoints |
| `internal/search/styles.go` | Add priority badge styles (4 levels) | Small -- 4 styles |
| `internal/search/model.go` | Render priority badges in results | Small -- ~15 lines |
| `internal/calendar/styles.go` | Add IndicatorP1-P4 and TodayIndicatorP1-P4 styles | Medium -- 8 styles |
| `internal/calendar/grid.go` | Use priority-per-day data for indicator coloring | Medium -- both RenderGrid and RenderWeekGrid |
| `internal/calendar/model.go` | Store and refresh priority-per-day data alongside indicators | Small -- mirror existing pattern |
| `internal/recurring/generate_test.go` | Add `HighestPriorityPerDay` to fakeStore | Small -- 1 method stub |

### Files NOT Modified

| File | Why Not |
|------|---------|
| `internal/store/todo.go` | `HasPriority()` and `PriorityLabel()` already added in Phase 31 |
| `internal/app/model.go` | No structural changes needed -- priority flows through existing edit/refresh paths |
| `internal/config/config.go` | No priority-related config (PRIO-10 deferred to future) |

## Open Questions

1. **Priority field UI: textinput vs. inline selector?**
   - What we know: Other fields use `textinput.Model` or `textarea.Model`. Priority has only 5 valid values (0-4).
   - What's unclear: Whether a textinput (user types "1"-"4") or an inline selector (left/right arrows cycle through options) is better UX.
   - Recommendation: Use an inline selector (no textinput). The priority field displays `[none] P1 P2 P3 P4` and left/right arrows move the selection. This is faster than typing, prevents invalid input, and avoids needing a new textinput model. Store the selection as `editPriority int` on the Model. In the edit form, handle left/right keys when `editField` is the priority field.

2. **editField numbering: insert priority before or after body?**
   - What we know: Current cycle is title(0) -> date(1) -> body(2). Priority is a quick selection, body is a multi-line textarea.
   - What's unclear: Users might expect body right after date, or priority right after date.
   - Recommendation: Insert priority BETWEEN date and body. Rationale: priority is a quick one-tap selection that groups logically with the "metadata" fields (title, date), while body is the extended content field. This gives cycle: title(0) -> date(1) -> priority(2) -> body(3) [-> template(4) in inputMode].

3. **Should the badge width include the space, making it exactly 5 chars?**
   - What we know: `[P1]` is 4 chars. With a trailing space separator, the badge slot is 5 chars. For no-priority, 5 spaces maintains alignment.
   - Recommendation: Yes, fixed 5-char slot. `[P1] ` or `     ` (5 spaces). This is the simplest approach and matches `PRIO-04`.

## Sources

### Primary (HIGH confidence)
- Codebase: `internal/theme/theme.go` -- Theme struct pattern, all 4 theme definitions, color naming conventions
- Codebase: `internal/todolist/model.go` -- editField cycling, renderTodo, saveAdd/saveEdit with hardcoded priority 0
- Codebase: `internal/todolist/styles.go` -- Styles struct pattern, NewStyles from theme
- Codebase: `internal/search/model.go` -- search result rendering (View method, line formatting)
- Codebase: `internal/search/styles.go` -- search Styles pattern
- Codebase: `internal/calendar/grid.go` -- RenderGrid/RenderWeekGrid indicator logic, hasPending/hasAllDone cascade
- Codebase: `internal/calendar/styles.go` -- Indicator/IndicatorDone/TodayIndicator style pattern
- Codebase: `internal/calendar/model.go` -- RefreshIndicators pattern, indicators map storage
- Codebase: `internal/store/iface.go` -- TodoStore interface, IncompleteTodosPerDay method signature
- Codebase: `internal/store/sqlite.go` -- IncompleteTodosPerDay SQL query pattern (GROUP BY day)
- Codebase: `internal/store/todo.go` -- HasPriority(), PriorityLabel() helpers (from Phase 31)
- Codebase: `.planning/REQUIREMENTS.md` -- PRIO-01 through PRIO-07 definitions, out-of-scope list
- Codebase: `.planning/phases/31-priority-data-layer/31-RESEARCH.md` -- Phase 31 decisions and data layer context

### Secondary (MEDIUM confidence)
- Nord color palette: https://www.nordtheme.com (referenced in existing theme comments)
- Solarized palette: https://ethanschoonover.com/solarized (referenced in existing theme comments)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all changes within existing packages following established patterns
- Architecture: HIGH -- every pattern (theme colors, styles, editField cycling, calendar indicators, search rendering) has direct precedent in the codebase
- Pitfalls: HIGH -- derived from direct codebase analysis with specific line numbers and file paths
- Theme colors: MEDIUM -- color hex values are based on established palette families (Nord, Solarized) but exact values need visual verification in a real terminal

**Research date:** 2026-02-13
**Valid until:** 2026-03-15 (stable -- TUI rendering patterns do not change rapidly)
