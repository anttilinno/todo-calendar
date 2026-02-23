# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.8: Shipped (see MILESTONES.md)
- v1.9 Fuzzy Date Todos: Shipped 2026-02-12 (see MILESTONES.md)
- v2.0 Settings UX: Shipped 2026-02-12 (see MILESTONES.md)
- v2.1 Priorities: Shipped 2026-02-13 (see MILESTONES.md)
- v2.2 Google Calendar Events: Shipped 2026-02-14 (see MILESTONES.md)
- v2.3 Polybar Status: In progress

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

<details>
<summary>v2.1 Priorities (Phases 31-32) — SHIPPED 2026-02-13</summary>

- [x] Phase 31: Priority Data Layer (1/1 plan) — completed 2026-02-13
- [x] Phase 32: Priority UI + Theme (2/2 plans) — completed 2026-02-13

</details>

<details>
<summary>v2.2 Google Calendar Events (Phases 33-35) — SHIPPED 2026-02-14</summary>

- [x] Phase 33: OAuth & Offline Guard (2/2 plans) — completed 2026-02-14
- [x] Phase 34: Event Fetching & Async Integration (2/2 plans) — completed 2026-02-14
- [x] Phase 35: Event Display & Grid (3/3 plans) — completed 2026-02-14

</details>

### v2.3 Polybar Status (In Progress)

- [x] **Phase 36: Status Subcommand** - CLI subcommand that queries SQLite and writes Polybar-formatted status to state file (completed 2026-02-23)
- [ ] **Phase 37: TUI State File Integration** - TUI writes state file on startup and on every todo mutation

## Phase Details

### Phase 36: Status Subcommand
**Goal**: User can run `todo-calendar status` from a shell or Polybar script to get a current pending-todo indicator written to a state file
**Depends on**: Nothing (first phase in v2.3)
**Requirements**: BAR-01, BAR-02, BAR-03
**Success Criteria** (what must be TRUE):
  1. Running `todo-calendar status` queries SQLite for today's pending todos and writes output to `/tmp/.todo_status`
  2. State file contains `%{F#hex}ICON COUNT%{F-}` where hex color matches highest-priority pending todo's theme color
  3. State file contains empty string when there are zero pending todos today
  4. Subcommand exits immediately after writing (no TUI, no blocking)
**Plans**: 2 plans

Plans:
- [ ] 36-01-PLAN.md — Status formatting engine (TDD: PriorityColorHex + FormatStatus + WriteStatusFile)
- [ ] 36-02-PLAN.md — Wire status subcommand in main.go

### Phase 37: TUI State File Integration
**Goal**: Polybar status stays current while the TUI is running, without requiring periodic `status` subcommand invocations
**Depends on**: Phase 36
**Requirements**: BAR-04, BAR-05
**Success Criteria** (what must be TRUE):
  1. TUI writes state file on startup so Polybar reflects current state when the app opens
  2. Adding, completing, deleting, or editing a todo in the TUI immediately updates the state file
  3. State file output format is identical to what the `status` subcommand produces
**Plans**: TBD

Plans:
- [ ] 37-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 36 -> 37

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 27. Date Precision & Input | v1.9 | 2/2 | Complete | 2026-02-12 |
| 28. Display & Indicators | v1.9 | 2/2 | Complete | 2026-02-12 |
| 29. Settings & View Filtering | v1.9 | 1/1 | Complete | 2026-02-12 |
| 30. Save-on-Close Settings | v2.0 | 1/1 | Complete | 2026-02-12 |
| 31. Priority Data Layer | v2.1 | 1/1 | Complete | 2026-02-13 |
| 32. Priority UI + Theme | v2.1 | 2/2 | Complete | 2026-02-13 |
| 33. OAuth & Offline Guard | v2.2 | 2/2 | Complete | 2026-02-14 |
| 34. Event Fetching & Async | v2.2 | 2/2 | Complete | 2026-02-14 |
| 35. Event Display & Grid | v2.2 | 3/3 | Complete | 2026-02-14 |
| 36. Status Subcommand | 2/2 | Complete   | 2026-02-23 | - |
| 37. TUI State File Integration | v2.3 | 0/? | Not started | - |
