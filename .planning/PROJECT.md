# Todo Calendar

## What This Is

A terminal-based (TUI) application that combines a monthly/weekly calendar view with a todo list. The left panel shows a navigable calendar with national holidays, priority-colored date indicators for pending work, Google Calendar event indicators, color-coded overview counts, blended today+status highlighting, circle indicators for month/year todo status, and weekly view toggle. The right panel displays todos in 4 sections (dated, This Month, This Year, Floating) with inline filter support, configurable section visibility, color-coded P1-P4 priority badges, and read-only Google Calendar events mixed into dated sections. Includes full-screen cross-month search with priority badges, editing and reordering todos, configurable date format and first day of week, 4 color themes, and an in-app settings overlay with live preview. Todos support day, month, or year date precision via a segmented date input, and optional P1-P4 priority levels set via an inline selector. Google Calendar events are fetched via OAuth 2.0 with background polling and displayed with visual distinction (teal color, time prefix, non-selectable). Stored in SQLite with support for rich markdown bodies, reusable templates with placeholder prompting, and external editor integration ($EDITOR). Templates can be managed in a dedicated overlay and have recurring schedules (daily, weekdays, weekly, monthly) that auto-create todos on app launch. A unified add form (`a` key) with title, date, priority, body, and template picker fields replaces the previous multi-key entry points. A `status` subcommand writes Polybar-formatted todo status to a state file, and the TUI keeps it updated on every mutation for real-time external status bar integration. Built with Go and Bubble Tea for personal use.

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
- Unified add form with title, date, body, and template picker fields (single `a` key) — v1.7
- Template picker in add form with j/k navigation, placeholder prompting, and Title/Body pre-fill — v1.7
- Today calendar indicator blends pending/done status with today highlight — v1.7
- Dead code removed: JSON store, old A/t keybindings, unused modes — v1.7
- In weekly view, todo panel filters to show only todos dated within the visible week — v1.8
- Floating (undated) todos remain visible in weekly view — v1.8
- Todo panel updates immediately when navigating weeks with h/l keys — v1.8
- Improved contrast for MutedFg/EmptyFg in Nord and Solarized themes — v1.8
- ✓ Month-level todos with segmented date input and dedicated "This Month" section — v1.9
- ✓ Year-level todos with dedicated "This Year" section — v1.9
- ✓ Calendar circle indicators: left (month) and right (year) status, red pending / green done — v1.9
- ✓ Show/hide toggles in settings for month and year todo sections with live preview — v1.9
- ✓ Segmented date input (dd/mm/yyyy) with format-aware ordering and precision derivation — v1.9
- ✓ Fuzzy todos visible in monthly view only, excluded from weekly view — v1.9
- ✓ Settings overlay: Esc saves and closes (no save button, no cancel flow) — v2.0
- ✓ Priority levels (P1-P4) with color-coded badge prefix in todo list, calendar, and search — v2.1
- ✓ Priority theme colors for all 4 themes (Dark, Light, Nord, Solarized) — v2.1
- ✓ Calendar day indicators reflect highest-priority incomplete todo's color — v2.1
- ✓ Inline priority selector in add/edit forms with left/right arrow cycling — v2.1
- ✓ Priority stored as INTEGER in SQLite with migration v7, existing todos default to 0 — v2.1
- ✓ OAuth 2.0 authentication with PKCE loopback redirect for Google Calendar — v2.2
- ✓ Token persistence with atomic writes (0600 permissions) and transparent auto-refresh — v2.2
- ✓ App works fully offline when Google account is not configured — v2.2
- ✓ Google Calendar events fetched via REST API with syncToken delta sync — v2.2
- ✓ Background polling re-fetches events every 5 minutes without TUI freeze — v2.2
- ✓ Events displayed in todo panel with HH:MM time prefix or "all day" label, teal color, non-selectable — v2.2
- ✓ Multi-day events expanded to show on each day they span — v2.2
- ✓ Calendar grid bracket indicators for days with Google Calendar events — v2.2
- ✓ Google Calendar enable/disable toggle in settings without removing credentials — v2.2
- ✓ `todo-calendar status` subcommand writes Polybar-formatted status to state file and exits — v2.3
- ✓ TUI updates state file on todo mutations (add, complete, delete, edit) — v2.3
- ✓ Polybar output format `%{F#hex}ICON COUNT%{F-}` colored by highest priority, empty string when zero pending — v2.3
- ✓ State file at `/tmp/.todo_status` initialized by subcommand and kept current by TUI — v2.3

### Active

(None — next milestone requirements TBD via `/gsd:new-milestone`)

### v2 Candidates

- Complex recurring cadences ("every 2nd Tuesday", "last Friday of month")
- Completed tasks archive (browse/review past completions)
- Natural language date input ("tomorrow", "next fri", "jan 15", "in 3 days")
- Inline priority cycling in normal mode (press 1-4 on selected todo)
- Default priority configurable in settings

### Out of Scope

- Individual day selection / day-by-day arrow navigation — month-level navigation is sufficient
- Syncing / cloud storage — local todos remain local-only (read-only Google Calendar pull is scoped exception)
- Tags / labels — keep it minimal (priorities are sufficient categorization)
- CalDAV write operations — read-only pull is sufficient; 2-way sync is complexity explosion
- Subtasks / nesting — flat list is sufficient
- Notifications / reminders — desktop notifications out of scope; passive Polybar status is the scoped exception (shipped v2.3)
- Time-blocked appointments — this is a todo app, not a scheduler
- Auto-sort by priority — conflicts with manual J/K reordering; priority is visual only
- Click actions in Polybar — status is read-only

## Context

- **Stack:** Go 1.25.6, Bubble Tea v1.3.10, Lipgloss v1.1.0, Bubbles v0.21.1, Glamour v0.10.0
- **Holidays:** rickar/cal/v2 with 11-country registry (de, dk, ee, es, fi, fr, gb, it, no, se, us)
- **Config:** TOML at ~/.config/todo-calendar/config.toml (BurntSushi/toml v1.6.0)
- **Storage:** SQLite at ~/.config/todo-calendar/todos.db (modernc.org/sqlite, pure Go, WAL mode)
- **Google Calendar:** OAuth 2.0 via golang.org/x/oauth2, REST API via google.golang.org/api/calendar/v3, token at ~/.config/todo-calendar/token.json
- **Codebase:** 9,823 lines of Go across 37 source files
- **Architecture:** Elm Architecture (Bubble Tea), pure rendering functions, constructor DI, TodoStore interface

## Constraints

- **Stack**: Go + Bubble Tea — chosen for ergonomic component model and ecosystem
- **Storage**: Local SQLite for todos — no cloud sync for user data; Google Calendar events cached in-memory only
- **Network**: Google Calendar pull is optional; app must work fully offline when unconfigured
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
| Settings as full-screen overlay with live preview | User wants to see changes immediately; overlay avoids cramming into split pane | ✓ Good — clean UX, save-on-close since v2.0 |
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
| Unified 4-field add form replacing 3 separate entry points | Single form is more intuitive in TUI context (research-backed) | ✓ Good — cleaner UX, fewer keybindings |
| Sub-state booleans (pickingTemplate) over new mode constants | Picker is transient sub-interaction of inputMode, not standalone | ✓ Good — clean mode enum, no mode switch proliferation |
| Shared m.input for title field and placeholder prompting | Avoids extra textinput; explicit state restore after pre-fill | ✓ Good — no duplication, state properly managed |
| Blended today+status styles (foreground=status, background=today) | Users see pending/done at a glance without losing today context | ✓ Good — clean visual hierarchy |
| Nord/Solarized MutedFg/EmptyFg use lighter colors (nord9, base00) | nord3 (#4C566A) and base01 (#586E75) too dark on typical dark backgrounds | ✓ Good — empty states and hints now readable |
| date_precision column: 'day', 'month', 'year', '' | Clean schema for multi-precision todos | ✓ Good — extensible, clean queries |
| Fuzzy dates stored as first-of-period (YYYY-MM-01, YYYY-01-01) | Valid SQL dates for range queries | ✓ Good — no special date handling needed |
| Segmented date input with auto-advance | More intuitive than typing separators | ✓ Good — natural tab flow |
| sectionID enum for 4-section todo panel | Boundary-aware reordering without HasDate() | ✓ Good — extensible section model |
| Reuse PendingFg/CompletedCountFg for circle indicators | No new theme roles needed | ✓ Good — consistent palette |
| boolIndex() helper for settings toggle mapping | Clean bool-to-option-index conversion | ✓ Good — reusable pattern |
| Settings save-on-close (no Enter/cancel) | Immediate feedback, fewer keystrokes, simpler mental model | ✓ Good — SettingChangedMsg on every cycle |
| Priority is visual-only, no auto-sort | Conflicts with manual J/K reordering — priority is visual indicator only | ✓ Good — preserves user-defined order |
| Inline selector (left/right arrows) for priority | Faster UX for 5 fixed options, prevents invalid input | ✓ Good — natural TUI interaction |
| Fixed 5-char badge slot for priority | "[P1] " or 5 spaces ensures column alignment regardless of priority | ✓ Good — clean visual alignment |
| HighestPriorityPerDay MIN/GROUP BY query | Single-pass efficient lookup for calendar indicators | ✓ Good — no N+1 queries |
| Named field constants replacing magic numbers | fieldTitle=0..fieldTemplate=4 for editField safety | ✓ Good — maintainable, prevents off-by-one |
| Priority cache in RenderWeekGrid | Cross-month week spans need independent store queries, like indicator cache | ✓ Good — consistent with existing patterns |
| OAuth 2.0 with PKCE over app passwords | Google disabled app passwords for Calendar in Sept 2024 | ✓ Good — secure desktop OAuth flow |
| Google REST API over CalDAV | Google-specific integration, simpler API, better docs | ✓ Good — clean JSON API, syncToken support |
| Events cached in-memory only | Events are ephemeral, rebuilt on each sync; no SQLite schema complexity | ✓ Good — simple, no cache invalidation |
| PKCE with S256 + ephemeral loopback port | Desktop OAuth security best practices | ✓ Good — no fixed port conflicts |
| persistingTokenSource wrapper | Auto-save token on refresh transparently | ✓ Good — zero user interaction for token refresh |
| All-day event dates as raw YYYY-MM-DD string | Prevents timezone conversion off-by-one errors | ✓ Good — consistent with existing date handling |
| SyncToken delta sync with 410 GONE retry | Efficient incremental updates, automatic full-sync fallback | ✓ Good — minimal API calls |
| Teal/cyan EventFg color family | Visually distinct from accent (indigo) and muted (grey) | ✓ Good — clear event/todo separation |
| Events inserted before todos in dated section | Calendar-driven items get visual priority over user-driven todos | ✓ Good — natural reading order |
| Non-selectable eventItem kind | Events are read-only; cursor skips automatically via selectableIndices | ✓ Good — no code changes to selection logic |
| Settings toggle Enabled/Disabled when AuthReady | Clean UX: action row for auth, cycling toggle for display control | ✓ Good — context-appropriate UI |
| PriorityColorHex casts lipgloss.Color to string | Avoids adding lipgloss dependency to status package | ✓ Good — clean package boundary |
| Restructured main() for subcommand routing | Config+DB load before branch, shared setup for TUI and status | ✓ Good — no duplicate initialization |
| Value receiver for refreshStatusFile | Matches Init/Update pattern; silent errors for best-effort writes | ✓ Good — consistent with Elm Architecture |

## Known Tech Debt

None.

## Current State

Shipped v2.3 with 10,506 LOC Go across 38 source files. Polybar status integration complete — `todo-calendar status` subcommand and TUI-driven state file updates provide real-time external status bar display.

---
*Last updated: 2026-02-23 after v2.3 milestone*
