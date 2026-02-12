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

// fuzzyStatus returns "pending", "done", or "" based on whether any todos are incomplete.
func fuzzyStatus(todos []store.Todo) string {
	if len(todos) == 0 {
		return ""
	}
	for _, t := range todos {
		if !t.Done {
			return "pending"
		}
	}
	return "done"
}

// RenderGrid produces a 34-character wide calendar month grid with 4-char cells.
// It is a pure function with no side effects.
//
// Parameters:
//   - year, month: the month to render
//   - today: day number to highlight as today (0 for none)
//   - holidays: map of day numbers that are holidays
//   - mondayStart: if true, weeks start on Monday; otherwise Sunday
//   - indicators: map of day numbers to count of incomplete todos (nil safe)
//   - st: store for querying month/year fuzzy todos (nil safe)
func RenderGrid(year int, month time.Month, today int, holidays map[int]bool, mondayStart bool, indicators map[int]int, totals map[int]int, st store.TodoStore, showMonthTodos bool, showYearTodos bool, s Styles) string {
	var b strings.Builder

	// Title line: month and year, centered in grid width.
	title := fmt.Sprintf("%s %d", month.String(), year)

	// Determine circle indicators for fuzzy-date todos.
	var monthCircle, yearCircle string
	// visibleExtra tracks the number of extra visible characters added by circles + spaces.
	visibleExtra := 0
	if st != nil {
		if showMonthTodos {
			ms := fuzzyStatus(st.MonthTodos(year, month))
			switch ms {
			case "pending":
				monthCircle = s.FuzzyPending.Render("\u25cf")
			case "done":
				monthCircle = s.FuzzyDone.Render("\u25cf")
			}
		}

		if showYearTodos {
			ys := fuzzyStatus(st.YearTodos(year))
			switch ys {
			case "pending":
				yearCircle = s.FuzzyPending.Render("\u25cf")
			case "done":
				yearCircle = s.FuzzyDone.Render("\u25cf")
			}
		}
	}

	// Build composed title with optional circles.
	var composed string
	if monthCircle != "" {
		composed = monthCircle + " " + s.Header.Render(title)
		visibleExtra += 2 // "● " = 1 circle + 1 space
	} else {
		composed = s.Header.Render(title)
	}
	if yearCircle != "" {
		composed = composed + " " + yearCircle
		visibleExtra += 2 // " ●" = 1 space + 1 circle
	}

	// Center based on visible width (title text + circle chars + spaces).
	visibleWidth := len(title) + visibleExtra
	pad := (gridWidth - visibleWidth) / 2
	if pad < 0 {
		pad = 0
	}
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(composed)
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
		hasPending := indicators[day] > 0
		hasAllDone := !hasPending && totals[day] > 0
		var cell string
		if hasPending || hasAllDone {
			cell = fmt.Sprintf("[%2d]", day)
		} else {
			cell = fmt.Sprintf(" %2d ", day)
		}

		// Apply style based on priority: today+indicator > today+done > today > holiday > indicator > done > normal.
		switch {
		case day == today && hasPending:
			cell = s.TodayIndicator.Render(cell)
		case day == today && hasAllDone:
			cell = s.TodayDone.Render(cell)
		case day == today:
			cell = s.Today.Render(cell)
		case holidays[day]:
			cell = s.Holiday.Render(cell)
		case hasPending:
			cell = s.Indicator.Render(cell)
		case hasAllDone:
			cell = s.IndicatorDone.Render(cell)
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
func RenderWeekGrid(weekStart time.Time, today time.Time, hp *holidays.Provider, mondayStart bool, st store.TodoStore, s Styles) string {
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
	totalsCache := make(map[monthKey]map[int]int)

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

	getTotals := func(y int, m time.Month) map[int]int {
		k := monthKey{y, m}
		if v, ok := totalsCache[k]; ok {
			return v
		}
		v := st.TotalTodosPerDay(y, m)
		totalsCache[k] = v
		return v
	}

	// Day cells (single row of 7 days).
	for i := 0; i < 7; i++ {
		d := weekStart.AddDate(0, 0, i)
		dy, dm, dd := d.Year(), d.Month(), d.Day()

		hols := getHolidays(dy, dm)
		inds := getIndicators(dy, dm)
		tots := getTotals(dy, dm)

		isToday := d.Year() == today.Year() && d.Month() == today.Month() && d.Day() == today.Day()

		// Format cell to 4 visible characters.
		hasPending := inds[dd] > 0
		hasAllDone := !hasPending && tots[dd] > 0
		var cell string
		if hasPending || hasAllDone {
			cell = fmt.Sprintf("[%2d]", dd)
		} else {
			cell = fmt.Sprintf(" %2d ", dd)
		}

		// Apply style based on priority: today+indicator > today+done > today > holiday > indicator > done > normal.
		switch {
		case isToday && hasPending:
			cell = s.TodayIndicator.Render(cell)
		case isToday && hasAllDone:
			cell = s.TodayDone.Render(cell)
		case isToday:
			cell = s.Today.Render(cell)
		case hols[dd]:
			cell = s.Holiday.Render(cell)
		case hasPending:
			cell = s.Indicator.Render(cell)
		case hasAllDone:
			cell = s.IndicatorDone.Render(cell)
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
