# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.8: Shipped (see MILESTONES.md)
- v1.9 Fuzzy Date Todos: Shipped 2026-02-12 (see MILESTONES.md)
- v2.0 Settings UX: Shipped 2026-02-12 (see MILESTONES.md)
- v2.1 Priorities: Shipped 2026-02-13 (see MILESTONES.md)

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

### v2.2 Google Calendar Events (In Progress)

**Milestone Goal:** Display read-only Google Calendar events alongside todos in the TUI via Google REST API with OAuth 2.0 authentication

- [x] **Phase 33: OAuth & Offline Guard** - OAuth 2.0 authentication with token persistence and graceful offline fallback — completed 2026-02-14
- [ ] **Phase 34: Event Fetching & Async Integration** - Google Calendar API client with background polling and in-memory cache
- [ ] **Phase 35: Event Display & Grid** - Events rendered in todo list and calendar grid with settings toggle

## Phase Details

### Phase 33: OAuth & Offline Guard
**Goal**: Users can authenticate with Google and the app handles unconfigured/offline states gracefully
**Depends on**: Nothing (first phase of v2.2)
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, CONF-02
**Success Criteria** (what must be TRUE):
  1. User can complete OAuth flow (browser opens, user signs in, app receives token) on first setup
  2. App remembers authentication across restarts without re-prompting (token persisted securely)
  3. Token refreshes transparently — user never sees "re-authenticate" unless token is revoked
  4. App launches and works fully when Google account is not configured (no errors, no prompts)
  5. OAuth setup can be triggered from settings overlay or detected on first run
**Plans:** 2 plans
Plans:
- [x] 33-01-PLAN.md — OAuth core package (config, token persistence, auth flow, PKCE)
- [x] 33-02-PLAN.md — Settings integration and app wiring (Google Calendar row, auth trigger, offline guard)

### Phase 34: Event Fetching & Async Integration
**Goal**: Events are fetched from Google Calendar without freezing the TUI, with efficient incremental updates
**Depends on**: Phase 33
**Requirements**: FETCH-01, FETCH-02, FETCH-03, FETCH-04, FETCH-05, FETCH-06, FETCH-07
**Success Criteria** (what must be TRUE):
  1. Events from user's primary Google Calendar appear in the app after startup
  2. Events update automatically every 5 minutes without user action or TUI freeze
  3. Network errors show last known data gracefully — no crash, no blank screen, no hang
  4. Recurring events appear as individual occurrences (not collapsed into one entry)
  5. All-day events show on the correct calendar day regardless of timezone
**Plans:** 2 plans
Plans:
- [ ] 34-01-PLAN.md — CalendarEvent type, Google Calendar API client, syncToken delta sync, event conversion
- [ ] 34-02-PLAN.md — Bubble Tea app model wiring (startup fetch, 5-min polling, error resilience, auth guard)

### Phase 35: Event Display & Grid
**Goal**: Users see their Google Calendar events alongside todos with clear visual distinction in both the todo list and calendar grid
**Depends on**: Phase 34
**Requirements**: DISP-01, DISP-02, DISP-03, DISP-04, DISP-05, DISP-06, DISP-07, DISP-08, GRID-01, CONF-01
**Success Criteria** (what must be TRUE):
  1. Timed events show with HH:MM prefix and all-day events show "all day" label in the todo panel
  2. Events are visually distinct from todos (no checkbox, different color, not selectable)
  3. Calendar grid days with events show indicators even when no todos exist for that day
  4. Events respect the current view — monthly view shows month's events, weekly view shows week's events
  5. Google Calendar can be toggled on/off in settings without removing credentials
**Plans**: TBD

## Progress

**Execution Order:** 33 -> 34 -> 35

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 27. Date Precision & Input | v1.9 | 2/2 | Complete | 2026-02-12 |
| 28. Display & Indicators | v1.9 | 2/2 | Complete | 2026-02-12 |
| 29. Settings & View Filtering | v1.9 | 1/1 | Complete | 2026-02-12 |
| 30. Save-on-Close Settings | v2.0 | 1/1 | Complete | 2026-02-12 |
| 31. Priority Data Layer | v2.1 | 1/1 | Complete | 2026-02-13 |
| 32. Priority UI + Theme | v2.1 | 2/2 | Complete | 2026-02-13 |
| 33. OAuth & Offline Guard | v2.2 | 2/2 | Complete | 2026-02-14 |
| 34. Event Fetching & Async | v2.2 | 0/2 | Not started | - |
| 35. Event Display & Grid | v2.2 | 0/TBD | Not started | - |
