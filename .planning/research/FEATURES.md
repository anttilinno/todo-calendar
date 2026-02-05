# Feature Research

**Domain:** TUI Calendar + Todo Application
**Researched:** 2026-02-05
**Confidence:** HIGH (based on analysis of calcurse, calcure, taskwarrior, todoman, khal, dooit, todo.txt ecosystem, and HN/community sentiment)

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete or broken.

#### Calendar

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Monthly calendar grid view | Every calendar app has this. It is the fundamental UI element. | LOW | Left pane in our layout. Standard 7-column grid with day numbers. |
| Navigate between months | Users need to look ahead and behind. Calcurse, khal, calcure all support this. | LOW | Arrow keys or h/l (vim). Must feel instant. |
| Today highlight | Users need immediate orientation -- "where am I?" | LOW | Bold, color, or inverse video on today's date. |
| Day-of-week headers | Without these the grid is unreadable. | LOW | Mo Tu We Th Fr Sa Su (locale-configurable start day is a nice-to-have, not table stakes). |
| National holidays in red | Explicitly in our project spec. calcure shows holidays. Users scanning a month need to see non-working days at a glance. | MEDIUM | Use `rickar/cal/v2` Go library -- supports Finland and 40+ countries. Holidays render in red/distinct color on the calendar grid. |

#### Todo

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Add a todo | Core CRUD. Every tool has this. | LOW | Inline entry in the right pane. Minimal friction -- type and press Enter. |
| Mark todo as done | The fundamental dopamine hit. Calcurse, taskwarrior, todoman all have completion. | LOW | Single keypress toggle (e.g., `x` or `Enter`). Visual strikethrough or checkmark. |
| Delete a todo | Users make mistakes or finish things. Calcurse uses `D`, taskwarrior uses `delete`. | LOW | Single keypress with optional confirmation (configurable). |
| Date-bound todos | Our spec requires showing todos for a selected month. Without date binding, the right pane has no calendar connection. | LOW | Each todo optionally tied to a date. Displayed when that date/month is in view. |
| Floating (undated) todos | Taskwarrior's default. Calcurse separates todos from appointments. Users need a place for "do this sometime" items. | LOW | Always visible in a separate section below date-bound items, or in a dedicated area. |
| Persist data across sessions | Every tool saves to disk. Data loss is unacceptable. | LOW | Local file storage (JSON, plain text, or similar). Auto-save on every mutation. |

#### UI/UX

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Split-pane layout | In our spec. Calcurse uses this exact pattern (calendar left, content right). It is the established pattern for this type of app. | MEDIUM | Calendar left, todo list right. Bubble Tea viewport management. |
| Keyboard-driven navigation | TUI users expect keyboard-first. Calcurse, khal, taskwarrior are all keyboard-only. Mouse is optional. | LOW | Vi-style (hjkl) or arrow keys. Tab to switch panes. |
| Responsive to terminal resize | Users resize terminals constantly. Calcure explicitly advertises this. | MEDIUM | Bubble Tea handles this via WindowSizeMsg, but layout math needs to adapt. |
| Visual feedback on actions | Adding, completing, deleting should provide immediate visual confirmation. | LOW | Highlight changes, status bar messages, or brief flash. |
| Quit without data loss | Auto-save or save-on-quit. Calcurse has `general.autosave` option. | LOW | Save on every write operation. No explicit "save" step needed. |
| Help / keybinding reference | Calcurse has `?` for help. Users need to discover available actions. | LOW | `?` key shows overlay or footer with keybindings. |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required for v1, but create real value.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Integrated calendar+todo in one view | Calcurse has this but is aging (ncurses, C). Khal is calendar-only. Todoman is todo-only. Taskwarrior is todo-only. A modern Go/Bubble Tea app combining both with a clean aesthetic is genuinely rare. | N/A (this IS the product) | This is our core value prop. Not a feature to add -- it is the product concept itself. |
| Holiday-aware calendar out of the box | Calcurse does NOT show national holidays. Khal does not either. Calcure shows them. Having holidays visible by default with zero config (auto-detect locale or simple country code setting) is a real differentiator vs calcurse/khal. | MEDIUM | `rickar/cal/v2` with Finnish holidays by default. Configurable country code. |
| Clean, modern TUI aesthetic | Calcurse looks dated (1990s ncurses). Bubble Tea + Lip Gloss enable modern styling that calcure (Python) achieves but calcurse does not. | MEDIUM | Lip Gloss borders, colors, spacing. This is about polish, not features. |
| Zero-config startup | Taskwarrior needs `.taskrc`. Khal needs `khal configure`. Todoman needs config for vdir paths. Calcurse works out of the box but with ugly defaults. Launch and immediately use -- no config file needed. | LOW | Sensible defaults. Config file optional for customization. Data directory auto-created. |
| Fast startup and rendering | Go compiles to native binary. Calcure is Python (slow startup). Calcurse is C (fast but old). A Go binary starts instantly and renders smoothly. | LOW | Inherent to Go + Bubble Tea. No runtime dependency. |
| Todo items visible on calendar dates | Show dots, marks, or count indicators on calendar dates that have todos. Calcure shows icons on event dates. This connects the two panes visually. | MEDIUM | Small dot or number badge on calendar dates that have associated todos. Powerful wayfinding feature. |

### Anti-Features (Deliberately NOT Building in v1)

Features that seem good but create complexity disproportionate to their value, or contradict the "intentionally minimal" v1 philosophy.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Recurring todos | Taskwarrior, calcurse, dooit all support recurrence. Seems essential. | Recurrence rules (daily, weekly, monthly, yearly, nth weekday) are a complexity explosion. RFC 5545 RRULE is notoriously hard to implement correctly. Edge cases around completion, regeneration, and skipped occurrences are endless. | v1: Users manually re-add recurring items. v2: Add simple recurrence (daily/weekly/monthly) once the core is solid. |
| Todo priorities / urgency | Taskwarrior has H/M/L priorities and urgency scores. Calcurse has 1-9 priority. | Priority systems add UI complexity (how to set, how to display, sort order interactions) and cognitive overhead ("Is this a 3 or a 4?"). HN consensus: most people ignore priority after initial setup. | v1: Order implies priority -- items at the top are more important. Users reorder manually. v2: Consider simple High/Normal/Low if users request it. |
| Subtasks / nested todos | Calcure and dooit support subtasks. Seems natural. | Subtask rendering in a TUI is tricky (indentation, collapse/expand, completion propagation). It also fundamentally changes the data model from flat list to tree. | v1: Flat todo list only. Users can use naming conventions like "Project: task" to group mentally. v2: Consider one level of nesting if validated. |
| CalDAV / cloud sync | Khal, todoman, calcurse all support CalDAV or sync. | Sync is an entire project unto itself -- conflict resolution, authentication, network error handling, partial sync states. It violates "local file storage" simplicity. | v1: Local files only. Users can put the data directory in Dropbox/Syncthing for DIY sync. v2+: Consider CalDAV if there is demand. |
| iCalendar (.ics) import/export | Calcurse, khal, todoman all support iCal. Standard interop. | Parsing iCal/RFC 5545 is complex (timezones, RRULE, VTODO, VEVENT, VALARM). It is a large surface area for a v1. | v1: Simple proprietary format (JSON or plain text). v2: Add .ics export. v3: Add .ics import. |
| Multiple calendar sources | Calcure shows events from multiple .ics files. Khal handles multiple CalDAV calendars. | Multiple sources means a merging layer, source attribution, color coding per source, and conflict handling. | v1: Single data source. One file for todos, one for any events. |
| Notifications / reminders | Calcurse has a full notification system with daemon mode. | Desktop notifications from a TUI require platform-specific integration (libnotify, osascript, etc.). Daemon mode is a separate architecture concern. | v1: No notifications. The app shows what is due when you open it. v2+: Consider terminal bell or system notification for overdue items. |
| Appointment / event scheduling | Calcurse's core feature. Time-blocked appointments with start/end times. | Our spec says "todo list" not "appointment scheduler." Adding time-blocked events means a completely different data model, conflict detection, and calendar rendering (time slots). | v1: Todos only (date-bound or floating). No time-of-day scheduling. This is a todo app with a calendar view, not a scheduling app. |
| Weekly / daily views | Calcurse has weekly view. Khal has day view. | Additional views multiply the rendering code and navigation complexity. Monthly view is sufficient for the "see your month, manage your todos" use case. | v1: Monthly view only. v2: Add week view if users want finer granularity. |
| Search / filtering | Taskwarrior has powerful regex filtering. Calcurse supports search. | Search requires a search UI, result highlighting, and filter state management. For a minimal todo list, scrolling is sufficient. | v1: No search. List is short enough to scan visually. v2: Add `/` search if lists get long. |
| Tags / categories / projects | Taskwarrior tags (+tag), contexts (@context), projects. | Tags require a tagging UI, filter-by-tag, tag management, and colored rendering. Significant UI surface area. | v1: No tags. Naming conventions suffice. v2: Consider simple color-coded categories. |
| Undo / redo | Valuable safety net for destructive actions. | Undo requires a command history stack, inverse operations for each action, and memory management. Non-trivial to implement correctly. | v1: Delete confirmation prompt. No undo. v2: Simple one-level undo. |
| Timer / pomodoro | Calcure has timers. Trendy productivity feature. | Timers are a fundamentally different feature (real-time countdown, notifications, state management). Scope creep disguised as productivity. | v1: Not a timer app. Use a dedicated pomodoro tool. |
| Vim-mode text editing | Calcurse supports vim-style editing in input fields. Calcure uses vim keys for navigation. | Full vim emulation in input fields is a rabbit hole. Navigation vim keys (hjkl) are table stakes; vim text editing (dd, yy, p, etc.) is not. | v1: Vi-style navigation (hjkl). Standard line editing in input fields (readline-style). |
| Weather / moon phases | Calcure shows weather and moon phases. | Requires external API calls, network dependencies, and UI space for non-core information. | v1: Not a weather app. Keep the UI focused on calendar + todos. |

## Feature Dependencies

```
[Monthly Calendar Grid]
    |
    +--requires--> [Date Navigation] (need to move between months)
    |
    +--requires--> [Today Highlight] (orientation within grid)
    |
    +--enhanced-by--> [Holiday Display] (needs calendar grid to render on)
    |
    +--enhanced-by--> [Todo Indicators on Dates] (dots/badges on calendar dates)

[Todo List - Right Pane]
    |
    +--requires--> [Add Todo] (core CRUD)
    |
    +--requires--> [Complete Todo] (core CRUD)
    |
    +--requires--> [Delete Todo] (core CRUD)
    |
    +--requires--> [Data Persistence] (must save to disk)
    |
    +--requires--> [Date-Bound Todos] (connects todos to calendar)
    |
    +--requires--> [Floating Todos] (undated catch-all section)

[Split-Pane Layout]
    |
    +--requires--> [Monthly Calendar Grid] (left pane content)
    |
    +--requires--> [Todo List] (right pane content)
    |
    +--requires--> [Pane Focus Switching] (Tab or similar)
    |
    +--requires--> [Terminal Resize Handling] (responsive layout)

[Date-Bound Todos] --connects--> [Monthly Calendar Grid]
    (selecting a date in calendar filters/highlights todos for that date)

[Todo Indicators on Dates] --requires--> [Date-Bound Todos] + [Monthly Calendar Grid]
    (can only show indicators if both systems exist)
```

### Dependency Notes

- **Split-Pane Layout requires both Calendar Grid and Todo List:** The layout is the container; without content for both panes it is meaningless. Build the panes first, then compose them.
- **Date-Bound Todos connect the two panes:** This is the critical integration point. Without it, the calendar and todo list are just two unrelated widgets sharing screen space. The date binding is what makes this product coherent.
- **Holiday Display enhances Calendar Grid:** Holidays are rendered as colored dates within the existing grid. The grid must exist and be navigable first.
- **Todo Indicators on Dates require both subsystems:** Dots/badges on calendar dates depend on both the calendar rendering and the date-bound todo data model.

## MVP Definition

### Launch With (v1)

Minimum viable product -- what is needed to validate the concept of a combined calendar+todo TUI.

- [x] Monthly calendar grid with day-of-week headers -- the foundational left pane
- [x] Navigate months forward/backward -- essential calendar interaction
- [x] Today highlighted -- user orientation
- [x] National holidays displayed in red -- key differentiator, low marginal cost with `rickar/cal/v2`
- [x] Split-pane layout (calendar left, todos right) -- the core product concept
- [x] Add todo with optional date -- core CRUD
- [x] Mark todo as complete -- core CRUD
- [x] Delete todo -- core CRUD
- [x] Date-bound todos shown for selected month -- the integration that makes calendar+todo valuable
- [x] Floating todos section -- catch-all for undated items
- [x] Local file persistence (JSON) -- data must survive restarts
- [x] Keyboard navigation (arrows/hjkl, Tab for pane switch) -- TUI standard
- [x] Help overlay (?) -- discoverability
- [x] Terminal resize handling -- Bubble Tea provides this, just need to wire it up

### Add After Validation (v1.x)

Features to add once core is working and users provide feedback.

- [ ] Todo indicators on calendar dates (dots/counts) -- triggered by: users saying "I can't tell which dates have todos without scrolling"
- [ ] Configurable country for holidays -- triggered by: non-Finnish users wanting their own holidays
- [ ] Edit todo text/date after creation -- triggered by: users making typos or rescheduling
- [ ] Simple todo reordering (move up/down) -- triggered by: users wanting manual priority via ordering
- [ ] Color themes / customization -- triggered by: users wanting to match their terminal aesthetic
- [ ] Configurable first day of week (Monday vs Sunday) -- triggered by: international users

### Future Consideration (v2+)

Features to defer until the core product is validated and users actively request them.

- [ ] Weekly view -- defer because: monthly view covers the primary use case; weekly adds rendering complexity
- [ ] Simple recurrence (daily/weekly/monthly) -- defer because: RRULE complexity; needs careful data model design
- [ ] One-level subtasks -- defer because: changes data model from flat list to tree
- [ ] Search / filter todos -- defer because: only needed when lists grow beyond visual scanning
- [ ] iCalendar export -- defer because: format complexity; only needed for interop
- [ ] Tags / categories with colors -- defer because: significant UI surface area
- [ ] Undo (one level) -- defer because: command history architecture

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Monthly calendar grid | HIGH | LOW | P1 |
| Month navigation | HIGH | LOW | P1 |
| Today highlight | HIGH | LOW | P1 |
| National holidays (red) | HIGH | MEDIUM | P1 |
| Split-pane layout | HIGH | MEDIUM | P1 |
| Add todo | HIGH | LOW | P1 |
| Complete todo | HIGH | LOW | P1 |
| Delete todo | HIGH | LOW | P1 |
| Date-bound todos | HIGH | LOW | P1 |
| Floating todos | MEDIUM | LOW | P1 |
| Local file persistence | HIGH | LOW | P1 |
| Keyboard navigation | HIGH | LOW | P1 |
| Help overlay | MEDIUM | LOW | P1 |
| Terminal resize | MEDIUM | MEDIUM | P1 |
| Todo indicators on dates | MEDIUM | MEDIUM | P2 |
| Edit todo | MEDIUM | LOW | P2 |
| Todo reordering | MEDIUM | LOW | P2 |
| Country config for holidays | LOW | LOW | P2 |
| Color themes | LOW | MEDIUM | P2 |
| First day of week config | LOW | LOW | P2 |
| Weekly view | MEDIUM | HIGH | P3 |
| Recurrence | MEDIUM | HIGH | P3 |
| Subtasks | LOW | HIGH | P3 |
| Search / filter | LOW | MEDIUM | P3 |
| iCal export | LOW | HIGH | P3 |
| Tags / categories | LOW | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible (v1.x)
- P3: Nice to have, future consideration (v2+)

## Competitor Feature Analysis

| Feature | calcurse | calcure | khal | taskwarrior | todoman | Our Approach (v1) |
|---------|----------|---------|------|-------------|---------|-------------------|
| Calendar view | Monthly + Weekly | Monthly | Monthly + Day | None | None | Monthly only |
| Todo list | Yes (priority 1-9) | Yes (subtasks, deadlines, timers) | No | Yes (full GTD) | Yes (basic) | Yes (minimal: add/complete/delete) |
| Combined calendar+todo | Yes (split pane) | Yes (side by side) | No (calendar only) | No (todo only) | No (todo only) | Yes (split pane, core concept) |
| National holidays | No | Yes | No | N/A | N/A | Yes (via rickar/cal) |
| Recurring items | Yes (RRULE) | No | Yes (iCal) | Yes | No | No (v1) |
| Priorities | Yes (1-9) | No info | N/A | Yes (H/M/L + urgency) | Yes (basic) | No (v1, ordering implies priority) |
| Subtasks | No | Yes | No | No (dependencies instead) | No | No (v1) |
| CalDAV sync | Experimental | Via .ics files | Yes (via vdirsyncer) | Taskserver | Yes (via vdirsyncer) | No (local only) |
| iCal support | Import + Export | Import (.ics) | Full | No (JSON) | Full (RFC 5545) | No (v1, JSON storage) |
| Vim keybindings | Partial | Yes | Yes (ikhal) | N/A (CLI) | N/A (CLI) | Yes (navigation) |
| Notifications | Yes (daemon) | No | No | No | No | No |
| Data format | Plain text (custom) | Plain text (custom) | iCal/vdir | JSON | iCal/vdir | JSON |
| Language | C | Python | Python | C++ (v3: Rust) | Python | Go |
| TUI framework | ncurses | Python curses | urwid | N/A (CLI) | N/A (CLI) | Bubble Tea |
| Zero-config startup | Mostly | Yes | No (needs configure) | No (needs .taskrc) | No (needs config) | Yes |
| Modern aesthetic | No (dated) | Yes | Functional | N/A | N/A | Yes (Lip Gloss) |

## Sources

- [calcurse official site and manual](https://calcurse.org/files/manual.html) -- HIGH confidence, authoritative
- [calcure GitHub and documentation](https://github.com/anufrievroman/calcure) -- HIGH confidence, primary source
- [taskwarrior documentation](https://taskwarrior.org/docs/) -- HIGH confidence, authoritative
- [todoman documentation](https://todoman.readthedocs.io/en/stable/) -- HIGH confidence, authoritative
- [khal GitHub](https://github.com/pimutils/khal) -- HIGH confidence, primary source
- [dooit on Terminal Trove](https://terminaltrove.com/dooit/) -- MEDIUM confidence, secondary source
- [rickar/cal Go holiday library](https://pkg.go.dev/github.com/rickar/cal/v2) -- HIGH confidence, Go package docs (supports Finland)
- [todo.txt format specification](https://github.com/todotxt/todo.txt) -- HIGH confidence, primary source
- [HN discussion on todo app design](https://news.ycombinator.com/item?id=44864134) -- MEDIUM confidence, community sentiment (key insight: simplicity beats features)
- [calcurse man page](https://www.mankier.com/1/calcurse) -- HIGH confidence, official documentation
- [khal usage examples](https://commandmasters.com/commands/khal-common/) -- MEDIUM confidence, tutorial site

---
*Feature research for: TUI Calendar + Todo Application*
*Researched: 2026-02-05*
