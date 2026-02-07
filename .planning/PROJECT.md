# Todo Calendar

## What This Is

A terminal-based (TUI) application that combines a monthly/weekly calendar view with a todo list. The left panel shows a navigable calendar with national holidays, date indicators for pending work, color-coded overview counts, and weekly view toggle. The right panel displays todos for the visible month alongside undated (floating) items with inline filter support. Includes full-screen cross-month search, editing and reordering todos, configurable date format and first day of week, 4 color themes, and an in-app settings overlay with live preview. Todos are stored in SQLite with support for rich markdown bodies, reusable templates with placeholder prompting, and external editor integration ($EDITOR). Templates can be managed in a dedicated overlay and have recurring schedules (daily, weekdays, weekly, monthly) that auto-create todos on app launch. Built with Go and Bubble Tea for personal use.

## Core Value

See your month at a glance — calendar with holidays and todos in one terminal screen.

## Requirements

### Validated

- Monthly calendar view (left panel) resembling `cal` output — v1.0
- Navigate between months (next/prev) — v1.0
- National holidays highlighted in red on calendar — v1.0
- Configurable country for holidays (11 countries supported) — v1.0
- Todo list (right panel) showing month's date-bound todos + floating todos — v1.0
- Add todo with optional date — v1.0
- Check off (complete) a todo — v1.0
- Delete a todo — v1.0
- Persist todos to a local JSON file on disk — v1.0
- Split-pane TUI layout (calendar left, todos right) — v1.0
- Keyboard navigation with vim keys and Tab focus switching — v1.0
- Context-sensitive help bar — v1.0
- Responsive terminal resize handling — v1.0
- TOML configuration file — v1.0
- XDG-compliant data paths — v1.0
- Calendar dates with incomplete todos display bracket indicators `[N]` — v1.1
- Edit todo text and date after creation — v1.1
- Configurable first day of week (Monday/Sunday) — v1.1
- 4 preset color themes (Dark, Light, Nord, Solarized) selectable in config — v1.1
- Reorder todos (move up/down with keybindings, persist order) — v1.2
- In-app settings overlay (full-screen, live theme preview, save/cancel) — v1.2
- Settings: change theme, country, first day of week with immediate feedback — v1.2
- Overview panel below calendar showing todo counts per month + undated count — v1.2
- Overview color coding: pending (themed red) and completed (themed green) counts — v1.3
- Date format setting: 3 presets (ISO/EU/US) configurable in settings with format-aware input — v1.3
- Weekly calendar view with `w` toggle, week navigation, and auto-select current week — v1.3
- Inline todo filter (`/`) with real-time case-insensitive narrowing — v1.3
- Full-screen search overlay (Ctrl+F) for cross-month todo discovery with jump-to-month — v1.3
- SQLite database backend replacing JSON storage — v1.4
- TodoStore interface decoupling consumers from storage backend — v1.4
- Markdown todo bodies with glamour-rendered preview overlay — v1.4
- Reusable markdown templates with {{.Variable}} placeholder prompting — v1.4
- External editor integration ($VISUAL/$EDITOR/vi fallback) — v1.4
- Todo pane visual overhaul with styled checkboxes, section separators, and vertical spacing — v1.5
- Full-pane editing for add/edit todo with dual-field dated-add and Tab switching — v1.5
- Mode-aware help bar showing 5 keys in normal mode, full list via ? toggle — v1.5
- 7 pre-built markdown templates (3 general + 4 dev) seeded on first launch — v1.5
- Template management overlay with CRUD (list, preview, rename, delete, edit) — v1.6
- Recurring schedules on templates (daily, weekdays, weekly, monthly) with schedule picker UI — v1.6
- Auto-creation of scheduled todos on app launch (rolling 7-day window with dedup) — v1.6
- Placeholder defaults prompting at schedule creation for auto-created todos — v1.6
- [R] indicator on recurring todos and schedule cadence suffix in template overlay — v1.6
- Unified edit mode with title + date + body fields (Tab to cycle) — v1.6+
- Preview works on all items including those without bodies — v1.6+
- Distinct pending (yellow) vs completed (green) calendar indicator colors in Nord and Solarized — v1.6+
- Template and placeholder modes render as full-pane views — v1.6+

### Active

- [ ] Unified add flow — single `a` key opens full-pane form (title, date, body, template picker) replacing separate `a`/`A`/`t` entry points
- [ ] Today calendar indicator blends with pending/done colors — currently today style overrides indicator status
- [ ] Remove dead code — unused JSON store, old keybindings
- [ ] Document recent v1.6+ changes — unified edit mode, preview on all items, indicator color fix, full-pane template modes

## Current Milestone: v1.7 Unified Add Flow & Polish

**Goal:** Unify the three separate todo creation flows (quick add, dated add, template add) into a single full-pane form, plus polish and cleanup.

**Target features:**
- Single add form with title, date (optional), body (optional), template picker (optional)
- Template selection pre-fills title/body, user can edit before saving
- Today calendar indicator shows pending/done status alongside today highlight
- Remove unused JSON store and old keybindings
- Update validated requirements for recent commits

### v2 Candidates

- Complex recurring cadences ("every 2nd Tuesday", "last Friday of month")

### Out of Scope

- Individual day selection / day-by-day arrow navigation — month-level navigation is sufficient
- Syncing / cloud storage — local file only
- Priority levels or tags — keep it minimal
- CalDAV integration — complexity explosion
- Subtasks / nesting — flat list is sufficient
- Notifications / reminders — out of scope for TUI
- Time-blocked appointments — this is a todo app, not a scheduler

## Context

- **Stack:** Go 1.25.6, Bubble Tea v1.3.10, Lipgloss v1.1.0, Bubbles v0.21.1, Glamour v0.10.0
- **Holidays:** rickar/cal/v2 with 11-country registry (de, dk, ee, es, fi, fr, gb, it, no, se, us)
- **Config:** TOML at ~/.config/todo-calendar/config.toml (BurntSushi/toml v1.6.0)
- **Storage:** SQLite at ~/.config/todo-calendar/todos.db (modernc.org/sqlite, pure Go, WAL mode)
- **Codebase:** 7,624 lines of Go across 35 source files
- **Architecture:** Elm Architecture (Bubble Tea), pure rendering functions, constructor DI, TodoStore interface

## Constraints

- **Stack**: Go + Bubble Tea — chosen for ergonomic component model and ecosystem
- **Storage**: Local SQLite only — no network dependencies, no cloud sync
- **Holidays**: Must work offline using bundled Go library, not an external API

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go + Bubble Tea over Rust + Ratatui | Better component model for split-pane layout, gentler learning curve | ✓ Good — clean architecture, fast development |
| Month-level navigation, no day selection | User doesn't have many items — showing all month todos is simpler | ✓ Good — keeps UI simple |
| Local JSON file → SQLite in v1.4 | JSON was simpler initially; SQLite needed for body/templates | ✓ Good — migrated in v1.4 via TodoStore interface |
| Configurable country holidays via Go library | Offline, no API dependency, flexible | ✓ Good — 11 countries supported |
| String dates (YYYY-MM-DD) over time.Time | Prevents timezone corruption during JSON round-trips | ✓ Good — research-informed decision |
| Atomic file writes (CreateTemp+Sync+Rename) | Data safety from day one | ✓ Good — prevents corruption |
| Pure rendering functions | RenderGrid has no side effects, testable | ✓ Good — clean separation |
| Three-mode input state machine | Cleanly separates key handling for normal/text/date | ✓ Good — input isolation works well |
| `first_day_of_week` string over `monday_start` bool | More extensible, clearer semantics | ✓ Good — clean config field |
| 4-char calendar cells (grid 34 chars wide) | Room for bracket indicators `[N]` without breaking alignment | ✓ Good — fits well |
| Semantic theme color roles (14 fields) | Named by role not component, decoupled from UI structure | ✓ Good — clean theme propagation |
| Styles struct + constructor DI over package-level vars | Enables runtime theme switching, testable | ✓ Good — no global state |
| Empty string = terminal default in Dark theme | Respects user's terminal palette | ✓ Good — non-invasive default |
| Settings as full-screen overlay with live preview | User wants to see changes immediately; overlay avoids cramming into split pane | ✓ Good — clean UX with save/cancel |
| SortOrder field with gap-10 spacing | Efficient reordering without renumbering all items | ✓ Good — simple swap-based reorder |
| No caching of overview data; fresh from store | Tiny dataset, no cache invalidation complexity | ✓ Good — always correct |
| Dedicated PendingFg/CompletedCountFg theme roles | Avoid coupling unrelated UI elements by reusing colors | ✓ Good — clean separation |
| FormatDate/ParseUserDate in config package | Co-located with DateLayout/DatePlaceholder, single bridge between ISO storage and display | ✓ Good — clean abstraction |
| ViewMode enum with weekStart tracking | year/month auto-sync from weekStart enables seamless todolist integration | ✓ Good — zero changes to todolist |
| Keys() returns mode-aware copies | Avoids mutating stored key bindings; clean contextual help | ✓ Good — no side effects |
| Search overlay creates fresh model on Ctrl+F | No stale state; simple initialization | ✓ Good — clean lifecycle |
| Inline filter preserves section headers | Headers always visible with "(no matches)" placeholder for empty sections | ✓ Good — clear UX |
| modernc.org/sqlite over mattn/go-sqlite3 | Pure Go, no CGo required, simpler cross-compilation | ✓ Good — zero build complexity |
| TodoStore interface in store package | Decouple consumers from backend, enable future backends | ✓ Good — clean DI, all 5 consumers updated |
| PRAGMA user_version for schema migration | Lightweight, no external tool, sufficient for single-user app | ✓ Good — v1→v2 migration seamless |
| Hand-written SQL over sqlc/ORM | Single-table CRUD, scan helpers sufficient | ✓ Good — clear and debuggable |
| text/template/parse AST walk for placeholders | Correct handling of all node types vs fragile regex | ✓ Good — handles If/Range/With correctly |
| Glamour for markdown rendering | Charmbracelet ecosystem, theme-matched light/dark styles | ✓ Good — clean terminal markdown |
| editing bool flag + View() empty guard | Prevents Bubble Tea alt-screen teardown leak to scrollback | ✓ Good — clean editor lifecycle |
| $VISUAL → $EDITOR → vi fallback | POSIX standard editor resolution chain | ✓ Good — works on all Unix systems |
| Styled checkboxes independent from text | Accent [ ], green [x], strikethrough only on text | ✓ Good — clean visual separation |
| Mode-branched View() with editView()/normalView() | Clean separation of edit vs list rendering | ✓ Good — each mode fully owns the pane |
| SetSize(w,h) replacing WindowSizeMsg | Todolist gets dimensions from parent, not global messages | ✓ Good — cleaner ownership |
| Migration-based template seeding (PRAGMA user_version) | Run-once, idempotent, no runtime checks | ✓ Good — follows existing migration pattern |
| Raw text preview in template overlay (not glamour) | Reveals placeholder syntax {{.Variable}} | ✓ Good — users see what they'll edit |
| UpdateTemplate returns error for UNIQUE constraint | Rename UI shows error for duplicate names | ✓ Good — clean error handling |
| tmplmgr overlay as separate package | Follows search/settings/preview pattern | ✓ Good — consistent architecture |
| CadenceType/CadenceValue flexible string columns | Extensible without schema changes | ✓ Good — simple and clean |
| PlaceholderDefaults as JSON string in schedules | Arbitrary key-value pairs per schedule | ✓ Good — flexible storage |
| FK CASCADE templates→schedules, SET NULL schedules→todos | Template delete cleans schedules; schedule delete preserves todos | ✓ Good — correct cascade semantics |
| UNIQUE index on (schedule_id, schedule_date) | Database-level deduplication for auto-created todos | ✓ Good — prevents duplicates |
| AutoCreate runs synchronously before TUI | Simple, no goroutine complexity | ✓ Good — fast for small schedule counts |
| Placeholder defaults prompting intercepts schedule confirm | Single flow: pick cadence → fill defaults → save | ✓ Good — natural UX progression |

## Known Tech Debt

- JSON Store still exists but unused (main.go uses SQLiteStore exclusively) — targeted for removal in v1.7
- JSON Store template and schedule methods are stubs (return error/nil/no-op)

---
*Last updated: 2026-02-07 after v1.7 milestone start*
