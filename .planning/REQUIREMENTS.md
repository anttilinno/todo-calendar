# Requirements: Todo Calendar

**Defined:** 2026-02-07
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen.

## v1.5 Requirements

Requirements for UX Polish milestone. Each maps to roadmap phases.

### Visual Overhaul

- [x] **VIS-01**: Todo items have increased vertical spacing between them for easier scanning
- [x] **VIS-02**: Section headers (month name, "Floating") are visually distinct from items (separator line, padding, or stronger styling)
- [x] **VIS-03**: Dates and status indicators have clear visual differentiation from todo text (color, alignment, or positioning)

### Full-Pane Editing

- [x] **EDIT-01**: Adding a todo (title input) takes over the full right pane with centered/prominent input field
- [x] **EDIT-02**: Adding a dated todo (title + date input) uses the full right pane with clear field labels
- [x] **EDIT-03**: Editing a todo title uses the full right pane layout
- [x] **EDIT-04**: Editing a todo date uses the full right pane layout
- [x] **EDIT-05**: Full-pane edit mode shows only minimal help (Enter/Esc/Tab)

### Mode-Aware Help Bar

- [x] **HELP-01**: Normal mode shows max 5 most-used keys (e.g., a/add, x/done, d/delete, e/edit, ?/more)
- [x] **HELP-02**: Input modes show only Enter/Esc
- [x] **HELP-03**: Expanded help (via ?) shows full keybinding list

### Pre-Built Templates

- [ ] **TMPL-01**: App ships with 3-4 general templates (e.g., meeting notes, checklist, daily plan)
- [ ] **TMPL-02**: App ships with 3-4 dev templates (e.g., bug report, feature spec, PR checklist)
- [ ] **TMPL-03**: Pre-built templates are available on first launch (seeded into DB)
- [ ] **TMPL-04**: User can delete pre-built templates (not forced)

## v2 Candidates

- Simple recurring todos

## Out of Scope

| Feature | Reason |
|---------|--------|
| Inline body preview in list | Complexity of multi-line height calculation, defer to future |
| Built-in markdown editor | External $EDITOR is the correct boundary |
| Custom keybinding configuration | Keep keybindings hardcoded for simplicity |
| Animation/transitions | Terminal TUI, keep it snappy |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| VIS-01 | Phase 17 | Complete |
| VIS-02 | Phase 17 | Complete |
| VIS-03 | Phase 17 | Complete |
| EDIT-01 | Phase 18 | Complete |
| EDIT-02 | Phase 18 | Complete |
| EDIT-03 | Phase 18 | Complete |
| EDIT-04 | Phase 18 | Complete |
| EDIT-05 | Phase 18 | Complete |
| HELP-01 | Phase 17 | Complete |
| HELP-02 | Phase 17 | Complete |
| HELP-03 | Phase 17 | Complete |
| TMPL-01 | Phase 19 | Pending |
| TMPL-02 | Phase 19 | Pending |
| TMPL-03 | Phase 19 | Pending |
| TMPL-04 | Phase 19 | Pending |

**Coverage:**
- v1.5 requirements: 15 total
- Mapped to phases: 15
- Unmapped: 0

---
*Requirements defined: 2026-02-07*
*Last updated: 2026-02-07 after Phase 18 completion*
