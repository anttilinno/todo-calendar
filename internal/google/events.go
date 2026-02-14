package google

import (
	"context"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// CalendarEvent is the app's representation of a Google Calendar event,
// decoupled from the Google API types.
type CalendarEvent struct {
	ID      string
	Summary string
	Date    string    // "2006-01-02" for ALL events (used as lookup key)
	EndDate string    // "2006-01-02" end date for all-day events (exclusive, Google convention)
	Start   time.Time // parsed RFC3339 for timed events; zero for all-day
	End     time.Time // parsed RFC3339 for timed events; zero for all-day
	AllDay  bool
	Status  string // "confirmed", "tentative", "cancelled"
}

// NewCalendarService creates a Google Calendar API service using the
// persisted OAuth token from auth.go.
func NewCalendarService() (*calendar.Service, error) {
	ts, err := TokenSource()
	if err != nil {
		return nil, err
	}
	return calendar.NewService(context.Background(), option.WithTokenSource(ts))
}

// FetchEvents fetches calendar events using the Google Calendar API.
// If syncToken is empty, a full sync is performed (past 1 month to future 3 months).
// If syncToken is non-empty, a delta sync is performed using the token.
// Returns the fetched events, the new sync token, and any error.
func FetchEvents(srv *calendar.Service, syncToken string) ([]CalendarEvent, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var allEvents []CalendarEvent
	var nextSyncToken string
	pageToken := ""

	for {
		call := srv.Events.List("primary").
			Context(ctx).
			SingleEvents(true).
			ShowDeleted(true).
			MaxResults(2500)

		if syncToken == "" {
			now := time.Now()
			call = call.
				TimeMin(now.AddDate(0, -1, 0).Format(time.RFC3339)).
				TimeMax(now.AddDate(0, 3, 0).Format(time.RFC3339)).
				OrderBy("startTime")
		} else {
			call = call.SyncToken(syncToken)
		}

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 410 {
				// 410 GONE: sync token expired, retry with full sync
				return FetchEvents(srv, "")
			}
			return nil, "", err
		}

		for _, e := range resp.Items {
			allEvents = append(allEvents, convertEvent(e))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			nextSyncToken = resp.NextSyncToken
			break
		}
	}

	return allEvents, nextSyncToken, nil
}

// convertEvent converts a Google Calendar API event to our CalendarEvent type.
func convertEvent(e *calendar.Event) CalendarEvent {
	ce := CalendarEvent{
		ID:      e.Id,
		Summary: e.Summary,
		Status:  e.Status,
	}

	if e.Start == nil {
		return ce
	}

	if e.Start.Date != "" {
		// All-day event: store date as raw string, no timezone conversion
		ce.AllDay = true
		ce.Date = e.Start.Date
		if e.End != nil && e.End.Date != "" {
			ce.EndDate = e.End.Date
		}
	} else if e.Start.DateTime != "" {
		// Timed event: parse RFC3339 and derive date
		t, err := time.Parse(time.RFC3339, e.Start.DateTime)
		if err == nil {
			ce.Start = t
			ce.Date = t.Format("2006-01-02")
		}
		if e.End != nil && e.End.DateTime != "" {
			t2, err := time.Parse(time.RFC3339, e.End.DateTime)
			if err == nil {
				ce.End = t2
			}
		}
	}

	return ce
}

// ExpandMultiDay expands multi-day all-day events into per-day entries.
// Each expanded entry gets its Date set to the corresponding day.
// EndDate in Google is exclusive (a 2-day event on Jan 1 has EndDate Jan 3).
func ExpandMultiDay(events []CalendarEvent) []CalendarEvent {
	const layout = "2006-01-02"
	var result []CalendarEvent

	for _, ev := range events {
		if !ev.AllDay || ev.EndDate == "" {
			result = append(result, ev)
			continue
		}

		start, err := time.Parse(layout, ev.Date)
		if err != nil {
			result = append(result, ev)
			continue
		}
		end, err := time.Parse(layout, ev.EndDate)
		if err != nil {
			result = append(result, ev)
			continue
		}

		// Single-day all-day event: EndDate is the day after Date
		if end.Equal(start.AddDate(0, 0, 1)) {
			result = append(result, ev)
			continue
		}

		// Multi-day: create a copy for each day (end is exclusive)
		for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
			expanded := ev
			expanded.Date = d.Format(layout)
			result = append(result, expanded)
		}
	}

	return result
}

// --- Bubble Tea message types and commands ---

// EventsFetchedMsg is sent when event fetching completes.
type EventsFetchedMsg struct {
	Events    []CalendarEvent
	SyncToken string
	Err       error
}

// EventTickMsg is sent periodically to trigger event re-fetching.
type EventTickMsg time.Time

// FetchEventsCmd returns a tea.Cmd that fetches events in a goroutine
// and returns an EventsFetchedMsg.
func FetchEventsCmd(srv *calendar.Service, syncToken string) tea.Cmd {
	return func() tea.Msg {
		events, newToken, err := FetchEvents(srv, syncToken)
		return EventsFetchedMsg{
			Events:    events,
			SyncToken: newToken,
			Err:       err,
		}
	}
}

// ScheduleEventTick returns a tea.Cmd that sends an EventTickMsg after 5 minutes.
func ScheduleEventTick() tea.Cmd {
	return tea.Tick(5*time.Minute, func(t time.Time) tea.Msg {
		return EventTickMsg(t)
	})
}

// --- Event merging ---

// MergeEvents merges delta events into an existing event list.
// Cancelled events are removed, others are upserted by ID.
// The result is sorted by Date, then Start time.
func MergeEvents(existing []CalendarEvent, delta []CalendarEvent) []CalendarEvent {
	m := make(map[string]CalendarEvent, len(existing))
	for _, e := range existing {
		m[e.ID] = e
	}

	for _, e := range delta {
		if e.Status == "cancelled" {
			delete(m, e.ID)
		} else {
			m[e.ID] = e
		}
	}

	result := make([]CalendarEvent, 0, len(m))
	for _, e := range m {
		result = append(result, e)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Date != result[j].Date {
			return result[i].Date < result[j].Date
		}
		// All-day events come before timed events on the same date
		if result[i].AllDay != result[j].AllDay {
			return result[i].AllDay
		}
		return result[i].Start.Before(result[j].Start)
	})

	return result
}
