# Stack Research: Google Calendar Integration via CalDAV/REST API

**Domain:** Read-only Google Calendar event fetching, CalDAV client, OAuth 2.0 auth, background polling in Bubble Tea
**Researched:** 2026-02-13
**Confidence:** HIGH (Google auth requirements verified via official docs, library APIs verified via pkg.go.dev)

## Executive Summary

The original milestone assumption -- CalDAV with app password/basic auth for Google Calendar -- is **dead**. Google disabled "less secure apps" (basic auth, app passwords) for CalDAV on September 30, 2024. Google Calendar CalDAV now requires OAuth 2.0 exclusively. This fundamentally changes the stack requirements.

Two viable paths exist:

1. **Google Calendar REST API** (`google.golang.org/api/calendar/v3`) with OAuth 2.0 -- purpose-built Go client, native syncToken support, official Google library
2. **CalDAV via `emersion/go-webdav`** with OAuth 2.0 HTTP client -- protocol-level CalDAV, works with Google and self-hosted servers (Nextcloud, Radicale, Fastmail), but sync-token support in the CalDAV client is incomplete

**Recommendation: Dual-path architecture.** Use the Google Calendar REST API for Google accounts (better API, native sync tokens, official support) and `emersion/go-webdav` CalDAV client for self-hosted/third-party servers that still support basic auth (Nextcloud, Radicale, Baikal, Fastmail). Both paths share a common internal event interface. This avoids locking users into Google while avoiding the pain of OAuth-over-CalDAV for Google.

If the milestone scope is strictly "Google Calendar only," use path 1 alone. If future milestones will add Nextcloud/Fastmail/self-hosted CalDAV, build the abstraction now.

---

## Critical Finding: App Passwords Do Not Work for Google CalDAV

**Confidence:** HIGH (verified via official Google documentation)

Google's CalDAV endpoint (`https://apidata.googleusercontent.com/caldav/v2/`) refuses Basic Authentication with HTTP 401. This was enforced starting September 30, 2024 as part of the "less secure apps" shutdown. App-specific passwords generated in Google Account settings no longer work for CalDAV, IMAP, or SMTP.

The CalDAV API Developer's Guide explicitly states: "The CalDAV server will refuse to authenticate a request unless it arrives over HTTPS with OAuth 2.0 authentication of a Google account."

**Impact on milestone:** The "app password authentication" approach described in the milestone context is not possible for Google Calendar. OAuth 2.0 is mandatory.

**Source:** [Google CalDAV API Guide](https://developers.google.com/workspace/calendar/caldav/v2/guide), [Less Secure Apps Transition](https://support.google.com/a/answer/14114704)

---

## Recommended Stack

### Path 1: Google Calendar REST API (Primary -- for Google accounts)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `google.golang.org/api/calendar/v3` | v0.266.0 | Google Calendar client | Official Google client library. Purpose-built Events.List with native syncToken, time range filtering, SingleEvents expansion. In maintenance mode but stable and feature-complete. |
| `golang.org/x/oauth2` | latest | OAuth 2.0 token management | Official Go OAuth2 library. Handles token refresh automatically via TokenSource. Required by Google Calendar client. |
| `golang.org/x/oauth2/google` | (part of x/oauth2) | Google-specific OAuth endpoints | Provides Google endpoint configuration, credential file parsing, and scope definitions. |

**Why REST API over CalDAV for Google:**

1. **Same auth requirement.** Both Google CalDAV and REST API require OAuth 2.0. Zero auth advantage to CalDAV.
2. **Better sync support.** REST API has native `syncToken` on `Events.List` with well-documented 410-gone handling. The go-webdav CalDAV client lacks `SyncCollection` for CalDAV (only CardDAV has it).
3. **Official Go client.** `google.golang.org/api/calendar/v3` is maintained by Google, typed, and documented. CalDAV returns raw iCalendar data requiring manual parsing.
4. **Simpler event model.** REST API returns structured JSON with `Event.Summary`, `Event.Start.DateTime`, etc. CalDAV returns VCALENDAR/VEVENT iCal blobs that need `emersion/go-ical` to parse.
5. **Read-only scope.** `calendar.readonly` scope is purpose-built for this use case.

### Path 2: CalDAV Client (Secondary -- for self-hosted/third-party servers)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/emersion/go-webdav` | v0.7.0 | CalDAV client (WebDAV/CalDAV/CardDAV) | Only serious Go CalDAV client library. 455 stars, active development (last release Oct 2024). Provides `caldav.Client` with `FindCalendars`, `QueryCalendar`, `GetCalendarObject`. |
| `github.com/emersion/go-ical` | v0.0.0 (untagged) | iCalendar parsing | Required companion to go-webdav. Parses VEVENT from CalDAV responses. Same author (emersion). Provides `Event.DateTimeStart()`, `Event.Props.Text(ical.PropSummary)`. |

**Self-hosted servers that support basic auth:**
- **Nextcloud** -- username + app password at `https://nextcloud.example.com/remote.php/dav`
- **Radicale** -- username + password, flexible auth backends
- **Baikal** -- username + password at `https://baikal.example.com/cal.php/calendars/`
- **Fastmail** -- username + app password (Fastmail generates app-specific passwords)

For these servers, `webdav.HTTPClientWithBasicAuth(nil, user, pass)` works directly. No OAuth complexity.

### OAuth 2.0 Flow (for Google)

| Technology | Purpose | Why |
|------------|---------|-----|
| `golang.org/x/oauth2` | OAuth2 Config, token exchange, auto-refresh | Standard Go OAuth2 library. Creates `http.Client` with auto-refreshing token. |
| Loopback redirect (`http://127.0.0.1:PORT`) | Desktop/CLI OAuth flow | Google's recommended approach for native apps. App starts local HTTP server, opens browser, receives auth code via redirect. Manual copy-paste (OOB) is deprecated. |
| PKCE (S256) | Security enhancement | Required for public clients (no client_secret enforcement). Generate code_verifier + code_challenge. Supported by `golang.org/x/oauth2` via `oauth2.SetAuthURLParam`. |

**OAuth flow for TUI app:**

1. App generates PKCE code_verifier and code_challenge
2. App starts temporary HTTP server on `127.0.0.1` (random port)
3. App opens browser to Google auth URL (or prints URL for user to open)
4. User consents in browser
5. Google redirects to `127.0.0.1:PORT/callback` with auth code
6. App exchanges code for access_token + refresh_token
7. Tokens stored in `~/.config/todo-calendar/google-token.json` (mode 0600)
8. Subsequent launches: load token, auto-refresh via `oauth2.TokenSource`

**Scope:** `https://www.googleapis.com/auth/calendar.readonly` (read-only, narrowest scope)

**Google API Console requirement:** User must create a project in Google API Console and download `credentials.json` (client_id + client_secret for "Desktop" application type). This is a one-time setup step.

### Background Polling in Bubble Tea

| Pattern | Purpose | Why |
|---------|---------|-----|
| `tea.Cmd` (command pattern) | Async calendar fetch | Bubble Tea executes commands in goroutines internally. Never use raw goroutines. Return a `tea.Cmd` from `Init()` for first fetch and from `Update()` when poll timer fires. |
| `tea.Tick` | Periodic polling | `tea.Tick(5*time.Minute, func(t time.Time) tea.Msg { return calendarPollMsg(t) })` fires every 5 minutes. Returned from `Update()` when `calendarPollMsg` is received, creating a self-renewing poll loop. |

**Polling architecture in Elm/Bubble Tea:**

```go
// Message types
type calendarFetchStartMsg struct{}
type calendarFetchDoneMsg struct {
    events []CalendarEvent
    err    error
}
type calendarPollMsg time.Time

// Init -- trigger first fetch
func (m Model) Init() tea.Cmd {
    return fetchCalendarEventsCmd(m.calendarService)
}

// Command -- runs in goroutine, returns message
func fetchCalendarEventsCmd(svc CalendarService) tea.Cmd {
    return func() tea.Msg {
        events, err := svc.FetchEvents(context.Background(), time.Now(), time.Now().AddDate(0, 1, 0))
        return calendarFetchDoneMsg{events: events, err: err}
    }
}

// Update -- handle results, schedule next poll
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case calendarFetchDoneMsg:
        if msg.err != nil {
            m.calendarError = msg.err
        } else {
            m.calendarEvents = msg.events
        }
        // Schedule next poll
        return m, tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
            return calendarPollMsg(t)
        })
    case calendarPollMsg:
        // Poll fired, do another fetch
        return m, fetchCalendarEventsCmd(m.calendarService)
    }
    return m, nil
}
```

**Key rules:**
- Never use `go func()` inside Bubble Tea. Always return `tea.Cmd`.
- Error handling via message types, not panics.
- First fetch in `Init()` (or from `main.go` before TUI, storing results in model).
- Subsequent fetches via `tea.Tick` -> `calendarPollMsg` -> `fetchCalendarEventsCmd` cycle.
- SyncToken stored in model (or SQLite), passed to fetch command for delta sync.

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/json` | stdlib | Token persistence, event caching | Store OAuth tokens and cached events in JSON files |
| `net/http` | stdlib | Loopback OAuth server, HTTP client | Temporary server for OAuth callback, base for OAuth HTTP client |
| `os/exec` | stdlib | Open browser for OAuth | `xdg-open` / `open` to launch browser for consent |
| `crypto/sha256` | stdlib | PKCE code challenge | Generate S256 code challenge from code verifier |
| `crypto/rand` | stdlib | PKCE code verifier | Generate random code verifier for OAuth flow |

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Google Calendar access | REST API (`calendar/v3`) | CalDAV via go-webdav | Same OAuth requirement, worse sync support, requires iCal parsing, no advantage |
| CalDAV library | `emersion/go-webdav` v0.7.0 | `samedi/caldav-go` | Server-focused, less maintained, no client discovery methods |
| CalDAV library | `emersion/go-webdav` v0.7.0 | `azoff/caldav-go` / `cj123/caldav-go` | Forks of old project, minimal maintenance, fewer features |
| OAuth library | `golang.org/x/oauth2` | Manual HTTP token exchange | Standard library handles refresh, expiry, token reuse. Rolling your own is error-prone. |
| iCal parsing | `emersion/go-ical` | `arran4/golang-ical` | go-ical is the companion to go-webdav (same author), consistent API, used in CalDAV response parsing |
| Polling | `tea.Tick` + `tea.Cmd` | Raw goroutine + channel | Violates Bubble Tea architecture. Commands are the idiomatic way. |
| Background sync | `tea.Tick` every 5 min | `tea.Every` / filesystem watcher | `tea.Tick` is the standard polling primitive. No filesystem to watch (data is remote). |
| Token storage | JSON file (0600 perms) | System keychain | Keychain requires platform-specific code (Linux: libsecret, macOS: Keychain). JSON file matches existing TOML config simplicity. |
| Event caching | SQLite (existing DB) | Separate JSON file | Events are date-indexed data, belong in the same DB as todos. SQLite queries for "events on date X" are efficient. |

---

## Deliberately NOT Adding

| Consideration | Decision | Rationale |
|---------------|----------|-----------|
| **Full OAuth server** (e.g., `ory/fosite`) | Not adding | We only need a single-user OAuth flow with one provider (Google). The standard `golang.org/x/oauth2` library handles everything. A full OAuth server is for multi-tenant applications. |
| **Google API client library generator** (`google-api-go-client`) | Using directly | `google.golang.org/api/calendar/v3` is the generated library. No code generation step needed by us. |
| **CalDAV server** (e.g., go-webdav's `Handler`) | Not adding | We are a CalDAV client, not a server. The `caldav.Handler` and `Backend` types are for running a CalDAV server. |
| **Push notifications** (`Events.Watch` / webhooks) | Not adding | Watch requires a publicly-accessible webhook endpoint. A TUI app behind NAT has no public URL. Polling is the correct approach for a local desktop app. |
| **Google Cloud client libraries** (`cloud.google.com/go`) | Not adding | Google Calendar API is a Workspace API, not a Cloud API. The `google.golang.org/api/calendar/v3` package is the correct client. |
| **Browser automation / headless OAuth** | Not adding | OAuth consent must happen in the user's real browser (Google blocks embedded webviews). `xdg-open` / `open` to launch the system browser is the standard approach. |
| **Multiple Google account support** | Not adding (v1) | Scope to single account initially. Multi-account adds token management complexity. Can be extended later by storing tokens per-account. |
| **Write access** | Not adding | Read-only scope (`calendar.readonly`). No event creation, modification, or deletion. Reduces permission scope and risk. |

---

## Installation

```bash
# Google Calendar REST API client + OAuth
go get google.golang.org/api/calendar/v3
go get golang.org/x/oauth2
go get golang.org/x/oauth2/google

# CalDAV client (for non-Google providers)
go get github.com/emersion/go-webdav@v0.7.0

# iCalendar parsing (companion to go-webdav, for CalDAV responses)
go get github.com/emersion/go-ical
```

**Note on dependency weight:**
- `google.golang.org/api/calendar/v3` brings in `google.golang.org/api` (large but well-maintained). This is the heaviest new dependency.
- `golang.org/x/oauth2` is lightweight, already a transitive dependency in many Go projects.
- `emersion/go-webdav` v0.7.0 has minimal dependencies (just `emersion/go-ical` and Go stdlib).
- If deferring CalDAV support to a future milestone, skip go-webdav and go-ical entirely. Only `calendar/v3` + `x/oauth2` are needed for Google-only.

---

## Integration Points with Existing Codebase

### Config (TOML)

New TOML config section for calendar integration:

```toml
[calendar]
enabled = false
provider = "google"  # "google" | "caldav"

# Google-specific
google_credentials_file = ""  # path to credentials.json from API Console

# CalDAV-specific (for Nextcloud/Radicale/Fastmail)
caldav_url = ""
caldav_username = ""
caldav_password_cmd = ""  # command to retrieve password (e.g., "pass show caldav")

# Sync settings
poll_interval_minutes = 5
sync_window_days = 30
```

**Why `caldav_password_cmd` instead of `caldav_password`:** Never store passwords in config files. A command like `pass show caldav` or `security find-generic-password -w -s caldav` retrieves the password from a secure store at runtime.

### Store (SQLite)

New table for cached calendar events (schema migration v8 or higher, depending on what version the project is at when this milestone starts):

```sql
CREATE TABLE calendar_events (
    id          TEXT PRIMARY KEY,  -- provider-specific event ID
    provider    TEXT NOT NULL,     -- "google" | "caldav"
    summary     TEXT NOT NULL,
    start_time  TEXT NOT NULL,     -- ISO 8601
    end_time    TEXT NOT NULL,     -- ISO 8601
    all_day     INTEGER NOT NULL DEFAULT 0,
    location    TEXT NOT NULL DEFAULT '',
    calendar_id TEXT NOT NULL,     -- which calendar this belongs to
    updated_at  TEXT NOT NULL      -- when this cache entry was last updated
);

CREATE INDEX idx_calendar_events_date ON calendar_events(start_time, end_time);
```

Sync metadata (syncToken, last poll time) stored in a separate table or in the config:

```sql
CREATE TABLE calendar_sync (
    calendar_id TEXT PRIMARY KEY,
    sync_token  TEXT NOT NULL DEFAULT '',
    last_synced TEXT NOT NULL
);
```

### Bubble Tea Model

The calendar fetch system integrates as a background service in `app.Model`:

- `calendarEvents []CalendarEvent` -- cached events for display
- `calendarError error` -- last fetch error (shown in status bar)
- `calendarSyncing bool` -- true during fetch (show spinner)
- `calendarService CalendarService` -- interface for Google/CalDAV fetcher

Events are rendered alongside todos in the todolist view for the selected date. A `CalendarEvent` is visually distinct from a `Todo` (different style, not toggleable, not editable).

---

## Sync-Token Delta Fetching

### Google REST API (native support)

**Confidence:** HIGH (documented in official Google sync guide)

1. **Initial sync:** `Events.List("primary").TimeMin(now).TimeMax(now+30d).SingleEvents(true).Do()` -- returns all events + `NextSyncToken`
2. **Delta sync:** `Events.List("primary").SyncToken(stored).Do()` -- returns only changed events since last sync
3. **Token expiry:** Server returns HTTP 410 Gone. Clear local cache, do full sync again.
4. **Deleted events:** Delta response includes events with `status: "cancelled"`. Remove from local cache.

### CalDAV (partial support in go-webdav)

**Confidence:** MEDIUM (CardDAV has SyncCollection, CalDAV does not in go-webdav v0.7.0)

The go-webdav library implements `SyncCollection` for CardDAV (`carddav.Client.SyncCollection`) but **not** for CalDAV. The CalDAV client has `QueryCalendar` with time-range filtering but no sync-token-based delta sync.

**Workarounds for CalDAV delta sync:**
1. **CTag comparison:** Fetch the calendar's CTag (collection tag) via PROPFIND. If CTag unchanged, skip sync. If changed, do full `QueryCalendar` for the time window. This is a coarse "has anything changed" check, not a true delta.
2. **Full re-fetch with dedup:** Always do `QueryCalendar` for the time window, compare with cached events by UID, update/insert/delete as needed. Works for small calendars (< 1000 events in 30 days).
3. **Contribute SyncCollection to go-webdav:** The protocol support exists in RFC 6578, and the CardDAV side is implemented. Adding CalDAV SyncCollection is possible but out of scope for this milestone.

**Recommendation for CalDAV path:** Use full re-fetch with time-range filtering. For personal calendars (tens of events per month), the performance difference between delta and full sync is negligible. Store events in SQLite, diff against cached data.

---

## OAuth Token Lifecycle

**First-time setup:**
1. User places `credentials.json` (from Google API Console) in config directory
2. User enables calendar in TOML config: `[calendar] enabled = true, provider = "google"`
3. On next app launch, if no `google-token.json` exists:
   - App prints "Opening browser for Google Calendar authorization..."
   - Starts loopback server on `127.0.0.1:PORT`
   - Opens browser with `xdg-open` (Linux) / `open` (macOS)
   - User consents, redirected back to loopback
   - Token saved to `~/.config/todo-calendar/google-token.json`
4. TUI launches normally with calendar events

**Subsequent launches:**
1. Load `google-token.json`
2. `oauth2.TokenSource` auto-refreshes expired access tokens using refresh token
3. If refresh token is revoked/expired, prompt re-authorization (same flow as first-time)

**Token file location:** Same directory as the TOML config and SQLite database. The project already uses `~/.config/todo-calendar/` (or `$XDG_CONFIG_HOME`).

---

## Risk: Google API Console Requirement

**Confidence:** HIGH (this is a fundamental requirement)

Unlike many integrations, Google Calendar REST API requires the user to:
1. Go to console.cloud.google.com
2. Create a project
3. Enable the Google Calendar API
4. Create OAuth 2.0 credentials (Desktop application type)
5. Download `credentials.json`

This is a non-trivial setup step for non-technical users. Alternatives:
- **Ship a client_id with the app:** Possible for open-source projects, but Google may rate-limit or require verification for many users. Google's policy allows unverified apps for <100 users.
- **Use a Google Workspace service account:** Only works for Workspace (paid) accounts, not personal Gmail.
- **Document the setup clearly:** Most realistic approach. A one-time 5-minute setup guide.

**Recommendation:** For a personal TUI app, document the setup process. This is the same approach used by `gcalcli`, `khal+vdirsyncer`, and other CLI calendar tools.

---

## Sources

### Official Documentation (HIGH confidence)
- [Google CalDAV API Guide](https://developers.google.com/workspace/calendar/caldav/v2/guide) -- confirms OAuth 2.0 requirement, no basic auth
- [Google Calendar REST API Go Quickstart](https://developers.google.com/workspace/calendar/api/quickstart/go) -- OAuth flow for Go CLI apps
- [Google OAuth 2.0 for Native Apps](https://developers.google.com/identity/protocols/oauth2/native-app) -- loopback redirect, PKCE, no OOB
- [Google Calendar API Sync Guide](https://developers.google.com/workspace/calendar/api/guides/sync) -- syncToken flow, 410 handling
- [Less Secure Apps Transition](https://support.google.com/a/answer/14114704) -- app passwords disabled Sept 2024
- [Google Workspace Blog: LSA Shutdown](https://workspaceupdates.googleblog.com/2023/09/winding-down-google-sync-and-less-secure-apps-support.html)

### Package Documentation (HIGH confidence)
- [google.golang.org/api/calendar/v3 on pkg.go.dev](https://pkg.go.dev/google.golang.org/api/calendar/v3) -- v0.266.0, Events.List, SyncToken, scopes
- [golang.org/x/oauth2 on pkg.go.dev](https://pkg.go.dev/golang.org/x/oauth2) -- TokenSource, Config, auto-refresh
- [emersion/go-webdav on pkg.go.dev](https://pkg.go.dev/github.com/emersion/go-webdav) -- v0.7.0, HTTPClient interface, HTTPClientWithBasicAuth
- [emersion/go-webdav caldav package](https://pkg.go.dev/github.com/emersion/go-webdav/caldav) -- Client methods: FindCalendars, QueryCalendar
- [emersion/go-ical on pkg.go.dev](https://pkg.go.dev/github.com/emersion/go-ical) -- Event, Decoder, DateTimeStart, PropSummary

### GitHub Repositories (MEDIUM confidence)
- [emersion/go-webdav releases](https://github.com/emersion/go-webdav/releases) -- v0.7.0 Oct 2024, 455 stars
- [emersion/go-ical](https://github.com/emersion/go-ical) -- untagged, actively maintained

### Community/Ecosystem (MEDIUM confidence)
- [Bubble Tea Commands Blog Post](https://charm.land/blog/commands-in-bubbletea/) -- tea.Cmd pattern, no raw goroutines
- [Bubble Tea DeepWiki: HTTP and Async Operations](https://deepwiki.com/charmbracelet/bubbletea/6.4-step-by-step-tutorials) -- fetch command pattern
- [Bubble Tea DeepWiki: Core Components](https://deepwiki.com/charmbracelet/bubbletea/2-core-components) -- tea.Tick for polling
