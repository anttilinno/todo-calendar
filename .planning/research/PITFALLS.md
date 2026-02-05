# Pitfalls Research

**Domain:** Go TUI Calendar + Todo App (Bubble Tea)
**Researched:** 2026-02-05
**Confidence:** HIGH (verified via official docs, GitHub discussions, and multiple community sources)

## Critical Pitfalls

### Pitfall 1: View() Called Before Terminal Dimensions Are Known

**What goes wrong:**
Bubble Tea calls `View()` before `tea.WindowSizeMsg` arrives. If your View function depends on `width` and `height` to calculate the split-pane layout, the first render produces garbage output -- misaligned columns, overflowing lines, or panics from zero-division when computing pane widths.

**Why it happens:**
The window size query runs asynchronously so it does not block program startup. `View()` is called immediately after `Init()`, but `WindowSizeMsg` has not yet arrived. This is documented behavior, not a bug (see [bubbletea#282](https://github.com/charmbracelet/bubbletea/issues/282)).

**How to avoid:**
Track initialization state on your model. Do not attempt layout calculations until dimensions arrive:
```go
type model struct {
    state  int // 0 = initializing, 1 = ready
    width  int
    height int
}

func (m model) View() string {
    if m.width == 0 || m.height == 0 {
        return "Loading..."
    }
    // Safe to compute split-pane layout now
}
```
Set `width`/`height` in `Update` when you receive `tea.WindowSizeMsg`, and propagate the dimensions to all child components at the same time.

**Warning signs:**
- Momentary layout flash on startup
- Panic stack trace referencing View with zero dimensions
- Layout arithmetic using hardcoded terminal sizes

**Phase to address:**
Phase 1 (scaffold) -- this must be in the initial application skeleton before any layout code is written.

---

### Pitfall 2: Frame Size Accounting Errors in Split-Pane Layout

**What goes wrong:**
When calculating pane widths for the calendar (left) and todo list (right), developers subtract the raw pane content width from terminal width but forget to account for borders, padding, and margins added by Lipgloss styles. The result: content overflows the terminal width, causing line wrapping that destroys the layout. This is the single most common layout bug in Bubble Tea applications with side-by-side panes.

**Why it happens:**
Lipgloss styles (borders, padding, margins) consume characters that are invisible in the mental model of "my pane is 40 columns wide." A 1-character border on each side of both panes eats 4 columns. A vertical separator eats 1 more. Developers hardcode `leftWidth = termWidth / 2` without subtracting frame overhead.

**How to avoid:**
Always use `style.GetFrameSize()` to measure the overhead of each style, and subtract it from available space before splitting:
```go
// Measure the overhead from borders, padding, margins
hFrame, vFrame := calendarStyle.GetFrameSize()
availableWidth := m.width - hFrame
// Then split availableWidth between panes
```
Never hardcode dimension arithmetic. Use `lipgloss.Width()` and `lipgloss.Height()` to measure rendered strings rather than guessing their size. If the calendar pane is a fixed width (e.g., always 22 columns for a `cal`-like grid), compute the todo pane as `remainingWidth = totalWidth - calendarRenderedWidth - separatorWidth - frameOverhead`.

**Warning signs:**
- Content wrapping to the next line in the terminal
- Layout that "works on my terminal" but breaks at different sizes
- Repeated manual tweaking of `+1` / `-1` constants

**Phase to address:**
Phase 1 (scaffold) -- establish the split-pane layout with correct frame accounting from the start. Regression-test at multiple terminal widths (80, 120, 200 columns).

---

### Pitfall 3: Choosing Bubble Tea v1 vs v2 at the Wrong Time

**What goes wrong:**
Starting a new project on v1 means eventually migrating to v2, which has significant breaking changes (new import path `charm.land/bubbletea/v2`, `View()` returns `tea.View` instead of `string`, `tea.KeyMsg` split into `KeyPressMsg`/`KeyReleaseMsg`, mouse API restructured). Starting on v2 RC means building on pre-release software where APIs may still shift and community examples are sparse.

**Why it happens:**
As of February 2026, Bubble Tea v2 is at release candidate stage (v2.0.0-rc.2, released November 2025). The latest stable is v1.3.10 (September 2025). v2 has been in alpha/beta since March 2025. The stable release appears imminent but has not landed.

**How to avoid:**
**Recommendation: Start with v1 (v1.3.10).** Rationale:
- v1 has stable APIs, comprehensive examples, and proven community patterns.
- The `bubbles` component library (list, textinput, viewport) is battle-tested on v1.
- For a personal todo app, the v2 improvements (synchronized output Mode 2026, better key handling) are nice-to-have, not critical.
- Migration from v1 to v2 is mechanical and well-documented (see [discussion#1374](https://github.com/charmbracelet/bubbletea/discussions/1374)).
- If v2 goes stable during development, migration is a bounded task -- not an architectural rewrite.

Keep architecture clean (small models, message-driven updates) so that migration is a find-and-replace exercise, not a redesign.

**Warning signs:**
- Go module path pointing at pre-release version with `@v2.0.0-rc.X`
- Struggling to find working examples for v2 APIs
- `bubbles` components not yet updated for v2

**Phase to address:**
Phase 1 (scaffold) -- version choice is locked in with `go mod init`. Document the decision and the migration plan.

---

### Pitfall 4: Mutating Model State Outside of Update()

**What goes wrong:**
Directly modifying the model from within a `tea.Cmd` goroutine or from a helper method called outside `Update()` causes race conditions. The Bubble Tea runtime replaces the entire model after each `Update` call. Mutations made concurrently are silently lost or, worse, cause data races that corrupt state.

**Why it happens:**
Go developers are accustomed to pointer-receiver methods that mutate state in place. Bubble Tea's Elm Architecture requires all state changes to flow through `Update()` via messages. Commands (`tea.Cmd`) run in separate goroutines and must return `tea.Msg` values, never touch the model directly. This is the most common architectural mistake in Bubble Tea applications.

**How to avoid:**
- All I/O and async operations go in `tea.Cmd` functions. They return `tea.Msg`, not modify the model.
- All model mutations happen in `Update()` in response to messages.
- Run `go vet -race` during development to catch data races early.
- Never capture `*model` in a `tea.Cmd` closure. Capture only the data the command needs (e.g., a filename, a todo ID), not the model itself.

```go
// WRONG: captures model pointer in Cmd closure
func (m *model) saveTodos() tea.Cmd {
    return func() tea.Msg {
        m.saving = true  // RACE CONDITION
        err := save(m.todos)
        return savedMsg{err: err}
    }
}

// CORRECT: captures only data, returns a message
func saveTodosCmd(todos []Todo, path string) tea.Cmd {
    return func() tea.Msg {
        err := saveTodosToFile(todos, path)
        return savedMsg{err: err}
    }
}
```

**Warning signs:**
- `-race` flag detects data races
- Intermittent state corruption (todo appears then vanishes)
- Model fields changing without a corresponding message in Update

**Phase to address:**
Phase 1 (scaffold) -- establish the Cmd/Msg pattern in the initial skeleton. Every subsequent phase inherits this discipline.

---

### Pitfall 5: Non-Atomic File Writes Causing Data Loss

**What goes wrong:**
Using `os.WriteFile()` to persist todos means a crash or power loss mid-write produces a truncated or empty file. The user loses all their todo data. `os.WriteFile` is explicitly not atomic -- the Go standard library documents this (see [golang/go#56173](https://github.com/golang/go/issues/56173)).

**Why it happens:**
`os.WriteFile` is the obvious stdlib choice and works fine in development. The failure mode only manifests under real-world conditions: process killed during write, filesystem full, power loss. By the time data is lost, there is no recovery.

**How to avoid:**
Use the write-to-temp-then-rename pattern:
1. Write to a temporary file in the same directory.
2. `fsync` the temporary file.
3. `os.Rename` the temp file over the target (atomic on Linux/macOS).

Use an existing library like `github.com/google/renameio` or `github.com/natefinch/atomic` rather than hand-rolling this. Example:
```go
import "github.com/google/renameio"

func saveTodos(path string, data []byte) error {
    return renameio.WriteFile(path, data, 0644)
}
```

Note: On Windows, `os.Rename` is not atomic. If Windows support matters, use `natefinch/atomic` which handles this. For a personal Linux TUI tool, `renameio` is sufficient.

**Warning signs:**
- Direct `os.WriteFile` or `ioutil.WriteFile` calls in the save path
- No temporary file in the save logic
- Missing `fsync` before rename
- No backup/recovery mechanism

**Phase to address:**
Phase 2 (data persistence) -- this must be correct from the first time data is saved to disk. Do not ship a "write directly" approach and plan to fix it later.

---

### Pitfall 6: Calendar Grid Alignment Broken by ANSI Color Codes

**What goes wrong:**
When highlighting holidays in red using ANSI escape codes (via Lipgloss), the escape characters are invisible but consume bytes. If you compute string widths using `len()` instead of a display-width-aware function, column alignment in the calendar grid breaks. Days shift right, the grid becomes ragged, and the calendar is unreadable.

**Why it happens:**
A `cal`-like calendar grid depends on exact column alignment. Each day number must be exactly 2-3 characters wide, with precise spacing. ANSI escape codes for red text add ~11 bytes per colored cell (`\033[31m12\033[0m` vs `12`). `len()` counts bytes, not visible characters. `utf8.RuneCountInString()` counts runes but still includes escape sequence runes.

**How to avoid:**
Always use `lipgloss.Width()` to measure the visible width of styled strings. Build the calendar grid by constructing each cell as a styled string of a fixed visible width, then join cells with `lipgloss.JoinHorizontal`. Do NOT build the grid as a raw string and apply styles after the fact -- style each cell individually, then compose.

```go
// Each cell is exactly 3 visible characters wide
cell := lipgloss.NewStyle().
    Foreground(lipgloss.Color("9")). // red for holidays
    Width(3).
    Render(fmt.Sprintf("%2d", day))
```

**Warning signs:**
- Calendar columns misalign when holidays are present but align when they are absent
- Visual width of rows differs between holiday and non-holiday rows
- Manual space-padding that works for plain text but breaks for styled text

**Phase to address:**
Phase 2 (calendar rendering) -- this is the core of calendar display. Get the cell-based rendering pattern right from the first calendar implementation.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Hardcoding terminal width (e.g., 80 columns) | Fast initial layout | Breaks on any other terminal size; must rewrite layout | Never -- `WindowSizeMsg` is trivial to use |
| Single monolithic model struct | Everything in one place, no message routing | Unmaintainable after ~300 lines; all concerns tangled | MVP only if refactored before adding features |
| Storing todos as plain text instead of structured JSON | Simpler initial implementation | Cannot add fields (dates, completion status) without parsing migration | Never -- JSON is equally simple to implement from day one |
| Saving on every keystroke | Data is always persisted | Excessive disk I/O; perceptible lag on slow storage; file contention | Never -- save on meaningful actions (add, complete, delete) |
| Skipping error handling in Cmds | Less code | Silent data loss; user never knows save failed | Never |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| `rickar/cal` holiday library | Assuming all countries are supported; importing the entire library | Check the [v2 subdirectories](https://github.com/rickar/cal/tree/master/v2) for your country's ISO code. ~46 countries supported. Finland (FI), US, GB, DE, etc. are present. Import only your country's subpackage. |
| `rickar/cal` holiday dates | Assuming holiday dates are always the calendar date, ignoring observed dates | Holidays have both `ActualDate` and `ObservedDate`. A holiday on Saturday may be observed on Friday. Decide which to display and be consistent. |
| `bubbles/list` component | Using the list component for the todo panel and fighting its built-in filtering/search behavior | The `list` bubble includes filtering by default. For a simple todo list, either disable filtering or build a custom list from scratch with a viewport. Evaluate whether the built-in list actually matches your needs before committing to it. |
| `bubbles/textinput` for adding todos | Forgetting to call `Focus()` on the textinput when entering add mode | The textinput is blurred by default. Without calling `Focus()`, it silently swallows all keyboard input. Toggle `Focus()`/`Blur()` explicitly when switching between navigation and input modes. |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Rebuilding the entire calendar grid on every `View()` call | Slight lag on low-powered terminals or SSH sessions | Cache the rendered calendar string; only regenerate when the viewed month changes or terminal resizes | Noticeable on Raspberry Pi or high-latency SSH connections |
| Rendering all todos in the View even when list exceeds viewport | Long render times; invisible content below the fold still computed | Use a viewport/scrollable container; only render visible items | Unlikely for personal use (few hundred todos), but good practice |
| Re-reading the todo file from disk on every Update cycle | Disk I/O on every keypress | Load once into memory on startup; write-back on mutations only | Immediate lag if file is on network storage or slow disk |
| Calling `lipgloss.Render` with complex styles in tight loops | Measurable rendering overhead from style computation | Pre-compute styles at initialization; reuse style objects rather than creating new ones per render | Dozens of styled cells per frame (calendar grid with 42 cells is fine; hundreds would show lag) |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Storing todo file with world-readable permissions (0644) | Other users on shared system can read personal todos | Use `0600` permissions. The `renameio` or `atomic` libraries accept permissions as a parameter. |
| Not sanitizing todo text before rendering | Control characters or ANSI escape sequences in todo text could corrupt terminal display | Strip or escape control characters when accepting input; Lipgloss rendering handles most cases, but raw string injection in View could break layout |
| Hardcoding file path (e.g., `~/.todos.json`) instead of using XDG | File stored in unexpected location; conflicts with other tools; no standard cleanup path | Use `os.UserConfigDir()` for config, `os.UserHomeDir()` + `.local/share/` for data, or the `adrg/xdg` library for full XDG compliance |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No visual feedback when saving | User unsure if todo was actually saved | Brief status message in footer: "Saved" that fades after 2 seconds, or a subtle indicator |
| No confirmation before delete | Accidental deletion with no undo | Either add a confirmation prompt, or implement soft-delete (mark deleted, purge on exit), or show a brief "Undo? Press z" message |
| Calendar navigation overwrites todo panel focus | User typing a todo accidentally navigates months | Maintain explicit focus state: calendar pane vs todo pane. Only the focused pane responds to keys. Tab or a dedicated key switches focus. |
| Holiday names not visible anywhere | User sees red dates but does not know which holiday it is | Show holiday name in a footer bar or tooltip area when the month contains holidays. Even a simple list below the calendar: "5 -- Independence Day" |
| Adding a todo requires too many keystrokes | Friction discourages use | Single keypress (e.g., `a`) enters add mode. Type text, optional date prefix, Enter to save. Minimal ceremony. |

## "Looks Done But Isn't" Checklist

- [ ] **Split-pane layout:** Often missing resize handling -- verify layout recalculates on `WindowSizeMsg` and looks correct at 80, 120, and 200+ column widths
- [ ] **Calendar rendering:** Often missing the 6th week row -- months can span 6 weeks (e.g., a month starting on Saturday). Verify February and months starting late in the week render correctly with 4, 5, and 6 week rows
- [ ] **Holiday highlighting:** Often missing edge cases -- verify holidays that fall on weekends, observed dates vs actual dates, and that the correct year's holidays are calculated (not just the current year, but also for navigated months in other years)
- [ ] **Todo persistence:** Often missing error handling -- verify what happens when the file is missing on first launch (should create it), when the file is corrupted (should not crash), and when disk is full (should show error, not silently fail)
- [ ] **Keyboard handling:** Often missing edge cases -- verify that pressing keys rapidly does not drop inputs, that Ctrl+C always exits cleanly, and that the escape key works consistently to cancel operations
- [ ] **Month navigation:** Often missing year boundary -- verify navigating backward from January goes to December of the previous year, and forward from December goes to January of the next year
- [ ] **Empty states:** Often missing entirely -- verify what the app looks like with zero todos for a month (should show a helpful message, not a blank pane) and what the todo pane shows for months with only floating items

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Data loss from non-atomic write | HIGH (data gone) | Implement atomic writes. If data already lost, no recovery possible. Consider adding periodic backup (copy to `.bak` before overwrite) as defense-in-depth. |
| Layout miscalculated due to frame size errors | LOW | Fix the arithmetic using `GetFrameSize()`. No data impact, purely visual. |
| Race condition from model mutation in Cmd | MEDIUM | Refactor Cmd to return messages instead of mutating model. May require restructuring several handlers. Run with `-race` to verify fix. |
| Calendar grid misalignment from ANSI width | LOW | Switch to cell-based rendering with `lipgloss.Width()`. Isolated to calendar component. |
| Wrong Bubble Tea version choice | MEDIUM | Migration from v1 to v2 is documented and mechanical, but touches every file with Update/View/KeyMsg. Budget a focused sprint for it. |
| Holidays missing for user's country | LOW | `rickar/cal` supports ~46 countries. If missing, define custom holidays using the library's `Holiday` struct. The API supports exact dates, floating rules, and custom functions. |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| View() before WindowSizeMsg | Phase 1: Scaffold | App shows "Loading..." briefly then renders correctly; no panics on startup |
| Frame size accounting errors | Phase 1: Scaffold | Layout tested at 80, 120, 200 columns with no line wrapping |
| v1 vs v2 version choice | Phase 1: Scaffold | `go.mod` pins `bubbletea v1.3.10`; migration plan documented |
| Model mutation outside Update | Phase 1: Scaffold | `go vet -race` passes; all Cmds return Msg, never reference model |
| Non-atomic file writes | Phase 2: Data persistence | Save uses write-temp-rename pattern; verify by killing process during save |
| Calendar ANSI alignment | Phase 2: Calendar rendering | Calendar grid aligned with and without holidays; tested with different terminal themes |
| Focus management between panes | Phase 3: Todo interaction | Tab switches focus; typing in todo mode does not navigate calendar; calendar keys do not corrupt todo input |
| Holiday edge cases | Phase 2: Calendar rendering | Holidays render correctly for navigated months across year boundaries; observed dates handled |
| Empty states | Phase 3: Todo interaction | Every state has a visual representation: no todos, no dated todos, no floating todos |
| Delete confirmation / undo | Phase 3: Todo interaction | Accidental delete is recoverable; user tested the delete flow |

## Sources

- [Bubble Tea GitHub repository](https://github.com/charmbracelet/bubbletea) -- official docs, issues, discussions (HIGH confidence)
- [View() called before WindowSizeMsg - Issue #282](https://github.com/charmbracelet/bubbletea/issues/282) (HIGH confidence)
- [Pointer receivers discussion #434](https://github.com/charmbracelet/bubbletea/discussions/434) (HIGH confidence)
- [Tips for building Bubble Tea programs - leg100](https://leg100.github.io/en/posts/building-bubbletea-programs/) (MEDIUM confidence -- community, but well-sourced)
- [Bubble Tea v2 migration guide - Discussion #1374](https://github.com/charmbracelet/bubbletea/discussions/1374) (HIGH confidence)
- [Layout handling discussion #307](https://github.com/charmbracelet/bubbletea/discussions/307) (HIGH confidence)
- [View Layout Issues discussion #943](https://github.com/charmbracelet/bubbletea/discussions/943) (HIGH confidence)
- [Commands in Bubble Tea - charm.land blog](https://charm.land/blog/commands-in-bubbletea/) (HIGH confidence -- official)
- [Loss of input in Bubble Tea - dr-knz.net](https://dr-knz.net/bubbletea-control-inversion.html) (MEDIUM confidence)
- [rickar/cal v2 -- Go holiday library](https://github.com/rickar/cal) (HIGH confidence)
- [google/renameio -- atomic file writes](https://pkg.go.dev/github.com/google/renameio) (HIGH confidence)
- [natefinch/atomic](https://github.com/natefinch/atomic) (HIGH confidence)
- [golang/go#56173 -- os.WriteFile not atomic](https://github.com/golang/go/issues/56173) (HIGH confidence)
- [Atomically writing files in Go - Michael Stapelberg](https://michael.stapelberg.ch/posts/2017-01-28-golang_atomically_writing/) (MEDIUM confidence)
- [adrg/xdg -- XDG Base Directory Specification for Go](https://github.com/adrg/xdg) (HIGH confidence)
- [Flickering on Windows - Issue #1019](https://github.com/charmbracelet/bubbletea/issues/1019) (HIGH confidence)
- [Resize truncation - Discussion #661](https://github.com/charmbracelet/bubbletea/discussions/661) (HIGH confidence)
- [Lipgloss GitHub repository](https://github.com/charmbracelet/lipgloss) (HIGH confidence)
- [Bubbles component library](https://github.com/charmbracelet/bubbles) (HIGH confidence)
- [Bubble Tea bubbletea DeepWiki - Core Components](https://deepwiki.com/charmbracelet/bubbletea/2-core-components) (MEDIUM confidence)
- [Bubble Tea bubbletea DeepWiki - Component Integration](https://deepwiki.com/charmbracelet/bubbletea/6.5-component-integration) (MEDIUM confidence)

---
*Pitfalls research for: Go TUI Calendar + Todo App (Bubble Tea)*
*Researched: 2026-02-05*
