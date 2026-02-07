# Roadmap: Todo Calendar

## Milestones

- ✅ **v1.0 MVP** - Phases 1-3 (shipped 2026-02-05)
- ✅ **v1.1 Polish & Personalization** - Phases 4-6 (shipped 2026-02-05)
- ✅ **v1.2 Reorder & Settings** - Phases 7-9 (shipped 2026-02-06)
- ✅ **v1.3 Views & Usability** - Phases 10-13 (shipped 2026-02-06)
- ✅ **v1.4 Data & Editing** - Phases 14-16 (shipped 2026-02-06)
- ✅ **v1.5 UX Polish** - Phases 17-19 (shipped 2026-02-07)
- 🚧 **v1.6 Templates & Recurring** - Phases 20-22 (in progress)

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1-3) - SHIPPED 2026-02-05</summary>

### Phase 1: TUI Scaffold
**Goal**: Split-pane terminal layout with navigation
**Plans**: 1 plan

Plans:
- [x] 01-01: Scaffold split-pane TUI with Bubble Tea

### Phase 2: Calendar & Holidays
**Goal**: Monthly calendar with national holiday display
**Plans**: 2 plans

Plans:
- [x] 02-01: Render monthly calendar grid
- [x] 02-02: Integrate holiday highlighting

### Phase 3: Todo CRUD & Persistence
**Goal**: Create, complete, delete todos with JSON persistence
**Plans**: 2 plans

Plans:
- [x] 03-01: Todo list component with add/complete/delete
- [x] 03-02: Atomic JSON persistence

</details>

<details>
<summary>✅ v1.1 Polish & Personalization (Phases 4-6) - SHIPPED 2026-02-05</summary>

### Phase 4: Date Indicators & Editing
**Goal**: Calendar shows pending work, todos are editable
**Plans**: 2 plans

Plans:
- [x] 04-01: Calendar bracket indicators for dates with todos
- [x] 04-02: Todo text and date editing

### Phase 5: First Day of Week
**Goal**: Configurable week start day
**Plans**: 2 plans

Plans:
- [x] 05-01: Config field and calendar grid adjustment
- [x] 05-02: Wire through UI and help bar

### Phase 6: Color Themes
**Goal**: 4 preset themes with semantic color roles
**Plans**: 2 plans

Plans:
- [x] 06-01: Theme system with Styles struct and constructor DI
- [x] 06-02: Wire theme through app layer and main.go

</details>

<details>
<summary>✅ v1.2 Reorder & Settings (Phases 7-9) - SHIPPED 2026-02-06</summary>

### Phase 7: Todo Reordering
**Goal**: Users can arrange todos in their preferred order
**Plans**: 2 plans

Plans:
- [x] 07-01: Add SortOrder field, migration, SwapOrder method, updated sort logic
- [x] 07-02: Wire MoveUp/MoveDown keybindings and help bar integration

### Phase 8: Settings Overlay
**Goal**: Users can configure theme, holiday country, and first day of week from inside the app with live preview
**Plans**: 2 plans

Plans:
- [x] 08-01: Config.Save, theme.Names, settings model, SetTheme methods
- [x] 08-02: Wire settings overlay into app with live preview and save/cancel

### Phase 9: Overview Panel
**Goal**: Calendar panel shows at-a-glance todo counts so users know where work is concentrated
**Plans**: 1 plan

Plans:
- [x] 09-01: Store aggregation methods, overview styles, and calendar overview rendering

</details>

<details>
<summary>✅ v1.3 Views & Usability (Phases 10-13) - SHIPPED 2026-02-06</summary>

### Phase 10: Overview Color Coding
**Goal**: Users see completion progress at a glance in the overview panel
**Plans**: 1 plan

Plans:
- [x] 10-01: Split overview counts with theme-aware color roles

### Phase 11: Date Format Setting
**Goal**: Users see dates in their preferred regional format
**Plans**: 1 plan

Plans:
- [x] 11-01: Date format config, settings integration, and display propagation

### Phase 12: Weekly Calendar View
**Goal**: Users can zoom into a single week for focused daily planning
**Plans**: 1 plan

Plans:
- [x] 12-01: Weekly view mode with toggle, grid rendering, and week navigation

### Phase 13: Search & Filter
**Goal**: Users can find any todo regardless of which month it lives in
**Plans**: 2 plans

Plans:
- [x] 13-01: Inline todo filter with `/` activation and Esc clear
- [x] 13-02: Full-screen search overlay with cross-month results and navigation

</details>

<details>
<summary>✅ v1.4 Data & Editing (Phases 14-16) - SHIPPED 2026-02-06</summary>

### Phase 14: Database Backend
**Goal**: Todos persist reliably in a SQLite database with zero behavior changes for the user
**Plans**: 2 plans

Plans:
- [x] 14-01: Extract TodoStore interface from concrete Store struct
- [x] 14-02: Implement SQLite backend and wire into main.go

### Phase 15: Markdown Templates
**Goal**: Todos support rich markdown bodies created from reusable templates
**Plans**: 3 plans

Plans:
- [x] 15-01: Store foundation: Body field, templates table, template utilities
- [x] 15-02: Preview overlay with glamour rendering and body indicator
- [x] 15-03: Template creation and usage flow with placeholder prompting

### Phase 16: External Editor
**Goal**: Users can edit todo bodies in their preferred terminal editor
**Plans**: 1 plan

Plans:
- [x] 16-01: Editor package, keybinding, and app lifecycle wiring

</details>

<details>
<summary>✅ v1.5 UX Polish (Phases 17-19) - SHIPPED 2026-02-07</summary>

### Phase 17: Visual Polish & Help
**Goal**: The todo pane is easy to scan and the help bar shows only what matters for the current mode
**Plans**: 2 plans

Plans:
- [x] 17-01: Todo pane visual polish (spacing, separators, styled checkboxes)
- [x] 17-02: Mode-aware help bar with ? toggle and dynamic height

### Phase 18: Full-Pane Editing
**Goal**: Adding and editing todos uses a clean, focused full-pane layout instead of cramped inline inputs
**Plans**: 2 plans

Plans:
- [x] 18-01: Full-pane edit infrastructure and single-field views
- [x] 18-02: Simultaneous two-field dated-add flow

### Phase 19: Pre-Built Templates
**Goal**: Users have useful markdown templates available from first launch without needing to create their own
**Plans**: 1 plan

Plans:
- [x] 19-01: Seed 7 pre-built templates via version-3 migration

</details>

### v1.6 Templates & Recurring (In Progress)

**Milestone Goal:** Users can manage templates in a dedicated overlay and attach recurring schedules that auto-create todos on app launch.

#### Phase 20: Template Management Overlay
**Goal**: Users can browse, view, edit, rename, and delete templates in a dedicated full-screen overlay
**Depends on**: Phase 19
**Requirements**: REQ-20, REQ-21, REQ-22, REQ-23, REQ-24, REQ-25
**Success Criteria** (what must be TRUE):
  1. User can press M in normal mode to open a full-screen template list with cursor navigation
  2. Selecting a template shows its raw content (including placeholder syntax) below the list
  3. User can delete a template with d, rename with r (pre-filled input, duplicate name handled), and edit content with e (opens external editor)
  4. Esc closes the overlay and returns to the main view
**New/Modified files**: NEW internal/tmplmgr/ (model.go, keys.go, styles.go), MOD store/store.go (UpdateTemplate), MOD store/sqlite.go, MOD app/model.go (overlay routing), MOD app/keys.go
**Risk**: LOW -- follows established overlay pattern (settings, search, preview)
**Plans**: 2 plans

Plans:
- [x] 20-01-PLAN.md -- Store extension (UpdateTemplate) and tmplmgr overlay package
- [x] 20-02-PLAN.md -- Wire overlay into app.Model with external editor integration

#### Phase 21: Schedule Schema & CRUD
**Goal**: The data layer supports recurring schedule definitions attached to templates with deduplication tracking
**Depends on**: Phase 20
**Requirements**: REQ-26, REQ-27, REQ-28
**Success Criteria** (what must be TRUE):
  1. Migration v4 creates a schedules table with FK CASCADE to templates and columns for cadence_type, cadence_value, and placeholder_defaults (JSON)
  2. Migration v5 adds schedule_id (FK SET NULL) and schedule_date columns to todos table with a unique index for deduplication
  3. ScheduleRule in internal/recurring can parse and match all four cadence types (daily, weekdays, weekly, monthly) including monthly day clamping for short months
  4. TodoStore interface has schedule CRUD methods (Add/List/Delete/Update) plus TodoExistsForSchedule and AddScheduledTodo, implemented in SQLite and stubbed in JSON store
**New/Modified files**: NEW internal/recurring/rule.go + rule_test.go, MOD store/todo.go (Schedule struct, Todo schedule fields), MOD store/store.go (interface), MOD store/sqlite.go (migrations v4+v5, schedule methods)
**Risk**: MEDIUM -- schema design is straightforward but monthly edge cases and migration ordering need careful testing
**Plans**: 2 plans

Plans:
- [x] 21-01-PLAN.md -- ScheduleRule TDD (parse/match all cadence types in internal/recurring)
- [x] 21-02-PLAN.md -- Schedule struct, migrations v4+v5, interface extension, SQLite CRUD

#### Phase 22: Auto-Creation & Schedule UI
**Goal**: Scheduled todos are automatically created on app launch and users can attach/manage schedules from the template overlay
**Depends on**: Phase 21
**Requirements**: REQ-29, REQ-30, REQ-31, REQ-32, REQ-33
**Success Criteria** (what must be TRUE):
  1. User can press S on a template in the overlay to open a schedule picker that cycles cadence types with arrows, toggles weekdays with space, and accepts a day number for monthly
  2. Templates with schedules show a dimmed suffix in the overlay list (e.g., "(daily)", "(Mon/Wed/Fri)", "(15th of month)")
  3. On app launch, recurring.AutoCreate() runs in main.go before the TUI starts, creating todos for matching dates in a rolling 7-day window with deduplication (multiple launches on the same day produce no duplicates)
  4. Auto-created todos display an [R] indicator after the todo text, styled in a muted color
  5. When scheduling a template with placeholders, the user is prompted to fill default values once; auto-created todos use these stored defaults
**New/Modified files**: NEW internal/recurring/generate.go + generate_test.go, MOD main.go (AutoCreate call), MOD internal/tmplmgr/model.go (schedule modes), MOD internal/todolist/view.go ([R] indicator)
**Risk**: HIGH -- integrates phases 20 and 21, schedule picker adds multiple sub-modes, placeholder defaults need a prompting sub-flow

**Plans**: TBD

## Requirement Coverage

| REQ | Phase | Description |
|-----|-------|-------------|
| REQ-20 | 20 | Template management overlay |
| REQ-21 | 20 | Template content preview |
| REQ-22 | 20 | Delete template from overlay |
| REQ-23 | 20 | Rename template |
| REQ-24 | 20 | Edit template content |
| REQ-25 | 20 | Template overlay keybinding (M) |
| REQ-26 | 21 | Schedule schema (migrations v4+v5) |
| REQ-27 | 21 | Schedule rule types (daily/weekdays/weekly/monthly) |
| REQ-28 | 21 | Schedule CRUD in store |
| REQ-29 | 22 | Schedule picker UI |
| REQ-30 | 22 | Schedule display in template list |
| REQ-31 | 22 | Auto-create on app launch |
| REQ-32 | 22 | Recurring todo visual indicator [R] |
| REQ-33 | 22 | Placeholder defaults at schedule creation |

**Coverage: 14/14 requirements mapped.**

## Phase Dependencies

```
Phase 20 (Template Overlay) --> Phase 21 (Schedule Schema) --> Phase 22 (Auto-Creation + UI)
```

Strict chain: each phase depends on the previous. Phase 20 is standalone value (template management). Phase 21 is backend-only (no UI changes). Phase 22 integrates everything.

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. TUI Scaffold | v1.0 | 1/1 | Complete | 2026-02-05 |
| 2. Calendar & Holidays | v1.0 | 2/2 | Complete | 2026-02-05 |
| 3. Todo CRUD & Persistence | v1.0 | 2/2 | Complete | 2026-02-05 |
| 4. Date Indicators & Editing | v1.1 | 2/2 | Complete | 2026-02-05 |
| 5. First Day of Week | v1.1 | 2/2 | Complete | 2026-02-05 |
| 6. Color Themes | v1.1 | 2/2 | Complete | 2026-02-05 |
| 7. Todo Reordering | v1.2 | 2/2 | Complete | 2026-02-06 |
| 8. Settings Overlay | v1.2 | 2/2 | Complete | 2026-02-06 |
| 9. Overview Panel | v1.2 | 1/1 | Complete | 2026-02-06 |
| 10. Overview Color Coding | v1.3 | 1/1 | Complete | 2026-02-06 |
| 11. Date Format Setting | v1.3 | 1/1 | Complete | 2026-02-06 |
| 12. Weekly Calendar View | v1.3 | 1/1 | Complete | 2026-02-06 |
| 13. Search & Filter | v1.3 | 2/2 | Complete | 2026-02-06 |
| 14. Database Backend | v1.4 | 2/2 | Complete | 2026-02-06 |
| 15. Markdown Templates | v1.4 | 3/3 | Complete | 2026-02-06 |
| 16. External Editor | v1.4 | 1/1 | Complete | 2026-02-06 |
| 17. Visual Polish & Help | v1.5 | 2/2 | Complete | 2026-02-07 |
| 18. Full-Pane Editing | v1.5 | 2/2 | Complete | 2026-02-07 |
| 19. Pre-Built Templates | v1.5 | 1/1 | Complete | 2026-02-07 |
| 20. Template Management Overlay | v1.6 | 2/2 | Complete | 2026-02-07 |
| 21. Schedule Schema & CRUD | v1.6 | 2/2 | Complete | 2026-02-07 |
| 22. Auto-Creation & Schedule UI | v1.6 | 0/TBD | Not started | - |
