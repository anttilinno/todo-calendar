---
phase: 22-auto-creation-schedule-ui
verified: 2026-02-07T15:30:00Z
status: passed
score: 5/5 must-haves verified
---

# Phase 22: Auto-Creation & Schedule UI Verification Report

**Phase Goal:** Users can attach recurring schedules to templates and scheduled todos are auto-created on app launch.
**Verified:** 2026-02-07T15:30:00Z
**Status:** passed
**Re-verification:** No -- initial verification

## Build & Test

- `go build ./...` -- passes, zero errors
- `go test ./...` -- all 45 tests pass (9 recurring generate, 26 recurring rule, 9 store schedule, 1 store seed)

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can press S to open schedule picker that cycles cadence types with arrows, toggles weekdays with space, and accepts day number for monthly | VERIFIED | `tmplmgr/model.go:211-258` handles `keys.Schedule` in listMode, enters scheduleMode. `updateScheduleMode` (lines 312-457) handles Left/Right to cycle cadence index across `["none","daily","weekdays","weekly","monthly"]`, Up/Down+Toggle for weekly weekday selection, monthly text input via `monthlyInput`, Enter saves, Esc cancels. Existing schedules are loaded from store (lines 223-256). |
| 2 | Templates with schedules show dimmed suffix in overlay list | VERIFIED | `tmplmgr/model.go:678-714` `scheduleLabel()` builds human-readable suffix from schedule data -- `(daily)`, `(weekdays)`, `(Mon/Wed/Fri)`, `(15th of month)`. Rendered with `ScheduleSuffix` style (line 528-531 in View) using `MutedFg` color (styles.go:37). Templates without schedules return `""` (line 681). |
| 3 | On app launch, recurring.AutoCreate() runs before TUI starts, creating todos for matching dates in rolling 7-day window with deduplication | VERIFIED | `main.go:42` calls `recurring.AutoCreate(s)` after store creation but before `tea.NewProgram`. `generate.go:19-21` delegates to `AutoCreateForDate` which iterates schedules (line 27), builds rule strings (line 28), checks cadence matching (line 50), deduplicates via `TodoExistsForSchedule` (line 54), and calls `AddScheduledTodo` (line 57). Window is 7 days (lines 48-49). 9 dedicated tests cover daily/weekly/monthly/dedup/placeholder-defaults/orphan/bad-cadence scenarios, all passing. |
| 4 | Auto-created todos display [R] indicator after todo text in muted color | VERIFIED | `todolist/model.go:1067-1069` in `renderTodo()`: `if t.ScheduleID > 0 { b.WriteString(" " + m.styles.RecurringIndicator.Render("[R]")) }`. `RecurringIndicator` style at `todolist/styles.go:34` uses `t.MutedFg`. `store/todo.go:19` has `ScheduleID int` field. `store/sqlite.go:643` sets `ScheduleID: scheduleID` in `AddScheduledTodo`. |
| 5 | When scheduling a template with placeholders, user is prompted to fill default values once; auto-created todos use stored defaults | VERIFIED | `tmplmgr/model.go:414-431` after Enter in scheduleMode, calls `tmpl.ExtractPlaceholders(sel.Content)` and if placeholders exist, transitions to `placeholderDefaultsMode` with pending cadence stored. `updatePlaceholderDefaultsMode` (lines 460-497) steps through each placeholder with input, stores values, and on final Enter serializes to JSON and saves schedule. Pre-fills from existing schedule defaults (lines 422-424). `generate.go:41-46` parses defaults JSON and passes to `tmpl.ExecuteTemplate`. Test `TestAutoCreatePlaceholderDefaults` confirms `{"Project":"Alpha","Owner":"Alice"}` produces `"Project: Alpha\nOwner: Alice"` in todo body. |

**Score:** 5/5 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/recurring/generate.go` | AutoCreate engine | VERIFIED | 86 lines, substantive implementation, no stubs. Imported and called from main.go. |
| `internal/recurring/generate_test.go` | Test coverage | VERIFIED | 303 lines, 9 test functions covering all cadence types + edge cases. All pass. |
| `internal/recurring/rule.go` | ParseRule + MatchesDate | VERIFIED | 134 lines, complete implementation with daily/weekdays/weekly/monthly support and month-end clamping. 26 tests pass. |
| `internal/todolist/model.go` | [R] indicator in renderTodo | VERIFIED | Lines 1067-1069 render `[R]` when `t.ScheduleID > 0`. |
| `internal/todolist/styles.go` | RecurringIndicator style | VERIFIED | Line 34, uses `t.MutedFg` for muted color. |
| `internal/tmplmgr/model.go` | scheduleMode + placeholderDefaultsMode | VERIFIED | 743 lines. scheduleMode (lines 312-457) with full cadence cycling, weekday toggle, monthly input. placeholderDefaultsMode (lines 460-497) with multi-step prompting. |
| `internal/tmplmgr/keys.go` | Schedule/Left/Right/Toggle bindings | VERIFIED | Lines 62-76 define Schedule (s), Left (left/h), Right (right/l), Toggle (space). |
| `internal/tmplmgr/styles.go` | Schedule-related styles | VERIFIED | Lines 18-23 define ScheduleSuffix, ScheduleActive, ScheduleInactive, ScheduleDay, ScheduleDaySelected, SchedulePrompt. |
| `internal/store/todo.go` | Schedule struct + Todo.ScheduleID | VERIFIED | Lines 37-44 define Schedule struct. Lines 19-20 add ScheduleID and ScheduleDate to Todo. |
| `internal/store/sqlite.go` | Schedule CRUD + dedup + AddScheduledTodo | VERIFIED | Lines 540-646. Full CRUD with cascade on template delete, SET NULL on schedule delete, unique index for dedup. 9 store tests pass. |
| `main.go` | recurring.AutoCreate(s) call | VERIFIED | Line 42, called after store creation, before TUI program creation. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `main.go` | `recurring.AutoCreate` | Direct function call | WIRED | Line 42: `recurring.AutoCreate(s)` with store argument |
| `recurring.AutoCreate` | `store.TodoStore` | Interface methods | WIRED | Calls `ListSchedules`, `FindTemplate`, `TodoExistsForSchedule`, `AddScheduledTodo` |
| `recurring.AutoCreate` | `tmpl.ExecuteTemplate` | Function call | WIRED | Line 42 of generate.go: fills placeholders from parsed defaults |
| `tmplmgr.updateScheduleMode` | `store.TodoStore` | Schedule CRUD | WIRED | Calls `ListSchedulesForTemplate`, `AddSchedule`, `UpdateSchedule`, `DeleteSchedule` |
| `tmplmgr.updatePlaceholderDefaultsMode` | `store.TodoStore` | Schedule save | WIRED | Lines 483-486: saves schedule with JSON defaults via `AddSchedule` or `UpdateSchedule` |
| `tmplmgr.View` | `scheduleLabel` | Function call | WIRED | Line 527: calls for each template, result styled with `ScheduleSuffix` |
| `todolist.renderTodo` | `Todo.ScheduleID` | Field access | WIRED | Line 1067: checks `t.ScheduleID > 0` to render `[R]` |
| `store.SQLiteStore` | DB schema | SQL migrations | WIRED | schedules table (migration 4), schedule_id/schedule_date columns + unique dedup index (migration 5) |

### Requirements Coverage

| Requirement | Status | Evidence |
|-------------|--------|----------|
| REQ-29: Schedule picker UI | SATISFIED | scheduleMode with cadence cycling, weekday toggle, monthly input, Enter/Esc handling |
| REQ-30: Schedule display in template list | SATISFIED | `scheduleLabel()` renders "(daily)", "(Mon/Wed/Fri)", "(15th of month)" with muted style |
| REQ-31: Auto-create on app launch | SATISFIED | `recurring.AutoCreate(s)` in main.go before TUI, 7-day window, dedup via unique index |
| REQ-32: Recurring todo visual indicator [R] | SATISFIED | `renderTodo()` checks `ScheduleID > 0`, renders "[R]" with `RecurringIndicator` (MutedFg) |
| REQ-33: Placeholder defaults at schedule creation | SATISFIED | `placeholderDefaultsMode` prompts per placeholder, stores JSON, pre-fills on edit |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | - | - | - | No anti-patterns found in production code |

### Human Verification Required

### 1. Schedule picker visual flow
**Test:** Open template overlay (T), create a template, press S, cycle through None/Daily/Weekdays/Weekly/Monthly with left/right arrows. In weekly mode, use j/k to move between days and space to toggle. In monthly mode, type a number.
**Expected:** Cadence type bar shows active type highlighted with accent color, inactive in muted. Weekly shows checkboxes with cursor. Monthly shows text input. Enter saves, Esc cancels.
**Why human:** Visual layout, cursor movement feel, and input responsiveness cannot be verified programmatically.

### 2. Schedule suffix display
**Test:** After attaching a schedule to a template, return to list mode in the overlay.
**Expected:** Template name is followed by a dimmed suffix like "(daily)" or "(Mon/Wed/Fri)" or "(15th of month)".
**Why human:** Verify actual visual dimming and suffix formatting in context.

### 3. Recurring [R] indicator on auto-created todos
**Test:** Create a template with a daily schedule, restart the app, view the todo list for today.
**Expected:** Auto-created todos show "[R]" after the text in a muted color, distinct from the body indicator "[+]".
**Why human:** Verify visual distinction between indicators and that muted styling is clearly visible.

### 4. Placeholder defaults prompting flow
**Test:** Create a template with `{{.Project}}` and `{{.Owner}}` placeholders. Press S to schedule it. After choosing cadence, fill in default values for each placeholder.
**Expected:** Each placeholder gets its own input step showing "Set default for 'Project' (1/2):", then "Set default for 'Owner' (2/2):". Values are remembered when editing the schedule later.
**Why human:** Multi-step input flow and pre-fill behavior need interactive testing.

### 5. Auto-creation deduplication across restarts
**Test:** Start the app (triggers AutoCreate), verify todos are created. Close and restart the app.
**Expected:** No duplicate todos are created on second launch.
**Why human:** Requires actual app restart to verify runtime deduplication behavior.

### Gaps Summary

No gaps found. All 5 success criteria are met with substantive, wired implementations backed by comprehensive test coverage (45 tests across recurring and store packages). The AutoCreate engine correctly generates scheduled todos in a 7-day rolling window with deduplication. The schedule picker UI supports all cadence types with appropriate input modes. The [R] indicator renders for todos with a non-zero ScheduleID. Placeholder defaults are prompted during schedule creation and used during auto-creation.

---

_Verified: 2026-02-07T15:30:00Z_
_Verifier: Claude (gsd-verifier)_
