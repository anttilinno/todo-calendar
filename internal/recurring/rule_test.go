package recurring

import (
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// --- ParseRule tests ---

func TestParseRuleDaily(t *testing.T) {
	r, err := ParseRule("daily")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "daily" {
		t.Errorf("expected Type=daily, got %s", r.Type)
	}
}

func TestParseRuleWeekdays(t *testing.T) {
	r, err := ParseRule("weekdays")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "weekdays" {
		t.Errorf("expected Type=weekdays, got %s", r.Type)
	}
}

func TestParseRuleWeekly(t *testing.T) {
	r, err := ParseRule("weekly:mon,fri")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "weekly" {
		t.Errorf("expected Type=weekly, got %s", r.Type)
	}
	if len(r.Days) != 2 || r.Days[0] != "mon" || r.Days[1] != "fri" {
		t.Errorf("expected Days=[mon fri], got %v", r.Days)
	}
}

func TestParseRuleWeeklySingle(t *testing.T) {
	r, err := ParseRule("weekly:tue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "weekly" {
		t.Errorf("expected Type=weekly, got %s", r.Type)
	}
	if len(r.Days) != 1 || r.Days[0] != "tue" {
		t.Errorf("expected Days=[tue], got %v", r.Days)
	}
}

func TestParseRuleMonthly(t *testing.T) {
	r, err := ParseRule("monthly:15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "monthly" {
		t.Errorf("expected Type=monthly, got %s", r.Type)
	}
	if r.DayOfMonth != 15 {
		t.Errorf("expected DayOfMonth=15, got %d", r.DayOfMonth)
	}
}

func TestParseRuleMonthly31(t *testing.T) {
	r, err := ParseRule("monthly:31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.DayOfMonth != 31 {
		t.Errorf("expected DayOfMonth=31, got %d", r.DayOfMonth)
	}
}

// --- ParseRule error cases ---

func TestParseRuleEmpty(t *testing.T) {
	_, err := ParseRule("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseRuleUnknown(t *testing.T) {
	_, err := ParseRule("unknown")
	if err == nil {
		t.Fatal("expected error for unknown cadence")
	}
}

func TestParseRuleWeeklyNoDays(t *testing.T) {
	_, err := ParseRule("weekly:")
	if err == nil {
		t.Fatal("expected error for weekly with no days")
	}
}

func TestParseRuleWeeklyInvalidDay(t *testing.T) {
	_, err := ParseRule("weekly:xyz")
	if err == nil {
		t.Fatal("expected error for invalid day name")
	}
}

func TestParseRuleMonthlyZero(t *testing.T) {
	_, err := ParseRule("monthly:0")
	if err == nil {
		t.Fatal("expected error for monthly:0")
	}
}

func TestParseRuleMonthly32(t *testing.T) {
	_, err := ParseRule("monthly:32")
	if err == nil {
		t.Fatal("expected error for monthly:32")
	}
}

func TestParseRuleMonthlyNoDay(t *testing.T) {
	_, err := ParseRule("monthly:")
	if err == nil {
		t.Fatal("expected error for monthly with no day")
	}
}

func TestParseRuleMonthlyNotNumber(t *testing.T) {
	_, err := ParseRule("monthly:abc")
	if err == nil {
		t.Fatal("expected error for monthly:abc")
	}
}

// --- MatchesDate tests ---

func TestMatchesDailyAlwaysTrue(t *testing.T) {
	r, _ := ParseRule("daily")
	dates := []time.Time{
		date(2026, time.January, 1),
		date(2026, time.February, 14),
		date(2026, time.December, 31),
		date(2026, time.March, 7), // Saturday
	}
	for _, d := range dates {
		if !r.MatchesDate(d) {
			t.Errorf("daily should match %s", d.Format("2006-01-02"))
		}
	}
}

func TestMatchesWeekdaysMonToFri(t *testing.T) {
	r, _ := ParseRule("weekdays")
	// 2026-02-02 is Monday, 2026-02-06 is Friday
	for d := 2; d <= 6; d++ {
		dt := date(2026, time.February, d)
		if !r.MatchesDate(dt) {
			t.Errorf("weekdays should match %s (%s)", dt.Format("2006-01-02"), dt.Weekday())
		}
	}
}

func TestMatchesWeekdaysNotWeekend(t *testing.T) {
	r, _ := ParseRule("weekdays")
	// 2026-02-07 is Saturday, 2026-02-08 is Sunday
	sat := date(2026, time.February, 7)
	sun := date(2026, time.February, 8)
	if r.MatchesDate(sat) {
		t.Errorf("weekdays should not match Saturday")
	}
	if r.MatchesDate(sun) {
		t.Errorf("weekdays should not match Sunday")
	}
}

func TestMatchesWeeklyMonFri(t *testing.T) {
	r, _ := ParseRule("weekly:mon,fri")
	mon := date(2026, time.February, 2)  // Monday
	fri := date(2026, time.February, 6)  // Friday
	wed := date(2026, time.February, 4)  // Wednesday
	sat := date(2026, time.February, 7)  // Saturday

	if !r.MatchesDate(mon) {
		t.Error("weekly:mon,fri should match Monday")
	}
	if !r.MatchesDate(fri) {
		t.Error("weekly:mon,fri should match Friday")
	}
	if r.MatchesDate(wed) {
		t.Error("weekly:mon,fri should not match Wednesday")
	}
	if r.MatchesDate(sat) {
		t.Error("weekly:mon,fri should not match Saturday")
	}
}

func TestMatchesWeeklyTue(t *testing.T) {
	r, _ := ParseRule("weekly:tue")
	tue := date(2026, time.February, 3)  // Tuesday
	wed := date(2026, time.February, 4)  // Wednesday
	if !r.MatchesDate(tue) {
		t.Error("weekly:tue should match Tuesday")
	}
	if r.MatchesDate(wed) {
		t.Error("weekly:tue should not match Wednesday")
	}
}

func TestMatchesMonthly15(t *testing.T) {
	r, _ := ParseRule("monthly:15")
	if !r.MatchesDate(date(2026, time.January, 15)) {
		t.Error("monthly:15 should match Jan 15")
	}
	if !r.MatchesDate(date(2026, time.February, 15)) {
		t.Error("monthly:15 should match Feb 15")
	}
	if r.MatchesDate(date(2026, time.January, 14)) {
		t.Error("monthly:15 should not match Jan 14")
	}
}

func TestMatchesMonthly31Clamping(t *testing.T) {
	r, _ := ParseRule("monthly:31")
	// January has 31 days
	if !r.MatchesDate(date(2026, time.January, 31)) {
		t.Error("monthly:31 should match Jan 31")
	}
	// February 2026 has 28 days (non-leap) - clamp to 28
	if !r.MatchesDate(date(2026, time.February, 28)) {
		t.Error("monthly:31 should match Feb 28 (clamped, non-leap)")
	}
	if r.MatchesDate(date(2026, time.February, 27)) {
		t.Error("monthly:31 should not match Feb 27")
	}
	// April has 30 days - clamp to 30
	if !r.MatchesDate(date(2026, time.April, 30)) {
		t.Error("monthly:31 should match Apr 30 (clamped)")
	}
	if r.MatchesDate(date(2026, time.April, 29)) {
		t.Error("monthly:31 should not match Apr 29")
	}
	// March has 31 days - exact match
	if !r.MatchesDate(date(2026, time.March, 31)) {
		t.Error("monthly:31 should match Mar 31")
	}
}

func TestMatchesMonthly29LeapYear(t *testing.T) {
	r, _ := ParseRule("monthly:29")
	// 2024 is a leap year
	if !r.MatchesDate(date(2024, time.February, 29)) {
		t.Error("monthly:29 should match Feb 29 in leap year")
	}
	// 2026 is not a leap year - clamp to 28
	if !r.MatchesDate(date(2026, time.February, 28)) {
		t.Error("monthly:29 should match Feb 28 in non-leap year (clamped)")
	}
	if r.MatchesDate(date(2026, time.February, 27)) {
		t.Error("monthly:29 should not match Feb 27 in non-leap year")
	}
}

// --- String() round-trip tests ---

func TestStringDaily(t *testing.T) {
	r, _ := ParseRule("daily")
	if r.String() != "daily" {
		t.Errorf("expected 'daily', got %q", r.String())
	}
}

func TestStringWeekdays(t *testing.T) {
	r, _ := ParseRule("weekdays")
	if r.String() != "weekdays" {
		t.Errorf("expected 'weekdays', got %q", r.String())
	}
}

func TestStringWeekly(t *testing.T) {
	r, _ := ParseRule("weekly:mon,wed,fri")
	if r.String() != "weekly:mon,wed,fri" {
		t.Errorf("expected 'weekly:mon,wed,fri', got %q", r.String())
	}
}

func TestStringMonthly(t *testing.T) {
	r, _ := ParseRule("monthly:15")
	if r.String() != "monthly:15" {
		t.Errorf("expected 'monthly:15', got %q", r.String())
	}
}

func TestStringRoundTrip(t *testing.T) {
	inputs := []string{
		"daily",
		"weekdays",
		"weekly:mon,fri",
		"weekly:tue",
		"weekly:mon,wed,fri",
		"monthly:1",
		"monthly:15",
		"monthly:31",
	}
	for _, s := range inputs {
		r, err := ParseRule(s)
		if err != nil {
			t.Fatalf("ParseRule(%q) error: %v", s, err)
		}
		got := r.String()
		if got != s {
			t.Errorf("round-trip failed: input=%q, String()=%q", s, got)
		}
		// Parse again from String output
		r2, err := ParseRule(got)
		if err != nil {
			t.Fatalf("ParseRule(%q) from String() error: %v", got, err)
		}
		if r2.String() != s {
			t.Errorf("double round-trip failed: %q -> %q -> %q", s, got, r2.String())
		}
	}
}
