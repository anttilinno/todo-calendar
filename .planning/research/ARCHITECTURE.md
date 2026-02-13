# Architecture Research: CalDAV Google Calendar Integration

**Domain:** Read-only Google Calendar event display via CalDAV in existing Go/Bubble Tea TUI app
**Researched:** 2026-02-13
**Confidence:** HIGH for Bubble Tea integration patterns (well-documented, consistent with existing codebase); MEDIUM for CalDAV client library (API verified via pkg.go.dev, OAuth2 flow verified via Google docs); MEDIUM for Google CalDAV OAuth2 (CLI token flow is standard but CalDAV-specific scoping has a known caveat about read-only vs read-write)

## Current Architecture Summary (Post v2.1)

```
main.go
  |
  app.Model (root orchestrator)
  |-- calendar.Model    (left pane: grid + overview)
  |-- todolist.Model    (right pane: 4-section todo list + 5-field edit form)
  |-- settings.Model    (full-screen overlay, showSettings bool)
  |-- search.Model      (full-screen overlay, showSearch bool)
  |-- preview.Model     (full-screen overlay, showPreview bool)
  |-- tmplmgr.Model     (full-screen overlay, showTmplMgr bool)
  |-- editor            (external process, editing bool)
  |
  Dependencies injected via constructors:
  |-- store.TodoStore   (interface, SQLite implementation)
  |-- holidays.Provider (struct, holiday lookup)
  |-- config.Config     (TOML struct)
  |-- theme.Theme       (color palette struct)
```

**Key architecture patterns already in use:**
- Elm Architecture: Init/Update/View on all models
- Constructor DI: `app.New(provider, mondayStart, store, theme, cfg)`
- Pure rendering: `RenderGrid()` and `RenderWeekGrid()` are pure functions taking data, returning strings
- Message routing: Parent `app.Model.Update()` handles cross-component messages via type switch
- Overlay pattern: `showX bool` + `x.Model` field pair for modal overlays
- Store interface: `store.TodoStore` decouples all consumers from SQLite

## Recommended Architecture for CalDAV Integration

### High-Level Component Diagram

```
main.go
  |
  +-- [NEW] OAuth2 token bootstrap (one-time browser flow on first run)
  |
  app.Model (root orchestrator)
  |-- calendar.Model
  |-- todolist.Model
  |-- [MODIFIED] app.Model fields: caldavEvents map, caldavSyncing bool
  |
  +-- [NEW] internal/caldav/          (CalDAV client + event cache)
  |     |-- client.go                 (CalDAV client wrapper)
  |     |-- event.go                  (CalendarEvent domain type)
  |     |-- cache.go                  (in-memory event cache by date)
  |     |-- oauth.go                  (Google OAuth2 token management)
  |     |-- messages.go               (tea.Msg types for sync results)
  |
  +-- [MODIFIED] internal/config/     (add CalDAV config fields)
  +-- [MODIFIED] internal/calendar/   (render events on grid days)
  +-- [MODIFIED] internal/todolist/   (show events in mixed list)
```

### New Package: `internal/caldav/`

This is the core new package. It follows the same patterns as existing packages:
- Constructor DI (takes config, returns struct)
- Exports message types for Bubble Tea
- No Bubble Tea model -- it is a service, not a UI component

**Rationale for separate package (not extending store):** CalDAV events are read-only, ephemeral, and externally sourced. They do not belong in `store.TodoStore` which manages mutable local data. The CalDAV package owns its own data lifecycle (fetch, cache, expire).

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| `caldav.Client` | Wraps go-webdav CalDAV client, handles OAuth2 HTTP client creation, fetches calendar objects, parses into domain events | `caldav.Cache`, `caldav.OAuth` |
| `caldav.Cache` | Thread-safe in-memory map of `date -> []CalendarEvent`, expiry tracking, date range queries | `caldav.Client` (populated by), `app.Model` (queried by) |
| `caldav.OAuth` | Manages Google OAuth2 token lifecycle: load from disk, refresh, first-run browser flow | `caldav.Client` (provides http.Client to) |
| `caldav.CalendarEvent` | Domain type: Summary, Start, End, AllDay bool, UID, CalendarName | All consumers |
| `app.Model` | Starts sync commands, receives sync results, distributes events to child models | `caldav.Client`, `calendar.Model`, `todolist.Model` |
| `calendar.Model` | Renders event indicators on grid days alongside todo indicators | `app.Model` (receives event data via setter) |
| `todolist.Model` | Renders calendar events in a new "Events" section alongside todo sections | `app.Model` (receives event data via setter) |

### Data Flow: CalDAV Events from Network to Screen

```
STARTUP:
  main.go
    -> config.Load() reads CalDAV config (enabled, calendar IDs, token path)
    -> if caldav enabled:
         caldav.NewOAuth(tokenPath) loads/validates token
         caldav.NewClient(oauthClient, endpoint) creates CalDAV client
         caldav.NewCache() creates empty cache
    -> app.New(..., caldavClient, caldavCache)
    -> app.Init() returns tea.Batch(syncCalDAVCmd(client, cache, dateRange))

SYNC CYCLE (every 5 minutes + on month navigation):
  app.Init() or app.Update(monthChanged)
    -> returns syncCalDAVCmd(client, cache, year, month)
    |
    | [runs in goroutine via tea.Cmd]
    |
    -> caldav.Client.FetchEvents(ctx, calendarPath, startDate, endDate)
         -> go-webdav QueryCalendar() with CompFilter for VEVENT + time-range
         -> parse CalendarObjects into []CalendarEvent
         -> cache.Store(events)
    -> returns CalDAVSyncDoneMsg{events, err}
    |
    | [back in main event loop]
    |
  app.Update(CalDAVSyncDoneMsg)
    -> if err: set caldavError string, log, continue (non-fatal)
    -> cache events in app.Model or pass cache reference
    -> calendar.SetCalendarEvents(cache.EventsForMonth(year, month))
    -> todolist.SetCalendarEvents(cache.EventsForDateRange(start, end))
    -> return tea.Tick(5*time.Minute, func(t) tea.Msg { return CalDAVPollMsg{} })

POLL TICK:
  app.Update(CalDAVPollMsg{})
    -> return syncCalDAVCmd(client, cache, currentYear, currentMonth)

MONTH NAVIGATION:
  app.Update(tea.KeyMsg) [after calendar navigates]
    -> syncTodoView() [existing]
    -> calendar.SetCalendarEvents(cache.EventsForMonth(newYear, newMonth))
    -> todolist.SetCalendarEvents(cache.EventsForDateRange(...))
    -> if cache miss for new month: return syncCalDAVCmd for new range
```

### Bubble Tea Message Types (in `caldav/messages.go`)

```go
package caldav

import "time"

// CalDAVSyncDoneMsg is sent when a background CalDAV sync completes.
type CalDAVSyncDoneMsg struct {
    Events []CalendarEvent
    Err    error
}

// CalDAVPollMsg triggers the next periodic sync.
type CalDAVPollMsg struct {
    At time.Time
}

// CalDAVAuthRequiredMsg indicates OAuth token is missing or expired
// and cannot be auto-refreshed.
type CalDAVAuthRequiredMsg struct {
    AuthURL string
}
```

### Bubble Tea Command Pattern (in `app.Model`)

The sync command follows the standard Bubble Tea async pattern. It is a `tea.Cmd` (a function returning a `tea.Msg`) that runs in its own goroutine. This is the same pattern used by `editor.Open()` in the existing codebase.

```go
// syncCalDAVCmd returns a tea.Cmd that fetches CalDAV events in the background.
func syncCalDAVCmd(client *caldav.Client, year int, month time.Month) tea.Cmd {
    return func() tea.Msg {
        start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
        end := start.AddDate(0, 1, 0)
        events, err := client.FetchEvents(context.Background(), start, end)
        return caldav.CalDAVSyncDoneMsg{Events: events, Err: err}
    }
}
```

The poll is a standard `tea.Tick` returning to the Update loop:

```go
func pollCalDAVCmd() tea.Cmd {
    return tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
        return caldav.CalDAVPollMsg{At: t}
    })
}
```

**Why tea.Cmd and not raw channels:** Bubble Tea's `tea.Cmd` already runs in a goroutine and delivers the result as a `tea.Msg` into the main event loop. Using raw channels (like the `realtime` example) is unnecessary here because each sync is a discrete request-response, not a continuous stream. The `tea.Tick` pattern handles periodic scheduling. Using `tea.Cmd` keeps the architecture consistent with how `editor.Open()` already works in this codebase.

### CalDAV Client Wrapper (in `caldav/client.go`)

```go
package caldav

import (
    "context"
    "time"

    gocaldav "github.com/emersion/go-webdav/caldav"
    "github.com/emersion/go-ical"
)

// Client wraps the go-webdav CalDAV client for Google Calendar access.
type Client struct {
    inner        *gocaldav.Client
    calendarPath string  // e.g., "/caldav/v2/{calid}/events"
}

// NewClient creates a CalDAV client using the provided authenticated HTTP client.
func NewClient(httpClient webdav.HTTPClient, endpoint string, calendarPath string) (*Client, error) {
    inner, err := gocaldav.NewClient(httpClient, endpoint)
    if err != nil {
        return nil, err
    }
    return &Client{inner: inner, calendarPath: calendarPath}, nil
}

// FetchEvents queries VEVENT objects in the given time range.
func (c *Client) FetchEvents(ctx context.Context, start, end time.Time) ([]CalendarEvent, error) {
    query := &gocaldav.CalendarQuery{
        CompRequest: gocaldav.CalendarCompRequest{
            Name:  "VCALENDAR",
            Props: []string{"VERSION"},
            Comps: []gocaldav.CalendarCompRequest{{
                Name:  "VEVENT",
                Props: []string{"SUMMARY", "DTSTART", "DTEND", "UID", "DESCRIPTION"},
            }},
        },
        CompFilter: gocaldav.CompFilter{
            Name: "VCALENDAR",
            Comps: []gocaldav.CompFilter{{
                Name:  "VEVENT",
                Start: start,
                End:   end,
            }},
        },
    }

    objects, err := c.inner.QueryCalendar(ctx, c.calendarPath, query)
    if err != nil {
        return nil, err
    }

    var events []CalendarEvent
    for _, obj := range objects {
        for _, ev := range obj.Data.Events() {
            event, err := parseEvent(ev)
            if err != nil {
                continue // skip unparseable events
            }
            events = append(events, event)
        }
    }
    return events, nil
}
```

### CalendarEvent Domain Type (in `caldav/event.go`)

```go
package caldav

import "time"

// CalendarEvent represents a single Google Calendar event for display.
// This is a read-only display type -- no mutation methods.
type CalendarEvent struct {
    UID          string
    Summary      string    // event title
    Start        time.Time
    End          time.Time
    AllDay       bool
    Description  string    // optional body text
    CalendarName string    // which calendar this came from
}

// Date returns the event's start date as "YYYY-MM-DD" for consistency
// with the todo date format used throughout the codebase.
func (e CalendarEvent) Date() string {
    return e.Start.Format("2006-01-02")
}

// InMonth reports whether the event starts in the given year and month.
func (e CalendarEvent) InMonth(year int, month time.Month) bool {
    return e.Start.Year() == year && e.Start.Month() == month
}

// TimeRange returns a formatted time range string for display.
// All-day events return "All day". Timed events return "HH:MM - HH:MM".
func (e CalendarEvent) TimeRange() string {
    if e.AllDay {
        return "All day"
    }
    return e.Start.Format("15:04") + " - " + e.End.Format("15:04")
}
```

### In-Memory Event Cache (in `caldav/cache.go`)

```go
package caldav

import (
    "sync"
    "time"
)

// Cache stores CalDAV events in memory indexed by date string.
// Thread-safe for concurrent read/write from Bubble Tea commands.
type Cache struct {
    mu     sync.RWMutex
    events map[string][]CalendarEvent // "2026-02-13" -> events
    lastSync time.Time
}

func NewCache() *Cache {
    return &Cache{events: make(map[string][]CalendarEvent)}
}

// Store replaces cached events with new data.
func (c *Cache) Store(events []CalendarEvent) {
    c.mu.Lock()
    defer c.mu.Unlock()
    // Clear existing and repopulate
    c.events = make(map[string][]CalendarEvent)
    for _, e := range events {
        key := e.Date()
        c.events[key] = append(c.events[key], e)
    }
    c.lastSync = time.Now()
}

// EventsForDate returns events on a specific date.
func (c *Cache) EventsForDate(date string) []CalendarEvent {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.events[date]
}

// EventsForMonth returns all events in a given month.
func (c *Cache) EventsForMonth(year int, month time.Month) []CalendarEvent {
    c.mu.RLock()
    defer c.mu.RUnlock()
    var result []CalendarEvent
    for _, evts := range c.events {
        for _, e := range evts {
            if e.InMonth(year, month) {
                result = append(result, e)
            }
        }
    }
    return result
}
```

**Why in-memory, not SQLite:** Calendar events are ephemeral -- they change on the server and must be re-fetched regularly. Persisting to SQLite adds complexity (schema, migrations, stale data cleanup) with no benefit for a read-only display. The cache is rebuilt on every sync. Memory usage is trivial (a month of events is typically 50-200 structs).

### OAuth2 Token Management (in `caldav/oauth.go`)

Google CalDAV requires OAuth2 -- app passwords are NOT supported (Google disabled all non-OAuth auth in June 2023).

The token flow for a CLI app:

1. **First run:** No token file exists. Print an auth URL to the terminal. User visits URL in browser, authorizes, gets auth code. User pastes code into terminal. Exchange code for token. Save token to disk.
2. **Subsequent runs:** Load token from disk. The `oauth2.Config.Client()` auto-refreshes expired access tokens using the stored refresh token.
3. **Token expired and refresh fails:** Print auth URL again (re-authorize).

```go
package caldav

import (
    "context"
    "encoding/json"
    "net/http"
    "os"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

const (
    // Google CalDAV requires this scope. The readonly scope reportedly causes
    // 400 errors with CalDAV protocol, so full calendar scope is needed even
    // for read-only usage.
    calendarScope = "https://www.googleapis.com/auth/calendar"
)

// OAuthConfig holds the Google OAuth2 configuration for CalDAV.
type OAuthConfig struct {
    ClientID     string
    ClientSecret string
    TokenPath    string // path to save/load oauth2.Token JSON
}

// NewOAuthHTTPClient creates an http.Client with OAuth2 credentials.
// Returns (client, nil) on success, or (nil, CalDAVAuthRequiredMsg) if
// interactive auth is needed.
func NewOAuthHTTPClient(ctx context.Context, cfg OAuthConfig) (*http.Client, error) {
    oauthCfg := &oauth2.Config{
        ClientID:     cfg.ClientID,
        ClientSecret: cfg.ClientSecret,
        Endpoint:     google.Endpoint,
        Scopes:       []string{calendarScope},
        RedirectURL:  "urn:ietf:wg:oauth:2.0:oob", // out-of-band for CLI
    }

    tok, err := loadToken(cfg.TokenPath)
    if err != nil {
        // No token -- need interactive auth
        return nil, err
    }

    // oauth2 auto-refreshes the token if expired
    return oauthCfg.Client(ctx, tok), nil
}

func loadToken(path string) (*oauth2.Token, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    var tok oauth2.Token
    return &tok, json.NewDecoder(f).Decode(&tok)
}

func saveToken(path string, token *oauth2.Token) error {
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        return err
    }
    defer f.Close()
    return json.NewEncoder(f).Encode(token)
}
```

**Where the OAuth client ID/secret come from:** The user must create a Google Cloud project and OAuth2 credentials. The client ID and secret are stored in the TOML config file. This is the standard approach for CLI apps (gcalcli, calcurse-caldav, khal/vdirsyncer all do this). The token file (containing access+refresh tokens) is stored alongside the config in XDG config dir with 0600 permissions.

### Config Extension (in `config/config.go`)

```go
// Config holds the application configuration.
type Config struct {
    // ... existing fields ...

    // CalDAV integration
    CalDAV CalDAVConfig `toml:"caldav"`
}

// CalDAVConfig holds Google Calendar CalDAV settings.
type CalDAVConfig struct {
    Enabled      bool   `toml:"enabled"`
    CalendarID   string `toml:"calendar_id"`   // e.g., "primary" or email
    ClientID     string `toml:"client_id"`
    ClientSecret string `toml:"client_secret"`
    PollMinutes  int    `toml:"poll_minutes"`  // default: 5
}
```

The TOML config file would look like:

```toml
# Existing settings
country = "us"
first_day_of_week = "sunday"
theme = "dark"
date_format = "iso"
show_month_todos = true
show_year_todos = true

# CalDAV integration
[caldav]
enabled = true
calendar_id = "primary"
client_id = "YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com"
client_secret = "YOUR_GOOGLE_CLIENT_SECRET"
poll_minutes = 5
```

**Token path is derived, not configured:** Stored as `~/.config/todo-calendar/caldav-token.json` (same XDG directory as the existing config and database). This follows the existing `config.Path()` / `config.DBPath()` pattern in `config/paths.go`.

### Calendar Grid Integration (in `calendar/grid.go`)

The calendar grid needs to show event indicators on days with Google Calendar events. This parallels the existing todo indicator system (`indicators map[int]int`, `totals map[int]int`).

**Approach: Add event count map to grid rendering.**

```go
// In calendar.Model, add:
type Model struct {
    // ... existing fields ...
    eventCounts map[int]int // day -> count of CalDAV events (set by app.Model)
}

// SetCalendarEvents updates the event counts for the current month.
func (m *Model) SetCalendarEvents(events []caldav.CalendarEvent) {
    m.eventCounts = make(map[int]int)
    for _, e := range events {
        if e.InMonth(m.year, m.month) {
            m.eventCounts[e.Start.Day()]++
        }
    }
}
```

In `RenderGrid()`, add a new parameter `eventCounts map[int]int` and render a small indicator (like a dot character) for days with events. This keeps RenderGrid pure -- it receives all data as parameters and returns a string.

**Visual indicator suggestion:** A small dot or calendar symbol next to the day number for days with CalDAV events. The existing `[NN]` bracket indicator is for todos. Events could use a different character like a small diamond or underline to distinguish them.

### Todo List Integration (in `todolist/model.go`)

The todolist panel needs a new "Events" section showing CalDAV events for the viewed date range, visually distinct from todos (not interactive -- no toggle/delete/edit).

**Approach: Extend `visibleItems()` with a new section.**

```go
// Add new section and item kind constants
const (
    sectionEvents sectionID = 4 // after sectionFloating
)

// In visibleItems(), after floating section, add:
if len(m.calendarEvents) > 0 {
    items = append(items, visibleItem{
        kind:    headerItem,
        label:   "Calendar",
        section: sectionEvents,
    })
    for i := range m.calendarEvents {
        items = append(items, visibleItem{
            kind:    eventItem, // new itemKind
            event:   &m.calendarEvents[i],
            section: sectionEvents,
        })
    }
}
```

**Event items are NOT selectable.** The `selectableIndices()` function already filters to `todoItem` only. Event items use a new `eventItem` kind and are rendered differently (no checkbox, different color, time range shown).

```go
// renderEvent writes a single calendar event line to the builder.
func (m Model) renderEvent(b *strings.Builder, e *caldav.CalendarEvent) {
    b.WriteString("  ")                        // no cursor (not selectable)
    b.WriteString("     ")                     // priority slot (empty for events)
    b.WriteString(m.styles.EventIcon.Render("[E]"))
    b.WriteString(" ")
    b.WriteString(m.styles.EventTitle.Render(e.Summary))
    b.WriteString(" ")
    b.WriteString(m.styles.EventTime.Render(e.TimeRange()))
    b.WriteString("\n")
}
```

### app.Model Changes

```go
type Model struct {
    // ... existing fields ...

    // CalDAV state
    caldavClient *caldav.Client  // nil if CalDAV disabled
    caldavCache  *caldav.Cache
    caldavError  string          // last sync error for display
}

// Modified New() constructor:
func New(provider *holidays.Provider, mondayStart bool, s store.TodoStore, t theme.Theme, cfg config.Config, caldavClient *caldav.Client, caldavCache *caldav.Cache) Model {
    // ... existing setup ...
    return Model{
        // ... existing fields ...
        caldavClient: caldavClient,
        caldavCache:  caldavCache,
    }
}

// Modified Init():
func (m Model) Init() tea.Cmd {
    if m.caldavClient != nil {
        return tea.Batch(
            syncCalDAVCmd(m.caldavClient, m.calendar.Year(), m.calendar.Month()),
            pollCalDAVCmd(m.cfg.CalDAV.PollMinutes),
        )
    }
    return nil
}
```

### Modified main.go

```go
func main() {
    cfg, err := config.Load()
    // ... existing setup ...

    var caldavClient *caldav.Client
    var caldavCache *caldav.Cache

    if cfg.CalDAV.Enabled {
        tokenPath, _ := config.CalDAVTokenPath()
        oauthCfg := caldav.OAuthConfig{
            ClientID:     cfg.CalDAV.ClientID,
            ClientSecret: cfg.CalDAV.ClientSecret,
            TokenPath:    tokenPath,
        }

        httpClient, err := caldav.NewOAuthHTTPClient(context.Background(), oauthCfg)
        if err != nil {
            // First run: need interactive auth
            fmt.Println("Google Calendar authorization required.")
            fmt.Printf("Visit: %s\n", caldav.AuthURL(oauthCfg))
            fmt.Print("Enter authorization code: ")
            // ... read code, exchange, save token ...
            httpClient, _ = caldav.NewOAuthHTTPClient(context.Background(), oauthCfg)
        }

        endpoint := "https://apidata.googleusercontent.com/caldav/v2/" + cfg.CalDAV.CalendarID
        caldavClient, _ = caldav.NewClient(httpClient, endpoint, endpoint+"/events")
        caldavCache = caldav.NewCache()
    }

    model := app.New(provider, cfg.MondayStart(), s, t, cfg, caldavClient, caldavCache)
    // ... rest of main ...
}
```

## Patterns to Follow

### Pattern 1: tea.Cmd for Async I/O (Existing Pattern)

**What:** All I/O operations return `tea.Cmd` functions that execute in goroutines.
**When:** Any CalDAV network call.
**Why:** This is already how `editor.Open()` works. Keeps the main event loop responsive. Bubble Tea manages goroutine lifecycle.

### Pattern 2: Constructor DI (Existing Pattern)

**What:** Pass dependencies into constructors, not global state.
**When:** Creating the CalDAV client, cache, passing to app.Model.
**Why:** The entire codebase uses this pattern. `app.New()` already takes store, provider, config, theme.

### Pattern 3: SetX Methods for Cross-Component Data (Existing Pattern)

**What:** Parent model calls setter methods on child models to push data down.
**When:** Distributing calendar events to calendar.Model and todolist.Model.
**Why:** This is how `calendar.RefreshIndicators()`, `todolist.SetViewMonth()`, `todolist.SetWeekFilter()` all work. The parent owns the data flow.

### Pattern 4: Pure Rendering Functions (Existing Pattern)

**What:** `RenderGrid()` takes all data as parameters and returns a string. No side effects.
**When:** Adding event indicators to calendar grid.
**Why:** `RenderGrid()` and `RenderWeekGrid()` already take indicators, totals, priorities as maps. Adding eventCounts follows the same pattern.

### Pattern 5: Graceful Degradation

**What:** CalDAV failures are non-fatal. The app works identically without CalDAV.
**When:** Network errors, OAuth token expiry, CalDAV disabled.
**Why:** This is a read-only enhancement. The core todo functionality must never be blocked by CalDAV issues.

## Anti-Patterns to Avoid

### Anti-Pattern 1: Storing CalDAV Events in SQLite

**What:** Persisting fetched events to the local database.
**Why bad:** Events are owned by Google, change externally, and must be re-fetched regularly. SQLite adds migration complexity, stale data problems, and blurs the boundary between local data (todos) and remote data (events).
**Instead:** In-memory cache, rebuilt on every sync. If the app restarts, events are re-fetched within seconds.

### Anti-Pattern 2: Adding CalDAV Methods to TodoStore Interface

**What:** Extending the existing `store.TodoStore` interface with event methods.
**Why bad:** Violates interface segregation. TodoStore is for mutable local todos. CalDAV events are read-only remote data. Every existing TodoStore consumer would see methods they don't need. SQLite tests would need CalDAV mocking.
**Instead:** Separate `caldav.Cache` type with its own query methods.

### Anti-Pattern 3: Raw Goroutines + Channels for Polling

**What:** Spawning a long-running goroutine with a channel to feed events into the Bubble Tea loop.
**Why bad:** Bubble Tea already provides `tea.Cmd` (for one-shot async) and `tea.Tick` (for periodic scheduling). Raw channels add synchronization complexity and bypass Bubble Tea's goroutine lifecycle management. The "realtime" example uses channels for continuous streams (like IRC messages), but CalDAV sync is periodic request-response.
**Instead:** `tea.Cmd` for each sync, `tea.Tick` for scheduling the next one. This matches the existing `editor.Open()` pattern.

### Anti-Pattern 4: Making CalDAV Events Selectable/Editable

**What:** Allowing cursor navigation, toggling, or editing of CalDAV events in the todo list.
**Why bad:** Events are read-only from Google Calendar. Editing would require CalDAV PUT operations, which dramatically increases complexity (conflict resolution, error handling, partial updates). Scope creep turns a read-only display into a full calendar client.
**Instead:** Events are display-only. The `[E]` prefix and distinct styling make it clear they are not actionable.

### Anti-Pattern 5: OAuth2 Flow Inside the TUI

**What:** Handling the OAuth2 browser flow while the Bubble Tea TUI is running.
**Why bad:** The TUI uses alt-screen mode. Opening a browser and prompting for a code requires terminal I/O that conflicts with Bubble Tea's input handling. The `editor.Open()` pattern works by suspending the TUI with `tea.ExecProcess`, but OAuth needs the user to paste text back, which doesn't fit that model.
**Instead:** Handle OAuth2 flow in `main.go` BEFORE starting the Bubble Tea program. If the token is missing or needs re-auth, do the interactive flow in plain terminal, then start the TUI.

## Scalability Considerations

| Concern | 1 Calendar | 5 Calendars | 20 Calendars |
|---------|------------|-------------|--------------|
| Fetch time | < 1s | 2-5s (sequential) | 10-20s or use parallel |
| Memory | ~10KB | ~50KB | ~200KB |
| Rendering | Negligible | Negligible | List scrolling needed |
| API quota | No issue | No issue | Google quota: 1M req/day, no issue |

For v1 of CalDAV integration, support a single calendar ("primary"). Multiple calendar support can be added later by iterating over a list of calendar IDs and merging events into the same cache.

## Files Changed Summary

### New Files

| File | Purpose |
|------|---------|
| `internal/caldav/client.go` | CalDAV client wrapper using go-webdav |
| `internal/caldav/event.go` | CalendarEvent domain type |
| `internal/caldav/cache.go` | Thread-safe in-memory event cache |
| `internal/caldav/oauth.go` | Google OAuth2 token management |
| `internal/caldav/messages.go` | Bubble Tea message types |

### Modified Files

| File | Change |
|------|--------|
| `internal/config/config.go` | Add `CalDAVConfig` struct to `Config` |
| `internal/config/paths.go` | Add `CalDAVTokenPath()` function |
| `internal/app/model.go` | Add caldav fields, sync commands, message handling, modified Init() |
| `internal/calendar/model.go` | Add `eventCounts map[int]int`, `SetCalendarEvents()` |
| `internal/calendar/grid.go` | Add `eventCounts` parameter to RenderGrid/RenderWeekGrid, render event indicators |
| `internal/todolist/model.go` | Add `calendarEvents []CalendarEvent`, `SetCalendarEvents()`, event section in visibleItems(), renderEvent() |
| `internal/todolist/styles.go` | Add event-related styles (EventIcon, EventTitle, EventTime) |
| `main.go` | OAuth bootstrap, CalDAV client creation, pass to app.New() |
| `go.mod` | Add go-webdav, go-ical, golang.org/x/oauth2 dependencies |

## Build Order (Considering Dependencies)

1. **Phase 1: CalDAV data layer** -- `caldav/event.go`, `caldav/cache.go`, `caldav/messages.go`
   - No external dependencies beyond stdlib. Defines types other code depends on.
   - Testable in isolation with unit tests.

2. **Phase 2: OAuth + Client** -- `caldav/oauth.go`, `caldav/client.go`, `config/config.go` changes, `config/paths.go` changes
   - Depends on go-webdav, go-ical, oauth2 packages (go.mod update).
   - Depends on Phase 1 types.
   - Testable with mock CalDAV server or integration test against Google.

3. **Phase 3: Bubble Tea integration** -- `app/model.go` changes, `main.go` changes
   - Depends on Phase 1 + 2 (client, cache, messages).
   - Wires CalDAV into the app lifecycle: Init, Update, sync commands.

4. **Phase 4: Calendar grid rendering** -- `calendar/model.go`, `calendar/grid.go` changes
   - Depends on Phase 1 (CalendarEvent type) and Phase 3 (data flow).
   - Adds event indicators to the grid.

5. **Phase 5: Todo list rendering** -- `todolist/model.go`, `todolist/styles.go` changes
   - Depends on Phase 1 (CalendarEvent type) and Phase 3 (data flow).
   - Adds Events section to the todo list.

**Phases 4 and 5 can be done in parallel** since they modify different packages and both depend only on Phase 3 being complete.

## Sources

- [emersion/go-webdav](https://github.com/emersion/go-webdav) -- CalDAV client library (v0.7.0, MIT, 455 stars) -- HIGH confidence
- [go-webdav/caldav pkg.go.dev](https://pkg.go.dev/github.com/emersion/go-webdav/caldav) -- CalDAV client API docs -- HIGH confidence
- [emersion/go-ical](https://github.com/emersion/go-ical) -- iCalendar parser (used by go-webdav) -- HIGH confidence
- [go-ical pkg.go.dev](https://pkg.go.dev/github.com/emersion/go-ical) -- go-ical API docs (Event.DateTimeStart, etc.) -- HIGH confidence
- [Google CalDAV API Guide](https://developers.google.com/workspace/calendar/caldav/v2/guide) -- Endpoint URLs, OAuth2 requirement -- HIGH confidence
- [Google CalDAV OAuth Scopes](https://developers.google.com/calendar/caldav/v2/auth) -- OAuth scope requirements -- HIGH confidence
- [Bubble Tea realtime example](https://github.com/charmbracelet/bubbletea/blob/main/examples/realtime/main.go) -- Channel pattern (not recommended for this use case, but informed decision) -- HIGH confidence
- [Commands in Bubble Tea](https://charm.land/blog/commands-in-bubbletea/) -- tea.Cmd, tea.Batch, tea.Tick patterns -- HIGH confidence
- [go-webdav Issue #68](https://github.com/emersion/go-webdav/issues/68) -- CalDAV event query example -- MEDIUM confidence
- [Google OAuth2 Go quickstart](https://developers.google.com/calendar/api/quickstart/go) -- Token management pattern -- HIGH confidence
- [CalDAV read-only scope caveat](https://support.google.com/calendar/thread/302141783/caldav-transition-to-oauth2-0) -- Read-only scope may cause 400 errors -- LOW confidence (community report, needs validation)
