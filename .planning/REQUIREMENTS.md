# Requirements: Todo Calendar

**Defined:** 2026-02-07
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen

## v1.7 Requirements

Requirements for the Unified Add Flow & Polish milestone.

### Unified Add Flow

- [ ] **ADD-01**: Single `a` key opens full-pane add form replacing separate `a`/`A`/`t` entry points
- [ ] **ADD-02**: Add form has Title, Date (optional), Body (optional), and Template fields with Tab cycling
- [ ] **ADD-03**: Template field opens template picker; selecting a template pre-fills Title and Body
- [ ] **ADD-04**: User can edit pre-filled Title and Body after template selection before saving
- [ ] **ADD-05**: Enter saves from Title/Date fields; Ctrl+D saves from Body/Template fields
- [ ] **ADD-06**: Empty date creates floating todo; filled date creates dated todo
- [ ] **ADD-07**: Remove `A` (dated add) and `t` (template use) key bindings

### Calendar Indicators

- [x] **CAL-01**: Today's date blends today highlight style with pending/done indicator status

### Cleanup

- [x] **CLN-01**: Remove unused JSON store implementation
- [x] **CLN-02**: Remove obsolete key bindings and dead code
- [x] **CLN-03**: Update PROJECT.md validated requirements for recent v1.6+ commits

## Future Requirements

- Complex recurring cadences ("every 2nd Tuesday", "last Friday of month")

## Out of Scope

| Feature | Reason |
|---------|--------|
| Multi-step wizard for add flow | Research showed single form is more intuitive in TUI context |
| Template creation from add form | Use existing `T` key or template overlay (`M`) for creation |
| Inline add (no full-pane) | Consistency with edit mode; full-pane is the established pattern |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| ADD-01 | Phase 24 | Pending |
| ADD-02 | Phase 24 | Pending |
| ADD-03 | Phase 25 | Pending |
| ADD-04 | Phase 25 | Pending |
| ADD-05 | Phase 24 | Pending |
| ADD-06 | Phase 24 | Pending |
| ADD-07 | Phase 24 | Pending |
| CAL-01 | Phase 23 | Complete |
| CLN-01 | Phase 23 | Complete |
| CLN-02 | Phase 23 | Complete |
| CLN-03 | Phase 23 | Complete |

**Coverage:**
- v1.7 requirements: 11 total
- Mapped to phases: 11
- Unmapped: 0

---
*Requirements defined: 2026-02-07*
*Last updated: 2026-02-07 after Phase 23 completion*
