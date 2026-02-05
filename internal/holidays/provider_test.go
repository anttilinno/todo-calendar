package holidays

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"
)

// riigipyhad mirrors the JSON structure from riigipühad.ee
type riigipyhad struct {
	Date   string  `json:"date"`
	Title  string  `json:"title"`
	Kind   string  `json:"kind"`
	KindID string  `json:"kind_id"`
	Notes  *string `json:"notes"`
}

// loadFixture loads the Estonian holiday fixture and returns only public
// holidays (kind_id "1" = Riigipüha, "2" = Rahvuspüha). These are the
// official days off that the rickar/cal library should know about.
func loadFixture(t *testing.T) []riigipyhad {
	t.Helper()
	data, err := os.ReadFile("testdata/ee_riigipyhad.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var all []riigipyhad
	if err := json.Unmarshal(data, &all); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	var public []riigipyhad
	for _, h := range all {
		if h.KindID == "1" || h.KindID == "2" {
			public = append(public, h)
		}
	}
	return public
}

func TestEstonianHolidays_AgainstRiigipyhad(t *testing.T) {
	provider, err := NewProvider("ee")
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	fixture := loadFixture(t)

	// Group fixture holidays by year+month for efficient comparison.
	type yearMonth struct {
		year  int
		month time.Month
	}
	expected := make(map[yearMonth]map[int]string) // day -> title
	for _, h := range fixture {
		var y, m, d int
		fmt.Sscanf(h.Date, "%d-%d-%d", &y, &m, &d)
		ym := yearMonth{y, time.Month(m)}
		if expected[ym] == nil {
			expected[ym] = make(map[int]string)
		}
		expected[ym][d] = h.Title
	}

	// Collect all year-months to test.
	var months []yearMonth
	for ym := range expected {
		months = append(months, ym)
	}
	sort.Slice(months, func(i, j int) bool {
		if months[i].year != months[j].year {
			return months[i].year < months[j].year
		}
		return months[i].month < months[j].month
	})

	var missing []string
	var extra []string

	for _, ym := range months {
		got := provider.HolidaysInMonth(ym.year, ym.month)
		want := expected[ym]

		// Check each expected holiday exists in library output.
		for day, title := range want {
			if !got[day] {
				missing = append(missing, fmt.Sprintf(
					"%d-%02d-%02d %s", ym.year, ym.month, day, title))
			}
		}

		// Check library doesn't report extra holidays for this month.
		for day := range got {
			if _, ok := want[day]; !ok {
				missing := fmt.Sprintf("%d-%02d-%02d", ym.year, ym.month, day)
				extra = append(extra, missing)
			}
		}
	}

	// Also check months with no fixture holidays don't produce library hits.
	years := make(map[int]bool)
	for ym := range expected {
		years[ym.year] = true
	}
	for year := range years {
		for m := time.January; m <= time.December; m++ {
			ym := yearMonth{year, m}
			if expected[ym] != nil {
				continue // already checked above
			}
			got := provider.HolidaysInMonth(year, m)
			for day := range got {
				extra = append(extra, fmt.Sprintf(
					"%d-%02d-%02d (no fixture holidays in month)", year, m, day))
			}
		}
	}

	if len(missing) > 0 {
		t.Errorf("holidays in fixture but NOT returned by library (%d):", len(missing))
		for _, m := range missing {
			t.Errorf("  MISSING: %s", m)
		}
	}
	if len(extra) > 0 {
		t.Errorf("holidays returned by library but NOT in fixture (%d):", len(extra))
		for _, e := range extra {
			t.Errorf("  EXTRA: %s", e)
		}
	}
}
