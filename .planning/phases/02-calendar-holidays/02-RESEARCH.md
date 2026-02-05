# Phase 2: Calendar + Holidays - Research

**Researched:** 2026-02-05
**Domain:** Calendar grid rendering, holiday integration, TOML configuration in Go TUI
**Confidence:** HIGH

## Summary

Phase 2 transforms the calendar placeholder pane into a fully functional monthly calendar grid with today-highlight, month navigation, and national holidays displayed in red. It also introduces TOML-based configuration for country selection. The phase builds on the existing Bubble Tea scaffold from Phase 1, extending the `internal/calendar/model.go` placeholder into a real component.

The calendar grid rendering is pure Go `time` package math -- no external calendar rendering library is needed. The algorithm computes the first weekday of the month, pads leading blanks, and fills a 7-column grid with 3-character-wide day cells. This produces output identical to `cal` (20 chars wide), which fits comfortably within the 24-char `Width()` set on the calendar pane (22 chars content area after padding). Each day cell is individually styled with Lip Gloss: normal, today-highlight (reverse/bold), or holiday (red foreground).

For holidays, `rickar/cal/v2` (v2.1.27) is the standard Go library -- offline, zero API dependencies, with 48 country subpackages organized by ISO code. Each country subpackage exports a `Holidays` slice of `*cal.Holiday` pointers, making it trivial to load all holidays for a country. The country code comes from a TOML config file read with `BurntSushi/toml` (v1.6.0), the most widely adopted TOML library for Go (36,500+ importers). The config file path follows `os.UserConfigDir()` for XDG compliance.

**Primary recommendation:** Build the calendar grid as pure `time` package math with per-cell Lip Gloss styling. Use `rickar/cal/v2` for holidays with a registry map that maps country codes to holiday slices. Use `BurntSushi/toml` for config parsing. Keep the first weekday configurable (Sunday or Monday start) alongside the country code.

## Standard Stack

### Core (Phase 2 additions)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `rickar/cal/v2` | v2.1.27 | Holiday definitions by country | 401 stars, 40 releases, 48 country subpackages, offline, BSD-3 license; the only mature Go holiday library that works without an API |
| `BurntSushi/toml` | v1.6.0 | TOML config file parsing | 36,500+ importers, MIT license, simple `DecodeFile` API, TOML v1.1.0 compliant |

### Existing (from Phase 1, unchanged)

| Library | Version | Purpose |
|---------|---------|---------|
| Bubble Tea | v1.3.10 | TUI framework |
| Lip Gloss | v1.1.0 | Terminal styling (per-cell calendar coloring) |
| Bubbles | v0.21.1 | `key.Binding` for month navigation keys |

### Supporting (stdlib)

| Package | Purpose | When to Use |
|---------|---------|-------------|
| `time` | Calendar math: first day of month, days in month, weekday | All date calculations |
| `fmt` | `Sprintf` for day number formatting (`%2d`) | Grid cell formatting |
| `strings` | `Builder` for efficient grid string construction | View rendering |
| `os` | `UserConfigDir()` for XDG-compliant config path | Config file location |
| `path/filepath` | `Join` for platform-safe paths | Config file path |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `BurntSushi/toml` | `pelletier/go-toml/v2` | pelletier is 2-5x faster but has 24x fewer importers; BurntSushi is the ecosystem standard and simpler API for our read-only use case |
| `rickar/cal/v2` | `holidays/go-holidays` | go-holidays is less mature, fewer countries, fewer stars |
| `rickar/cal/v2` | `omidnikrah/go-holidays` | Requires Nager API (network dependency) -- violates offline constraint |
| Custom holiday registry | Reflection/plugin loading | Compile-time country imports with a switch/map is simpler, safer, and sufficient for ~48 countries |

**Installation:**

```bash
go get github.com/rickar/cal/v2@v2.1.27
go get github.com/BurntSushi/toml@v1.6.0
```

## Architecture Patterns

### Recommended Project Structure (Phase 2 additions)

```
internal/
├── calendar/
│   ├── model.go       # Calendar Bubble Tea model (state, Update, View)
│   ├── grid.go        # Pure function: renderGrid(year, month, today, holidays, width) string
│   ├── keys.go        # Calendar-specific key bindings (left/right month nav)
│   └── styles.go      # Calendar-specific Lip Gloss styles (today, holiday, header)
├── holidays/
│   ├── registry.go    # Map[string][]*cal.Holiday country code -> holiday slice
│   └── provider.go    # HolidayProvider: loads holidays, checks dates
└── config/
    ├── config.go      # Config struct, Load/Save functions
    └── paths.go       # XDG config directory resolution
```

### Pattern 1: Calendar Grid as Pure Function

**What:** The grid rendering is a pure function that takes year, month, today's date, a set of holiday dates, and available width, then returns a styled string. It has no side effects and no dependency on Bubble Tea.

**When to use:** Always -- this makes the grid testable without the TUI framework.

**Confidence:** HIGH (standard functional rendering pattern in Elm Architecture)

**Example:**

```go
// internal/calendar/grid.go

// RenderGrid produces a styled calendar grid string.
// holidays is a set of day-of-month numbers that are holidays.
// today is 0 if the current month is not being displayed.
func RenderGrid(year int, month time.Month, today int, holidays map[int]bool, mondayStart bool) string {
    var b strings.Builder

    // Header: "   February 2026"
    title := fmt.Sprintf("%s %d", month.String(), year)
    // Center the title in 20 chars
    padding := (20 - len(title)) / 2
    b.WriteString(strings.Repeat(" ", padding))
    b.WriteString(headerStyle.Render(title))
    b.WriteString("\n")

    // Weekday headers
    if mondayStart {
        b.WriteString("Mo Tu We Th Fr Sa Su")
    } else {
        b.WriteString("Su Mo Tu We Th Fr Sa")
    }
    b.WriteString("\n")

    // First day of month
    first := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
    startWeekday := int(first.Weekday()) // Sunday=0
    if mondayStart {
        startWeekday = (startWeekday + 6) % 7 // Monday=0
    }

    // Days in month
    daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

    // Leading blanks
    for i := 0; i < startWeekday; i++ {
        b.WriteString("   ")
    }

    col := startWeekday
    for day := 1; day <= daysInMonth; day++ {
        cell := fmt.Sprintf("%2d", day)

        // Apply style based on day type
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
    if col != 0 {
        b.WriteString("\n")
    }

    return b.String()
}
```

### Pattern 2: Calendar Model with Month State

**What:** The calendar model tracks the currently displayed year/month and navigates with left/right keys. It recalculates the holiday set when the month changes.

**When to use:** Always for the calendar pane.

**Confidence:** HIGH (standard Bubble Tea state management pattern)

**Example:**

```go
// internal/calendar/model.go

type Model struct {
    focused  bool
    width    int
    height   int
    year     int
    month    time.Month
    today    time.Time
    holidays map[int]bool // day-of-month -> is holiday
    provider *holidays.Provider
    keys     KeyMap
    mondayStart bool
}

func New(provider *holidays.Provider, mondayStart bool) Model {
    now := time.Now()
    m := Model{
        year:        now.Year(),
        month:       now.Month(),
        today:       now,
        provider:    provider,
        keys:        DefaultKeyMap(),
        mondayStart: mondayStart,
    }
    m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
    return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if !m.focused {
            return m, nil
        }
        switch {
        case key.Matches(msg, m.keys.PrevMonth):
            m.month--
            if m.month < time.January {
                m.month = time.December
                m.year--
            }
            m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
        case key.Matches(msg, m.keys.NextMonth):
            m.month++
            if m.month > time.December {
                m.month = time.January
                m.year++
            }
            m.holidays = m.provider.HolidaysInMonth(m.year, m.month)
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}

func (m Model) View() string {
    todayDay := 0
    now := time.Now()
    if now.Year() == m.year && now.Month() == m.month {
        todayDay = now.Day()
    }
    return RenderGrid(m.year, m.month, todayDay, m.holidays, m.mondayStart)
}
```

### Pattern 3: Holiday Provider with Country Registry

**What:** A provider wraps `rickar/cal/v2` Calendar, pre-loaded with holidays for the configured country. A registry maps ISO country codes to their holiday slices. The provider offers a `HolidaysInMonth` method that returns a `map[int]bool` of holiday day numbers.

**When to use:** To decouple holiday logic from the calendar model.

**Confidence:** HIGH (straightforward wrapper pattern)

**Example:**

```go
// internal/holidays/registry.go

import (
    "github.com/rickar/cal/v2"
    "github.com/rickar/cal/v2/fi"
    "github.com/rickar/cal/v2/us"
    // ... import more as needed
)

// Registry maps ISO country codes to their holiday slices.
var Registry = map[string][]*cal.Holiday{
    "fi": fi.Holidays,
    "us": us.Holidays,
    // Add more countries as needed
}

// internal/holidays/provider.go

type Provider struct {
    cal     *cal.Calendar
    country string
}

func NewProvider(countryCode string) (*Provider, error) {
    holidays, ok := Registry[countryCode]
    if !ok {
        return nil, fmt.Errorf("unsupported country code: %s", countryCode)
    }
    c := &cal.Calendar{}
    c.AddHoliday(holidays...)
    return &Provider{cal: c, country: countryCode}, nil
}

func (p *Provider) HolidaysInMonth(year int, month time.Month) map[int]bool {
    result := make(map[int]bool)
    daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
    for day := 1; day <= daysInMonth; day++ {
        date := time.Date(year, month, day, 12, 0, 0, 0, time.Local)
        actual, observed, _ := p.cal.IsHoliday(date)
        if actual || observed {
            result[day] = true
        }
    }
    return result
}
```

### Pattern 4: TOML Config with Defaults

**What:** Config struct with TOML tags, loaded from `~/.config/todo-calendar/config.toml`. If file doesn't exist, use defaults and optionally create it.

**When to use:** For the country and first-weekday settings.

**Confidence:** HIGH (verified BurntSushi/toml DecodeFile API)

**Example:**

```go
// internal/config/config.go

import (
    "os"
    "path/filepath"
    "github.com/BurntSushi/toml"
)

type Config struct {
    Country     string `toml:"country"`
    MondayStart bool   `toml:"monday_start"`
}

func DefaultConfig() Config {
    return Config{
        Country:     "us",
        MondayStart: false,
    }
}

func Load() (Config, error) {
    cfg := DefaultConfig()

    path, err := configPath()
    if err != nil {
        return cfg, nil // fall back to defaults
    }

    if _, err := os.Stat(path); os.IsNotExist(err) {
        return cfg, nil // file doesn't exist, use defaults
    }

    _, err = toml.DecodeFile(path, &cfg)
    return cfg, err
}

func configPath() (string, error) {
    dir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, "todo-calendar", "config.toml"), nil
}
```

**Config file example (`~/.config/todo-calendar/config.toml`):**

```toml
# Country code for national holidays (ISO 3166-1 alpha-2, lowercase)
country = "fi"

# Start week on Monday (true) or Sunday (false)
monday_start = true
```

### Pattern 5: Per-Cell Lip Gloss Styling

**What:** Each day number in the calendar grid is individually styled using small Lip Gloss styles. Normal days get default foreground, today gets a reversed/bold style, holidays get red foreground. Styles are applied per-cell, then cells are concatenated with spaces.

**When to use:** For the calendar grid rendering.

**Confidence:** HIGH (verified Lip Gloss Foreground, Bold, Reverse from v1.1.0 docs)

**Example:**

```go
// internal/calendar/styles.go

import "github.com/charmbracelet/lipgloss"

var (
    headerStyle  = lipgloss.NewStyle().Bold(true)
    normalStyle  = lipgloss.NewStyle()
    todayStyle   = lipgloss.NewStyle().Bold(true).Reverse(true)
    holidayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
    weekdayHeaderStyle = lipgloss.NewStyle().Faint(true)
)
```

### Anti-Patterns to Avoid

- **Hand-rolling holiday calculations:** Do NOT compute Easter, floating holidays, or observed-day shifts manually. `rickar/cal/v2` handles all of this including leap years, Easter algorithms, and observation rules.
- **Storing holidays as a flat list of dates:** Use `rickar/cal/v2`'s `IsHoliday` method or pre-compute a `map[int]bool` per month. Don't maintain a global date list.
- **Using `time.Month` arithmetic without normalization:** `time.January - 1` is `0`, not December. Always guard month overflow/underflow explicitly or use `time.Date` normalization.
- **Mixing `time.Local` and `time.UTC`:** Pick one consistently. Use `time.Local` for the calendar since holidays are location-aware.
- **Rendering the entire grid as one styled string:** Style each cell individually and concatenate. Lip Gloss styles applied to the whole grid would color everything uniformly.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Holiday dates for any country | Manual holiday tables, Easter algorithms | `rickar/cal/v2` with country subpackages | Easter alone requires a 19-step algorithm; floating holidays (3rd Monday of X) have edge cases; observed-day rules vary by country |
| TOML parsing | Manual string splitting, regex, or JSON | `BurntSushi/toml` `DecodeFile` | TOML has complex rules for tables, arrays, dates, multiline strings |
| Days in a given month | Hardcoded `[31,28,31,...]` lookup table | `time.Date(year, month+1, 0, ...).Day()` | time.Date handles February in leap years automatically via normalization |
| First weekday of a month | Manual Zeller's congruence or lookup | `time.Date(year, month, 1, ...).Weekday()` | Built into Go's time package, handles all calendar quirks |
| XDG config directory | Hardcoded `~/.config` | `os.UserConfigDir()` | Returns correct path on Linux, macOS, Windows; respects `$XDG_CONFIG_HOME` |
| Terminal color support detection | Manual `$TERM` checking | Lip Gloss handles this via `termenv` | Lip Gloss/termenv auto-detects color profile and degrades gracefully |

**Key insight:** The calendar grid is 100% stdlib `time` math. The only external libraries needed are `rickar/cal/v2` for holiday data and `BurntSushi/toml` for config. Everything else is built-in Go.

## Common Pitfalls

### Pitfall 1: Month Arithmetic Overflow

**What goes wrong:** `time.January - 1` evaluates to `time.Month(0)`, not `time.December`. Similarly `time.December + 1` is `time.Month(13)`, not `time.January`.
**Why it happens:** `time.Month` is an `int` type with no auto-wrapping.
**How to avoid:** Explicitly check bounds in month navigation:
```go
m.month--
if m.month < time.January {
    m.month = time.December
    m.year--
}
```
Alternatively, use `time.Date` normalization: `time.Date(year, month-1, 1, 0, 0, 0, 0, time.Local)` will normalize month 0 to December of the previous year. But explicit checks are clearer.
**Warning signs:** Calendar shows "month 0" or panics on January prev-month.
**Confidence:** HIGH (verified Go time.Month is a plain int)

### Pitfall 2: Off-by-One in Days-in-Month Calculation

**What goes wrong:** Using `time.Date(year, month, 0, ...).Day()` gives the last day of the *previous* month. The correct idiom is `time.Date(year, month+1, 0, ...).Day()`.
**Why it happens:** Day 0 of month M = last day of month M-1 (Go normalization).
**How to avoid:** Always use `month+1, day 0` pattern. Write a helper function `daysInMonth(year int, month time.Month) int`.
**Warning signs:** February shows 31 days, or months are shifted by one.
**Confidence:** HIGH (verified in Go time.Date documentation)

### Pitfall 3: ANSI Escape Sequences Break Column Alignment

**What goes wrong:** When using `fmt.Sprintf("%2d", day)` on a Lip Gloss-styled string, the ANSI escape codes count toward the width, making the column wider than expected.
**Why it happens:** Lip Gloss's `Render()` wraps the string in ANSI escape sequences. `fmt.Sprintf("%*s", width, styled)` counts these invisible bytes.
**How to avoid:** Format the number FIRST (`fmt.Sprintf("%2d", day)`), THEN apply the Lip Gloss style. Never use `fmt.Sprintf` width formatting on pre-styled strings.
**Warning signs:** Calendar grid columns misalign, especially for single-digit vs double-digit days.
**Confidence:** HIGH (fundamental ANSI/terminal behavior)

### Pitfall 4: Holiday Check Using Wrong Time Zone

**What goes wrong:** `cal.IsHoliday` checks the date in the provided timezone. If you pass `time.UTC` but the holiday is defined for a local date, you might get wrong results near midnight.
**Why it happens:** A date at `00:00 UTC` might be the previous day in UTC-N timezones.
**How to avoid:** Always construct dates with `time.Local` for holiday checks. Use noon (`12:00`) as the time to avoid any midnight edge cases:
```go
date := time.Date(year, month, day, 12, 0, 0, 0, time.Local)
```
**Warning signs:** Holidays appear on wrong days near timezone boundaries.
**Confidence:** MEDIUM (logical deduction from timezone behavior; not directly documented as a rickar/cal pitfall)

### Pitfall 5: Config File Missing Crashes App

**What goes wrong:** `toml.DecodeFile` returns an error if the file doesn't exist. If not handled, the app panics on first run.
**Why it happens:** No config file exists until the user creates one.
**How to avoid:** Check `os.Stat` before `DecodeFile`. Return default config when file is missing. Only propagate errors for malformed TOML (syntax errors).
**Warning signs:** App crashes on first launch with "no such file or directory".
**Confidence:** HIGH (standard file handling pattern)

### Pitfall 6: Lip Gloss Width() Excludes Borders in v1

**What goes wrong:** In lipgloss v1.1.0, `style.Width(24)` sets content width to 24 (including padding but EXCLUDING borders). The total outer width is 24 + 2 (borders) = 26 chars. Developers expecting `Width` to be the total outer width will miscalculate layout.
**Why it happens:** This was a known issue fixed only in lipgloss v2.0.0 (issue #449).
**How to avoid:** The current Phase 1 code already handles this correctly -- `Width(calendarInnerWidth)` with `calendarInnerWidth=24` gives 24 chars including padding (22 content + 2 padding), plus 2 for borders = 26 outer. The 22-char content area comfortably fits the 20-char `cal` grid. Do not change the existing width calculation.
**Warning signs:** Layout overflow or unexpected wrapping.
**Confidence:** HIGH (verified via lipgloss issue #449, confirmed fix is v2-only)

## Code Examples

### Calendar Math: First Weekday and Days in Month

```go
// Source: Go time package (https://pkg.go.dev/time)

// firstWeekday returns the weekday (0-6) of the first day of the month.
// If mondayStart is true, Monday=0..Sunday=6.
// If mondayStart is false, Sunday=0..Saturday=6.
func firstWeekday(year int, month time.Month, mondayStart bool) int {
    wd := int(time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Weekday())
    if mondayStart {
        wd = (wd + 6) % 7 // rotate Sunday(0) to 6, Monday(1) to 0
    }
    return wd
}

// daysInMonth returns the number of days in the given month.
func daysInMonth(year int, month time.Month) int {
    // Day 0 of the next month = last day of this month
    return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
```

### Holiday Lookup for a Month

```go
// Source: rickar/cal/v2 API (https://pkg.go.dev/github.com/rickar/cal/v2)

import (
    "time"
    "github.com/rickar/cal/v2"
)

// holidaysInMonth returns a set of day numbers that are holidays.
func holidaysInMonth(c *cal.Calendar, year int, month time.Month) map[int]bool {
    result := make(map[int]bool)
    days := daysInMonth(year, month)
    for day := 1; day <= days; day++ {
        // Use noon to avoid timezone edge cases
        date := time.Date(year, month, day, 12, 0, 0, 0, time.Local)
        actual, observed, _ := c.IsHoliday(date)
        if actual || observed {
            result[day] = true
        }
    }
    return result
}
```

### Country Holiday Loading

```go
// Source: rickar/cal/v2 country subpackages
// https://pkg.go.dev/github.com/rickar/cal/v2/fi
// https://pkg.go.dev/github.com/rickar/cal/v2/us

import (
    "github.com/rickar/cal/v2"
    "github.com/rickar/cal/v2/fi"
    "github.com/rickar/cal/v2/us"
)

// countryHolidays returns the holiday slice for a country code.
func countryHolidays(code string) ([]*cal.Holiday, bool) {
    registry := map[string][]*cal.Holiday{
        "fi": fi.Holidays,
        "us": us.Holidays,
        // Each country subpackage exports a Holidays slice
    }
    h, ok := registry[code]
    return h, ok
}

// Example: Finland holidays
// fi.Holidays contains 15 holidays:
//   fi.Uudenvuodenpaiva (New Year), fi.Loppiainen (Epiphany),
//   fi.Pitkaperjantai (Good Friday), fi.Paasiaispaiva (Easter),
//   fi.ToinenPaasiaispaiva (Easter Monday), fi.Vappu (May Day),
//   fi.Helatorstai (Ascension), fi.Helluntaipaiva (Pentecost),
//   fi.Juhannusaatto (Midsummer Eve), fi.Juhannuspaiva (Midsummer),
//   fi.Pyhainpaiva (All Saints), fi.Itsenaisyyspaiva (Independence Day),
//   fi.Jouluaatto (Christmas Eve), fi.Joulupaiva (Christmas),
//   fi.Tapaninpaiva (St. Stephen's Day)
```

### TOML Config File Parsing

```go
// Source: BurntSushi/toml v1.6.0
// https://pkg.go.dev/github.com/BurntSushi/toml

import "github.com/BurntSushi/toml"

type Config struct {
    Country     string `toml:"country"`
    MondayStart bool   `toml:"monday_start"`
}

func loadConfig(path string) (Config, error) {
    cfg := Config{
        Country:     "us",
        MondayStart: false,
    }
    _, err := toml.DecodeFile(path, &cfg)
    return cfg, err
}
```

### Calendar Key Bindings

```go
// Source: bubbles/key v0.21.1
// https://pkg.go.dev/github.com/charmbracelet/bubbles@v0.21.1/key

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
    PrevMonth key.Binding
    NextMonth key.Binding
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        PrevMonth: key.NewBinding(
            key.WithKeys("left", "h"),
            key.WithHelp("<-/h", "prev month"),
        ),
        NextMonth: key.NewBinding(
            key.WithKeys("right", "l"),
            key.WithHelp("->/l", "next month"),
        ),
    }
}
```

### Width Budget Analysis

```
Total calendar pane outer width in v1.1.0:
  Width(24) = 24 chars (content + padding, excluding borders)
  + 2 chars border (left + right) = 26 chars total outer

Content area calculation:
  24 (Width value) - 2 (Padding 0,1 = left 1 + right 1) = 22 chars usable content

cal output width: 20 chars ("Su Mo Tu We Th Fr Sa")
Available: 22 chars
Surplus: 2 chars (calendar is left-aligned within content area, 2 chars right margin)

Grid layout:
  7 columns x 3 chars each = 21 chars
  Minus trailing space on last column = 20 chars per row
  Header "Su Mo Tu We Th Fr Sa" = 20 chars
  Title "   February 2026   " = 20 chars (centered)

All rows fit within 22-char content area. No wrapping issues.
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `rickar/cal` v1 (single package) | `rickar/cal/v2` (subpackages by ISO code) | v2.0.0 | Smaller binaries, per-country imports, observation rules per-holiday |
| `BurntSushi/toml` v0.x | `BurntSushi/toml` v1.6.0 | 2022+ | TOML v1.0.0/v1.1.0 compliance, better error reporting |
| Manual `~/.config` paths | `os.UserConfigDir()` | Go 1.12+ | Cross-platform XDG compliance built into stdlib |
| lipgloss `Width()` includes borders | `Width()` excludes borders in v1 | v2.0.0 will fix | In v1.1.0, Width sets content+padding width; borders are additional |

**Deprecated/outdated:**
- `rickar/cal` v1 (import `github.com/rickar/cal`): Replaced by v2; v1 has calendar-wide observation rules instead of per-holiday, no ISO subpackages
- `pelletier/go-toml` v1: Replaced by v2; the v1 API is different and deprecated
- `ioutil.ReadFile`: Use `os.ReadFile` (Go 1.16+)

## Open Questions

1. **How many countries to include in the registry?**
   - What we know: `rickar/cal/v2` has 48 country subpackages. Each import adds to binary size.
   - What's unclear: Whether to import all 48 or start with a subset.
   - Recommendation: Start with a practical subset (5-10 countries: fi, us, gb, de, se, no, dk, fr, es, it). Add more on request. Each import is small (just holiday definitions), so including all 48 is also viable if binary size is not a concern.

2. **Should the config file be auto-created with defaults on first run?**
   - What we know: The app works with defaults when no config file exists.
   - What's unclear: Whether users expect a config file to be created automatically.
   - Recommendation: Do NOT auto-create. Document the config file location and format. Users who want holidays create the file manually. This avoids writing files unexpectedly.

3. **Monday vs Sunday start default**
   - What we know: US convention is Sunday start, most European countries use Monday start. Finnish `cal -m` shows Monday start.
   - What's unclear: What the "right" default is for a configurable app.
   - Recommendation: Default to `monday_start = false` (Sunday, matching `cal` default). The TOML config allows users to change this. Alternatively, consider detecting locale, but that adds complexity for little gain.

4. **Should both actual and observed holiday dates be highlighted?**
   - What we know: `rickar/cal/v2` `IsHoliday` returns both `actual` and `observed` booleans. Some holidays are observed on a different day (e.g., if July 4 falls on Saturday, observed on Friday).
   - What's unclear: Whether to show the actual date, the observed date, or both.
   - Recommendation: Highlight BOTH actual and observed dates. The user sees when the holiday "really" is and when it's observed. Both are useful for planning.

## Sources

### Primary (HIGH confidence)
- [rickar/cal/v2 pkg.go.dev](https://pkg.go.dev/github.com/rickar/cal/v2) -- Full API: Calendar, Holiday, IsHoliday, country subpackages, v2.1.27
- [rickar/cal GitHub](https://github.com/rickar/cal) -- README, releases, 401 stars, 48 country subpackages confirmed
- [rickar/cal/v2/fi pkg.go.dev](https://pkg.go.dev/github.com/rickar/cal/v2/fi) -- All 15 Finnish holidays verified with variable names and Holidays slice
- [rickar/cal/v2/us pkg.go.dev](https://pkg.go.dev/github.com/rickar/cal/v2/us) -- 12 US holidays verified with Holidays slice
- [BurntSushi/toml pkg.go.dev](https://pkg.go.dev/github.com/BurntSushi/toml) -- DecodeFile, Decode, Marshal, struct tags, v1.6.0 API verified
- [Go time package](https://pkg.go.dev/time) -- Weekday constants (Sunday=0), Date normalization, AddDate, Month type
- [lipgloss issue #449](https://github.com/charmbracelet/lipgloss/issues/449) -- Confirmed Width() excludes borders in v1, fix scheduled for v2

### Secondary (MEDIUM confidence)
- [pelletier/go-toml/v2 pkg.go.dev](https://pkg.go.dev/github.com/pelletier/go-toml/v2) -- v2.2.4, compared API and adoption against BurntSushi/toml
- [lkyuchukov/go-cal GitHub](https://github.com/lkyuchukov/go-cal) -- Terminal calendar rendering approach in Go (reference for grid algorithm)
- [lipgloss issue #298](https://github.com/charmbracelet/lipgloss/issues/298) -- Width() includes padding but not borders in v1

### Tertiary (LOW confidence)
- None -- all findings verified with primary or secondary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- `rickar/cal/v2` API verified on pkg.go.dev including country subpackages and IsHoliday signature; `BurntSushi/toml` API verified with DecodeFile; Go `time` package is stdlib
- Architecture: HIGH -- patterns follow established Bubble Tea composition from Phase 1; grid algorithm verified against `cal` output format (20 chars, 7 columns)
- Pitfalls: HIGH (5/6), MEDIUM (1/6) -- timezone pitfall is logical deduction, all others verified from documentation
- Code examples: HIGH -- all API calls verified against pkg.go.dev documentation for specific library versions

**Width budget:** Verified by measuring actual `cal` output (20 chars) against lipgloss v1.1.0 Width behavior (22 chars content area). 2-char surplus confirmed.

**Research date:** 2026-02-05
**Valid until:** 2026-03-07 (30 days -- stable libraries, no imminent breaking changes)
