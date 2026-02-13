# Requirements: Todo Calendar

**Defined:** 2026-02-12
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen

## v2.1 Requirements

Requirements for milestone v2.1: Priorities.

### Priority Levels

- [ ] **PRIO-01**: User can set priority (P1-P4 or none) on a todo via the edit form
- [ ] **PRIO-02**: Todos display a colored [P1]-[P4] badge prefix with priority-colored text
- [ ] **PRIO-03**: Completed prioritized todos show colored badge but grey strikethrough text
- [ ] **PRIO-04**: Priority badge uses fixed-width slot for consistent column alignment
- [ ] **PRIO-05**: Priority colors defined for all 4 themes (Dark, Light, Nord, Solarized)
- [ ] **PRIO-06**: Calendar day indicators reflect highest-priority incomplete todo's color
- [ ] **PRIO-07**: Search results display priority badges
- [ ] **PRIO-08**: Priority stored as INTEGER (0=none, 1-4) in SQLite with migration v7
- [ ] **PRIO-09**: Existing todos default to priority 0 (no priority) after migration

## Future Requirements

Deferred to future releases.

### Complex Recurring

- **RECR-01**: Complex recurring cadences ("every 2nd Tuesday", "last Friday of month")

### Completed Tasks Archive

- **ARCH-01**: User can browse/review past completed todos by date

### Natural Language Dates

- **NLDT-01**: NL date expressions ("tomorrow", "next friday", "jan 15", "in 3 days")
- **NLDT-02**: Parsed date preview in date field
- **NLDT-03** through **NLDT-10**: Full NL date input replacing segmented fields

### Priority Enhancements

- **PRIO-10**: Default priority configurable in settings
- **PRIO-11**: Inline priority cycling in normal mode (press 1-4 on selected todo)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Auto-sort by priority | Conflicts with manual J/K reordering -- priority is visual only |
| Priority cycling in normal mode | Risk of accidental changes; set via edit form only |
| Default priority setting | Polish item -- defer to future |
| NL parsing for complex recurrence | "every 2nd tuesday" is a v2 candidate, not NL date input |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| PRIO-01 | Phase 32 | Pending |
| PRIO-02 | Phase 32 | Pending |
| PRIO-03 | Phase 32 | Pending |
| PRIO-04 | Phase 32 | Pending |
| PRIO-05 | Phase 32 | Pending |
| PRIO-06 | Phase 32 | Pending |
| PRIO-07 | Phase 32 | Pending |
| PRIO-08 | Phase 31 | Pending |
| PRIO-09 | Phase 31 | Pending |

**Coverage:**
- v2.1 requirements: 9 total (PRIO-01 through PRIO-09)
- Mapped to phases: 9
- Unmapped: 0

---
*Requirements defined: 2026-02-12*
*Last updated: 2026-02-12 after roadmap creation*
