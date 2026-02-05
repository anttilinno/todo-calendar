# Project Milestones: Todo Calendar

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
