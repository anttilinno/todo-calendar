package store

import "time"

// dateFormat is the canonical date layout for todo dates (YYYY-MM-DD).
const dateFormat = "2006-01-02"

// Todo represents a single todo item.
// Date is stored as a plain string ("YYYY-MM-DD") to avoid timezone
// corruption during JSON round-trips.
type Todo struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Date      string `json:"date,omitempty"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	SortOrder int    `json:"sort_order,omitempty"`
}

// Data is the top-level JSON envelope persisted to disk.
type Data struct {
	NextID int    `json:"next_id"`
	Todos  []Todo `json:"todos"`
}

// HasDate reports whether the todo has a date assigned.
func (t Todo) HasDate() bool {
	return t.Date != ""
}

// InMonth reports whether the todo's date falls in the given year and month.
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
	return y == year && m == month
}
