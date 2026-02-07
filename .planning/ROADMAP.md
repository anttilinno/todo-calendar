# Roadmap: Todo Calendar

## Milestones

- ✅ **v1.0 MVP** - Phases 1-3 (shipped 2026-02-05)
- ✅ **v1.1 Polish & Personalization** - Phases 4-6 (shipped 2026-02-05)
- ✅ **v1.2 Reorder & Settings** - Phases 7-9 (shipped 2026-02-06)
- ✅ **v1.3 Views & Usability** - Phases 10-13 (shipped 2026-02-06)
- ✅ **v1.4 Data & Editing** - Phases 14-16 (shipped 2026-02-06)
- 🚧 **v1.5 UX Polish** - Phases 17-19 (in progress)

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

### 🚧 v1.5 UX Polish (In Progress)

**Milestone Goal:** Improve visual clarity, input ergonomics, help discoverability, and out-of-box template experience

#### Phase 17: Visual Polish & Help
**Goal**: The todo pane is easy to scan and the help bar shows only what matters for the current mode
**Depends on**: Phase 16 (builds on existing todolist and help rendering)
**Requirements**: VIS-01, VIS-02, VIS-03, HELP-01, HELP-02, HELP-03
**Success Criteria** (what must be TRUE):
  1. Todo items have visible breathing room -- vertical spacing separates individual items
  2. Section headers (month name, "Floating") stand apart from todo items through separators, padding, or stronger styling
  3. Dates and completion status are visually distinct from todo text (not just inline plaintext)
  4. Normal mode help bar shows at most 5 key bindings instead of the full list
  5. Pressing ? in normal mode reveals the complete keybinding list; input modes show only Enter/Esc
**Plans**: 2 plans

Plans:
- [x] 17-01: Todo pane visual polish (spacing, separators, styled checkboxes)
- [x] 17-02: Mode-aware help bar with ? toggle and dynamic height

#### Phase 18: Full-Pane Editing
**Goal**: Adding and editing todos uses a clean, focused full-pane layout instead of cramped inline inputs
**Depends on**: Phase 17 (VIS layout settled before EDIT takes over the pane)
**Requirements**: EDIT-01, EDIT-02, EDIT-03, EDIT-04, EDIT-05
**Success Criteria** (what must be TRUE):
  1. Pressing "a" to add a todo replaces the todo list with a full-pane input showing a prominent title field
  2. Adding a dated todo shows both title and date fields with clear labels in the full pane
  3. Editing an existing todo (title or date) uses the same full-pane layout with the current value pre-filled
  4. The full-pane edit view shows only minimal contextual help (Enter to confirm, Esc to cancel, Tab to switch fields)
**Plans**: 2 plans

Plans:
- [x] 18-01: Full-pane edit infrastructure and single-field views
- [x] 18-02: Simultaneous two-field dated-add flow

#### Phase 19: Pre-Built Templates
**Goal**: Users have useful markdown templates available from first launch without needing to create their own
**Depends on**: Phase 16 (uses existing template infrastructure)
**Requirements**: TMPL-01, TMPL-02, TMPL-03, TMPL-04
**Success Criteria** (what must be TRUE):
  1. First launch (empty DB) seeds 6-8 templates covering general and dev use cases
  2. General templates include practical items like meeting notes, checklist, and daily plan
  3. Dev templates include practical items like bug report, feature spec, and PR checklist
  4. Users can delete any pre-built template -- none are locked or force-retained
**Plans**: 1 plan

Plans:
- [ ] 19-01: Seed 7 pre-built templates via version-3 migration

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
| 19. Pre-Built Templates | v1.5 | 0/1 | Not started | - |
