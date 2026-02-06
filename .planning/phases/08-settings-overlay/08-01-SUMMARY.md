---
phase: 08-settings-overlay
plan: 01
subsystem: settings
tags: [settings, config, theme, overlay, bubbletea]

dependency_graph:
  requires: [06-theming]
  provides: [settings-model, config-save, runtime-reconfiguration]
  affects: [08-02]

tech_stack:
  added: []
  patterns: [cycling-options-model, atomic-config-write, runtime-style-replacement]

file_tracking:
  key_files:
    created:
      - internal/settings/model.go
      - internal/settings/keys.go
      - internal/settings/styles.go
    modified:
      - internal/config/config.go
      - internal/theme/theme.go
      - internal/calendar/model.go
      - internal/todolist/model.go

decisions:
  - Settings model uses cycling options (not free-text input) for all 3 fields
  - countryLabels uses hardcoded map of 11 entries (no external ISO library)
  - SetTheme methods are pointer receivers to modify in place without recreation

metrics:
  duration: 2 min
  completed: 2026-02-06
---

# Phase 08 Plan 01: Settings Infrastructure Summary

Settings overlay component and supporting infrastructure: Config.Save() for atomic TOML write-back, theme.Names() for theme enumeration, settings Model with 3 cycling options emitting ThemeChangedMsg/SaveMsg/CancelMsg, and SetTheme/SetProvider/SetMondayStart on calendar/todolist for runtime reconfiguration.

## Task Commits

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Add Config.Save() and theme.Names() | 8c2b3c6 | internal/config/config.go, internal/theme/theme.go |
| 2 | Create settings package | bfc2f17 | internal/settings/model.go, keys.go, styles.go |
| 3 | Add SetTheme/SetProvider/SetMondayStart | c7fff1c | internal/calendar/model.go, internal/todolist/model.go |

## What Was Built

### Config.Save() (internal/config/config.go)
- Atomic TOML write using temp-file-rename pattern
- Creates config directory with os.MkdirAll on first use
- Proper error handling: close and remove temp file on any failure

### theme.Names() (internal/theme/theme.go)
- Returns `["dark", "light", "nord", "solarized"]`
- Single source of truth for theme enumeration
- Settings model imports this rather than hardcoding

### Settings Package (internal/settings/)
- **model.go**: Model with 3 options (theme, country, first day of week), cycling with wrapping
  - `New(cfg, theme)` initializes from current config with indexOf matching
  - `Config()` returns config.Config from current selections
  - `Update()` handles j/k navigation, h/l cycling, enter save, esc cancel
  - `View()` renders centered overlay with selected/unselected row styling
  - Emits `ThemeChangedMsg` on theme cycle for live preview
  - Emits `SaveMsg` with config on enter, `CancelMsg` on escape
  - Country display: "XX - Country Name" format via hardcoded 11-entry map
- **keys.go**: KeyMap with Up/Down/Left/Right/Save/Cancel, implements help.KeyMap
- **styles.go**: Styles with Title/Label/Value/SelectedLabel/SelectedValue/Hint, built from theme

### Runtime Reconfiguration Methods
- `calendar.Model.SetTheme(t)` -- replaces styles in place
- `calendar.Model.SetProvider(p)` -- swaps provider and refreshes holidays
- `calendar.Model.SetMondayStart(v)` -- toggles week start
- `todolist.Model.SetTheme(t)` -- replaces styles in place
- All pointer receivers, preserving cursor/month/mode state

## Deviations from Plan

None -- plan executed exactly as written.

## Decisions Made

1. **Cycling options pattern**: All 3 settings use predefined value lists with left/right cycling and wraparound. No free-text input needed since all valid values are known.
2. **Hardcoded country map**: 11 entries mapping codes to names. Adding a dependency for 11 strings would be wasteful.
3. **Pointer receivers for Set methods**: SetTheme/SetProvider/SetMondayStart use pointer receivers to modify fields in place, avoiding model recreation and state loss.

## Verification

```
go build ./...  -- PASS
go vet ./...    -- PASS
go test ./...   -- PASS
```

## Next Phase Readiness

Plan 08-02 can proceed immediately. It will:
- Wire the settings Model into the app model
- Add showSettings flag with conditional routing
- Add "s" keybinding to open settings
- Handle ThemeChangedMsg/SaveMsg/CancelMsg in the app
- Use SetTheme/SetProvider/SetMondayStart for live preview and save

## Self-Check: PASSED
