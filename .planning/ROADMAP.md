# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.8: Shipped (see MILESTONES.md)
- v1.9 Fuzzy Date Todos: Shipped 2026-02-12 (see MILESTONES.md)
- v2.0 Settings UX: In progress

## Phases

<details>
<summary>v1.9 Fuzzy Date Todos (Phases 27-29) — SHIPPED 2026-02-12</summary>

- [x] Phase 27: Date Precision & Input (2/2 plans) — completed 2026-02-12
- [x] Phase 28: Display & Indicators (2/2 plans) — completed 2026-02-12
- [x] Phase 29: Settings & View Filtering (1/1 plan) — completed 2026-02-12

</details>

### v2.0 Settings UX (In Progress)

**Milestone Goal:** Make settings overlay save-on-close -- Esc saves and dismisses, no explicit save button needed.

#### Phase 30: Save-on-Close Settings
**Goal**: Settings overlay saves automatically when user presses Esc
**Depends on**: Nothing (standalone UX change)
**Requirements**: SET-01, SET-02
**Success Criteria** (what must be TRUE):
  1. User presses Esc in settings overlay and all changed settings persist to config.toml
  2. No save button is visible anywhere in the settings overlay
  3. No cancel flow or confirmation dialog exists -- Esc simply saves and closes
  4. Live preview continues to work while adjusting settings (no regression)
**Plans**: TBD

Plans:
- [ ] 30-01: TBD

## Progress

**Execution Order:** Phase 30

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 27. Date Precision & Input | v1.9 | 2/2 | Complete | 2026-02-12 |
| 28. Display & Indicators | v1.9 | 2/2 | Complete | 2026-02-12 |
| 29. Settings & View Filtering | v1.9 | 1/1 | Complete | 2026-02-12 |
| 30. Save-on-Close Settings | v2.0 | 0/? | Not started | - |
