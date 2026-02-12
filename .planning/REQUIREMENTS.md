# Requirements: Todo Calendar

**Defined:** 2026-02-12
**Core Value:** See your month at a glance — calendar with holidays and todos in one terminal screen

## v1.9 Requirements

Requirements for fuzzy date todos milestone. Each maps to roadmap phases.

### Date Precision

- [x] **DATE-01**: User can create a month-level todo by filling only mm + yyyy fields (dd left blank)
- [x] **DATE-02**: User can create a year-level todo by filling only yyyy field (mm and dd left blank)
- [x] **DATE-03**: Date input uses segmented fields (dd / mm / yyyy) with Tab navigation between them instead of typed separators
- [x] **DATE-04**: Segmented date field respects configured date format (ISO: yyyy-mm-dd, EU: dd.mm.yyyy, US: mm/dd/yyyy) for field order

### Todo Sections

- [ ] **SECT-01**: Month-level todos appear in a dedicated "This Month" section in the todo panel
- [ ] **SECT-02**: Year-level todos appear in a dedicated "This Year" section in the todo panel
- [ ] **SECT-03**: Month section shows todos matching the currently viewed month
- [ ] **SECT-04**: Year section shows todos matching the currently viewed year

### Calendar Indicators

- [ ] **INDIC-01**: Left-side circle indicator on calendar shows month-todo status (red = pending, green = all done)
- [ ] **INDIC-02**: Right-side circle indicator on calendar shows year-todo status (red = pending, green = all done)
- [ ] **INDIC-03**: Indicators only appear when there are month/year todos for the viewed period

### Settings

- [ ] **SET-01**: Setting to show/hide month-level todo section in the todo panel
- [ ] **SET-02**: Setting to show/hide year-level todo section in the todo panel
- [ ] **SET-03**: Settings accessible in the existing settings overlay with live preview

### View Integration

- [ ] **VIEW-01**: Fuzzy date todos (month/year) only appear in monthly calendar view, not weekly view

## Future Requirements

### Complex Recurring

- **RECUR-01**: Complex recurring cadences ("every 2nd Tuesday", "last Friday of month")

## Out of Scope

| Feature | Reason |
|---------|--------|
| Quarter-level todos | Overcomplicates date precision hierarchy |
| Fuzzy todos in weekly view | User decided monthly view only |
| Drag-and-drop date refinement | TUI constraint, manual edit sufficient |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| DATE-01 | Phase 27 | Done |
| DATE-02 | Phase 27 | Done |
| DATE-03 | Phase 27 | Done |
| DATE-04 | Phase 27 | Done |
| SECT-01 | Phase 28 | Pending |
| SECT-02 | Phase 28 | Pending |
| SECT-03 | Phase 28 | Pending |
| SECT-04 | Phase 28 | Pending |
| INDIC-01 | Phase 28 | Pending |
| INDIC-02 | Phase 28 | Pending |
| INDIC-03 | Phase 28 | Pending |
| SET-01 | Phase 29 | Pending |
| SET-02 | Phase 29 | Pending |
| SET-03 | Phase 29 | Pending |
| VIEW-01 | Phase 28 | Pending |

**Coverage:**
- v1.9 requirements: 15 total
- Mapped to phases: 15
- Unmapped: 0

---
*Requirements defined: 2026-02-12*
*Last updated: 2026-02-12 after roadmap creation*
