package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines semantic color roles for all UI elements.
// Empty string ("") means use terminal default.
type Theme struct {
	// Panel borders
	BorderFocused   lipgloss.Color // focused panel border
	BorderUnfocused lipgloss.Color // unfocused panel border

	// Calendar
	HeaderFg    lipgloss.Color // month/year header
	WeekdayFg   lipgloss.Color // weekday labels (Mo, Tu, ...)
	TodayFg     lipgloss.Color // today's date foreground
	TodayBg     lipgloss.Color // today's date background
	HolidayFg   lipgloss.Color // holiday names and dates
	IndicatorFg lipgloss.Color // todo-count indicators

	// Todo list
	AccentFg    lipgloss.Color // selected item, headings
	MutedFg     lipgloss.Color // secondary text
	CompletedFg lipgloss.Color // completed todo text
	EmptyFg     lipgloss.Color // "no todos" placeholder

	// General
	NormalFg lipgloss.Color // default foreground
	NormalBg lipgloss.Color // default background
}

// Dark returns the default dark theme matching the original hardcoded colors.
func Dark() Theme {
	return Theme{
		BorderFocused:   lipgloss.Color("#5F5FD7"), // ANSI 62
		BorderUnfocused: lipgloss.Color("#585858"), // ANSI 240
		HeaderFg:        lipgloss.Color(""),         // terminal default
		WeekdayFg:       lipgloss.Color(""),         // terminal default
		TodayFg:         lipgloss.Color(""),         // terminal default
		TodayBg:         lipgloss.Color(""),         // terminal default
		HolidayFg:       lipgloss.Color("#AF0000"),  // ANSI 1 red
		IndicatorFg:     lipgloss.Color(""),         // terminal default
		AccentFg:        lipgloss.Color("#5F5FD7"),  // ANSI 62
		MutedFg:         lipgloss.Color("#585858"),  // ANSI 240
		CompletedFg:     lipgloss.Color("#585858"),  // ANSI 240
		EmptyFg:         lipgloss.Color("#585858"),  // ANSI 240
		NormalFg:        lipgloss.Color(""),         // terminal default
		NormalBg:        lipgloss.Color(""),         // terminal default
	}
}

// Light returns a theme for light terminal backgrounds.
func Light() Theme {
	return Theme{
		BorderFocused:   lipgloss.Color("#5F5FD7"), // indigo
		BorderUnfocused: lipgloss.Color("#BCBCBC"), // light grey
		HeaderFg:        lipgloss.Color("#303030"), // dark grey
		WeekdayFg:       lipgloss.Color("#8A8A8A"), // medium grey
		TodayFg:         lipgloss.Color("#FFFFFF"), // white
		TodayBg:         lipgloss.Color("#5F5FD7"), // indigo
		HolidayFg:       lipgloss.Color("#D70000"), // red
		IndicatorFg:     lipgloss.Color("#005FAF"), // blue
		AccentFg:        lipgloss.Color("#5F5FD7"), // indigo
		MutedFg:         lipgloss.Color("#8A8A8A"), // medium grey
		CompletedFg:     lipgloss.Color("#BCBCBC"), // light grey
		EmptyFg:         lipgloss.Color("#8A8A8A"), // medium grey
		NormalFg:        lipgloss.Color("#303030"), // dark grey
		NormalBg:        lipgloss.Color(""),         // terminal default
	}
}

// Nord returns a theme based on the Nord color palette.
// https://www.nordtheme.com
func Nord() Theme {
	return Theme{
		BorderFocused:   lipgloss.Color("#88C0D0"), // nord8 frost
		BorderUnfocused: lipgloss.Color("#4C566A"), // nord3 polar night
		HeaderFg:        lipgloss.Color("#ECEFF4"), // nord6 snow storm
		WeekdayFg:       lipgloss.Color("#4C566A"), // nord3 polar night
		TodayFg:         lipgloss.Color("#2E3440"), // nord0 polar night
		TodayBg:         lipgloss.Color("#88C0D0"), // nord8 frost
		HolidayFg:       lipgloss.Color("#BF616A"), // nord11 aurora red
		IndicatorFg:     lipgloss.Color("#A3BE8C"), // nord14 aurora green
		AccentFg:        lipgloss.Color("#88C0D0"), // nord8 frost
		MutedFg:         lipgloss.Color("#4C566A"), // nord3 polar night
		CompletedFg:     lipgloss.Color("#4C566A"), // nord3 polar night
		EmptyFg:         lipgloss.Color("#4C566A"), // nord3 polar night
		NormalFg:        lipgloss.Color("#D8DEE9"), // nord4 snow storm
		NormalBg:        lipgloss.Color(""),         // terminal default
	}
}

// Solarized returns a theme based on the Solarized Dark color palette.
// https://ethanschoonover.com/solarized
func Solarized() Theme {
	return Theme{
		BorderFocused:   lipgloss.Color("#268BD2"), // blue
		BorderUnfocused: lipgloss.Color("#586E75"), // base01
		HeaderFg:        lipgloss.Color("#93A1A1"), // base1
		WeekdayFg:       lipgloss.Color("#586E75"), // base01
		TodayFg:         lipgloss.Color("#FDF6E3"), // base3
		TodayBg:         lipgloss.Color("#268BD2"), // blue
		HolidayFg:       lipgloss.Color("#DC322F"), // red
		IndicatorFg:     lipgloss.Color("#859900"), // green
		AccentFg:        lipgloss.Color("#268BD2"), // blue
		MutedFg:         lipgloss.Color("#586E75"), // base01
		CompletedFg:     lipgloss.Color("#586E75"), // base01
		EmptyFg:         lipgloss.Color("#586E75"), // base01
		NormalFg:        lipgloss.Color("#839496"), // base0
		NormalBg:        lipgloss.Color(""),         // terminal default
	}
}

// ForName returns the theme matching the given name.
// Unknown or empty names default to Dark.
func ForName(name string) Theme {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "light":
		return Light()
	case "nord":
		return Nord()
	case "solarized":
		return Solarized()
	default:
		return Dark()
	}
}
