# Phase 6: Themes - Research

**Researched:** 2026-02-05
**Domain:** Lipgloss color theming for Bubble Tea TUI
**Confidence:** HIGH

## Summary

This phase adds four preset color themes (Dark, Light, Nord, Solarized) to the todo-calendar app. The app currently hardcodes ~12 distinct style definitions across three `styles.go` files using Lipgloss `Color()` values and style attributes (Bold, Faint, Reverse, Strikethrough). The theming approach is straightforward: define a `Theme` struct holding all color values, create four preset instances, load the theme name from `config.toml`, and replace all package-level `var` style definitions with functions that accept the theme.

The standard pattern in the Charm ecosystem (used by `charmbracelet/huh`) is: define a base theme struct with color fields, build variant themes by populating different color values, and construct Lipgloss styles from those colors at initialization. Since this app uses `lipgloss.Color` (hex strings), not `AdaptiveColor`, and themes are preset (not auto-detected), the implementation is simpler than adaptive theming -- each theme is just a fixed set of hex color strings.

**Primary recommendation:** Create a single `Theme` struct in a new `internal/theme` package with one color field per semantic role (e.g., `Border`, `BorderFocused`, `Holiday`, `Indicator`, `Accent`, `Muted`, `HeaderFg`), define four preset constructors, wire theme selection through `config.toml`, and convert the three `styles.go` files to accept the theme.

## Standard Stack

No new dependencies required. Theming uses only the existing Lipgloss API.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| lipgloss | v1.1.0 | Style definitions with color support | Already in use; `lipgloss.Color()` with hex strings is the standard approach |
| BurntSushi/toml | v1.6.0 | Config parsing for theme field | Already in use; theme is a new string field in config.toml |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| (none) | - | - | No new dependencies needed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hex color strings (`lipgloss.Color`) | `lipgloss.AdaptiveColor` | AdaptiveColor auto-detects light/dark terminal background, but we have explicit theme selection so it adds complexity without benefit |
| `lipgloss.CompleteColor` | `lipgloss.Color` | CompleteColor gives per-profile control (TrueColor/ANSI256/ANSI). Not needed since Lipgloss auto-degrades hex to nearest ANSI256 |

## Architecture Patterns

### Recommended Project Structure
```
internal/
  theme/
    theme.go          # Theme struct + 4 preset constructors
  app/
    styles.go         # Converts Theme -> app-level lipgloss.Style (focused/unfocused borders)
    model.go          # Stores theme, passes to children
  calendar/
    styles.go         # Converts Theme -> calendar lipgloss.Style (header, today, holiday, etc.)
    model.go          # Receives theme at construction
    grid.go           # Uses styles from model (not package-level vars)
  todolist/
    styles.go         # Converts Theme -> todolist lipgloss.Style (cursor, section header, etc.)
    model.go          # Receives theme at construction
  config/
    config.go         # New `Theme string` field with "dark" default
```

### Pattern 1: Theme as Value Struct with Semantic Color Roles

**What:** A plain Go struct with one `lipgloss.Color` field per semantic UI role. No methods, no interfaces -- just data.

**When to use:** When themes are preset and the color set is small (< 20 colors).

**Example:**
```go
// internal/theme/theme.go
package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines all colors used throughout the application.
// Each field represents a semantic UI role, not a specific component.
type Theme struct {
    // Panel borders
    BorderFocused   lipgloss.Color
    BorderUnfocused lipgloss.Color

    // Calendar
    HeaderFg        lipgloss.Color // month/year title
    WeekdayFg       lipgloss.Color // day-of-week header row
    TodayFg         lipgloss.Color // today highlight foreground
    TodayBg         lipgloss.Color // today highlight background
    HolidayFg       lipgloss.Color // holiday text color
    IndicatorFg     lipgloss.Color // bracket indicator for days with todos

    // Todo list
    AccentFg        lipgloss.Color // section headers, cursor, active elements
    MutedFg         lipgloss.Color // dates, secondary info
    CompletedFg     lipgloss.Color // done todos (faint + strikethrough)
    EmptyFg         lipgloss.Color // placeholder text

    // General
    NormalFg        lipgloss.Color // default text (empty string = terminal default)
    NormalBg        lipgloss.Color // default background
}
```

### Pattern 2: Preset Theme Constructors

**What:** Each theme is a function returning a populated `Theme` struct.

**Example:**
```go
func Dark() Theme {
    return Theme{
        BorderFocused:   lipgloss.Color("#7B68EE"), // medium slate blue (ANSI 62 equivalent)
        BorderUnfocused: lipgloss.Color("#585858"), // gray (ANSI 240 equivalent)
        HeaderFg:        lipgloss.Color("#FFFFFF"),
        // ...
    }
}

func Nord() Theme {
    return Theme{
        BorderFocused:   lipgloss.Color("#88C0D0"), // nord8 frost
        BorderUnfocused: lipgloss.Color("#4C566A"), // nord3 polar night
        HeaderFg:        lipgloss.Color("#ECEFF4"), // nord6 snow storm
        // ...
    }
}
```

### Pattern 3: Styles Struct Built from Theme

**What:** Each package defines a styles struct holding pre-built `lipgloss.Style` values, constructed from a Theme at init time.

**Example:**
```go
// internal/calendar/styles.go
package calendar

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/antti/todo-calendar/internal/theme"
)

type Styles struct {
    Header     lipgloss.Style
    WeekdayHdr lipgloss.Style
    Normal     lipgloss.Style
    Today      lipgloss.Style
    Holiday    lipgloss.Style
    Indicator  lipgloss.Style
}

func NewStyles(t theme.Theme) Styles {
    return Styles{
        Header:     lipgloss.NewStyle().Bold(true).Foreground(t.HeaderFg),
        WeekdayHdr: lipgloss.NewStyle().Foreground(t.WeekdayFg),
        Normal:     lipgloss.NewStyle().Foreground(t.NormalFg),
        Today:      lipgloss.NewStyle().Bold(true).Foreground(t.TodayFg).Background(t.TodayBg),
        Holiday:    lipgloss.NewStyle().Foreground(t.HolidayFg),
        Indicator:  lipgloss.NewStyle().Bold(true).Foreground(t.IndicatorFg),
    }
}
```

### Pattern 4: Theme Flows Through Constructor DI

**What:** Theme is passed from `main.go` -> `app.New()` -> child `New()` functions. Matches the existing pattern used for `store`, `provider`, and `mondayStart`.

**Example:**
```go
// main.go
t := theme.ForName(cfg.Theme) // returns Theme struct
model := app.New(provider, cfg.MondayStart(), s, t)

// app.New passes theme to children
func New(provider *holidays.Provider, mondayStart bool, s *store.Store, t theme.Theme) Model {
    cal := calendar.New(provider, mondayStart, s, t)
    tl := todolist.New(s, t)
    // ...
}
```

### Anti-Patterns to Avoid

- **Global mutable theme variable:** Do not use a package-level `var CurrentTheme theme.Theme` that styles read from. This creates hidden coupling and makes testing harder. Pass theme through constructors.
- **Styles as theme fields:** Do not put `lipgloss.Style` values in the Theme struct. Theme should hold colors only; styles are built per-package from colors. This keeps Theme serialization-friendly and package autonomy intact.
- **Rebuilding styles on every render:** Build `lipgloss.Style` values once at construction time, store them in a Styles struct on the model. Do not call `lipgloss.NewStyle().Foreground(...)` inside `View()`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Color degradation for limited terminals | Custom ANSI fallback mapping | `lipgloss.Color("#hex")` | Lipgloss auto-degrades hex to nearest ANSI256/ANSI color via termenv |
| Terminal background detection | Custom dark/light detection | Explicit theme selection in config | User picks their theme; auto-detection is unreliable and out of scope |
| Color parsing/validation | Custom hex parser | `lipgloss.Color()` accepts hex and ANSI strings | Already handles all formats |

**Key insight:** Lipgloss handles all color profile degradation automatically. Using `lipgloss.Color("#88C0D0")` works correctly on TrueColor, ANSI256, and ANSI terminals -- Lipgloss finds the nearest available color. No need to specify per-profile fallbacks with `CompleteColor`.

## Common Pitfalls

### Pitfall 1: Today Highlight Losing Visibility
**What goes wrong:** The current `todayStyle` uses `.Reverse(true)` which inverts fg/bg. With themed backgrounds, reverse may produce illegible results (e.g., light fg reversed on light bg = invisible).
**Why it happens:** `.Reverse(true)` depends on the terminal's default colors, not on theme colors.
**How to avoid:** Replace `.Reverse(true)` with explicit `.Foreground(t.TodayFg).Background(t.TodayBg)` so the theme controls both sides.
**Warning signs:** Today's date disappearing or blending into background on light themes.

### Pitfall 2: Faint/Strikethrough Behavior Varies
**What goes wrong:** `.Faint(true)` dims the text, but the dimming amount depends on the terminal emulator. On some terminals with light themes, faint text becomes invisible.
**Why it happens:** Faint is a terminal attribute, not a color -- it halves brightness. On a white background, dimmed white = light gray, which may be invisible.
**How to avoid:** For `completedStyle` and `emptyStyle`, use explicit foreground color from the theme instead of relying solely on `.Faint(true)`. Can keep `.Faint(true)` as a supplement for dark themes but always set a foreground color as the baseline.
**Warning signs:** Completed todos or empty placeholders vanishing on light/solarized themes.

### Pitfall 3: Package-Level Var Styles Break Theming
**What goes wrong:** If styles remain as package-level `var` declarations (current state), they are initialized once at import time before the theme is known.
**Why it happens:** Go `var` blocks run at package init, before `main()` runs and reads config.
**How to avoid:** Convert package-level `var` styles to a `Styles` struct created by a `NewStyles(theme)` function called after config is loaded. Store `Styles` on the model.
**Warning signs:** All themes rendering identically (using whatever default the `var` block picked).

### Pitfall 4: Forgetting to Theme the Help Bar
**What goes wrong:** The `help.Model` from `charmbracelet/bubbles` has its own styles that render independently.
**Why it happens:** The help bar uses `help.Styles` which has its own color settings (key color, description color, separator).
**How to avoid:** After creating `help.New()`, override `help.Styles` to use theme colors: `h.Styles.ShortKey`, `h.Styles.ShortDesc`, `h.Styles.ShortSeparator`, etc.
**Warning signs:** Help bar renders in default colors while the rest of the app is themed.

### Pitfall 5: Invalid Theme Name Silently Uses Wrong Colors
**What goes wrong:** User types `theme = "nrod"` (typo) in config.toml and gets unexpected behavior.
**Why it happens:** No validation of theme name string.
**How to avoid:** Validate theme name in the `theme.ForName()` function; return the Dark theme as default for unknown names. Optionally log a warning.
**Warning signs:** User reports "theme doesn't work" when they have a typo.

## Code Examples

### Complete Theme Definition (Dark -- Maps Current Hardcoded Colors)

```go
// Source: Derived from current codebase hardcoded values
func Dark() Theme {
    return Theme{
        BorderFocused:   lipgloss.Color("#5F5FD7"), // ANSI 62
        BorderUnfocused: lipgloss.Color("#585858"), // ANSI 240
        HeaderFg:        lipgloss.Color(""),         // terminal default (bold applied separately)
        WeekdayFg:       lipgloss.Color(""),         // terminal default (faint applied separately)
        TodayFg:         lipgloss.Color(""),         // uses reverse
        TodayBg:         lipgloss.Color(""),         // uses reverse
        HolidayFg:       lipgloss.Color("#AF0000"), // ANSI 1 (red)
        IndicatorFg:     lipgloss.Color(""),         // terminal default (bold applied separately)
        AccentFg:        lipgloss.Color("#5F5FD7"), // ANSI 62
        MutedFg:         lipgloss.Color("#585858"), // ANSI 240
        CompletedFg:     lipgloss.Color("#585858"), // dim gray
        EmptyFg:         lipgloss.Color("#585858"), // dim gray
        NormalFg:        lipgloss.Color(""),         // terminal default
        NormalBg:        lipgloss.Color(""),         // terminal default
    }
}
```

### Light Theme Definition

```go
func Light() Theme {
    return Theme{
        BorderFocused:   lipgloss.Color("#5F5FD7"), // purple accent
        BorderUnfocused: lipgloss.Color("#BCBCBC"), // light gray
        HeaderFg:        lipgloss.Color("#303030"), // dark gray
        WeekdayFg:       lipgloss.Color("#8A8A8A"), // medium gray
        TodayFg:         lipgloss.Color("#FFFFFF"), // white on dark bg
        TodayBg:         lipgloss.Color("#5F5FD7"), // purple highlight
        HolidayFg:       lipgloss.Color("#D70000"), // red
        IndicatorFg:     lipgloss.Color("#005FAF"), // blue
        AccentFg:        lipgloss.Color("#5F5FD7"), // purple
        MutedFg:         lipgloss.Color("#8A8A8A"), // gray
        CompletedFg:     lipgloss.Color("#BCBCBC"), // light gray
        EmptyFg:         lipgloss.Color("#8A8A8A"), // medium gray
        NormalFg:        lipgloss.Color("#303030"), // dark text
        NormalBg:        lipgloss.Color(""),         // terminal default
    }
}
```

### Nord Theme Definition

```go
// Source: https://www.nordtheme.com/docs/colors-and-palettes
func Nord() Theme {
    return Theme{
        // Polar Night: #2E3440, #3B4252, #434C5E, #4C566A
        // Snow Storm:  #D8DEE9, #E5E9F0, #ECEFF4
        // Frost:       #8FBCBB, #88C0D0, #81A1C1, #5E81AC
        // Aurora:      #BF616A, #D08770, #EBCB8B, #A3BE8C, #B48EAD
        BorderFocused:   lipgloss.Color("#88C0D0"), // nord8 frost
        BorderUnfocused: lipgloss.Color("#4C566A"), // nord3 polar night
        HeaderFg:        lipgloss.Color("#ECEFF4"), // nord6 snow storm bright
        WeekdayFg:       lipgloss.Color("#4C566A"), // nord3 muted
        TodayFg:         lipgloss.Color("#2E3440"), // nord0 dark bg as fg
        TodayBg:         lipgloss.Color("#88C0D0"), // nord8 frost
        HolidayFg:       lipgloss.Color("#BF616A"), // nord11 aurora red
        IndicatorFg:     lipgloss.Color("#A3BE8C"), // nord14 aurora green
        AccentFg:        lipgloss.Color("#88C0D0"), // nord8 frost
        MutedFg:         lipgloss.Color("#4C566A"), // nord3 polar night
        CompletedFg:     lipgloss.Color("#4C566A"), // nord3
        EmptyFg:         lipgloss.Color("#4C566A"), // nord3
        NormalFg:        lipgloss.Color("#D8DEE9"), // nord4 snow storm
        NormalBg:        lipgloss.Color(""),         // terminal default
    }
}
```

### Solarized Theme Definition

```go
// Source: https://ethanschoonover.com/solarized/
func Solarized() Theme {
    return Theme{
        // Base:    #002B36(bg), #073642(bg-hl), #586E75(muted), #657B83(secondary)
        //         #839496(body), #93A1A1(emphasis), #EEE8D5(light-bg), #FDF6E3(light-bg)
        // Accent: yellow #B58900, orange #CB4B16, red #DC322F, magenta #D33682
        //         violet #6C71C4, blue #268BD2, cyan #2AA198, green #859900
        BorderFocused:   lipgloss.Color("#268BD2"), // solarized blue
        BorderUnfocused: lipgloss.Color("#586E75"), // base01 muted
        HeaderFg:        lipgloss.Color("#93A1A1"), // base1 emphasis
        WeekdayFg:       lipgloss.Color("#586E75"), // base01
        TodayFg:         lipgloss.Color("#FDF6E3"), // base3 lightest
        TodayBg:         lipgloss.Color("#268BD2"), // blue
        HolidayFg:       lipgloss.Color("#DC322F"), // solarized red
        IndicatorFg:     lipgloss.Color("#859900"), // solarized green
        AccentFg:        lipgloss.Color("#268BD2"), // blue
        MutedFg:         lipgloss.Color("#586E75"), // base01
        CompletedFg:     lipgloss.Color("#586E75"), // base01
        EmptyFg:         lipgloss.Color("#586E75"), // base01
        NormalFg:        lipgloss.Color("#839496"), // base0 body text
        NormalBg:        lipgloss.Color(""),         // terminal default
    }
}
```

### Theme Selection from Config

```go
// internal/theme/theme.go
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
```

### Config Integration

```go
// internal/config/config.go
type Config struct {
    Country        string `toml:"country"`
    FirstDayOfWeek string `toml:"first_day_of_week"`
    Theme          string `toml:"theme"`
}

func DefaultConfig() Config {
    return Config{
        Country:        "us",
        FirstDayOfWeek: "sunday",
        Theme:          "dark",
    }
}
```

### Help Bar Theming

```go
// Source: charmbracelet/bubbles help package
import "github.com/charmbracelet/bubbles/help"

h := help.New()
h.Styles.ShortKey = lipgloss.NewStyle().Foreground(t.AccentFg)
h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)
```

## Current Color Inventory

Complete list of hardcoded colors that must be converted to theme values:

| File | Style | Current Value | Semantic Role |
|------|-------|---------------|---------------|
| `app/styles.go` | `focusedBorderColor` | `Color("62")` (purple) | `BorderFocused` |
| `app/styles.go` | `unfocusedBorderColor` | `Color("240")` (gray) | `BorderUnfocused` |
| `calendar/styles.go` | `headerStyle` | `.Bold(true)` (no color) | `HeaderFg` + Bold |
| `calendar/styles.go` | `weekdayHdrStyle` | `.Faint(true)` (no color) | `WeekdayFg` |
| `calendar/styles.go` | `normalStyle` | (empty style) | `NormalFg` |
| `calendar/styles.go` | `todayStyle` | `.Bold(true).Reverse(true)` | `TodayFg` + `TodayBg` + Bold |
| `calendar/styles.go` | `holidayStyle` | `.Foreground(Color("1"))` (red) | `HolidayFg` |
| `calendar/styles.go` | `indicatorStyle` | `.Bold(true)` (no color) | `IndicatorFg` + Bold |
| `todolist/styles.go` | `sectionHeaderStyle` | `.Bold(true).Foreground(Color("62"))` | `AccentFg` + Bold |
| `todolist/styles.go` | `completedStyle` | `.Faint(true).Strikethrough(true)` | `CompletedFg` + Strikethrough |
| `todolist/styles.go` | `cursorStyle` | `.Foreground(Color("62"))` | `AccentFg` |
| `todolist/styles.go` | `dateStyle` | `.Foreground(Color("240"))` | `MutedFg` |
| `todolist/styles.go` | `emptyStyle` | `.Faint(true)` | `EmptyFg` |
| `app/model.go` | `help.New()` | default help styles | `AccentFg`, `MutedFg` |

**Total: 13 style definitions + 1 help bar = 14 items to theme.**

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Package-level `var` styles with hardcoded colors | Styles struct built from Theme at init | Charm ecosystem convention as of huh v0.3+ | Enables runtime theme selection |
| `.Reverse(true)` for today highlight | Explicit fg/bg from theme | Best practice for themed apps | Consistent appearance across themes |
| `.Faint(true)` for dim text | Explicit foreground color + optional `.Faint(true)` | Needed for light themes | Prevents invisible text on light backgrounds |

**Deprecated/outdated:**
- `lipgloss.SetColorProfile()`: Still works in v1.x but renderers handle this automatically. No need to call it.

## Open Questions

1. **Background color handling**
   - What we know: The current app uses transparent background (terminal default). The `NormalBg` field is empty for all themes.
   - What's unclear: Whether any theme should set an explicit background color (e.g., Nord could set `#2E3440` to guarantee the right look).
   - Recommendation: Keep `NormalBg` empty (terminal default) for all themes. Setting explicit backgrounds causes issues with terminal transparency and padding areas. Users are expected to use terminal themes that complement their chosen app theme.

2. **Help bar style fields**
   - What we know: `help.Styles` has `ShortKey`, `ShortDesc`, `ShortSeparator`, `FullKey`, `FullDesc`, `FullSeparator`, `Ellipsis` fields.
   - What's unclear: Whether we use full help view at all (currently short only).
   - Recommendation: Theme only the `Short*` styles since the app uses `ShortHelp()` only.

## Sources

### Primary (HIGH confidence)
- Lipgloss v1.1.0 API: `lipgloss.Color`, `lipgloss.NewStyle()`, style methods -- verified via [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/lipgloss) and [GitHub](https://github.com/charmbracelet/lipgloss)
- Codebase audit: All 17 Go source files read directly, all color usages catalogued
- Nord palette: Official [nordtheme.com/docs/colors-and-palettes](https://www.nordtheme.com/docs/colors-and-palettes) -- 16 colors verified
- Solarized palette: Official [ethanschoonover.com/solarized](https://ethanschoonover.com/solarized/) -- 16 colors verified

### Secondary (MEDIUM confidence)
- Theme struct pattern: charmbracelet/huh [theme.go](https://github.com/charmbracelet/huh/blob/main/theme.go) -- verified as real-world Charm ecosystem pattern
- Help bar styling: Derived from bubbles help package convention

### Tertiary (LOW confidence)
- None -- all findings verified with primary or secondary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, uses existing Lipgloss API verified via official docs
- Architecture: HIGH -- pattern directly from charmbracelet/huh, adapted to this app's simpler needs; verified against existing codebase structure
- Pitfalls: HIGH -- derived from direct code audit (Reverse, Faint, package-level vars) and verified Lipgloss behavior
- Color palettes: HIGH -- Nord and Solarized hex values from official sources

**Research date:** 2026-02-05
**Valid until:** 2026-03-05 (stable domain, Lipgloss v1.x API unlikely to change)
