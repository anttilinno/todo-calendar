# Requirements: Todo Calendar

**Defined:** 2026-02-06
**Core Value:** See your month at a glance — calendar with holidays and todos in one terminal screen.

## v1.2 Requirements

Requirements for milestone v1.2 Reorder & Settings.

### Todo Reordering

- [x] **REORD-01**: User can move a selected todo up in the list
- [x] **REORD-02**: User can move a selected todo down in the list
- [x] **REORD-03**: Custom order persists across app restarts

### Settings Page

- [x] **SETT-01**: User can open a full-screen settings overlay via keybinding
- [x] **SETT-02**: User can change the color theme with live preview (app redraws immediately)
- [x] **SETT-03**: User can change the holiday country
- [x] **SETT-04**: User can change the first day of week
- [x] **SETT-05**: User can save settings to config.toml and dismiss the overlay
- [x] **SETT-06**: User can dismiss settings without saving (cancel)

### Overview Panel

- [x] **OVRVW-01**: Calendar panel shows todo counts per month below the calendar grid (e.g., `January [7]`)
- [x] **OVRVW-02**: Overview shows count of undated (floating) todos (e.g., `Unknown [12]`)

## Future Requirements

- Weekly calendar view
- Simple recurring todos
- Search/filter todos

## Out of Scope

| Feature | Reason |
|---------|--------|
| Drag-and-drop reordering | Keyboard up/down is sufficient for TUI |
| Settings as side panel | User chose full-screen overlay for clarity |
| Import/export settings | Single config.toml is sufficient |
| Custom theme creation in-app | Preset themes only; custom via config.toml editing |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| REORD-01 | Phase 7 | Complete |
| REORD-02 | Phase 7 | Complete |
| REORD-03 | Phase 7 | Complete |
| SETT-01 | Phase 8 | Complete |
| SETT-02 | Phase 8 | Complete |
| SETT-03 | Phase 8 | Complete |
| SETT-04 | Phase 8 | Complete |
| SETT-05 | Phase 8 | Complete |
| SETT-06 | Phase 8 | Complete |
| OVRVW-01 | Phase 9 | Complete |
| OVRVW-02 | Phase 9 | Complete |

**Coverage:**
- v1.2 requirements: 11 total
- Mapped to phases: 11
- Unmapped: 0

---
*Requirements defined: 2026-02-06*
*Last updated: 2026-02-06 after Phase 9 complete*
