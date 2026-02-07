package recurring

import (
	"encoding/json"
	"time"

	"github.com/antti/todo-calendar/internal/store"
	"github.com/antti/todo-calendar/internal/tmpl"
)

const dateFormat = "2006-01-02"

// windowDays is the number of days to look ahead (today + 6).
const windowDays = 7

// AutoCreate generates scheduled todos for the next 7 days (today through
// today+6). It iterates all schedules, checks cadence matching, deduplicates,
// and fills template placeholders from stored defaults.
func AutoCreate(s store.TodoStore) {
	AutoCreateForDate(s, time.Now())
}

// AutoCreateForDate is the testable core of AutoCreate, accepting an explicit
// "today" time so tests can pin the date.
func AutoCreateForDate(s store.TodoStore, today time.Time) {
	schedules := s.ListSchedules()
	for _, sched := range schedules {
		ruleStr := buildRuleString(sched.CadenceType, sched.CadenceValue)
		rule, err := ParseRule(ruleStr)
		if err != nil {
			// Bad cadence -- skip silently.
			continue
		}

		tpl := s.FindTemplate(sched.TemplateID)
		if tpl == nil {
			// Orphan schedule -- template deleted. Skip.
			continue
		}

		defaults := parseDefaults(sched.PlaceholderDefaults)
		body, err := tmpl.ExecuteTemplate(tpl.Content, defaults)
		if err != nil {
			// Template execution failed -- skip.
			continue
		}

		for i := 0; i < windowDays; i++ {
			d := today.AddDate(0, 0, i)
			if !rule.MatchesDate(d) {
				continue
			}
			dateStr := d.Format(dateFormat)
			if s.TodoExistsForSchedule(sched.ID, dateStr) {
				continue
			}
			s.AddScheduledTodo(tpl.Name, dateStr, body, sched.ID)
		}
	}
}

// buildRuleString concatenates cadence type and value into the rule string
// format expected by ParseRule.
//
//	"daily" + "" -> "daily"
//	"weekly" + "mon,fri" -> "weekly:mon,fri"
//	"monthly" + "15" -> "monthly:15"
func buildRuleString(cadenceType, cadenceValue string) string {
	if cadenceValue == "" {
		return cadenceType
	}
	return cadenceType + ":" + cadenceValue
}

// parseDefaults parses the JSON placeholder defaults string into a map.
// Returns an empty map on any parse error.
func parseDefaults(raw string) map[string]string {
	if raw == "" {
		return make(map[string]string)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return make(map[string]string)
	}
	return m
}
