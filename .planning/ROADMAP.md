# Roadmap: Todo Calendar

## Milestones

- v1.0 through v1.6: See MILESTONES.md
- **v1.7 Unified Add Flow & Polish** - Phases 23-25 (in progress)

## Phases

### v1.7 Unified Add Flow & Polish (In Progress)

**Milestone Goal:** Unify the three separate todo creation flows into a single full-pane form, fix today indicator blending, and clean up dead code from previous milestones.

**Phase Numbering:**
- Integer phases (23, 24, 25): Planned milestone work
- Decimal phases (23.1, 24.1): Urgent insertions (marked with INSERTED)

- [x] **Phase 23: Cleanup & Calendar Polish** - Remove dead code, fix today indicator, update docs
- [ ] **Phase 24: Unified Add Form** - Single `a` key opens full-pane add form with title/date/body fields
- [ ] **Phase 25: Template Picker Integration** - Template field in add form with pre-fill and editing

## Phase Details

### Phase 23: Cleanup & Calendar Polish
**Goal**: Dead code is removed and the calendar today indicator correctly blends with todo status
**Depends on**: Nothing (independent cleanup)
**Requirements**: CLN-01, CLN-02, CLN-03, CAL-01
**Success Criteria** (what must be TRUE):
  1. JSON store implementation files are deleted and no code references them
  2. Old `A` (dated add) and `t` (template use) key bindings are removed from the codebase
  3. Today's calendar date shows pending (yellow) or done (green) coloring blended with the today highlight, not just the today style alone
  4. PROJECT.md validated requirements section includes all v1.6+ features (unified edit mode, preview on all items, indicator colors, full-pane template modes)
**Plans**: 2 plans

Plans:
- [x] 23-01-PLAN.md -- Remove JSON store, blend today indicator with todo status
- [x] 23-02-PLAN.md -- Remove obsolete A and t keybindings and all dead code

### Phase 24: Unified Add Form
**Goal**: User creates any todo (floating, dated, or with body) through a single full-pane form
**Depends on**: Phase 23 (old keybindings removed first)
**Requirements**: ADD-01, ADD-02, ADD-05, ADD-06, ADD-07
**Success Criteria** (what must be TRUE):
  1. Pressing `a` in normal mode opens a full-pane form with Title, Date, Body, and Template fields
  2. User can Tab between fields (Title -> Date -> Body -> Template -> Title)
  3. Pressing Enter from Title or Date field saves the todo; pressing Ctrl+D from Body or Template field saves the todo
  4. Leaving Date empty creates a floating (undated) todo; filling Date creates a dated todo
  5. The old `A` and `t` key bindings no longer exist (removed in Phase 23)
**Plans**: TBD

Plans:
- [ ] 24-01: TBD
- [ ] 24-02: TBD

### Phase 25: Template Picker Integration
**Goal**: User can select a template from within the add form to pre-fill title and body, then edit before saving
**Depends on**: Phase 24 (add form exists)
**Requirements**: ADD-03, ADD-04
**Success Criteria** (what must be TRUE):
  1. Tabbing to the Template field and pressing Enter opens the template picker list
  2. Selecting a template pre-fills the Title field with the template name and the Body field with the rendered template content
  3. After template selection, user can navigate to Title or Body fields and edit the pre-filled content before saving
**Plans**: TBD

Plans:
- [ ] 25-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 23 -> 24 -> 25

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 23. Cleanup & Calendar Polish | v1.7 | 2/2 | ✓ Complete | 2026-02-07 |
| 24. Unified Add Form | v1.7 | 0/TBD | Not started | - |
| 25. Template Picker Integration | v1.7 | 0/TBD | Not started | - |
