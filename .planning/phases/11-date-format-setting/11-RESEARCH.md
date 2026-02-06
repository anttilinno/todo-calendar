# Phase 11: Date Format Setting - Research

**Researched:** 2026-02-06
**Domain:** Go time formatting, settings UI extension, config persistence
**Confidence:** HIGH

## Summary

This phase adds a date format preference to the app's settings overlay, allowing users to cycle through three preset date formats (ISO, European, US). The implementation is well-scoped: it touches config, settings, todolist, and the app model, but leaves the store and calendar entirely unchanged. The existing settings overlay pattern (cycling options with left/right arrows, save/cancel) directly supports this feature with no new UX patterns needed.

The primary technical challenge is the separation of storage format from display format. Internally, all dates remain `YYYY-MM-DD` (`"2006-01-02"`) in the JSON store. The display format applies only in View/render functions and date input parsing. Go's `time.Format`/`time.Parse` layout system handles the conversion cleanly, but the layout strings are unintuitive (e.g., `"01/02/2006"` is MM/DD/YYYY because it IS January 2, 2006). The three presets eliminate any user exposure to Go layout strings.

A complete audit of the codebase shows dates are displayed in exactly one place: `todolist.renderTodo()` at line 478, which renders `t.Date` as a raw string with the `Date` style. Date input occurs in two modes: `dateInputMode` (adding new dated todos) and `editDateMode` (editing existing todo dates). Both currently validate against `"2006-01-02"` and show a `"YYYY-MM-DD"` placeholder. All three sites must be updated. The calendar grid shows day numbers only, and the overview shows month names only -- neither displays full dates, so neither needs changes.

**Primary recommendation:** Add a `DateFormat` string field to `config.Config` with preset identifiers (`"iso"`, `"eu"`, `"us"`), add a 4th settings row, propagate the Go layout string to todolist via a `SetDateFormat()` method, and convert dates at render time and input time.

## Standard Stack

No new libraries needed. This feature uses only Go standard library and existing project dependencies.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go `time` package | Go 1.25.6 | `time.Parse` and `time.Format` for date conversion | Built-in, no dependencies needed |
| `BurntSushi/toml` v1.6.0 | Already in go.mod | Config persistence of `date_format` field | Already used for all config |
| `charmbracelet/bubbletea` v1.3.10 | Already in go.mod | Settings overlay message passing | Already used for all UI |

### Supporting
No additional libraries needed.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Go layout strings for presets | strftime-style format specifiers | Would require a third-party library or custom parser; Go stdlib handles this natively |
| Preset identifiers in config | Raw Go layout string in config | Raw strings enable custom formats but are unintuitive for users; presets are clearer and REQUIREMENTS explicitly puts custom format out of scope for settings UI |

**Installation:** No new dependencies.

## Architecture Patterns

### Config Field Design

Store preset identifiers, not raw Go layout strings, in config.toml. This matches the existing pattern where `theme` stores `"dark"` / `"light"` / `"nord"` / `"solarized"` and `first_day_of_week` stores `"sunday"` / `"monday"`.

```toml
# config.toml
country = "fi"
first_day_of_week = "monday"
theme = "nord"
date_format = "iso"
```

The `Config` struct gains a `DateFormat` field and two helper methods:

```go
// config.go additions
type Config struct {
    Country        string `toml:"country"`
    FirstDayOfWeek string `toml:"first_day_of_week"`
    Theme          string `toml:"theme"`
    DateFormat     string `toml:"date_format"`
}

func DefaultConfig() Config {
    return Config{
        Country:        "us",
        FirstDayOfWeek: "sunday",
        Theme:          "dark",
        DateFormat:     "iso",  // YYYY-MM-DD, matches storage format
    }
}

// DateLayout returns the Go time layout string for the configured date format.
func (c Config) DateLayout() string {
    switch c.DateFormat {
    case "eu":
        return "02.01.2006" // DD.MM.YYYY
    case "us":
        return "01/02/2006" // MM/DD/YYYY -- NOTE: this IS the Go reference date
    default:
        return "2006-01-02" // YYYY-MM-DD (ISO)
    }
}

// DatePlaceholder returns a human-readable placeholder for date input prompts.
func (c Config) DatePlaceholder() string {
    switch c.DateFormat {
    case "eu":
        return "DD.MM.YYYY"
    case "us":
        return "MM/DD/YYYY"
    default:
        return "YYYY-MM-DD"
    }
}
```

**Confidence:** HIGH -- follows established patterns in the existing codebase.

### Date Conversion Helper

A single function handles storage-to-display conversion. Place it in the `config` package or as a standalone helper in `todolist`:

```go
// FormatDate converts an ISO date string ("2006-01-02") to the display format.
// Returns the original string unchanged if parsing fails.
func FormatDate(isoDate, layout string) string {
    t, err := time.Parse("2006-01-02", isoDate)
    if err != nil {
        return isoDate // graceful fallback
    }
    return t.Format(layout)
}

// ParseUserDate parses a date string in the display format and returns the ISO
// storage format. Returns an error if the input cannot be parsed.
func ParseUserDate(input, layout string) (string, error) {
    t, err := time.Parse(layout, input)
    if err != nil {
        return "", err
    }
    return t.Format("2006-01-02"), nil
}
```

**Confidence:** HIGH -- Go `time.Parse`/`time.Format` are well-documented and this is the standard approach.

### Settings Overlay Extension

Add a 4th option row to the settings overlay using the existing `option` struct pattern. The display values should include a preview of today's date in each format for clarity.

```go
// In settings.New(), add after the dayValues/dayDisplay block:
formatValues := []string{"iso", "eu", "us"}
formatDisplay := []string{
    fmt.Sprintf("ISO (%s)", time.Now().Format("2006-01-02")),
    fmt.Sprintf("European (%s)", time.Now().Format("02.01.2006")),
    fmt.Sprintf("US (%s)", time.Now().Format("01/02/2006")),
}

// Add to options slice (index 3):
{label: "Date Format", values: formatValues, display: formatDisplay, index: indexOf(formatValues, cfg.DateFormat)},
```

The `Config()` method must be updated to include the DateFormat field from `m.options[3]`.

**Confidence:** HIGH -- the existing option cycling mechanism handles this without modification.

### Todolist Date Format Propagation

The todolist model needs a `dateLayout` field and a `SetDateFormat()` setter, following the `SetTheme()` pattern:

```go
// In todolist.Model:
dateLayout string // Go time layout for date display/input

// In todolist.New():
dateLayout: "2006-01-02", // default ISO

// Setter:
func (m *Model) SetDateFormat(layout string) {
    m.dateLayout = layout
}
```

Three call sites in todolist need to use the format:

1. **renderTodo()** -- display: `FormatDate(t.Date, m.dateLayout)` instead of raw `t.Date`
2. **updateDateInputMode()** -- input: parse with `ParseUserDate(input, m.dateLayout)`, placeholder shows `cfg.DatePlaceholder()`
3. **updateEditDateMode()** -- input: same parse logic, same placeholder

**Confidence:** HIGH -- the setter pattern is already established in the codebase.

### App Model Wiring

The app model must call `SetDateFormat()` on the todolist at:

1. **Initialization** (`app.New`): `tl.SetDateFormat(cfg.DateLayout())`
2. **Settings save** (`settings.SaveMsg` handler): `m.todoList.SetDateFormat(m.cfg.DateLayout())`

**Confidence:** HIGH -- follows the exact same pattern as `SetMondayStart()` and `SetTheme()`.

### File Modification Summary

| File | Changes |
|------|---------|
| `internal/config/config.go` | Add `DateFormat` field, update `DefaultConfig()`, add `DateLayout()` and `DatePlaceholder()` methods |
| `internal/settings/model.go` | Add 4th option row for date format, update `Config()` to include `DateFormat`, add `time` import |
| `internal/todolist/model.go` | Add `dateLayout` field, `SetDateFormat()` setter, update `renderTodo()`, `updateDateInputMode()`, `updateEditDateMode()` |
| `internal/app/model.go` | Call `SetDateFormat()` in `New()` and `SaveMsg` handler |

**Files NOT modified:** `store/`, `calendar/`, `theme/`, `holidays/`, `main.go`.

### Anti-Patterns to Avoid
- **Passing display format to store methods:** The store must never see the display format. All store operations use `"2006-01-02"` internally. The conversion happens only in View functions and input handling.
- **Storing Go layout strings in config:** Use preset identifiers (`"iso"`, `"eu"`, `"us"`) and convert to layout strings via `DateLayout()`. Raw layout strings are unintuitive and error-prone in TOML files.
- **Exposing custom format input in settings UI:** Requirements explicitly list "Custom date format in settings UI" as out of scope. Only the 3 presets are available via the overlay.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Date format conversion | Custom string manipulation (regex, substring replacement) | `time.Parse` + `time.Format` | Go's time package handles all edge cases (leap years, month boundaries, zero-padding) correctly |
| Settings option cycling | New cycling widget | Existing `option` struct in settings/model.go | The cycling mechanism already works perfectly for this use case |
| Config persistence | Manual TOML writing | Existing `config.Save()` with `BurntSushi/toml` encoder | Already handles atomic writes and correct TOML encoding |

**Key insight:** This feature requires zero new patterns or components. Every piece of the implementation slots into an existing codebase pattern. The only "new" thing is the `FormatDate`/`ParseUserDate` helper functions.

## Common Pitfalls

### Pitfall 1: Date Format Round-Trip Corruption
**What goes wrong:** Display format layout string accidentally used for store reads/writes. For example, storing `"06.02.2026"` (EU format for Feb 6) and later parsing it as ISO produces a parse error -- or worse, storing `"02/06/2026"` (US format for Feb 6) and parsing as `"01/02/2006"` silently reads it as February 6, but storing US-formatted `"06/02/2026"` and parsing as ISO reads it as nothing (error) or as June 2.
**Why it happens:** Go layout `"01/02/2006"` and `"02/01/2006"` look nearly identical in code. A developer might pass the wrong layout to `time.Parse`.
**How to avoid:** Hard rule: `time.Parse("2006-01-02", ...)` is the ONLY layout used when reading from the store. The display layout is used ONLY in `FormatDate()` for rendering and `ParseUserDate()` for input. The store never imports or references the display format. Add a comment at each conversion site.
**Warning signs:** `time.Parse` called with any layout other than `"2006-01-02"` on data read from store/JSON.

### Pitfall 2: Settings Index Off-By-One After Adding 4th Option
**What goes wrong:** The `Config()` method in `settings/model.go` currently accesses options by hardcoded indices: `m.options[0]` = theme, `m.options[1]` = country, `m.options[2]` = first day of week. Adding a 4th option at index 3 is straightforward, but if the new option is inserted in the middle instead of appended at the end, all subsequent indices shift.
**Why it happens:** The option list is ordered and index-based. The `Config()` method and the theme-change detection in `Update()` reference options by numeric index.
**How to avoid:** Always append the new "Date Format" option at the end of the options slice (index 3). Update `Config()` to read `m.options[3].values[m.options[3].index]` for DateFormat. Do not change the order of existing options.
**Warning signs:** After adding date format, theme cycling breaks (wrong theme applied) or country cycling returns wrong value.

### Pitfall 3: Date Input Placeholder Not Matching Active Format
**What goes wrong:** The date input placeholder still shows `"YYYY-MM-DD"` after user changes format to European. User types `06.02.2026` but the parser expects `"2006-01-02"` and rejects it.
**Why it happens:** The placeholder strings in `updateNormalMode()` (lines 275, 298) are hardcoded. The parse validation in `updateDateInputMode()` (line 332) and `updateEditDateMode()` (line 394) uses hardcoded `"2006-01-02"`.
**How to avoid:** Store the active date layout and placeholder in the todolist model. Update both placeholder strings and parse calls to use the model's `dateLayout` field. The `DatePlaceholder()` method on config provides the human-readable placeholder.
**Warning signs:** User changes date format in settings, but the date input prompt still says "YYYY-MM-DD".

### Pitfall 4: Invalid or Missing DateFormat in Config
**What goes wrong:** User manually edits config.toml and enters an invalid `date_format` value (e.g., `date_format = "european"` instead of `"eu"`). The `DateLayout()` method returns the default ISO layout, which is a safe fallback, but the settings overlay shows the wrong selection.
**Why it happens:** The `indexOf()` function returns 0 when the value is not found in the slice. This means an invalid config value silently defaults to the first option ("iso") in the settings overlay, which is actually correct behavior.
**How to avoid:** The existing `indexOf()` fallback to 0 is sufficient. The `DateLayout()` method's `default` case returns ISO format. No additional validation needed.
**Warning signs:** None -- the existing fallback behavior is correct.

### Pitfall 5: Edit Date Mode Pre-Populates With ISO Format, User Expects Display Format
**What goes wrong:** When editing a todo's date (`E` key), the input is pre-populated with `todo.Date` which is always in ISO format (`"2026-02-06"`). If the user's display format is European, they see `"2026-02-06"` in the input but expect `"06.02.2026"`. They might add characters to the end or try to edit it as European format.
**Why it happens:** `todo.Date` stores the ISO string, and the current code does `m.input.SetValue(todo.Date)` directly.
**How to avoid:** When entering editDateMode, convert the stored date to the display format before setting the input value: `m.input.SetValue(FormatDate(todo.Date, m.dateLayout))`. The parse logic then converts back to ISO on confirm.
**Warning signs:** User presses `E` to edit a date and sees ISO format despite having European format selected.

## Code Examples

### Converting Storage Date to Display Format

```go
// FormatDate converts an ISO date string to the active display format.
// Returns the original string unchanged if parsing fails (graceful degradation).
func FormatDate(isoDate, layout string) string {
    t, err := time.Parse("2006-01-02", isoDate)
    if err != nil {
        return isoDate
    }
    return t.Format(layout)
}
```

### Parsing User Input Back to Storage Format

```go
// ParseUserDate parses a date string in the user's display format
// and returns the canonical ISO storage format ("2006-01-02").
func ParseUserDate(input, layout string) (string, error) {
    t, err := time.Parse(layout, input)
    if err != nil {
        return "", err
    }
    return t.Format("2006-01-02"), nil
}
```

### Updated renderTodo With Date Formatting

```go
func (m Model) renderTodo(b *strings.Builder, t *store.Todo, selected bool) {
    if selected {
        b.WriteString(m.styles.Cursor.Render("> "))
    } else {
        b.WriteString("  ")
    }

    check := "[ ] "
    if t.Done {
        check = "[x] "
    }

    text := t.Text
    if t.HasDate() {
        displayDate := FormatDate(t.Date, m.dateLayout)
        text += " " + m.styles.Date.Render(displayDate)
    }

    if t.Done {
        b.WriteString(m.styles.Completed.Render(check + text))
    } else {
        b.WriteString(check + text)
    }
    b.WriteString("\n")
}
```

### Updated Date Input With Format-Aware Parsing

```go
// In updateDateInputMode:
case key.Matches(msg, m.keys.Confirm):
    date := strings.TrimSpace(m.input.Value())
    if date == "" {
        return m, nil
    }
    // Parse in user's configured format, convert to ISO for storage
    isoDate, err := ParseUserDate(date, m.dateLayout)
    if err != nil {
        return m, nil // invalid date -- stay in input mode
    }
    m.store.Add(m.pendingText, isoDate)
    m.mode = normalMode
    m.input.Blur()
    m.input.SetValue("")
    m.pendingText = ""
    return m, nil
```

### Settings Option With Date Preview

```go
formatValues := []string{"iso", "eu", "us"}
now := time.Now()
formatDisplay := []string{
    fmt.Sprintf("ISO (%s)", now.Format("2006-01-02")),
    fmt.Sprintf("European (%s)", now.Format("02.01.2006")),
    fmt.Sprintf("US (%s)", now.Format("01/02/2006")),
}
```

### Config() Update in Settings

```go
func (m Model) Config() config.Config {
    return config.Config{
        Theme:          m.options[0].values[m.options[0].index],
        Country:        m.options[1].values[m.options[1].index],
        FirstDayOfWeek: m.options[2].values[m.options[2].index],
        DateFormat:     m.options[3].values[m.options[3].index],
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hardcoded `t.Date` display (ISO only) | Format-aware display via `FormatDate()` | This phase | All date displays reflect user preference |
| Hardcoded `"YYYY-MM-DD"` input placeholder | Format-aware placeholder via `DatePlaceholder()` | This phase | Users know what format to type |
| Hardcoded `"2006-01-02"` input parsing | Format-aware parsing via `ParseUserDate()` | This phase | Users can type dates in their preferred format |
| 3-option settings overlay | 4-option settings overlay | This phase | Date format is user-configurable |

## Design Decision: Date Input Format

The prior milestone research contains a contradiction:

- **PITFALLS.md "Looks Done" #9:** Recommends keeping date input as `YYYY-MM-DD` always, regardless of display format. Rationale: avoids ambiguous input parsing.
- **ARCHITECTURE.md Feature 4:** Recommends adapting date input to the configured display format. Rationale: better UX -- users should type dates in the format they chose.

**Recommendation: Adapt input to display format.** Rationale:

1. The 3 presets are all unambiguous -- each uses different separators (`-`, `.`, `/`), so there is no parsing ambiguity between them.
2. Forcing ISO input while displaying EU format creates cognitive dissonance: user sees `06.02.2026` next to their todos but must type `2026-02-06` to add a new one.
3. The conversion is straightforward: `ParseUserDate(input, displayLayout)` returns ISO for storage.
4. The placeholder tells the user exactly what format to use (e.g., `"DD.MM.YYYY"`).

The only risk would be if custom format strings were allowed (where ambiguity could arise), but custom format in settings UI is explicitly out of scope.

## Open Questions

1. **Where to place FormatDate/ParseUserDate helpers?**
   - What we know: These are utility functions used by todolist model.
   - Options: (a) In `config` package alongside `DateLayout()`/`DatePlaceholder()`, (b) In `todolist` package as private helpers, (c) In a new `internal/dateutil` package.
   - Recommendation: Place in `config` package since they are closely related to the config's date format concept and may be needed by future components (search results display).

2. **Should the calendar header use the date format?**
   - What we know: The calendar header shows `"February 2026"` (month name + year). This is not a full date.
   - Answer: No. The date format setting applies only to full dates (day+month+year). Month name + year is unaffected. The prior research agrees on this.
   - Confidence: HIGH.

## Sources

### Primary (HIGH confidence)
- **Codebase audit** -- All source files in `internal/` read and analyzed for date display/input/storage points
- **Prior milestone research** -- `.planning/research/STACK.md`, `ARCHITECTURE.md`, `PITFALLS.md`, `FEATURES.md` contain detailed analysis of this exact feature
- **Go `time` package** -- Layout string system verified against Go standard library documentation (reference time: `Mon Jan 2 15:04:05 MST 2006`)

### Secondary (MEDIUM confidence)
- **Requirements** -- `.planning/REQUIREMENTS.md` defines DTFMT-01/02/03 and explicitly scopes out custom format in settings UI

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new libraries, all Go stdlib
- Architecture: HIGH -- follows established patterns in codebase, verified against prior research
- Pitfalls: HIGH -- prior research already identified the key pitfalls (round-trip corruption, input format, config fallback)
- Code examples: HIGH -- based on actual codebase patterns and Go stdlib documentation

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (stable domain, no external dependencies to change)
