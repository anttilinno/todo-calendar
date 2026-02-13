# Domain Pitfalls: Google Calendar CalDAV Integration

**Domain:** Adding read-only Google Calendar CalDAV integration to existing Go/Bubble Tea TUI todo-calendar app
**Researched:** 2026-02-13
**Confidence:** HIGH (pitfalls derived from codebase analysis of 8,695 LOC across 35+ Go files, official Google CalDAV documentation, go-webdav/go-ical library API analysis, Bubble Tea async patterns, and CalDAV protocol RFC 4791 investigation)

This document covers pitfalls specific to ADDING network features (CalDAV) to a previously offline-only TUI app. This is the app's first network feature. Each pitfall identifies what breaks, why, and how to prevent it.

---

## Critical Pitfalls

Mistakes that cause authentication failures, data corruption, app freezes, or require architecture redesigns.

### Pitfall 1: Google CalDAV Requires OAuth 2.0 -- App Passwords Do NOT Work

**What goes wrong:** The project plan assumes CalDAV with app password authentication (basic auth). Google disabled basic authentication for CalDAV entirely. As of mid-2024, the CalDAV endpoint at `apidata.googleusercontent.com` returns HTTP 401 Unauthorized for ANY request not authenticated with OAuth 2.0. App passwords, which still work for IMAP/SMTP, do NOT work for CalDAV. This is not a deprecation warning -- it is a hard block.

**Why it happens:** Google phased out basic auth for CalDAV in stages: new connections blocked in Summer 2024, existing connections terminated by Autumn 2024. The official CalDAV Developer's Guide states: "The CalDAV server will refuse to authenticate a request unless it arrives over HTTPS with OAuth 2.0 authentication of a Google account." The DAVx5 project (the most popular open-source CalDAV client for Android) confirms: "Since June 2023, Google has disabled all authentication methods except OAuth for their services."

**Consequences:** If the implementation assumes basic auth with app passwords:
- Every request to `apidata.googleusercontent.com/caldav/v2/` returns 401
- Zero calendar data is ever fetched
- The entire CalDAV integration is dead on arrival
- Users who generate app passwords and configure them get cryptic auth errors

**Prevention:**

The authentication strategy must be OAuth 2.0 from day one. For a CLI/TUI app, there are two viable flows:

1. **Device Authorization Flow** (recommended for TUI): The app displays a URL and a short code. The user opens the URL in their browser, enters the code, and authorizes. The app polls Google's token endpoint until authorization is granted. No localhost server needed. Ideal for headless/SSH environments.

2. **Localhost Redirect Flow**: The app starts a temporary HTTP server on localhost, opens the browser with the auth URL, and receives the callback. Works well on desktop but fails in SSH sessions.

Implementation with Go:
```go
import "golang.org/x/oauth2"
import "golang.org/x/oauth2/google"

// Use readonly scope for read-only CalDAV access
conf := &oauth2.Config{
    ClientID:     "...",
    ClientSecret: "...",
    Scopes:       []string{"https://www.googleapis.com/auth/calendar.readonly"},
    Endpoint:     google.Endpoint,
}

// The oauth2.Config produces an http.Client with automatic token refresh
httpClient := conf.Client(ctx, token)

// Pass to go-webdav's caldav.NewClient
caldavClient, err := caldav.NewClient(webdav.HTTPClientWithBasicAuth(httpClient, "", ""), endpoint)
// NOTE: go-webdav's NewClient takes an HTTPClient interface -- you provide
// the OAuth2-authenticated http.Client, NOT basic auth credentials
```

**Requirements this creates:**
- Google Cloud Console project with OAuth consent screen
- OAuth client ID (type: "TVs and Limited Input devices" for device flow, or "Desktop" for localhost redirect)
- Token storage (access token + refresh token) persisted locally
- Token refresh logic (access tokens expire in ~1 hour)
- First-run authorization UX in the TUI

**Warning signs:**
- Any code using `webdav.HTTPClientWithBasicAuth` with a username/password
- Configuration asking for "app password" or "Google password"
- No OAuth token storage or refresh logic
- No Google Cloud Console project setup instructions

**Detection:** Attempt to connect to `https://apidata.googleusercontent.com/caldav/v2/{email}/user` with basic auth credentials. You will get HTTP 401 immediately.

**Phase to address:** This is the FIRST thing to solve. Authentication must work before any CalDAV fetching logic can be developed or tested. The entire milestone architecture depends on this decision.

---

### Pitfall 2: Blocking Network Calls in Bubble Tea Update Loop Freeze the TUI

**What goes wrong:** The app is currently 100% offline. Every operation in `Update()` and `View()` is synchronous and fast (SQLite reads, in-memory data). CalDAV network calls take 500ms-5s (DNS + TLS + HTTP request + response parsing). If a CalDAV fetch runs synchronously in `Update()`, the entire TUI freezes -- no keyboard input, no screen redraw, no quit handling -- until the network call completes or times out.

**Why it happens:** Bubble Tea uses a single-threaded event loop. `Update()` must return quickly (< 16ms for 60fps rendering). The existing codebase has zero async operations -- `Init()` returns nil, all store operations are synchronous SQLite calls that complete in microseconds. A developer unfamiliar with Bubble Tea's async model might write:

```go
// WRONG: blocks the entire TUI
case tickMsg:
    events, err := caldavClient.QueryCalendar(ctx, calendarPath, &query)
    m.events = events
    return m, scheduleNextTick()
```

**Consequences:**
- TUI appears hung for 1-5 seconds during every poll
- User presses 'q' during a fetch -- nothing happens until fetch completes
- If DNS is slow or server is down, TUI freezes for the full timeout (default: no timeout in Go's http.Client)
- Multiple rapid ticks can queue up, causing cascading freezes

**Where it breaks in the codebase:**
- `app/model.go:118-326` -- `Update()` currently has zero async operations; every case returns immediately
- `app/model.go:113-115` -- `Init()` returns nil; no initial async commands exist
- `calendar/model.go:85-146` -- `Update()` directly calls `m.store.IncompleteTodosPerDay()` synchronously; a developer might follow this pattern for CalDAV

**Prevention:**

ALL network operations must use Bubble Tea's `tea.Cmd` pattern:

```go
// Define message types for CalDAV results
type caldavFetchMsg struct {
    events []CalendarEvent
    err    error
}

// The Cmd runs in a goroutine managed by Bubble Tea
func fetchCalendarEvents(client *caldav.Client, calPath string, start, end time.Time) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        query := &caldav.CalendarQuery{
            CompFilter: caldav.CompFilter{
                Name:  "VCALENDAR",
                Start: start,
                End:   end,
                Comps: []caldav.CompFilter{{Name: "VEVENT"}},
            },
        }
        objects, err := client.QueryCalendar(ctx, calPath, query)
        if err != nil {
            return caldavFetchMsg{err: err}
        }
        return caldavFetchMsg{events: parseEvents(objects)}
    }
}

// In Update(), handle the result message:
case caldavFetchMsg:
    if msg.err != nil {
        m.caldavError = msg.err  // show error in status bar
    } else {
        m.calendarEvents = msg.events
    }
    return m, scheduleNextPoll()  // tea.Tick for next poll
```

For background polling at 5-minute intervals, use `tea.Tick`:
```go
func scheduleNextPoll() tea.Cmd {
    return tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
        return pollTickMsg(t)
    })
}
```

**Critical detail:** The `tea.Cmd` function must NOT access or modify the model. It receives only the data it needs via closure and returns results as a message. This is already how the app handles external editor launches (`tea.ExecProcess` in `app/model.go:444`), so the pattern has precedent in this codebase.

**Warning signs:**
- CalDAV client methods called directly inside `Update()` or `View()`
- No `context.WithTimeout` on network calls
- No `tea.Cmd` wrapper around network operations
- Network fetch and result handling in the same function

**Detection:** Start the app with network disabled (airplane mode). If the TUI takes > 1 second to show the initial screen, a blocking network call is in the initialization path.

**Phase to address:** Must be the foundational pattern established in the first networking phase. Every subsequent CalDAV operation must follow this pattern.

---

### Pitfall 3: OAuth Token Storage in Plaintext Config File Leaks Credentials

**What goes wrong:** The existing config uses plaintext TOML (`~/.config/todo-calendar/config.toml`). A developer might store OAuth tokens (access_token, refresh_token) in the same file or a sibling JSON file with 0644 permissions. OAuth refresh tokens are long-lived credentials that grant indefinite access to the user's Google Calendar. A plaintext refresh token is equivalent to a stored password.

**Why it happens:** The existing `config.Save()` in `config/config.go:109-147` writes TOML with `os.CreateTemp` and `os.Rename`. Following this pattern for token storage would create `~/.config/todo-calendar/token.json` with default permissions. The Google Calendar Go quickstart literally does this:

```go
// Google's example -- insecure for real apps
f, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
json.NewEncoder(f).Encode(token)
```

Even with 0600 permissions, the token is readable by any process running as the same user, any backup tool, any file sync service, and shows up in `grep -r` searches.

**Consequences:**
- Refresh token leaked via backup, git, or file sync gives attacker permanent calendar read access
- Token stolen from a shared machine grants calendar access without the user's knowledge
- No way to detect that a token has been compromised (Google does not notify on token use)

**Prevention:**

For a personal TUI tool, there are pragmatic security levels:

1. **Minimum viable (acceptable for personal use):** Store tokens in a separate file with 0600 permissions in the XDG config directory. Use `os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)`. Add the token file to `.gitignore`. This is what Google's own quickstart recommends.

2. **Better -- system keyring:** Use `zalando/go-keyring` to store the refresh token in the OS keyring (macOS Keychain, Linux Secret Service via D-Bus, Windows Credential Manager). BUT: on Linux, this requires D-Bus and `gnome-keyring-daemon` running, which fails on headless servers and minimal distros. The fallback must be graceful.

3. **Recommended approach for this app:** File-based storage with 0600 permissions, stored at `~/.config/todo-calendar/oauth-token.json`. Keep it separate from `config.toml`. Add to any `.gitignore` patterns. Document that the file contains sensitive credentials.

**Keyring pitfall on Linux:** `zalando/go-keyring` depends on the Secret Service D-Bus interface. On headless Linux, WSL, Docker, or minimal distros without GNOME Keyring, it fails with "The name org.freedesktop.secrets was not provided by any .service files." This would make the entire CalDAV feature unusable on those environments. For a personal TUI tool, the system keyring is overkill and creates more problems than it solves.

**Warning signs:**
- Tokens stored in `config.toml` alongside non-sensitive settings
- Token file created with default permissions (0644 or 0666)
- No `.gitignore` entry for token files
- Token file path hardcoded without using XDG conventions
- System keyring used without headless/minimal-distro fallback

**Detection:** After completing OAuth setup, run `ls -la ~/.config/todo-calendar/` and verify the token file has `-rw-------` permissions. Run `cat ~/.config/todo-calendar/config.toml` and verify no tokens appear there.

**Phase to address:** The OAuth/authentication phase. Token storage design must happen alongside the OAuth flow implementation.

---

### Pitfall 4: Missing HTTP Timeout Causes Indefinite TUI Hangs on Network Failure

**What goes wrong:** Go's default `http.Client` has no timeout. If Google's CalDAV server is unreachable, the TCP connection attempt blocks until the OS TCP timeout (often 2+ minutes on Linux). Even with Bubble Tea's `tea.Cmd` pattern running the request in a goroutine, the goroutine leaks and accumulates -- each 5-minute poll creates a new goroutine that blocks for 2 minutes, creating a goroutine leak.

**Why it happens:** The app has never made network calls. There is no HTTP client configuration anywhere in the codebase. A developer creates `&http.Client{}` or uses the OAuth2 library's default client without setting timeouts.

**Consequences:**
- On flaky WiFi: goroutines pile up, each waiting for TCP timeout
- On DNS failure: 30-second DNS timeout per request
- On server outage: 2+ minute TCP timeout per request
- Memory grows with each leaked goroutine holding response buffers
- After 10 failed polls (50 minutes), 10+ goroutines are blocking simultaneously
- App becomes sluggish as the runtime manages many blocked goroutines

**Prevention:**

```go
// Always create the HTTP client with explicit timeouts
httpClient := &http.Client{
    Timeout: 30 * time.Second,  // Total request timeout
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,  // TCP connection timeout
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 15 * time.Second,
        IdleConnTimeout:       90 * time.Second,
    },
}

// When wrapping with OAuth2:
ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
oauthClient := conf.Client(ctx, token)
```

Additionally, EVERY `tea.Cmd` that makes a network call must use `context.WithTimeout`:
```go
func fetchEvents(client *caldav.Client, ...) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        // ... use ctx for all operations
    }
}
```

**Warning signs:**
- `&http.Client{}` with no Timeout field
- `context.Background()` passed directly to network calls without timeout wrapping
- No Transport configuration
- No logging/counting of failed requests

**Detection:** Start the app, then disconnect the network. The app should show a fetch error within 30 seconds, not hang.

**Phase to address:** HTTP client setup in the networking foundation phase. The timeout-configured client must be created once and reused.

---

### Pitfall 5: All-Day Event Timezone Off-by-One Puts Events on Wrong Calendar Day

**What goes wrong:** Google Calendar stores all-day events using `VALUE=DATE` (e.g., `DTSTART;VALUE=DATE:20260215` for February 15). This is a date without time or timezone. But when this app converts it to display alongside dated todos (which use `"2006-01-02"` string format), timezone misinterpretation can shift the event forward or backward by one day.

**Why it happens:** The iCalendar spec (RFC 5545) says `VALUE=DATE` is a "floating date" -- it means that calendar day in whatever timezone the user is in. But if Go code parses it as UTC midnight (`2026-02-15T00:00:00Z`) and then converts to local time, a user in UTC-5 sees it as `2026-02-14T19:00:00-05:00` -- February 14, not February 15. The event appears on the wrong day.

This is the single most common CalDAV implementation bug. Real-world examples:
- Home Assistant issue #126448: "Recurring all day events are displayed 1 day too early"
- Home Assistant issue #25814: "CalDAV all day events are UTC only"
- GLPI issue #20414: "All-day events incorrectly shifted by timezone offset"

**Where it breaks in the codebase:**
- `store/todo.go:9` -- `dateFormat = "2006-01-02"` is the canonical date format; CalDAV events must produce this exact format
- `store/todo.go:71-88` -- `InMonth()` and `InDateRange()` parse with this format
- `calendar/model.go:28-35` -- `weekStartFor()` uses `time.Local` for date construction

**Prevention:**

When parsing iCalendar events, detect `VALUE=DATE` properties and handle them differently from `VALUE=DATE-TIME`:

```go
func extractEventDate(event *ical.Event) (string, error) {
    dtstart := event.Props.Get(ical.PropDateTimeStart)
    if dtstart == nil {
        return "", errors.New("no DTSTART")
    }

    // Check if this is a date-only value (all-day event)
    if dtstart.Params.Get(ical.ParamValue) == "DATE" ||
       !strings.Contains(dtstart.Value, "T") {
        // All-day event: the date string IS the date, no timezone conversion
        // "20260215" -> "2026-02-15"
        t, err := time.Parse("20060102", dtstart.Value)
        if err != nil {
            return "", err
        }
        return t.Format("2006-01-02"), nil
    }

    // Timed event: parse with timezone, then extract local date
    t, err := dtstart.DateTime(nil)  // go-ical handles TZID
    if err != nil {
        return "", err
    }
    return t.In(time.Local).Format("2006-01-02"), nil
}
```

The key insight: for `VALUE=DATE`, parse the date string directly WITHOUT going through `time.Time` timezone conversion. The 8-digit string `20260215` is reformatted to `2026-02-15` as a pure string operation, bypassing timezone entirely.

**Warning signs:**
- All dates parsed through `time.Parse` with a time-containing layout
- `time.UTC` used for all-day event dates
- No distinction between `VALUE=DATE` and `VALUE=DATE-TIME` in parsing code
- All-day events showing up one day early or late

**Detection:** Create an all-day event in Google Calendar for a specific date. Fetch via CalDAV and verify the displayed date matches. Test in a timezone with significant UTC offset (e.g., UTC+12 or UTC-12).

**Phase to address:** The event parsing phase. This must be correct from the start because the date is the primary display attribute.

---

## Moderate Pitfalls

These cause UX confusion, incorrect displays, or require rework but are recoverable.

### Pitfall 6: Stale Calendar Data Displayed as Current After Network Outage

**What goes wrong:** The app polls every 5 minutes. If the network is down for an hour, the last successful fetch from an hour ago is still displayed. The user sees events that may have been cancelled or moved, with no indication that the data is stale.

**Why it happens:** The offline-first architecture (SQLite todos are always current) creates an expectation that everything on screen is live. CalDAV events break this assumption because they can be arbitrarily stale. Unlike todos (which the user controls), calendar events are controlled by others and change without local notification.

**Consequences:**
- User sees a meeting at 3 PM, but it was moved to 4 PM 30 minutes ago
- User sees a cancelled event and shows up
- No visual distinction between "fetched 2 minutes ago" and "fetched 2 hours ago"

**Prevention:**

1. **Show a sync status indicator** somewhere in the TUI (e.g., in the help bar or calendar pane header):
   ```
   Calendar: synced 3m ago    -- normal, green/dim
   Calendar: synced 47m ago   -- warning, yellow
   Calendar: sync failed      -- error, red
   Calendar: offline          -- not configured or auth failed
   ```

2. **Track last successful sync time** in the model:
   ```go
   type CalDAVState struct {
       LastSync    time.Time
       LastError   error
       Events      []CalendarEvent
       Syncing     bool
   }
   ```

3. **Dim or annotate stale events** if the last sync is > 15 minutes old. This gives the user visual feedback that the data may not be current.

4. **Provide a manual refresh key** (e.g., `r` or `R`) that triggers an immediate CalDAV fetch outside the polling schedule. This gives the user control.

**Warning signs:**
- No sync timestamp stored or displayed
- No visual difference between freshly synced and hours-old data
- No manual refresh capability
- No error display when sync fails

**Phase to address:** The CalDAV sync/polling phase, after basic fetching works.

---

### Pitfall 7: Token Refresh Failure During Background Poll Silently Breaks All Future Fetches

**What goes wrong:** OAuth2 access tokens expire after ~1 hour. Go's `oauth2` library automatically refreshes them using the refresh token. But if the refresh fails (network error during refresh, refresh token revoked by user in Google security settings, Google project verification expired), every subsequent CalDAV request fails with 401. The background poll silently fails every 5 minutes without ever recovering or alerting the user.

**Why it happens:** The `oauth2.Config.Client()` returns an `http.Client` with a `TokenSource` that auto-refreshes. If refresh fails, it returns an error on the HTTP request, not a distinct token error. The CalDAV response looks like an auth error, but the actual problem is the token refresh, not the CalDAV request itself.

**Consequences:**
- Calendar events disappear from the display (last successful data ages out)
- Error message says "401 Unauthorized" which the user interprets as wrong credentials
- User regenerates app password (futile -- OAuth is in use) or reconfigures
- Actual fix is re-authorizing, but nothing tells the user this

**Prevention:**

1. **Detect token refresh failures specifically:**
   ```go
   type tokenErrorMsg struct {
       err error
       needsReauth bool
   }

   // Custom TokenSource that detects refresh failure
   type persistingTokenSource struct {
       base    oauth2.TokenSource
       storage TokenStorage
   }

   func (s *persistingTokenSource) Token() (*oauth2.Token, error) {
       tok, err := s.base.Token()
       if err != nil {
           // Check if this is a refresh failure
           return nil, fmt.Errorf("token refresh failed (re-authorize with Ctrl+G): %w", err)
       }
       // Persist new token after successful refresh
       s.storage.Save(tok)
       return tok, nil
   }
   ```

2. **Distinguish between transient network errors and auth errors** in the error display:
   - Network timeout: "Calendar sync failed (network timeout, retrying in 5m)"
   - 401 after token refresh: "Calendar authentication expired -- press Ctrl+G to re-authorize"
   - 403: "Calendar access denied -- check permissions"

3. **Auto-persist refreshed tokens.** The default `oauth2.TokenSource` refreshes in-memory. If the app crashes after a refresh but before the new token is saved, the old (now-invalid) refresh token is loaded on next start, causing immediate auth failure.

**Warning signs:**
- Using `oauth2.Config.Client()` directly without custom TokenSource
- No token persistence after refresh
- Generic "sync failed" error for all failure types
- No re-authorization UX path

**Phase to address:** The OAuth/authentication phase. Token refresh handling is part of the auth implementation, not an afterthought.

---

### Pitfall 8: Google CalDAV Calendar Discovery Requires Correct Principal URL Construction

**What goes wrong:** The CalDAV protocol uses a discovery sequence: find principal URL -> find calendar home set -> list calendars. Google's CalDAV API has specific URL patterns that differ from standard CalDAV servers. If the discovery code uses `caldav.DiscoverContextURL()` (DNS-based SRV record lookup), it will not find Google's CalDAV endpoint because Google does not publish SRV records for CalDAV.

**Why it happens:** Standard CalDAV discovery (RFC 6764) uses DNS SRV records at `_caldavs._tcp.gmail.com`. Google does not implement this. Instead, Google provides a hardcoded base URL: `https://apidata.googleusercontent.com/caldav/v2/`. The calendar home set for a user is `https://apidata.googleusercontent.com/caldav/v2/{calendarId}/`, where `calendarId` is the user's email for the primary calendar.

The go-webdav library's `DiscoverContextURL()` performs DNS-based discovery, which returns nothing for Google. The developer then assumes the library does not work with Google.

**Prevention:**

Skip automatic discovery for Google Calendar. Use direct URL construction:

```go
const googleCalDAVBase = "https://apidata.googleusercontent.com/caldav/v2/"

func googleCalendarURL(email string) string {
    // Primary calendar: email is the calendar ID
    return googleCalDAVBase + url.PathEscape(email) + "/events"
}

func googlePrincipalURL(email string) string {
    return googleCalDAVBase + url.PathEscape(email) + "/user"
}

// To list all calendars for the user:
// 1. Create client with principal URL
// 2. Call FindCalendarHomeSet with principal URL
// 3. Call FindCalendars with the home set
client, _ := caldav.NewClient(httpClient, googlePrincipalURL(email))
homeSet, _ := client.FindCalendarHomeSet(ctx, googlePrincipalURL(email))
calendars, _ := client.FindCalendars(ctx, homeSet)
```

**Google-specific URL quirks:**
- Calendar IDs with special characters must be URL-encoded
- Shared calendars have long hex IDs, not email addresses
- The principal URL uses `/user` suffix, the events URL uses `/events` suffix
- `FindCalendars` returns ALL calendars the user has access to (including subscribed calendars, holidays, birthdays)

**Warning signs:**
- Using `DiscoverContextURL` for Google Calendar
- Hardcoding only the primary calendar (missing shared/subscribed calendars)
- Not URL-encoding calendar IDs
- Confusing the `/user` and `/events` URL patterns

**Phase to address:** The CalDAV client setup phase, immediately after authentication works.

---

### Pitfall 9: Timed Events From Different Timezones Display at Wrong Local Time

**What goes wrong:** A calendar event created in timezone "America/New_York" at 3:00 PM has DTSTART `20260215T150000` with TZID `America/New_York`. If the user viewing it is in `Europe/Helsinki` (UTC+2), the event should display as 10:00 PM local time. But if the parser ignores TZID and treats the time as local, it shows 3:00 PM -- 7 hours wrong.

**Why it happens:** iCalendar events represent times in three ways:
1. **UTC (Zulu):** `DTSTART:20260215T200000Z` -- the trailing Z means UTC
2. **With TZID:** `DTSTART;TZID=America/New_York:20260215T150000` -- time in specified zone
3. **Floating:** `DTSTART:20260215T150000` -- no timezone, treated as local

Google Calendar typically uses TZID for timed events and UTC for internally generated events. The go-ical library's `DateTime()` method handles TZID resolution, but only if the IANA timezone database is available.

**Additional complication for this app:** The existing todo system uses date-only strings (`"2006-01-02"`) with no time component. CalDAV events have both date AND time. For display purposes, the event's date must be extracted in the user's local timezone:

```
Event: 11:00 PM Feb 15 in America/New_York = 6:00 AM Feb 16 in Europe/Helsinki
```

The DATE portion changes depending on timezone. If you extract the date in the wrong timezone, the event shows up on the wrong calendar day.

**Prevention:**

1. **Always convert to local time before extracting the date:**
   ```go
   // CORRECT: convert to local time, then extract date
   localTime := eventTime.In(time.Local)
   dateStr := localTime.Format("2006-01-02")

   // WRONG: extract date from original timezone
   dateStr := eventTime.Format("2006-01-02")  // wrong day if event crosses midnight locally
   ```

2. **Embed the IANA timezone database** in the binary to ensure `time.LoadLocation()` works everywhere:
   ```go
   import _ "time/tzdata"  // Embeds IANA tzdb in binary
   ```
   Without this import, `time.LoadLocation("America/New_York")` fails on minimal Linux systems that lack `/usr/share/zoneinfo/`. The go-ical library calls `time.LoadLocation()` internally when resolving TZID values.

3. **For display in the todo list,** store the event's local date and optionally the time:
   ```go
   type CalendarEvent struct {
       Summary   string
       Date      string    // "2006-01-02" in user's local timezone
       StartTime string    // "15:04" in user's local timezone, empty for all-day
       EndTime   string    // "15:04" in user's local timezone, empty for all-day
       AllDay    bool
       CalendarName string
   }
   ```

**Warning signs:**
- No `import _ "time/tzdata"` in the binary
- Using `time.UTC` instead of `time.Local` for date extraction
- Ignoring TZID parameters in VEVENT parsing
- Not testing with events in different timezones

**Detection:** Create events in Google Calendar in different timezones. One event at 11 PM Eastern, another at 1 AM Tokyo time (same UTC instant, different dates in some timezones). Verify they appear on the correct local date.

**Phase to address:** The event parsing phase, alongside Pitfall 5 (all-day events).

---

### Pitfall 10: Recurring Events Require RRULE Expansion That Google May or May Not Do

**What goes wrong:** A weekly meeting event has a single VEVENT with `RRULE:FREQ=WEEKLY;BYDAY=MO`. This represents infinite Mondays. When you query with a time-range filter, the server should either expand recurrences (returning individual VEVENTs for each occurrence) or return the RRULE for client-side expansion. Google's CalDAV supports the `EXPAND` property in calendar-query requests, but the behavior and reliability varies.

**Why it happens:** CalDAV RFC 4791 defines the `calendar-data` element with an optional `expand` element that asks the server to expand recurring events within the requested time range. Google supports this. However:
- If `EXPAND` is not requested, Google returns the master VEVENT with the RRULE, and the client must expand it
- RRULE expansion is complex: `RRULE` + `RDATE` - `EXDATE` + exception VEVENTs with `RECURRENCE-ID`
- Some providers (documented: Zoho) silently fail when EXPAND is requested, returning 200 with empty results

Client-side RRULE expansion requires:
- Parsing RRULE syntax (FREQ, INTERVAL, BYDAY, BYMONTH, COUNT, UNTIL, etc.)
- Handling EXDATE (deleted occurrences)
- Handling exception VEVENTs (modified single occurrences)
- Timezone-aware recurrence generation (DST shifts change occurrence times)

**Prevention:**

Use server-side expansion via the EXPAND property in CalendarQuery:

```go
query := &caldav.CalendarQuery{
    CompRequest: caldav.CalendarCompRequest{
        Name: "VCALENDAR",
        Comps: []caldav.CalendarCompRequest{{
            Name:  "VEVENT",
            Props: []string{"SUMMARY", "DTSTART", "DTEND", "DURATION", "DESCRIPTION"},
        }},
        Expand: &caldav.CalendarExpandRequest{
            Start: startTime,
            End:   endTime,
        },
    },
    CompFilter: caldav.CompFilter{
        Name:  "VCALENDAR",
        Comps: []caldav.CompFilter{{
            Name:  "VEVENT",
            Start: startTime,
            End:   endTime,
        }},
    },
}
```

With EXPAND, Google returns individual expanded VEVENTs, each with concrete DTSTART/DTEND. No client-side RRULE parsing needed.

**BUT: verify EXPAND works with Google CalDAV.** If Google returns the master event instead of expanded instances, you need a fallback. The `emersion/go-ical` library does not include RRULE expansion. You would need a separate library or implement it.

For the initial read-only implementation, a pragmatic approach:
1. Request EXPAND in the query
2. If the response contains RRULE properties (meaning expansion did not happen), skip those events and log a warning
3. Add RRULE expansion as a future enhancement if needed

**Warning signs:**
- No EXPAND in the CalendarQuery
- Client-side RRULE expansion attempted as the primary strategy (massive complexity)
- Recurring events showing only the first occurrence
- Modified occurrences (single instance moved to different time) not handled

**Phase to address:** The event fetching phase. Test EXPAND behavior with Google CalDAV early.

---

### Pitfall 11: Sync-Collection Report and CTag Ignored, Causing Full Re-fetch Every Poll

**What goes wrong:** The naive approach fetches ALL events in the visible time range on every 5-minute poll. For a user with a busy calendar (100+ events/month), this transfers 50-200KB of iCalendar data every 5 minutes. This wastes bandwidth, battery (on laptops), and creates unnecessary server load that may trigger Google's rate limits.

**Why it happens:** The developer implements the simplest working approach (full fetch with time-range filter) and moves on. Google's CalDAV supports efficient change detection via CTag and sync-token (RFC 6578), but implementing incremental sync is significantly more complex than a full fetch.

**Prevention:**

Implement a two-level caching strategy:

1. **CTag check before full fetch:** Before querying events, check if the calendar's CTag has changed:
   ```go
   // PROPFIND to get current CTag
   calendar, _ := client.FindCalendars(ctx, homeSet)
   // Compare calendar.CTag with stored CTag
   if calendar.CTag == m.storedCTag {
       return  // Nothing changed, skip fetch
   }
   ```
   CTag is a single value that changes whenever any event in the calendar changes. This turns most polls into a single lightweight PROPFIND request.

2. **Sync-token for incremental updates (advanced, defer):** After the first full fetch, store the sync-token from the response. Subsequent requests send the sync-token and receive only changed/deleted events. Google supports this via RFC 6578. BUT: sync-tokens can be invalidated (expired, server-side change), requiring a fallback to full fetch.

**For the initial implementation:** CTag check is sufficient. It reduces 95% of unnecessary full fetches with minimal code complexity. Sync-token can be added later for optimization.

**Warning signs:**
- Full CalendarQuery on every poll regardless of changes
- No CTag or ETag storage between polls
- No consideration of rate limits
- Polling interval shorter than necessary (< 5 minutes)

**Phase to address:** Optimization phase, after basic polling works. CTag check should be added early; sync-token can be deferred.

---

## Minor Pitfalls

### Pitfall 12: CalDAV Events Mixed With Todos Create Confusing Sort Order

**What goes wrong:** The todo list currently shows todos sorted by `sort_order, date, id`. Calendar events injected into the same list need a sensible position. If events are appended at the end of the day's todos, users see todos first and events last, which may not match their mental model (events at the top of the day, todos below). If events are interleaved by time, the sort order logic becomes complex.

**Prevention:** Display calendar events in a separate section within the todo list pane, visually distinct from todos. Use a section header like "Calendar" with a different style. Events within the section are sorted by start time. This avoids contaminating the todo sort order and makes it clear which items are local todos vs external calendar events.

**Phase to address:** The display integration phase.

---

### Pitfall 13: Google CalDAV Does Not Support VTODO -- Cannot Sync Tasks

**What goes wrong:** A developer might plan to eventually sync Google Tasks alongside calendar events via CalDAV. Google's CalDAV implementation explicitly does not support VTODO or VJOURNAL components. Only VEVENT is supported. This is documented in Google's CalDAV Developer's Guide but easy to miss.

**Prevention:** Document this limitation in the architecture. If Google Tasks integration is desired in the future, it requires the Google Tasks REST API (`tasks/v1`), which is a completely separate API from CalDAV. Do not plan CalDAV-based task sync with Google.

**Phase to address:** Architecture/planning phase. Document the limitation and do not promise task sync.

---

### Pitfall 14: The go-webdav Library May Return Opaque Errors From Google's Non-Standard Responses

**What goes wrong:** Google's CalDAV implementation has quirks that deviate from strict RFC compliance. The go-webdav library is primarily tested against standard CalDAV servers (Radicale, Nextcloud, etc.). Google-specific error responses may not parse correctly, returning generic "bad request" or XML parsing errors instead of actionable error messages.

**Known Google quirks:**
- Google may set cookies on client applications for "security and abuse prevention"
- Google does not support `MKCALENDAR`, `LOCK`, `UNLOCK`, `COPY`, `MOVE`, or `MKCOL`
- Google's scheduling model differs from CalDAV (invitations delivered directly to events collection, no inbox)
- Google may return non-standard error XML for rate limiting

**Prevention:** Wrap all go-webdav client calls with error handling that:
1. Logs the full HTTP status code and response body for debugging
2. Translates common HTTP errors (401, 403, 404, 429, 503) into user-friendly messages
3. Handles Google's cookie setting gracefully (the `http.Client` does this by default, but be aware)

**Phase to address:** The CalDAV client wrapper phase.

---

### Pitfall 15: First-Run OAuth Authorization UX Is Jarring in a TUI

**What goes wrong:** The app currently starts instantly with `Init()` returning nil. Adding CalDAV with OAuth means the first launch after configuration requires an authorization flow: display a URL, wait for the user to authorize in browser, receive the token. This interrupts the instant-launch experience that TUI users expect.

**Prevention:**

1. **Make CalDAV opt-in, not default.** The app starts without CalDAV. The user enables it via Settings (existing overlay at `S` key) by entering their Google email. Only then does the OAuth flow trigger.

2. **Show the auth flow as a modal/overlay,** not a blocking screen. The user can cancel and use the app without CalDAV.

3. **After authorization, persist the token.** Subsequent launches should be silent (auto-refresh handles token expiry).

4. **If the auth flow fails or is cancelled,** the app works normally without calendar events. CalDAV is an enhancement, not a requirement.

**Warning signs:**
- App blocks on startup waiting for OAuth
- No way to cancel the auth flow
- No way to disable CalDAV after enabling it
- Auth failure prevents app from launching

**Phase to address:** The OAuth UX phase, designed alongside the Settings integration.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation | Severity |
|-------------|---------------|------------|----------|
| Authentication | App passwords do NOT work with Google CalDAV | Must use OAuth 2.0 from day one | Critical |
| Authentication | Token refresh failure silently breaks all fetches | Custom TokenSource with persistence + error classification | Moderate |
| Authentication | First-run OAuth UX is jarring | Opt-in via Settings, modal overlay, cancellable | Minor |
| Token storage | Plaintext refresh tokens in config dir | Separate file, 0600 permissions, not in config.toml | Critical |
| HTTP client | No timeout = indefinite hangs | Explicit timeouts on Client, Transport, and context | Critical |
| CalDAV discovery | DNS-based discovery fails for Google | Hardcode Google endpoint URLs, skip DiscoverContextURL | Moderate |
| Event parsing | All-day events off by one day | Parse VALUE=DATE as string, no timezone conversion | Critical |
| Event parsing | Timed events in wrong timezone | Convert to local time before date extraction, embed tzdata | Moderate |
| Event parsing | Recurring events not expanded | Use EXPAND in CalendarQuery, verify Google support | Moderate |
| Async pattern | Blocking network call in Update() | All network ops via tea.Cmd, never in Update/View | Critical |
| Sync/polling | Full re-fetch every 5 minutes | CTag check before fetch, incremental sync later | Minor |
| Error UX | No indication of stale data or sync failure | Status indicator, last-sync timestamp, manual refresh | Moderate |
| Display | Events mixed with todos confuse sort order | Separate "Calendar" section in todo list | Minor |
| Google limits | No VTODO support in Google CalDAV | Document limitation, do not plan CalDAV task sync | Minor |
| Library compat | go-webdav may not handle Google quirks | Error wrapping, logging, user-friendly messages | Minor |

---

## Integration Risks with Existing System

### Risk 1: CalDAV Events Must Coexist With Todo Date System Without Corruption

The existing `store.Todo` struct uses date-only strings (`"2006-01-02"`) with date precision ("day", "month", "year"). CalDAV events have date+time with timezone. These must remain separate data types. CalDAV events should NOT be stored in the todos SQLite table or represented as `store.Todo` objects. A separate in-memory struct (`CalendarEvent`) avoids schema contamination.

**Mitigation:** Create a `calendarEvent` package separate from `store`. Events are never persisted locally (they come from Google on every sync). The display layer merges todos and events for rendering but they remain distinct types.

### Risk 2: Background Polling Must Not Interfere With User Input

The existing `Update()` loop routes messages based on active pane and overlay state (`app/model.go:118-326`). CalDAV poll results arrive as messages that must be handled regardless of which pane is active or which overlay is open. If CalDAV messages are routed through the pane-specific update functions, they get swallowed when an overlay is open.

**Mitigation:** Handle CalDAV messages at the TOP of `Update()`, before any overlay/pane routing, similar to how `settings.SettingChangedMsg` is handled at lines 123-138. CalDAV messages are model-level concerns, not pane-level.

### Risk 3: OAuth Dependency Adds Significant New Dependencies

Adding `golang.org/x/oauth2` and potentially `google.golang.org/api` brings in a substantial dependency tree. The current app has a lean `go.mod` with 12 direct dependencies. OAuth2 adds Google's auth libraries, gRPC dependencies (if using Google API client), and potentially `cloud.google.com/go` packages.

**Mitigation:** Use `golang.org/x/oauth2` directly (minimal dependency) rather than the full Google API client library. For CalDAV, `emersion/go-webdav` + `golang.org/x/oauth2` is sufficient. Do not import `google.golang.org/api/calendar/v3` (the REST API client) -- CalDAV replaces it.

### Risk 4: The Calendar Pane Needs to Show Events But Currently Only Shows Indicators

The calendar grid (`calendar/grid.go`) currently shows day-level indicators: brackets for pending todos, check style for all done. CalDAV events need representation in the calendar grid. But the grid cells are 4 characters wide and already dense. Adding event indicators requires careful visual design.

**Mitigation:** For the initial implementation, CalDAV events contribute to the day indicator (a day with events shows an indicator even if there are no todos). A distinct marker (e.g., a dot character instead of brackets) can differentiate "has events" from "has todos." Detailed event information is shown in the todo list pane, not the calendar grid.

---

## Sources

- [Google CalDAV API Developer's Guide](https://developers.google.com/workspace/calendar/caldav/v2/guide) -- endpoint URLs, supported methods, authentication requirements, iCalendar limitations (HIGH confidence)
- [DAVx5 Google Calendar documentation](https://www.davx5.com/tested-with/google) -- confirms basic auth disabled, OAuth required (HIGH confidence)
- [Google Workspace: Transition from less secure apps](https://support.google.com/a/answer/14114704) -- timeline for basic auth deprecation (HIGH confidence)
- [emersion/go-webdav CalDAV client API](https://pkg.go.dev/github.com/emersion/go-webdav/caldav) -- Client methods, CalendarQuery, CompFilter, types (HIGH confidence)
- [Bubble Tea async patterns](https://deepwiki.com/charmbracelet/bubbletea/6.4-step-by-step-tutorials) -- tea.Cmd for background operations (HIGH confidence)
- [Bubble Tea two-way goroutine communication](https://github.com/charmbracelet/bubbletea/issues/1244) -- current limitations, no built-in duplex pattern (MEDIUM confidence)
- [CalDAV timezone/all-day event issues](https://github.com/home-assistant/core/issues/126448) -- off-by-one day bug in multiple implementations (HIGH confidence)
- [Cal.com CalDAV implementation challenges](https://cal.com/blog/the-intricacies-and-challenges-of-implementing-a-caldav-supporting-system-for-cal) -- RRULE expansion, timezone complexity, provider quirks (MEDIUM confidence)
- [zalando/go-keyring](https://github.com/zalando/go-keyring) -- Linux D-Bus/Secret Service requirement, headless failure mode (HIGH confidence)
- [Go time/tzdata package](https://pkg.go.dev/time/tzdata) -- embedded IANA timezone database for portable binaries (HIGH confidence)
- [RFC 4791: CalDAV](https://datatracker.ietf.org/doc/html/rfc4791) -- CalendarQuery, time-range filter, EXPAND (HIGH confidence)
- [RFC 6578: Collection Synchronization](https://datatracker.ietf.org/doc/html/rfc6578) -- sync-token based incremental sync (HIGH confidence)
- Codebase analysis of `internal/app/model.go` -- Update loop structure, message routing, no async ops (HIGH confidence)
- Codebase analysis of `internal/config/config.go` -- TOML config, atomic save pattern, plaintext storage (HIGH confidence)
- Codebase analysis of `internal/store/todo.go` -- date format, precision system, Todo struct (HIGH confidence)
- Codebase analysis of `internal/calendar/model.go` -- synchronous store calls, no network operations (HIGH confidence)
