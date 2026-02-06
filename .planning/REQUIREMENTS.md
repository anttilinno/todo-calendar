# Requirements: Todo Calendar

**Defined:** 2026-02-06
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen.

## v1.3 Requirements

Requirements for milestone v1.3 Views & Usability.

### Overview Color Coding

- [x] **OVCLR-01**: Overview shows split count per month: pending (red) and completed (green)
- [x] **OVCLR-02**: Overview colors follow the active theme (not hardcoded red/green)

### Date Format

- [x] **DTFMT-01**: User can choose date display format from 3 presets (YYYY-MM-DD, DD.MM.YYYY, MM/DD/YYYY) in settings
- [x] **DTFMT-02**: All date displays in the app use the chosen format
- [x] **DTFMT-03**: Date format preference persists in config.toml

### Weekly Calendar View

- [ ] **WKVIEW-01**: User can toggle between monthly and weekly calendar view via keybinding
- [ ] **WKVIEW-02**: Weekly view shows 7 days with day numbers, holidays, and todo indicators
- [ ] **WKVIEW-03**: User can navigate forward/backward by week in weekly mode
- [ ] **WKVIEW-04**: Current week is auto-selected when switching from monthly to weekly view

### Search/Filter

- [ ] **SRCH-01**: User can activate inline filter with `/` to filter visible todos by text
- [ ] **SRCH-02**: User can clear inline filter with Esc to return to normal mode
- [ ] **SRCH-03**: User can open full-screen search overlay to find todos across all months
- [ ] **SRCH-04**: Search results show matching todos with their dates
- [ ] **SRCH-05**: User can navigate search results and jump to a selected todo's month

## Future Requirements

Deferred to later milestones.

### Recurring Todos

- **RECUR-01**: User can mark a todo as recurring (daily/weekly/monthly)
- **RECUR-02**: Completed recurring todos auto-generate the next occurrence

## Out of Scope

| Feature | Reason |
|---------|--------|
| Day selection / day-by-day arrow navigation | Month/week-level navigation is sufficient |
| Syncing / cloud storage | Local file only |
| Priority levels or tags | Keep it minimal |
| CalDAV integration | Complexity explosion |
| Subtasks / nesting | Flat list is sufficient |
| Notifications / reminders | Out of scope for TUI |
| Time-blocked appointments | This is a todo app, not a scheduler |
| Fuzzy matching in search | Overkill for short todo text; substring matching is clearer |
| Custom date format in settings UI | 3 presets cover most users; custom via config.toml only |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| OVCLR-01 | Phase 10 | Complete |
| OVCLR-02 | Phase 10 | Complete |
| DTFMT-01 | Phase 11 | Complete |
| DTFMT-02 | Phase 11 | Complete |
| DTFMT-03 | Phase 11 | Complete |
| WKVIEW-01 | Phase 12 | Pending |
| WKVIEW-02 | Phase 12 | Pending |
| WKVIEW-03 | Phase 12 | Pending |
| WKVIEW-04 | Phase 12 | Pending |
| SRCH-01 | Phase 13 | Pending |
| SRCH-02 | Phase 13 | Pending |
| SRCH-03 | Phase 13 | Pending |
| SRCH-04 | Phase 13 | Pending |
| SRCH-05 | Phase 13 | Pending |

**Coverage:**
- v1.3 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0

---
*Requirements defined: 2026-02-06*
*Last updated: 2026-02-06 after phase 11 completion*
