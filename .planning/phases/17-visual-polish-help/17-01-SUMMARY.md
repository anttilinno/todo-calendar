---
phase: 17-visual-polish-help
plan: 01
subsystem: todolist-rendering
tags: [lipgloss, styles, visual-polish, checkboxes, separators]
depends_on: []
provides: [styled-checkboxes, section-separators, item-spacing]
affects: [17-02]
tech_stack:
  added: []
  patterns: [separated-checkbox-styling, section-separator-lines]
key_files:
  created: []
  modified:
    - internal/todolist/styles.go
    - internal/todolist/model.go
decisions:
  - id: VIS-checkbox-styling
    description: "Checkbox styled independently from text: accent for [ ], green for [x], strikethrough only on text"
  - id: VIS-separator-width
    description: "Fixed 10-char Unicode box-drawing separator under headers, not dynamic width"
metrics:
  duration: "1 min"
  completed: "2026-02-07"
---

# Phase 17 Plan 01: Todo List Visual Polish Summary

**One-liner:** Styled checkboxes (accent/green), separator lines under section headers, and blank-line spacing between todo items.

## What Was Done

### Task 1: Add Separator, Checkbox, CheckboxDone styles
Added three new style fields to the `Styles` struct in `styles.go`:
- `Separator` -- `MutedFg` foreground for thin horizontal rules under section headers
- `Checkbox` -- `AccentFg` foreground for unchecked `[ ]`
- `CheckboxDone` -- `CompletedCountFg` foreground for checked `[x]`

All three use theme colors already defined in all 4 themes (Dark, Light, Nord, Solarized). No theme changes needed.

### Task 2: Update View and renderTodo for visual polish
Three rendering changes in `model.go`:

1. **VIS-01 (Breathing room):** Added `b.WriteString("\n")` after each `todoItem` in `View()`, producing a blank line between todo items.

2. **VIS-02 (Section separators):** Added `m.styles.Separator.Render("──────────")` after each section header, plus `b.WriteString("\n")` before non-first headers (detected via `b.Len() > 0`).

3. **VIS-03 (Styled checkboxes):** Rewrote `renderTodo()` to style checkbox separately from text. Unchecked uses `m.styles.Checkbox`, checked uses `m.styles.CheckboxDone`. `m.styles.Completed` (strikethrough) is applied only to the text string, not the checkbox. Body indicator `[+]` and date remain outside completed styling.

## Task Commits

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Add Separator, Checkbox, CheckboxDone styles | 77f6f57 | internal/todolist/styles.go |
| 2 | Update View and renderTodo for visual polish | e660392 | internal/todolist/model.go |

## Decisions Made

| ID | Decision | Rationale |
|----|----------|-----------|
| VIS-checkbox-styling | Checkbox styled independently from text | Prevents strikethrough from bleeding into checkbox characters |
| VIS-separator-width | Fixed 10-char separator line | Simple, predictable, avoids width-calculation complexity |

## Deviations from Plan

None -- plan executed exactly as written.

## Verification

- `go build ./...` -- passes
- `go vet ./...` -- passes
- All 4 themes use the same color fields (AccentFg, CompletedCountFg, MutedFg) so all render correctly

## Next Phase Readiness

No blockers. Plan 17-02 can proceed independently.

## Self-Check: PASSED
