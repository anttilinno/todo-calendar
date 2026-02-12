# Requirements: Todo Calendar

**Defined:** 2026-02-12
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen

## v2.1 Requirements

Requirements for milestone v2.1: Priorities & Smart Dates.

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

### Natural Language Dates

- [ ] **NLDT-01**: User can type NL expressions ("tomorrow", "next friday", "jan 15", "in 3 days") in the date field
- [ ] **NLDT-02**: Date field shows parsed/resolved date preview after input
- [ ] **NLDT-03**: NL parser uses Future direction (dates resolve to upcoming, not past)
- [ ] **NLDT-04**: Year-only input ("2027") produces year precision
- [ ] **NLDT-05**: Month+year input ("march 2026") produces month precision
- [ ] **NLDT-06**: All other NL expressions produce day precision
- [ ] **NLDT-07**: Formatted date input still works ("2026-02-15", "15.02.2026", "02/15/2026" per date format setting)
- [ ] **NLDT-08**: Empty date input creates floating (undated) todo
- [ ] **NLDT-09**: Segmented date input (3-field dd/mm/yyyy) replaced by single NL text field
- [ ] **NLDT-10**: Parsing logic isolated in `internal/nldate` package with unit tests

## Future Requirements

Deferred to future releases.

### Complex Recurring

- **RECR-01**: Complex recurring cadences ("every 2nd Tuesday", "last Friday of month")

### Completed Tasks Archive

- **ARCH-01**: User can browse/review past completed todos by date

### Priority Enhancements

- **PRIO-10**: Default priority configurable in settings
- **PRIO-11**: Inline priority cycling in normal mode (press 1-4 on selected todo)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Auto-sort by priority | Conflicts with manual J/K reordering -- priority is visual only |
| Priority cycling in normal mode | Risk of accidental changes; set via edit form only |
| Default priority setting | Polish item -- defer to future |
| Multi-language NL date parsing | English only, personal use app |
| NL parsing for complex recurrence | "every 2nd tuesday" is a v2 candidate, not NL date input |
| Segmented date input (fallback mode) | Fully replaced by NL text field with precision detection |

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
| NLDT-01 | Phase 33 | Pending |
| NLDT-02 | Phase 33 | Pending |
| NLDT-03 | Phase 33 | Pending |
| NLDT-04 | Phase 33 | Pending |
| NLDT-05 | Phase 33 | Pending |
| NLDT-06 | Phase 33 | Pending |
| NLDT-07 | Phase 33 | Pending |
| NLDT-08 | Phase 33 | Pending |
| NLDT-09 | Phase 33 | Pending |
| NLDT-10 | Phase 33 | Pending |

**Coverage:**
- v2.1 requirements: 19 total
- Mapped to phases: 19
- Unmapped: 0

---
*Requirements defined: 2026-02-12*
*Last updated: 2026-02-12 after roadmap creation*
