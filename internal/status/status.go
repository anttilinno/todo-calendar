package status

import (
	"encoding/json"
	"time"

	"github.com/antti/todo-calendar/internal/store"
)

// StatusResult holds pending todo counts by date granularity.
type StatusResult struct {
	Daily   int `json:"daily"`
	Monthly int `json:"monthly"`
	Yearly  int `json:"yearly"`
}

// countPending returns the number of incomplete todos in the slice.
func countPending(todos []store.Todo) int {
	var n int
	for _, td := range todos {
		if !td.Done {
			n++
		}
	}
	return n
}

// FormatStatus returns pending todo counts as a JSON object.
func FormatStatus(s store.TodoStore) string {
	now := time.Now()
	today := now.Format("2006-01-02")

	result := StatusResult{
		Daily:   countPending(s.TodosForDateRange(today, today)),
		Monthly: countPending(s.MonthTodos(now.Year(), now.Month())),
		Yearly:  countPending(s.YearTodos(now.Year())),
	}

	b, _ := json.Marshal(result)
	return string(b)
}
