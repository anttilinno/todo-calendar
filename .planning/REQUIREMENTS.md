# Requirements: Todo Calendar

**Defined:** 2026-02-05
**Core Value:** See your month at a glance — calendar with holidays and todos in one terminal screen.

## v1 Requirements

### Calendar

- [ ] **CAL-01**: App displays a monthly calendar grid with day-of-week headers (like `cal`)
- [ ] **CAL-02**: User can navigate between months (next/prev)
- [ ] **CAL-03**: Today's date is visually highlighted on the calendar
- [ ] **CAL-04**: National holidays are displayed in red on the calendar
- [ ] **CAL-05**: Country for holidays is configurable

### Todo

- [ ] **TODO-01**: User can add a todo with text and optional date
- [ ] **TODO-02**: User can mark a todo as complete (visual checkmark/strikethrough)
- [ ] **TODO-03**: User can delete a todo
- [ ] **TODO-04**: Date-bound todos are shown for the currently viewed month
- [ ] **TODO-05**: Floating (undated) todos are shown in a separate section

### UI/UX

- [ ] **UI-01**: Split-pane layout with calendar on left, todo list on right
- [ ] **UI-02**: Keyboard navigation (arrows/vim keys, Tab to switch panes)
- [ ] **UI-03**: Help bar showing available keybindings
- [ ] **UI-04**: Layout responds to terminal resize

### Data

- [ ] **DATA-01**: Todos persist to a local JSON file
- [ ] **DATA-02**: Configuration stored in TOML file (country, preferences)
- [ ] **DATA-03**: Data stored in XDG-compliant paths (~/.config/todo-calendar/)

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
| CAL-01 | — | Pending |
| CAL-02 | — | Pending |
| CAL-03 | — | Pending |
| CAL-04 | — | Pending |
| CAL-05 | — | Pending |
| TODO-01 | — | Pending |
| TODO-02 | — | Pending |
| TODO-03 | — | Pending |
| TODO-04 | — | Pending |
| TODO-05 | — | Pending |
| UI-01 | — | Pending |
| UI-02 | — | Pending |
| UI-03 | — | Pending |
| UI-04 | — | Pending |
| DATA-01 | — | Pending |
| DATA-02 | — | Pending |
| DATA-03 | — | Pending |

**Coverage:**
- v1 requirements: 17 total
- Mapped to phases: 0
- Unmapped: 17

---
*Requirements defined: 2026-02-05*
*Last updated: 2026-02-05 after initial definition*
