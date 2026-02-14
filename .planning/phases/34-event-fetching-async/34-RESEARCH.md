# Phase 34: Event Fetching & Async Integration - Research

**Researched:** 2026-02-14
**Domain:** Google Calendar REST API event fetching, Bubble Tea async patterns, incremental sync
**Confidence:** HIGH

## Summary

This phase adds Google Calendar event fetching to the TUI app, building on Phase 33's `internal/google/auth.go` which provides `TokenSource()`. The core challenge is threefold: (1) calling the Google Calendar Events.List REST API with the correct parameters for recurring events and all-day events, (2) integrating async HTTP calls into the Bubble Tea event loop without freezing the TUI, and (3) implementing efficient incremental sync via syncToken to minimize API calls during 5-minute polling.

The Google Calendar Go client library (`google.golang.org/api/calendar/v3`) provides a type-safe wrapper around the REST API. It accepts an `oauth2.TokenSource` directly via `option.WithTokenSource()`, which fits perfectly with Phase 33's `TokenSource()` function. The `Events.List` call with `SingleEvents(true)` expands recurring events server-side (FETCH-06), and the `EventDateTime` struct's `Date` vs `DateTime` fields cleanly distinguish all-day from timed events (FETCH-07).

For Bubble Tea integration, the standard pattern is: `tea.Cmd` functions run HTTP calls in goroutines, return result messages to `Update()`, and `tea.Tick` or `tea.Every` schedules periodic re-fetches. The app already uses this pattern for OAuth (`google.StartAuthFlow()` returns a `tea.Cmd`). The key addition is a polling loop: on startup fire a fetch command, on success store events and schedule the next tick, on error preserve last-known data and schedule retry.

**Primary recommendation:** Use `google.golang.org/api/calendar/v3` with `option.WithTokenSource()`, `SingleEvents(true)`, syncToken-based delta sync, and a `tea.Tick`-based 5-minute polling loop. New file: `internal/google/events.go` in the existing `google` package.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| google.golang.org/api/calendar/v3 | latest (module google.golang.org/api) | Type-safe Google Calendar API client | Official Google-maintained Go client, auto-generated from API discovery doc |
| google.golang.org/api/option | (same module) | `WithTokenSource()` to create authenticated service | Standard way to pass credentials to Google API services |
| golang.org/x/oauth2 | v0.35.0 (already in go.mod) | `TokenSource` interface consumed by the calendar client | Already used by Phase 33 |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/charmbracelet/bubbletea | v1.3.10 (already in go.mod) | `tea.Tick` for polling, `tea.Cmd` for async fetch | Scheduling periodic fetches |
| time (stdlib) | - | Duration for polling interval, RFC3339 formatting | TimeMin/TimeMax parameters, tick interval |
| context (stdlib) | - | Request cancellation, timeout for API calls | Each API call should have a context with timeout |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| google.golang.org/api/calendar/v3 | Raw net/http REST calls | Lose type safety, pagination helpers, field builders; no benefit |
| tea.Tick polling | Channel-based goroutine (realtime example) | tea.Tick is simpler for fixed intervals; channel pattern is for variable-rate events |
| In-memory cache | SQLite persistence | Requirements explicitly say in-memory only (FETCH-04) |

**Installation:**
```bash
go get google.golang.org/api/calendar/v3
go get google.golang.org/api/option
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
  google/
    auth.go           # [existing] OAuth config, TokenSource(), auth flow
    auth_test.go      # [existing] Auth tests
    events.go         # [new] CalendarEvent type, FetchEvents(), EventStore
    events_test.go    # [new] Tests for event parsing, sync logic
```

Keep events in the same `google` package since it shares `TokenSource()` and the auth state. No need for a separate package.

### Pattern 1: Event Data Type (Decoupled from Google API)
**What:** Define a local `CalendarEvent` struct that the rest of the app uses, decoupled from `calendar.Event`.
**When to use:** Always. Don't leak Google API types into the TUI layer.

```go
// internal/google/events.go

// CalendarEvent is a normalized calendar event for display in the TUI.
type CalendarEvent struct {
    ID       string
    Summary  string
    Date     string // "2006-01-02" for all-day events
    Start    time.Time // zero for all-day events
    End      time.Time // zero for all-day events
    AllDay   bool
    Status   string // "confirmed", "tentative", "cancelled"
}
```

### Pattern 2: Service Creation from TokenSource
**What:** Create a `calendar.Service` from the existing `TokenSource()`.
**When to use:** When initializing the event fetcher.

```go
// Source: Google Calendar API Go docs, option.WithTokenSource pattern
import (
    "context"
    "google.golang.org/api/calendar/v3"
    "google.golang.org/api/option"
)

func newCalendarService(ts oauth2.TokenSource) (*calendar.Service, error) {
    ctx := context.Background()
    return calendar.NewService(ctx, option.WithTokenSource(ts))
}
```

### Pattern 3: Fetch with SyncToken Delta Sync
**What:** First call does full fetch with time bounds; subsequent calls use syncToken for incremental updates.
**When to use:** Every fetch cycle (FETCH-03).

```go
func fetchEvents(srv *calendar.Service, syncToken string) ([]CalendarEvent, string, error) {
    call := srv.Events.List("primary").
        SingleEvents(true).
        ShowDeleted(true). // Required when using syncToken
        MaxResults(2500)

    if syncToken != "" {
        // Incremental sync: only get changes
        // Note: timeMin/timeMax CANNOT be used with syncToken
        call = call.SyncToken(syncToken)
    } else {
        // Full sync: fetch a reasonable window
        now := time.Now()
        timeMin := now.AddDate(0, -1, 0).Format(time.RFC3339)
        timeMax := now.AddDate(0, 3, 0).Format(time.RFC3339)
        call = call.TimeMin(timeMin).TimeMax(timeMax).OrderBy("startTime")
    }

    var allEvents []*calendar.Event
    var nextSyncToken string

    // Paginate through all results
    for {
        events, err := call.Do()
        if err != nil {
            return nil, "", err
        }
        allEvents = append(allEvents, events.Items...)

        if events.NextPageToken != "" {
            call = call.PageToken(events.NextPageToken)
        } else {
            nextSyncToken = events.NextSyncToken
            break
        }
    }

    // Convert to local type
    result := make([]CalendarEvent, 0, len(allEvents))
    for _, e := range allEvents {
        result = append(result, convertEvent(e))
    }
    return result, nextSyncToken, nil
}
```

### Pattern 4: All-Day Event Date Parsing (FETCH-07)
**What:** All-day events use `Start.Date` (YYYY-MM-DD string), timed events use `Start.DateTime` (RFC3339).
**When to use:** Converting Google API events to local CalendarEvent.

```go
func convertEvent(e *calendar.Event) CalendarEvent {
    ce := CalendarEvent{
        ID:      e.Id,
        Summary: e.Summary,
        Status:  e.Status,
    }

    if e.Start.Date != "" {
        // All-day event: Date is "2006-01-02" string
        // CRITICAL: Do NOT parse with timezone - use as-is to prevent off-by-one
        ce.AllDay = true
        ce.Date = e.Start.Date // Store raw date string
    } else if e.Start.DateTime != "" {
        // Timed event: DateTime is RFC3339
        t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
        ce.Start = t
        ce.Date = t.Format("2006-01-02")
        if e.End.DateTime != "" {
            ce.End, _ = time.Parse(time.RFC3339, e.End.DateTime)
        }
    }

    return ce
}
```

### Pattern 5: Bubble Tea Polling Loop
**What:** Use `tea.Cmd` for async fetch, `tea.Tick` for scheduling next poll.
**When to use:** FETCH-01 (startup fetch) and FETCH-02 (5-minute polling).

```go
// Message types
type EventsFetchedMsg struct {
    Events    []google.CalendarEvent
    SyncToken string
    Err       error
}

type eventTickMsg time.Time

// Schedule next fetch
func scheduleEventTick() tea.Cmd {
    return tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
        return eventTickMsg(t)
    })
}

// Async fetch command
func fetchEventsCmd(srv *calendar.Service, syncToken string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        // Use ctx for the API call
        events, newToken, err := fetchEvents(srv, syncToken)
        return EventsFetchedMsg{Events: events, SyncToken: newToken, Err: err}
    }
}

// In app.Model Update():
case EventsFetchedMsg:
    if msg.Err != nil {
        // FETCH-05: Keep last known data, don't crash
        // Optionally log error, schedule retry
        return m, scheduleEventTick()
    }
    m.calendarEvents = msg.Events
    m.eventsSyncToken = msg.SyncToken
    return m, scheduleEventTick()

case eventTickMsg:
    if m.googleAuthState != google.AuthReady {
        return m, scheduleEventTick() // Skip fetch, retry later
    }
    return m, fetchEventsCmd(m.calendarSvc, m.eventsSyncToken)
```

### Pattern 6: 410 GONE Recovery
**What:** When syncToken expires, Google returns 410. Clear token and do full re-fetch.
**When to use:** In error handling of fetchEvents.

```go
import "google.golang.org/api/googleapi"

func fetchEvents(srv *calendar.Service, syncToken string) ([]CalendarEvent, string, error) {
    // ... build call ...
    events, err := call.Do()
    if err != nil {
        if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 410 {
            // SyncToken expired - do full sync
            return fetchEvents(srv, "") // Retry without syncToken
        }
        return nil, "", err
    }
    // ...
}
```

### Anti-Patterns to Avoid
- **Blocking the TUI with HTTP calls:** Never call `Events.List.Do()` synchronously in `Update()`. Always wrap in a `tea.Cmd`.
- **Parsing all-day event dates with `time.Parse` in a timezone:** The `Date` field is "2006-01-02" without timezone. Parsing with `time.ParseInLocation` using local timezone can shift the date. Store as plain string.
- **Using timeMin/timeMax with syncToken:** The API returns 400 if you combine these. SyncToken replaces time-based filtering for incremental updates.
- **Ignoring pagination:** A user with many events can have paginated results. Always loop until `NextPageToken` is empty.
- **Dropping cancelled events during sync:** When using syncToken, cancelled events (status="cancelled") appear in results to signal deletion. Must remove them from the local cache.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Calendar API HTTP client | Manual REST calls with net/http | `google.golang.org/api/calendar/v3` | Handles auth headers, pagination, error types, field masking |
| Event instance expansion | Manual RRULE parsing for recurring events | `SingleEvents(true)` API parameter | Server-side expansion is correct, handles exceptions, time zones |
| Token refresh during API calls | Manual refresh before each call | `option.WithTokenSource()` with auto-refreshing TokenSource | Phase 33's persistingTokenSource handles this transparently |
| Delta sync protocol | Manual change tracking | syncToken + 410 GONE handling | Google's protocol handles all edge cases |

**Key insight:** The Google Calendar Go client library does all the heavy lifting. The only custom code is: (1) converting API types to local types, (2) managing the syncToken, (3) Bubble Tea integration (messages and commands), and (4) the in-memory event store.

## Common Pitfalls

### Pitfall 1: timeMin/timeMax with syncToken Returns 400
**What goes wrong:** API returns 400 Bad Request.
**Why it happens:** `timeMin`, `timeMax`, `orderBy`, `iCalUID`, `q`, `sharedExtendedProperty`, `privateExtendedProperty`, and `updatedMin` cannot be combined with `syncToken`.
**How to avoid:** Only set time bounds on the initial full sync (when syncToken is empty). On incremental sync, use only `syncToken` + `singleEvents` + `showDeleted`.
**Warning signs:** 400 error on the second fetch cycle.

### Pitfall 2: All-Day Events Off by One Day
**What goes wrong:** An all-day event on Feb 14 shows on Feb 13 or Feb 15.
**Why it happens:** The `Date` field is "2026-02-14" with no timezone. If you parse it with `time.Parse(time.RFC3339, ...)` or `time.ParseInLocation("2006-01-02", date, time.Local)`, timezone offsets can shift the date.
**How to avoid:** Treat all-day event dates as opaque strings. Store them as "2006-01-02" strings and compare directly. Never convert to `time.Time` for display purposes.
**Warning signs:** Events appear on wrong days for users in non-UTC timezones.

### Pitfall 3: Missing ShowDeleted During Sync
**What goes wrong:** Deleted events remain in local cache forever.
**Why it happens:** Without `ShowDeleted(true)`, the API omits cancelled events from sync results.
**How to avoid:** Always set `ShowDeleted(true)` when using syncToken. Filter out `status == "cancelled"` events from display but use them to remove from cache.
**Warning signs:** Deleted events keep appearing; event count only grows.

### Pitfall 4: TUI Freezes During Network Calls
**What goes wrong:** The TUI becomes unresponsive for 1-30 seconds during event fetch.
**Why it happens:** API call runs synchronously in `Update()` instead of as a `tea.Cmd`.
**How to avoid:** All HTTP calls must be wrapped in `tea.Cmd` (which runs in a goroutine). Use `context.WithTimeout` to bound wait time.
**Warning signs:** TUI stops responding to keypresses during fetch.

### Pitfall 5: Goroutine Leak on App Exit
**What goes wrong:** Pending HTTP request goroutine hangs after user quits.
**Why it happens:** `tea.Cmd` goroutines run independently. If one is blocked on an HTTP call when the user presses 'q', it continues until timeout.
**How to avoid:** Use `context.WithTimeout` (30 seconds max) on all API calls. The goroutine will clean up when the context expires. This is acceptable for a short-lived CLI app.
**Warning signs:** Process hangs briefly after quit.

### Pitfall 6: SyncToken Expiration Not Handled
**What goes wrong:** After the app has been closed for a long time, the next fetch with the old syncToken returns 410 GONE and the app shows no events.
**Why it happens:** SyncTokens have an undocumented expiration period.
**How to avoid:** Catch `googleapi.Error` with code 410, clear the syncToken, and retry with a full sync. Since FETCH-04 says in-memory only, the syncToken is naturally cleared on app restart.
**Warning signs:** 410 errors in logs, empty event list.

### Pitfall 7: Polling Starts Before Auth is Ready
**What goes wrong:** Fetch command runs but TokenSource returns an error because auth is not configured.
**Why it happens:** Polling timer fires regardless of auth state.
**How to avoid:** Check `googleAuthState == AuthReady` before creating the calendar service and firing fetch commands. If not ready, just schedule the next tick and skip.
**Warning signs:** Error messages on startup for users who haven't configured Google Calendar.

## Code Examples

### Complete Service Initialization
```go
// Source: google.golang.org/api/calendar/v3 + option.WithTokenSource pattern
import (
    "google.golang.org/api/calendar/v3"
    "google.golang.org/api/option"
)

// NewCalendarService creates an authenticated Calendar API service.
func NewCalendarService() (*calendar.Service, error) {
    ts, err := TokenSource()
    if err != nil {
        return nil, fmt.Errorf("token source: %w", err)
    }
    return calendar.NewService(
        context.Background(),
        option.WithTokenSource(ts),
    )
}
```

### In-Memory Event Store
```go
// EventStore holds fetched calendar events in memory.
// Thread-safe for concurrent access from tea.Cmd goroutines.
type EventStore struct {
    mu        sync.RWMutex
    events    []CalendarEvent
    syncToken string
}

func (s *EventStore) Update(events []CalendarEvent, syncToken string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // For incremental sync: merge changes
    // For full sync: replace all
    s.events = events
    s.syncToken = syncToken
}

func (s *EventStore) EventsForDate(date string) []CalendarEvent {
    s.mu.RLock()
    defer s.mu.RUnlock()
    var result []CalendarEvent
    for _, e := range s.events {
        if e.Date == date {
            result = append(result, e)
        }
    }
    return result
}
```

### App Model Integration Points
```go
// Fields to add to app.Model:
type Model struct {
    // ... existing fields ...
    calendarSvc      *calendar.Service // nil if auth not ready
    calendarEvents   []google.CalendarEvent
    eventsSyncToken  string
    eventsFetchErr   error
}

// In Init() - start fetch if auth is ready:
func (m Model) Init() tea.Cmd {
    if m.googleAuthState == google.AuthReady {
        svc, err := google.NewCalendarService()
        if err == nil {
            m.calendarSvc = svc
            return tea.Batch(
                fetchEventsCmd(svc, ""),
                scheduleEventTick(),
            )
        }
    }
    return scheduleEventTick() // Still schedule tick for when auth becomes ready
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| google-api-go-client as separate module | Part of google.golang.org/api monorepo | 2020+ | Single `go get google.golang.org/api` for all Google APIs |
| Manual HTTP client construction | `option.WithTokenSource()` | Available since early versions | Cleaner integration with oauth2.TokenSource |
| Polling only | Push notifications via webhooks | Available | Not applicable for desktop TUI apps; polling is correct for this use case |

**Deprecated/outdated:**
- The old `google.golang.org/api/googleapi/transport` package for auth is deprecated; use `option.WithTokenSource()` instead.

## Open Questions

1. **Time window for full sync**
   - What we know: Full sync needs timeMin/timeMax bounds. Requirements don't specify how far back/forward to fetch.
   - What's unclear: Should we fetch 1 month back + 3 months forward? 1 week back + 1 month forward?
   - Recommendation: 1 month back + 3 months forward. This covers the visible calendar range and is a reasonable default. Can be adjusted later.

2. **Incremental sync merge strategy**
   - What we know: SyncToken returns only changed events. Cancelled events need removal.
   - What's unclear: Since we store in-memory only and syncToken resets on restart, do we even need the merge logic during a single session? The first fetch is always full.
   - Recommendation: Implement merge for correctness (handle mid-session changes). On full sync replace all events; on incremental sync, update/add changed events and remove cancelled ones by ID.

3. **Where to display events in the TUI**
   - What we know: Phase 34 is about fetching and caching. Display integration may be Phase 35.
   - What's unclear: Whether this phase should also add event display to the todolist/calendar views.
   - Recommendation: This phase should focus on fetch infrastructure and expose events via a queryable interface. Minimal display (e.g., event count indicator) can validate the fetch works, but full TUI rendering may be a separate phase.

4. **Rate limiting / quota**
   - What we know: Google Calendar API has a quota (typically 1,000,000 queries/day for most projects, but per-user rate limits apply).
   - What's unclear: Exact per-user rate limits for personal-use GCP projects.
   - Recommendation: 5-minute polling (FETCH-02) is well within any rate limit. No special handling needed.

## Sources

### Primary (HIGH confidence)
- [Google Calendar Events.list API reference](https://developers.google.com/workspace/calendar/api/v3/reference/events/list) - Parameters, response format, syncToken constraints
- [Google Calendar Sync Guide](https://developers.google.com/workspace/calendar/api/guides/sync) - Full/incremental sync protocol, 410 handling
- [google.golang.org/api/calendar/v3 Go package](https://pkg.go.dev/google.golang.org/api/calendar/v3) - Event struct, EventDateTime, EventsListCall methods
- [Google Calendar Go Quickstart](https://developers.google.com/workspace/calendar/api/quickstart/go) - Service creation, Events.List usage pattern
- [Bubble Tea realtime example](https://github.com/charmbracelet/bubbletea/blob/main/examples/realtime/main.go) - Async command pattern for background fetching

### Secondary (MEDIUM confidence)
- [Bubble Tea commands documentation](https://charm.land/blog/commands-in-bubbletea/) - tea.Cmd, tea.Tick, tea.Every patterns
- [googleapis/google-api-go-client source](https://github.com/googleapis/google-api-go-client/blob/main/calendar/v3/calendar-gen.go) - Event struct fields, EventDateTime definition

### Tertiary (LOW confidence)
- [singleEvents + syncToken compatibility issue (PHP)](https://github.com/googleapis/google-api-php-client/issues/2637) - Reports singleEvents=true prevents syncToken return. However, official Google sync guide shows them used together in example code, and the parameters-prohibited-with-syncToken list does not include singleEvents. Likely a PHP client library issue, not an API limitation. Flag for validation during implementation.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - `google.golang.org/api/calendar/v3` is the official Go client, verified via pkg.go.dev and Google's own quickstart
- Architecture: HIGH - Follows existing project patterns (tea.Cmd for async, same google package) and Google's documented sync protocol
- Pitfalls: HIGH - syncToken constraints, all-day date handling, and 410 GONE are all documented in official API reference
- Bubble Tea integration: HIGH - Pattern matches existing auth flow integration and official examples

**Research date:** 2026-02-14
**Valid until:** 2026-03-14 (stable domain: Google Calendar API v3 and Bubble Tea patterns are mature)
