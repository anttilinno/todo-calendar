package store

import "time"

// TodoStore defines the contract for todo persistence.
// Consumers depend on this interface, not the concrete backend.
type TodoStore interface {
	Add(text string, date string) Todo
	Toggle(id int)
	Delete(id int)
	Find(id int) *Todo
	Update(id int, text string, date string)
	Todos() []Todo
	TodosForMonth(year int, month time.Month) []Todo
	TodosForDateRange(startDate, endDate string) []Todo
	FloatingTodos() []Todo
	IncompleteTodosPerDay(year int, month time.Month) map[int]int
	TotalTodosPerDay(year int, month time.Month) map[int]int
	TodoCountsByMonth() []MonthCount
	FloatingTodoCounts() FloatingCount
	UpdateBody(id int, body string)
	AddTemplate(name, content string) (Template, error)
	ListTemplates() []Template
	FindTemplate(id int) *Template
	DeleteTemplate(id int)
	UpdateTemplate(id int, name, content string) error
	// Schedule operations
	AddSchedule(templateID int, cadenceType, cadenceValue, placeholderDefaults string) (Schedule, error)
	ListSchedules() []Schedule
	ListSchedulesForTemplate(templateID int) []Schedule
	DeleteSchedule(id int)
	UpdateSchedule(id int, cadenceType, cadenceValue, placeholderDefaults string) error
	TodoExistsForSchedule(scheduleID int, date string) bool
	AddScheduledTodo(text, date, body string, scheduleID int) Todo
	SwapOrder(id1, id2 int)
	SearchTodos(query string) []Todo
	EnsureSortOrder()
	Save() error
}

// MonthCount holds pending and completed todo counts for a given year and month.
type MonthCount struct {
	Year      int
	Month     time.Month
	Pending   int
	Completed int
}

// FloatingCount holds pending and completed counts for undated todos.
type FloatingCount struct {
	Pending   int
	Completed int
}
