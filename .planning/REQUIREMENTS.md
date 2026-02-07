# Requirements: Todo Calendar

**Defined:** 2026-02-07
**Core Value:** See your month at a glance — calendar with holidays and todos in one terminal screen.

## v1.5 Requirements

Requirements for UX Polish milestone. Each maps to roadmap phases.

### Visual Overhaul

- [ ] **VIS-01**: Todo items have increased vertical spacing between them for easier scanning
- [ ] **VIS-02**: Section headers (month name, "Floating") are visually distinct from items (separator line, padding, or stronger styling)
- [ ] **VIS-03**: Dates and status indicators have clear visual differentiation from todo text (color, alignment, or positioning)

### Full-Pane Editing

- [ ] **EDIT-01**: Adding a todo (title input) takes over the full right pane with centered/prominent input field
- [ ] **EDIT-02**: Adding a dated todo (title + date input) uses the full right pane with clear field labels
- [ ] **EDIT-03**: Editing a todo title uses the full right pane layout
- [ ] **EDIT-04**: Editing a todo date uses the full right pane layout
- [ ] **EDIT-05**: Full-pane edit mode shows only minimal help (Enter/Esc/Tab)

### Mode-Aware Help Bar

- [ ] **HELP-01**: Normal mode shows max 5 most-used keys (e.g., a/add, x/done, d/delete, e/edit, ?/more)
- [ ] **HELP-02**: Input modes show only Enter/Esc
- [ ] **HELP-03**: Expanded help (via ?) shows full keybinding list

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
| VIS-01 | — | Pending |
| VIS-02 | — | Pending |
| VIS-03 | — | Pending |
| EDIT-01 | — | Pending |
| EDIT-02 | — | Pending |
| EDIT-03 | — | Pending |
| EDIT-04 | — | Pending |
| EDIT-05 | — | Pending |
| HELP-01 | — | Pending |
| HELP-02 | — | Pending |
| HELP-03 | — | Pending |
| TMPL-01 | — | Pending |
| TMPL-02 | — | Pending |
| TMPL-03 | — | Pending |
| TMPL-04 | — | Pending |

**Coverage:**
- v1.5 requirements: 15 total
- Mapped to phases: 0
- Unmapped: 15

---
*Requirements defined: 2026-02-07*
*Last updated: 2026-02-07 after initial definition*
