# Features Research: v1.6 Template Management & Recurring Todos

**Domain:** Template CRUD overlay, schedule attachment, auto-creation of recurring todos
**Researched:** 2026-02-07
**Confidence:** HIGH for template management (well-understood CRUD overlay pattern, existing codebase has 3 overlays to follow), MEDIUM for recurring todos (novel for this app, but well-established patterns in Taskwarrior, Todoist, Things)

---

## Template Management Overlay

### Table Stakes

Features that are absolutely required for a template management overlay. Without these, the overlay feels half-baked -- users can currently create templates (T) and use them (t) but cannot see, edit, rename, or manage them except by deleting during the template-select flow.

| Feature | Why Expected | Complexity | Depends On | Notes |
|---------|--------------|------------|------------|-------|
| Full-screen overlay listing all templates | Users need to see what templates exist. The current template-select mode is embedded inline in the todo pane and only shows name + truncated content. A proper overlay (matching settings/search/preview patterns) gives room for a complete list with metadata. | MEDIUM | Existing overlay infrastructure in `app/model.go` (showSettings, showSearch, showPreview pattern) | Follow the same pattern: `showTemplates bool`, `templates templates.Model`, route messages in Update, render in View. The app already has 3 overlays; this is overlay #4. |
| Template content preview in the list | When browsing templates, users need to see what each template contains -- not just its name. A side-by-side or below-list preview of the selected template's full content (rendered or raw). | MEDIUM | Glamour for markdown rendering (already a dependency), viewport for scrollable content | Two viable UX patterns: (1) split the overlay into list-left/preview-right like file managers, or (2) show preview below the list. Given the app is personal-use TUI, a simpler approach is highlighting a template and showing its full content below the list. |
| Delete template from the overlay | Users must be able to remove templates they no longer want. The current delete-during-select (d in templateSelectMode) is hidden and undiscoverable. | LOW | `store.DeleteTemplate(id)` already exists | Keybinding: `d` to delete selected template. No confirmation dialog needed for a personal app -- this matches the todo delete pattern (d deletes immediately). |
| Rename template | Users misspell names or want to reorganize. Currently there is no rename capability at all -- the only option is delete and recreate. | LOW | Need new `UpdateTemplate` or `RenameTemplate` store method. Current templates table has UNIQUE constraint on name. | Inline rename: press `r`, text input appears with current name, Enter saves, Esc cancels. Matches the edit-text pattern in the todo list. New store method: `UPDATE templates SET name = ? WHERE id = ?` |
| Edit template content | Templates evolve. A "Daily Plan" template might need a new section added. Currently, editing requires deleting and recreating. | MEDIUM | External editor integration (already built in Phase 16) or in-app textarea (already used for template creation) | Two options: (1) open in external editor via `$EDITOR` (reuse Phase 16 pattern), or (2) use the in-app textarea. Recommendation: use external editor because template content is multi-line markdown and the textarea is awkward for editing existing content. The external editor pattern is proven. |
| Overlay keybinding from normal mode | Users need a discoverable way to open the template management overlay. Currently `t` = use template, `T` = create template, but there is no "manage templates" entry point. | LOW | New key binding in app-level KeyMap | Use `ctrl+t` or add to settings. Recommendation: use `M` (Manage templates) since `m` is unused and uppercase letters are used for "heavier" operations (A = add dated, E = edit date, K/J = move). |
| Close overlay with Esc | Standard pattern across all existing overlays. | LOW | Built into the overlay pattern | Emit a `CloseMsg` like search and preview do. |

### Differentiators

Features that go beyond basic CRUD and make template management genuinely useful.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Template reordering | Let users arrange templates in preferred order (most-used first). Matches the todo reorder pattern (K/J). | MEDIUM | Need `sort_order` column on templates table (new migration). Reuse the SwapOrder pattern from todos. Currently templates are ordered by name (alphabetical). |
| Template count / usage indicator | Show how many times each template has been used, or when it was last used. Helps users identify unused templates for cleanup. | MEDIUM | Would require a `usage_count` or `last_used` column. Minor schema change, but unclear value for a personal app with 7-15 templates. |
| Duplicate template | Create a copy of an existing template with "(copy)" suffix. Useful when creating a variation of an existing template. | LOW | `AddTemplate(name + " (copy)", original.Content)`. Simple store operation. |
| Filter/search within template list | Type to narrow the template list. Useful if template count grows beyond 15-20. | LOW | Reuse the filter pattern from the todo list. For a personal app with <20 templates, this is marginal. |
| Template categories / tags | Group templates (e.g., "Work", "Personal", "Dev"). | HIGH | Requires new data model (tags table or category column), UI for managing categories, filtering by category. Over-scoped for this milestone. |

### Anti-Features

Things to deliberately NOT build for template management.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Template import/export (files, JSON, YAML) | Personal single-user app. Import/export adds file picker UI, format validation, conflict handling. Zero real users need this. | Templates are managed in-app. Users can copy/paste content via external editor. |
| Template versioning / history | Tracking changes to templates adds significant schema complexity (versions table, diff storage) for a feature nobody asked for. | Edit overwrites. If the user makes a mistake, they can edit again. The original content is not precious enough to version. |
| Template sharing between devices | Would require sync infrastructure, cloud storage, or file-based export. Massively out of scope. | Single-device app. If users want to move templates, they can copy the SQLite database. |
| Rich template editor in TUI | Building a proper multi-line editor with syntax highlighting, cursor movement, selection, undo/redo is a massive undertaking. The existing textarea component is adequate for creation but not for editing complex documents. | Use external editor for editing template content. This is already the pattern for todo bodies. |
| Template variables with validation types | Typed placeholders like `{{.Date:date}}` or `{{.Priority:enum:high,medium,low}}` that validate input. | All placeholders are free-form text input. The template provides structure, not validation. This was already decided in v1.4 research and remains correct. |

---

## Recurring / Scheduled Todos

### Table Stakes

Features that are required for a recurring todo system to feel usable. Based on patterns from Taskwarrior (template/instance model), Todoist (recurring dates with visual scheduler), and Things (template-based repeating with auto-population).

| Feature | Why Expected | Complexity | Depends On | Notes |
|---------|--------------|------------|------------|-------|
| Attach a schedule to a template | The core concept: a template can optionally have a recurrence rule. "Daily Plan" runs every weekday. "Weekly Review" runs every Friday. The schedule lives on the template, not on individual todos. | MEDIUM | Templates table needs new columns: `schedule_type` (NULL, "daily", "weekdays", "weekly", "monthly"), `schedule_value` (day-of-week number, day-of-month number, comma-separated weekday list). | This is the Taskwarrior approach: the template IS the recurring task definition. Todoist and Things also use a template/instance model internally. NULL schedule_type means the template is not recurring. |
| Schedule types: daily, specific weekdays, monthly Nth | Three schedule cadences cover 95% of personal recurring task use cases. Daily = every day. Weekdays = pick one or more days of the week (Mon, Wed, Fri). Monthly = specific day of month (1st, 15th, last). | MEDIUM | Schedule storage, date math for determining if a schedule fires on a given date | Todoist supports dozens of recurrence patterns. Taskwarrior supports arbitrary durations. For a personal TUI app, three types is the right scope. Avoid "every N days" or "every N weeks" -- they add complexity without proportional value for this use case. |
| Schedule picker UI in template management overlay | When a template is selected, user can press a key (e.g., `S`) to set/change/remove its schedule. This opens a sub-flow: pick type (daily/weekday/monthly), then configure (which days/which date). | MEDIUM | Template overlay must be built first | UX: Press `S` on a template. Cycle through schedule types with left/right arrows. For weekdays, toggle individual days (MTWTFSS). For monthly, enter a number (1-28). Press Enter to save, Esc to cancel. This is similar to the settings overlay's cycle-through-options pattern. |
| Auto-create todos from scheduled templates on app launch | When the app starts, check all scheduled templates. For each, determine which dates in a rolling window (today + 6 days = 7 day window) should have a todo. Create missing todos. This is the core automation. | HIGH | Schedule storage, todo creation, deduplication logic | This is the most complex feature in the milestone. The critical challenge is deduplication: do NOT create duplicate todos if the app is launched multiple times in the same day. See Pitfalls section for details. |
| Deduplication of auto-created todos | If the user launches the app 5 times on Monday, the "Daily Plan" todo for Monday must only be created once. | MEDIUM | Need a way to track which (template, date) pairs have already been created. | Two approaches: (1) Track in a `schedule_log` table: `(template_id, date, created_at)`. Before creating, check if (template_id, date) exists. (2) Add `template_id` and `source_date` columns to the todos table itself. Approach 2 is simpler and also enables the "visual indicator" feature below. Recommendation: add `template_id INTEGER REFERENCES templates(id) ON DELETE SET NULL` and `source_date TEXT` to todos. A todo with non-null template_id + source_date was auto-created. |
| Placeholder prompting for auto-created todos | When a scheduled template has placeholders (e.g., `{{.Focus}}` in "Daily Plan"), the auto-created todo needs those values filled in. The project context says: prompt on first launch. | HIGH | Placeholder extraction (already built in `tmpl` package), sequential prompting (already built in `placeholderInputMode`) | This is the UX challenge. If 3 templates fire for today and each has 2 placeholders, the user faces 6 prompts on launch. Options: (1) Prompt sequentially on first launch -- annoying but complete. (2) Create todos with unfilled placeholders and let user fill them later -- less friction but bodies have raw `{{.Focus}}` text. (3) Hybrid: create todos, mark them as "needs attention", prompt when user first views/selects them. Recommendation: option 2 with a visual indicator showing the todo needs placeholder values filled. Users can press a key to fill placeholders on-demand. This avoids blocking app startup. |
| Visual indicator for auto-created / recurring todos | Users need to distinguish auto-created recurring todos from manually created ones. Todoist uses a recurring icon. Things shows repeating items differently. | LOW | `template_id` column on todos (from deduplication feature) | Show a small indicator like `[R]` (recurring) or a repeat symbol next to auto-created todos, similar to how `[+]` indicates a todo has a body. Check `todo.TemplateID != 0` to determine if it came from a schedule. |
| Schedule display in template list | When browsing templates in the overlay, scheduled templates should show their cadence (e.g., "Daily", "Mon/Wed/Fri", "15th of month"). Non-scheduled templates show nothing. | LOW | Schedule data on template rows | Render schedule as a dimmed suffix after the template name: `Daily Plan  (every weekday)` or `Weekly Review  (every Fri)`. |

### Differentiators

Features that would make the recurring system feel polished but are not strictly required.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| "Fill placeholders" action on existing todos | For auto-created todos with unfilled `{{.Placeholder}}` text, press a key to enter the placeholder-fill flow retroactively. This completes the deferred-prompting UX. | MEDIUM | Parse body for `{{.VarName}}` patterns, prompt for values, re-execute template, update body. Reuses existing placeholder infrastructure. |
| Pause/resume schedule | Temporarily disable a schedule without deleting it. Things supports this: "You can pause repeating to-dos without losing their settings." | LOW | Add `schedule_enabled BOOLEAN DEFAULT 1` column. Paused templates are skipped during auto-creation. Toggle with a key in the overlay. |
| Skip individual occurrences | Mark a specific day's auto-created todo as "skipped" rather than deleting it (which might cause it to be recreated). | MEDIUM | Need to distinguish "deleted by user" from "never created". The dedup log handles this if we record skips. Adds complexity to the dedup logic. |
| Configurable rolling window | Let users set the look-ahead window (default 7 days) in settings. Some users might want 14 days, others just today. | LOW | New config field `recurring_window_days`. Default 7. Used in auto-creation logic. |
| Recurring todo summary on launch | On first launch of the day, show a brief message: "Created 3 recurring todos for today." Gives visibility into what the automation did. | LOW | Count created todos during auto-creation, emit a message. Display as a transient status bar notification. |
| Schedule from ISO weekday names | Allow typing "Monday" or "Mon" instead of number 1 when configuring weekly schedules. Natural language for weekday selection. | LOW | Map string names to weekday numbers. Already know the pattern from `config.go` day-of-week handling. |

### Anti-Features

Things to deliberately NOT build for recurring todos.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Natural language date parsing ("every other Tuesday starting March 1") | Todoist's NLP date parsing is their competitive advantage built over years with a dedicated team. Implementing even a subset is a large NLP problem. A TUI app with 3 schedule types does not need this. | Structured picker: select type, configure with arrow keys/number input. Three types, three UIs. Clear and unambiguous. |
| Calendar-based schedule picker | A visual calendar where you click dates to set recurrence. Requires a complex interactive calendar widget beyond what the existing calendar component provides. | Simple type/value configuration. For weekdays, toggle checkboxes. For monthly, enter a number. |
| "Every N days/weeks/months" flexible intervals | "Every 3 days" or "every 2 weeks" adds significant date math complexity (tracking the last occurrence, computing the next). The three base types (daily, weekdays, monthly) cover the practical use cases. | Stick to daily, specific weekdays, and monthly Nth. If a user wants "every other week on Monday", they can create todos manually for the off-weeks. |
| Recurring todos without templates | Could allow any todo to be marked as recurring independently of the template system. This creates two parallel systems. | All recurring todos flow through templates. A template is the "definition" of a recurring task. This is the Taskwarrior model and it is architecturally clean. If the user wants a recurring todo without a template body, they create a template with no placeholders and minimal content. |
| End dates on schedules | "Repeat daily until March 31" adds temporal scoping that complicates the auto-creation logic (must check end date, handle expired schedules). | Schedules run indefinitely until the user pauses or removes them. For a personal app, manual control is sufficient. |
| Time-of-day scheduling | "Create todo at 9am every day." The app has no concept of time, only dates. Adding time would require a timer/scheduler daemon running in the background. | Auto-creation happens on app launch. Todos are dated, not timed. The user opens the app when they want to see their tasks. |
| Completion-based recurrence ("every 3 days after completion") | Todoist's `every!` pattern where the next occurrence is relative to when you completed the last one. Requires tracking completion timestamps and computing relative offsets. | All recurrence is calendar-based (periodic). "Every Monday" means every Monday, regardless of when you completed last Monday's todo. This is simpler and matches the calendar-centric design of the app. |
| Batch operations on recurring todos | "Delete all future occurrences" or "edit all future occurrences" like Google Calendar. Requires parent-child relationship management and propagation logic. | Each auto-created todo is independent after creation. Deleting one does not affect others. To stop future creation, pause or remove the schedule on the template. |
| Notification / reminder system | Recurring tasks in Todoist and Things can trigger notifications. A TUI app has no notification infrastructure. | The app creates todos; users see them when they open the app. No push notifications, no system tray, no daemon. |

---

## Feature Dependencies

```
Template Management Overlay (Phase A)
    |
    +-- REQUIRES: existing overlay infrastructure (settings, search, preview patterns)
    +-- REQUIRES: existing store.TodoStore interface (ListTemplates, FindTemplate, DeleteTemplate, AddTemplate)
    +-- NEW: store methods for UpdateTemplate (rename), UpdateTemplateContent
    +-- NEW: templates package (internal/templates/model.go, keys.go, styles.go)
    +-- MODIFIES: app/model.go (add showTemplates, templates fields, routing)
    +-- MODIFIES: app/keys.go (add TemplateManage keybinding)
    +-- INDEPENDENT OF: recurring/scheduling (can ship without it)

Schedule Attachment (Phase B)
    |
    +-- REQUIRES: Template Management Overlay (Phase A) -- schedule picker lives in the overlay
    +-- REQUIRES: schema migration (new columns on templates table)
    +-- NEW: schedule type/value columns on templates
    +-- NEW: schedule display in template list
    +-- NEW: schedule picker sub-flow in overlay
    +-- MODIFIES: store interface (schedule CRUD on templates)
    +-- INDEPENDENT OF: auto-creation (schedules can be set without auto-creation working yet)

Auto-Creation of Recurring Todos (Phase C)
    |
    +-- REQUIRES: Schedule Attachment (Phase B) -- must have schedules to create from
    +-- REQUIRES: schema migration (template_id, source_date on todos table)
    +-- NEW: auto-creation logic (run on app startup)
    +-- NEW: deduplication via (template_id, source_date) uniqueness
    +-- NEW: visual indicator for recurring todos [R]
    +-- MODIFIES: Todo struct (add TemplateID, SourceDate fields)
    +-- MODIFIES: store interface (new method for checking existing recurring todos)
    +-- MODIFIES: app model Init() or startup sequence
    +-- OPTIONAL: placeholder prompting (can defer to "fill later" UX)
```

**Strict dependency chain: Phase A -> Phase B -> Phase C.** Template management must exist before schedules can be attached. Schedules must exist before auto-creation can use them.

---

## MVP Recommendation

### Phase A MVP (Template Management Overlay)
1. Full-screen overlay listing all templates (name + schedule indicator + content preview)
2. Delete template (d key)
3. Rename template (r key, inline text input)
4. Edit template content via external editor (e/o key, reuse Phase 16 pattern)
5. Open overlay with a keybinding (M or ctrl+t from normal mode)
6. Close with Esc

**Defer:** Template reordering, duplicate template, filter/search within list. These are nice-to-haves that can ship in a later milestone.

### Phase B MVP (Schedule Attachment)
1. Schema migration adding `schedule_type` and `schedule_value` to templates table
2. Schedule picker in overlay: press S on a template, cycle through types (none/daily/weekdays/monthly), configure value
3. Schedule display in template list row
4. Store methods for reading/writing schedule data

**Defer:** Pause/resume schedule (can add `schedule_enabled` column later). Configurable window size.

### Phase C MVP (Auto-Creation)
1. Schema migration adding `template_id` and `source_date` to todos table
2. On app startup: iterate scheduled templates, compute dates in 7-day rolling window, create missing todos
3. Deduplication: check (template_id, source_date) before creating
4. Visual indicator [R] on auto-created todos
5. Auto-created todos with placeholders get unfilled `{{.Var}}` text in body (defer prompting)

**Defer:** Interactive placeholder prompting on launch, "fill placeholders" retroactive action, skip individual occurrences, launch summary message.

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Risk | Priority |
|---------|-----------|-------------------|------|----------|
| Template list overlay (view all) | HIGH | MEDIUM | LOW | P0 |
| Delete template from overlay | HIGH | LOW | LOW | P0 |
| Rename template | HIGH | LOW | LOW | P0 |
| Edit template content | HIGH | MEDIUM | LOW | P0 |
| Schedule attachment to template | HIGH | MEDIUM | MEDIUM | P0 |
| Schedule picker UI | MEDIUM | MEDIUM | MEDIUM | P1 |
| Auto-create recurring todos | HIGH | HIGH | HIGH (dedup) | P1 |
| Deduplication logic | CRITICAL | MEDIUM | HIGH | P0 (part of auto-create) |
| Visual indicator for recurring todos | MEDIUM | LOW | LOW | P1 |
| Schedule display in template list | MEDIUM | LOW | LOW | P1 |
| Placeholder prompting for auto-created | MEDIUM | HIGH | HIGH (UX) | P2 (defer) |
| Fill placeholders retroactively | MEDIUM | MEDIUM | LOW | P2 |
| Template reordering | LOW | MEDIUM | LOW | P3 |
| Pause/resume schedule | LOW | LOW | LOW | P3 |
| Configurable rolling window | LOW | LOW | LOW | P3 |

---

## UX Patterns: How Things Should Look and Feel

### Template Management Overlay Layout

The overlay should follow the existing overlay pattern (vertically centered, full-screen takeover):

```
                    Templates

  > Daily Plan              (every weekday)
    Weekly Review            (every Fri)
    Bug Report
    Checklist
    Code Review
    Feature Spec
    Meeting Notes
    PR Checklist

  ─────────────────────────────────────────

  ## Daily Plan

  ### Top Priorities
  1.
  2.
  3.

  ### Tasks
  - [ ]
  - [ ]

  r rename | e edit | d delete | S schedule | esc close
```

Key UX decisions:
- **List on top, preview on bottom** (simpler than side-by-side, works at any terminal width)
- **Selected template's content shown below separator** (rendered with Glamour or shown as raw text)
- **Schedule shown as dimmed suffix** on templates that have one
- **Hint bar at bottom** showing available actions

### Schedule Picker Sub-Flow

When pressing `S` on a template, a small inline picker appears:

```
  Schedule: < Weekdays >

  [ ] Mon  [x] Tue  [ ] Wed  [x] Thu  [x] Fri  [ ] Sat  [ ] Sun

  enter save | esc cancel
```

For monthly:
```
  Schedule: < Monthly >

  Day of month: 15

  enter save | esc cancel
```

For daily:
```
  Schedule: < Daily >

  No additional configuration needed.

  enter save | esc cancel
```

UX decisions:
- **Left/right arrows cycle schedule type** (None / Daily / Weekdays / Monthly) -- matches settings overlay pattern
- **Weekday selection uses space to toggle** individual days (checkbox pattern)
- **Monthly uses text input** for day number (1-28, or "last")
- **None removes the schedule** (sets schedule_type to NULL)

### Auto-Created Todo Appearance

Auto-created todos appear in the normal todo list with a subtle indicator:

```
  February 2026
  ──────────
  > [ ] Daily Plan [R] [+]              2026-02-07
    [ ] Weekly Review [R] [+]            2026-02-07
    [x] Write documentation              2026-02-07
    [ ] Fix login bug                    2026-02-06
```

UX decisions:
- **[R] indicator** shown after todo text, before [+] body indicator, styled in a muted/accent color
- **Auto-created todos get the template name as their title** (e.g., "Daily Plan")
- **Auto-created todos are dated** to their scheduled date
- **Auto-created todos sort like normal todos** (by sort_order within their date)
- **Auto-created todos are fully editable** -- once created, they are regular todos. Users can rename, change date, delete, toggle completion.

---

## Competitor Feature Comparison

| Feature | Taskwarrior | Todoist | Things | Our Approach |
|---------|-------------|---------|--------|-------------|
| Recurrence model | Template/instance with mask tracking | Single task with shifting due date | Template creates copies on schedule | Template/instance: scheduled templates auto-create independent todos |
| Schedule types | Arbitrary durations (daily, weekly, monthly, yearly, Ndays, Nweeks, etc.) | Natural language (any pattern expressible in English) | Fixed/after-completion with weekday selection | Three types: daily, specific weekdays, monthly Nth |
| Instance generation | On report display, controlled by `recurrence.limit` | On completion of current instance | Auto-populates Today list on scheduled date | On app launch, rolling 7-day window |
| Visual indicator | Recurring status shown in task attributes | Recurring icon in task row | Distinct in Quick Find "repeating" filter | [R] indicator next to todo text |
| Template management | N/A (no template UI) | N/A (no templates) | N/A (templates are internal) | Full CRUD overlay with preview, rename, edit, schedule |
| Placeholder prompting | N/A | N/A | N/A | Deferred: unfilled placeholders in body, fill on demand |

**Key differentiator:** No competitor combines template management (CRUD overlay with preview) with schedule attachment on templates. Taskwarrior has recurrence but no template UI. Todoist has recurrence but no markdown templates. Our approach of "template = recurring task definition with rich content" is unique.

---

## Sources

### HIGH Confidence (official docs, authoritative, verified)
- [Taskwarrior recurrence documentation](https://taskwarrior.org/docs/recurrence/) -- Template/instance model, mask tracking, periodic vs chained recurrence
- [Todoist recurring dates documentation](https://www.todoist.com/help/articles/introduction-to-recurring-dates-YUYVJJAV) -- Supported patterns: every/every!, daily/weekly/monthly/yearly, custom intervals
- [Todoist visual recurring scheduler](https://www.todoist.com/help/articles/new-visual-interface-for-recurring-tasks-jun-30-nMZ4CZjHb) -- Visual picker with presets and custom menu
- [Things repeating to-dos](https://culturedcode.com/things/support/articles/2803564/) -- Template-based repeating, auto-population to Today list, pause/resume
- Existing codebase: `internal/app/model.go` -- 3 overlay patterns (settings, search, preview) to follow
- Existing codebase: `internal/todolist/model.go` -- Template select/create/placeholder flow already built
- Existing codebase: `internal/store/sqlite.go` -- Templates table schema, migration pattern with PRAGMA user_version

### MEDIUM Confidence (multiple sources agree)
- [Amplenote recurring tasks comparison](https://www.amplenote.com/blog/five_best_todo_list_apps_for_recurring_tasks) -- Survey of 5 todo apps' recurring task implementations
- [Bubble Tea overlay pattern](https://github.com/charmbracelet/bubbletea) -- Reactive message-driven architecture for overlay/modal management
- [Building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) -- TUI architecture patterns: each view does one thing, keyboard shortcuts always visible

### LOW Confidence (single source, needs validation during implementation)
- Deduplication via (template_id, source_date) uniqueness constraint -- Logical approach but needs testing for edge cases (what if user manually creates a todo with the same title on the same day?)
- Placeholder deferral UX (leaving `{{.Var}}` in body) -- No precedent found in other apps. May look confusing to users. Should validate in implementation whether this needs a better UX (e.g., replacing placeholders with `[FILL: VarName]` human-readable markers instead of Go template syntax).
- Rolling 7-day window as default -- No prior art for this specific approach. Taskwarrior uses `recurrence.limit` (number of instances). Todoist generates on completion. Things generates for "today". The 7-day window is our invention and may need adjustment based on real usage.

---
*Feature research for: TUI Calendar v1.6 Template Management & Recurring Todos*
*Researched: 2026-02-07*
