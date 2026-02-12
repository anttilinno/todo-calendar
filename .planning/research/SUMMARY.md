# Project Research Summary

**Project:** todo-calendar v2.1 - Priority Levels & Natural Language Date Input
**Domain:** TUI enhancement for existing Go/Bubble Tea todo application
**Researched:** 2026-02-12
**Confidence:** HIGH

## Executive Summary

This research covers adding two features to an existing 8,177 LOC Go/Bubble Tea TUI todo-calendar app: P1-P4 priority levels and natural language date input ("tomorrow", "next friday"). Both features integrate into an established codebase with well-defined patterns for schema migrations, edit form handling, theme management, and date precision (day/month/year).

The recommended approach is to treat priority as a visual indicator (colored badge + label) without affecting manual sort order, and to replace the existing 3-segment date input (dd/mm/yyyy) with a single textinput that accepts both natural language and formatted dates. Priority requires a schema migration (v7), theme color additions, and edit form integration. Natural language parsing requires a precision-aware wrapper around `tj/go-naturaldate` to preserve the existing date precision system.

Key risks include priority auto-sort conflicting with manual J/K reordering (mitigation: visual-only priority), NL parsers discarding precision information (mitigation: detect precision from input text before parsing), and ambiguous date interpretations (mitigation: show parsed date confirmation). Both features are additive and can be built incrementally without breaking existing functionality.

## Key Findings

### Recommended Stack

**Core technologies already in use:**
- Go 1.24 with Bubble Tea TUI framework
- SQLite with WAL mode, schema at PRAGMA user_version=6
- Lipgloss for styling with 4 themes (Dark/Light/Nord/Solarized)
- TOML config (6 settings currently)

**New dependency for NL date:**
- `github.com/tj/go-naturaldate` (313 stars, MIT, zero transitive deps) — lightweight parser with Future/Past direction support, handles "tomorrow", "next friday", "in 2 weeks", "jan 15"

**Why this parser:** Pure Go, minimal API (single Parse function), no locale data bloat. Alternatives rejected: `olebedev/when` (too heavy, complex rule system), `markusmobius/go-dateparser` (200+ locale files, inappropriate for personal TUI), `araddon/dateparse` (format parsing only, not natural language).

### Expected Features

**Priority levels (P1-P4):**
- Visual indicator (colored badge + text label) on each todo
- Set during add/edit via dedicated form field
- Display in todo list, search results, and optionally calendar indicators
- Theme-aware colors (4 priority colors × 4 themes = 16 color values)
- Integer storage (0=none, 1=P1, 2=P2, 3=P3, 4=P4) for sortability

**Natural language date input:**
- Single textinput replacing 3-segment (dd/mm/yyyy) date fields
- Accepts NL expressions: "tomorrow", "next friday", "jan 15", "in 2 weeks"
- Accepts formatted dates: "2026-02-15" (ISO), "15.02.2026" (EU), "02/15/2026" (US)
- Preserves date precision system (day/month/year) via input pattern detection
- Shows parsed date confirmation before saving

**Defer (future milestones):**
- Priority-based auto-sort (conflicts with manual reordering, needs separate view mode)
- Inline priority toggling in normal mode (accidental changes risk, edit form is safer)
- Calendar grid priority colors (nice-to-have, complex integration)

### Architecture Approach

The existing architecture has a root `app.Model` orchestrating 6 overlays (calendar, todolist, settings, search, preview, template manager) with a SQLite store implementing the `TodoStore` interface. Priority and NL date integrate as follows:

**Major components:**

1. **Store layer** — Add `priority INTEGER NOT NULL DEFAULT 0` column via migration v7, extend `todoColumns` constant, update `scanTodo()` to read priority, expand `Add()` and `Update()` signatures to accept priority parameter

2. **Theme layer** — Add 4 priority color fields to `Theme` struct (PriorityP1/P2/P3/P4), map to existing semantic colors where possible (P1→PendingFg red, P2→IndicatorFg orange, P3→AccentFg blue, P4→MutedFg grey) to minimize palette expansion

3. **Todolist model** — Add `editPriority int` field, insert priority as field 2 in edit form (Title→Date→Priority→Body→Template), render priority badge in `renderTodo()` before todo text, pass priority to store on save

4. **NL date package** (new: `internal/nldate/`) — Pure function `Parse(input, ref, userLayout) (isoDate, precision, error)` that detects precision from input patterns BEFORE calling `go-naturaldate`, handles year-only ("2026"), month-year ("feb 2026"), ISO dates, and NL expressions, returns day precision for all NL dates

5. **Edit form refactor** — Replace 3 date segment textinputs (day/month/year) with single `dateInput` textinput, remove ~200 lines of segment cycling/auto-advance logic, replace `deriveDateFromSegments()` with `nldate.Parse()`

**Data flow:** User types "tomorrow" in date field → Tab triggers parse → `nldate.Parse()` returns ("2026-02-13", "day", nil) → User types "2" in priority field → Enter saves → `store.Add(text, date, precision, priority)` → INSERT with priority=2 → Display shows `[P2]` badge in theme's P2 color

### Critical Pitfalls

1. **Priority auto-sort destroys manual reorder state** — The existing system uses `sort_order` with manual J/K reordering. Adding `ORDER BY priority` means users cannot move P3 above P1. **Avoidance:** Priority is visual-only (badge + color), does NOT affect SQL ORDER BY or manual J/K reordering. Auto-sort can be a separate view mode in future.

2. **NL parser discards date precision** — All Go NL parsers return `time.Time` (always day-precise). Parsing "March" or "2027" loses month/year precision, breaking section assignment. **Avoidance:** Build a wrapper that detects precision from input text BEFORE parsing (regex for year-only, month-only patterns), then delegates to `go-naturaldate`.

3. **Ambiguous NL dates parsed wrong** — "Friday" could mean last Friday or next Friday depending on parser direction. Users have no confirmation. **Avoidance:** Configure parser with Future direction for todo app, show parsed date confirmation ("Parsed: Friday, February 14, 2026") before saving.

4. **Schema migration default value breaks semantics** — Adding `priority INTEGER NOT NULL DEFAULT 0` works, but 0 must mean "no priority" not "P1". **Avoidance:** 0=none, 1=P1, 2=P2, 3=P3, 4=P4. Existing todos get 0 (unset), which sorts after P4 if priority sorting is ever added.

5. **Theme color palette exhaustion** — Adding 4 new priority colors per theme (16 total) may conflict with existing reds (HolidayFg, PendingFg), yellows (IndicatorFg), greys (MutedFg, CompletedFg). **Avoidance:** Map priorities to existing semantic colors where possible (reuse PendingFg for P1, IndicatorFg for P2, AccentFg for P3, MutedFg for P4). Add text labels `[P1]` so color is not the only indicator (accessibility).

## Implications for Roadmap

Based on research, suggested 3-phase structure with clear dependency ordering:

### Phase 1: Priority Data Layer

**Rationale:** Schema migration must exist before any UI can read/write priority. This phase is backend-only (store + schema), easily testable in isolation, and establishes the data contract for all downstream phases.

**Delivers:**
- Migration v7: `ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0`
- Extended `Todo` struct with `Priority int` field and helper methods (`PriorityLabel()`, `HasPriority()`)
- Updated `todoColumns` constant and `scanTodo()` function
- Expanded `Add()` and `Update()` interface signatures with priority parameter
- All store tests passing with priority roundtrip

**Addresses features:**
- Priority storage and retrieval (prerequisite for UI)

**Avoids pitfalls:**
- Pitfall 4 (migration default semantics) — 0=none established early
- Pitfall 8 (store interface changes) — all signatures updated atomically

**Research flag:** Standard pattern (follows v6 migration for date_precision), no additional research needed.

---

### Phase 2: Priority UI + Theme

**Rationale:** Depends on Phase 1 (store must read/write priority). Self-contained UI feature that does not affect date input. Can be tested independently with manual priority entry.

**Delivers:**
- 4 priority color fields in `Theme` struct, values for all 4 themes
- 4 priority styles in `todolist.Styles`
- Priority field in edit form (position 2, between date and body)
- Priority badge in `renderTodo()` before todo text
- Priority display in search results
- Priority input handling (keypress 1-4 to set, 0/backspace to clear)

**Addresses features:**
- Priority visual indicator (colored badge + label)
- Priority setting during add/edit

**Avoids pitfalls:**
- Pitfall 1 (auto-sort conflicts) — priority is visual-only, no ORDER BY changes
- Pitfall 5 (color palette) — map to existing semantic colors where feasible
- Pitfall 9 (completed priority confusion) — badge retains color, text gets CompletedFg
- Pitfall 11 (alignment) — fixed-width priority slot (5 chars: `[P1] ` or `     `)

**Research flag:** Standard form integration pattern, no additional research needed. Theme colors need manual tuning but follow existing 4-theme structure.

---

### Phase 3: Natural Language Date Input

**Rationale:** Independent of priority (could be parallel), but placing it last means priority form changes are stable before refactoring date input. The date input replacement is the largest refactor, touching edit form rendering, Tab cycling, and removing ~200 lines of segment logic.

**Delivers:**
- New package `internal/nldate/` with `Parse()` function
- Precision-aware wrapper around `tj/go-naturaldate`
- Single `dateInput` textinput replacing 3 date segments
- Multi-strategy date parsing (year-only, month-year, ISO, NL)
- Removal of `dateSegDay/Month/Year`, `dateSegFocus`, `renderDateSegments()`, `deriveDateFromSegments()`, and all segment helpers
- Comprehensive tests for NL parsing edge cases

**Addresses features:**
- Natural language date input ("tomorrow", "next friday")
- Simplified date entry (1 field instead of 3)
- Format-aware parsing (respects ISO/EU/US config)

**Avoids pitfalls:**
- Pitfall 2 (precision loss) — wrapper detects precision from input text before parsing
- Pitfall 3 (ambiguous dates) — Future direction configured, parsed date shown before save
- Pitfall 7 (field cycling conflicts) — single textinput, no segment focus state
- Pitfall 10 (heavy dependencies) — `tj/go-naturaldate` is lightweight (zero transitive deps)
- Pitfall 13 (locale format) — numeric fragments disambiguated via user's dateFormat setting

**Research flag:** Needs thorough testing of NL parser edge cases (month-only, year-only, relative expressions). The precision detection wrapper is novel for this codebase and requires validation with diverse inputs.

---

### Phase Ordering Rationale

- **Phase 1 first:** Schema changes must precede UI. Migration v7 establishes priority storage contract.
- **Phase 2 before Phase 3:** Priority adds one field to edit form (simple). NL date replaces an entire subsystem (complex). Staging complexity reduces merge conflicts.
- **Phases 2 and 3 could be parallel:** No shared code beyond the edit form field count. But sequential is safer given that both touch edit form Tab cycling.
- **No calendar indicator phase:** Calendar priority colors (Pitfall 6) are deferred. Requires new store method `HighestPriorityPerDay()` and `RenderGrid()` parameter expansion. Not essential for v2.1.

### Research Flags

**Needs testing validation:**
- **Phase 3:** NL date precision detection wrapper (novel design, needs edge case validation: "March", "2027", "next month", "in 2 weeks")

**Standard patterns (no additional research):**
- **Phase 1:** Schema migration follows established v1-v6 pattern
- **Phase 2:** Edit form field addition and theme color extension follow existing patterns

**Deferred for future research:**
- Calendar indicator priority coloring (moderate complexity, unclear UX value)
- Priority-based auto-sort as separate view mode (requires sort strategy research)
- Inline priority toggling (needs keybinding conflict analysis)

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Codebase patterns well-established (8,177 LOC reviewed), migration pattern verified across v1-v6, `tj/go-naturaldate` API confirmed via pkg.go.dev |
| Features | HIGH | Priority visual indicator and NL date input are well-scoped, implementation paths clear from existing edit form and theme systems |
| Architecture | HIGH | Component boundaries identified via deep codebase analysis, integration points mapped (store interface, theme struct, edit form fields, render methods) |
| Pitfalls | HIGH | All 14 pitfalls derived from codebase analysis (specific line numbers cited), validated against existing patterns (sort_order, date_precision, migration v6) |

**Overall confidence:** HIGH

This is not a new project but an enhancement to an existing, stable codebase with well-established patterns. The risk profile is low because:
1. Both features are additive (no removal of existing functionality)
2. Schema migration follows proven pattern (v6 added date_precision identically)
3. NL parser is isolated in a pure function (easily testable)
4. Priority is visual-only (does not break sort logic)

### Gaps to Address

**NL date precision detection edge cases:**
- How to handle "march" vs "march 15" (month-precision vs day-precision)?
- Does "next month" mean month-precision or day-precision (same day next month)?
- Can year-only ("2027") be reliably distinguished from numeric ID or typo?

**Mitigation:** Comprehensive unit tests in `nldate_test.go` covering all input patterns (relative, named months, year-only, ISO dates). Add integration test: type each pattern in TUI, verify section assignment (Dated/This Month/This Year).

**Theme color tuning:**
- Do mapped priority colors (P1→PendingFg, P2→IndicatorFg) create confusion with existing indicators?
- Are P3/P4 distinguishable enough from MutedFg and AccentFg in all 4 themes?

**Mitigation:** Manual testing in all 4 themes. Create todos at all 4 priority levels, view alongside holidays, completed todos, and section headers. Adjust colors if collisions occur.

**Edit form field count:**
- Does adding priority (5th field) make Tab cycling too long?
- Should Body textarea remain the second-to-last field for quick access?

**Mitigation:** Use priority during dogfooding. If Tab cycling feels tedious, consider reordering fields or adding Shift+Tab backward cycling.

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `internal/store/sqlite.go` — 27 TodoStore methods, migration v1-v6 pattern, sort_order queries, todoColumns/scanTodo (lines 54-189)
- Codebase analysis: `internal/store/todo.go` — Todo struct (10 fields), precision methods (IsMonthPrecision/IsYearPrecision), InMonth/InDateRange logic
- Codebase analysis: `internal/todolist/model.go` — Edit form (4 fields, Tab cycling lines 666-780), 3-segment date input (~200 lines), renderTodo (lines 1054-1093), save methods (saveAdd line 852, saveEdit line 814)
- Codebase analysis: `internal/theme/theme.go` — 16 semantic color roles, 4 theme constructors (Dark/Light/Nord/Solarized), palette analysis
- Codebase analysis: `internal/calendar/grid.go` — 4-char cell rendering, indicator style priority chain (lines 144-183), RenderGrid 12-parameter signature
- [tj/go-naturaldate pkg.go.dev](https://pkg.go.dev/github.com/tj/go-naturaldate) — Parse() API, WithDirection option, zero dependencies

### Secondary (MEDIUM confidence)
- [tj/go-naturaldate GitHub](https://github.com/tj/go-naturaldate) — 313 stars, MIT license, 15 commits, stable
- [olebedev/when](https://github.com/olebedev/when) — evaluated for comparison (1.5k stars, rule-based, heavier)
- [markusmobius/go-dateparser](https://github.com/markusmobius/go-dateparser) — evaluated and rejected (200+ locale files, heavy dependencies)
- [araddon/dateparse](https://github.com/araddon/dateparse) — evaluated and rejected (format parsing only, not NL)

### Tertiary (LOW confidence, informational)
- [Todoist sort documentation](https://www.todoist.com/help/articles/sort-or-group-tasks-in-todoist-WFWD0hrb) — auto-sort vs manual reorder pattern (informational, not prescriptive)
- [Colorblind-safe design guide](https://www.smashingmagazine.com/2024/02/designing-for-colorblindness/) — never rely on color alone (general UX principle)

---
*Research completed: 2026-02-12*
*Ready for roadmap: yes*
