# Roadmap: Todo Calendar

## Milestones

- ✅ **v1.0 MVP** - Phases 1-3 (shipped 2026-02-05)
- ✅ **v1.1 Polish & Personalization** - Phases 4-6 (shipped 2026-02-05)
- ✅ **v1.2 Reorder & Settings** - Phases 7-9 (shipped 2026-02-06)
- ✅ **v1.3 Views & Usability** - Phases 10-13 (shipped 2026-02-06)
- 📋 **v1.4 Data & Editing** - Phases 14-16 (planned)

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

### 📋 v1.4 Data & Editing (Planned)

**Milestone Goal:** Persistent database storage, rich markdown todo bodies, and external editor integration for power-user workflows.

#### Phase 14: Database Backend
**Goal**: Migrate todo persistence from JSON to SQLite for reliability and query capability
**Depends on**: Nothing (replaces existing JSON store)
**Requirements**: DB-01, DB-02, DB-03
**Success Criteria** (what must be TRUE):
  1. Todos are stored in a SQLite database instead of JSON files
  2. Existing JSON data is migrated automatically on first launch
  3. All existing CRUD operations work identically with the new backend
**Plans**: TBD

#### Phase 15: Markdown Templates
**Goal**: Todo bodies use markdown templates with placeholders so users can define reusable todo structures
**Depends on**: Phase 14 (database supports richer todo body storage)
**Requirements**: MDTPL-01, MDTPL-02, MDTPL-03
**Success Criteria** (what must be TRUE):
  1. Todos have a markdown body field beyond the single-line title
  2. User can create markdown templates with placeholder variables
  3. Creating a todo from a template fills in placeholders interactively
**Plans**: TBD

#### Phase 16: External Editor Integration
**Goal**: Users can open and edit todo content in their preferred $EDITOR (neovim, vim, etc.)
**Depends on**: Phase 15 (markdown body exists to edit)
**Requirements**: EDITOR-01, EDITOR-02, EDITOR-03
**Success Criteria** (what must be TRUE):
  1. User can press a key to open the selected todo in $EDITOR
  2. App suspends TUI, launches editor, and resumes cleanly after editor exits
  3. Edits made in the external editor are saved back to the todo
**Plans**: TBD

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
| 14. Database Backend | v1.4 | 0/? | Not started | - |
| 15. Markdown Templates | v1.4 | 0/? | Not started | - |
| 16. External Editor | v1.4 | 0/? | Not started | - |
