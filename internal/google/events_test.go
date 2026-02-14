package google

import (
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
)

func TestConvertEvent_AllDay(t *testing.T) {
	e := &calendar.Event{
		Id:      "allday1",
		Summary: "All Day Event",
		Status:  "confirmed",
		Start:   &calendar.EventDateTime{Date: "2026-02-14"},
		End:     &calendar.EventDateTime{Date: "2026-02-15"},
	}

	ce := convertEvent(e)

	if !ce.AllDay {
		t.Error("expected AllDay to be true")
	}
	if ce.Date != "2026-02-14" {
		t.Errorf("expected Date '2026-02-14', got %q", ce.Date)
	}
	if !ce.Start.IsZero() {
		t.Error("expected Start to be zero for all-day event")
	}
	if !ce.End.IsZero() {
		t.Error("expected End to be zero for all-day event")
	}
}

func TestConvertEvent_Timed(t *testing.T) {
	e := &calendar.Event{
		Id:      "timed1",
		Summary: "Timed Event",
		Status:  "confirmed",
		Start:   &calendar.EventDateTime{DateTime: "2026-02-14T10:00:00+02:00"},
		End:     &calendar.EventDateTime{DateTime: "2026-02-14T11:00:00+02:00"},
	}

	ce := convertEvent(e)

	if ce.AllDay {
		t.Error("expected AllDay to be false")
	}
	if ce.Date != "2026-02-14" {
		t.Errorf("expected Date '2026-02-14', got %q", ce.Date)
	}
	if ce.Start.IsZero() {
		t.Error("expected Start to be non-zero")
	}
	if ce.Start.Hour() != 10 {
		t.Errorf("expected Start hour 10, got %d", ce.Start.Hour())
	}
	if ce.End.IsZero() {
		t.Error("expected End to be non-zero")
	}
}

func TestConvertEvent_CopiesFields(t *testing.T) {
	e := &calendar.Event{
		Id:      "copy1",
		Summary: "Meeting with Team",
		Status:  "tentative",
		Start:   &calendar.EventDateTime{DateTime: "2026-02-14T14:00:00Z"},
		End:     &calendar.EventDateTime{DateTime: "2026-02-14T15:00:00Z"},
	}

	ce := convertEvent(e)

	if ce.ID != "copy1" {
		t.Errorf("expected ID 'copy1', got %q", ce.ID)
	}
	if ce.Summary != "Meeting with Team" {
		t.Errorf("expected Summary 'Meeting with Team', got %q", ce.Summary)
	}
	if ce.Status != "tentative" {
		t.Errorf("expected Status 'tentative', got %q", ce.Status)
	}
}

func TestMergeEvents_Upsert(t *testing.T) {
	existing := []CalendarEvent{
		{ID: "a", Summary: "Event A", Date: "2026-02-14", Status: "confirmed"},
		{ID: "b", Summary: "Event B", Date: "2026-02-15", Status: "confirmed"},
	}
	delta := []CalendarEvent{
		{ID: "b", Summary: "Updated B", Date: "2026-02-15", Status: "confirmed"},
		{ID: "c", Summary: "Event C", Date: "2026-02-16", Status: "confirmed"},
	}

	result := MergeEvents(existing, delta)

	if len(result) != 3 {
		t.Fatalf("expected 3 events, got %d", len(result))
	}

	// Find event B and verify it was updated
	found := false
	for _, e := range result {
		if e.ID == "b" {
			found = true
			if e.Summary != "Updated B" {
				t.Errorf("expected Summary 'Updated B', got %q", e.Summary)
			}
		}
	}
	if !found {
		t.Error("event B not found in result")
	}

	// Verify sorting by date
	for i := 1; i < len(result); i++ {
		if result[i].Date < result[i-1].Date {
			t.Errorf("events not sorted by date: %s < %s", result[i].Date, result[i-1].Date)
		}
	}
}

func TestMergeEvents_Cancelled(t *testing.T) {
	existing := []CalendarEvent{
		{ID: "a", Summary: "Event A", Date: "2026-02-14", Start: time.Date(2026, 2, 14, 9, 0, 0, 0, time.UTC), Status: "confirmed"},
		{ID: "b", Summary: "Event B", Date: "2026-02-15", Start: time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC), Status: "confirmed"},
	}
	delta := []CalendarEvent{
		{ID: "b", Status: "cancelled"},
	}

	result := MergeEvents(existing, delta)

	if len(result) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result))
	}
	if result[0].ID != "a" {
		t.Errorf("expected remaining event ID 'a', got %q", result[0].ID)
	}
}
