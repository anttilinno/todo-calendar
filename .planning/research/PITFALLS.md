# Domain Pitfalls: Priority Levels & Natural Language Date Input

**Domain:** Adding P1-P4 priority and NL date input to existing Go/Bubble Tea TUI todo-calendar
**Researched:** 2026-02-12
**Confidence:** HIGH (pitfalls derived from deep codebase analysis of 8,177 LOC across 35 Go files, existing sort/theme/calendar/input systems, and NL date parser library evaluation)

This document covers pitfalls specific to ADDING priority levels (P1-P4) and natural language date input to the existing system. Each pitfall identifies what breaks, why, and exactly which existing code is at risk.

---

## Critical Pitfalls

Mistakes that cause data loss, broken sorting, or require schema/architecture redesigns.

### Pitfall 1: Priority Auto-Sort Destroys Manual Reorder State

**What goes wrong:** The existing system uses `SortOrder` (gap-10 spacing) with manual J/K reordering via `SwapOrder()`. Every SQL query orders by `sort_order, date, id`. Adding priority-based auto-sort (P1 at top, P4 at bottom) conflicts fundamentally: if the user manually moves a P3 item above a P1 item, does the app respect the manual order or re-sort by priority? Neither answer is satisfying without careful design.

**Why it happens:** The existing `ORDER BY sort_order, date, id` clause in `TodosForMonth()`, `FloatingTodos()`, and 5 other query methods hard-codes a single sort dimension. Adding `priority` to the ORDER BY (e.g., `ORDER BY priority, sort_order, id`) means manual reordering within a priority group works but users can never move a P3 above a P1. Conversely, keeping `ORDER BY sort_order, id` means priority is just metadata with no automatic positioning.

**Consequences:** If auto-sort wins, the J/K reorder keys (`SwapOrder` in `sqlite.go:470-488`) become unpredictable -- the user moves an item up, the next render re-sorts it back down by priority. If manual order wins, priority is just a color label with no sort behavior, which defeats the purpose.

**Where it breaks in the codebase:**
- `sqlite.go:284-291` -- `Todos()` uses `ORDER BY sort_order, id`
- `sqlite.go:296-311` -- `TodosForMonth()` uses `ORDER BY sort_order, date, id`
- `sqlite.go:362-370` -- `FloatingTodos()` uses `ORDER BY sort_order, id`
- `todolist/model.go:471-496` -- `updateNormalMode` J/K swap checks `curItem.section == prevItem.section` but not priority boundaries
- `sqlite.go:470-488` -- `SwapOrder()` blindly swaps sort_order values with no priority awareness

**Prevention:**

The cleanest pattern for this app is **priority as a sort tiebreaker, not a sort override**:

```
ORDER BY sort_order, priority, date, id
```

But this still does not group by priority. The better approach:

1. **Manual reorder stays king within each section.** Priority is a visual indicator (color, label) that informs the user but does NOT override their manual ordering. This matches Todoist's "Manual" sort mode where priority is just a colored flag.

2. **If grouping by priority is desired later**, it becomes a separate view mode (like the existing month/week toggle), not a default behavior. The current section structure (Dated / This Month / This Year / Floating) already groups logically -- adding priority sub-groups within sections creates a 4x4 matrix (4 sections x 4 priorities) that is too deep for a TUI.

3. **Provide a one-time "sort by priority" action** (not a persistent mode) that reassigns sort_order values based on priority, then returns to manual mode. The user can invoke it to organize, then continue manual reordering.

**Warning signs:**
- Adding `priority` to SQL ORDER BY clauses globally
- J/K reorder crossing priority boundaries without user confirmation
- Priority changing causes items to jump positions unexpectedly

**Detection:** Add a P1 todo and a P3 todo. Manually reorder P3 above P1 with J/K. Navigate away and back. If P3 jumped back below P1, auto-sort is overriding manual order.

**Phase to address:** Must be decided in the FIRST phase (schema + store), because the sort strategy determines the database query structure, the store interface, and all downstream rendering.

---

### Pitfall 2: NL Date Parser Cannot Produce Date Precision

**What goes wrong:** The existing date precision system (`date_precision` column: 'day', 'month', 'year', '') is deeply integrated -- it controls which section a todo appears in, how dates render, which SQL queries match, and how calendar indicators work. A natural language date parser that returns only a `time.Time` throws away precision information. Parsing "March" or "2027" must produce month-precision and year-precision respectively, not `2027-03-01T00:00:00` with day precision.

**Why it happens:** Every Go NL date library (`tj/go-naturaldate`, `olebedev/when`, `ijt/go-anytime`, `markusmobius/go-dateparser`) returns `time.Time` which is always day+time precise. None of them expose what the user actually typed -- they all normalize to a concrete timestamp. So "March" becomes "2026-03-12T00:00:00" (current day of March), "2027" becomes "2027-02-12T00:00:00" (current month/day of 2027), and the precision information is lost.

**Consequences:** Without precision information:
- "March" creates a day-precision todo for March 12 instead of a month-precision todo
- "2027" creates a day-precision todo for Feb 12, 2027 instead of a year-precision todo
- "tomorrow" correctly creates a day-precision todo (this case works)
- The todo appears in the wrong section (Dated instead of This Month or This Year)
- Calendar indicators show it on the wrong specific day
- Editing the todo shows a full date instead of just month/year

**Where it breaks in the codebase:**
- `store/todo.go:47-60` -- `IsMonthPrecision()`, `IsYearPrecision()`, `IsFuzzy()` methods
- `store/todo.go:71-88` -- `InMonth()` switches behavior on precision
- `store/todo.go:93-113` -- `InDateRange()` excludes fuzzy todos
- `sqlite.go:296-311` -- `TodosForMonth()` filters on `date_precision = 'day'`
- `sqlite.go:331-343` -- `MonthTodos()` filters on `date_precision = 'month'`
- `sqlite.go:347-359` -- `YearTodos()` filters on `date_precision = 'year'`
- `todolist/model.go:1229-1249` -- `renderFuzzyDate()` switches on precision

**Prevention:**

Build a thin wrapper around the NL parser that infers precision from the input text BEFORE parsing:

```go
type ParsedDate struct {
    Date      string // ISO "YYYY-MM-DD"
    Precision string // "day", "month", "year"
}

func ParseNaturalDate(input string, ref time.Time) (ParsedDate, error) {
    input = strings.TrimSpace(strings.ToLower(input))

    // Detect precision from input patterns BEFORE delegating to NL parser
    if isMonthOnly(input) {        // "march", "march 2026", "next month"
        return parseMonthPrecision(input, ref)
    }
    if isYearOnly(input) {         // "2027", "next year"
        return parseYearPrecision(input, ref)
    }
    // Default: day precision
    t, err := nlparser.Parse(input, ref)
    if err != nil {
        return ParsedDate{}, err
    }
    return ParsedDate{
        Date:      t.Format("2006-01-02"),
        Precision: "day",
    }, nil
}
```

The precision detection must happen at the string level (regex/pattern matching), not after the NL parser has already discarded the information by normalizing to `time.Time`.

**Warning signs:**
- NL parser used directly, its `time.Time` result passed to `store.Add()` with hardcoded "day" precision
- No test cases for "March", "2027", "next month", "this year"
- Precision detection attempted after parsing (impossible -- the information is gone)

**Detection:** Type "March" in the NL date input. If the todo appears in the "February 2026" dated section on March 12 instead of the "This Month" section, precision inference is broken.

**Phase to address:** Must be designed in the NL parser integration phase. The precision wrapper is the core of the NL date feature -- without it, NL input is strictly worse than the existing segmented input for fuzzy dates.

---

### Pitfall 3: NL Parser Returns Ambiguous or Wrong Dates Without User Confirmation

**What goes wrong:** The user types "next Friday" and the NL parser interprets it as a date. But which Friday? If today is Thursday, does "next Friday" mean tomorrow or 8 days from now? Different libraries disagree. The user has no way to verify what date was parsed until after the todo is created and they notice it is on the wrong day.

**Why it happens:** NL date parsing is inherently ambiguous. Libraries handle ambiguity differently:
- `tj/go-naturaldate` defaults to Past direction, so "Friday" means last Friday
- `olebedev/when` uses rule clustering with configurable distance
- `ijt/go-anytime` parses "next Friday" vs "Friday" differently but the distinction is subtle

The existing segmented date input (`dateSegDay`, `dateSegMonth`, `dateSegYear`) is unambiguous -- the user types exact numbers. Replacing it with NL parsing introduces an entirely new class of errors: correct syntax, wrong meaning.

**Consequences:** Todos end up on wrong dates. The user must edit them to fix, which is worse UX than typing the date correctly the first time. Trust in the NL input erodes quickly after a few wrong-date experiences.

**Where it breaks in the codebase:**
- `todolist/model.go:852-892` -- `saveAdd()` currently calls `deriveDateFromSegments()` which returns exact values; replacing with NL output that might be wrong changes the trust model
- `todolist/model.go:814-849` -- `saveEdit()` same issue

**Prevention:**

1. **Show the parsed date before saving.** After the user types NL text and presses Tab (to move to next field), display the resolved date in a confirmation line:
   ```
   Date: next friday
   Parsed: Friday, February 14, 2026 [day precision]
   ```
   This lets the user see what the parser understood before committing.

2. **Use the NL parser with a Forward direction** for this app. Todo dates are almost always in the future. Configure the parser accordingly:
   ```go
   // tj/go-naturaldate
   t, err := naturaldate.Parse(input, time.Now(), naturaldate.WithDirection(naturaldate.Future))
   ```

3. **Provide instant fallback to segmented input.** If the NL parse fails or the user is not confident, they can press a key to switch to the existing segmented date fields. Do NOT remove the segmented input.

**Warning signs:**
- No parsed-date confirmation shown to user
- NL parser configured with default (Past) direction for a todo app
- Segmented date input removed entirely in favor of NL-only

**Detection:** Type "Friday" on a Thursday. Check whether the todo ends up on tomorrow (correct for forward-facing todo) or last Friday (wrong).

**Phase to address:** The NL input UI phase. The confirmation display is a UX requirement, not a nice-to-have.

---

### Pitfall 4: Schema Migration Breaks Existing Data When Adding Priority Column

**What goes wrong:** Adding a `priority` column with `ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0` fails because SQLite's `ALTER TABLE ADD COLUMN` with a NOT NULL constraint requires a DEFAULT value -- but the semantic question is: what priority do existing todos get? If DEFAULT 0 means "no priority" but the UI interprets 0 as P1 (highest), all existing todos suddenly become high priority.

**Why it happens:** Priority value encoding must be decided before the migration. Common schemes:
- 0=none, 1=P1, 2=P2, 3=P3, 4=P4 (0 is special "unset")
- 1=P1, 2=P2, 3=P3, 4=P4 (no unset state)
- 4=P1, 3=P2, 2=P3, 1=P4 (higher number = higher priority, SQL `ORDER BY priority DESC` sorts P1 first)

If the encoding is wrong, sorting inverts (P4 at top instead of P1) or existing unprioritized todos sort incorrectly.

**Where it breaks in the codebase:**
- `sqlite.go:54-157` -- `migrate()` function, currently at version 6
- `sqlite.go:164-165` -- `todoColumns` constant must include `priority`
- `sqlite.go:168-189` -- `scanTodo()` must read the new column
- `store/todo.go:8-22` -- `Todo` struct must include Priority field

**Prevention:**

1. **Use 0 as "no priority" with explicit semantics.** Existing todos get priority=0 meaning "unset." In sort order, unset sorts AFTER all prioritized items (or equivalently, "unset" is lower than P4):
   ```sql
   ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0
   -- 0=none, 1=P1(highest), 2=P2, 3=P3, 4=P4(lowest)
   ```

2. **In queries that sort by priority,** handle 0 specially:
   ```sql
   ORDER BY CASE WHEN priority = 0 THEN 5 ELSE priority END, sort_order, id
   ```
   This puts unprioritized todos after P4 without changing their priority value.

3. **The migration itself is simple** -- just one ALTER TABLE. But test with an existing database that has todos to verify they retain correct behavior.

**Warning signs:**
- Priority encoding not documented before implementation
- DEFAULT value chosen without considering existing data semantics
- No explicit handling of priority=0 ("unset") in sort queries

**Detection:** Open the app after migration. If existing todos now show a priority badge or sort differently, the default value semantics are wrong.

**Phase to address:** The schema migration phase (must be first).

---

## Moderate Pitfalls

These cause UX confusion, technical debt, or rework but are recoverable.

### Pitfall 5: Four Priority Colors x Four Themes = 16 Color Decisions That Can Clash

**What goes wrong:** The theme system has 16 semantic color roles defined in `theme.Theme` struct. Adding 4 priority colors (`PriorityP1Fg`, `PriorityP2Fg`, `PriorityP3Fg`, `PriorityP4Fg`) means 4 new colors per theme, or 16 new color values total. These must be visually distinct from each other AND from existing semantic colors (AccentFg, CompletedFg, MutedFg, HolidayFg, PendingFg, etc.) across all 4 themes.

**Why it happens:** The existing theme palettes are carefully chosen from established palettes (Nord, Solarized) with limited color slots. Nord has exactly 8 aurora colors. Solarized has exactly 8 accent colors. Four new distinct colors may exhaust the palette, forcing colors that are too similar to existing roles.

Specific collision risks:
- P1 (urgent red) vs `HolidayFg` (red) vs `PendingFg` (red-family) -- already 2 reds in every theme
- P3 or P4 vs `MutedFg` -- low-priority colors that look like disabled text
- P2 (yellow/orange) vs `IndicatorFg` (yellow in Nord/Solarized) -- calendar indicators
- All priority colors vs `CompletedFg` (grey-family) when a prioritized todo is marked done

**Where it breaks in the codebase:**
- `theme/theme.go:10-37` -- Theme struct, 16 existing fields, grows to 20
- `theme/theme.go:40-59` -- Dark(), must pick 4 new colors
- `theme/theme.go:62-81` -- Light(), must pick 4 new colors
- `theme/theme.go:85-104` -- Nord(), only 8 aurora colors available
- `theme/theme.go:108-127` -- Solarized(), only 8 accent colors available
- `todolist/styles.go:9-24` -- Styles struct, needs 4 new priority styles
- `calendar/styles.go:9-26` -- Calendar Styles, needs priority-aware indicator styles

**Prevention:**

1. **Use a 4-color priority palette that works across light and dark backgrounds.** Recommended approach:
   - P1: Red (already exists as HolidayFg/PendingFg -- reuse PendingFg)
   - P2: Orange/Yellow (theme-specific accent)
   - P3: Blue/Cyan (use existing AccentFg or IndicatorFg)
   - P4: Grey (use existing MutedFg)

   This reuses 2-3 existing colors per theme, only requiring 1-2 genuinely new color slots.

2. **Do NOT add 4 new fields to Theme struct if you can reuse existing roles.** Define priority colors as a mapping from priority level to existing theme role:
   ```go
   func PriorityColor(t Theme, level int) lipgloss.Color {
       switch level {
       case 1: return t.PendingFg      // red-family, already exists
       case 2: return t.IndicatorFg    // yellow/orange, already exists
       case 3: return t.AccentFg       // blue/indigo, already exists
       case 4: return t.MutedFg        // grey, already exists
       default: return t.NormalFg
       }
   }
   ```
   This adds ZERO new theme fields and leverages the existing carefully-chosen palette.

3. **If dedicated priority colors are required,** add them but keep count low. Consider 2 new colors (P1 and P2) and map P3 and P4 to existing roles. Nord's aurora palette has red, orange, yellow, green, purple -- P1=red(nord11), P2=orange(nord12), P3=yellow(nord13 already used for IndicatorFg), P4=purple(nord15). Solarized: P1=red, P2=orange, P3=yellow(already IndicatorFg), P4=violet.

4. **Never rely on color alone.** Priority should also be shown as text: "[P1]", "[P2]", etc. This handles colorblind users and monochrome terminals.

**Warning signs:**
- 4 new arbitrary hex colors per theme with no derivation from the palette
- Priority colors that are indistinguishable from each other in the terminal
- P1 red that is identical to holiday red, making holidays look like priority indicators
- No text-based priority indicator, color-only

**Detection:** Set theme to each of the 4 themes. Create one todo at each priority. Check that all 4 priority levels are visually distinguishable from each other AND from existing UI elements (holidays, completed todos, section headers).

**Phase to address:** The theme integration phase. Color choices should be made by examining all 4 palettes simultaneously, not one at a time.

---

### Pitfall 6: Calendar Indicators Cannot Show Priority Color in 4-Character Cells

**What goes wrong:** The calendar grid uses 4-character cells (`[12]` for dates with todos, ` 12 ` for dates without). The bracket `[ ]` notation indicates "has pending todos." Adding priority color means the bracket should reflect the highest priority todo on that day. But the existing `IncompleteTodosPerDay()` returns `map[int]int` (day -> count), not priority information. The calendar grid render code (`grid.go:144-183`) uses a simple boolean `hasPending` check, not a priority-aware color selection.

**Why it happens:** The calendar indicator system was designed for binary state (has incomplete todos / all done / no todos). Priority adds a third dimension that the existing data pipeline does not carry.

**Where it breaks in the codebase:**
- `sqlite.go:374-395` -- `IncompleteTodosPerDay()` returns count, not max priority
- `sqlite.go:400-421` -- `TotalTodosPerDay()` same issue
- `calendar/grid.go:144-183` -- render loop uses `hasPending` bool, not priority level
- `calendar/grid.go:157-172` -- style priority chain: `TodayIndicator > TodayDone > Today > Holiday > Indicator > IndicatorDone > Normal`
- `calendar/styles.go:15-16` -- `Indicator` and `IndicatorDone` are single styles, not priority-indexed

**Prevention:**

1. **Add a new store method** that returns the highest priority per day:
   ```go
   HighestPriorityPerDay(year int, month time.Month) map[int]int
   // Returns: day -> highest priority (1-4, 0 for no priority)
   ```

2. **Keep the existing indicator system and layer priority on top.** The bracket `[12]` stays. The COLOR of the bracket changes based on the highest-priority incomplete todo. This requires the grid render to receive priority data alongside count data:
   ```go
   func RenderGrid(..., priorities map[int]int, ...) string {
   ```

3. **Do NOT try to show all 4 priority levels per day.** The 4-char cell cannot convey "2 P1s, 1 P3, and 3 P4s." Show only the highest priority color. If the day has any P1, the bracket is P1-colored.

4. **If priority colors are not distinguishable enough in the small calendar cell,** fall back to the existing single indicator color and show priority only in the todo list. Calendar priority coloring is a nice-to-have, not a requirement.

**Warning signs:**
- Trying to show per-todo priority colors in the calendar grid (impossible in 4 chars)
- Calendar grid receiving 4 separate maps (one per priority level)
- Grid render function signature growing beyond 12 parameters (already at 12)

**Detection:** Create a day with both P1 and P4 todos. Check that the calendar indicator shows P1's color (highest priority wins).

**Phase to address:** Should be deferred to AFTER basic priority + todolist rendering works. Calendar indicator priority is a polish item, not a core requirement.

---

### Pitfall 7: NL Date Input Mode Conflicts with Existing Segmented Date Field Cycling

**What goes wrong:** The existing edit form has 4 fields cycled with Tab: Title(0) -> Date Segments(1) -> Body(2) -> Template(3). The date field at position 1 is actually 3 sub-fields (day/month/year segments) that Tab advances through internally before moving to Body. Adding NL date input as the primary mode means field 1 is now a free-text input, but switching to segmented fallback re-introduces the 3-sub-field behavior. The Tab cycling logic in `updateInputMode()` (lines 666-691) has hardcoded field indices and segment advancement.

**Why it happens:** The segmented date input has its own internal focus state (`dateSegFocus int`, `dateSegOrder [3]int`) separate from the form's `editField int`. NL mode would use a single text input. Switching between modes mid-edit requires swapping which widgets are active, resetting focus state, and preserving any partially-entered date across the mode switch.

**Where it breaks in the codebase:**
- `todolist/model.go:666-691` -- `updateInputMode` Tab cycling with hardcoded `editField` values and `dateSegFocus` advancement
- `todolist/model.go:761-780` -- `updateEditMode` same Tab cycling logic
- `todolist/model.go:1362-1392` -- `updateDateSegment` handles separator chars, auto-advance
- `todolist/model.go:1216-1225` -- `renderDateSegments()` renders the 3-field layout
- `todolist/model.go:1254-1359` -- `deriveDateFromSegments()` complex 100-line precision derivation

**Prevention:**

1. **NL input and segmented input should NOT coexist in the same form simultaneously.** Use a single date field that accepts NL text. The segmented input becomes a FALLBACK accessed via a toggle key (e.g., Ctrl+D to switch modes), not a parallel UI element.

2. **When in NL mode,** field 1 is a single `textinput.Model`. Tab moves from Title to Date (single field) to Body. The `dateSegFocus` advancement is skipped entirely.

3. **When switching to segmented fallback,** parse any NL text already entered and pre-populate the segments. "March 15" should pre-fill month=03, day=15 in the segments.

4. **Store the current date input mode** in the model:
   ```go
   dateInputMode int // 0=natural language, 1=segmented
   ```
   The Tab cycling and `deriveDateFromSegments()` logic branches on this mode.

5. **Do NOT try to parse-as-you-type.** Wait until the user presses Tab/Enter to parse the NL input. Real-time parsing creates flickering resolved dates and confusing UX.

**Warning signs:**
- Both NL text input and segmented inputs visible simultaneously
- Tab sometimes advances date segments, sometimes moves to Body, depending on hidden mode state
- `deriveDateFromSegments()` called when in NL mode (ignores the NL input entirely)

**Detection:** Type "tomorrow" in the NL date field, press Tab. If focus goes to the first date segment instead of Body, the mode is not properly routing.

**Phase to address:** The NL input UI phase. This is the most complex integration point -- design the field switching BEFORE writing any NL parsing code.

---

### Pitfall 8: TodoStore Interface Requires Priority-Aware Methods Throughout

**What goes wrong:** Adding priority to the Todo struct and store is not just one column -- it ripples through the `TodoStore` interface. `Add()` needs a priority parameter. `Update()` needs a priority parameter. If priority affects sort order, all query methods need priority in their ORDER BY. The interface (16 methods currently) grows, and every call site must be updated.

**Where it breaks in the codebase:**
- `store/iface.go:8` -- `Add(text string, date string, datePrecision string) Todo` -- needs priority
- `store/iface.go:12` -- `Update(id int, text string, date string, datePrecision string)` -- needs priority
- `store/iface.go:36` -- `AddScheduledTodo(text, date, body string, scheduleID int) Todo` -- needs priority?
- `todolist/model.go:868` -- `m.store.Add(text, isoDate, precision)` -- call site
- `todolist/model.go:831` -- `m.store.Update(m.editingID, text, isoDate, precision)` -- call site

**Prevention:**

1. **Add a separate `UpdatePriority(id int, priority int)` method** instead of expanding existing signatures. This keeps `Add()` and `Update()` backward-compatible. New todos get priority 0 (default), and priority is set via a separate action.

2. **Alternatively, expand `Add()` and `Update()` signatures** -- this is also fine since there is only one implementation. But update ALL call sites at once, including `AddScheduledTodo()`.

3. **Choose one approach and be consistent.** Do not mix "priority as parameter" and "priority as separate method."

**Recommendation:** Expand `Add()` and `Update()` to include priority. This is a single-implementation app. Add the parameter, update the 3 call sites, done. A separate `UpdatePriority()` method is only needed if you want inline priority toggling (press a key to cycle priority on the selected todo without entering edit mode), which is a good UX pattern worth having.

**Warning signs:**
- `Add()` signature unchanged but priority somehow gets set (magic default that is never overridden)
- Some call sites pass priority, others do not, leading to inconsistent defaults
- `AddScheduledTodo()` forgotten -- scheduled todos always get priority 0

**Phase to address:** The schema + store phase. Interface changes must happen first.

---

### Pitfall 9: Completed Prioritized Todos Create Visual Confusion

**What goes wrong:** A completed todo currently renders with `CompletedFg` (grey) and strikethrough. A prioritized completed todo should arguably show its priority color BUT also look completed. If the priority color overrides the completed style, users cannot tell at a glance which priority items are done. If the completed style overrides priority, there is no point showing priority on completed items.

**Where it breaks in the codebase:**
- `todolist/model.go:1054-1093` -- `renderTodo()` applies styles sequentially: cursor, checkbox, text (with done override), body indicator, recurring indicator, date
- `todolist/model.go:1071-1075` -- Done text styling: `text = m.styles.Completed.Render(text)` -- this applies CompletedFg which is grey

**Prevention:**

1. **Priority color applies to the checkbox or a prefix badge, NOT to the text.** The text still gets CompletedFg when done. This way:
   ```
   > [x] [P1] Buy groceries  2026-02-15
   ```
   The `[P1]` badge retains its color. The text "Buy groceries" is grey with strikethrough. This gives both signals.

2. **Alternatively, dim the priority color when done.** Use the priority color but at reduced saturation or with the completed foreground. But this requires 8 new styles (4 priorities x 2 states) which is excessive.

3. **Simplest approach:** Priority badge `[P1]` is always colored by priority. Todo text follows existing done/undone styling. No new style combinations needed.

**Warning signs:**
- Priority color applied to the entire todo line (overrides completed styling)
- No visual distinction between "done P1" and "undone P1" at a glance
- Style combinations growing exponentially (priority x done x selected)

**Phase to address:** The todo rendering phase. Must decide the visual language before implementing.

---

### Pitfall 10: NL Date Parser Dependency Adds Heavy Transitive Dependencies

**What goes wrong:** The project currently has a lean dependency tree (Bubble Tea, lipgloss, TOML, cal, sqlite). Adding `markusmobius/go-dateparser` brings in 200+ locale data files and multiple transitive dependencies. Even `tj/go-naturaldate` is relatively light but may not handle month/year expressions needed for the precision system.

**Why it happens:** NL date parsing for "real" multi-language support is inherently heavy. But this app has a specific, narrow need: English-language relative dates and month/year expressions, not full i18n date parsing.

**Prevention:**

1. **For this app, build a minimal custom parser** that handles the specific expressions needed:
   - Relative: "today", "tomorrow", "yesterday", "next week", "in 3 days"
   - Named days: "monday", "next friday", "this sunday"
   - Named months: "march", "next march", "march 2026"
   - Year: "2027", "next year"
   - Passthrough: existing format strings ("2026-02-15", "15.02.2026", "02/15/2026")

   This is ~200 lines of Go with regex patterns. It avoids any external dependency and integrates perfectly with the precision system because you control the parsing.

2. **If using a library, choose `tj/go-naturaldate`** -- it is lightweight (no locale data), handles relative expressions well, and is actively maintained. BUT wrap it with precision detection (see Pitfall 2). Do not use `markusmobius/go-dateparser` -- its dependency footprint is inappropriate for a TUI tool.

3. **If using `olebedev/when`** -- it returns matched text ranges which can help with precision detection, but its rule-based system is more complex to configure. Good for extensibility, overkill for a personal app.

**Warning signs:**
- `go.sum` grows by 50+ entries after adding the NL parser
- Build time increases noticeably
- Binary size jumps (from ~15MB to ~25MB+)

**Detection:** Run `go mod tidy && wc -l go.sum` before and after adding the dependency. If it grew by more than 20 lines, the dependency is too heavy.

**Phase to address:** The NL parser selection phase. Evaluate dependency weight before committing.

---

## Minor Pitfalls

### Pitfall 11: Priority Badge Misaligns Todo List Columns

**What goes wrong:** Adding a `[P1]` badge before or after the checkbox shifts all text rightward by 4-5 characters. Todos without priority (legacy or unset) do not have this badge, creating ragged alignment in the todo list.

**Prevention:** Always render a fixed-width priority slot. If no priority, render spaces:
```
  [x] [P1] Buy groceries    2026-02-15
  [ ]      Write report      2026-02-16
  [ ] [P3] Clean kitchen     2026-02-17
```
The 5-char slot `[P1] ` or `     ` keeps alignment consistent.

**Phase to address:** Todo rendering phase.

---

### Pitfall 12: Settings Overlay Needs Priority Default Option

**What goes wrong:** Users expect to set a default priority for new todos (e.g., "all new todos start as P3"). Without this, every new todo is P0 (no priority) and the user must manually set priority on each one, which defeats the purpose.

**Prevention:** Add a "Default Priority" option in the settings overlay with values: None, P1, P2, P3, P4. Store in config.toml. Use in `saveAdd()` when creating new todos.

**Where it integrates:**
- `config/config.go:13-20` -- Config struct needs `DefaultPriority`
- `settings/model.go:44-80` -- New() needs new option row
- `todolist/model.go:852-892` -- saveAdd() uses default priority

**Phase to address:** Settings integration phase (after core priority works).

---

### Pitfall 13: NL Parser and Locale Date Format Conflict

**What goes wrong:** The app supports 3 date formats: ISO (2026-02-15), EU (15.02.2026), US (02/15/2026). If the NL parser interprets "3/4" as March 4 but the user's format is EU (day/month), they meant April 3. Numeric date fragments in NL input inherit the format ambiguity.

**Prevention:** When the NL input contains numeric-only date fragments (like "3/4" or "3-4"), use the app's configured `dateFormat` to disambiguate:
- ISO/US: month/day (3/4 = March 4)
- EU: day/month (3/4 = April 3)

For named expressions ("March 4", "tomorrow"), there is no ambiguity. Only numeric fragments need this disambiguation.

**Phase to address:** NL parser integration phase.

---

### Pitfall 14: Priority Field in Edit Form Adds a Fifth Tab Stop

**What goes wrong:** The edit form currently has Title -> Date -> Body -> Template (4 fields). Adding Priority creates Title -> Date -> Priority -> Body -> Template (5 fields). The Tab cycling code has hardcoded field count and index checks throughout `updateInputMode()` and `updateEditMode()`.

**Prevention:** Add Priority as a simple cycling field (like settings options: press h/l to cycle P1-P4-None). It should NOT be a text input. It should be positioned after Date and before Body. Update all `editField` index references.

**Phase to address:** The edit form integration phase.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation | Severity |
|-------------|---------------|------------|----------|
| Schema migration | Priority default breaks existing data | Use 0=none, handle in sort queries | Critical |
| Schema migration | Migration version must be 7 | Follow existing `if version < 7` pattern | Minor |
| Store interface | Add/Update signatures need priority | Expand signatures, update all call sites | Moderate |
| Store interface | Priority sort vs manual sort conflict | Priority as visual indicator, not sort override | Critical |
| Theme colors | 4 priorities x 4 themes = 16 new colors | Map to existing semantic roles where possible | Moderate |
| Theme colors | Priority red clashes with holiday red | Reuse PendingFg for P1 | Moderate |
| Todo rendering | Priority badge misaligns columns | Fixed-width priority slot | Minor |
| Todo rendering | Done + prioritized visual conflict | Badge colored, text follows done styling | Moderate |
| Calendar indicators | No priority data in existing pipeline | Add HighestPriorityPerDay() method | Moderate |
| Calendar indicators | 4 priority colors in 4-char cell | Show highest priority only | Moderate |
| NL parser selection | Heavy dependencies | Build minimal custom parser or use tj/go-naturaldate | Moderate |
| NL parser precision | Parser discards precision info | Wrapper that detects precision from input text | Critical |
| NL parser ambiguity | Wrong dates from ambiguous input | Show parsed date confirmation before saving | Critical |
| NL input mode | Conflicts with segmented Tab cycling | Single date field in NL mode, toggle to segmented | Moderate |
| NL input + locale | Numeric date fragment format conflict | Use app's dateFormat for disambiguation | Minor |
| Edit form | Priority adds 5th Tab stop | Cycling field, not text input | Minor |
| Settings | No default priority option | Add setting, use in saveAdd() | Minor |

---

## Integration Risks with Existing System

### Risk 1: ORDER BY Clause Changes Affect 7 Query Methods

Adding priority to sort order requires updating SQL in `Todos()`, `TodosForMonth()`, `TodosForDateRange()`, `MonthTodos()`, `YearTodos()`, `FloatingTodos()`, and `SearchTodos()`. Missing even one creates inconsistent sort behavior between views.

**Mitigation:** If priority does NOT affect sort order (recommended -- see Pitfall 1), no SQL changes are needed. If it does, grep for `ORDER BY` in `sqlite.go` and update all 7 queries atomically.

### Risk 2: Todo Struct Change Ripples Through scanTodo

The `todoColumns` constant (`sqlite.go:165`) and `scanTodo()` function (`sqlite.go:168-189`) must both be updated for the new priority column. The column must be added in the correct position in BOTH the constant string and the Scan call, or every query will return corrupted data.

**Mitigation:** Add `priority` at the end of `todoColumns` and at the end of `scanTodo()`. Never insert in the middle of the column list.

### Risk 3: Section-Aware Reorder Boundaries Need Priority Awareness

`updateNormalMode()` (lines 471-496) checks `curItem.section == prevItem.section` for reorder boundaries. If priority grouping is added within sections, reorder must also respect priority boundaries. But if priority is visual-only (recommended), no change is needed here.

**Mitigation:** Keep priority as visual-only, do not create priority sub-sections.

### Risk 4: The RenderGrid Signature Is Already at 12 Parameters

`RenderGrid()` in `grid.go:40` takes 12 parameters. Adding priority data pushes it to 13+. Go does not have named parameters, making this increasingly error-prone.

**Mitigation:** If calendar priority indicators are added, wrap parameters in a `GridConfig` struct. But defer this refactor until it is actually needed -- premature abstraction adds complexity.

### Risk 5: NL Date Input Must Work with All 3 Date Formats

The existing segmented input respects `dateFormat` config (ISO/EU/US) for segment ordering. NL input must also respect this for numeric fragment disambiguation (Pitfall 13) and for displaying the parsed date confirmation. The `SetDateFormat()` method on the todolist model must propagate to the NL parser configuration.

**Mitigation:** Pass the date format to the NL parser wrapper. Named expressions ("tomorrow", "March") are format-independent; only numeric fragments need format awareness.

---

## Sources

- Codebase analysis of `internal/store/sqlite.go` -- sort_order queries, migration pattern, scanTodo, 7 ORDER BY clauses (HIGH confidence)
- Codebase analysis of `internal/store/todo.go` -- Todo struct, precision methods, InMonth/InDateRange (HIGH confidence)
- Codebase analysis of `internal/theme/theme.go` -- 4 theme definitions, 16 semantic roles, color palettes (HIGH confidence)
- Codebase analysis of `internal/calendar/grid.go` -- 4-char cell rendering, indicator style priority chain (HIGH confidence)
- Codebase analysis of `internal/todolist/model.go` -- edit form field cycling, date segment handling, renderTodo (HIGH confidence)
- Codebase analysis of `internal/calendar/styles.go` -- Indicator/IndicatorDone styles (HIGH confidence)
- Codebase analysis of `internal/settings/model.go` -- option cycling pattern (HIGH confidence)
- [tj/go-naturaldate](https://github.com/tj/go-naturaldate) -- lightweight NL parser, Future/Past direction support (MEDIUM confidence)
- [olebedev/when](https://github.com/olebedev/when) -- rule-based parser with match text extraction (MEDIUM confidence)
- [markusmobius/go-dateparser](https://github.com/markusmobius/go-dateparser) -- heavy dependency, Confidence return type (MEDIUM confidence)
- [ijt/go-anytime](https://github.com/ijt/go-anytime) -- range parsing, standard+natural format support (MEDIUM confidence)
- [Todoist sort/group documentation](https://www.todoist.com/help/articles/sort-or-group-tasks-in-todoist-WFWD0hrb) -- auto-sort vs manual reorder mutual exclusivity pattern (MEDIUM confidence)
- [Todoist default sorting order](https://www.todoist.com/help/articles/default-sorting-order-in-todoist-mqmgerY7) -- priority sort defaults to descending (MEDIUM confidence)
- [Reorderable list design patterns](https://www.darins.page/articles/designing-a-reorderable-list-component) -- explicit order handling in UIs (MEDIUM confidence)
- [Colorblind-safe design guide](https://www.smashingmagazine.com/2024/02/designing-for-colorblindness/) -- never rely on color alone, use labels (MEDIUM confidence)
