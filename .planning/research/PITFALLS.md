# Pitfalls Research: v1.3 Feature Integration

**Domain:** TUI Calendar v1.3 -- weekly view, search/filter, overview colors, date format
**Researched:** 2026-02-06
**Confidence:** HIGH (pitfalls derived from analysis of actual codebase + verified Bubble Tea patterns)

This document covers pitfalls specific to ADDING four v1.3 features to the existing stable codebase. It replaces the prior v1.0 scaffold-focused pitfalls document. The existing app already handles the v1.0 pitfalls (atomic writes, frame sizing, WindowSizeMsg guard, Elm Architecture discipline).

---

## Critical Pitfalls

### Pitfall 1: Weekly View Grid Width Differs From Monthly Grid Width

**What goes wrong:** The existing `RenderGrid` function produces a 34-character-wide grid (7 columns x 4-char cells + 6 separators). A weekly view shows only 7 days but typically needs MORE horizontal space per cell (to show todo previews or date labels), not less. If the weekly view renders at a different width than 34 characters, the layout in `app.View()` breaks because `calendarInnerWidth` is hardcoded to 38 (34 grid + padding). The calendar pane either overflows or has dead space, and the todo pane width calculation `todoInnerWidth := m.width - calendarInnerWidth - (frameH * 2)` produces wrong values.

**Why it happens:** The app model treats the calendar pane as fixed-width (line 255 of `app/model.go`: `calendarInnerWidth := 38`). Monthly and weekly grids have fundamentally different space needs. Developers add the weekly view to `RenderGrid` or a new `RenderWeekGrid` function but forget to update the width contract at the app level.

**How to avoid:**
- Make the calendar pane width dynamic, queried from the calendar model based on current view mode: `calendarInnerWidth := m.calendar.ContentWidth()`
- Alternatively, keep both views at exactly 34 characters wide. The weekly view renders 7 cells in the same 4-char format but with a single row instead of 5-6 rows. This is simpler and avoids the width problem entirely.
- Decide the approach BEFORE implementation. The "same width" approach is strongly recommended -- it preserves the existing layout contract and only changes the vertical dimension.

**Warning signs:**
- Todo pane suddenly shrinks or grows when toggling views
- Calendar content wraps to next line in weekly mode
- Hardcoded `38` still present in app model after adding weekly view

**Recovery cost:** LOW -- layout arithmetic fix, no data impact.

**Phase to address:** Weekly view phase (first).

---

### Pitfall 2: View Mode State Not Synced Across Calendar and Todo List

**What goes wrong:** When toggling between monthly and weekly view, the calendar model changes its view mode but the todo list model still shows the full month of todos. Or worse: the calendar shows "week of Feb 3-9" but the todo list shows all of February, creating a confusing mismatch. The `SetViewMonth(year, month)` API on the todo list has no concept of "week" -- it only accepts year+month.

**Why it happens:** The existing sync point is in `app/model.go` line 137 and 170: `m.todoList.SetViewMonth(m.calendar.Year(), m.calendar.Month())`. This works because the calendar only navigates at month granularity. Adding weekly navigation means the calendar now has sub-month positioning (which week), but the todo list API cannot express "show only this week's todos."

**How to avoid:**
- Decide upfront whether weekly view filters the todo list to that week's todos or not.
- **Recommended approach:** Weekly view is a CALENDAR-ONLY visual change. The todo list always shows the full month regardless of view mode. This avoids the sync problem entirely and keeps the existing `SetViewMonth` API unchanged.
- If week-level todo filtering IS desired: extend the todo list API to `SetViewRange(startDate, endDate string)` and update the sync points in app.Update() to pass the week boundaries.

**Warning signs:**
- Calendar shows week view but todo list shows todos from dates not visible in the grid
- Toggling view mode does not update the todo list at all
- Navigation in weekly mode (prev/next week) does not call the todo list sync point

**Recovery cost:** MEDIUM -- requires API redesign of the calendar-todolist sync if discovered late.

**Phase to address:** Weekly view phase. Decide the sync strategy during planning, not during coding.

---

### Pitfall 3: Date Format Round-Trip Corruption

**What goes wrong:** The store uses `YYYY-MM-DD` internally (`store.dateFormat = "2006-01-02"`). A "date format" setting changes how dates are DISPLAYED (e.g., `DD/MM/YYYY`, `MM/DD/YYYY`). If the display format is accidentally used for storage or parsing, dates get silently corrupted. For example, storing `"06/02/2026"` (Feb 6 in DD/MM/YYYY) and later parsing it as MM/DD/YYYY produces June 2 -- a silent, data-corrupting bug with no error.

**Why it happens:** Go's `time.Format` and `time.Parse` use layout strings, not format specifiers. The layout `"01/02/2006"` means MM/DD/YYYY while `"02/01/2006"` means DD/MM/YYYY. These look nearly identical in code. A developer might pass the display format layout to `time.Parse` when reading from the store, or pass the display format to `store.Add()` instead of the canonical format.

**How to avoid:**
- **Hard rule:** The store NEVER changes. All dates in the store remain `YYYY-MM-DD` strings. The format setting is display-only.
- The date format conversion happens exclusively in `View()` functions and `renderTodo()`. Never in `Update()`, never in store methods.
- Create a single conversion function:
  ```
  func FormatDate(isoDate string, layout string) string
  ```
  This function takes the ISO date from the store and returns the display string. There is no reverse function needed -- user input for dates always uses `YYYY-MM-DD` (the existing dateInputMode already enforces this).
- Store the display format layout string in config, NOT the formatted date.

**Warning signs:**
- `time.Parse` called with any layout other than `"2006-01-02"` on store data
- Display format layout string passed to any store method
- User-entered dates parsed with the display format instead of `"2006-01-02"`
- Dates appearing correct in one format setting but wrong after changing the setting

**Recovery cost:** HIGH -- silent data corruption. Dates in the JSON file become ambiguous. If `"03/04/2026"` is stored, you cannot determine if it was March 4 or April 3 without external context. Recovery may require user to manually fix their data file.

**Phase to address:** Date format phase. Must be enforced from the first line of implementation.

---

### Pitfall 4: Custom Date Format Layout String Injection

**What goes wrong:** The date format feature offers "3 presets + custom." If the custom format allows arbitrary Go time layout strings, users can create formats that produce ambiguous output (e.g., `"01 02"` -- is that month-day or day-month?) or formats that cannot round-trip at all (e.g., `"Jan 2"` -- loses the year). Even worse, certain characters in Go layout strings have magic meaning: `1`, `2`, `3`, `4`, `5`, `6`, `7`, `01`, `02`, `15`, `Jan`, `Mon`, `MST`, etc. A user typing `"My date: 12/25"` as a custom format would see garbled output because Go interprets `1`, `2`, `5` as format verbs.

**Why it happens:** Go's time layout system is notoriously unintuitive. Every digit and many words in the reference time `Mon Jan 2 15:04:05 MST 2006` are format specifiers. Users unfamiliar with Go's system (which is everyone) will expect strftime-style `%Y-%m-%d` or moment.js-style `YYYY-MM-DD`.

**How to avoid:**
- **Recommended:** Do NOT expose raw Go layout strings to the user. Instead, offer a small set of validated presets:
  - `YYYY-MM-DD` (ISO) -> layout `"2006-01-02"`
  - `DD/MM/YYYY` (European) -> layout `"02/01/2006"`
  - `MM/DD/YYYY` (US) -> layout `"01/02/2006"`
- If a "custom" option is truly needed, present it as a preset-picker with perhaps 5-6 curated options, not a free-text field.
- If free-text custom IS implemented: validate the layout by formatting a known reference date and checking the output looks reasonable. Show a live preview in the settings overlay (the existing live-preview pattern supports this).

**Warning signs:**
- Free-text input field for date format with no validation
- Settings overlay does not preview the formatted date
- Layout string stored in config but never validated on load (corrupt config produces garbled dates on next startup)

**Recovery cost:** LOW -- display-only, no data corruption. But poor UX if dates render as nonsense.

**Phase to address:** Date format phase. Design the preset list during planning.

---

### Pitfall 5: Search Mode Conflicts With Existing Input State Machine

**What goes wrong:** The todolist already has a 5-mode state machine: `normalMode`, `inputMode`, `dateInputMode`, `editTextMode`, `editDateMode`. Adding search introduces at least one new mode (`searchMode`), possibly two (inline filter + full-screen overlay). If the search textinput reuses the same `m.input` field as the add/edit modes, entering search while mid-edit (or vice versa) corrupts the input state -- the user's half-typed todo text gets replaced by a search query, or the search query gets saved as a todo.

**Why it happens:** The current model has a single `input textinput.Model` field shared across all modes. The mode enum prevents concurrent use in normal operation, but adding a new mode creates edge cases: what happens if the user presses the search key while in `editTextMode`? The `IsInputting()` check at the app level (line 122 of `app/model.go`) suppresses most keys during input, but search might be bound to a key not covered by `IsInputting()`.

**How to avoid:**
- **Option A (recommended):** Use a separate `searchInput textinput.Model` field for search, distinct from the existing `input` field. This eliminates any state sharing between search and add/edit.
- **Option B:** Continue sharing `input` but make the mode transitions explicit. Search key is blocked when `IsInputting()` returns true. Exiting search clears the input and resets to `normalMode`.
- Either way: ensure `IsInputting()` returns true for the search mode too, so that `q` (quit) and `tab` (switch pane) are suppressed during search input.

**Warning signs:**
- Single `textinput.Model` field used for both search and add/edit
- Search key works while user is mid-edit of a todo
- `IsInputting()` does not cover the new search mode
- App quits when user types `q` during search

**Recovery cost:** MEDIUM -- requires refactoring the input field ownership if discovered after both features are built.

**Phase to address:** Search/filter phase. Design input ownership before coding.

---

### Pitfall 6: Full-Screen Search Overlay Message Routing Conflict

**What goes wrong:** The app already has a full-screen overlay pattern (`showSettings`). Adding a full-screen search overlay creates a second overlay state. If both `showSettings` and `showSearch` can be true simultaneously, the app has two overlays competing for message routing and View rendering. Even if they are mutually exclusive, the routing logic in `app.Update()` becomes a growing chain of `if showSettings { ... } else if showSearch { ... }` that is easy to get wrong.

**Why it happens:** The settings overlay was the first and only overlay, so it was implemented as a simple boolean flag with inline routing. Each new overlay adds another flag and another routing branch. The code in `app.Update()` lines 114-117 shows the pattern: `if m.showSettings { return m.updateSettings(msg) }`. Adding search duplicates this pattern.

**How to avoid:**
- **Recommended:** Generalize the overlay pattern to an enum or interface before adding the second overlay:
  ```
  type overlay int
  const (
      noOverlay overlay = iota
      settingsOverlay
      searchOverlay
  )
  ```
  The `Update()` function routes based on `m.activeOverlay`, and only one overlay can be active at a time. This prevents the "two overlays open simultaneously" bug by design.
- Ensure the search overlay key binding is suppressed when settings is open, and vice versa.

**Warning signs:**
- Two separate boolean flags (`showSettings`, `showSearch`) instead of a single enum
- Both overlays can be opened simultaneously
- Copy-pasted routing logic for each overlay

**Recovery cost:** LOW -- refactoring booleans to an enum is straightforward. But the two-overlays-open bug can be confusing to debug if it manifests as garbled rendering.

**Phase to address:** Search/filter phase (or earlier if architecture is being cleaned up).

---

### Pitfall 7: Overview Color Calculation Done in View() on Every Render

**What goes wrong:** The overview panel (`renderOverview()` in `calendar/model.go`) already calls `m.store.TodoCountsByMonth()` on every render. Adding color coding (red for overdue/incomplete, green for all-done) requires ADDITIONAL per-month computation: not just "how many todos" but "how many done vs incomplete." If this computation iterates all todos for every month on every render frame, it creates O(months x todos) work per frame. For a personal app this is unlikely to cause visible lag, but it sets a bad pattern.

**Why it happens:** The existing code already computes overview data fresh on every `View()` call (line 112: `months := m.store.TodoCountsByMonth()`). This was acceptable because it was a simple count. Adding completion-status color coding requires a second pass or a richer data structure, doubling the per-render cost.

**How to avoid:**
- Extend `TodoCountsByMonth()` to return completion data in the same pass: a struct with `{Total, Done, Incomplete int}` per month. One iteration, richer data.
- Alternatively, cache the overview data and only recompute after store mutations (when `RefreshIndicators()` is called). The current `RefreshIndicators()` only refreshes the current month's day-level indicators -- extend it to also refresh the overview cache.

**Warning signs:**
- Two separate store methods called in `renderOverview()` (one for counts, one for completion status)
- Noticeable delay when navigating months with many todos
- `TodoCountsByMonth()` signature unchanged but new `CompletionByMonth()` method added alongside it

**Recovery cost:** LOW -- refactoring the store method to return richer data is a minor change.

**Phase to address:** Overview colors phase.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | Avoid? |
|----------|-------------------|----------------|--------|
| Hardcoded `calendarInnerWidth := 38` surviving into weekly view | No layout refactor needed | Breaks when weekly view needs different width | YES -- make it dynamic or verify both views fit in 38 |
| Reusing single `textinput.Model` for search + add/edit | Fewer fields on model | State corruption when modes overlap | YES -- use separate textinput for search |
| Separate boolean flags for each overlay | Quick to add | Combinatorial state explosion with each new overlay | YES -- use enum from the second overlay onward |
| Storing display format in store (even "temporarily") | Simpler display logic | Data corruption if canonical format is lost | NEVER -- display format is View-only |
| `strings.ToLower` on every keystroke during search | Correct filtering | Allocates new strings per keystroke per todo | OK for personal use -- premature optimization to avoid |
| Free-text custom date format | Maximum flexibility | Users will create broken formats | YES -- use curated presets instead |

---

## UX Pitfalls

### UX Pitfall 1: View Toggle Has No Visual Indicator

**What goes wrong:** User presses the toggle key and the calendar changes, but there is no label or indicator showing which view mode is active. Users forget which mode they are in, especially if weekly and monthly views look similar for weeks with few events.

**How to avoid:** Add a mode indicator to the calendar header. The existing header line (`"February 2026"`) could become `"February 2026 [month]"` or `"Feb 3-9, 2026 [week]"`. This also solves the problem of knowing WHICH week is shown.

---

### UX Pitfall 2: Search Clears on Mode Exit With No Way to Resume

**What goes wrong:** User types a search query, reviews results, exits search to interact with a todo, then wants to resume searching -- but the query is gone. They have to retype it.

**How to avoid:** Preserve the last search query. When re-entering search mode, pre-populate the textinput with the previous query. The existing `editTextMode` pattern already does this (line 265: `m.input.SetValue(todo.Text)`).

---

### UX Pitfall 3: Overview Colors Invisible on Some Themes

**What goes wrong:** Red/green color coding for completion status is invisible or unreadable on certain terminal themes. Red-on-dark-red is invisible. Green that is close to the terminal's normal text color provides no signal.

**How to avoid:** Use the theme's semantic color roles, not hardcoded red/green. The theme already has 14 roles. Add two new roles (e.g., `OverviewDoneFg`, `OverviewPendingFg`) and define them per theme with sufficient contrast. Test all four themes (dark, light, nord, solarized).

---

### UX Pitfall 4: Date Format Setting Not Previewed in Settings Overlay

**What goes wrong:** User changes the date format setting but cannot see the effect until they close settings. They may cycle through formats without knowing which one produces `DD/MM/YYYY` vs `MM/DD/YYYY`.

**How to avoid:** The settings overlay already supports live preview for themes (via `ThemeChangedMsg`). Apply the same pattern: show a sample date formatted with the currently selected format directly in the settings overlay. e.g., `"Date Format    < DD/MM/YYYY >  (today: 06/02/2026)"`.

---

### UX Pitfall 5: Weekly Navigation Overloads Existing Keys

**What goes wrong:** The calendar currently uses `left`/`h` for previous month and `right`/`l` for next month. If weekly view reuses these keys to mean "previous week" / "next week", there is no way to jump to the previous/next month while in weekly view. Users get trapped navigating week-by-week through a 12-month range.

**How to avoid:** Define navigation semantics per view mode:
- **Monthly view:** `left`/`right` = previous/next month (unchanged)
- **Weekly view:** `left`/`right` = previous/next week; add `[`/`]` or `H`/`L` for previous/next month jump
- OR keep `left`/`right` as month navigation in both modes, and add new keys for week navigation in weekly mode

The key design must be decided before implementation. Update the help bar to reflect the current mode's key bindings.

---

## "Looks Done But Isn't" Checklist

- [ ] **Weekly view toggle:** Does the view mode persist across month navigation? (Switching to next month should stay in weekly view, not reset to monthly)
- [ ] **Weekly view boundary:** What happens at month boundaries? Week of Jan 27 - Feb 2 spans two months. Which month's todos are shown? Does the overview highlight both months?
- [ ] **Weekly view + monday start:** Does the weekly view respect the `mondayStart` config setting? The week must start on the configured day, not always Monday or always Sunday.
- [ ] **Search across all months:** Does the full-screen search actually search ALL todos in the store, not just the current month? The `visibleItems()` method only returns current-month + floating todos. Search needs a different data source.
- [ ] **Search results navigation:** After finding a todo via search, can the user navigate to it? Does selecting a search result switch the calendar to that todo's month?
- [ ] **Search with special characters:** Does searching for `[` or `]` work? These characters appear in the calendar grid as todo indicators `[12]`. If search uses regex internally, special characters will cause panics or wrong results.
- [ ] **Overview colors update after toggle:** When a todo is toggled complete/incomplete, does the overview color update immediately? The existing `RefreshIndicators()` call in app.Update() only refreshes day-level indicators. Overview colors need to refresh too.
- [ ] **Date format in todo list:** Does the date format setting affect dates shown in the todo list (line 478 of `todolist/model.go`: `m.styles.Date.Render(t.Date)`)? Currently it renders the raw ISO date string. The format setting must be propagated to the todolist renderer.
- [ ] **Date format in date input:** When the user enters a date for a new todo, the prompt says `"YYYY-MM-DD"`. This should always say `YYYY-MM-DD` regardless of the display format setting, because the input format is always ISO. Do NOT change the input format to match the display format -- that would require parsing ambiguous user input.
- [ ] **Date format on config load:** What happens if the config file has an invalid `date_format` value? Ensure `FormatDate` falls back to ISO format, not panic.
- [ ] **Theme color roles added:** Are the new overview color roles defined for ALL four themes (dark, light, nord, solarized)? Missing a theme causes zero-value colors (empty string = terminal default), which may be invisible against the background.
- [ ] **Help bar updated:** Does the help bar show the correct keys for the current mode? Weekly view has different navigation keys. Search mode has its own keys. The `currentHelpKeys()` method in app.Model must account for these.

---

## Pitfall-to-Phase Mapping

| Phase | Pitfall | Prevention Strategy |
|-------|---------|---------------------|
| Weekly View | Grid width mismatch (#1) | Keep weekly grid at 34 chars wide; verify `calendarInnerWidth` still correct |
| Weekly View | View mode not synced to todo list (#2) | Decide week-level todo filtering strategy upfront; recommended: month-level stays |
| Weekly View | Navigation key overload (UX #5) | Design key bindings before coding; update help bar |
| Weekly View | Month boundary weeks ("Looks Done" #2) | Test with Jan 27 - Feb 2 type weeks |
| Weekly View | Monday start respected ("Looks Done" #3) | Pass `mondayStart` to weekly renderer |
| Search/Filter | Input state machine conflict (#5) | Use separate `searchInput` textinput field |
| Search/Filter | Overlay routing conflict (#6) | Convert boolean overlay flags to enum |
| Search/Filter | Special characters in search ("Looks Done" #6) | Use `strings.Contains`, not regex |
| Search/Filter | Cross-month search data source ("Looks Done" #4) | Query `store.Todos()` directly, not `visibleItems()` |
| Overview Colors | Per-render computation cost (#7) | Extend `TodoCountsByMonth` to include completion data |
| Overview Colors | Colors invisible on themes (UX #3) | Add semantic color roles; test all four themes |
| Overview Colors | Colors not refreshing ("Looks Done" #7) | Extend `RefreshIndicators` or add separate refresh |
| Date Format | Round-trip corruption (#3) | Store is ALWAYS ISO; format conversion in View only |
| Date Format | Custom format injection (#4) | Use curated presets, not free-text |
| Date Format | Format not applied in todo list ("Looks Done" #8) | Propagate format setting to todolist renderer |
| Date Format | Input prompt unchanged ("Looks Done" #9) | Keep input prompt as YYYY-MM-DD always |
| Date Format | Preview in settings (UX #4) | Show sample formatted date in settings overlay |

---

## Recovery Strategies

| Pitfall | Recovery Cost | What To Do |
|---------|---------------|------------|
| Grid width mismatch (#1) | LOW | Fix `calendarInnerWidth` to be dynamic or verify both views fit |
| View mode sync (#2) | MEDIUM | Add `SetViewRange` API or accept month-level sync |
| Date format corruption (#3) | HIGH | Manual data file repair; no automated recovery possible |
| Custom format injection (#4) | LOW | Replace free-text with preset picker |
| Input state conflict (#5) | MEDIUM | Add separate textinput field; refactor mode transitions |
| Overlay routing (#6) | LOW | Convert booleans to enum |
| Overview computation (#7) | LOW | Extend store method to return richer struct |

---

## Sources

- Codebase analysis: `internal/app/model.go`, `internal/calendar/model.go`, `internal/calendar/grid.go`, `internal/todolist/model.go`, `internal/store/store.go`, `internal/store/todo.go`, `internal/config/config.go`, `internal/theme/theme.go`, `internal/settings/model.go` (HIGH confidence -- primary source)
- [Go time package documentation](https://pkg.go.dev/time) -- date format layout system (HIGH confidence)
- [Go time.Format reference date explanation](https://yourbasic.org/golang/format-parse-string-time-date-example/) -- "regrettable historic error" of American date convention (MEDIUM confidence)
- [Tips for building Bubble Tea programs - leg100](https://leg100.github.io/en/posts/building-bubbletea-programs/) -- state management patterns (MEDIUM confidence)
- [Managing nested models with Bubble Tea - Roman Parykin](https://donderom.com/posts/managing-nested-models-with-bubble-tea/) -- overlay routing complexity (MEDIUM confidence)
- [Overlay composition using Bubble Tea - Leon Mika](https://lmika.org/2022/09/24/overlay-composition-using.html) -- rendering challenges (MEDIUM confidence)
- [ISO week date - Wikipedia](https://en.wikipedia.org/wiki/ISO_week_date) -- week boundary edge cases (HIGH confidence)
- [Bubble Tea GitHub - charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) -- framework reference (HIGH confidence)
- [Lipgloss GitHub - charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) -- color profile and adaptive colors (HIGH confidence)
- [Case-insensitive string search in Go](https://programming-idioms.org/idiom/133/case-insensitive-string-contains/1723/go) -- search performance (MEDIUM confidence)
- Prior v1.0 pitfalls research (`.planning/research/PITFALLS.md` dated 2026-02-05) -- foundational patterns already addressed (HIGH confidence)

---

*Pitfalls research for: TUI Calendar v1.3 feature integration*
*Researched: 2026-02-06*
