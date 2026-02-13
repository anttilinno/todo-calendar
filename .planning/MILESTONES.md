# Project Milestones: Todo Calendar

## v2.0 Settings UX (Shipped: 2026-02-12)

**Delivered:** Settings overlay now applies changes immediately on value cycle and closes with Esc — no save button or cancel flow

**Phases completed:** 30 (1 phase, implemented directly)

**Key accomplishments:**
- Settings save-on-close: every value change (h/l arrows) immediately persists to config.toml via SettingChangedMsg
- Removed Save/Cancel key bindings and savedConfig revert mechanism
- Footer updated from "enter save / esc cancel" to "esc close"

**Stats:**
- 3 files changed, 22 insertions(+), 54 deletions(-)
- 8,177 lines of Go across 35 source files
- 1 phase, 2 requirements

**Git range:** `cfd7fad` → `1f47fd3`

**What's next:** New feature milestones

---

## v1.9 Fuzzy Date Todos (Shipped: 2026-02-12)

**Delivered:** Month-level and year-level todo precision with dedicated sections, calendar circle indicators, and configurable show/hide toggles in the settings overlay

**Phases completed:** 27-29 (5 plans total)

**Key accomplishments:**
- Date precision storage schema with SQLite migration, MonthTodos/YearTodos queries, and day-query exclusion for clean separation
- Segmented 3-field date input (dd/mm/yyyy) with format-aware ordering (ISO/EU/US) and precision derivation from empty segments
- 4-section todo panel (dated, This Month, This Year, Floating) with section-aware reordering boundaries
- Calendar circle indicators on title line showing month/year todo status (red = pending, green = all done)
- Show/hide toggles in settings overlay with TOML persistence and live preview on save

**Stats:**
- 29 files changed, 2,458 insertions(+), 209 deletions(-)
- 8,179 lines of Go across 35 source files
- 3 phases, 5 plans, 8 tasks
- 7 days (2026-02-05 → 2026-02-12)

**Git range:** `929eb3a` → `854f160`

**What's next:** v2 candidates or new feature milestones

---

## v1.8 Weekly Todo Filtering (Shipped: 2026-02-08)

**Delivered:** Weekly view now filters todo panel to show only that week's dated todos plus floating items, with instant updates on navigation

**Phases completed:** 26 (1 plan total)

**Key accomplishments:**
- TodosForDateRange store method for date-range query enabling week-scoped todo retrieval
- Week filter state in todolist model with SetWeekFilter/ClearWeekFilter and conditional visibleItems logic
- syncTodoView app helper centralizing view-mode-aware sync, replacing 3 scattered SetViewMonth calls
- Improved contrast for MutedFg/EmptyFg in Nord and Solarized themes for better readability

**Stats:**
- 11 files changed, 390 insertions(+), 27 deletions(-)
- 7,469 lines of Go across 35 source files
- 1 phase, 1 plan, 2 tasks
- Same day (2026-02-08, ~17 minutes)

**Git range:** `7957a04` → `1ca2379`

**What's next:** Additional weekly view enhancements or new feature milestones

---

## v1.7 Unified Add Flow & Polish (Shipped: 2026-02-07)

**Delivered:** Unified todo creation into a single full-pane form with template picker integration, fixed today indicator blending, and removed dead code

**Phases completed:** 23-25 (4 plans total)

**Key accomplishments:**
- Unified add flow: single `a` key opens 4-field form (title, date, body, template) replacing 3 separate entry points (`a`/`A`/`t`)
- Template picker integrated into add form with j/k navigation, placeholder prompting, and Title/Body pre-fill
- Today calendar indicator now blends pending/done status colors with today highlight
- Removed 799+ lines of dead code (JSON store, old keybindings, 3 unused modes, 10 struct fields)
- TodoStore interface extracted to dedicated `iface.go` for clean dependency injection

**Stats:**
- 17 files changed, 2,394 insertions(+), 359 deletions(-)
- 7,239 lines of Go across 35 source files
- 3 phases, 4 plans
- 1 day (2026-02-07)

**Git range:** `54e50a7` → `70f99c8`

**What's next:** v2 candidates or new feature milestones

---

## v1.6 Templates & Recurring (Shipped: 2026-02-07)

**Delivered:** Template management overlay with CRUD operations, recurring schedule system with 4 cadence types, and auto-creation engine that generates scheduled todos on app launch

**Phases completed:** 20-22 (8 plans total)

**Key accomplishments:**
- Full-screen template management overlay (M key) with list, preview, rename, delete, and external editor integration
- ScheduleRule engine supporting daily, weekdays, weekly (day selection), and monthly (day-of-month with clamping) cadences
- SQLite schema extensions: schedules table with FK CASCADE, todos schedule columns with dedup UNIQUE index
- AutoCreate engine running on startup, generating todos for 7-day rolling window with deduplication
- Schedule picker UI with cadence cycling, weekly day toggling, and monthly day input
- Placeholder defaults prompting at schedule creation with JSON serialization and pre-fill on edit

**Stats:**
- 16 files changed, 2,362 insertions(+)
- 7,624 lines of Go across 35 source files
- 3 phases, 8 plans
- 2 days (2026-02-05 to 2026-02-07)

**Git range:** `4aa445a` → `768b6f3`

**What's next:** v2 candidates or new feature milestones

---

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


## v2.1 Priorities (Shipped: 2026-02-13)

**Delivered:** P1-P4 priority levels with color-coded badges across todo list, calendar, and search — inline selector in edit form, theme-aware colors for all 4 themes

**Phases completed:** 31-32 (2 phases, 3 plans, 6 tasks)

**Key accomplishments:**
- SQLite schema v7 with priority INTEGER column and zero-migration defaults for existing todos
- Priority theme colors (P1=red, P2=orange, P3=blue, P4=grey) for all 4 themes (Dark, Light, Nord, Solarized)
- Inline priority selector in add/edit forms with left/right arrow cycling (none/P1-P4)
- Colored [P1]-[P4] badge rendering in todo list with fixed 5-char slot for column alignment
- Priority-colored calendar day indicators in both monthly and weekly grids
- Priority badge rendering in search results matching todolist pattern

**Stats:**
- 22 files changed, 2,075 insertions(+), 161 deletions(-)
- 8,644 lines of Go across 35 source files
- 2 phases, 3 plans, 6 tasks
- Same day (2026-02-13)

**Git range:** `9dc7855` → `a730bd7`

**What's next:** New feature milestones

---

