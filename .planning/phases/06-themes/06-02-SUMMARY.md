---
phase: "06-themes"
plan: "02"
title: "Wire Theme Through Application"
status: "complete"
subsystem: "theming"
tags: ["lipgloss", "theme", "styles", "wiring"]
dependency-graph:
  requires: ["06-01"]
  provides: ["full-theme-support"]
  affects: []
tech-stack:
  added: []
  patterns: ["struct-based styles", "constructor injection", "theme propagation"]
key-files:
  created: []
  modified:
    - "internal/calendar/styles.go"
    - "internal/calendar/grid.go"
    - "internal/calendar/model.go"
    - "internal/todolist/styles.go"
    - "internal/todolist/model.go"
    - "internal/app/styles.go"
    - "internal/app/model.go"
    - "main.go"
decisions:
  - id: "styles-as-struct"
    description: "All three styles.go files converted from package-level vars to Styles struct + NewStyles(theme.Theme) constructor"
  - id: "pane-method"
    description: "paneStyle() free function replaced with Styles.Pane(focused) method receiver"
  - id: "help-bar-themed"
    description: "Help bar ShortKey uses AccentFg, ShortDesc/ShortSeparator use MutedFg"
metrics:
  duration: "3 min"
  completed: "2026-02-05"
---

# Phase 6 Plan 02: Wire Theme Through Application Summary

Theme constructor injection from config.toml through main.go to all three packages, replacing all package-level style vars with struct-based styles built from Theme.

## What Was Done

### Task 1: Convert calendar and todolist styles to Styles structs (27cd47b)

**Calendar package:**
- `styles.go`: Replaced 6 package-level `var` declarations with `Styles` struct and `NewStyles(theme.Theme)` constructor
- `grid.go`: `RenderGrid` now accepts `Styles` parameter; all bare style references replaced with `s.Header`, `s.Today`, etc.
- `model.go`: Added `styles Styles` field; `New()` accepts `theme.Theme`, calls `NewStyles(t)`

**Todolist package:**
- `styles.go`: Replaced 5 package-level `var` declarations with `Styles` struct and `NewStyles(theme.Theme)` constructor
- `model.go`: Added `styles Styles` field; `New()` accepts `theme.Theme`; `View()` and `renderTodo()` use `m.styles.*`

Critical pitfalls avoided:
- No `.Reverse(true)` on Today style (uses explicit Foreground/Background)
- No `.Faint(true)` anywhere (uses explicit foreground colors from theme)

### Task 2: Wire theme through app layer and main.go (5e45736)

**App package:**
- `styles.go`: Replaced `focusedStyle`/`unfocusedStyle` vars and `paneStyle()` function with `Styles` struct, `NewStyles(theme.Theme)`, and `Styles.Pane(focused)` method
- `model.go`: Added `styles Styles` field; `New()` accepts `theme.Theme`, passes to `calendar.New` and `todolist.New`; help bar themed with `AccentFg` and `MutedFg`

**main.go:**
- Imports `theme` package
- Resolves theme: `t := theme.ForName(cfg.Theme)`
- Passes to app: `app.New(provider, cfg.MondayStart(), s, t)`

## Deviations from Plan

None -- plan executed exactly as written.

## Verification Results

- `go build .` -- passes
- `go vet ./...` -- passes clean
- No package-level `var` style declarations remain in any `styles.go`
- No `.Reverse(true)` in calendar styles
- No `.Faint(true)` in any styles file
- `theme.ForName` called in `main.go` line 40
- Theme flows: `main.go` -> `app.New` -> `calendar.New` + `todolist.New`

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 27cd47b | Convert calendar and todolist styles to themed Styles structs |
| 2 | 5e45736 | Wire theme through app layer and main.go, theme help bar |
