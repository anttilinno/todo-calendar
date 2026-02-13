# Feature Landscape: Read-Only Google Calendar Event Display

**Domain:** Read-only Google Calendar event display in a Go TUI calendar+todo app
**Researched:** 2026-02-13
**Confidence:** HIGH for event rendering patterns (well-established in calcure, gcal-tui, Todoist, Things 3, WTF dashboard), MEDIUM for TUI-specific visual distinction (fewer direct precedents for events-mixed-with-todos in terminal apps), HIGH for Google Calendar API behavior (official docs verified)

---

## Table Stakes

Features users expect when calendar events appear alongside todos. Missing any of these makes the integration feel broken or confusing.

| Feature | Why Expected | Complexity | Depends On | Notes |
|---------|--------------|------------|------------|-------|
| Timed events with time prefix | Every calendar app shows event times. Without time display, events are indistinguishable from todos. Format: `09:00 Team Standup` or `09:00-09:30 Team Standup`. Users cannot plan their day around events without seeing when they occur. | LOW | Event data model with parsed start/end times | Use 24h format by default (matches TUI conventions in calcure, khal, gcalcli). Consider making 12h/24h configurable in settings. The time prefix is the primary visual cue that distinguishes events from todos. |
| All-day events without time prefix | Google Calendar distinguishes all-day events (`start.date`) from timed events (`start.dateTime`). All-day events should display without a time prefix, similar to how Todoist shows them in an "All day" section and Things 3 shows them at the top. | LOW | Detection of all-day vs timed events from API response (check `start.date` vs `start.dateTime`) | All-day events should render like `  All day: Company Offsite` or simply with a distinct marker. They sort before timed events within a day, matching Google Calendar and Todoist behavior. |
| Visual distinction from todos | The core UX requirement. Events must be instantly recognizable as "not a todo" at a glance. Todoist uses color-coded time blocks. Things 3 groups events at the top. Calcure uses separate colors (`color_events` vs `color_todo`). WTF dashboard uses color-coded event titles. | LOW | New styles in `todolist/styles.go` and `theme/theme.go` for event rendering | Recommendation: Use a distinct foreground color (e.g., cyan/teal family) for event text, plus a `[CAL]` or clock icon prefix. No checkbox -- events are not completable. The absence of `[ ]` plus a different color is a strong double signal. This matches how `[R]` and `[+]` indicators work in the existing codebase. |
| Events appear in the dated section | Events belong alongside dated todos -- they are things happening on specific dates. Todoist groups events at the top of the day's task list. Things 3 shows events at the top of the Today view. | LOW | Modification to `visibleItems()` in `todolist/model.go` to merge events into the dated section | Events should sort above todos within the same date section. Rationale: events are fixed commitments (externally scheduled), while todos are flexible. Showing events first gives the user a "what's locked in" context before the "what I need to do" list. |
| Events in weekly view filtering | The app already filters todos by week range when in WeekView. Events must respect the same filter. If viewing the week of Feb 10-16, only events in that range appear. | LOW | Reuse existing `weekFilterStart`/`weekFilterEnd` mechanism for event date filtering | This is structural -- events share the same date-range filtering that todos use. The `visibleItems()` method already handles week filtering for todos; events just need to be included in the same logic. |
| Events in monthly view | When viewing a full month, events for that month appear in the dated section alongside that month's todos. | LOW | Events filtered by `viewYear`/`viewMonth`, same as `TodosForMonth` | Match the existing pattern: `store.TodosForMonth(year, month)` returns todos; a parallel `EventsForMonth(year, month)` or equivalent returns events. |
| Multi-day event display | Google Calendar supports multi-day events (e.g., "Conference" spanning Feb 10-14). Users expect to see these events listed on each day they span, not just the start day. Todoist and Google Calendar both show multi-day events on every day within the range. | MEDIUM | Expanding multi-day events into per-day entries during event processing | When a multi-day event spans Feb 10-14 and the user views February, it should appear under each day (or at least show on each day with a visual hint like "Day 2/5"). The simplest approach: expand multi-day events into one display entry per day during the API-to-local-model transformation. This avoids complex rendering logic. |
| Recurring event expansion | Google Calendar API with `singleEvents=true` expands recurring events into individual instances. Users expect to see "Team Standup" on Monday, Tuesday, Wednesday, etc. as separate entries, not a single "recurring event" object. | LOW | Pass `SingleEvents(true)` when calling the Google Calendar API `Events.List` | This is an API-level concern, not a rendering concern. With `singleEvents=true`, the API returns individual instances already. No client-side recurrence expansion needed. HIGH confidence -- verified in official Google Calendar API docs. |
| Event time respects user timezone | Events must display in the user's local timezone, not UTC or the calendar's timezone. A meeting at "09:00 EST" should show as "09:00" for an EST user and "15:00" for a CET user. | MEDIUM | Timezone conversion using Go's `time.LoadLocation` and the event's timezone info | The Google Calendar API returns events in the requested `timeZone` parameter. Pass the user's local timezone (from `time.Now().Location()` or a config setting). If the API handles timezone conversion server-side, the client just parses the returned times. LOW risk if we rely on API-side conversion. |
| Polling refresh (5-minute interval) | Events change externally (someone reschedules a meeting). The app needs to periodically re-fetch events without user intervention. The project context specifies 5-minute polling. | MEDIUM | Background ticker using Bubble Tea's `tea.Tick` or `tea.Every`, re-fetch events on tick | Bubble Tea supports `tea.Tick` for periodic commands. On each tick: call API, compare with cached events, update display if changed. The refresh should be non-blocking -- fetch in a goroutine, deliver results via a message. |
| Not selectable / not actionable | Events are read-only. The cursor should skip over events when navigating the todo list, or if events are selectable, pressing Enter/toggle/delete/edit should do nothing (or show "read-only" hint). | LOW | Modify `selectableIndices()` to skip event items, or add a new `itemKind` for events | Recommendation: Add `eventItem` as a new `itemKind` in `todolist/model.go`. The `selectableIndices()` function already filters to `todoItem` kind only -- events will be naturally skipped. This is the cleanest approach and matches how `headerItem` and `emptyItem` are already non-selectable. |
| Calendar grid indicators include events | The calendar grid already shows `[15]` brackets and priority colors for days with todos. Days with events should also show an indicator so the user knows something is happening on that day even before looking at the todo list. | MEDIUM | Extend `indicators`/`totals` maps in calendar model to include event counts, or add a parallel `eventDays` map | The calendar grid's cell styling logic (`hasPending`, `hasAllDone`) needs to account for events. Simplest approach: a day with events gets a bracket indicator even if it has no todos. A new indicator style (e.g., a dot or different bracket style) could distinguish "has events only" from "has todos only" from "has both". |

## Differentiators

Features that go beyond basic event display and make the integration feel polished. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Event color from Google Calendar | Google Calendar assigns colors to calendars (Work, Personal, etc.) and individual events. Reflecting these colors in the TUI gives the user the same visual language they are used to from the web/mobile app. | MEDIUM | The API returns `colorId` on events and calendar-level colors. Map Google's 11 event colors to terminal-safe ANSI colors. Needs a color mapping table. May clash with the app's theme system -- needs careful integration. |
| Event location display | Show location on a separate line or as a suffix when the user highlights an event. Useful for "where is this meeting?" context. Example: `09:00 Team Standup  [Room 3B]` | LOW | Parse `event.Location` field. Render as dimmed suffix after event title, similar to how `[+]` and `[R]` indicators work. Only show when space permits (truncate for narrow terminals). |
| Event preview on highlight | When the cursor is near an event (even though events are not selectable), show event details in a preview pane or status bar: full title, time, location, description snippet. | MEDIUM | Could reuse the existing preview overlay pattern (`PreviewMsg`), but would need to work differently since events are not selectable. Alternative: show event details in the help bar area when the cursor is adjacent to an event. |
| Events in search results | The full-screen search currently searches todos only. Including events in search results lets the user find "When is the dentist appointment?" across all months. | MEDIUM | Extend `search/model.go` to also search through cached events. Events would need a different visual treatment in results (no checkbox, event color, time prefix). The fuzzy search infrastructure works on text; event titles are text. |
| Events in inline filter | The `/` filter mode currently narrows visible todos by fuzzy match. If events are also visible items, the filter should also match against event titles. | LOW | The filter already operates on `visibleItems()`. If events are included in `visibleItems()` with a title field, the fuzzy match can include them. Just need to match on the event's display text. |
| Multiple calendar support | Google users often have multiple calendars (Work, Personal, Birthdays, Holidays). The API can list events from all calendars. Showing events from multiple calendars with distinct visual treatment (e.g., dimmed personal events during work hours). | HIGH | Need calendar list API call, per-calendar color mapping, settings for which calendars to show. Significantly expands the settings surface. Recommend deferring to after single-calendar works. |
| "Open in browser" action | When an event is highlighted or selected, press a key to open the event in Google Calendar web UI. Provides an escape hatch for seeing full event details, attendees, video call links. | LOW | Each event has an `htmlLink` field. Open with `xdg-open` (Linux), `open` (macOS), or `cmd /c start` (Windows). Existing `editor.go` shows the pattern for exec'ing external programs. |
| Stale data indicator | Show a visual indicator when event data is stale (last fetch was >10 minutes ago, or API call failed). Prevents the user from trusting outdated information. | LOW | Track last-successful-fetch timestamp. Compare against current time. Show a dimmed "(stale)" or warning icon in the event section header if data is old. |
| Event count in calendar overview | The calendar overview section shows todo counts per month. Could include event counts: `February 2026   3 todos   12 events`. | LOW | Count events per month from the cache. Render alongside existing `pending`/`completed` counts. |

## Anti-Features

Features to explicitly NOT build. These would add complexity without proportional value for a read-only TUI integration.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Event creation from TUI | The entire milestone is read-only display. Creating events requires write scopes (broader OAuth permissions), attendee management, recurrence rule authoring, and conflict handling. Massively out of scope. | Events are created in Google Calendar web/mobile. The TUI is a read-only viewer. |
| Event editing from TUI | Same as creation: write operations require broader permissions, undo logic, conflict resolution with other attendees' changes, and a complex editing UI for time, title, location, recurrence. | Read-only. To edit, use "open in browser" action (differentiator). |
| RSVP / response from TUI | Accepting/declining events requires write API access and a confirmation UI. gcal-tui supports this but it is a full Google Calendar client, not a sidebar display. | The TUI shows events as information. RSVP happens in the web/mobile app. |
| Event reminders / notifications | The TUI has no daemon process, no system tray, no notification infrastructure. Adding reminders would require a background service and OS-level notification integration. | Events are visible when the user opens the app. The app is not a reminder system. |
| Free/busy overlay on calendar grid | Showing time blocks as colored regions on the calendar grid (like Google Calendar's day/week view) would require a completely different grid rendering approach. The current grid is 4-char-per-day cells -- too small for time blocks. | Events display as a list in the todo panel, not as blocks on the calendar grid. Calendar grid shows day-level indicators only. |
| CalDAV / generic calendar protocol support | CalDAV would enable iCloud Calendar, Outlook.com, Fastmail, etc. But CalDAV is a complex protocol with authentication variations, sync tokens, VTIMEZONE handling, and partial update semantics. | Google Calendar API only, via OAuth2. Users of other providers are out of scope for this milestone. CalDAV could be a future milestone. |
| Offline event caching with sync | Persisting events to SQLite for offline access adds schema migrations, sync conflict handling (what if an event was deleted server-side?), and stale-data management. | Events are in-memory only. Fetched on startup and every 5 minutes. If offline, show "unable to fetch" message with last-known data from current session. No persistence between app restarts. |
| Event attachments / video call links | Parsing and displaying Google Meet links, Zoom links, or file attachments from events. | Not displayed. If the user needs this detail, they use "open in browser". |
| Drag-and-drop time blocking | Dragging todos onto event time slots to time-block your day. This is a Todoist premium feature that requires a completely different interaction model (mouse-driven, visual time grid). | Not applicable in a keyboard-driven TUI. |

---

## Event Rendering Specification

### Rendering Format

Based on patterns observed in gcal-tui (time range + status marker), WTF dashboard (time + title + color), calcure (separate color for events), and Todoist (events grouped at top, color-coded):

**Timed event:**
```
     09:00 Team Standup
     10:30 Design Review
```

**All-day event:**
```
     all day  Company Offsite
```

**Key rendering decisions:**
- No checkbox (`[ ]` or `[x]`) -- events are not completable
- No priority badge slot -- events do not have priorities
- 5-char empty padding in place of priority badge slot to maintain alignment with todos
- Time displayed in `HH:MM` format, left-aligned where the checkbox would be
- Event text colored with a distinct "event" color (recommend cyan/teal family)
- `all day` label for all-day events, styled in the same muted color as the time
- No `[+]`, `[R]`, or date suffix -- those are todo-specific indicators

**Mixed list appearance (events above todos within a section):**
```
  February 2026
  ──────────
     09:00 Team Standup
     10:30 Design Review
     all day  Company Offsite
  > [P2] [ ] Write documentation                2026-02-13
         [ ] Fix login bug                       2026-02-13
         [x] Review PR                           2026-02-12
```

### Sorting Within a Date Section

1. All-day events first (alphabetical by title)
2. Timed events second (chronological by start time)
3. Todos third (existing sort order by `sort_order` field)

This matches Todoist's behavior: events at top, tasks below. It also matches Things 3: "calendar events at the top" of the Today view.

### The `eventItem` Kind

Add to `todolist/model.go`:

```go
const (
    headerItem itemKind = iota
    todoItem
    emptyItem
    eventItem  // NEW: Google Calendar event (read-only, not selectable)
)
```

The `visibleItem` struct gains an event field:

```go
type visibleItem struct {
    kind    itemKind
    label   string
    todo    *store.Todo
    event   *CalendarEvent  // non-nil only for eventItem
    section sectionID
}
```

Where `CalendarEvent` is a local display model (not a Google API type):

```go
type CalendarEvent struct {
    Title     string
    StartTime string    // "HH:MM" or "" for all-day
    EndTime   string    // "HH:MM" or "" for all-day
    AllDay    bool
    Date      string    // "YYYY-MM-DD"
    Location  string    // optional
    Color     string    // optional, from Google Calendar
    Source    string    // calendar name, e.g., "Work"
}
```

---

## Feature Dependencies on Existing Features

```
Event Data Layer (fetching + caching)
    |
    +-- REQUIRES: Google OAuth2 authentication (new)
    +-- REQUIRES: Google Calendar API client (new)
    +-- NEW: internal/gcal package (auth, client, event model)
    +-- INDEPENDENT OF: existing store/SQLite (events are in-memory)
    |
    v
Event Display in Todo Panel
    |
    +-- REQUIRES: Event Data Layer
    +-- MODIFIES: todolist/model.go -- new eventItem kind, event field on visibleItem
    +-- MODIFIES: todolist/model.go -- visibleItems() merges events into dated section
    +-- MODIFIES: todolist/model.go -- selectableIndices() already skips non-todoItem
    +-- MODIFIES: todolist/styles.go -- new EventTime, EventTitle, EventAllDay styles
    +-- MODIFIES: theme/theme.go -- new EventFg, EventTimeFg color roles
    +-- INTERACTS WITH: weekFilterStart/weekFilterEnd (events respect week filtering)
    +-- INTERACTS WITH: filterMode (events matched by inline filter)
    |
    v
Calendar Grid Event Indicators
    |
    +-- REQUIRES: Event Display (to have events in memory)
    +-- MODIFIES: calendar/model.go -- new eventDays map[int]bool
    +-- MODIFIES: calendar/grid.go -- cell styling accounts for event presence
    +-- INTERACTS WITH: existing indicator/priority/holiday styling cascade
    |
    v
Polling Refresh
    |
    +-- REQUIRES: Event Data Layer
    +-- MODIFIES: app/model.go -- tea.Tick subscription, refresh command
    +-- MODIFIES: app/model.go -- handle EventsUpdatedMsg, propagate to todolist + calendar
    +-- INTERACTS WITH: overlay routing (refresh should work even with overlays open)
```

**Integration with existing features:**

| Existing Feature | How Events Interact | Complexity |
|-----------------|---------------------|------------|
| Monthly view | Events for current month appear in dated section | LOW |
| Weekly view | Events filtered to current week's date range | LOW |
| Inline filter (`/`) | Events matched by fuzzy filter on title text | LOW |
| Full-screen search | Events optionally searchable (differentiator, not table stakes) | MEDIUM |
| Settings overlay | New settings: Google account, calendar selection, 12h/24h time | MEDIUM |
| Preview overlay | Could preview event details (differentiator) | MEDIUM |
| Priorities | No interaction -- events have no priority | NONE |
| Templates/Recurring | No interaction -- events are external, not template-based | NONE |
| Todo reordering (K/J) | No interaction -- events are not reorderable | NONE |

---

## MVP Recommendation

### Must-Have for v3.0 (minimum viable Google Calendar integration)

1. **Timed events with HH:MM prefix** -- the core visual format
2. **All-day events with "all day" label** -- handles the two event types
3. **Visual distinction from todos** -- no checkbox, distinct color, time prefix
4. **Events in dated section above todos** -- proper sorting
5. **Events in both monthly and weekly views** -- respects existing view modes
6. **Multi-day event expansion** -- one entry per day
7. **Recurring event expansion via API** -- `singleEvents=true`
8. **Timezone handling via API parameter** -- local timezone
9. **5-minute polling refresh** -- background timer
10. **Events not selectable** -- cursor skips event items
11. **Calendar grid day indicators include events** -- brackets on days with events

**Defer to later:**

- **Event color from Google Calendar** -- nice polish but not required. Ship with a single event color first.
- **Events in search results** -- valuable but search works fine without events initially.
- **Multiple calendar support** -- start with primary calendar only. Add calendar selection later.
- **Open in browser** -- low effort but not blocking for MVP.
- **Event location display** -- minor enhancement after core rendering works.
- **Event preview on highlight** -- requires UX design for non-selectable item interaction.

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Risk | Priority |
|---------|-----------|-------------------|------|----------|
| Timed event display with HH:MM | CRITICAL | LOW | LOW | P0 |
| All-day event display | CRITICAL | LOW | LOW | P0 |
| Visual distinction (color + no checkbox) | CRITICAL | LOW | LOW | P0 |
| Events above todos in dated section | HIGH | LOW | LOW | P0 |
| Week view filtering for events | HIGH | LOW | LOW | P0 |
| Multi-day event expansion | HIGH | MEDIUM | LOW | P0 |
| Recurring expansion (singleEvents=true) | HIGH | LOW | LOW | P0 |
| Timezone conversion | HIGH | MEDIUM | MEDIUM | P0 |
| 5-minute polling refresh | HIGH | MEDIUM | LOW | P0 |
| Events not selectable (cursor skip) | HIGH | LOW | LOW | P0 |
| Calendar grid event indicators | MEDIUM | MEDIUM | LOW | P1 |
| Inline filter matches events | MEDIUM | LOW | LOW | P1 |
| Event color from Google | MEDIUM | MEDIUM | MEDIUM | P2 |
| Events in search results | MEDIUM | MEDIUM | LOW | P2 |
| Multiple calendar support | LOW | HIGH | HIGH | P3 |
| Open in browser | LOW | LOW | LOW | P2 |
| Event location display | LOW | LOW | LOW | P2 |
| Event preview on highlight | LOW | MEDIUM | MEDIUM | P3 |
| Stale data indicator | LOW | LOW | LOW | P2 |
| Event count in overview | LOW | LOW | LOW | P3 |

---

## Competitor / Prior Art Summary

| App | Events + Tasks Together? | Visual Distinction | All-Day Handling | Event Position |
|-----|-------------------------|-------------------|-----------------|----------------|
| Todoist | Yes, events at top of day | Color-coded time blocks | Separate "All day" section at top | Events grouped in "event stack" above tasks |
| Things 3 | Yes, events at top of Today view | Distinct section, calendar colors | At the top, no time shown | Calendar events at the very top |
| Calcure | Events left (calendar), tasks right (journal) | Separate panels, `color_events` vs `color_todo` | Shown on calendar date | Separate panel entirely |
| gcal-tui | Events only (no todos) | Status markers (checkmark/X/dot), color | Filtered as "no specific start/end times" | List sorted by start time |
| WTF dashboard | Events only (widget) | Color-coded by regex match, icons | N/A (agenda widget) | Chronological list |
| Fantastical | Yes, events + tasks in unified list | Color pills for events, checkboxes for tasks | At top of day view | Events integrated into vertical list |

**Key insight:** Every app that mixes events and tasks puts events above tasks. Every app uses color and/or structural cues (no checkbox, time prefix) to distinguish events from tasks. Our approach follows this consensus.

---

## Sources

### HIGH Confidence (official docs, verified)
- [Google Calendar API: Events concepts](https://developers.google.com/workspace/calendar/api/concepts/events-calendars) -- All-day vs timed events, timezone handling, recurring event expansion
- [Google Calendar API: Events.list](https://developers.google.com/workspace/calendar/api/v3/reference/events/list) -- singleEvents parameter, timeZone parameter, timeMin/timeMax filtering
- [Google Calendar API: Recurring events](https://developers.google.com/workspace/calendar/api/guides/recurringevents) -- singleEvents=true expands recurrences
- [Google Calendar API Go package](https://pkg.go.dev/google.golang.org/api/calendar/v3) -- Go client library for Calendar API
- [Todoist: Use Calendar with Todoist](https://todoist.com/help/articles/use-calendar-with-todoist-rCqwLCt3G) -- Events grouped in "event stack" at top, read-only, color-coded
- [Things 3: Today, Upcoming, Anytime, Someday](https://culturedcode.com/things/support/articles/4001304/) -- Calendar events at the top of Today view
- Existing codebase: `internal/todolist/model.go` lines 49-56 -- `itemKind` enum pattern for extending visible item types
- Existing codebase: `internal/todolist/model.go` lines 292-402 -- `visibleItems()` assembly pattern
- Existing codebase: `internal/todolist/model.go` lines 404-413 -- `selectableIndices()` filtering pattern

### MEDIUM Confidence (multiple sources agree)
- [gcal-tui Google Calendar Integration (DeepWiki)](https://deepwiki.com/motemen/gcal-tui/4-google-calendar-integration) -- Event rendering with status markers, time range display, lipgloss styling
- [WTF Dashboard: Google Calendar module](https://wtfutil.com/modules/google/gcal/) -- Event display format, color coding, conflict indicators
- [Calcure GitHub](https://github.com/anufrievroman/calcure) -- Split layout, separate colors for events vs tasks
- [Fantastical calendar views](https://flexibits.com/fantastical/help/calendar-views) -- Unified event+task list, color pills

### LOW Confidence (needs validation during implementation)
- Optimal sort order (all-day first, then timed, then todos) -- consensus from Todoist/Things but not verified in TUI context. May feel wrong if all-day events push todos too far down.
- 5-char padding for priority badge alignment when events have no priority -- visual alignment needs testing in real terminal to confirm it looks right.
- Inline filter matching events -- logical extension of existing pattern but may cause confusion if filtered results mix events and todos without clear headers.

---
*Feature research for: Read-Only Google Calendar Event Display in Todo Calendar TUI*
*Researched: 2026-02-13*
