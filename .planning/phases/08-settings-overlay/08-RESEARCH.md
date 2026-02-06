# Phase 8: Settings Overlay - Research

**Researched:** 2026-02-06
**Domain:** Bubble Tea full-screen overlay, config write-back, live theme preview
**Confidence:** HIGH

## Summary

This phase adds an in-app settings overlay where users can configure theme, holiday country, and first day of week without editing config.toml by hand. The codebase already has all the building blocks: `config.Config` struct with TOML tags, `theme.ForName()` for theme resolution, `holidays.Registry` with `SupportedCountries()`, and the constructor DI pattern that passes theme through `app.New()` to children.

The core challenge is implementing a full-screen overlay within Bubble Tea's Elm Architecture. The standard pattern is simple: add a boolean `showSettings` flag on the root `Model`, and when true, `View()` renders the settings model instead of the normal two-pane layout. `Update()` routes all input to the settings model while the overlay is open. This is exactly the pattern used in Bubble Tea's own "composable-views" example. No third-party overlay library is needed -- the app already uses alt-screen mode, so the settings view simply replaces the main view content.

Live preview requires rebuilding all `Styles` structs when the user cycles through themes. The existing architecture supports this well: each component has a `NewStyles(theme.Theme)` constructor. A new method on each model (or a `SetTheme` approach) can replace the styles in-place. For cancel/revert, the app snapshots the original config before opening settings and restores it on Escape. For save, `BurntSushi/toml` v1.6.0 provides `toml.NewEncoder(w).Encode(cfg)` to write the Config struct back to disk.

**Primary recommendation:** Implement settings as a state-machine overlay on the root app model. Use a new `internal/settings/model.go` component with its own Update/View cycle. Route all keys to it when active. Support live preview by propagating theme changes via a custom `tea.Msg`. Save with atomic TOML write. Cancel by restoring a snapshot.

## Standard Stack

No new dependencies required. Everything needed is already in the project.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| bubbletea | v1.3.10 | TUI framework, Elm Architecture | Already in use; overlay is just conditional rendering in View() |
| lipgloss | v1.1.0 | Styled rendering for settings UI | Already in use; settings panel needs styled rows |
| BurntSushi/toml | v1.6.0 | Encode Config struct back to TOML file | Already in use for decode; `toml.NewEncoder(w).Encode(cfg)` handles write-back |
| bubbles/key | v0.21.1 | Keybinding handling | Already in use for all other key handling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| bubbles/help | v0.21.1 | Help bar for settings-specific keys | Already in use; settings overlay shows its own help keys |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom overlay rendering | rmhubbert/bubbletea-overlay | That library composites foreground on background (useful for partial overlays). Full-screen replacement is simpler and matches the "composable views" pattern -- no new dependency needed |
| charmbracelet/huh (form library) | Hand-built settings rows | huh is a full form framework with validation, groups, etc. Overkill for 3 settings fields that just cycle through predefined options. Hand-building keeps the codebase consistent with existing patterns |
| toml.Marshal() | toml.NewEncoder().Encode() | Marshal returns []byte which then needs file I/O separately. NewEncoder writing to a buffer is equivalent; either works. Use the atomic-write-to-temp-file pattern already established in store.Save() |

## Architecture Patterns

### Recommended Project Structure
```
internal/
  settings/
    model.go       # Settings overlay Model with Update/View
    keys.go        # KeyMap for settings navigation
    styles.go      # Styles struct built from theme (consistent with other packages)
  config/
    config.go      # Add Save() method for writing TOML back to disk
  app/
    model.go       # Add showSettings flag, settings Model field, routing logic
    keys.go        # Add Settings key binding ("s")
```

### Pattern 1: Overlay as Conditional View Routing

**What:** The root app model holds a `showSettings bool` and a `settings.Model`. When `showSettings` is true, both `Update()` and `View()` delegate entirely to the settings model. When false, normal two-pane rendering occurs.

**When to use:** For full-screen overlays that completely replace the main view.

**Example:**
```go
// app/model.go

type Model struct {
    // ... existing fields ...
    showSettings bool
    settings     settings.Model
    cfg          config.Config  // current config, needed for save/revert
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Settings overlay intercepts ALL input when active
    if m.showSettings {
        return m.updateSettings(msg)
    }
    // ... existing update logic ...
}

func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }
    if m.showSettings {
        return m.settings.View()
    }
    // ... existing two-pane rendering ...
}
```

### Pattern 2: Live Preview via Theme Propagation

**What:** When the user cycles the theme option in settings, a custom message (`ThemeChangedMsg`) is sent. The root model catches this, rebuilds all styles for itself and its children.

**When to use:** When a setting change must visually affect the entire app immediately.

**Example:**
```go
// settings/model.go
type ThemeChangedMsg struct {
    Theme theme.Theme
}

// In settings Update, when user cycles theme:
return m, func() tea.Msg {
    return ThemeChangedMsg{Theme: theme.ForName(m.themeName)}
}

// app/model.go -- in updateSettings:
case settings.ThemeChangedMsg:
    m.applyTheme(msg.Theme)
    // Continue routing to settings model too
```

**Key implementation detail:** Each component needs a method to replace its styles at runtime:
```go
// calendar/model.go
func (m *Model) SetTheme(t theme.Theme) {
    m.styles = NewStyles(t)
}
```

### Pattern 3: Snapshot and Restore for Cancel

**What:** Before opening settings, snapshot the current config. If the user cancels (Escape), restore the snapshot and rebuild all styles from the original theme.

**When to use:** When live-preview changes must be revertible.

**Example:**
```go
// When opening settings:
m.savedConfig = m.cfg  // snapshot
m.showSettings = true

// On cancel (Escape from settings):
m.cfg = m.savedConfig
m.applyTheme(theme.ForName(m.savedConfig.Theme))
m.showSettings = false

// On save (Enter from settings):
m.cfg = m.settings.Config()  // get modified config
m.cfg.Save()                 // write to disk
m.showSettings = false
// Theme already applied via live preview -- no additional rebuild needed
```

### Pattern 4: Settings as Cycling Options (Not Free-Text)

**What:** Each setting is a list of predefined values. The user cycles through them with left/right arrow keys (or h/l). No free-text input needed.

**When to use:** When all valid values are known ahead of time.

**Example:**
```go
type option struct {
    label   string     // "Theme", "Country", "First Day"
    values  []string   // ["dark", "light", "nord", "solarized"]
    display []string   // ["Dark", "Light", "Nord", "Solarized"]
    index   int        // currently selected index
}

// Cycle with left/right:
case key.Matches(msg, m.keys.Left):
    m.options[m.cursor].index--
    if m.options[m.cursor].index < 0 {
        m.options[m.cursor].index = len(m.options[m.cursor].values) - 1
    }
```

### Anti-Patterns to Avoid

- **Rebuilding child models from scratch on theme change:** Do NOT re-create `calendar.Model` and `todolist.Model` when the theme changes. This loses state (current month, cursor position, input mode). Instead, add `SetTheme()` methods that only replace the `styles` field.
- **Using text input for settings values:** All three settings (theme, country, first-day) have a fixed set of valid values. Cycling through a list is faster and prevents invalid input.
- **Coupling settings model to app internals:** The settings model should work with a plain config struct and emit messages. It should not directly mutate the calendar or todolist models.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| TOML file writing | Custom string formatting/templating | `toml.NewEncoder(w).Encode(cfg)` | Handles escaping, formatting, struct tag mapping automatically |
| Config directory creation | Manual mkdir | `os.MkdirAll(dir, 0755)` before write | Already used in store.Save() pattern |
| Country name display | External ISO 3166 library | Hardcoded map of 11 entries | Only 11 countries in Registry; adding a dependency for 11 strings is wasteful |
| Overlay compositing | bubbletea-overlay library | Conditional rendering in View() | Full-screen overlay is simpler than partial overlay compositing |

**Key insight:** The settings overlay is architecturally simple because it is full-screen. A partial overlay (floating dialog on top of the app) would require string-level compositing. A full-screen replacement is just an if/else in View().

## Common Pitfalls

### Pitfall 1: Holiday Provider Not Rebuilt on Country Change
**What goes wrong:** User changes country in settings, saves, but holidays still show the old country.
**Why it happens:** The `holidays.Provider` is created once in `main.go` and stored on the calendar model. Changing `cfg.Country` in config doesn't rebuild the provider.
**How to avoid:** After saving settings with a changed country, create a new `holidays.Provider` and pass it to the calendar model via a new `SetProvider()` method (or rebuild it in-place). Similarly, rebuild if `first_day_of_week` changes (the calendar model stores `mondayStart bool`).
**Warning signs:** Country dropdown changes in settings but calendar still shows old country's holidays.

### Pitfall 2: Losing Calendar/Todo State on Theme Change
**What goes wrong:** Live theme preview causes the calendar to jump back to current month, or todo cursor resets to 0.
**Why it happens:** If theme change triggers re-creation of child models (via `calendar.New()` etc.), all runtime state is lost.
**How to avoid:** Add `SetTheme(t theme.Theme)` methods to calendar.Model and todolist.Model that ONLY update the `styles` field. Never re-create the models.
**Warning signs:** Cycling themes in settings causes the calendar to snap back to the current month.

### Pitfall 3: Settings Keybinding Conflicts During Input Mode
**What goes wrong:** Pressing "s" while typing a todo name opens settings instead of inserting the letter "s".
**Why it happens:** The settings keybinding is checked before routing to the todolist's input mode.
**How to avoid:** Check `isInputting` before handling the settings key, exactly as the existing Quit keybinding does: `key.Matches(msg, m.keys.Settings) && !isInputting`.
**Warning signs:** Cannot type the letter "s" in todo text.

### Pitfall 4: Config File Not Created on First Save
**What goes wrong:** User has never manually created config.toml. Settings save fails because the directory does not exist.
**Why it happens:** `config.Load()` gracefully returns defaults when the file is missing, but `Save()` needs the directory to exist.
**How to avoid:** In `config.Save()`, call `os.MkdirAll(filepath.Dir(path), 0755)` before writing, following the same pattern as `store.Save()`.
**Warning signs:** "Config error" or silent failure when saving settings for the first time.

### Pitfall 5: TOML Encoder Changes Field Order
**What goes wrong:** User had a hand-edited config.toml with comments. After saving from settings, comments are gone and field order changes.
**Why it happens:** `toml.Encode()` serializes the struct fields in order, ignoring any pre-existing file content or comments.
**How to avoid:** Accept this as a known limitation. The Config struct has only 3 fields with `toml` tags, so the output is clean and predictable. Document that comments in config.toml may be lost when saving from the settings overlay.
**Warning signs:** User complaints about lost comments in config.toml.

### Pitfall 6: Help Bar Not Updated for Settings Context
**What goes wrong:** Settings overlay is open but help bar still shows calendar/todo keybindings.
**Why it happens:** `currentHelpKeys()` in app/model.go only considers calendarPane and todoPane.
**How to avoid:** When `showSettings` is true, return settings-specific help keys (j/k navigate, left/right change value, enter save, esc cancel).
**Warning signs:** Help bar shows irrelevant keys when in settings.

## Code Examples

### Config Save Method (Atomic Write)
```go
// internal/config/config.go
// Source: Follows same atomic-write pattern as store.Save()

func Save(cfg Config) error {
    path, err := Path()
    if err != nil {
        return err
    }

    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    var buf bytes.Buffer
    if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
        return err
    }

    // Atomic write: temp file -> sync -> rename
    tmp, err := os.CreateTemp(dir, ".config-*.tmp")
    if err != nil {
        return err
    }
    tmpName := tmp.Name()

    if _, err := tmp.Write(buf.Bytes()); err != nil {
        tmp.Close()
        os.Remove(tmpName)
        return err
    }
    if err := tmp.Sync(); err != nil {
        tmp.Close()
        os.Remove(tmpName)
        return err
    }
    if err := tmp.Close(); err != nil {
        os.Remove(tmpName)
        return err
    }
    return os.Rename(tmpName, path)
}
```

### Settings Model Core Structure
```go
// internal/settings/model.go

type option struct {
    label   string   // displayed label: "Theme", "Country", "First Day of Week"
    values  []string // config values: ["dark", "light", "nord", "solarized"]
    display []string // display values: ["Dark", "Light", "Nord", "Solarized"]
    index   int      // currently selected
}

type Model struct {
    options []option
    cursor  int // which option row is selected (0, 1, or 2)
    width   int
    height  int
    keys    KeyMap
    styles  Styles
}

func New(cfg config.Config, t theme.Theme) Model {
    themeNames := []string{"dark", "light", "nord", "solarized"}
    themeDisplay := []string{"Dark", "Light", "Nord", "Solarized"}

    countries := holidays.SupportedCountries() // returns sorted []string
    countryDisplay := countryLabels(countries)  // "us" -> "US - United States"

    dayValues := []string{"sunday", "monday"}
    dayDisplay := []string{"Sunday", "Monday"}

    return Model{
        options: []option{
            {label: "Theme", values: themeNames, display: themeDisplay, index: indexOf(themeNames, cfg.Theme)},
            {label: "Country", values: countries, display: countryDisplay, index: indexOf(countries, cfg.Country)},
            {label: "First Day of Week", values: dayValues, display: dayDisplay, index: indexOf(dayValues, cfg.FirstDayOfWeek)},
        },
        keys:   DefaultKeyMap(),
        styles: NewStyles(t),
    }
}

func (m Model) Config() config.Config {
    return config.Config{
        Theme:          m.options[0].values[m.options[0].index],
        Country:        m.options[1].values[m.options[1].index],
        FirstDayOfWeek: m.options[2].values[m.options[2].index],
    }
}
```

### Settings View Rendering
```go
// internal/settings/model.go

func (m Model) View() string {
    var b strings.Builder

    title := "Settings"
    b.WriteString(m.styles.Title.Render(title))
    b.WriteString("\n\n")

    for i, opt := range m.options {
        isSelected := i == m.cursor

        label := fmt.Sprintf("  %-20s", opt.label)
        value := fmt.Sprintf("<  %s  >", opt.display[opt.index])

        if isSelected {
            label = m.styles.SelectedLabel.Render(fmt.Sprintf("> %-20s", opt.label))
            value = m.styles.SelectedValue.Render(value)
        } else {
            label = m.styles.Label.Render(label)
            value = m.styles.Value.Render(value)
        }

        b.WriteString(label + value + "\n")
    }

    b.WriteString("\n")
    b.WriteString(m.styles.Hint.Render("  enter save  |  esc cancel  |  <-/-> change value"))

    return b.String()
}
```

### Theme Propagation - SetTheme Methods
```go
// calendar/model.go
func (m *Model) SetTheme(t theme.Theme) {
    m.styles = NewStyles(t)
}

// calendar/model.go - also need for country/first-day changes
func (m *Model) SetProvider(p *holidays.Provider) {
    m.provider = p
    m.holidays = p.HolidaysInMonth(m.year, m.month)
}

func (m *Model) SetMondayStart(v bool) {
    m.mondayStart = v
}

// todolist/model.go
func (m *Model) SetTheme(t theme.Theme) {
    m.styles = NewStyles(t)
}

// app/model.go
func (m *Model) applyTheme(t theme.Theme) {
    m.styles = NewStyles(t)
    m.calendar.SetTheme(t)
    m.todoList.SetTheme(t)
    m.settings.SetTheme(t)  // settings itself is also themed
    // Re-theme help bar
    m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(t.AccentFg)
    m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(t.MutedFg)
    m.help.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(t.MutedFg)
}
```

### Country Display Labels
```go
// internal/settings/model.go

var countryNames = map[string]string{
    "de": "Germany",
    "dk": "Denmark",
    "ee": "Estonia",
    "es": "Spain",
    "fi": "Finland",
    "fr": "France",
    "gb": "United Kingdom",
    "it": "Italy",
    "no": "Norway",
    "se": "Sweden",
    "us": "United States",
}

func countryLabels(codes []string) []string {
    labels := make([]string, len(codes))
    for i, code := range codes {
        name := countryNames[code]
        if name == "" {
            name = strings.ToUpper(code)
        }
        labels[i] = fmt.Sprintf("%s - %s", strings.ToUpper(code), name)
    }
    return labels
}
```

### Keybinding for Opening Settings
```go
// app/keys.go
type KeyMap struct {
    Quit     key.Binding
    Tab      key.Binding
    Settings key.Binding  // NEW
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        // ... existing ...
        Settings: key.NewBinding(
            key.WithKeys("s"),
            key.WithHelp("s", "settings"),
        ),
    }
}
```

**Note:** "s" is unused by all existing keybindings (verified: calendar uses h/l/left/right, todolist uses j/k/J/K/a/A/x/d/e/E/enter/esc, app uses q/ctrl+c/tab). The "s" key is intuitive for "settings" and works from both panes.

### Custom Messages for Settings Communication
```go
// settings/model.go

// ThemeChangedMsg is emitted when the user cycles the theme option.
// The parent model uses this to trigger live preview.
type ThemeChangedMsg struct {
    Theme theme.Theme
}

// SaveMsg is emitted when the user presses Enter to save.
type SaveMsg struct {
    Cfg config.Config
}

// CancelMsg is emitted when the user presses Escape to cancel.
type CancelMsg struct{}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Config file only (edit config.toml by hand) | In-app settings with live preview | This phase | Users don't need to know file location or TOML syntax |
| One-time theme load at startup | Runtime theme switching via SetTheme() | This phase | Enables live preview; all components must support style replacement |
| Read-only config.Load() | Bidirectional config.Load() + config.Save() | This phase | App can now persist user preferences from UI |

**No deprecated APIs involved.** All patterns use current Bubble Tea v1.3.x and Lipgloss v1.1.x APIs.

## Open Questions

1. **Help bar in settings overlay**
   - What we know: The app uses `help.Model` at the bottom. Settings needs its own help text (enter/esc/arrows).
   - What's unclear: Whether to use the same `help.Model` with different bindings, or render help text directly in the settings View().
   - Recommendation: Render settings help as a styled string at the bottom of the settings view. Simpler than swapping help.Model bindings. The help bar is part of the app View, which is skipped when settings is shown. So settings renders its own footer.

2. **Window size propagation to settings model**
   - What we know: `tea.WindowSizeMsg` is broadcast to children in the root Update.
   - What's unclear: Whether settings model needs width/height for centered rendering.
   - Recommendation: Pass WindowSizeMsg to settings model. Settings should use width/height to center its content vertically and horizontally for a clean full-screen look.

3. **Theme name list maintenance**
   - What we know: Theme presets are defined as functions in theme.go (Dark, Light, Nord, Solarized). There is no `ThemeNames()` function.
   - What's unclear: Whether to add a ThemeNames() function to theme package or hardcode the list in settings.
   - Recommendation: Add a `Names() []string` function to theme package that returns `["dark", "light", "nord", "solarized"]`. This keeps the theme package as the single source of truth and avoids settings getting out of sync if themes are added later.

## Sources

### Primary (HIGH confidence)
- Codebase audit: All 19 Go source files read directly (config.go, paths.go, theme.go, app/model.go, app/keys.go, app/styles.go, calendar/model.go, calendar/styles.go, calendar/keys.go, calendar/grid.go, todolist/model.go, todolist/styles.go, todolist/keys.go, holidays/registry.go, holidays/provider.go, store/store.go, store/todo.go, main.go, go.mod)
- BurntSushi/toml v1.6.0 Encoder API: [pkg.go.dev/github.com/BurntSushi/toml](https://pkg.go.dev/github.com/BurntSushi/toml) -- NewEncoder, Encode, struct tag support verified
- Bubble Tea composable views example: [github.com/charmbracelet/bubbletea/blob/main/examples/views/main.go](https://github.com/charmbracelet/bubbletea/blob/main/examples/views/main.go) -- conditional view switching pattern verified

### Secondary (MEDIUM confidence)
- Bubble Tea overlay patterns: [pkg.go.dev/github.com/rmhubbert/bubbletea-overlay](https://pkg.go.dev/github.com/quickphosphat/bubbletea-overlay) -- confirmed full-screen replacement is simpler than library-based compositing for this use case

### Tertiary (LOW confidence)
- None -- all findings verified with primary or secondary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies; all APIs verified against existing codebase and official docs
- Architecture: HIGH -- overlay pattern directly from Bubble Tea examples; config save verified against BurntSushi/toml API; all component signatures verified from codebase
- Pitfalls: HIGH -- derived from direct code audit (provider lifecycle, style rebuild, input mode conflict, directory creation)
- Code examples: HIGH -- based on actual codebase patterns (constructor DI, Styles struct, KeyMap, atomic write)

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (stable domain, all libraries at current versions)
