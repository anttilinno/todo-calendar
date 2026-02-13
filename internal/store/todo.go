package store

import (
	"fmt"
	"time"
)

// dateFormat is the canonical date layout for todo dates (YYYY-MM-DD).
const dateFormat = "2006-01-02"

// Todo represents a single todo item.
// Date is stored as a plain string ("YYYY-MM-DD") to avoid timezone
// corruption during JSON round-trips.
type Todo struct {
	ID            int    `json:"id"`
	Text          string `json:"text"`
	Body          string `json:"body,omitempty"`
	Date          string `json:"date,omitempty"`
	Done          bool   `json:"done"`
	CreatedAt     string `json:"created_at"`
	SortOrder     int    `json:"sort_order,omitempty"`
	ScheduleID    int    `json:"schedule_id,omitempty"`
	ScheduleDate  string `json:"schedule_date,omitempty"`
	DatePrecision string `json:"date_precision"`
	Priority      int    `json:"priority"`
}

// HasPriority reports whether the todo has a valid priority level (1-4).
func (t Todo) HasPriority() bool {
	return t.Priority >= 1 && t.Priority <= 4
}

// PriorityLabel returns a short label like "P1" for priority 1-4, or "" otherwise.
func (t Todo) PriorityLabel() string {
	if t.Priority >= 1 && t.Priority <= 4 {
		return fmt.Sprintf("P%d", t.Priority)
	}
	return ""
}

// HasBody reports whether the todo has a non-empty markdown body.
func (t Todo) HasBody() bool {
	return t.Body != ""
}

// Template represents a reusable markdown template with placeholders.
type Template struct {
	ID        int
	Name      string
	Content   string
	CreatedAt string
}

// Schedule represents a recurring schedule linked to a template.
type Schedule struct {
	ID                 int
	TemplateID         int
	CadenceType        string
	CadenceValue       string
	PlaceholderDefaults string // JSON object of default placeholder values
	CreatedAt          string
}

// IsMonthPrecision reports whether this todo has month-level date precision.
func (t Todo) IsMonthPrecision() bool {
	return t.DatePrecision == "month"
}

// IsYearPrecision reports whether this todo has year-level date precision.
func (t Todo) IsYearPrecision() bool {
	return t.DatePrecision == "year"
}

// IsFuzzy reports whether this todo has a fuzzy (non-day) date precision.
func (t Todo) IsFuzzy() bool {
	return t.DatePrecision == "month" || t.DatePrecision == "year"
}

// HasDate reports whether the todo has a date assigned.
func (t Todo) HasDate() bool {
	return t.Date != ""
}

// InMonth reports whether the todo's date falls in the given year and month.
// Month-precision todos match if year and month match.
// Year-precision todos match if the year matches.
// Returns false if the todo has no date or the date cannot be parsed.
func (t Todo) InMonth(year int, month time.Month) bool {
	if t.Date == "" {
		return false
	}
	d, err := time.Parse(dateFormat, t.Date)
	if err != nil {
		return false
	}
	y, m, _ := d.Date()
	switch t.DatePrecision {
	case "year":
		return y == year
	case "month":
		return y == year && m == month
	default:
		return y == year && m == month
	}
}

// InDateRange reports whether the todo's date falls within [startDate, endDate] inclusive.
// Fuzzy-date todos (month/year precision) are excluded from date range matching.
// Returns false if the todo has no date or any date cannot be parsed.
func (t Todo) InDateRange(startDate, endDate string) bool {
	if t.Date == "" {
		return false
	}
	if t.IsFuzzy() {
		return false
	}
	d, err := time.Parse(dateFormat, t.Date)
	if err != nil {
		return false
	}
	s, err := time.Parse(dateFormat, startDate)
	if err != nil {
		return false
	}
	e, err := time.Parse(dateFormat, endDate)
	if err != nil {
		return false
	}
	return !d.Before(s) && !d.After(e)
}
