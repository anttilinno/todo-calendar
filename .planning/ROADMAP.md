# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.8: Shipped (see MILESTONES.md)
- v1.9 Fuzzy Date Todos: Shipped 2026-02-12 (see MILESTONES.md)
- v2.0 Settings UX: Shipped 2026-02-12 (see MILESTONES.md)

## Phases

<details>
<summary>v1.9 Fuzzy Date Todos (Phases 27-29) — SHIPPED 2026-02-12</summary>

- [x] Phase 27: Date Precision & Input (2/2 plans) — completed 2026-02-12
- [x] Phase 28: Display & Indicators (2/2 plans) — completed 2026-02-12
- [x] Phase 29: Settings & View Filtering (1/1 plan) — completed 2026-02-12

</details>

<details>
<summary>v2.0 Settings UX (Phase 30) — SHIPPED 2026-02-12</summary>

- [x] Phase 30: Save-on-Close Settings — completed 2026-02-12

</details>

### v2.1 Priorities & Smart Dates (In Progress)

**Milestone Goal:** Add P1-P4 priority levels with color-coded badges and calendar indicators, plus natural language date input replacing the segmented 3-field date entry.

- [ ] **Phase 31: Priority Data Layer** - Schema migration v7, Todo struct extension, store interface updates
- [ ] **Phase 32: Priority UI + Theme** - Edit form field, badge rendering, theme colors, calendar indicators, search
- [ ] **Phase 33: Natural Language Date Input** - NL parser package, single text field replacing segmented input, precision detection

## Phase Details

### Phase 31: Priority Data Layer
**Goal**: Todos have a priority field that persists through the full store roundtrip
**Depends on**: Phase 30 (v2.0 complete)
**Requirements**: PRIO-08, PRIO-09
**Success Criteria** (what must be TRUE):
  1. SQLite schema is at version 7 with a priority INTEGER column on the todos table
  2. Existing todos have priority 0 (no priority) after migration -- no data loss
  3. Store Add() and Update() accept priority and persist it correctly
  4. Store queries return todos with their priority value populated in the Todo struct
**Plans**: TBD

Plans:
- [ ] 31-01: TBD

### Phase 32: Priority UI + Theme
**Goal**: Users can set, see, and distinguish priority levels across the entire interface
**Depends on**: Phase 31
**Requirements**: PRIO-01, PRIO-02, PRIO-03, PRIO-04, PRIO-05, PRIO-06, PRIO-07
**Success Criteria** (what must be TRUE):
  1. User can set priority (P1-P4 or none) on any todo via the edit/add form
  2. Todos display a colored [P1]-[P4] badge prefix with aligned text across all priority levels including no-priority
  3. Completed prioritized todos show the colored badge but greyed-out strikethrough text
  4. Calendar day indicators reflect the highest-priority incomplete todo's color for that day
  5. Search results display priority badges matching the todo list rendering
**Plans**: TBD

Plans:
- [ ] 32-01: TBD
- [ ] 32-02: TBD

### Phase 33: Natural Language Date Input
**Goal**: Users type dates naturally in a single text field with automatic precision detection
**Depends on**: Phase 32
**Requirements**: NLDT-01, NLDT-02, NLDT-03, NLDT-04, NLDT-05, NLDT-06, NLDT-07, NLDT-08, NLDT-09, NLDT-10
**Success Criteria** (what must be TRUE):
  1. User can type "tomorrow", "next friday", "jan 15", "in 3 days" and get the correct resolved date
  2. Date field shows a parsed date preview so the user can confirm the interpretation before saving
  3. Year-only input produces a This Year todo and month-year input produces a This Month todo
  4. Formatted dates (ISO/EU/US per user setting) still work in the same field
  5. Empty date input creates a floating (undated) todo, same as before
**Plans**: TBD

Plans:
- [ ] 33-01: TBD
- [ ] 33-02: TBD

## Progress

**Execution Order:** 31 -> 32 -> 33

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 27. Date Precision & Input | v1.9 | 2/2 | Complete | 2026-02-12 |
| 28. Display & Indicators | v1.9 | 2/2 | Complete | 2026-02-12 |
| 29. Settings & View Filtering | v1.9 | 1/1 | Complete | 2026-02-12 |
| 30. Save-on-Close Settings | v2.0 | 1/1 | Complete | 2026-02-12 |
| 31. Priority Data Layer | v2.1 | 0/? | Not started | - |
| 32. Priority UI + Theme | v2.1 | 0/? | Not started | - |
| 33. Natural Language Date Input | v2.1 | 0/? | Not started | - |
