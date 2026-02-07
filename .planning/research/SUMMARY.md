# Project Research Summary

**Project:** Todo Calendar v1.6 Templates & Recurring
**Domain:** TUI Calendar with template management and recurring todo automation
**Researched:** 2026-02-07
**Confidence:** HIGH

## Executive Summary

v1.6 adds template management and recurring todos to an existing Go Bubble Tea TUI app with proven architecture (16 completed phases, SQLite backend, 3 established overlay patterns). The research confirms this is a **zero new dependencies** milestone -- everything needed (schedule storage, auto-creation logic, overlay UI, date matching) can be built with the existing stack (Go stdlib, SQLite, Bubble Tea, existing overlay patterns).

The recommended approach uses three architectural pillars: (1) a dedicated template management overlay following the established settings/search/preview pattern, (2) a custom schedule representation (daily/weekdays/monthly types with TEXT storage) rather than cron or RRULE libraries, and (3) synchronous auto-creation in main.go before the TUI launches. This avoids complexity from external parsers, background goroutines, or TUI-blocking prompts while delivering the full feature set.

The critical risk is duplicate todo creation. Auto-creation logic that runs on every app launch will create duplicate todos unless deduplication tracking exists from day one. This is addressed via a schedule_instances ledger table with a PRIMARY KEY(schedule_id, date) constraint. Secondary risks include monthly edge cases (day 31 in February), placeholder prompting blocking startup, and template deletion orphaning schedules. All are addressable with upfront schema design and following the established overlay pattern.

## Key Findings

### Recommended Stack

**Zero new external dependencies.** Every capability needed for v1.6 exists in the current stack. The key decisions are about data modeling and integration patterns, not library selection.

**Core technologies:**
- **Go stdlib**: time.Weekday() for day matching, encoding/json for placeholder storage, strconv for monthly day parsing -- all existing dependencies
- **SQLite (modernc.org/sqlite v1.44.3)**: New schedules and schedule_instances tables, foreign key CASCADE enforcement (already enabled via DSN)
- **Bubble Tea overlay pattern**: Template management follows the exact pattern from settings/search/preview (showX bool, dedicated package, CloseMsg routing)
- **Custom schedule format**: Structured TEXT columns (cadence_type, cadence_value) instead of cron or RRULE -- simpler, debuggable, sufficient for three cadence types

**Deliberately NOT adding:**
- robfig/cron: Cron expressions encode time-of-day precision irrelevant to a date-level todo app
- teambition/rrule-go: RRULE library is inactive (Snyk advisory), massive overkill for 3 cadence types
- Background goroutines/tickers: Auto-creation runs once at startup, no need for concurrent complexity
- Separate ORM: Hand-written SQL is the established pattern

### Expected Features

**Must have (table stakes):**
- **Template CRUD overlay**: Full-screen overlay for listing, viewing, editing, renaming, deleting templates (users can create templates but cannot manage them except during inline selection)
- **Schedule attachment to templates**: Recurring rules (daily, specific weekdays, monthly Nth) stored on templates
- **Auto-creation on app launch**: Rolling 7-day window creates missing scheduled todos synchronously before TUI starts
- **Deduplication**: Multiple app launches per day must NOT create duplicate todos (PRIMARY KEY constraint on ledger table)
- **Visual indicator**: [R] marker on auto-created todos distinguishes them from manual todos

**Should have (differentiators):**
- **Template content preview**: Show full template content when browsing the overlay (uses existing Glamour renderer)
- **External editor for template content**: Reuse Phase 16 external editor pattern for multi-line editing
- **Placeholder pre-fill at schedule creation**: User fills placeholders once when creating schedule, not on every occurrence
- **Schedule display in template list**: Show cadence ("every weekday", "Mon/Wed/Fri") next to template name

**Defer (v2+):**
- Template reordering: Needs sort_order column, matches todo reorder pattern but not essential
- Complex cadences: "every 2nd Tuesday", "last Friday of month" -- explicitly out of scope per milestone definition
- Pause/resume schedule: Needs enabled column, useful but can be added later
- Natural language date parsing: Todoist's competitive advantage, not worth the NLP complexity
- Notification/reminder system: TUI has no daemon infrastructure

### Architecture Approach

v1.6 integrates cleanly into the existing architecture via three new capabilities built on established patterns: a template management overlay (new internal/tmplmgr package following settings/search/preview structure), schedule CRUD in the store layer (new methods on TodoStore interface with SQLite implementation), and auto-creation orchestration in a pure business logic package (internal/recurring) called from main.go before the TUI launches.

**Major components:**

1. **internal/tmplmgr** (new package): Template management overlay with Model/Update/View, four view modes (list/view/edit/rename), own keybindings and styles, emits CloseMsg to dismiss. Integrates into app.Model via showTmplMgr bool and dedicated updateTmplMgr() routing method (position 6.5 in overlay priority order). Uses existing bubbles textinput/textarea, Glamour for preview, external editor for content editing.

2. **Schedule storage (SQLite)**: Two new tables added via migrations v4 and v5. schedules table (template_id FK with CASCADE, cadence_type TEXT, cadence_value TEXT, placeholder_defaults JSON, enabled flag). schedule_instances ledger table (schedule_id+date PRIMARY KEY for deduplication, todo_id FK for lineage). TodoStore interface gains 8 new methods: UpdateTemplate, AddSchedule, ListSchedules, FindSchedule, UpdateSchedule, DeleteSchedule, TodoExistsForSchedule, CreateScheduledTodos.

3. **internal/recurring** (new package): Pure business logic for schedule rule parsing and auto-creation orchestration. ScheduleRule struct with ParseRule/MatchesDate methods handles three cadence types. AutoCreate() function called from main.go after store initialization but before app.New(), iterates schedules, checks matches for today+7 days, calls store to create missing todos with dedup checks. No TUI dependencies.

**Data flow:**
```
App Launch -> config.Load() -> store.NewSQLiteStore() (runs migrations v4+v5)
  -> recurring.AutoCreate(store, time.Now(), 7)  // creates scheduled todos
  -> app.New() (TUI sees fully populated dataset)
  -> tea.NewProgram().Run()
```

**Integration points:**
- app.Model gains showTmplMgr, tmplMgr fields, updateTmplMgr() routing at priority 6.5
- TodoStore interface extended with 8 methods (SQLite implements, JSON stubs)
- Theme propagation: m.tmplMgr.SetTheme(t) added to applyTheme()
- Help bar: m.tmplMgr.HelpBindings() routed when overlay active
- Key binding: New key (M for "manage templates") in app.KeyMap

### Critical Pitfalls

1. **Duplicate Todo Creation on Repeated Launch** (CRITICAL): Auto-creation without deduplication creates duplicate todos every time the app launches. Prevention: schedule_instances ledger table with PRIMARY KEY(schedule_id, date). Before creating a todo for schedule S on date D, check if row exists. The constraint makes this idempotent at the database level. Must be in the schema from day one.

2. **Monthly Day Edge Cases** (CRITICAL): "Monthly on the 31st" silently fails in February/April/June/September/November. Prevention: Clamp to last day of month when schedule's day-of-month exceeds month length. Do NOT use SQLite date arithmetic (ceiling behavior wraps to next month). Use Go's own date logic. Requires explicit test cases for all short months.

3. **Template Deletion Orphans Schedules** (CRITICAL): Deleting a template without CASCADE leaves orphaned schedule rows. Prevention: Declare FK with ON DELETE CASCADE in schedules table schema: `template_id INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE`. App already has foreign_keys(ON) in DSN. Also cascade to schedule_instances ledger so dedup state is cleaned.

4. **Placeholder Prompting Blocks Startup** (CRITICAL): Batch auto-creation tries to prompt for placeholders before TUI renders, causing hang or skipped prompts. Prevention: Pre-fill placeholder values at schedule creation time, store as JSON in schedules.placeholder_defaults column. Auto-created todos use these stored values. "{{.Date}}" auto-fills with scheduled date. Other placeholders use stored values. Keeps auto-creation non-interactive.

5. **Rolling Window Creates Future Todo Flood** (MODERATE): 7-day window creates todos for days the user hasn't navigated to yet, cluttering current view. Mitigation: Start with today-only or today+1, make window configurable. Simpler for personal use. Can expand if users request longer horizons.

## Implications for Roadmap

Based on research, suggested **3-phase structure** with strict dependencies:

### Phase 1: Template Management Overlay (Foundation)

**Rationale:** Standalone feature with no dependency on recurring/scheduling. Provides immediate value (proper template management UI) and establishes the UI infrastructure that schedule management will integrate into. Lowest risk -- follows established overlay pattern exactly.

**Delivers:**
- Full-screen template list with content preview
- Delete/rename/edit operations
- External editor integration for content editing
- Keybinding to open overlay (M key)

**Addresses:**
- Table stakes: template CRUD overlay, content preview, external editor
- Fills gap: users can create templates (T) and use them (t) but cannot manage except during select flow

**Avoids:**
- Pitfall: overlay state leaks (separate package with fresh model on open)
- Pitfall: rename to duplicate name (return error, show in UI)

**Stack elements:**
- Bubble Tea overlay pattern (existing)
- bubbles textinput/textarea (existing)
- Glamour for preview (existing)
- store.UpdateTemplate (new method)

**New/Modified:**
- NEW: internal/tmplmgr/model.go, keys.go, styles.go
- MOD: store/store.go (add UpdateTemplate to interface)
- MOD: store/sqlite.go (implement UpdateTemplate)
- MOD: app/model.go (overlay routing, keybinding)

**Research needed:** NO (well-documented overlay pattern, 3 existing examples in codebase)

### Phase 2: Schedule Schema + CRUD

**Rationale:** Schema and data layer must exist before either the schedule management UI or auto-creation engine can work. This is all backend -- no UI changes. Establishes the foundation for both Phase 1 (schedule UI in overlay) and Phase 3 (auto-creation).

**Delivers:**
- schedules table (migration v4) with FK CASCADE
- schedule_instances ledger table (migration v4, PRIMARY KEY deduplication)
- ScheduleRule parsing (daily/weekdays/monthly types)
- Schedule CRUD methods on TodoStore
- Date matching logic with monthly clamping

**Uses:**
- Custom schedule format (cadence_type + cadence_value TEXT columns)
- Go stdlib time.Weekday() for matching
- SQLite foreign keys (already enabled)

**Implements:**
- Schedule storage architecture component
- internal/recurring/rule.go (parsing + matching)

**Avoids:**
- Pitfall: duplicate creation (ledger table with PRIMARY KEY from start)
- Pitfall: monthly day > days-in-month (clamp to last day)
- Pitfall: template deletion orphans schedules (FK CASCADE in DDL)
- Pitfall: missing schema fields (include enabled, placeholder_defaults upfront)

**New/Modified:**
- NEW: internal/recurring/rule.go, rule_test.go
- MOD: store/todo.go (Schedule struct, Todo.ScheduleID/SourceDate fields)
- MOD: store/store.go (add 6 schedule methods to interface)
- MOD: store/sqlite.go (migrations v4+v5, implement schedule methods)

**Research needed:** NO (straightforward schema design, date logic is stdlib)

### Phase 3: Auto-Creation + Schedule UI Integration

**Rationale:** Depends on both the overlay (Phase 1) and the schema/CRUD (Phase 2). This is where everything comes together. Schedule management modes added to tmplmgr overlay, auto-creation wired into main.go startup.

**Delivers:**
- Auto-creation on app launch (rolling 7-day window)
- Schedule management UI in template overlay (view/add/delete schedules)
- Visual indicator [R] on auto-created todos
- Schedule display in template list ("every weekday")
- Placeholder pre-fill at schedule creation

**Integrates:**
- recurring.AutoCreate() called in main.go after store init, before app.New()
- Schedule picker sub-modes in tmplmgr overlay
- Template overlay shows schedules for selected template

**Addresses:**
- Table stakes: auto-creation on launch, deduplication, visual indicator, schedule attachment
- Differentiator: placeholder pre-fill, schedule display

**Avoids:**
- Pitfall: placeholder prompting blocking startup (pre-fill at schedule creation, store in JSON)
- Pitfall: too many future todos (start with today-only, configurable window)
- Pitfall: weekday numbering mismatch (store Go time.Weekday values, explicit display mapping)

**New/Modified:**
- NEW: internal/recurring/generate.go, generate_test.go
- MOD: main.go (add auto-creation call)
- MOD: internal/tmplmgr/model.go (schedule management modes)
- MOD: internal/todolist/view.go (render [R] indicator)

**Research needed:** NO (integration testing needed, but patterns are established)

### Phase Ordering Rationale

- **Strict dependency chain:** Phase 1 (overlay) -> Phase 2 (schema) -> Phase 3 (auto-creation + schedule UI). Template management must exist before schedules can be managed in the overlay. Schedule schema must exist before auto-creation can use it.
- **Risk minimization:** Phase 1 is lowest risk (established pattern). Phase 2 is backend-only (testable in isolation). Phase 3 integrates everything (highest complexity, deferred to last).
- **Incremental value:** Phase 1 ships standalone value (template management). Phase 2 enables Phase 3. Phase 3 delivers the full recurring system.
- **Pitfall avoidance:** Critical pitfalls (duplication, FK cascade, placeholder blocking) are addressed in Phases 2 and 3 via upfront schema design and pre-fill strategy.

### Research Flags

**No phases need deeper research.** All three phases use established patterns:

- **Phase 1:** Overlay pattern has 3 existing examples (settings, search, preview). External editor pattern proven in Phase 16. Template CRUD is straightforward SQLite.
- **Phase 2:** Schema migrations follow existing PRAGMA user_version pattern (currently at v3). Schedule rule parsing is simple string parsing. Date matching is stdlib time.Weekday().
- **Phase 3:** Auto-creation is a linear scan with dedup checks. Schedule UI adds modes to the overlay (same pattern as todolist's 10 modes). Visual indicator is a string suffix.

**Standard patterns (skip research-phase):** All phases. The architecture research confirmed every component maps to an existing pattern or stdlib capability.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Zero new dependencies, all capabilities exist in current stack. Custom schedule format simpler than cron/RRULE. |
| Features | HIGH (overlay), MEDIUM (recurring) | Template CRUD overlay is well-understood (3 existing overlays to follow). Recurring todos novel for this app but patterns well-established in Taskwarrior/Todoist/Things. |
| Architecture | HIGH | Integration points verified against existing codebase. Overlay pattern proven 3x. Store pattern established (16 phases). main.go startup hook is straightforward. |
| Pitfalls | HIGH | Critical pitfalls identified with database-level solutions (PRIMARY KEY, FK CASCADE). Monthly edge cases documented with stdlib solution (clamp to last day). |

**Overall confidence:** HIGH

v1.6 is architecturally conservative -- it extends existing patterns (overlay, store interface, migrations) rather than introducing new ones. The three-phase roadmap has clear dependencies and incremental value. The critical risks (duplicate creation, monthly edge cases, FK cascade) have upfront database-level solutions.

### Gaps to Address

**1. Rolling window size validation:** Research recommends starting with today-only or today+1 instead of 7 days to avoid future todo flood. This needs UX validation during Phase 3 implementation. Start conservative, expand if users request it.

**2. Placeholder pre-fill UX:** The placeholder_defaults JSON approach avoids blocking startup but requires UI for filling values at schedule creation time. The overlay needs a "configure placeholders" sub-mode when adding a schedule to a template with {{.Var}} fields. This is straightforward (reuse existing placeholder extraction from tmpl package) but wasn't detailed in research.

**3. Weekday numbering with first_day_of_week:** The app supports both Sunday-start and Monday-start weeks (config setting). Schedule UI must map display order (Mon/Tue/Wed if Monday-start) to storage order (Go time.Weekday: Sunday=0, Monday=1). This requires explicit mapping layer in Phase 3. Test with both week start configurations.

**4. Template deletion confirmation with schedules:** When deleting a template that has schedules, the overlay should warn the user that schedules will cascade-delete. This is a UX polish item for Phase 1 (requires checking for schedules before delete).

## Sources

### Primary (HIGH confidence)
- **Existing codebase analysis:** internal/app/model.go (overlay routing), internal/settings/model.go (overlay component pattern), internal/store/sqlite.go (migration pattern, FK enforcement), internal/todolist/model.go (template workflow, mode state machine), internal/tmpl/tmpl.go (placeholder extraction)
- **Go stdlib documentation:** time.Weekday, encoding/json, text/template
- **SQLite documentation:** sqlite.org/foreignkeys.html (CASCADE behavior), sqlite.org/lang_datefunc.html (month arithmetic edge cases)
- **PROJECT.md:** Active requirements, v2 candidates (complex cadences deferred)

### Secondary (MEDIUM confidence)
- **Taskwarrior recurrence docs:** taskwarrior.org/docs/recurrence/ (template/instance model, mask tracking)
- **Todoist recurring dates docs:** todoist.com/help/articles/introduction-to-recurring-dates (supported patterns, visual scheduler)
- **Things repeating to-dos docs:** culturedcode.com/things/support/articles/2803564/ (template-based repeating, pause/resume)
- **Recurring event database patterns:** red-gate.com/blog/managing-recurring-events-in-a-data-model (schema patterns)
- **RFC 5545 RRULE spec:** icalendar.org (reviewed to confirm overkill for v1.6)

### Tertiary (LOW confidence)
- **teambition/rrule-go GitHub:** Snyk advisory, inactive maintenance status
- **stephens2424/rrule GitHub:** 17 stars, author notes no production usage
- **robfig/cron pkg.go.dev:** API review (time-of-day focus irrelevant for date-level todos)

---
*Research completed: 2026-02-07*
*Ready for roadmap: yes*
