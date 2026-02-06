# Feature Research: v1.3 Views & Usability

**Domain:** TUI Calendar v1.3 features (weekly view, search/filter, overview colors, date format)
**Researched:** 2026-02-06
**Confidence:** HIGH for UX patterns, MEDIUM for implementation details

## Feature Landscape

### Table Stakes

Features users expect when these capabilities are announced. Missing any = feature feels half-baked.

#### Weekly Calendar View

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| 7-day grid showing one week | The whole point of a weekly view. Users expect to see Mon-Sun (or Sun-Sat) for the selected week. | MEDIUM | Must respect `first_day_of_week` setting. |
| Toggle between monthly/weekly with single key | Calcurse toggles views. Calcure uses `v` key. Users expect seamless switching. | LOW | Single keybinding (e.g., `w` or `v`). Must remember which month/week context to return to. |
| Navigate between weeks | If you can see a week, you need to move to next/previous week. Calcurse uses j/k in weekly mode. | LOW | Same keys as month navigation (h/l) but move by week instead of month. |
| Current week highlighted / today visible | Orientation is critical. Users must know "where am I in time?" | LOW | Today gets the same highlight style as in monthly view. |
| Weekday headers | Same as monthly view -- day-of-week labels at top of columns. | LOW | Already exist in monthly view; reuse the pattern. |
| Todo indicators on days | Already exist as `[N]` in monthly view. Weekly view must show them too. | LOW | Same data source (`IncompleteTodosPerDay`), just different rendering. |
| Holiday display on days | Already exist in monthly view. Weekly view must show holidays. | LOW | Same data source, different rendering context. |

#### Search/Filter Todos

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Inline filter in todo panel | Taskwarrior-tui uses `/` to filter. This is the standard TUI pattern. Users type, list narrows in real time. | MEDIUM | Filter applies to visible todos only (current month + floating). |
| Clear filter with Escape | Standard pattern across all TUI apps with search. Esc exits filter mode and restores full list. | LOW | Must cleanly restore cursor position and full list. |
| Full-screen search overlay across all months | Differentiator -- searching beyond the current view. Users need to find "where did I put that todo?" | HIGH | Requires scanning all todos, displaying results with month context, and navigating to a result. |
| Substring matching (case-insensitive) | Users expect typing "buy" to find "Buy groceries". Case-insensitive is the default expectation. | LOW | `strings.Contains(strings.ToLower(...))` -- trivial. |
| Visual indication of active filter | When a filter is active, users must know the list is filtered, not empty. | LOW | Show filter text in a status line or change the section header. |

#### Overview Color Coding

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Distinct colors for incomplete vs complete | If you color-code the overview, red/green (or equivalent semantic colors) for incomplete/complete is the obvious mapping. | LOW | Two new style fields per theme. |
| Theme-aware colors | The app has 4 themes. Overview colors must work with all of them. | LOW | Add new semantic color roles to `Theme` struct (e.g., `OverviewIncompleteFg`, `OverviewCompleteFg`). |
| Current month still visually distinct | Overview already highlights current month in bold. This must not regress. | LOW | Keep existing `OverviewActive` style; layer color on top. |

#### Date Format Setting

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| At least 3 common presets | ISO (YYYY-MM-DD), European (DD.MM.YYYY), US (MM/DD/YYYY) cover the vast majority of users. Calcurse offers 4 formats. | LOW | Map preset names to Go layout strings. |
| Setting accessible in settings overlay | The app already has a settings overlay. Adding a new option row is the natural place. | LOW | Same pattern as theme/country/first-day-of-week cycling. |
| Dates update everywhere immediately | Calendar header, todo dates, date input prompts -- all must respect the format. | MEDIUM | Format string must propagate to all rendering code that displays dates. |
| Persistence in config.toml | Same as other settings. | LOW | New `date_format` field in config. |

### Differentiators

Features that go beyond what users minimally expect, creating competitive advantage.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Full-screen search across all months | Neither calcurse nor calcure offer cross-month todo search in a TUI overlay. Taskwarrior-tui has it, but that is a different product class. For a calendar+todo app, being able to find "where is that todo?" across all time is genuinely useful. | HIGH | This is the expensive part of search/filter. |
| Week view showing todos per day | Calcurse weekly view shows time-slotted appointments. Our weekly view would show todo counts or todo names per day -- a different and arguably more useful approach for a todo app (not a scheduler). | MEDIUM | This is the key design decision: our weekly view is todo-centric, not time-slot-centric. |
| Overview showing split counts (incomplete/complete) | Current overview shows `[N]` total count. Showing `[3/5]` or color-coded incomplete vs complete gives much richer at-a-glance information. | LOW | Small change, high information density gain. |
| Custom date format string | Beyond the 3 presets, allowing a custom Go layout string (e.g., `02 Jan 2006`) is power-user friendly. | LOW | Just a textinput in settings; validate by attempting `time.Format()`. |

### Anti-Features

Things to deliberately NOT build for v1.3, even though they seem related.

| Anti-Feature | Why It Seems Related | Why Avoid | What to Do Instead |
|--------------|---------------------|-----------|-------------------|
| Day selection / cursor on calendar | Weekly view might tempt adding per-day cursor navigation in the calendar grid. | This was explicitly ruled out in PROJECT.md: "Individual day selection / day-by-day arrow navigation -- month-level navigation is sufficient." Adding it contradicts the core design. Weekly view navigates by week, not by day. | Navigate weeks (prev/next week), not individual days. |
| Time-slotted weekly view | Calcurse shows 4-hour time slots in weekly view. Seems like the standard. | This app has no concept of time-of-day. Todos have dates, not times. Time slots are meaningless without appointments. Building time slots would require a data model change for zero benefit. | Show a simple 7-column grid with day numbers, holidays, and todo indicators. The todo panel shows todos for the visible week (all 7 days). |
| Fuzzy matching in search | Taskwarrior-tui supports regex. Fuzzy finders like fzf are popular. | Fuzzy matching adds complexity (scoring, ranking) for minimal gain when the dataset is small (personal todo list). Substring match covers 95% of use cases. | Use case-insensitive substring matching. If users complain, add regex support later. |
| Search result ranking / scoring | Results could be ranked by relevance, date proximity, completion status. | Over-engineering for a small dataset. Chronological order (by date) is the natural and expected ordering. | Show results in chronological order (dated todos by date, then floating todos). |
| Locale-based auto-detection of date format | Could detect system locale and set date format automatically. | Go's locale detection is unreliable. The app already requires manual country selection for holidays. Consistency: let users choose explicitly. | Offer presets + custom. Default to ISO 8601 (YYYY-MM-DD) which is the existing internal format. |
| Separate overview panels for incomplete/complete | Could split the overview into two sections. | Doubles the vertical space used. The overview already competes for space below the calendar grid. | Use color coding within the existing single-line-per-month format. |

## Feature Details

### Weekly Calendar View

**Expected UX Behavior:**

The weekly view replaces the monthly grid in the calendar pane (left panel). It shows 7 days as columns, similar to the monthly view but zoomed in to one week. The critical design insight: this is a **todo calendar**, not a scheduling app. The weekly view shows day numbers, holiday indicators, and todo count indicators -- NOT time slots.

**Layout concept (34 chars wide, matching monthly grid width):**

```
     Week 6 -- February 2026
 Mo   Tu   We   Th   Fr   Sa   Su
  2    3    4    5    6    7    8
     [1]             [2]
```

Each day column shows: day number (top), todo indicator `[N]` if applicable (below), and holiday coloring on the day number. The todo panel (right) shows todos for all 7 days of the visible week, grouped by day, plus the floating section.

**Interaction patterns:**

| Action | Key | Behavior |
|--------|-----|----------|
| Toggle monthly/weekly | `w` | Switch view mode. Remember position: if viewing Feb 2026 monthly, switch to the week containing today (or the first of the month if not current month). |
| Next week | `l` or `right` | Advance by 1 week. If crossing month boundary, update month context. |
| Previous week | `h` or `left` | Go back 1 week. If crossing month boundary, update month context. |
| Return to monthly | `w` | Toggle back. Return to the month that contains the current week. |

**Edge cases:**

1. **Week spanning two months:** A week starting Mon Jan 26 includes days in both January and February. The week view must show all 7 days regardless of month boundaries. The todo panel should show todos for all 7 days (from both months).
2. **First/last week of year:** Week containing Jan 1 or Dec 31 may span years. Must handle year boundaries.
3. **Week numbering:** ISO 8601 week numbers (1-53) are standard in Europe. US uses different conventions. Since the app has `first_day_of_week`, use it: Monday-start = ISO week numbers, Sunday-start = simple "Week N" count.
4. **Returning to monthly view:** When toggling back from weekly to monthly, display the month that contains the majority of the current week (or the month of the first day of the week).
5. **Overview panel in weekly mode:** The overview panel below the calendar should still show per-month todo counts. It does not change between views.
6. **Todo panel grouping:** In weekly mode, the todo panel should group todos by day within the week (7 sections for dated, plus floating), rather than showing the entire month. Each day section shows the day name and date.

**Dependencies on existing features:**
- Calendar grid rendering (`RenderGrid`) -- needs a new `RenderWeekGrid` function or a mode parameter
- `first_day_of_week` setting -- determines which day starts the week
- `IncompleteTodosPerDay` -- needs to work across month boundaries for cross-month weeks
- Todo panel (`visibleItems`) -- needs a "week mode" that shows todos for 7 specific dates instead of a whole month
- Help bar -- needs to show week-specific navigation hints when in weekly mode

**Complexity assessment:** MEDIUM-HIGH. The rendering is straightforward, but the cross-month week boundary handling and the todo panel regrouping add real complexity. The biggest challenge is making the todo panel work with a 7-day range that may span two months.

---

### Search/Filter Todos

**Expected UX Behavior:**

Two distinct modes with different scopes:

**Mode 1: Inline filter (todo panel)**
- User presses `/` while focused on the todo panel
- A filter input appears at the top of the todo panel (or below the section headers)
- As the user types, todos are filtered in real time (case-insensitive substring match)
- Only matching todos are shown; non-matching todos are hidden
- Headers ("February 2026", "Floating") remain visible even if their section is empty after filtering
- Pressing `Escape` clears the filter and restores the full list
- Pressing `Enter` accepts the filter and returns to normal mode with the filter active
- A visual indicator shows that a filter is active (e.g., "Filter: buy" in the header area)

**Mode 2: Full-screen search overlay**
- User presses a different key (e.g., `Ctrl+f` or `?` -- needs to not conflict with existing bindings) from any context
- A full-screen overlay appears (similar to the settings overlay pattern)
- A search input at the top
- Results show all matching todos across ALL months, grouped by month
- Each result shows: todo text, date, completion status
- User can scroll through results
- Pressing `Enter` on a result navigates to that month and highlights/selects that todo
- Pressing `Escape` closes the overlay without navigating

**Interaction patterns for inline filter:**

| Action | Key | Behavior |
|--------|-----|----------|
| Start filter | `/` | Enter filter mode. Show text input at top of todo panel. |
| Type filter | (any text) | Filter list in real time. Case-insensitive substring match on todo text. |
| Accept filter | `Enter` | Lock in the current filter. Return to normal mode. Filter indicator stays visible. |
| Clear filter | `Escape` | Remove filter text, restore full list, return to normal mode. |
| Navigate filtered list | `j`/`k` | Move cursor within filtered results only. |

**Interaction patterns for full-screen search:**

| Action | Key | Behavior |
|--------|-----|----------|
| Open search | `/` (from calendar pane) or `Ctrl+f` (from either pane) | Show full-screen search overlay. |
| Type query | (any text) | Search all todos. Results update as you type. |
| Navigate results | `j`/`k` or `up`/`down` | Move cursor through search results. |
| Go to result | `Enter` | Close overlay, navigate calendar to the result's month, select the todo in the todo panel. |
| Close search | `Escape` | Close overlay, no navigation change. |

**Edge cases:**

1. **Empty search results:** Show "No matching todos" message, not a blank screen.
2. **Filter + add todo:** If a filter is active and user adds a todo, should the new todo appear even if it does not match the filter? Recommendation: Yes, temporarily show it (clear the filter on add).
3. **Filter + month navigation:** If user navigates to a different month while filter is active, should the filter persist? Recommendation: Yes, keep the filter active across month changes.
4. **Search result in a month far from current:** Navigating to the result must update both the calendar month and the todo panel.
5. **Multiple matches in same month:** Group them together in the search results with the month as header.
6. **Floating todos in search results:** Show them in a separate "Floating" section at the bottom of results.
7. **Completed todos in search:** Include them by default. Users searching across months may be looking for completed work.

**Dependencies on existing features:**
- Todo panel input modes (already have `inputMode`, `dateInputMode`, etc.) -- add `filterMode`
- Store methods -- need `SearchTodos(query string) []Todo` that scans all todos
- Settings overlay pattern -- reuse for full-screen search overlay
- Calendar month navigation -- search "go to result" must trigger month change
- Help bar -- needs filter-specific bindings when filter is active

**Complexity assessment:** MEDIUM for inline filter (reuses existing textinput and mode patterns). HIGH for full-screen search overlay (new component, cross-month data access, result navigation).

**Recommendation:** Build inline filter first (Phase 1 of search). Full-screen search can be a separate phase. Inline filter alone provides 80% of the value for 30% of the cost.

---

### Overview Color Coding

**Expected UX Behavior:**

The overview panel currently shows lines like:

```
Overview
 February 2026   [5]
 March 2026      [2]
 Unknown         [3]
```

With color coding, the counts would show completion information:

```
Overview
 February 2026   [3] [2]    (3 incomplete in red/warm, 2 complete in green/cool)
 March 2026      [2]        (2 incomplete, 0 complete -- no green indicator)
 Unknown         [1] [2]    (1 incomplete, 2 complete)
```

**Design decisions:**

1. **Format options:**
   - Option A: `[3] [2]` -- two separate brackets, red and green respectively
   - Option B: `[3/5]` -- fraction format (incomplete/total)
   - Option C: `[3+2]` -- additive format

   **Recommendation: Option A** (`[3] [2]`). Reasons:
   - Consistent with existing `[N]` bracket indicator pattern on calendar dates
   - Each bracket can be independently colored
   - Clearer at a glance than fraction notation
   - If complete count is 0, just show `[3]` in red (no green bracket) -- cleaner than `[3/3]` or `[3+0]`

2. **Color mapping across themes:**

   | Theme | Incomplete (pending work) | Complete (done) |
   |-------|--------------------------|-----------------|
   | Dark | Red (`#AF0000`) | Green (`#5F8700`) |
   | Light | Red (`#D70000`) | Green (`#008700`) |
   | Nord | Aurora Red (`#BF616A`) | Aurora Green (`#A3BE8C`) |
   | Solarized | Red (`#DC322F`) | Green (`#859900`) |

   These colors already exist in the theme palettes (red is used for holidays, green/aurora green is used for indicators in Nord/Solarized). This is deliberate: reuse established palette colors for semantic consistency.

3. **Colorblind accessibility:** Red-green is the most common form of color blindness (8% of men). Mitigation:
   - The bracket format `[3] [2]` provides positional information (first bracket = incomplete, second = complete) even without color
   - The existing `[N]` indicator on calendar dates is already position-dependent (it only appears on dates with incomplete todos)
   - Optional: use bold on incomplete counts for additional visual differentiation

**Edge cases:**

1. **Month with only completed todos:** Show `[5]` in green only (no red bracket). This is the "all done" state.
2. **Month with only incomplete todos:** Show `[3]` in red only (no green bracket). Current behavior, just colored.
3. **Month with zero todos:** Should not appear in overview (current behavior, no change needed).
4. **Current month emphasis:** The active month is already bold. Colors layer on top of bold.
5. **Floating section:** Same coloring applies to the "Unknown" (floating) line.

**Dependencies on existing features:**
- Theme struct -- add 2 new color fields: `OverviewIncompleteFg`, `OverviewCompleteFg`
- Calendar styles -- add 2 new styles: `OverviewIncomplete`, `OverviewComplete`
- Store -- need `TodoCountsByMonthSplit()` returning incomplete AND complete counts (current `TodoCountsByMonth()` returns total count only)
- Overview rendering in `calendar/model.go` -- update `renderOverview()` to use new data and styles

**Complexity assessment:** LOW. This is mostly adding 2 new theme colors, a new store method, and updating the render function. Small, well-scoped change.

---

### Date Format Setting

**Expected UX Behavior:**

Users select a date format from the settings overlay. The format affects how dates are displayed throughout the app. The internal storage format remains `YYYY-MM-DD` (ISO 8601) -- only the display changes.

**Three presets:**

| Preset Name | Display Format | Go Layout | Regions |
|-------------|---------------|-----------|---------|
| ISO | 2026-02-06 | `2006-01-02` | International, developer preference |
| European | 06.02.2026 | `02.01.2006` | Germany, Finland, most of EU |
| US | 02/06/2026 | `01/02/2006` | United States |

**Custom format:** Allow entering a Go time layout string directly. The settings overlay shows a preview of the current date in the selected format for immediate feedback.

**Where dates appear in the app (places that must update):**

1. **Calendar header:** "February 2026" -- this is month+year, not a full date. The date format setting should NOT change this. Month names are always English (no locale). Only full dates (day+month+year) are affected.
2. **Todo date display:** `2026-02-06` shown next to todo text in the todo panel. THIS is the primary place the format matters.
3. **Date input prompt:** When adding a dated todo or editing a date, the placeholder should show the active format (e.g., "DD.MM.YYYY" or "YYYY-MM-DD") so users know what to type.
4. **Date input parsing:** The input parser must accept the configured format AND the ISO format (as fallback). Users may paste ISO dates regardless of display setting.
5. **Search results:** If full-screen search is implemented, dates in results must use the configured format.

**Settings overlay integration:**

Add a fourth row to the settings overlay:

```
> Theme              <  Dark  >
  Country            <  FI - Finland  >
  First Day of Week  <  Monday  >
  Date Format        <  ISO (2026-02-06)  >
```

The display value shows both the preset name and a preview of today's date in that format. For custom format, show the format string and preview:

```
  Date Format        <  Custom: 02 Jan 2006  >
```

**Interaction for custom format:**
- Cycling through options: ISO -> European -> US -> Custom
- When "Custom" is selected, pressing Enter (or right arrow) opens a text input for the format string
- The preview updates as the user types

**Edge cases:**

1. **Invalid custom format:** If user enters a non-sensical format string, validate by attempting `time.Now().Format(layout)` and checking the result contains expected components. If invalid, reject and keep previous.
2. **Date input parsing ambiguity:** `01/02/2026` -- is that Jan 2 or Feb 1? The parser must use the configured format, not guess. The placeholder text must make the expected format clear.
3. **Existing data:** Stored dates are always ISO 8601 (`YYYY-MM-DD`). The format setting is display-only. No data migration needed.
4. **Config serialization:** Store the Go layout string in config.toml: `date_format = "2006-01-02"` (default), `date_format = "02.01.2006"`, etc. Presets are just convenient names for specific layout strings.
5. **Date input must accept configured format:** If user's format is `DD.MM.YYYY`, typing `06.02.2026` in the date input must parse correctly. The input parser should try the configured format first, then fall back to ISO.

**Dependencies on existing features:**
- Config struct -- add `DateFormat string` field with default `"2006-01-02"`
- Settings overlay -- add fourth option row with cycling + custom input
- Todo rendering (`renderTodo`) -- format `todo.Date` using config format instead of raw string
- Date input mode (`dateInputMode`, `editDateMode`) -- update placeholder and parser to use config format
- Theme -- no change needed (this is not a visual styling feature)

**Complexity assessment:** LOW-MEDIUM. The config/settings/display parts are straightforward. The tricky part is date input parsing -- accepting the configured format while remaining robust against typos and ambiguous inputs.

---

## Feature Dependencies

```
[Weekly Calendar View]
    |
    +--requires--> [first_day_of_week] (determines week start)
    |
    +--requires--> [Calendar grid rendering] (new RenderWeekGrid or mode)
    |
    +--requires--> [IncompleteTodosPerDay] (indicators in weekly grid)
    |
    +--requires--> [Holiday provider] (holidays in weekly grid)
    |
    +--modifies--> [Todo panel grouping] (show todos for 7 days, not full month)
    |
    +--modifies--> [Help bar] (show week navigation hints)
    |
    +--independent-of--> [Search/filter] (no dependency)
    +--independent-of--> [Overview colors] (no dependency)
    +--independent-of--> [Date format] (uses same format for display)

[Search/Filter Todos]
    |
    +-- Inline filter:
    |   +--requires--> [Todo panel] (renders filtered list)
    |   +--requires--> [textinput component] (already used for add/edit)
    |   +--requires--> [mode state machine] (add filterMode)
    |   +--independent-of--> [Calendar pane] (filter is todo-panel only)
    |
    +-- Full-screen search:
        +--requires--> [Store.SearchTodos()] (scan all todos)
        +--requires--> [Overlay pattern] (reuse settings overlay approach)
        +--requires--> [Calendar month navigation] (go-to-result)
        +--enhanced-by--> [Date format] (display dates in configured format)

[Overview Color Coding]
    |
    +--requires--> [Theme struct] (new color fields)
    |
    +--requires--> [Store split counts] (incomplete + complete per month)
    |
    +--requires--> [Overview rendering] (update renderOverview)
    |
    +--independent-of--> [Weekly view] (overview works same in both views)
    +--independent-of--> [Search/filter] (no dependency)

[Date Format Setting]
    |
    +--requires--> [Config struct] (new field)
    |
    +--requires--> [Settings overlay] (new option row)
    |
    +--modifies--> [Todo date rendering] (format display dates)
    |
    +--modifies--> [Date input parsing] (accept configured format)
    |
    +--independent-of--> [Weekly view] (format applies to both views)
    +--independent-of--> [Search/filter] (format used in search results)
    +--independent-of--> [Overview colors] (overview shows counts, not dates)
```

**Key insight: All four features are independent of each other.** They can be built in any order. The only soft dependency is that date format affects how dates appear in search results -- but search can launch with ISO dates and pick up the format setting when it arrives.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Risk | Priority | Rationale |
|---------|-----------|-------------------|------|----------|-----------|
| Overview color coding | HIGH | LOW | LOW | P1 (build first) | Highest value-to-cost ratio. Small, self-contained change. Instant visual improvement. |
| Date format setting | MEDIUM | LOW-MEDIUM | LOW | P2 | Well-understood problem. Settings overlay pattern is proven. Main risk is date input parsing. |
| Inline filter (todo panel) | HIGH | MEDIUM | LOW | P3 | Reuses existing textinput and mode patterns. High utility as todo lists grow. |
| Weekly calendar view | MEDIUM | MEDIUM-HIGH | MEDIUM | P4 | Cross-month boundary handling is the hardest problem. Todo panel regrouping is non-trivial. |
| Full-screen search overlay | MEDIUM | HIGH | MEDIUM | P5 (build last) | Most complex feature. New component. Can be deferred if time-constrained. |

**Recommended build order rationale:**

1. **Overview colors first** -- quick win, builds confidence, no risk
2. **Date format second** -- adds a settings row (warm-up for settings pattern), date parsing is a foundation for any future date work
3. **Inline filter third** -- adds a new mode to an existing component (known pattern), provides immediate search value
4. **Weekly view fourth** -- most architectural change, benefits from having date format and filter already working
5. **Full-screen search last** -- highest cost, most optional, inline filter already covers 80% of search needs

## Competitor Feature Comparison (v1.3 Scope)

| Feature | calcurse | calcure | taskwarrior-tui | Our v1.3 Approach |
|---------|----------|---------|-----------------|-------------------|
| Weekly view | Yes (time-slotted, 4hr blocks) | Daily view (not weekly) | N/A (task list) | Todo-centric 7-day grid (no time slots) |
| Search/filter | CLI flags only (no TUI search) | Not documented | `/` inline filter | Inline filter + full-screen overlay |
| Overview colors | N/A | `color_todo`, `color_done` | Taskwarrior colors | Semantic red/green per theme |
| Date format | 4 presets (mm/dd, dd/mm, yyyy/mm/dd, yyyy-mm-dd) | Not configurable | Taskwarrior rc | 3 presets + custom Go layout |

## Sources

- [calcurse manual -- weekly view, date formats, key bindings](https://calcurse.org/files/manual.html) -- HIGH confidence, authoritative official docs
- [calcure documentation and key bindings](https://anufrievroman.gitbook.io/calcure) -- MEDIUM confidence, official but sparse on search details
- [calcure settings -- color configuration, view options](https://anufrievroman.gitbook.io/calcure/settings) -- HIGH confidence, confirmed "weekly view is not supported" in calcure
- [taskwarrior-tui keybindings -- `/` filter pattern, Esc to exit](https://kdheepak.com/taskwarrior-tui/keybindings/) -- HIGH confidence, authoritative
- [Bubble Tea framework and bubbles components](https://github.com/charmbracelet/bubbles) -- HIGH confidence, list component has built-in fuzzy filtering
- [Go time.Format reference -- layout patterns](https://pkg.go.dev/time) -- HIGH confidence, stdlib docs
- [Go date formatting guide](https://yourbasic.org/golang/format-parse-string-time-date-example/) -- HIGH confidence, well-verified reference
- [Color blind accessibility guidelines](https://rgblind.com/blog/color-blind-friendly-palette) -- MEDIUM confidence, general UX guidance

---
*Feature research for: TUI Calendar v1.3 Views & Usability*
*Researched: 2026-02-06*
