# Project Milestones: Todo Calendar

## v1.5 UX Polish (Shipped: 2026-02-07)

**Delivered:** Visual overhaul with styled checkboxes and section separators, full-pane editing with dual-field forms, mode-aware help bar, and 7 pre-built markdown templates seeded on first launch

**Phases completed:** 17-19 (5 plans total)

**Key accomplishments:**
- Styled checkboxes (accent/green) with section separators and vertical spacing for easier scanning
- Mode-aware help bar showing 5 keys in normal mode, Enter/Esc in input modes, full list via ? toggle
- Full-pane editing forms replacing inline inputs, with dual-field dated-add and Tab field switching
- 7 pre-built templates (Meeting Notes, Checklist, Daily Plan, Bug Report, Feature Spec, PR Checklist, Code Review) seeded via version-3 migration

**Stats:**
- 28 files changed, 4326 insertions(+), 71 deletions(-)
- 5,209 lines of Go across 33 source files
- 3 phases, 5 plans
- 1 day (2026-02-07)

**Git range:** `77f6f57` → `777db03`

**What's next:** v2 candidates — simple recurring todos

---

## v1.4 Data & Editing (Shipped: 2026-02-06)

**Delivered:** SQLite database backend, markdown todo bodies with reusable templates, and external editor integration for power-user editing workflows

**Phases completed:** 14-16 (6 plans total)

**Key accomplishments:**
- Replaced JSON storage with SQLite (modernc.org/sqlite pure Go driver, WAL mode, PRAGMA user_version migrations)
- Extracted TodoStore interface decoupling all consumers from storage backend
- Markdown todo bodies with glamour-rendered preview overlay and [+] body indicators
- Reusable markdown templates with {{.Variable}} placeholder prompting and multi-line textarea input
- External editor integration ($VISUAL/$EDITOR/vi fallback) with content change detection and clean TUI lifecycle

**Stats:**
- 44 files changed, 6040 insertions(+), 1475 deletions(-)
- 4,670 lines of Go across the codebase
- 3 phases, 6 plans
- 1 day (2026-02-06)

**Git range:** `16860da` → `299f5c4`

**What's next:** v1.5 or v2 candidates — recurring todos, FTS5 search, or other enhancements

---

## v1.3 Views & Usability (Shipped: 2026-02-06)

**Delivered:** Enhanced calendar views (weekly toggle), search/filter (inline + full-screen), date format presets, and color-coded overview counts

**Phases completed:** 10-13 (5 plans total)

**Key accomplishments:**
- Overview panel shows color-coded pending/completed counts per month across all 4 themes
- 3-preset date format (ISO/EU/US) configurable in settings with format-aware input
- Weekly calendar view with `w` toggle, week navigation, and auto-select current week
- Inline todo filter (`/`) with real-time case-insensitive narrowing and Esc to clear
- Full-screen search overlay (Ctrl+F) for cross-month todo discovery with jump-to-month

**Stats:**
- 31 files changed
- 3,263 lines of Go across 23 source files
- 4 phases, 5 plans
- 2 days (2026-02-05 → 2026-02-06)

**Git range:** `50c009d` → `e43b5f4`

**What's next:** v1.4 Data & Editing — SQLite backend, markdown templates, external editor integration

---

## v1.2 Reorder & Settings (Shipped: 2026-02-06)

**Delivered:** Todo reordering, in-app settings overlay with live theme preview, and calendar overview panel with per-month todo counts

**Phases completed:** 7-9 (5 plans total)

**Key accomplishments:**
- Todo reordering with J/K keybindings, SortOrder persistence, and section boundary checks
- Full-screen settings overlay with live theme preview, country and week-start config
- Save/cancel settings with config.toml persistence and theme revert on cancel
- Calendar overview panel showing per-month todo counts and floating todo count below grid
- All 11 requirements shipped without scope changes

**Stats:**
- 30 files changed
- 2,492 lines of Go across 20 source files
- 3 phases, 5 plans
- 1 day (2026-02-06)

**Git range:** `5823244` → `cd70d07`

**What's next:** v2 candidates — weekly calendar view, recurring todos, search/filter

---

## v1.1 Polish & Personalization (Shipped: 2026-02-05)

**Delivered:** Calendar date indicators, todo editing, configurable first day of week, and 4 preset color themes for a more informative and personalized experience

**Phases completed:** 4-6 (6 plans total)

**Key accomplishments:**
- Calendar dates with incomplete todos display bracket indicators `[N]` for at-a-glance task awareness
- Todo text and date editing without delete-and-recreate workflow
- Configurable first day of week (Monday/Sunday) in config.toml
- 4 preset color themes (Dark, Light, Nord, Solarized) with semantic color roles
- Styles struct + constructor DI pattern replacing all package-level style vars
- Theme propagation through all UI layers: calendar, todolist, app borders, and help bar

**Stats:**
- 28 files changed
- 1,695 lines of Go across 18 source files
- 3 phases, 6 plans
- 1 day from v1.0 to v1.1

**Git range:** `f8d644b` → `5e45736`

**What's next:** Todo reordering, weekly view, recurring todos, or search/filter.

---

## v1.0 MVP (Shipped: 2026-02-05)

**Delivered:** Terminal-based calendar+todo app with split-pane TUI, monthly calendar with holidays, and todo management with persistence

**Phases completed:** 1-3 (5 plans total)

**Key accomplishments:**
- Split-pane TUI scaffold with Bubble Tea, Tab focus routing, and responsive resize handling
- Monthly calendar grid with today highlight and configurable national holiday display (11 countries)
- Todo CRUD with three-mode input system (normal/text/date) and input isolation
- Atomic JSON persistence with XDG-compliant paths (~/.config/todo-calendar/)
- Context-sensitive help bar with calendar-todo month synchronization

**Stats:**
- 44 files created
- 1,325 lines of Go
- 3 phases, 5 plans
- 1 day from start to ship

**Git range:** `0a2acaf` → `b9bfca7`

**What's next:** Project functionally complete for v1. Potential v2 enhancements: todo editing, reordering, calendar date indicators, color themes, configurable first day of week.

---
