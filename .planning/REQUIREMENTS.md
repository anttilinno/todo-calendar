# Requirements: Todo Calendar

**Defined:** 2026-02-13
**Core Value:** See your month at a glance — calendar with holidays and todos in one terminal screen

## v2.2 Requirements

Requirements for Google Calendar Events milestone. Each maps to roadmap phases.

### Authentication

- [x] **AUTH-01**: User can authenticate with Google via OAuth 2.0 loopback redirect flow (browser opens, user signs in, token returned)
- [x] **AUTH-02**: OAuth refresh token persisted to disk (separate file, 0600 permissions) for re-use across app launches
- [x] **AUTH-03**: Token auto-refreshes transparently — user does not re-authenticate unless token is revoked
- [x] **AUTH-04**: App works fully offline when Google account is not configured (graceful no-op)

### Event Fetching

- [ ] **FETCH-01**: Events fetched from user's primary Google Calendar via REST API on app startup
- [ ] **FETCH-02**: Background polling re-fetches events every 5 minutes using tea.Tick
- [ ] **FETCH-03**: SyncToken-based delta sync for efficient incremental updates
- [ ] **FETCH-04**: Events cached in-memory (not persisted to SQLite)
- [ ] **FETCH-05**: Network errors handled gracefully — show last known data, no TUI freeze
- [ ] **FETCH-06**: Recurring events expanded server-side via singleEvents=true parameter
- [ ] **FETCH-07**: All-day events parsed as date strings without timezone conversion (prevents off-by-one)

### Event Display

- [ ] **DISP-01**: Timed events show HH:MM time prefix (e.g., "09:00 Team Standup")
- [ ] **DISP-02**: All-day events show "all day" label instead of time
- [ ] **DISP-03**: Events visually distinct from todos — no checkbox, distinct foreground color, time prefix
- [ ] **DISP-04**: Events sorted above todos within the dated section (events first, then todos)
- [ ] **DISP-05**: Events are not selectable — cursor skips over event items in the todo list
- [ ] **DISP-06**: Multi-day events expanded to show on each day they span
- [ ] **DISP-07**: Events respect monthly view filtering (only current month's events shown)
- [ ] **DISP-08**: Events respect weekly view filtering (only current week's events shown)

### Calendar Grid

- [ ] **GRID-01**: Calendar grid day indicators include event presence (days with events show indicator even without todos)

### Configuration

- [ ] **CONF-01**: Google Calendar toggle in settings overlay (enable/disable without removing credentials)
- [x] **CONF-02**: OAuth setup flow triggered from settings or first-run detection

## Future Requirements

Deferred to subsequent milestones. Tracked but not in current roadmap.

### Event Display Enhancements

- **DISP-09**: Event color from Google Calendar reflected in TUI
- **DISP-10**: Event location displayed as dimmed suffix
- **DISP-11**: Event preview on highlight (details in preview pane)
- **DISP-12**: Open event in browser action (xdg-open to Google Calendar)

### Integration Enhancements

- **INTG-01**: Events in full-screen search results (Ctrl+F)
- **INTG-02**: Events matched by inline filter (/)
- **INTG-03**: Event count in calendar overview panel
- **INTG-04**: Stale data indicator when last fetch is >10 minutes old

### Multi-Provider

- **PROV-01**: CalDAV support for Nextcloud/Radicale/Fastmail (app password auth)
- **PROV-02**: Multiple calendar support with calendar selection in settings

### Todo Enhancements

- **TODO-01**: Optional time field on todos for chronological sorting alongside events

## Out of Scope

| Feature | Reason |
|---------|--------|
| Event creation from TUI | Read-only display only — write scopes add complexity |
| Event editing from TUI | Read-only — edit in Google Calendar web/mobile |
| RSVP / attendee responses | Out of scope for read-only viewer |
| Event reminders / notifications | No daemon process, no notification infrastructure |
| CalDAV protocol (this milestone) | Google REST API only — CalDAV deferred to future milestone |
| Event persistence to SQLite | Events are ephemeral, rebuilt on each sync |
| Free/busy time blocks on calendar grid | 4-char cells too small for time blocks |
| App password authentication | Google disabled app passwords for Calendar in Sept 2024 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 33 | Done |
| AUTH-02 | Phase 33 | Done |
| AUTH-03 | Phase 33 | Done |
| AUTH-04 | Phase 33 | Done |
| CONF-02 | Phase 33 | Done |
| FETCH-01 | Phase 34 | Pending |
| FETCH-02 | Phase 34 | Pending |
| FETCH-03 | Phase 34 | Pending |
| FETCH-04 | Phase 34 | Pending |
| FETCH-05 | Phase 34 | Pending |
| FETCH-06 | Phase 34 | Pending |
| FETCH-07 | Phase 34 | Pending |
| DISP-01 | Phase 35 | Pending |
| DISP-02 | Phase 35 | Pending |
| DISP-03 | Phase 35 | Pending |
| DISP-04 | Phase 35 | Pending |
| DISP-05 | Phase 35 | Pending |
| DISP-06 | Phase 35 | Pending |
| DISP-07 | Phase 35 | Pending |
| DISP-08 | Phase 35 | Pending |
| GRID-01 | Phase 35 | Pending |
| CONF-01 | Phase 35 | Pending |

**Coverage:**
- v2.2 requirements: 22 total
- Mapped to phases: 22
- Unmapped: 0

---
*Requirements defined: 2026-02-13*
*Last updated: 2026-02-13 after roadmap creation*
