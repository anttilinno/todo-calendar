# Todo Calendar

## What This Is

A terminal-based (TUI) application that combines a monthly calendar view with a todo list. The left panel shows a navigable calendar with national holidays and date indicators showing pending work. The right panel displays todos for the visible month alongside undated (floating) items. Supports editing todos after creation, configurable first day of week, and 4 color themes. Built with Go and Bubble Tea for personal use.

## Core Value

See your month at a glance — calendar with holidays and todos in one terminal screen.

## Requirements

### Validated

- Monthly calendar view (left panel) resembling `cal` output — v1.0
- Navigate between months (next/prev) — v1.0
- National holidays highlighted in red on calendar — v1.0
- Configurable country for holidays (11 countries supported) — v1.0
- Todo list (right panel) showing month's date-bound todos + floating todos — v1.0
- Add todo with optional date — v1.0
- Check off (complete) a todo — v1.0
- Delete a todo — v1.0
- Persist todos to a local JSON file on disk — v1.0
- Split-pane TUI layout (calendar left, todos right) — v1.0
- Keyboard navigation with vim keys and Tab focus switching — v1.0
- Context-sensitive help bar — v1.0
- Responsive terminal resize handling — v1.0
- TOML configuration file — v1.0
- XDG-compliant data paths — v1.0
- Calendar dates with incomplete todos display bracket indicators `[N]` — v1.1
- Edit todo text and date after creation — v1.1
- Configurable first day of week (Monday/Sunday) — v1.1
- 4 preset color themes (Dark, Light, Nord, Solarized) selectable in config — v1.1

### Active

**Current Milestone: v1.2 — Reorder & Settings**

**Goal:** Let users reorder todos and configure the app from an in-app settings page instead of editing config.toml manually.

- [ ] Reorder todos (move up/down with keybindings, persist order)
- [ ] In-app settings page (full-screen overlay, live preview)
- [ ] Settings: change theme with immediate visual feedback
- [ ] Settings: change country for holidays
- [ ] Settings: change first day of week
- [ ] Settings: save changes to config.toml

### v2 Candidates

- Weekly calendar view
- Simple recurring todos
- Search/filter todos

### Out of Scope

- Individual day selection / day-by-day arrow navigation — month-level navigation is sufficient
- Syncing / cloud storage — local file only
- Priority levels or tags — keep it minimal
- CalDAV integration — complexity explosion
- Subtasks / nesting — flat list is sufficient
- Notifications / reminders — out of scope for TUI
- Time-blocked appointments — this is a todo app, not a scheduler

## Context

- **Stack:** Go 1.25.6, Bubble Tea v1.3.10, Lipgloss v1.1.0, Bubbles v0.21.1
- **Holidays:** rickar/cal/v2 with 11-country registry (de, dk, ee, es, fi, fr, gb, it, no, se, us)
- **Config:** TOML at ~/.config/todo-calendar/config.toml (BurntSushi/toml v1.6.0)
- **Storage:** JSON at ~/.config/todo-calendar/todos.json with atomic writes
- **Codebase:** 1,695 lines of Go across 18 source files
- **Architecture:** Elm Architecture (Bubble Tea), pure rendering functions, constructor DI

## Constraints

- **Stack**: Go + Bubble Tea — chosen for ergonomic component model and ecosystem
- **Storage**: Local file only — no database, no network dependencies
- **Holidays**: Must work offline using bundled Go library, not an external API

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go + Bubble Tea over Rust + Ratatui | Better component model for split-pane layout, gentler learning curve | ✓ Good — clean architecture, fast development |
| Month-level navigation, no day selection | User doesn't have many items — showing all month todos is simpler | ✓ Good — keeps UI simple |
| Local JSON file over SQLite | Simpler, more portable, sufficient for personal use | ✓ Good — atomic writes keep data safe |
| Configurable country holidays via Go library | Offline, no API dependency, flexible | ✓ Good — 11 countries supported |
| String dates (YYYY-MM-DD) over time.Time | Prevents timezone corruption during JSON round-trips | ✓ Good — research-informed decision |
| Atomic file writes (CreateTemp+Sync+Rename) | Data safety from day one | ✓ Good — prevents corruption |
| Pure rendering functions | RenderGrid has no side effects, testable | ✓ Good — clean separation |
| Three-mode input state machine | Cleanly separates key handling for normal/text/date | ✓ Good — input isolation works well |
| `first_day_of_week` string over `monday_start` bool | More extensible, clearer semantics | ✓ Good — clean config field |
| 4-char calendar cells (grid 34 chars wide) | Room for bracket indicators `[N]` without breaking alignment | ✓ Good — fits well |
| Semantic theme color roles (14 fields) | Named by role not component, decoupled from UI structure | ✓ Good — clean theme propagation |
| Styles struct + constructor DI over package-level vars | Enables runtime theme switching, testable | ✓ Good — no global state |
| Empty string = terminal default in Dark theme | Respects user's terminal palette | ✓ Good — non-invasive default |
| Settings as full-screen overlay with live preview | User wants to see changes immediately; overlay avoids cramming into split pane | — Pending |

## Known Tech Debt

- Store.Save() errors ignored in CRUD methods (silent persistence failures on disk errors)

---
*Last updated: 2026-02-06 after v1.2 milestone started*
