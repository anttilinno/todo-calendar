package calendar

import (
	"fmt"
	"strings"
	"time"
)

// RenderGrid produces a 20-character wide calendar month grid.
// It is a pure function with no side effects.
//
// Parameters:
//   - year, month: the month to render
//   - today: day number to highlight as today (0 for none)
//   - holidays: map of day numbers that are holidays
//   - mondayStart: if true, weeks start on Monday; otherwise Sunday
func RenderGrid(year int, month time.Month, today int, holidays map[int]bool, mondayStart bool) string {
	var b strings.Builder

	// Title line: month and year, centered in 20 chars.
	title := fmt.Sprintf("%s %d", month.String(), year)
	pad := (20 - len(title)) / 2
	if pad < 0 {
		pad = 0
	}
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(headerStyle.Render(title))
	b.WriteString("\n")

	// Weekday header.
	if mondayStart {
		b.WriteString(weekdayHdrStyle.Render("Mo Tu We Th Fr Sa Su"))
	} else {
		b.WriteString(weekdayHdrStyle.Render("Su Mo Tu We Th Fr Sa"))
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

	// Leading blanks.
	col := startCol
	for i := 0; i < startCol; i++ {
		b.WriteString("   ")
	}

	// Day cells.
	for day := 1; day <= daysInMonth; day++ {
		// Format number BEFORE styling (research pitfall #3).
		cell := fmt.Sprintf("%2d", day)

		// Apply style based on priority: today > holiday > normal.
		switch {
		case day == today:
			cell = todayStyle.Render(cell)
		case holidays[day]:
			cell = holidayStyle.Render(cell)
		default:
			cell = normalStyle.Render(cell)
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
