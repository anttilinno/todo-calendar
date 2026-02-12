# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.8: Shipped (see MILESTONES.md)
- **v1.9 Fuzzy Date Todos** - Phases 27-29 (in progress)

## Phases

### v1.9 Fuzzy Date Todos (In Progress)

**Milestone Goal:** Support month-level and year-level todos with calendar indicators and visibility toggles

**Phase Numbering:**
- Integer phases (27, 28, 29): Planned milestone work
- Decimal phases (27.1, 27.2): Urgent insertions (marked with INSERTED)

- [ ] **Phase 27: Date Precision & Input** - Storage schema for date precision and segmented date field UI
- [ ] **Phase 28: Display & Indicators** - Todo panel sections and calendar circle indicators for fuzzy todos
- [ ] **Phase 29: Settings & View Filtering** - Show/hide toggles and weekly view exclusion

## Phase Details

### Phase 27: Date Precision & Input
**Goal**: Users can create todos with month-level or year-level precision using a segmented date field
**Depends on**: Nothing (first phase of v1.9)
**Requirements**: DATE-01, DATE-02, DATE-03, DATE-04
**Success Criteria** (what must be TRUE):
  1. User can create a month-level todo by entering only month and year in the date field (leaving day blank)
  2. User can create a year-level todo by entering only the year (leaving day and month blank)
  3. Date input shows three separate segments (dd / mm / yyyy) with Tab moving between them instead of typing separators
  4. Segment order matches the configured date format (ISO shows yyyy-mm-dd segments, EU shows dd.mm.yyyy, US shows mm/dd/yyyy)
**Plans**: 2 plans

Plans:
- [ ] 27-01-PLAN.md -- Schema migration, date precision storage, and precision-aware store queries
- [ ] 27-02-PLAN.md -- Segmented date input UI with format-aware ordering and precision derivation

### Phase 28: Display & Indicators
**Goal**: Users can see their fuzzy-date todos in dedicated sections and spot month/year status at a glance on the calendar
**Depends on**: Phase 27
**Requirements**: SECT-01, SECT-02, SECT-03, SECT-04, INDIC-01, INDIC-02, INDIC-03, VIEW-01
**Success Criteria** (what must be TRUE):
  1. Month-level todos appear in a "This Month" section that updates when navigating months
  2. Year-level todos appear in a "This Year" section that updates when navigating to a different year
  3. Calendar displays a left circle indicator for month-todo status and a right circle for year-todo status (red = pending, green = all done), only when relevant todos exist
  4. Fuzzy-date todos (month/year) do not appear in weekly view
**Plans**: TBD

Plans:
- [ ] 28-01: TBD

### Phase 29: Settings & View Filtering
**Goal**: Users can toggle visibility of month and year todo sections from the settings overlay
**Depends on**: Phase 28
**Requirements**: SET-01, SET-02, SET-03
**Success Criteria** (what must be TRUE):
  1. Settings overlay has toggles to show or hide the month-level todo section
  2. Settings overlay has toggles to show or hide the year-level todo section
  3. Toggling settings takes effect immediately with live preview, and persists after save
**Plans**: TBD

Plans:
- [ ] 29-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 27 -> 28 -> 29

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 27. Date Precision & Input | 0/2 | In progress | - |
| 28. Display & Indicators | 0/? | Not started | - |
| 29. Settings & View Filtering | 0/? | Not started | - |
