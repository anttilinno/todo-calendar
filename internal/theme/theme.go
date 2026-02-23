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

	// Overview
	PendingFg        lipgloss.Color // pending todo count in overview
	CompletedCountFg lipgloss.Color // completed todo count in overview

	// Priority levels
	PriorityP1Fg lipgloss.Color // P1 (urgent/critical) -- red family
	PriorityP2Fg lipgloss.Color // P2 (high) -- orange family
	PriorityP3Fg lipgloss.Color // P3 (medium) -- blue family
	PriorityP4Fg lipgloss.Color // P4 (low) -- grey/muted family

	// Calendar events
	EventFg lipgloss.Color // calendar event text
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
		HolidayFg:       lipgloss.Color("#D75FAF"),  // magenta (distinct from P1 red)
		IndicatorFg:     lipgloss.Color(""),         // terminal default
		AccentFg:        lipgloss.Color("#5F5FD7"),  // ANSI 62
		MutedFg:         lipgloss.Color("#585858"),  // ANSI 240
		CompletedFg:     lipgloss.Color("#585858"),  // ANSI 240
		EmptyFg:         lipgloss.Color("#585858"),  // ANSI 240
		NormalFg:        lipgloss.Color(""),         // terminal default
		NormalBg:        lipgloss.Color(""),         // terminal default
		PendingFg:        lipgloss.Color("#D75F5F"), // warm rose
		CompletedCountFg: lipgloss.Color("#87AF87"), // soft sage green
		PriorityP1Fg:    lipgloss.Color("#FF5F5F"), // bright red
		PriorityP2Fg:    lipgloss.Color("#FFAF5F"), // orange
		PriorityP3Fg:    lipgloss.Color("#5F87FF"), // blue
		PriorityP4Fg:    lipgloss.Color("#808080"), // grey
		EventFg:         lipgloss.Color("#5FAFAF"), // teal
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
		HolidayFg:       lipgloss.Color("#AF005F"), // dark magenta (distinct from P1 red)
		IndicatorFg:     lipgloss.Color("#005FAF"), // blue
		AccentFg:        lipgloss.Color("#5F5FD7"), // indigo
		MutedFg:         lipgloss.Color("#8A8A8A"), // medium grey
		CompletedFg:     lipgloss.Color("#BCBCBC"), // light grey
		EmptyFg:         lipgloss.Color("#8A8A8A"), // medium grey
		NormalFg:        lipgloss.Color("#303030"), // dark grey
		NormalBg:        lipgloss.Color(""),         // terminal default
		PendingFg:        lipgloss.Color("#D70000"), // medium red
		CompletedCountFg: lipgloss.Color("#008700"), // forest green
		PriorityP1Fg:    lipgloss.Color("#D70000"), // dark red
		PriorityP2Fg:    lipgloss.Color("#AF5F00"), // dark orange
		PriorityP3Fg:    lipgloss.Color("#005FAF"), // dark blue
		PriorityP4Fg:    lipgloss.Color("#8A8A8A"), // medium grey
		EventFg:         lipgloss.Color("#00878A"), // dark teal
	}
}

// Nord returns a theme based on the Nord color palette.
// https://www.nordtheme.com
func Nord() Theme {
	return Theme{
		BorderFocused:   lipgloss.Color("#88C0D0"), // nord8 frost
		BorderUnfocused: lipgloss.Color("#4C566A"), // nord3 polar night
		HeaderFg:        lipgloss.Color("#ECEFF4"), // nord6 snow storm
		WeekdayFg:       lipgloss.Color("#81A1C1"), // nord9 muted frost (better contrast)
		TodayFg:         lipgloss.Color("#2E3440"), // nord0 polar night
		TodayBg:         lipgloss.Color("#88C0D0"), // nord8 frost
		HolidayFg:       lipgloss.Color("#B48EAD"), // nord15 aurora purple (distinct from P1 red)
		IndicatorFg:     lipgloss.Color("#EBCB8B"), // nord13 aurora yellow
		AccentFg:        lipgloss.Color("#88C0D0"), // nord8 frost
		MutedFg:         lipgloss.Color("#81A1C1"), // nord9 muted frost (better contrast)
		CompletedFg:     lipgloss.Color("#4C566A"), // nord3 polar night
		EmptyFg:         lipgloss.Color("#81A1C1"), // nord9 muted frost (better contrast)
		NormalFg:        lipgloss.Color("#D8DEE9"), // nord4 snow storm
		NormalBg:        lipgloss.Color(""),         // terminal default
		PendingFg:        lipgloss.Color("#BF616A"), // nord11 aurora red
		CompletedCountFg: lipgloss.Color("#A3BE8C"), // nord14 aurora green
		PriorityP1Fg:    lipgloss.Color("#BF616A"), // nord11 aurora red
		PriorityP2Fg:    lipgloss.Color("#D08770"), // nord12 aurora orange
		PriorityP3Fg:    lipgloss.Color("#5E81AC"), // nord10 frost blue
		PriorityP4Fg:    lipgloss.Color("#4C566A"), // nord3 polar night
		EventFg:         lipgloss.Color("#8FBCBB"), // nord7 frost teal
	}
}

// Solarized returns a theme based on the Solarized Dark color palette.
// https://ethanschoonover.com/solarized
func Solarized() Theme {
	return Theme{
		BorderFocused:   lipgloss.Color("#268BD2"), // blue
		BorderUnfocused: lipgloss.Color("#586E75"), // base01
		HeaderFg:        lipgloss.Color("#93A1A1"), // base1
		WeekdayFg:       lipgloss.Color("#657B83"), // base00 (better contrast)
		TodayFg:         lipgloss.Color("#FDF6E3"), // base3
		TodayBg:         lipgloss.Color("#268BD2"), // blue
		HolidayFg:       lipgloss.Color("#D33682"), // solarized magenta (distinct from P1 red)
		IndicatorFg:     lipgloss.Color("#B58900"), // yellow
		AccentFg:        lipgloss.Color("#268BD2"), // blue
		MutedFg:         lipgloss.Color("#657B83"), // base00 (better contrast)
		CompletedFg:     lipgloss.Color("#586E75"), // base01
		EmptyFg:         lipgloss.Color("#657B83"), // base00 (better contrast)
		NormalFg:        lipgloss.Color("#839496"), // base0
		NormalBg:        lipgloss.Color(""),         // terminal default
		PendingFg:        lipgloss.Color("#DC322F"), // solarized red
		CompletedCountFg: lipgloss.Color("#859900"), // solarized green
		PriorityP1Fg:    lipgloss.Color("#DC322F"), // solarized red
		PriorityP2Fg:    lipgloss.Color("#CB4B16"), // solarized orange
		PriorityP3Fg:    lipgloss.Color("#268BD2"), // solarized blue
		PriorityP4Fg:    lipgloss.Color("#586E75"), // solarized base01
		EventFg:         lipgloss.Color("#2AA198"), // solarized cyan
	}
}

// Names returns the list of available theme names.
// This is the single source of truth for theme enumeration.
func Names() []string {
	return []string{"dark", "light", "nord", "solarized"}
}

// PriorityColorHex returns the raw hex color string for the given priority level.
// Priority 1-4 maps to PriorityP1Fg through PriorityP4Fg.
// Priority 0 or any other value returns AccentFg as fallback.
func (t Theme) PriorityColorHex(priority int) string {
	switch priority {
	case 1:
		return string(t.PriorityP1Fg)
	case 2:
		return string(t.PriorityP2Fg)
	case 3:
		return string(t.PriorityP3Fg)
	case 4:
		return string(t.PriorityP4Fg)
	default:
		return string(t.AccentFg)
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
