# Stack Research

**Domain:** Go TUI calendar + todo application
**Researched:** 2026-02-05
**Confidence:** HIGH (v1 stack) / MEDIUM (v2 decision)

## Version Decision: Bubble Tea v1 vs v2

This is the most consequential stack decision. Here is the current state:

| Aspect | v1 (stable) | v2 (pre-release) |
|--------|-------------|-------------------|
| Latest version | v1.3.10 (Sep 2025) | v2.0.0-rc.2 (Nov 2025) |
| Stability | Stable, production-proven | RC stage, 77% milestone complete (6 open issues) |
| Import path | `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` |
| Ecosystem compat | Bubbles v0.21.1, Lip Gloss v1.1.0, Huh v0.8.0 | Bubbles v2.0.0-beta.1, Lip Gloss v2.0.0-beta.3 |
| View API | `View() string` | `View() tea.View` (declarative) |
| Renderer | Standard | Cursed renderer (ncurses-based, much faster) |
| Keyboard | Basic | Progressive enhancement, key release events |
| Mouse | `tea.MouseMsg` | Split into `MouseClickMsg`, `MouseReleaseMsg`, etc. |

**Recommendation: Use Bubble Tea v1 (stable).**

Rationale:
- v2 is still in RC with 6 open milestone issues and the main PR (#1118) in draft
- The companion libraries (Bubbles, Lip Gloss) are only at beta for v2, not RC
- For a greenfield project that ships in weeks, v1 is battle-tested with 10,000+ apps
- Migration to v2 later is straightforward (the team documented the path in Discussion #1374)
- v1 has everything this project needs: key handling, viewport, styling, list components

**When to reconsider:** If v2.0.0 stable drops before development begins, switch. The v2 declarative View API and cursed renderer are genuinely better, but not worth beta-quality companion libraries.

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| Go | 1.25.x (latest: 1.25.7) | Language runtime | Current stable release line; 1.24.x also supported but 1.25 is the primary supported line | HIGH |
| Bubble Tea | v1.3.10 | TUI framework (Elm Architecture) | De facto standard for Go TUIs with 10,000+ apps built; stable, well-documented, active maintenance | HIGH |
| Lip Gloss | v1.1.0 | Terminal styling and layout | Official companion to Bubble Tea; CSS-like declarative styling, color profiles, box model, horizontal/vertical joining | HIGH |
| Bubbles | v0.21.1 | Pre-built TUI components | Official component library; provides list, help, key bindings, viewport, text input -- all needed for this project | HIGH |

### Data & Storage

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| `encoding/json` (stdlib) | Go 1.25 stdlib | Todo data serialization | Zero dependencies; human-readable; every editor highlights it; sufficient for simple todo data | HIGH |
| BurntSushi/toml | v1.6.0 | Configuration file format | Native date/time types (perfect for calendar config); human-friendly for user-edited config; 36,500+ importers; TOML v1.1.0 compliant | HIGH |
| `os.UserConfigDir` (stdlib) | Go 1.25 stdlib | XDG-compliant config path | Returns `$XDG_CONFIG_HOME` on Linux, proper paths on macOS/Windows; no external dependency needed | HIGH |

### Holidays

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| rickar/cal | v2.1.27 | Holiday calculations | Built-in holiday definitions for 50+ countries including Finland (fi); no external API needed; supports exact days, floating days, Easter-relative offsets; actively maintained (Jan 2026 release) | HIGH |

### Supporting Libraries

| Library | Version | Purpose | When to Use | Confidence |
|---------|---------|---------|-------------|------------|
| charmbracelet/huh | v0.8.0 | Interactive form prompts | For todo add/edit dialogs if inline text input is insufficient; integrates with Bubble Tea as a model | MEDIUM |
| adrg/xdg | latest | Full XDG Base Directory spec | Only if `os.UserConfigDir` is insufficient (need data dir, cache dir separation); stdlib covers config dir | LOW |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `go vet` | Static analysis | Built-in, catches common mistakes |
| `go test` | Testing | Built-in test framework; use table-driven tests |
| `golangci-lint` | Comprehensive linting | Community standard meta-linter for Go projects |
| `goreleaser` | Build and release | If distributing binaries; cross-compilation for Linux/macOS/Windows |

## Installation

```bash
# Initialize module
go mod init github.com/antti/todo-calendar

# Core TUI framework
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/lipgloss@v1.1.0
go get github.com/charmbracelet/bubbles@v0.21.1

# Holidays
go get github.com/rickar/cal/v2@v2.1.27

# Configuration (TOML)
go get github.com/BurntSushi/toml@v1.6.0

# Optional: forms for todo input
go get github.com/charmbracelet/huh@v0.8.0
```

## Storage Format Decisions

### Todo Data: JSON

Store todo items in `~/.config/todo-calendar/todos.json` (or XDG equivalent).

**Why JSON over TOML for data:**
- Todos are structured data (arrays of objects), not configuration
- JSON marshaling is in Go stdlib (`encoding/json`) -- zero dependency
- Simpler programmatic read/write than TOML for dynamic data
- Every tool can inspect/debug JSON files

**Schema example:**
```json
{
  "todos": [
    {
      "id": "uuid-here",
      "text": "Doctor appointment",
      "date": "2026-02-15",
      "done": false,
      "created": "2026-02-05T10:30:00Z"
    },
    {
      "id": "uuid-here",
      "text": "Buy groceries",
      "date": "",
      "done": false,
      "created": "2026-02-05T10:31:00Z"
    }
  ]
}
```

Floating todos (no date) use an empty `date` field. Date-bound todos use `YYYY-MM-DD` format.

### Configuration: TOML

Store configuration in `~/.config/todo-calendar/config.toml`.

**Why TOML over JSON for config:**
- Supports comments (users can annotate their config)
- Native date types
- More human-readable for manual editing
- Standard for Go CLI tool configuration

**Schema example:**
```toml
# todo-calendar configuration

[calendar]
# ISO 3166-1 alpha-2 country code for holidays
country = "fi"
# First day of week: 0 = Sunday, 1 = Monday
first_day_of_week = 1

[display]
# Show week numbers in calendar
show_week_numbers = true
```

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| TUI framework | Bubble Tea v1 | Bubble Tea v2 (RC) | v2 companion libs still in beta; migration path exists when v2 stabilizes |
| TUI framework | Bubble Tea | tview/tcell | tview is imperative (not Elm Architecture); Bubble Tea's functional approach produces cleaner, more testable code |
| TUI framework | Bubble Tea | Termbox | Deprecated; last meaningful update years ago |
| Styling | Lip Gloss v1 | Manual ANSI codes | Lip Gloss handles color profile detection, box model, layout joining; manual codes are error-prone |
| Holidays | rickar/cal | go-holidays (Nager API) | go-holidays requires network API calls; rickar/cal has built-in definitions for 50+ countries including Finland |
| Holidays | rickar/cal | hardcoded holidays | rickar/cal handles floating holidays (Easter, Midsummer) correctly with complex date math |
| Data format | JSON | SQLite | Overkill for a flat list of todos; adds CGo dependency (or pure-Go driver complexity); JSON is sufficient for hundreds of items |
| Data format | JSON | TOML (for data) | TOML is better for config; JSON is simpler for programmatic read/write of structured data arrays |
| Config format | TOML | YAML | YAML has footguns (Norway problem, implicit type coercion); TOML is simpler and safer for config |
| Config format | TOML | JSON | JSON has no comments; users need to annotate config files |
| TOML library | BurntSushi/toml | pelletier/go-toml | BurntSushi/toml has 10x more importers (36,500 vs ~3,500); simpler API for our use case; more recently updated (Dec 2025) |
| XDG paths | os.UserConfigDir (stdlib) | adrg/xdg | Stdlib covers our needs (config dir only); external lib only needed for data/cache dir separation |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Bubble Tea v2 (for now) | RC stage; Bubbles/Lip Gloss v2 are beta-only; risk of breaking changes before stable | Bubble Tea v1.3.10 + Lip Gloss v1.1.0 + Bubbles v0.21.1 |
| tview | Imperative paradigm fights the Elm Architecture; harder to test; less composable | Bubble Tea (functional, Elm-based) |
| go-holidays (Nager API) | Requires network connectivity for a local TUI app; adds latency and failure modes | rickar/cal (built-in definitions, offline) |
| SQLite | CGo dependency complicates cross-compilation; overkill for simple todo storage | JSON file via encoding/json |
| YAML for config | Implicit type coercion bugs (e.g., `no` becomes boolean); more complex than needed | TOML via BurntSushi/toml |
| cobra/viper | Over-engineered for a single-command TUI app; viper pulls in many transitive deps | Direct TOML config loading + stdlib flag |
| `charm.land/*` import paths | These are for v2 pre-release only; mixing charm.land and github.com paths causes module conflicts | `github.com/charmbracelet/*` (v1 paths) |

## Stack Patterns

**For calendar grid rendering:**
- Use Lip Gloss `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical` for the split-pane layout
- Use Lip Gloss styles with `Foreground(lipgloss.Color("..."))` for holiday highlighting (red)
- Build calendar grid as a custom Bubble Tea model, not from Bubbles (no calendar component exists)

**For the todo list pane:**
- Use Bubbles `list.Model` for the todo list with filtering and selection
- Implement `list.Item` interface on your todo type
- Use `list.NewDefaultDelegate()` for standard look-and-feel, customize for done/undone styling

**For user input (adding todos):**
- Option A: Embed a Bubbles `textinput.Model` inline in the app
- Option B: Use `huh` forms for a richer dialog (date picker, text input)
- Recommend Option A for simplicity; Option B if date selection UX matters

**For data persistence:**
- Load JSON on startup, write on every mutation (add, check, delete)
- Use `os.UserConfigDir()` + `/todo-calendar/` for file paths
- Create directory with `os.MkdirAll` on first run

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| bubbletea v1.3.10 | lipgloss v1.1.0 | Both v1; stable compatibility |
| bubbletea v1.3.10 | bubbles v0.21.1 | Bubbles v0.x is designed for bubbletea v1 |
| bubbletea v1.3.10 | huh v0.8.0 | Huh v0.x integrates with bubbletea v1 |
| rickar/cal v2.1.27 | Go 1.25.x | No TUI dependencies; pure date library |
| BurntSushi/toml v1.6.0 | Go 1.18+ | Minimal Go version requirement |
| Go 1.25.7 | All above | Current stable Go release |

**WARNING:** Do NOT mix v1 and v2 Charm packages. Bubble Tea v1 is incompatible with Lip Gloss v2 or Bubbles v2. Stick to all v1 or (when stable) all v2.

## Key Bubbles Components for This Project

| Component | Import | Use Case |
|-----------|--------|----------|
| `list` | `github.com/charmbracelet/bubbles/list` | Todo list with selection, filtering, scrolling |
| `key` | `github.com/charmbracelet/bubbles/key` | Keybinding definitions and management |
| `help` | `github.com/charmbracelet/bubbles/help` | Bottom help bar showing available keys |
| `textinput` | `github.com/charmbracelet/bubbles/textinput` | Inline text input for adding todos |
| `viewport` | `github.com/charmbracelet/bubbles/viewport` | Scrollable content (if calendar overflows) |

Components NOT needed:
- `spinner` -- no async loading
- `progress` -- no progress bars
- `table` -- calendar is custom grid, not tabular data
- `filepicker` -- no file operations
- `textarea` -- single-line input is sufficient for todo text

## rickar/cal Finnish Holidays Detail

The `fi` package in rickar/cal v2 defines 15 Finnish public holidays:

| Holiday | Finnish Name | Type |
|---------|-------------|------|
| New Year's Day | Uudenvuodenpaiva | Fixed (Jan 1) |
| Epiphany | Loppiainen | Fixed (Jan 6) |
| Good Friday | Pitkaperjantai | Easter-relative |
| Easter Sunday | Paasiaispäivä | Easter-relative |
| Easter Monday | Toinen paasiaispäivä | Easter-relative |
| Labour Day | Vappu | Fixed (May 1) |
| Ascension Day | Helatorstai | Easter-relative (39 days) |
| Pentecost | Helluntaipäivä | Easter-relative (49 days) |
| Midsummer's Eve | Juhannusaatto | Floating (Fri before Midsummer) |
| Midsummer's Day | Juhannuspäivä | Floating (first Sat from Jun 20) |
| All Saints' Day | Pyhäinpäivä | Floating (first Sat from Oct 31) |
| Independence Day | Itsenäisyyspäivä | Fixed (Dec 6) |
| Christmas Eve | Jouluaatto | Fixed (Dec 24) |
| Christmas Day | Joulupäivä | Fixed (Dec 25) |
| St. Stephen's Day | Tapaninpäivä | Fixed (Dec 26) |

Usage:
```go
import (
    "github.com/rickar/cal/v2"
    "github.com/rickar/cal/v2/fi"
)

c := cal.NewBusinessCalendar()
c.AddHoliday(fi.Holidays...)

// Check if a date is a holiday
_, isHoliday, _ := c.GetHoliday(someDate)
```

The library handles all complex date calculations (Easter, floating Saturdays) automatically. To support other countries, swap `fi` for any of the 50+ country packages.

## Sources

- [Bubble Tea GitHub](https://github.com/charmbracelet/bubbletea) -- v1.3.10 stable, v2.0.0-rc.2 pre-release (verified via pkg.go.dev)
- [Bubble Tea v2 Discussion #1374](https://github.com/charmbracelet/bubbletea/discussions/1374) -- v2 changes and migration guide
- [Bubble Tea v2 Milestone](https://github.com/charmbracelet/bubbletea/milestone/2) -- 77% complete, 6 open issues (verified 2026-02-05)
- [Bubbles GitHub](https://github.com/charmbracelet/bubbles) -- v0.21.1 (Feb 2026), component list verified
- [Lip Gloss GitHub](https://github.com/charmbracelet/lipgloss) -- v1.1.0 stable (Mar 2025)
- [Huh GitHub](https://github.com/charmbracelet/huh) -- v0.8.0 (Oct 2025)
- [rickar/cal pkg.go.dev](https://pkg.go.dev/github.com/rickar/cal/v2) -- v2.1.27 (Jan 2026), 50+ countries verified
- [rickar/cal fi holidays source](https://github.com/rickar/cal/blob/master/v2/fi/fi_holidays.go) -- 15 Finnish holidays verified
- [BurntSushi/toml pkg.go.dev](https://pkg.go.dev/github.com/BurntSushi/toml) -- v1.6.0 (Dec 2025), TOML v1.1.0
- [Go Release History](https://go.dev/doc/devel/release) -- Go 1.25.7 (Feb 2026) confirmed latest stable
- [os.UserConfigDir](https://github.com/golang/go/issues/29960) -- stdlib XDG support since Go 1.13

---
*Stack research for: Go TUI calendar + todo application*
*Researched: 2026-02-05*
