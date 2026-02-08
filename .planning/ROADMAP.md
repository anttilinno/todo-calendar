# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.7: See MILESTONES.md

### v1.8 Weekly Todo Filtering (In Progress)

**Milestone Goal:** Weekly view shows only that week's todos, not the entire month.

## Phases

- [ ] **Phase 26: Weekly Todo Filtering** - Todo panel scopes to the visible week when weekly view is active

## Phase Details

### Phase 26: Weekly Todo Filtering
**Goal**: Users see only the current week's todos (plus floating items) when in weekly view, with instant updates on navigation
**Depends on**: Nothing (standalone feature on existing weekly view)
**Requirements**: WKLY-01, WKLY-02, WKLY-03
**Success Criteria** (what must be TRUE):
  1. When user presses `w` to enter weekly view, the todo panel shows only todos dated within that week's Monday-Sunday (or Sunday-Saturday per config) range
  2. Undated (floating) todos remain visible in the todo panel regardless of which week is selected
  3. When user presses `h`/`l` to navigate to a different week, the todo panel immediately updates to show only that week's dated todos
  4. When user presses `w` again to return to monthly view, the todo panel reverts to showing all todos for the full month
**Plans**: 1 plan

Plans:
- [ ] 26-01-PLAN.md -- Add date-range store query, calendar getter, todolist week filter, and app wiring

## Progress

**Execution Order:** Phase 26

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 23. Cleanup & Calendar Polish | v1.7 | 2/2 | Complete | 2026-02-07 |
| 24. Unified Add Form | v1.7 | 1/1 | Complete | 2026-02-07 |
| 25. Template Picker Integration | v1.7 | 1/1 | Complete | 2026-02-07 |
| 26. Weekly Todo Filtering | v1.8 | 0/1 | Not started | - |
