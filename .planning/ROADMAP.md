# Roadmap: Todo Calendar

## Milestones

- v1.0 MVP - Phases 1-3 (shipped 2026-02-05)
- v1.1 Polish & Personalization - Phases 4-6 (in progress)

## Phases

<details>
<summary>v1.0 MVP (Phases 1-3) - SHIPPED 2026-02-05</summary>

### Phase 1: TUI Scaffold
**Goal**: Split-pane terminal layout with calendar and todo panels
**Plans**: 1 plan

Plans:
- [x] 01-01: TUI scaffold with split-pane layout, calendar grid, and navigation

### Phase 2: Calendar Features
**Goal**: Monthly calendar with holidays and navigation
**Plans**: 2 plans

Plans:
- [x] 02-01: Holiday integration with configurable country
- [x] 02-02: Calendar-todo month synchronization

### Phase 3: Todo Management
**Goal**: Full todo CRUD with persistence
**Plans**: 2 plans

Plans:
- [x] 03-01: Todo CRUD with three-mode input system
- [x] 03-02: Persistence, help bar, and polish

</details>

### v1.1 Polish & Personalization (In Progress)

**Milestone Goal:** Make the calendar more informative at a glance and let users customize the experience.

**Phase Numbering:**
- Integer phases (4, 5, 6): Planned milestone work
- Decimal phases (4.1, 5.1): Urgent insertions (marked with INSERTED)

- [x] **Phase 4: Calendar Enhancements** - Date indicators and configurable first day of week
- [x] **Phase 5: Todo Editing** - Edit todo text and dates after creation
- [ ] **Phase 6: Themes** - Preset color themes selectable in config

## Phase Details

### Phase 4: Calendar Enhancements
**Goal**: Users see at a glance which dates have pending work, and can configure their preferred week layout
**Depends on**: Phase 3 (existing calendar grid and todo store)
**Requirements**: INDI-01, INDI-02, INDI-03, FDOW-01, FDOW-02, FDOW-03
**Success Criteria** (what must be TRUE):
  1. Calendar dates with incomplete todos display bracket indicators `[N]` around the date number
  2. Dates with only completed todos render identically to dates with no todos (no indicator)
  3. Calendar grid columns and rows remain properly aligned when indicators are present alongside non-indicated dates
  4. User can set `first_day_of_week = "monday"` or `"sunday"` in config.toml and the calendar grid starts on that day
  5. Day-of-week header row (Mo Tu We...) reflects the configured start day
**Plans**: 2 plans

Plans:
- [x] 04-01-PLAN.md -- Config migration, store query method, and grid rendering overhaul
- [x] 04-02-PLAN.md -- Wire indicators through calendar model and update app layout

### Phase 5: Todo Editing
**Goal**: Users can modify todos after creation without deleting and re-adding
**Depends on**: Phase 3 (existing todo CRUD and input system)
**Requirements**: EDIT-01, EDIT-02, EDIT-03
**Success Criteria** (what must be TRUE):
  1. User can press `e` on a selected todo to enter edit mode with existing text pre-filled, modify the text, and confirm
  2. User can change a todo's date -- adding a date to a floating todo, changing an existing date, or removing a date to make it floating
  3. Edited todos are persisted to disk immediately after confirmation (surviving app restart)
**Plans**: 2 plans

Plans:
- [x] 05-01-PLAN.md -- Store Update/Find methods and edit key bindings
- [x] 05-02-PLAN.md -- Edit text and edit date mode handlers

### Phase 6: Themes
**Goal**: Users can personalize the app's appearance by choosing a color theme
**Depends on**: Phases 4-5 (all UI elements exist to be themed)
**Requirements**: THME-01, THME-02, THME-03, THME-04
**Success Criteria** (what must be TRUE):
  1. App ships with 4 distinct preset themes: Dark, Light, Nord, and Solarized
  2. User can set `theme = "dark"` (or light, nord, solarized) in config.toml and the app renders with that theme on next launch
  3. All UI elements -- borders, panel backgrounds, calendar highlights, holiday text, todo text, help bar, date indicators -- render in colors consistent with the selected theme
  4. When no theme is configured in config.toml, the app defaults to the Dark theme
**Plans**: TBD

Plans:
- [ ] 06-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 4 -> 5 -> 6

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. TUI Scaffold | v1.0 | 1/1 | Complete | 2026-02-05 |
| 2. Calendar Features | v1.0 | 2/2 | Complete | 2026-02-05 |
| 3. Todo Management | v1.0 | 2/2 | Complete | 2026-02-05 |
| 4. Calendar Enhancements | v1.1 | 2/2 | Complete | 2026-02-05 |
| 5. Todo Editing | v1.1 | 2/2 | Complete | 2026-02-05 |
| 6. Themes | v1.1 | 0/TBD | Not started | - |
