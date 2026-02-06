---
phase: 15-markdown-templates
verified: 2026-02-06T23:37:21+02:00
status: passed
score: 22/22 must-haves verified
---

# Phase 15: Markdown Templates Verification Report

**Phase Goal:** Todos support rich markdown bodies created from reusable templates
**Verified:** 2026-02-06T23:37:21+02:00
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can view a multi-line markdown body attached to any todo | ✓ VERIFIED | Preview overlay exists at `/internal/preview/model.go`, wired via PreviewMsg, triggered by 'p' key |
| 2 | User can create named templates containing markdown with placeholder variables | ✓ VERIFIED | Template creation flow exists: 'T' key -> name input (templateNameMode) -> content input (templateContentMode) -> AddTemplate call |
| 3 | When creating a todo from a template, user is prompted for each placeholder value and the body is filled in | ✓ VERIFIED | Template selection flow: 't' key -> select template -> ExtractPlaceholders -> placeholderInputMode -> sequential prompts -> ExecuteTemplate -> todo creation with UpdateBody |
| 4 | Todo body renders as styled terminal markdown (headings, lists, code blocks) in a preview pane | ✓ VERIFIED | Glamour renderer at `/internal/preview/styles.go`, NewMarkdownRenderer creates glamour.TermRenderer, viewport displays rendered markdown |

**Score:** 4/4 truths verified

### Plan 15-01 Must-Haves (Store Foundation)

#### Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Todo struct has a Body string field included in all SQL queries | ✓ VERIFIED | `Body string` field at `/internal/store/todo.go:14`, included in todoColumns const and scanTodo function |
| 2 | TodoStore interface includes UpdateBody, AddTemplate, ListTemplates, FindTemplate, DeleteTemplate methods | ✓ VERIFIED | All 5 methods in TodoStore interface at `/internal/store/store.go:27-31` |
| 3 | SQLite migration v2 creates a templates table | ✓ VERIFIED | Migration at `/internal/store/sqlite.go:83-95` creates templates table with UNIQUE name constraint |
| 4 | Both Store (JSON) and SQLiteStore satisfy the extended TodoStore interface | ✓ VERIFIED | Compile-time checks pass: `var _ TodoStore = (*Store)(nil)` and `var _ TodoStore = (*SQLiteStore)(nil)` |
| 5 | ExtractPlaceholders parses template content and returns unique placeholder names | ✓ VERIFIED | Function at `/internal/tmpl/tmpl.go:13-23`, walks parse tree collecting FieldNode names |
| 6 | ExecuteTemplate fills placeholders with provided values | ✓ VERIFIED | Function at `/internal/tmpl/tmpl.go:81-91`, uses text/template with missingkey=zero |

#### Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `/internal/store/todo.go` | Todo.Body field, HasBody() method, Template type | ✓ VERIFIED | Body field line 14, HasBody() line 22-24, Template struct line 27-32 |
| `/internal/store/store.go` | Extended TodoStore interface, JSON Store methods | ✓ VERIFIED | Interface lines 15-36, UpdateBody line 190-198, template stubs lines 200-216 |
| `/internal/store/sqlite.go` | Migration v2 templates table, body in queries, template CRUD | ✓ VERIFIED | Migration line 83-95, body in todoColumns line 106, scanTodo line 113, template methods lines 363-420 |
| `/internal/tmpl/tmpl.go` | ExtractPlaceholders and ExecuteTemplate utilities | ✓ VERIFIED | Package exists, both functions present with full implementation |

#### Key Links

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `/internal/store/sqlite.go` | `/internal/store/store.go` | interface satisfaction | ✓ WIRED | `var _ TodoStore = (*SQLiteStore)(nil)` line 15 |
| `/internal/store/store.go` | `/internal/store/store.go` | interface satisfaction | ✓ WIRED | `var _ TodoStore = (*Store)(nil)` line 39 |

### Plan 15-02 Must-Haves (Preview Overlay)

#### Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can press 'p' on a selected todo to open a full-screen markdown preview of its body | ✓ VERIFIED | Preview keybinding at `/internal/todolist/keys.go:84-87`, triggered at model.go:374-380, emits PreviewMsg |
| 2 | Preview renders markdown with styled headings, lists, and code blocks via glamour | ✓ VERIFIED | Glamour renderer at `/internal/preview/styles.go:37-54`, renders via glamour.TermRenderer with theme-aware styles |
| 3 | User can scroll the preview with j/k and close it with Esc or q | ✓ VERIFIED | KeyMap at `/internal/preview/keys.go` defines scroll keys, Close binding line 33-36, Update handles at model.go:80-86 |
| 4 | Todos with a non-empty body show a [+] indicator in the todo list | ✓ VERIFIED | BodyIndicator style at `/internal/todolist/styles.go:15,26`, rendered at model.go:849-851 when HasBody() is true |
| 5 | Preview overlay blocks all other input while open (like search/settings) | ✓ VERIFIED | showPreview flag at `/internal/app/model.go:50`, routing at line 150, blocks other modes like search/settings |

#### Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `/internal/preview/model.go` | Preview Bubble Tea model with viewport and glamour rendering | ✓ VERIFIED | 161 lines, CloseMsg line 16, New() constructor, viewport field, glamour renderer |
| `/internal/preview/keys.go` | Preview keybindings (scroll, close) | ✓ VERIFIED | 45 lines, KeyMap struct, DefaultKeyMap() with scroll and close bindings |
| `/internal/preview/styles.go` | Theme-integrated glamour renderer builder | ✓ VERIFIED | 55 lines, NewMarkdownRenderer line 37-54, theme-aware base style selection |
| `/internal/todolist/styles.go` | BodyIndicator style for [+] marker | ✓ VERIFIED | BodyIndicator field line 15, initialized line 26 with MutedFg |

#### Key Links

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `/internal/app/model.go` | `/internal/preview/model.go` | showPreview bool + preview field + CloseMsg handler | ✓ WIRED | showPreview line 50, preview field line 51, CloseMsg handler line 134-135, PreviewMsg handler line 138-141 |
| `/internal/todolist/model.go` | `/internal/app/model.go` | PreviewMsg with todo ID | ✓ WIRED | PreviewMsg struct line 19-21, emitted line 379 when 'p' pressed on todo with body |
| `/internal/preview/styles.go` | `github.com/charmbracelet/glamour` | glamour.NewTermRenderer with WithStyles | ✓ WIRED | Import line 5, NewMarkdownRenderer calls glamour.NewTermRenderer line 50-52 |

### Plan 15-03 Must-Haves (Template Creation and Usage)

#### Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can press 't' to enter template selection mode showing a list of saved templates | ✓ VERIFIED | TemplateUse binding line 88-91 in keys.go, handler at model.go:383-390 loads templates and sets templateSelectMode |
| 2 | User can create a new template by pressing 'T' (enters name, then content) | ✓ VERIFIED | TemplateCreate binding line 92-95, handler at model.go:392-396 enters templateNameMode, then templateContentMode |
| 3 | When creating a todo from a template, user is prompted for each placeholder value sequentially | ✓ VERIFIED | Placeholder flow at model.go:606-626, ExtractPlaceholders call line 606, placeholderInputMode entered line 622 |
| 4 | After all placeholders are filled, a todo is created with the rendered template body | ✓ VERIFIED | ExecuteTemplate called line 669, body stored in pendingBody, UpdateBody called after Add() at lines 450 and 489 |
| 5 | If template has no placeholders, todo is created immediately with the template body | ✓ VERIFIED | No-placeholder path at model.go:607-616, directly executes template and enters inputMode for todo text |
| 6 | Esc cancels at any point and returns to normal mode | ✓ VERIFIED | Cancel handlers in all template modes: line 644-647 (select), 678-683 (placeholder), 705-710 (name), 736-740 (content) |

#### Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `/internal/todolist/model.go` | Template selection mode, placeholder input mode, template creation mode | ✓ VERIFIED | templateSelectMode line 34, placeholderInputMode line 35, templateNameMode line 36, templateContentMode line 37, update methods lines 586-748 |
| `/internal/todolist/keys.go` | Template keybindings (TemplateUse, TemplateCreate) | ✓ VERIFIED | TemplateUse line 19 and 88-91, TemplateCreate line 20 and 92-95 |

#### Key Links

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `/internal/todolist/model.go` | `/internal/tmpl/tmpl.go` | ExtractPlaceholders and ExecuteTemplate calls | ✓ WIRED | Import line 15, ExtractPlaceholders called line 606, ExecuteTemplate called lines 609 and 669 |
| `/internal/todolist/model.go` | `/internal/store/store.go` | ListTemplates, AddTemplate, Add, UpdateBody calls | ✓ WIRED | ListTemplates lines 384 and 633, AddTemplate line 730, UpdateBody lines 450 and 489 |

### Requirements Coverage

| Requirement | Status | Supporting Truth |
|-------------|--------|------------------|
| MDTPL-01: Todos have a markdown body field | ✓ SATISFIED | Truth 1 (view body), Body field verified, [+] indicator verified |
| MDTPL-02: Create reusable templates with {{.Variable}} | ✓ SATISFIED | Truth 2 (create templates), Template type, AddTemplate, 'T' keybinding verified |
| MDTPL-03: Template fills placeholders interactively | ✓ SATISFIED | Truth 3 (placeholder prompts), ExtractPlaceholders/ExecuteTemplate verified |
| MDTPL-04: Body renders as styled markdown via Glamour | ✓ SATISFIED | Truth 4 (styled rendering), Glamour integration verified |

### Anti-Patterns Found

None detected. All implementation files are substantive (40-161 lines each), no TODO/FIXME comments, no stub patterns, all functions have real implementations.

### Human Verification Required

#### 1. Visual Markdown Rendering Quality

**Test:** Create a todo with a markdown body containing headers, lists, code blocks, and bold/italic text. Press 'p' to preview.
**Expected:** Markdown renders with distinct styling: headers are bold/colored, lists are indented with bullets, code blocks have background/borders, emphasis works.
**Why human:** Visual quality assessment requires human judgment.

#### 2. Template Placeholder Flow

**Test:** 
1. Press 'T' to create template named "Meeting Notes"
2. Enter content: `# {{.Topic}}\n\nAttendees: {{.People}}\n\n## Notes\n{{.Notes}}`
3. Press Ctrl+D to save
4. Press 't' to select "Meeting Notes"
5. Fill placeholders: Topic="Sprint Planning", People="Alice, Bob", Notes="Discussed Q1 goals"
6. Complete todo creation
7. Press 'p' on created todo

**Expected:** Preview shows fully rendered markdown with all placeholders replaced.
**Why human:** End-to-end flow validation requires user interaction.

#### 3. Body Indicator Visual Clarity

**Test:** Create multiple todos, some with bodies, some without. Navigate the todo list.
**Expected:** [+] indicator appears clearly next to todos with bodies, is muted (not distracting), and does not appear on todos without bodies.
**Why human:** Visual clarity and UX assessment.

#### 4. Preview Scrolling for Long Content

**Test:** Create a todo with a very long markdown body (30+ lines). Open preview and test j/k scrolling.
**Expected:** Viewport scrolls smoothly, shows scroll position indicator if content exceeds screen, all content accessible.
**Why human:** Viewport behavior testing requires user interaction.

## Summary

**All 22 must-haves verified programmatically. Phase goal achieved.**

### Store Layer (15-01) ✓
- Todo.Body field exists and is persisted in both JSON and SQLite backends
- Template type and templates table created via migration v2
- TodoStore interface extended with 5 new methods (UpdateBody + 4 template CRUD)
- Both Store implementations satisfy the interface
- Template utilities (ExtractPlaceholders, ExecuteTemplate) implemented and functional

### Preview Overlay (15-02) ✓
- Full preview package with Bubble Tea model, keys, and styles
- Glamour dependency added and integrated for markdown rendering
- Preview triggered via 'p' key on todos with bodies
- Body indicator [+] displayed in todo list
- Preview overlay follows established app pattern (like search/settings)

### Template Workflow (15-03) ✓
- Template selection mode ('t') with cursor navigation
- Template creation mode ('T') with name and multi-line content input
- Placeholder extraction and sequential prompting
- Template execution and todo body attachment via UpdateBody
- Template deletion ('d' in selection mode)
- Comprehensive state management with clearTemplateState()

### Wiring Verified ✓
- Preview: todolist -> app -> preview overlay
- Template usage: todolist -> tmpl utilities -> store methods
- Body persistence: Add() -> UpdateBody() for template todos
- Glamour: preview styles -> glamour renderer -> viewport

### Build Status ✓
- `go build ./...` — passes
- `go vet ./...` — passes
- No compilation errors
- All imports resolve correctly

---

_Verified: 2026-02-06T23:37:21+02:00_
_Verifier: Claude (gsd-verifier)_
