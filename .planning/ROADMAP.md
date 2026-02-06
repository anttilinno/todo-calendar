# Roadmap: Todo Calendar

## Milestones

- ✅ **v1.0 MVP** - Phases 1-3 (shipped 2026-02-05)
- ✅ **v1.1 Polish & Personalization** - Phases 4-6 (shipped 2026-02-05)
- 🚧 **v1.2 Reorder & Settings** - Phases 7-9 (in progress)

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

### 🚧 v1.2 Reorder & Settings (In Progress)

**Milestone Goal:** Let users reorder todos, configure the app from an in-app settings page, and see todo count overview on the calendar.

#### Phase 7: Todo Reordering
**Goal**: Users can arrange todos in their preferred order
**Depends on**: Phase 6 (existing store and todolist component)
**Requirements**: REORD-01, REORD-02, REORD-03
**Success Criteria** (what must be TRUE):
  1. User can move the selected todo one position up in the list via keybinding
  2. User can move the selected todo one position down in the list via keybinding
  3. Custom todo order survives app restart (order is persisted in JSON)
  4. Reorder keybindings appear in the help bar when a todo is selected
**Plans**: 2 plans

Plans:
- [x] 07-01: Add SortOrder field, migration, SwapOrder method, updated sort logic
- [x] 07-02: Wire MoveUp/MoveDown keybindings and help bar integration

#### Phase 8: Settings Overlay
**Goal**: Users can configure theme, holiday country, and first day of week from inside the app with live preview
**Depends on**: Phase 7 (builds on store changes, existing theme/config systems)
**Requirements**: SETT-01, SETT-02, SETT-03, SETT-04, SETT-05, SETT-06
**Success Criteria** (what must be TRUE):
  1. User can open a full-screen settings overlay via a keybinding from any panel
  2. User can change the color theme and see the app redraw immediately (live preview)
  3. User can change the holiday country and first day of week within the settings overlay
  4. User can save all settings changes to config.toml and return to the main view
  5. User can dismiss settings without saving, reverting any previewed changes
**Plans**: 2 plans

Plans:
- [x] 08-01: Config.Save, theme.Names, settings model, SetTheme methods
- [x] 08-02: Wire settings overlay into app with live preview and save/cancel

#### Phase 9: Overview Panel
**Goal**: Calendar panel shows at-a-glance todo counts so users know where work is concentrated
**Depends on**: Phase 7 (store queries for counts; independent of settings)
**Requirements**: OVRVW-01, OVRVW-02
**Success Criteria** (what must be TRUE):
  1. Calendar panel displays todo count per month below the calendar grid (e.g., "January [7]")
  2. Overview shows count of undated (floating) todos (e.g., "Unknown [12]")
  3. Counts update live as todos are added, completed, or deleted
**Plans**: TBD

Plans:
- [ ] 09-01: TBD

## Progress

**Execution Order:** 7 → 8 → 9

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
| 9. Overview Panel | v1.2 | 0/TBD | Not started | - |
