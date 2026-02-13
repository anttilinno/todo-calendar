# Project Research Summary

**Project:** Google Calendar Integration for Todo-Calendar TUI
**Domain:** Read-only calendar event display via REST API/CalDAV in Go/Bubble Tea TUI
**Researched:** 2026-02-13
**Confidence:** HIGH

## Executive Summary

This research investigated adding read-only Google Calendar event display to an existing Go/Bubble Tea TUI todo-calendar application. The initial assumption of using CalDAV with app passwords is **fundamentally broken** — Google disabled basic authentication for CalDAV on September 30, 2024. OAuth 2.0 is now mandatory.

Two viable technical paths emerged: (1) Google Calendar REST API with native syncToken support and official Go client, or (2) CalDAV via emersion/go-webdav with OAuth2, which works with Google but also supports self-hosted servers (Nextcloud, Radicale, Fastmail). The recommended approach is a **dual-path architecture**: use REST API for Google accounts (better API, simpler OAuth, native sync) and CalDAV for future non-Google providers. This provides the best user experience for Google while maintaining extensibility.

The primary risks are architectural: this is the app's first network feature, requiring fundamental changes to the synchronous event loop (async commands via tea.Cmd), token management (OAuth refresh, secure storage), and data model (ephemeral events vs persistent todos). Critical pitfalls include blocking network calls that freeze the TUI, timezone handling for all-day events (off-by-one day bugs), and OAuth token security. These are all preventable with established Bubble Tea patterns and timezone-aware parsing.

## Key Findings

### Recommended Stack

**Authentication changed everything.** Google's deprecation of basic auth forces OAuth 2.0, which adds complexity but is well-supported in Go. The REST API path provides a better developer experience than CalDAV for Google-specific integration, while CalDAV remains valuable for future multi-provider support.

**Core technologies:**
- **google.golang.org/api/calendar/v3** (v0.266.0): Official Google Calendar client — native syncToken support, typed event model, simpler than parsing iCalendar
- **golang.org/x/oauth2**: Token management with automatic refresh — required for Google auth, handles access token expiry transparently
- **emersion/go-webdav** (v0.7.0): CalDAV client for non-Google providers — enables future Nextcloud/Fastmail support, protocol-level access
- **emersion/go-ical**: iCalendar parsing companion to go-webdav — required for CalDAV response parsing
- **Bubble Tea tea.Cmd pattern**: Async network operations — prevents TUI freezes, already used for editor launches

**Critical stack decision:** For Google-only v1, the REST API alone is sufficient. For multi-provider future, build the abstraction layer now with both REST and CalDAV clients behind a common interface.

### Expected Features

Research identified 11 table-stakes features and 9 differentiators based on patterns from Todoist, Things 3, gcal-tui, and calcure.

**Must have (table stakes):**
- Timed events with HH:MM time prefix (core visual distinction)
- All-day events without time prefix (separate handling from timed)
- Visual distinction from todos (no checkbox, distinct color, event-specific styling)
- Events appear in dated section above todos (fixed commitments first)
- Multi-day event expansion (show on each day of span)
- Recurring event expansion via API (singleEvents=true parameter)
- 5-minute background polling refresh (non-blocking fetch)
- Events not selectable/actionable (cursor skips over)
- Calendar grid day indicators include events (visual presence signal)
- Timezone-aware display (events in user's local time)
- Week/month view filtering (events respect view range)

**Should have (competitive):**
- Event color from Google Calendar (visual consistency)
- Event location display (context for "where")
- Events in search results (unified search across todos and events)
- Multiple calendar support (work/personal separation)
- Open in browser action (escape hatch for full details)
- Stale data indicator (show when last sync was)

**Defer (v2+):**
- Event creation/editing (requires write scope, massive complexity)
- RSVP/attendee responses (out of scope for read-only)
- Event reminders/notifications (requires background daemon)
- CalDAV protocol support (defer to separate milestone)

### Architecture Approach

The existing app is 100% offline with synchronous SQLite operations. CalDAV integration introduces the first async network feature. The architecture must preserve the offline-first model for todos while adding ephemeral network-sourced events.

**Major components:**
1. **internal/gcal/** (new package) — Google Calendar client wrapper, OAuth token management, CalendarEvent domain type, in-memory cache, Bubble Tea message types for sync results
2. **app.Model extensions** — Starts sync commands via tea.Cmd, receives sync result messages, distributes events to child models (calendar, todolist)
3. **calendar.Model modifications** — Adds eventCounts map, SetCalendarEvents() method, renders event indicators on grid days
4. **todolist.Model modifications** — Adds new Events section with eventItem kind (non-selectable), renders events with time prefix and distinct styling

**Key patterns to follow:**
- Constructor DI (pass dependencies, not globals) — matches existing store/provider/config pattern
- tea.Cmd for all network I/O — already used for editor launches
- Pure rendering functions (RenderGrid takes data as parameters) — existing pattern
- SetX methods for cross-component data flow — existing pattern in RefreshIndicators, SetViewMonth

**Anti-patterns to avoid:**
- Storing events in SQLite (events are ephemeral, rebuilt on each sync)
- Blocking network calls in Update() (freezes TUI)
- Making events selectable/editable (scope creep to full calendar client)
- OAuth flow inside running TUI (handle in main.go before Bubble Tea starts)

### Critical Pitfalls

Research identified 15 pitfalls across authentication, async patterns, parsing, and UX. The top 5 critical issues:

1. **App passwords do NOT work** — Google disabled basic auth for CalDAV. OAuth 2.0 is mandatory. Assumption of app-password authentication fails immediately with HTTP 401. Prevention: OAuth from day one via golang.org/x/oauth2 with device flow or localhost redirect.

2. **Blocking network calls freeze TUI** — Synchronous CalDAV fetch in Update() blocks entire event loop for 1-5 seconds. Prevention: ALL network operations via tea.Cmd (goroutine-based), with context.WithTimeout(30s) on every call.

3. **All-day event timezone off-by-one** — Parsing VALUE=DATE with timezone conversion shifts events to wrong calendar day. Prevention: Parse all-day dates as pure strings ("20260215" → "2026-02-15") without time.Time conversion.

4. **OAuth token storage in plaintext config** — Refresh tokens grant indefinite access. Prevention: Separate file with 0600 permissions at ~/.config/todo-calendar/oauth-token.json, NOT in config.toml.

5. **Missing HTTP timeout causes goroutine leaks** — Default http.Client has no timeout. Failed connections block for 2+ minutes. Prevention: Explicit 30s timeout on http.Client.Timeout and http.Transport.DialContext.

**Additional moderate pitfalls:**
- Stale data displayed as current after network outage (show sync status indicator)
- Token refresh failure silently breaks all fetches (custom TokenSource with error detection)
- Google CalDAV URL construction differs from standard discovery (hardcode endpoint URLs)
- Timed events from different timezones display at wrong local time (convert to time.Local before date extraction)
- Recurring events require RRULE expansion (use EXPAND in CalendarQuery, verify Google support)

## Implications for Roadmap

Based on combined research, the integration naturally decomposes into 5 phases with clear dependencies.

### Phase 1: OAuth 2.0 Authentication Foundation
**Rationale:** Authentication must work before any API interaction. OAuth is the most complex new dependency and blocks all subsequent phases. Handling this first validates the core assumption and enables iterative development.

**Delivers:** OAuth token acquisition, refresh, and persistence. Google API Console setup documentation.

**Stack:** golang.org/x/oauth2, device authorization flow (TUI-friendly), token file storage (0600 permissions)

**Addresses:** Pitfall 1 (app passwords don't work), Pitfall 3 (token storage security), Pitfall 7 (token refresh failure)

**Avoids:** Building CalDAV integration on broken authentication

**Research flag:** LOW — OAuth patterns are well-documented, Go library is mature, Google docs are official

### Phase 2: Google Calendar API Client & Event Model
**Rationale:** With auth working, establish the data fetching layer. REST API is simpler than CalDAV for Google and provides better sync primitives. Creating the domain model (CalendarEvent) early enables parallel UI development.

**Delivers:** Event fetching via google.golang.org/api/calendar/v3, CalendarEvent domain type, in-memory cache, syncToken-based delta sync

**Uses:** google.golang.org/api/calendar/v3 (official client), Events.List with timeMin/timeMax/singleEvents, syncToken for incremental updates

**Implements:** internal/gcal package (client wrapper, event model, cache)

**Addresses:** Pitfall 5 (timezone handling), Pitfall 10 (recurring expansion via singleEvents=true)

**Research flag:** LOW — Official Google API with documented sync patterns

### Phase 3: Bubble Tea Async Integration
**Rationale:** Network operations require async commands to prevent TUI freezes. This phase establishes the pattern for all future network features, including background polling.

**Delivers:** tea.Cmd wrappers for API calls, CalDAVSyncDoneMsg handling, 5-minute tea.Tick polling, graceful error handling

**Uses:** tea.Cmd (goroutine-based async), tea.Tick (periodic polling), context.WithTimeout (request timeouts)

**Implements:** Sync commands in app.Model, message routing, poll scheduling

**Addresses:** Pitfall 2 (blocking calls freeze TUI), Pitfall 4 (HTTP timeouts)

**Avoids:** TUI responsiveness degradation, goroutine leaks

**Research flag:** LOW — Bubble Tea async patterns are well-documented, matches existing editor.Open pattern

### Phase 4: Calendar Grid Event Indicators
**Rationale:** Visual feedback on the calendar grid shows which days have events. This is table-stakes UX (all competitors do this) and provides value before the full event list rendering.

**Delivers:** Day-level event indicators on calendar grid, event count tracking, visual distinction from todo indicators

**Uses:** Pure rendering pattern (eventCounts passed to RenderGrid), existing indicator system

**Implements:** calendar.Model.SetCalendarEvents(), eventCounts map, grid cell styling for events

**Addresses:** Feature: calendar grid indicators (table stakes)

**Research flag:** NONE — Extension of existing indicator pattern

### Phase 5: Todo List Event Section Rendering
**Rationale:** The primary event display surface. Shows event details (time, title, location) alongside dated todos with clear visual distinction.

**Delivers:** Events section in todo list, time-prefixed rendering, non-selectable event items, all-day vs timed styling

**Uses:** New eventItem kind, visibleItems() extension, event-specific styles

**Implements:** todolist.Model.SetCalendarEvents(), renderEvent(), event section in visibleItems()

**Addresses:** Features: timed events, all-day events, visual distinction, events above todos (all table stakes)

**Avoids:** Pitfall 12 (confusing sort order by keeping events in separate section)

**Research flag:** LOW — Extends existing section pattern (dated/month/year/floating)

### Phase Ordering Rationale

- **Phase 1 (OAuth) is foundational** — Nothing works without auth, validating this assumption early prevents wasted effort
- **Phase 2 (API client) before UI** — Domain model stabilizes before rendering logic depends on it
- **Phase 3 (async) before display** — Background polling must exist before event data is available to display
- **Phases 4 and 5 can be parallel** — Calendar and todolist are independent consumers of event data after Phase 3
- **Deferred: CalDAV abstraction** — Wait for user demand for non-Google providers before adding protocol complexity

**Dependency chain:** 1 → 2 → 3 → (4 | 5)

### Research Flags

Phases needing deeper research during planning:
- **None** — All critical patterns are well-documented (OAuth, Google API, Bubble Tea async, timezone handling)

Phases with standard patterns (skip research-phase):
- **Phase 1-5** — Official libraries (google.golang.org/api, golang.org/x/oauth2), established Bubble Tea patterns, timezone parsing in Go stdlib

**Implementation notes for planning:**
- Monitor Google Calendar API quotas (1M requests/day default, non-issue for single user)
- Test OAuth flow on headless/SSH environments (device flow is essential)
- Validate EXPAND parameter behavior with Google CalDAV if CalDAV path is chosen
- Test timezone edge cases (all-day events, DST transitions, events crossing midnight)

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Official Google libraries verified via pkg.go.dev, OAuth requirement confirmed via Google docs, Bubble Tea patterns documented |
| Features | HIGH | Table stakes derived from 4 competitor products (Todoist, Things 3, gcal-tui, calcure), Google API capabilities verified |
| Architecture | HIGH | Matches existing Bubble Tea patterns in codebase, async command pattern already used for editor launches |
| Pitfalls | HIGH | Critical issues (auth, timezone, async) verified via official docs and community reports, solutions tested in similar projects |

**Overall confidence:** HIGH

### Gaps to Address

Areas requiring validation during implementation (not blocking, but monitor):

- **Google API Console user setup friction** — Creating OAuth credentials is non-trivial for non-technical users. Document thoroughly or consider shipping a shared client_id for open-source use (Google allows <100 users unverified).

- **CTag/syncToken behavior with Google** — Incremental sync is documented but edge cases (token expiry, server-side changes) need real-world testing. Fallback to full fetch must be robust.

- **Multi-calendar selection UX** — Research assumed primary calendar only. If users request multiple calendars (Work, Personal), the Settings overlay needs calendar list/toggle UI. Defer to post-MVP.

- **CalDAV EXPAND parameter reliability** — If CalDAV path is chosen, verify that Google's CalDAV actually expands recurring events. Fallback plan if it doesn't: skip recurring events initially.

- **OAuth device flow vs localhost redirect** — Device flow is better for SSH/headless but requires more user interaction (visit URL, enter code). Localhost redirect is instant but fails in SSH. May need to support both flows.

## Sources

### Primary (HIGH confidence)
- Official Google CalDAV API Developer's Guide — endpoint URLs, OAuth requirement, feature support
- google.golang.org/api/calendar/v3 pkg.go.dev — Events.List API, syncToken, timeMin/timeMax, singleEvents parameter
- golang.org/x/oauth2 pkg.go.dev — TokenSource, Config, auto-refresh mechanics
- emersion/go-webdav pkg.go.dev — CalDAV client methods, CalendarQuery structure
- Bubble Tea official documentation — tea.Cmd pattern, tea.Tick polling, async examples
- Google Calendar API Sync Guide — syncToken flow, 410 gone handling
- RFC 5545 (iCalendar) — VALUE=DATE vs VALUE=DATE-TIME, RRULE syntax
- Codebase analysis (8,695 LOC across 35 Go files) — existing patterns, message routing, store interface

### Secondary (MEDIUM confidence)
- Todoist Calendar Integration docs — events-above-tasks pattern, color coding
- Things 3 feature documentation — events at top of Today view
- gcal-tui DeepWiki documentation — event rendering patterns, status markers
- calcure GitHub repository — dual-pane event/task layout
- WTF dashboard Google Calendar module — event display format
- DAVx5 Google Calendar compatibility notes — OAuth requirement confirmation
- Home Assistant CalDAV issues (#126448, #25814) — all-day event timezone bugs
- Cal.com blog post on CalDAV implementation — RRULE complexity, timezone challenges

### Tertiary (LOW confidence, needs validation)
- CalDAV read-only scope caveat (community report) — calendar.readonly may cause 400 errors with CalDAV protocol, needs testing
- Google CalDAV cookie behavior — reported in some CalDAV clients, needs verification
- Rate limit specifics for Google Calendar API — documented default is 1M req/day, but burst limits undocumented

---
*Research completed: 2026-02-13*
*Ready for roadmap: yes*
