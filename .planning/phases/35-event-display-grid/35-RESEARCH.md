# Phase 35: Event Display & Grid - Research

**Researched:** 2026-02-14
**Domain:** Bubble Tea TUI rendering, Google Calendar event display, lipgloss styling
**Confidence:** HIGH

## Summary

This phase integrates Google Calendar events into the existing todo list panel and calendar grid. The codebase already has full event fetching infrastructure (Phase 34) with `CalendarEvent` types stored in the app model. The work is purely UI/rendering: adding event items to the todolist's `visibleItems()`, rendering them distinctly, making them non-selectable, expanding multi-day events, and adding event presence indicators to the calendar grid.

The architecture is well-established. The todolist uses a `visibleItem` struct with `itemKind` discriminator (`headerItem`, `todoItem`, `emptyItem`). Adding an `eventItem` kind follows the exact same pattern. The calendar grid uses `indicators` and `totals` maps to determine day cell styling -- events just need to contribute to a similar presence signal.

**Primary recommendation:** Add a new `eventItem` kind to the todolist's `visibleItem` system, pass calendar events through the component hierarchy, and extend the grid's indicator logic to include event presence.

## Standard Stack

### Core (already in use)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charmbracelet/bubbletea | current | TUI framework | Project foundation |
| charmbracelet/lipgloss | current | Styling | All UI styling uses this |
| charmbracelet/bubbles | current | Input widgets | textinput, textarea, help |

No new dependencies are needed. This phase is purely rendering changes within existing architecture.

## Architecture Patterns

### Pattern 1: visibleItem Kind Extension

The todolist already has a well-defined item kind system:

```go
// Existing in internal/todolist/model.go
type itemKind int

const (
    headerItem itemKind = iota
    todoItem
    emptyItem
)

type visibleItem struct {
    kind    itemKind
    label   string
    todo    *store.Todo
    section sectionID
}
```

**Add `eventItem` kind:**

```go
const (
    headerItem itemKind = iota
    todoItem
    emptyItem
    eventItem  // NEW: Google Calendar event
)

type visibleItem struct {
    kind    itemKind
    label   string
    todo    *store.Todo
    event   *google.CalendarEvent  // NEW: non-nil only for eventItem
    section sectionID
}
```

**Key insight:** Events are NOT selectable. The existing `selectableIndices()` function only includes `todoItem` -- `eventItem` items are automatically skipped by the cursor, satisfying DISP-05 without extra logic.

### Pattern 2: Event Data Flow

Events flow from app model to child components:

```
App.calendarEvents []google.CalendarEvent
    |
    +---> todolist.Model.SetCalendarEvents(events)  // setter, stores reference
    +---> calendar.Model.SetCalendarEvents(events)  // setter for grid indicators
```

The app model calls these setters after `EventsFetchedMsg` processing and during `syncTodoView()`. This follows the existing pattern where `SetViewMonth()`, `SetWeekFilter()`, and `SetShowFuzzySections()` pass data from app to children.

### Pattern 3: Event Insertion in visibleItems()

Events are inserted into the dated section, BEFORE todos (DISP-04). The logic mirrors the existing dated section:

```go
// In visibleItems(), within the dated section, after the header:
// 1. Filter events to match the current view (month or week)
// 2. Insert eventItem entries for matching events
// 3. Then insert todoItem entries (existing code)
```

For multi-day events (DISP-06), events are expanded at display time: an event spanning Feb 14-16 produces three `eventItem` entries, one for each date. This expansion happens in `visibleItems()` or a helper called by it.

### Pattern 4: Event Rendering in normalView

```go
// New function alongside renderTodo:
func (m Model) renderEvent(b *strings.Builder, e *google.CalendarEvent) {
    b.WriteString("  ")     // No cursor indicator (not selectable)
    b.WriteString("     ")  // No priority badge slot (5 spaces)
    // No checkbox -- this is the key visual distinction

    // Time prefix
    if e.AllDay {
        b.WriteString(m.styles.EventTime.Render("all day"))
    } else {
        timeStr := e.Start.Format("15:04")
        b.WriteString(m.styles.EventTime.Render(timeStr))
    }
    b.WriteString(" ")
    b.WriteString(m.styles.EventText.Render(e.Summary))
    b.WriteString("\n")
}
```

### Pattern 5: Calendar Grid Event Indicators (GRID-01)

The grid currently uses `indicators` (incomplete todos per day) and `totals` (all todos per day) to decide bracket rendering `[14]` vs ` 14 `. Events should contribute to a separate presence signal:

```go
// In calendar Model:
hasEvents  map[int]bool  // day -> has calendar events

// In grid rendering, a day gets brackets if:
// hasPending || hasAllDone || hasEvents[day]
```

This ensures days with only events (no todos) still show `[14]` indicator brackets. The event indicator could use a distinct style (e.g., the event foreground color) when no todos are present.

### Pattern 6: Settings Toggle (CONF-01)

The settings overlay already has a Google Calendar row (row 6). The toggle adds an enable/disable control that:
- Persists to config as `google_calendar_enabled bool`
- When disabled: events are not displayed (but credentials remain)
- When enabled: events display normally

This requires adding a field to `config.Config` and a new settings row (or modifying the existing Google Calendar row to support cycling when auth state is `AuthReady`).

### Recommended Changes by File

```
internal/todolist/model.go
  - Add eventItem kind
  - Add event field to visibleItem
  - Add calendarEvents field to Model
  - Add SetCalendarEvents() setter
  - Modify visibleItems() to insert events in dated section
  - Add renderEvent() method

internal/todolist/styles.go
  - Add EventTime, EventText styles

internal/theme/theme.go
  - Add EventFg color to Theme struct
  - Add event color to all 4 themes

internal/calendar/model.go
  - Add calendarEvents field
  - Add SetCalendarEvents() setter
  - Add hasEventsPerDay() helper
  - Pass event presence to RenderGrid/RenderWeekGrid

internal/calendar/grid.go
  - Accept hasEvents map parameter
  - Use hasEvents in bracket/indicator logic

internal/google/events.go
  - Add EndDate string field to CalendarEvent for all-day events
  - Store e.End.Date in convertEvent for multi-day expansion
  - Add ExpandMultiDay() helper function

internal/app/model.go
  - Pass events to todolist and calendar after fetch
  - Call setters in syncTodoView() and after EventsFetchedMsg

internal/config/config.go
  - Add GoogleCalendarEnabled bool field

internal/settings/model.go
  - Add enable/disable toggle for Google Calendar (when connected)
```

### Anti-Patterns to Avoid
- **Filtering events in the app model:** Let child components handle their own view filtering. The app passes ALL events; todolist and calendar filter by their current view range.
- **Making events selectable:** Events must be entirely non-interactive in the todo list. No checkbox, no cursor stop, no edit/delete.
- **Storing expanded multi-day events:** Expand at render time only. The canonical event list stays as-is from the API.
- **Adding event counts to the overview section:** The overview shows todo counts only. Events are a separate concern.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Time formatting | Custom time parser | `time.Time.Format("15:04")` | Already have parsed `Start` time |
| Multi-day date iteration | Manual date arithmetic | `time.AddDate(0, 0, 1)` loop from start to end | Handles month boundaries correctly |
| Event-todo sorting | Custom sort | Insert events before todos in `visibleItems()` | MergeEvents already sorts all-day before timed |

## Common Pitfalls

### Pitfall 1: Multi-Day All-Day Event EndDate
**What goes wrong:** Google Calendar API uses exclusive end dates for all-day events. A 1-day event on Feb 14 has Start.Date="2026-02-14" and End.Date="2026-02-15". A 3-day event Feb 14-16 has End.Date="2026-02-17".
**Why it happens:** Off-by-one from exclusive end date convention.
**How to avoid:** When expanding multi-day events, iterate from Start.Date to End.Date EXCLUSIVE (i.e., `date < endDate`, not `date <= endDate`).
**Warning signs:** Events appearing on one extra day at the end.

### Pitfall 2: CalendarEvent.Date Only Stores Start Date
**What goes wrong:** The current `convertEvent()` for all-day events stores only `ce.Date = e.Start.Date` and leaves `ce.End` as zero time. Multi-day all-day events cannot be expanded without the end date.
**How to avoid:** Add an `EndDate string` field to `CalendarEvent` and populate it from `e.End.Date` in `convertEvent()`. This is needed for DISP-06.
**Warning signs:** Multi-day all-day events only showing on their first day.

### Pitfall 3: Cursor Index Mismatch After Event Insertion
**What goes wrong:** Events are inserted into `visibleItems()` but are not selectable. If cursor management doesn't account for this, cursor positions could be off.
**Why it happens:** `selectableIndices()` already only returns indices of `todoItem` items, so this should work automatically.
**How to avoid:** Verify that `selectableIndices()` continues to work correctly -- it checks `item.kind == todoItem`, so `eventItem` is naturally excluded.

### Pitfall 4: Weekly View Event Filtering
**What goes wrong:** Events span the week boundary but should only show within the filtered week range.
**Why it happens:** Multi-day events that start before or end after the week.
**How to avoid:** When inserting events into `visibleItems()`, filter by expanded date (each day's entry), not by the event's original date range. A multi-day event that starts before the week should still show its Wednesday entry if Wednesday is in the week.

### Pitfall 5: Event Style Must Work Across All 4 Themes
**What goes wrong:** Event color only tested in dark theme, looks bad in light/nord/solarized.
**How to avoid:** Add `EventFg` to all 4 theme definitions. Use a color that contrasts with both light and dark backgrounds. Cyan/teal family works well (distinct from todo accent/muted colors).

### Pitfall 6: Config Toggle Default
**What goes wrong:** Existing users who have already connected Google Calendar lose events after upgrade because new `GoogleCalendarEnabled` defaults to `false`.
**How to avoid:** Default `GoogleCalendarEnabled` to `true`. Users who haven't connected Google Calendar won't have events anyway. The toggle only matters for users who want to temporarily hide events.

## Code Examples

### Event Item Rendering (for todolist normalView)

```go
case eventItem:
    m.renderEvent(&b, item.event)
```

```go
func (m Model) renderEvent(b *strings.Builder, e *google.CalendarEvent) {
    // Indentation: 2 spaces (no cursor) + 5 spaces (no priority badge)
    b.WriteString("       ")

    // Time prefix (DISP-01, DISP-02)
    if e.AllDay {
        b.WriteString(m.styles.EventTime.Render("all day"))
    } else {
        b.WriteString(m.styles.EventTime.Render(e.Start.Format("15:04")))
    }
    b.WriteString("  ")

    // Event summary (DISP-03: distinct color, no checkbox)
    b.WriteString(m.styles.EventText.Render(e.Summary))
    b.WriteString("\n")
}
```

### Multi-Day Event Expansion

```go
// ExpandMultiDay expands multi-day all-day events into per-day entries.
// Each entry has Date set to the specific day it represents.
// Timed events and single-day events pass through unchanged.
func ExpandMultiDay(events []CalendarEvent) []CalendarEvent {
    var result []CalendarEvent
    for _, e := range events {
        if !e.AllDay || e.EndDate == "" || e.EndDate == "" {
            result = append(result, e)
            continue
        }
        start, err1 := time.Parse("2006-01-02", e.Date)
        end, err2 := time.Parse("2006-01-02", e.EndDate)
        if err1 != nil || err2 != nil {
            result = append(result, e)
            continue
        }
        // End date is EXCLUSIVE in Google Calendar API
        for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
            expanded := e
            expanded.Date = d.Format("2006-01-02")
            result = append(result, expanded)
        }
    }
    return result
}
```

### Event Filtering by Date Range

```go
// eventsForDateRange returns events whose Date falls within [start, end].
func eventsForDateRange(events []CalendarEvent, start, end string) []CalendarEvent {
    var result []CalendarEvent
    for i := range events {
        if events[i].Date >= start && events[i].Date <= end {
            result = append(result, events[i])
        }
    }
    return result
}
```

### Settings Toggle (CONF-01)

```go
// In settings, modify the Google Calendar row when AuthReady:
// Instead of static "Connected" text, show "< Enabled >" or "< Disabled >"
// that cycles with left/right arrows.
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No events in UI | Events fetched but not displayed | Phase 34 (current) | This phase adds display |
| Single Date field | Need EndDate for multi-day | Phase 35 (this phase) | Must extend CalendarEvent |

## Open Questions

1. **Event time format: 24h vs 12h?**
   - What we know: The app uses `time.Format("15:04")` which is 24-hour format.
   - What's unclear: Should this follow the user's date format preference? (ISO/EU = 24h, US = 12h with AM/PM)
   - Recommendation: Start with 24h always (simpler, consistent with ISO). Can add format option later. The requirement says "HH:MM" which implies 24h.

2. **Event indicator color in calendar grid?**
   - What we know: Todo indicators use priority-based coloring. Events have no priority concept.
   - What's unclear: Should event-only days use a specific indicator style, or just the default bracket style?
   - Recommendation: Use default bracket style with IndicatorFg. Events are "something happening" -- same visual weight as unprioritized todos.

3. **Timed events spanning midnight?**
   - What we know: A timed event from 23:00-01:00 has `Start` and `End` on different dates. `convertEvent` sets `Date` from `Start.DateTime`.
   - What's unclear: Should it also appear on the end date?
   - Recommendation: No. Use start date only for timed events. Multi-day expansion is only for all-day events (DISP-06 says "multi-day events" but Google Calendar treats midnight-spanning timed events differently from multi-day all-day events).

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `internal/todolist/model.go` -- visibleItem system, selectableIndices, renderTodo
- Codebase analysis: `internal/calendar/grid.go` -- RenderGrid indicator logic
- Codebase analysis: `internal/google/events.go` -- CalendarEvent type, MergeEvents
- Codebase analysis: `internal/app/model.go` -- event fetching, CalendarEvents accessor
- Codebase analysis: `internal/theme/theme.go` -- 4-theme color system
- Codebase analysis: `internal/settings/model.go` -- existing Google Calendar row
- Codebase analysis: `internal/config/config.go` -- Config struct pattern

### Secondary (MEDIUM confidence)
- Google Calendar API: all-day events use exclusive end dates (verified in test: 1-day event Feb 14 has End.Date="2026-02-15")

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all existing patterns
- Architecture: HIGH -- clear extension points in visibleItem, grid rendering, and settings
- Pitfalls: HIGH -- identified from concrete code analysis (EndDate gap, exclusive dates, cursor behavior)
- Multi-day events: HIGH -- verified exclusive end-date behavior from test code

**Research date:** 2026-02-14
**Valid until:** 2026-03-14 (stable codebase, no external dependency changes)
