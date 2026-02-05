# Todo Calendar

## What This Is

A terminal-based (TUI) application that combines a monthly calendar view with a todo list. The left panel shows a calendar similar to `cal`, with navigable months and national holidays highlighted in red. The right panel displays todos for the visible month alongside undated (floating) items. Built for personal use with Go and Bubble Tea.

## Core Value

See your month at a glance — calendar with holidays and todos in one terminal screen.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Monthly calendar view (left panel) resembling `cal` output
- [ ] Navigate between months (next/prev)
- [ ] National holidays highlighted in red on calendar
- [ ] Configurable country for holidays (using Go holidays library)
- [ ] Todo list (right panel) showing month's date-bound todos + floating todos
- [ ] Add todo with optional date
- [ ] Check off (complete) a todo
- [ ] Delete a todo
- [ ] Persist todos to a local file on disk
- [ ] Split-pane TUI layout (calendar left, todos right)

### Out of Scope

- Individual day selection / day-by-day arrow navigation — month-level navigation is sufficient
- Todo editing (title, date changes) — keep v1 simple: add, check, delete
- Recurring todos — complexity not warranted for v1
- Syncing / cloud storage — local file only
- Priority levels or tags — keep it minimal

## Context

- Go + Bubble Tea (Charm) for the TUI framework
- Bubbles component library for lists, text inputs
- Lipgloss for styling (colors, layout)
- Go holidays library for national holiday data
- Local file storage (format TBD — likely JSON or plain text)
- Target: single-user personal productivity tool

## Constraints

- **Stack**: Go + Bubble Tea — chosen for ergonomic component model and ecosystem
- **Storage**: Local file only — no database, no network dependencies
- **Holidays**: Must work offline using bundled Go library, not an external API
- **Simplicity**: v1 is add/check/delete only — no editing, no recurring, no priorities

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go + Bubble Tea over Rust + Ratatui | Better component model for split-pane layout, gentler learning curve, good ecosystem | — Pending |
| Month-level navigation, no day selection | User doesn't have many items — showing all month todos is simpler | — Pending |
| Local file over SQLite | Simpler, more portable, sufficient for personal use | — Pending |
| Configurable country holidays via Go library | Offline, no API dependency, flexible | — Pending |

---
*Last updated: 2026-02-05 after initialization*
