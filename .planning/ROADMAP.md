# Roadmap: Todo Calendar

## Milestones

- ✅ **v1.0 MVP** - Phases 1-3 (shipped 2026-02-05)
- ✅ **v1.1 Polish & Personalization** - Phases 4-6 (shipped 2026-02-05)
- ✅ **v1.2 Reorder & Settings** - Phases 7-9 (shipped 2026-02-06)
- 🚧 **v1.3 Views & Usability** - Phases 10-13 (in progress)

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

### 🚧 v1.3 Views & Usability (In Progress)

**Milestone Goal:** Enhanced calendar views, search, and regional preferences for faster daily use.

#### Phase 10: Overview Color Coding
**Goal**: Users see completion progress at a glance in the overview panel
**Depends on**: Phase 9 (overview panel exists)
**Requirements**: OVCLR-01, OVCLR-02
**Success Criteria** (what must be TRUE):
  1. Overview panel displays separate pending and completed counts per month
  2. Pending and completed counts are visually distinct via color
  3. Overview colors change when user switches themes
**Plans**: 1 plan

Plans:
- [ ] 10-01: Split overview counts with theme-aware color roles

#### Phase 11: Date Format Setting
**Goal**: Users see dates in their preferred regional format
**Depends on**: Nothing (independent feature)
**Requirements**: DTFMT-01, DTFMT-02, DTFMT-03
**Success Criteria** (what must be TRUE):
  1. User can cycle through 3 date format presets in settings overlay
  2. All date displays throughout the app reflect the chosen format
  3. Date format preference survives app restart
**Plans**: 1 plan

Plans:
- [ ] 11-01: Date format config, settings integration, and display propagation

#### Phase 12: Weekly Calendar View
**Goal**: Users can zoom into a single week for focused daily planning
**Depends on**: Nothing (independent feature, soft dependency on Phase 11 for date headers)
**Requirements**: WKVIEW-01, WKVIEW-02, WKVIEW-03, WKVIEW-04
**Success Criteria** (what must be TRUE):
  1. User can toggle between monthly and weekly view with a keybinding
  2. Weekly view displays 7 days with day numbers, holiday markers, and todo indicators
  3. User can navigate week-by-week in weekly mode
  4. Switching from monthly to weekly view auto-selects the current week
**Plans**: 1 plan

Plans:
- [ ] 12-01: Weekly view mode with toggle, grid rendering, and week navigation

#### Phase 13: Search & Filter
**Goal**: Users can find any todo regardless of which month it lives in
**Depends on**: Nothing (independent feature, soft dependency on Phase 11 for date display in results)
**Requirements**: SRCH-01, SRCH-02, SRCH-03, SRCH-04, SRCH-05
**Success Criteria** (what must be TRUE):
  1. User can type `/` to filter the current month's todo list by text
  2. User can press Esc to clear the filter and return to normal mode
  3. User can open a full-screen search overlay to search across all months
  4. Search results show matching todos with their associated dates
  5. User can select a search result and jump to that todo's month
**Plans**: 2 plans

Plans:
- [ ] 13-01: Inline todo filter with `/` activation and Esc clear
- [ ] 13-02: Full-screen search overlay with cross-month results and navigation

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
| 10. Overview Color Coding | v1.3 | 0/1 | Not started | - |
| 11. Date Format Setting | v1.3 | 0/1 | Not started | - |
| 12. Weekly Calendar View | v1.3 | 0/1 | Not started | - |
| 13. Search & Filter | v1.3 | 0/2 | Not started | - |
