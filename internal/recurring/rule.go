package recurring

import (
	"time"
)

// ScheduleRule represents a recurring cadence for todo creation.
type ScheduleRule struct {
	Type       string   // "daily", "weekdays", "weekly", "monthly"
	Days       []string // for weekly: lowercase day names (e.g., "mon", "fri")
	DayOfMonth int      // for monthly: 1-31 (0 means unused)
}

// ParseRule parses a cadence string into a ScheduleRule.
func ParseRule(_ string) (ScheduleRule, error) {
	return ScheduleRule{}, nil
}

// MatchesDate returns true if the given date matches this rule's cadence.
func (r ScheduleRule) MatchesDate(_ time.Time) bool {
	return false
}

// String returns the canonical string form of the rule.
func (r ScheduleRule) String() string {
	return ""
}
