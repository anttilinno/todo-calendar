package recurring

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ScheduleRule represents a recurring cadence for todo creation.
type ScheduleRule struct {
	Type       string   // "daily", "weekdays", "weekly", "monthly"
	Days       []string // for weekly: lowercase day names (e.g., "mon", "fri")
	DayOfMonth int      // for monthly: 1-31 (0 means unused)
}

// validDays maps lowercase day abbreviations to time.Weekday.
var validDays = map[string]time.Weekday{
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
	"sun": time.Sunday,
}

// ParseRule parses a cadence string into a ScheduleRule.
//
// Supported formats:
//   - "daily"
//   - "weekdays"
//   - "weekly:mon,wed,fri"
//   - "monthly:15"
func ParseRule(s string) (ScheduleRule, error) {
	if s == "" {
		return ScheduleRule{}, fmt.Errorf("empty schedule rule")
	}

	parts := strings.SplitN(s, ":", 2)
	typ := parts[0]

	switch typ {
	case "daily":
		return ScheduleRule{Type: "daily"}, nil

	case "weekdays":
		return ScheduleRule{Type: "weekdays"}, nil

	case "weekly":
		if len(parts) < 2 || parts[1] == "" {
			return ScheduleRule{}, fmt.Errorf("weekly rule requires day list (e.g., weekly:mon,fri)")
		}
		dayNames := strings.Split(parts[1], ",")
		for _, d := range dayNames {
			if _, ok := validDays[d]; !ok {
				return ScheduleRule{}, fmt.Errorf("invalid day name %q in weekly rule", d)
			}
		}
		return ScheduleRule{Type: "weekly", Days: dayNames}, nil

	case "monthly":
		if len(parts) < 2 || parts[1] == "" {
			return ScheduleRule{}, fmt.Errorf("monthly rule requires day number (e.g., monthly:15)")
		}
		day, err := strconv.Atoi(parts[1])
		if err != nil {
			return ScheduleRule{}, fmt.Errorf("invalid day number %q in monthly rule: %w", parts[1], err)
		}
		if day < 1 || day > 31 {
			return ScheduleRule{}, fmt.Errorf("monthly day must be 1-31, got %d", day)
		}
		return ScheduleRule{Type: "monthly", DayOfMonth: day}, nil

	default:
		return ScheduleRule{}, fmt.Errorf("unknown schedule type %q", typ)
	}
}

// MatchesDate returns true if the given date matches this rule's cadence.
func (r ScheduleRule) MatchesDate(d time.Time) bool {
	switch r.Type {
	case "daily":
		return true

	case "weekdays":
		wd := d.Weekday()
		return wd >= time.Monday && wd <= time.Friday

	case "weekly":
		wd := d.Weekday()
		for _, dayName := range r.Days {
			if validDays[dayName] == wd {
				return true
			}
		}
		return false

	case "monthly":
		// Get last day of the month for clamping.
		lastDay := lastDayOfMonth(d.Year(), d.Month())
		target := r.DayOfMonth
		if target > lastDay {
			target = lastDay
		}
		return d.Day() == target

	default:
		return false
	}
}

// String returns the canonical string form of the rule.
// ParseRule(r.String()) produces an equivalent rule.
func (r ScheduleRule) String() string {
	switch r.Type {
	case "daily":
		return "daily"
	case "weekdays":
		return "weekdays"
	case "weekly":
		return "weekly:" + strings.Join(r.Days, ",")
	case "monthly":
		return "monthly:" + strconv.Itoa(r.DayOfMonth)
	default:
		return ""
	}
}

// lastDayOfMonth returns the last day of the given month.
func lastDayOfMonth(year int, month time.Month) int {
	// Day 0 of next month is the last day of this month.
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
