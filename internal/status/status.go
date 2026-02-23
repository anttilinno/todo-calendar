package status

import (
	"fmt"

	"github.com/antti/todo-calendar/internal/store"
)

// FormatStatus returns the count of pending todos as a string.
// Returns an empty string when no pending todos exist.
func FormatStatus(todos []store.Todo) string {
	var count int
	for _, td := range todos {
		if !td.Done {
			count++
		}
	}
	if count == 0 {
		return ""
	}
	return fmt.Sprintf("%d", count)
}
