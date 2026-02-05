# Requirements: Todo Calendar

**Defined:** 2026-02-05
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen.

## v1 Requirements

### Calendar

- [x] **CAL-01**: App displays a monthly calendar grid with day-of-week headers (like `cal`)
- [x] **CAL-02**: User can navigate between months (next/prev)
- [x] **CAL-03**: Today's date is visually highlighted on the calendar
- [x] **CAL-04**: National holidays are displayed in red on the calendar
- [x] **CAL-05**: Country for holidays is configurable

### Todo

- [x] **TODO-01**: User can add a todo with text and optional date
- [x] **TODO-02**: User can mark a todo as complete (visual checkmark/strikethrough)
- [x] **TODO-03**: User can delete a todo
- [x] **TODO-04**: Date-bound todos are shown for the currently viewed month
- [x] **TODO-05**: Floating (undated) todos are shown in a separate section

### UI/UX

- [x] **UI-01**: Split-pane layout with calendar on left, todo list on right
- [x] **UI-02**: Keyboard navigation (arrows/vim keys, Tab to switch panes)
- [x] **UI-03**: Help bar showing available keybindings
- [x] **UI-04**: Layout responds to terminal resize

### Data

- [x] **DATA-01**: Todos persist to a local JSON file
- [x] **DATA-02**: Configuration stored in TOML file (country, preferences)
- [x] **DATA-03**: Data stored in XDG-compliant paths (~/.config/todo-calendar/)

## v2 Requirements

### Enhancements

- **EDIT-01**: User can edit todo text and date after creation
- **EDIT-02**: User can reorder todos (move up/down)
- **VIS-01**: Todo indicators (dots/counts) shown on calendar dates that have todos
- **THEME-01**: Color themes / customization
- **CFG-01**: Configurable first day of week (Monday vs Sunday)

### Future

- **VIEW-01**: Weekly calendar view
- **REC-01**: Simple recurring todos (daily/weekly/monthly)
- **SEARCH-01**: Search/filter todos
- **EXPORT-01**: iCalendar export
- **TAG-01**: Tags/categories with colors

## Out of Scope

| Feature | Reason |
|---------|--------|
| Day-by-day selection | Month-level navigation is sufficient; user has few items per month |
| CalDAV / cloud sync | Complexity explosion; local file storage only for v1 |
| Recurring todos | RRULE complexity; manual re-add is sufficient for v1 |
| Subtasks / nesting | Changes data model from flat list to tree; v1 stays flat |
| Notifications / reminders | Requires platform-specific integration; out of scope for TUI |
| Time-blocked appointments | This is a todo app with calendar view, not a scheduling app |
| OAuth / multi-user | Personal single-user tool |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CAL-01 | Phase 2 | Complete |
| CAL-02 | Phase 2 | Complete |
| CAL-03 | Phase 2 | Complete |
| CAL-04 | Phase 2 | Complete |
| CAL-05 | Phase 2 | Complete |
| TODO-01 | Phase 3 | Complete |
| TODO-02 | Phase 3 | Complete |
| TODO-03 | Phase 3 | Complete |
| TODO-04 | Phase 3 | Complete |
| TODO-05 | Phase 3 | Complete |
| UI-01 | Phase 1 | Complete |
| UI-02 | Phase 1 | Complete |
| UI-03 | Phase 3 | Complete |
| UI-04 | Phase 1 | Complete |
| DATA-01 | Phase 3 | Complete |
| DATA-02 | Phase 2 | Complete |
| DATA-03 | Phase 3 | Complete |

**Coverage:**
- v1 requirements: 17 total
- Mapped to phases: 17
- Unmapped: 0

---
*Requirements defined: 2026-02-05*
*Last updated: 2026-02-05 after roadmap creation*
