# Requirements: Todo Calendar

**Defined:** 2026-02-06
**Core Value:** See your month at a glance -- calendar with holidays and todos in one terminal screen.

## v1.4 Requirements

Requirements for milestone v1.4 Data & Editing.

### Database Backend

- [x] **DB-01**: Todos are stored in a SQLite database instead of JSON files
- [x] **DB-02**: Store interface extracted to decouple consumers from backend
- [x] **DB-03**: Database schema versioned via PRAGMA user_version with hand-written migrations
- [x] **DB-04**: Type-safe database queries via hand-written SQL with scan helpers
- [x] **DB-05**: All existing CRUD operations work identically with the new backend

### Markdown Templates

- [ ] **MDTPL-01**: Todos have a markdown body field beyond the single-line title
- [ ] **MDTPL-02**: User can create reusable markdown templates with {{.Variable}} placeholders
- [ ] **MDTPL-03**: Creating a todo from a template fills in placeholders interactively
- [ ] **MDTPL-04**: Todo body renders as styled markdown (via Glamour) in a preview

### External Editor

- [ ] **EDITOR-01**: User can press a key to open the selected todo body in $EDITOR
- [ ] **EDITOR-02**: App uses $VISUAL -> $EDITOR -> vi fallback chain
- [ ] **EDITOR-03**: Temp file uses .md extension for editor syntax highlighting
- [ ] **EDITOR-04**: Body is only saved if content actually changed

## Future Requirements

Deferred to later milestones.

### Recurring Todos

- **RECUR-01**: User can mark a todo as recurring (daily/weekly/monthly)
- **RECUR-02**: Completed recurring todos auto-generate the next occurrence

## Out of Scope

| Feature | Reason |
|---------|--------|
| JSON-to-SQLite migration | No existing data to migrate |
| ORM (GORM, Ent, etc.) | Hand-written SQL is clearer for single-table app |
| sqlc | Overkill for simple CRUD on one table |
| dbmate / goose / migrate | PRAGMA user_version sufficient for single-user desktop app |
| Built-in markdown editor | External editor is the correct solution |
| YAML frontmatter in templates | Unnecessary complexity |
| FTS5 full-text search | Defer to v1.5 if needed; LIKE sufficient for now |
| Embedded text editor widget | External $EDITOR handles all editing |
| Day selection / day-by-day navigation | Month/week-level navigation is sufficient |
| Syncing / cloud storage | Local file only |
| Priority levels or tags | Keep it minimal |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| DB-01 | Phase 14 | Complete |
| DB-02 | Phase 14 | Complete |
| DB-03 | Phase 14 | Complete |
| DB-04 | Phase 14 | Complete |
| DB-05 | Phase 14 | Complete |
| MDTPL-01 | Phase 15 | Pending |
| MDTPL-02 | Phase 15 | Pending |
| MDTPL-03 | Phase 15 | Pending |
| MDTPL-04 | Phase 15 | Pending |
| EDITOR-01 | Phase 16 | Pending |
| EDITOR-02 | Phase 16 | Pending |
| EDITOR-03 | Phase 16 | Pending |
| EDITOR-04 | Phase 16 | Pending |

**Coverage:**
- v1.4 requirements: 13 total
- Mapped to phases: 13
- Unmapped: 0

---
*Requirements defined: 2026-02-06*
*Last updated: 2026-02-06 after roadmap creation*
