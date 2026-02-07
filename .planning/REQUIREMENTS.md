# v1.6 Requirements: Templates & Recurring

## Template Management Overlay

### REQ-20: Template management overlay
Full-screen overlay listing all templates with cursor navigation. Shows template name and schedule indicator (if any). Follows the established overlay pattern (settings, search, preview). Opens with a keybinding from normal mode, closes with Esc.

### REQ-21: Template content preview
Selected template's full content displayed below the template list in the overlay, separated by a horizontal rule. Content rendered as text (raw markdown, not glamour-rendered, to show actual template syntax including placeholders).

### REQ-22: Delete template from overlay
Press `d` on a selected template to delete it permanently. No confirmation dialog (matches todo delete pattern). Cascade-deletes any attached schedules.

### REQ-23: Rename template
Press `r` on a selected template to enter rename mode. Text input pre-filled with current name. Enter saves, Esc cancels. Handles duplicate name error gracefully.

### REQ-24: Edit template content
Press `e` on a selected template to open its content in the external editor ($VISUAL/$EDITOR/vi). Reuses the Phase 16 external editor pattern. On save, template content is updated in the store.

### REQ-25: Template overlay keybinding
`M` key in normal mode (when no overlay is active) opens the template management overlay. Listed in the help bar expanded view.

## Recurring Schedules

### REQ-26: Schedule schema
New `schedules` table (migration v4) with: id, template_id (FK CASCADE to templates), cadence_type TEXT, cadence_value TEXT, placeholder_defaults TEXT (JSON), created_at. New columns on todos table (migration v5): schedule_id (FK SET NULL to schedules), schedule_date TEXT. Unique index on (schedule_id, schedule_date) for deduplication.

### REQ-27: Schedule rule types
Three cadence types supported:
- `daily` -- every day
- `weekdays` -- Monday through Friday
- `weekly` -- specific days of week (e.g., mon,wed,fri)
- `monthly` -- specific day of month (1-28, or clamp to last day for short months)

Rule stored as cadence_type + cadence_value columns (e.g., type="weekly", value="mon,fri").

### REQ-28: Schedule CRUD in store
TodoStore interface extended with schedule operations: AddSchedule, ListSchedules, ListSchedulesForTemplate, DeleteSchedule, UpdateSchedule, TodoExistsForSchedule, AddScheduledTodo. SQLite implements all; JSON store stubs.

### REQ-29: Schedule picker UI
Press `S` on a template in the management overlay to add/manage its schedule. Left/right arrows cycle cadence type (None/Daily/Weekdays/Weekly/Monthly). For weekly: toggle individual weekdays with space. For monthly: text input for day number. Enter saves, Esc cancels. Shows existing schedule if one exists.

### REQ-30: Schedule display in template list
Templates with schedules show a dimmed suffix in the overlay list: "(daily)", "(weekdays)", "(Mon/Wed/Fri)", "(15th of month)". Non-scheduled templates show no suffix.

### REQ-31: Auto-create on app launch
On startup, after store initialization and before TUI launch, iterate all schedules. For each schedule, check dates in a rolling 7-day window (today + 6 days). For matching dates where no todo exists yet (dedup check), create a todo with: title = template name, date = scheduled date, body = template content executed with placeholder defaults (or empty values). Runs synchronously in main.go.

### REQ-32: Recurring todo visual indicator
Auto-created todos (those with non-null schedule_id) display an `[R]` indicator after the todo text, styled in a muted/accent color. Shown alongside existing `[+]` body indicator.

### REQ-33: Placeholder defaults at schedule creation
When creating a schedule for a template that has `{{.Variable}}` placeholders, prompt the user to fill default values for each placeholder. Store as JSON in schedules.placeholder_defaults column. Auto-created todos use these defaults. Templates without placeholders skip this step.

## Out of Scope (v2+)

- Complex cadences ("every 2nd Tuesday", "every N days")
- Pause/resume schedules
- Template reordering in overlay
- Configurable rolling window size
- Natural language date parsing
- Completion-based recurrence
- End dates on schedules
- Retroactive creation for missed days
- Template import/export

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| REQ-20 | Phase 20 | Complete |
| REQ-21 | Phase 20 | Complete |
| REQ-22 | Phase 20 | Complete |
| REQ-23 | Phase 20 | Complete |
| REQ-24 | Phase 20 | Complete |
| REQ-25 | Phase 20 | Complete |
| REQ-26 | Phase 21 | Complete |
| REQ-27 | Phase 21 | Complete |
| REQ-28 | Phase 21 | Complete |
| REQ-29 | Phase 22 | Complete |
| REQ-30 | Phase 22 | Complete |
| REQ-31 | Phase 22 | Complete |
| REQ-32 | Phase 22 | Complete |
| REQ-33 | Phase 22 | Complete |
