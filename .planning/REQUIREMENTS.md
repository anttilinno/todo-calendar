# Requirements: Todo Calendar

**Defined:** 2026-02-05
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen.

## v1.1 Requirements

Requirements for v1.1 Polish & Personalization milestone.

### Date Indicators

- [x] **INDI-01**: Calendar dates with incomplete todos display with bracket indicators `[N]`
- [x] **INDI-02**: Dates with only completed todos render without indicators (normal display)
- [x] **INDI-03**: Calendar grid alignment is maintained when indicators are present

### Todo Editing

- [ ] **EDIT-01**: User can press `e` to edit selected todo's text in-place (reuses text input)
- [ ] **EDIT-02**: User can change a todo's date (add, modify, or remove date)
- [ ] **EDIT-03**: Edited todos persist to disk immediately

### First Day of Week

- [x] **FDOW-01**: User can set first day of week (Monday or Sunday) in config.toml
- [x] **FDOW-02**: Calendar grid renders with configured first day of week
- [x] **FDOW-03**: Day-of-week header row reflects the configured start day

### Themes

- [ ] **THME-01**: App ships with 4 preset themes: Dark, Light, Nord, Solarized
- [ ] **THME-02**: User can select theme in config.toml
- [ ] **THME-03**: All UI elements (borders, highlights, text, holidays) respect selected theme
- [ ] **THME-04**: Dark theme is the default when no theme is configured

## Future Requirements

Deferred to later milestones.

### Todo Management

- **REOR-01**: User can reorder todos (move up/down)

### Views

- **VIEW-01**: Weekly calendar view
- **RECR-01**: Simple recurring todos
- **SRCH-01**: Search/filter todos

## Out of Scope

| Feature | Reason |
|---------|--------|
| Full custom color config | Preset themes sufficient for v1.1; per-element overrides add complexity |
| Day selection / cursor | Month-level navigation is sufficient |
| Syncing / cloud storage | Local file only |
| Priority levels or tags | Keep it minimal |
| CalDAV integration | Complexity explosion |
| Subtasks / nesting | Flat list is sufficient |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| INDI-01 | Phase 4 | Complete |
| INDI-02 | Phase 4 | Complete |
| INDI-03 | Phase 4 | Complete |
| EDIT-01 | Phase 5 | Pending |
| EDIT-02 | Phase 5 | Pending |
| EDIT-03 | Phase 5 | Pending |
| FDOW-01 | Phase 4 | Complete |
| FDOW-02 | Phase 4 | Complete |
| FDOW-03 | Phase 4 | Complete |
| THME-01 | Phase 6 | Pending |
| THME-02 | Phase 6 | Pending |
| THME-03 | Phase 6 | Pending |
| THME-04 | Phase 6 | Pending |

**Coverage:**
- v1.1 requirements: 13 total
- Mapped to phases: 13
- Unmapped: 0

---
*Requirements defined: 2026-02-05*
*Last updated: 2026-02-05 after roadmap creation*
