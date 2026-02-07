# Phase 19: Pre-Built Templates - Research

**Researched:** 2026-02-07
**Domain:** SQLite seeding, Go template content, existing template infrastructure
**Confidence:** HIGH

## Summary

This phase adds 6-8 pre-built markdown templates that are available on first launch. The existing template infrastructure (Phase 15-16) is complete and fully functional: templates are stored in SQLite with `AddTemplate`, listed with `ListTemplates`, and deleted with `DeleteTemplate`. The only new work is (1) defining the template content, (2) adding a seeding mechanism that runs once on first launch, and (3) writing tests.

The recommended approach is a new migration step (version 3) in `sqlite.go` that inserts the seed templates using `INSERT OR IGNORE` to handle the UNIQUE name constraint idempotently. This is the simplest, most reliable mechanism and follows the existing pattern. No new packages, dependencies, or interfaces are needed.

**Primary recommendation:** Add a version-3 migration that seeds 7 templates (3 general + 4 dev) into the existing templates table, using `INSERT OR IGNORE` for idempotency.

## Standard Stack

No new libraries needed. This phase uses only existing infrastructure:

### Core (existing)
| Component | Location | Purpose | Status |
|-----------|----------|---------|--------|
| SQLiteStore | `internal/store/sqlite.go` | Template CRUD (Add/List/Find/Delete) | Complete |
| Template struct | `internal/store/todo.go` | ID, Name, Content, CreatedAt | Complete |
| tmpl package | `internal/tmpl/tmpl.go` | ExtractPlaceholders, ExecuteTemplate | Complete |
| todolist model | `internal/todolist/model.go` | Template selection/creation/deletion UI | Complete |

### No New Dependencies
This phase requires zero new Go modules. The seed data is just Go string constants inserted via existing `database/sql` calls.

## Architecture Patterns

### Pattern 1: Migration-Based Seeding (RECOMMENDED)

**What:** Add a version-3 migration in the existing `migrate()` function that inserts seed templates.
**When to use:** Always -- this is the only seeding mechanism needed.
**Why:** Follows the established pattern (version 1 = todos table, version 2 = templates table, version 3 = seed templates). Runs exactly once. Works for both new installs and upgrades from existing DBs.

```go
// In sqlite.go migrate() function, after the version < 2 block:
if version < 3 {
    for _, t := range defaultTemplates() {
        s.db.Exec(
            "INSERT OR IGNORE INTO templates (name, content, created_at) VALUES (?, ?, ?)",
            t.Name, t.Content, time.Now().Format(dateFormat),
        )
    }
    if _, err := s.db.Exec("PRAGMA user_version = 3"); err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

**Key details:**
- `INSERT OR IGNORE` handles the UNIQUE constraint on `name` gracefully -- if a user already created a template with the same name, it silently skips
- The `PRAGMA user_version` bump ensures this runs exactly once per database
- `defaultTemplates()` is a function (not package-level var) so it can be defined in a separate file for cleanliness

### Pattern 2: Seed Data in a Separate File

**What:** Define template content in a dedicated `internal/store/seed.go` file.
**When to use:** Always -- keeps sqlite.go clean and seed data maintainable.
**Why:** The template content strings are multi-line markdown. Putting 7 templates inline in sqlite.go would make the file unwieldy. A separate file keeps concerns separated.

```go
// internal/store/seed.go
package store

// defaultTemplate holds name and content for a seed template.
type defaultTemplate struct {
    Name    string
    Content string
}

// defaultTemplates returns the pre-built templates seeded on first launch.
func defaultTemplates() []defaultTemplate {
    return []defaultTemplate{
        {Name: "Meeting Notes", Content: meetingNotesContent},
        // ...
    }
}

const meetingNotesContent = `## Meeting: {{.Topic}}
...`
```

### Anti-Patterns to Avoid

- **Seeding in main.go or app.go:** The store owns its data; seeding belongs in the migration layer, not the application startup logic.
- **Checking template count at startup:** Fragile (user could delete all templates, then on next launch they'd be re-created). Migration version check is the correct mechanism.
- **Separate seeder function called from main:** Adds unnecessary complexity. Migration is already called once at startup.
- **Marking templates as "built-in" with a flag column:** The requirements explicitly say users can delete any template. No special treatment needed. Adding a column would be over-engineering.
- **Using `embed` for template files:** The templates are short (5-15 lines each). Go string constants are simpler and require no build changes.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Idempotent seeding | Custom "has_been_seeded" flag | `PRAGMA user_version` migration | Already exists, battle-tested in this codebase |
| Duplicate prevention | Check-then-insert logic | `INSERT OR IGNORE` | SQLite handles atomically, no race conditions |
| Template storage | New table or file format | Existing `templates` table | Phase 15 already built this |

## Common Pitfalls

### Pitfall 1: Re-seeding After User Deletion
**What goes wrong:** If seeding checks "are there templates?" instead of using migration version, deleted templates come back on restart.
**Why it happens:** Confusing "first launch" with "no templates exist."
**How to avoid:** Use `PRAGMA user_version` check (version < 3). Runs exactly once, regardless of what user does after.
**Warning signs:** Templates reappearing after user deletes them.

### Pitfall 2: Template Name Collision
**What goes wrong:** Seed template name matches a user-created template name, causing INSERT failure.
**Why it happens:** The `name` column has a UNIQUE constraint.
**How to avoid:** Use `INSERT OR IGNORE` which silently skips on UNIQUE constraint violation. The user's existing template is preserved.

### Pitfall 3: Templates Too Long for TUI Display
**What goes wrong:** Template content is so long that the template selection list becomes unwieldy or the resulting todo body doesn't fit well.
**Why it happens:** Designing templates in isolation without considering the terminal viewport.
**How to avoid:** Keep templates concise (5-15 lines). Focus on structure and placeholders, not boilerplate text. The user fills in details.

### Pitfall 4: Overly Complex Placeholders
**What goes wrong:** Templates with 5+ placeholders become tedious to fill out one at a time.
**Why it happens:** The current UI prompts for each placeholder sequentially via single text input.
**How to avoid:** Limit to 1-3 placeholders per template. Use static structure for the rest.

## Code Examples

### Template Content Design Principles

From the existing `tmpl` package, templates use Go's `text/template` syntax:
- Placeholders: `{{.VariableName}}` (PascalCase by convention)
- Missing values render as empty string (missingkey=zero option)
- Placeholders are extracted and prompted in order of first appearance

### General Templates (3 recommended)

```go
// Meeting Notes - 2 placeholders
const meetingNotesContent = `## {{.Topic}}

**Date:** {{.Date}}
**Attendees:**

### Agenda
-

### Notes
-

### Action Items
- [ ] `

// Checklist - 1 placeholder
const checklistContent = `## {{.Title}}

- [ ]
- [ ]
- [ ]
- [ ]
- [ ] `

// Daily Plan - 0 placeholders (pure structure)
const dailyPlanContent = `## Daily Plan

### Top Priorities
1.
2.
3.

### Tasks
- [ ]
- [ ]
- [ ]

### Notes
`
```

### Dev Templates (4 recommended)

```go
// Bug Report - 2 placeholders
const bugReportContent = `## Bug: {{.Summary}}

**Component:** {{.Component}}

### Steps to Reproduce
1.
2.
3.

### Expected Behavior


### Actual Behavior


### Environment
- OS:
- Version: `

// Feature Spec - 1 placeholder
const featureSpecContent = `## Feature: {{.Name}}

### Problem
What problem does this solve?

### Solution
How should it work?

### Acceptance Criteria
- [ ]
- [ ]
- [ ]

### Notes
`

// PR Checklist - 1 placeholder
const prChecklistContent = `## PR: {{.Title}}

### Changes
-

### Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes
- [ ] Self-reviewed code

### Testing
How was this tested?
`

// Code Review - 2 placeholders
const codeReviewContent = `## Review: {{.PR}}

**Author:** {{.Author}}

### Summary


### Feedback
-

### Approval
- [ ] Code quality
- [ ] Test coverage
- [ ] Documentation
`
```

### Migration Implementation

```go
// In sqlite.go, inside migrate() after version < 2 block:

if version < 3 {
    seeds := defaultTemplates()
    for _, t := range seeds {
        s.db.Exec(
            "INSERT OR IGNORE INTO templates (name, content, created_at) VALUES (?, ?, ?)",
            t.Name, t.Content, time.Now().Format(dateFormat),
        )
    }
    if _, err := s.db.Exec("PRAGMA user_version = 3"); err != nil {
        return fmt.Errorf("set user_version: %w", err)
    }
}
```

### Test Pattern

```go
// In internal/store/sqlite_test.go (new file)
func TestSeedTemplates(t *testing.T) {
    // Create temp DB
    dbPath := filepath.Join(t.TempDir(), "test.db")
    s, err := NewSQLiteStore(dbPath)
    if err != nil {
        t.Fatalf("create store: %v", err)
    }
    defer s.Close()

    // Verify templates were seeded
    templates := s.ListTemplates()
    if len(templates) < 6 || len(templates) > 8 {
        t.Errorf("expected 6-8 templates, got %d", len(templates))
    }

    // Verify each template has content with valid placeholder syntax
    for _, tmpl := range templates {
        if tmpl.Name == "" {
            t.Error("template has empty name")
        }
        if tmpl.Content == "" {
            t.Errorf("template %q has empty content", tmpl.Name)
        }
        // Verify placeholders can be extracted without error
        _, err := tmplpkg.ExtractPlaceholders(tmpl.Content)
        if err != nil {
            t.Errorf("template %q has invalid placeholder syntax: %v", tmpl.Name, err)
        }
    }
}

func TestSeedTemplates_Idempotent(t *testing.T) {
    dbPath := filepath.Join(t.TempDir(), "test.db")

    // First open: seeds templates
    s1, _ := NewSQLiteStore(dbPath)
    count1 := len(s1.ListTemplates())
    s1.Close()

    // Second open: should not duplicate
    s2, _ := NewSQLiteStore(dbPath)
    count2 := len(s2.ListTemplates())
    s2.Close()

    if count1 != count2 {
        t.Errorf("template count changed: %d -> %d", count1, count2)
    }
}

func TestSeedTemplates_DeletionPermanent(t *testing.T) {
    dbPath := filepath.Join(t.TempDir(), "test.db")

    s1, _ := NewSQLiteStore(dbPath)
    templates := s1.ListTemplates()
    for _, tmpl := range templates {
        s1.DeleteTemplate(tmpl.ID)
    }
    if len(s1.ListTemplates()) != 0 {
        t.Error("templates not deleted")
    }
    s1.Close()

    // Reopen: templates should NOT come back
    s2, _ := NewSQLiteStore(dbPath)
    if len(s2.ListTemplates()) != 0 {
        t.Error("deleted templates reappeared after reopen")
    }
    s2.Close()
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| JSON store (no template support) | SQLite with templates table | Phase 14-15 (v1.4) | Templates only work with SQLite backend |

**Key constraint:** The JSON `Store` implementation returns empty/nil for all template methods. Pre-built templates only apply to `SQLiteStore`, which is the only backend used in `main.go`. This is fine -- no action needed.

## Open Questions

1. **Exact template count: 7 vs 8?**
   - What we know: Requirements say "6-8 templates" with "3-4 general" and "3-4 dev"
   - Recommendation: 3 general + 4 dev = 7 total. This satisfies the range and provides good variety without bloat.

2. **Template naming convention**
   - What we know: Templates are listed alphabetically (`ORDER BY name`). Names appear in the selection UI.
   - Recommendation: Use descriptive, user-friendly names like "Meeting Notes", "Bug Report", "PR Checklist" -- no prefixes or categories in the name. The mix of general and dev templates will naturally sort well alphabetically.

## Sources

### Primary (HIGH confidence)
- `internal/store/sqlite.go` -- Migration system (PRAGMA user_version), template CRUD, UNIQUE constraint on name
- `internal/store/todo.go` -- Template struct definition (ID, Name, Content, CreatedAt)
- `internal/tmpl/tmpl.go` -- Placeholder extraction and execution using `text/template`
- `internal/todolist/model.go` -- Template selection UI, deletion flow, placeholder prompting
- `main.go` -- App initialization: `NewSQLiteStore` -> `migrate()` runs automatically

### Secondary (MEDIUM confidence)
- SQLite `INSERT OR IGNORE` behavior with UNIQUE constraints -- well-documented standard SQLite feature

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all infrastructure already exists, verified by reading source code
- Architecture: HIGH -- migration pattern is established in this codebase (2 versions already)
- Pitfalls: HIGH -- identified from reading existing code constraints (UNIQUE, UI flow)
- Template content: MEDIUM -- content is subjective; recommended templates are practical but could be adjusted

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (stable domain, no external dependencies)
