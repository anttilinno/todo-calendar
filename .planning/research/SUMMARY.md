# Project Research Summary

**Project:** Todo Calendar v1.3 - Views & Usability Enhancements
**Domain:** TUI Calendar Application Feature Integration
**Researched:** 2026-02-06
**Confidence:** HIGH

## Executive Summary

This is a feature expansion for an existing, stable Go TUI calendar+todo application (phases 1-9 shipped). Version 1.3 adds four usability enhancements: weekly calendar view toggle, search/filter capabilities, overview color coding for completion status, and configurable date formats. Research confirms all four features can be implemented with zero new dependencies - the existing stack (Go 1.25.6, Bubble Tea v1.3.10, Lipgloss v1.1.0, Bubbles v0.21.1) provides everything needed.

The recommended approach prioritizes quick wins before complex features. Start with overview color coding (smallest scope, highest value-to-cost ratio), followed by date format settings (establishes format propagation patterns), then weekly view (contained within calendar component), and finish with search/filter (most architecturally impactful, splits naturally into inline filter and full-screen overlay). All four features are architecturally independent - they can be built in any order without blocking dependencies.

Key risks center on state management rather than technology. Critical pitfalls include: weekly view grid width contract (must stay at 34 chars), date format round-trip corruption (display format must never touch storage), input state machine conflicts (search needs separate textinput), and overlay routing complexity (generalize before adding second overlay). All risks are mitigable with upfront design decisions documented in the pitfalls research.

## Key Findings

### Recommended Stack

No new dependencies required. All four v1.3 features integrate cleanly into the existing technology stack with zero library additions. The Go standard library (`strings`, `time`, `fmt`) covers substring search and date formatting. Bubble Tea's Elm Architecture handles new view modes and overlays through existing message routing. Lipgloss provides color coding via `Foreground()` styling. Bubbles textinput, already used for todo add/edit, serves search input needs.

**Core technologies (unchanged):**
- **Go 1.25.6**: Standard library (`strings.Contains` for search, `time.Format` for dates) - sufficient for all new features
- **Bubble Tea v1.3.10**: Elm Architecture handles view mode toggles and overlay state - no framework limitations
- **Lipgloss v1.1.0**: `Foreground()`, `Background()`, `Bold()` methods cover overview color coding - existing theme system extends naturally
- **Bubbles v0.21.1**: `textinput.Model` reusable for search filter - no new components needed
- **BurntSushi/toml v1.6.0**: New `date_format` config field is a simple string - no schema changes

**Deliberately NOT adding:**
- sahilm/fuzzy: Overkill for personal todo lists; `strings.Contains` substring matching is clearer and adds zero dependencies
- bubbles/list: Would require rewriting existing custom todo list rendering
- bubbles/viewport: Search results fit in manual scroll tracking, same pattern as current todo list
- charmbracelet/huh: No forms needed; existing textinput and settings cycling patterns suffice

### Expected Features

All four features follow established TUI patterns from calcurse, calcure, and taskwarrior-tui. Research identified clear table stakes vs differentiators.

**Must have (table stakes):**
- Weekly view: 7-day grid with single-key toggle, respects `first_day_of_week`, shows todo indicators and holidays
- Inline filter: `/` key activates, Escape clears, case-insensitive substring match (standard TUI pattern from taskwarrior-tui)
- Overview colors: Red/green semantic coloring for incomplete/complete months, theme-aware across all 4 themes
- Date format: 3 common presets (ISO YYYY-MM-DD, European DD.MM.YYYY, US MM/DD/YYYY), persisted in config

**Should have (competitive differentiators):**
- Full-screen search across all months: Neither calcurse nor calcure offer cross-month search; provides genuine "where did I put that?" utility
- Week view showing todo counts per day: Todo-centric design (not time-slot appointments like calcurse)
- Overview split counts: `[3] [2]` format showing incomplete/complete provides richer information than `[5]` total
- Custom date format: Beyond presets, allow Go layout strings for power users

**Defer (explicitly ruled out):**
- Day selection/cursor on calendar: Contradicts PROJECT.md core design; weekly view navigates by week, not day
- Time-slotted weekly view: App has no concept of time-of-day; todos have dates only
- Fuzzy matching in search: Over-engineering for small dataset; substring covers 95% of use cases
- Search result ranking: Chronological order is natural and expected

### Architecture Approach

The existing architecture follows Bubble Tea Elm patterns with clean component boundaries: app.Model orchestrates, calendar.Model and todolist.Model handle respective panes, settings.Model provides overlay, store.Store is pure data. All four features integrate into this structure without architectural changes - only component extensions.

**Major components and v1.3 modifications:**

1. **calendar.Model** - Add `viewMode` enum (monthly/weekly), `weekOffset` for navigation, `RenderWeekGrid()` pure function alongside existing `RenderGrid()`, extend `renderOverview()` with done/pending color styles
2. **todolist.Model** - Add `filterMode` to state machine, `filterQuery` field, date format propagation via `SetDateFormat()`, format conversion in `renderTodo()` and input validation
3. **search.Model (NEW)** - Full-screen overlay following settings pattern, own textinput for query, `Search()` method on store, `SelectMsg`/`CloseMsg` bubble to app
4. **store.Store** - Extend `TodoCountsByMonth()` to return done/pending breakdown (not just total), add `Search(query)` method for cross-month queries
5. **config.Config** - Add `DateFormat` string field, `DateDisplayFormat()` helper translating presets to Go layouts
6. **theme.Theme** - Add 2 new color roles (`OverviewDoneFg`, `OverviewPendingFg`) to existing 14, bringing total to 16 semantic roles

**Key architectural patterns preserved:**
- Pure rendering: `RenderGrid()` and new `RenderWeekGrid()` remain pure functions
- Message routing: Root routes to focused child; overlays intercept when active
- Theme propagation: New `SetDateFormat()` follows same setter pattern as `SetTheme()`
- Overlay pattern: Search follows settings precedent (bool flag, message bubbling, full-screen replacement)
- Fixed layout: Calendar grid stays exactly 34 chars wide in both views

### Critical Pitfalls

1. **Weekly View Grid Width Mismatch** - The calendar grid must remain exactly 34 characters wide in both monthly and weekly modes. The app layout hardcodes `calendarInnerWidth := 38`. Changing this breaks todo pane width calculation and causes rendering overflow. Avoid by keeping both views at 34-char width (same column format, just 1 row vs 5-6 rows).

2. **Date Format Round-Trip Corruption** - Display format must NEVER touch storage. Store always uses ISO `YYYY-MM-DD` (`"2006-01-02"` layout). If display format accidentally reaches store methods or parsing logic, dates get silently corrupted (e.g., `06/02/2026` as DD/MM/YYYY stored, later parsed as MM/DD/YYYY produces wrong date). Avoid by: hard rule that format conversion happens ONLY in View functions, create single `FormatDate(isoDate, layout)` helper, never pass display layout to store.

3. **Search Input State Machine Conflict** - todolist.Model has 5-mode state machine sharing one `textinput.Model` field. Adding search mode creates edge case: user presses search while mid-edit, input state corrupts. Avoid by using separate `searchInput textinput.Model` field distinct from add/edit input, or ensure search key is blocked when `IsInputting()` returns true.

4. **Full-Screen Overlay Routing Conflict** - Adding search as second overlay alongside settings creates dual-overlay risk if both `showSettings` and `showSearch` booleans exist. Two overlays open simultaneously causes message routing chaos. Avoid by converting to overlay enum before adding search: `type overlay int; const (noOverlay, settingsOverlay, searchOverlay)`.

5. **View Mode State Not Synced to Todo List** - Calendar switches to weekly view but todo list shows full month, creating UX mismatch. Existing `SetViewMonth(year, month)` API has no concept of week ranges. Recommended approach: weekly view is calendar-only visual change, todo list always shows full month regardless (keeps API simple). If week-level filtering desired, extend to `SetViewRange(startDate, endDate)`.

## Implications for Roadmap

Based on research, suggest 4 phases numbered 10-13 (continuing from shipped phase 9). All features are independent - no blocking dependencies - but ordering optimizes for risk and learning.

### Phase 10: Overview Color Coding
**Rationale:** Smallest scope (4 files modified, no new packages), highest value-to-cost ratio, builds confidence with quick win. Entirely contained within calendar + store + theme - no app-level routing changes. Zero risk of breaking existing functionality since it is purely additive.

**Delivers:**
- Overview panel shows completion status via color: red for months with incomplete todos, green for all-done months
- Two new theme color roles defined for all 4 themes (dark, light, nord, solarized)
- Extended `MonthCount` struct with done/pending breakdown

**Addresses:**
- Table stakes: Distinct colors for incomplete vs complete (FEATURES.md line 39)
- Architecture: Modify renderOverview() and extend TodoCountsByMonth() (ARCHITECTURE.md lines 232-286)

**Avoids:**
- Overview color calculation in View(): Extend store method to return rich data in one pass (PITFALLS.md #7)
- Colors invisible on themes: Use semantic color roles, test all themes (PITFALLS.md UX #3)

**Research needed:** NO - standard Lipgloss color application, well-documented pattern

---

### Phase 11: Date Format Setting
**Rationale:** Establishes format propagation pattern before features that display dates (search). Cross-cutting but mechanically straightforward - config field flows through settings to display. Should precede search so results display in configured format. Main complexity is bidirectional conversion (display vs storage), better to solve early.

**Delivers:**
- Settings overlay gains 4th option row for date format with 3 presets (ISO, European, US)
- `config.DateFormat` field with default `"2006-01-02"`
- Format propagation via `SetDateFormat()` to todolist
- Date display conversion in `renderTodo()`, input validation updated

**Addresses:**
- Table stakes: 3 presets, accessible in settings, dates update everywhere (FEATURES.md lines 46-51)
- Architecture: Config addition, todolist format conversion, settings integration (ARCHITECTURE.md lines 290-380)

**Avoids:**
- Round-trip corruption: Storage always ISO, conversion in View only (PITFALLS.md #3)
- Custom format injection: Use curated presets, validate layouts (PITFALLS.md #4)
- Input prompt ambiguity: Keep input as YYYY-MM-DD always (PITFALLS.md "Looks Done" #9)

**Research needed:** NO - Go time.Format is well-documented, pattern established in settings cycling

---

### Phase 12: Weekly Calendar View
**Rationale:** Self-contained within calendar package (no new packages). Calendar's pure-function pattern (`RenderGrid`) extends naturally to `RenderWeekGrid`. Should come after date format so week headers can use configured format. Main complexity is cross-month week boundaries, but contained in calendar component.

**Delivers:**
- `viewMode` enum in calendar.Model (monthly/weekly)
- `RenderWeekGrid()` pure function rendering 7-day row
- Toggle key binding (`w`) switching views
- Week navigation (left/right moves by week in weekly mode, by month in monthly mode)
- Week offset tracking with month boundary handling

**Addresses:**
- Table stakes: 7-day grid, single-key toggle, week navigation, respect first_day_of_week (FEATURES.md lines 16-23)
- Differentiators: Todo-centric weekly view (not time-slots) (FEATURES.md line 59)
- Architecture: Add viewMode to calendar, new RenderWeekGrid, navigation branching (ARCHITECTURE.md lines 42-104)

**Avoids:**
- Grid width mismatch: Keep weekly at 34 chars same as monthly (PITFALLS.md #1)
- View sync issues: Weekly is calendar-only, todo list stays month-level (PITFALLS.md #2)
- Navigation key overload: Define semantics per mode, update help bar (PITFALLS.md UX #5)
- Month boundary weeks: Handle cross-month weeks like Jan 27 - Feb 2 (PITFALLS.md "Looks Done" #2)

**Research needed:** NO - week calculation is stdlib time.Date, existing grid rendering patterns apply

---

### Phase 13: Search/Filter Todos
**Rationale:** Most complex - new package (search overlay) + todolist filter mode + store methods + app routing. Benefits from all prior phases being stable. Search overlay follows settings pattern established in phase 1-9. Search results should display dates in configured format (depends on phase 11). Consider splitting into 13a (inline filter) and 13b (full-screen overlay) if scope feels large.

**Delivers:**
- Inline filter mode in todolist: `/` key, textinput, real-time substring filtering in `visibleItems()`
- Full-screen search overlay: new `search.Model` component, cross-month `store.Search()` method
- Result navigation: `SelectMsg` navigates calendar to result's month, positions todolist cursor
- Separate `searchInput` textinput to avoid state machine conflicts

**Addresses:**
- Table stakes: Inline filter with `/` and Escape, substring matching, visual filter indicator (FEATURES.md lines 29-33)
- Differentiators: Full-screen cross-month search, result navigation (FEATURES.md line 58)
- Architecture: Add search.Model package, extend store, add app overlay routing (ARCHITECTURE.md lines 106-229)

**Avoids:**
- Input state conflict: Use separate searchInput textinput (PITFALLS.md #5)
- Overlay routing conflict: Convert overlays to enum before adding search (PITFALLS.md #6)
- Special characters in search: Use strings.Contains, not regex (PITFALLS.md "Looks Done" #6)
- Cross-month data source: Query store.Todos() directly, not visibleItems() (PITFALLS.md "Looks Done" #4)

**Research needed:** NO for inline filter (standard textinput mode). MAYBE for full-screen overlay (first time with dual overlays, test enum pattern) - suggest quick spike on overlay enum refactor

**Optional split:**
- **Phase 13a:** Inline filter in todolist (contained, ~2 files, LOW complexity)
- **Phase 13b:** Full-screen search overlay (new package, 5+ files, MEDIUM-HIGH complexity)

### Phase Ordering Rationale

- **Dependencies:** All features are architecturally independent. Soft dependency: search should come after date format so results display correctly.
- **Risk management:** Start with smallest scope (overview colors) to build momentum. Tackle most complex (search) last when codebase changes are well-understood.
- **Learning path:** Date format establishes propagation patterns reused in weekly view. Settings overlay experience informs search overlay design.
- **Component isolation:** Phases 10-12 modify existing components only. Phase 13 adds new package - deferred until foundation is solid.

### Research Flags

**Phases with standard patterns (skip research-phase):**
- **Phase 10 (Overview colors):** Lipgloss color application, theme system extension - well-documented, existing patterns in codebase
- **Phase 11 (Date format):** Go time.Format, settings cycling pattern - stdlib documented, pattern established
- **Phase 12 (Weekly view):** Pure rendering function, time.Date week calculation - standard Go patterns
- **Phase 13a (Inline filter):** textinput mode, strings.Contains - existing patterns, trivial search

**Phases possibly needing spike/investigation:**
- **Phase 13b (Search overlay):** First dual-overlay scenario. Suggest quick spike on converting `showSettings` bool to overlay enum before full implementation. Not deep research, just 30min validation that enum refactor doesn't break settings.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Full codebase review (20 files, 2492 LOC). All dependencies verified in go.mod. Bubble Tea patterns observed in existing code. |
| Features | HIGH | calcurse, calcure, taskwarrior-tui official docs reviewed. UX patterns well-established. All table stakes identified. |
| Architecture | HIGH | Existing architecture fully documented. Component boundaries clear. Integration points identified with line numbers. |
| Pitfalls | HIGH | Derived from actual codebase analysis + Bubble Tea community patterns. All critical pitfalls have concrete prevention strategies. |

**Overall confidence:** HIGH

### Gaps to Address

- **Weekly view todo panel grouping decision:** Must decide during phase 12 planning whether weekly view filters todo list to 7 days or shows full month. Research recommends month-level for simplicity (keeps existing `SetViewMonth` API), but team may prefer week-level filtering for UX. Decision affects todolist API design.

- **Search overlay enum refactor timing:** Converting `showSettings` bool to overlay enum could happen in phase 13 or earlier. Recommend doing it immediately before phase 13b (search overlay) as a small preparatory refactor rather than bundled into search implementation. Keeps phase 13b focused.

- **Custom date format scope:** Research identifies Go layout strings as unintuitive for users. Three approaches: (1) presets only, (2) curated 5-6 options, (3) free-text with validation. Decide during phase 11 planning based on target user sophistication. Research leans toward presets-only for simplicity.

- **Week numbering convention:** ISO 8601 weeks (Monday start, 1-53 numbering) vs US convention (Sunday start, different numbering). App already has `first_day_of_week` setting. Research suggests: if Monday start, use ISO week numbers; if Sunday start, use simple "Week N of month" count. Clarify during phase 12 planning.

## Sources

### Primary (HIGH confidence)
- Full codebase review: all 20 source files in internal/ (app, calendar, todolist, store, config, theme, settings, holidays) - 2492 total LOC
- Go standard library documentation (pkg.go.dev): time package Format/Parse, strings package Contains/ToLower
- Bubble Tea framework (github.com/charmbracelet/bubbletea): Elm Architecture patterns, message routing
- Lipgloss documentation (pkg.go.dev/github.com/charmbracelet/lipgloss): Style API, Foreground/Background methods
- Bubbles textinput (pkg.go.dev/github.com/charmbracelet/bubbles/textinput): CharLimit, Placeholder, Focus/Blur API

### Secondary (MEDIUM confidence)
- calcurse official manual (calcurse.org/files/manual.html): weekly view UX, date format presets, key bindings
- calcure documentation (anufrievroman.gitbook.io/calcure): view options, confirmed no weekly view
- taskwarrior-tui keybindings (kdheepak.com/taskwarrior-tui/keybindings/): `/` filter pattern
- Go time.Format guides (yourbasic.org, gosamples.dev): Reference time layout patterns, format cheatsheets
- Bubble Tea community patterns (leg100.github.io, donderom.com, lmika.org): State management, overlay composition

### Tertiary (LOW confidence)
- ISO week date Wikipedia: Week boundary edge cases, ISO 8601 numbering - used for background, not primary design driver

---
*Research completed: 2026-02-06*
*Ready for roadmap: yes*
