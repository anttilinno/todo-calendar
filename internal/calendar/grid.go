package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/store"
)

// gridWidth is the total character width of the 4-char-cell calendar grid.
// 7 cells x 4 chars + 6 separators x 1 char = 34.
const gridWidth = 34

// RenderGrid produces a 34-character wide calendar month grid with 4-char cells.
// It is a pure function with no side effects.
//
// Parameters:
//   - year, month: the month to render
//   - today: day number to highlight as today (0 for none)
//   - holidays: map of day numbers that are holidays
//   - mondayStart: if true, weeks start on Monday; otherwise Sunday
//   - indicators: map of day numbers to count of incomplete todos (nil safe)
func RenderGrid(year int, month time.Month, today int, holidays map[int]bool, mondayStart bool, indicators map[int]int, s Styles) string {
	var b strings.Builder

	// Title line: month and year, centered in grid width.
	title := fmt.Sprintf("%s %d", month.String(), year)
	pad := (gridWidth - len(title)) / 2
	if pad < 0 {
		pad = 0
	}
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(s.Header.Render(title))
	b.WriteString("\n")

	// Weekday header (4 chars per day label, 1 char separator = 34 chars total).
	if mondayStart {
		b.WriteString(s.WeekdayHdr.Render(" Mo   Tu   We   Th   Fr   Sa   Su "))
	} else {
		b.WriteString(s.WeekdayHdr.Render(" Su   Mo   Tu   We   Th   Fr   Sa "))
	}
	b.WriteString("\n")

	// Compute first weekday and days in month.
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Weekday()
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	// Determine starting column (0-indexed).
	startCol := int(firstDay) // Sunday=0
	if mondayStart {
		startCol = (int(firstDay) + 6) % 7 // Monday=0
	}

	// Leading blanks (4 chars each with 1 char separator).
	col := startCol
	for i := 0; i < startCol; i++ {
		b.WriteString("    ")
		if i < startCol-1 {
			b.WriteString(" ")
		}
	}
	// Separator between last blank and first day cell.
	if startCol > 0 {
		b.WriteString(" ")
	}

	// Day cells.
	for day := 1; day <= daysInMonth; day++ {
		// Format cell to 4 visible characters BEFORE styling.
		var cell string
		if indicators[day] > 0 {
			cell = fmt.Sprintf("[%2d]", day)
		} else {
			cell = fmt.Sprintf(" %2d ", day)
		}

		// Apply style based on priority: today > holiday > indicator > normal.
		switch {
		case day == today:
			cell = s.Today.Render(cell)
		case holidays[day]:
			cell = s.Holiday.Render(cell)
		case indicators[day] > 0:
			cell = s.Indicator.Render(cell)
		default:
			cell = s.Normal.Render(cell)
		}

		b.WriteString(cell)

		col++
		if col == 7 {
			b.WriteString("\n")
			col = 0
		} else {
			b.WriteString(" ")
		}
	}

	// Trailing newline if row didn't end at column 7.
	if col != 0 {
		b.WriteString("\n")
	}

	return b.String()
}

// RenderWeekGrid produces a 34-character wide single-week grid with 4-char cells.
// It is a pure function with no side effects.
//
// Parameters:
//   - weekStart: the first day of the week to render
//   - today: today's date for highlighting
//   - hp: holiday provider for holiday lookup
//   - mondayStart: if true, weeks start on Monday; otherwise Sunday
//   - st: store for incomplete todo indicator lookup
//   - s: calendar styles
func RenderWeekGrid(weekStart time.Time, today time.Time, hp *holidays.Provider, mondayStart bool, st *store.Store, s Styles) string {
	var b strings.Builder

	weekEnd := weekStart.AddDate(0, 0, 6)

	// Header line: date range centered in gridWidth.
	var title string
	if weekStart.Year() != weekEnd.Year() {
		// Cross year: "Dec 29, 2025 - Jan 4, 2026"
		title = fmt.Sprintf("%s %d, %d - %s %d, %d",
			weekStart.Month().String()[:3], weekStart.Day(), weekStart.Year(),
			weekEnd.Month().String()[:3], weekEnd.Day(), weekEnd.Year())
	} else if weekStart.Month() != weekEnd.Month() {
		// Cross month: "Jan 26 - Feb 1, 2026"
		title = fmt.Sprintf("%s %d - %s %d, %d",
			weekStart.Month().String()[:3], weekStart.Day(),
			weekEnd.Month().String()[:3], weekEnd.Day(),
			weekEnd.Year())
	} else {
		// Same month: "Feb 2 - 8, 2026"
		title = fmt.Sprintf("%s %d - %d, %d",
			weekStart.Month().String()[:3], weekStart.Day(),
			weekEnd.Day(), weekEnd.Year())
	}

	pad := (gridWidth - len(title)) / 2
	if pad < 0 {
		pad = 0
	}
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(s.Header.Render(title))
	b.WriteString("\n")

	// Weekday header (same as RenderGrid).
	if mondayStart {
		b.WriteString(s.WeekdayHdr.Render(" Mo   Tu   We   Th   Fr   Sa   Su "))
	} else {
		b.WriteString(s.WeekdayHdr.Render(" Su   Mo   Tu   We   Th   Fr   Sa "))
	}
	b.WriteString("\n")

	// Cache holiday and indicator data per (year, month) to avoid redundant lookups.
	type monthKey struct {
		year  int
		month time.Month
	}
	holidayCache := make(map[monthKey]map[int]bool)
	indicatorCache := make(map[monthKey]map[int]int)

	getHolidays := func(y int, m time.Month) map[int]bool {
		k := monthKey{y, m}
		if v, ok := holidayCache[k]; ok {
			return v
		}
		v := hp.HolidaysInMonth(y, m)
		holidayCache[k] = v
		return v
	}

	getIndicators := func(y int, m time.Month) map[int]int {
		k := monthKey{y, m}
		if v, ok := indicatorCache[k]; ok {
			return v
		}
		v := st.IncompleteTodosPerDay(y, m)
		indicatorCache[k] = v
		return v
	}

	// Day cells (single row of 7 days).
	for i := 0; i < 7; i++ {
		d := weekStart.AddDate(0, 0, i)
		dy, dm, dd := d.Year(), d.Month(), d.Day()

		hols := getHolidays(dy, dm)
		inds := getIndicators(dy, dm)

		isToday := d.Year() == today.Year() && d.Month() == today.Month() && d.Day() == today.Day()

		// Format cell to 4 visible characters.
		var cell string
		if inds[dd] > 0 {
			cell = fmt.Sprintf("[%2d]", dd)
		} else {
			cell = fmt.Sprintf(" %2d ", dd)
		}

		// Apply style based on priority: today > holiday > indicator > normal.
		switch {
		case isToday:
			cell = s.Today.Render(cell)
		case hols[dd]:
			cell = s.Holiday.Render(cell)
		case inds[dd] > 0:
			cell = s.Indicator.Render(cell)
		default:
			cell = s.Normal.Render(cell)
		}

		b.WriteString(cell)

		if i < 6 {
			b.WriteString(" ")
		}
	}
	b.WriteString("\n")

	return b.String()
}
