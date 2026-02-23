# Requirements: Todo Calendar

**Defined:** 2026-02-23
**Core Value:** See your month at a glance — calendar with holidays and todos in one terminal screen

## v2.3 Requirements

Requirements for Polybar Status milestone. Each maps to roadmap phases.

### Status Subcommand

- [x] **BAR-01**: User can run `todo-calendar status` to write today's pending todo count to a state file and exit
- [x] **BAR-02**: Output format is `%{F#hex}ICON COUNT%{F-}` where hex color reflects highest priority among today's pending todos

### Output Behavior

- [x] **BAR-03**: State file contains empty string when zero pending todos today (Polybar hides module)

### TUI Integration

- [x] **BAR-04**: TUI updates state file on todo add, complete, delete, and edit operations
- [x] **BAR-05**: TUI writes initial state file on startup

## Future Requirements

### Extended Bar Support

- **BAR-06**: Configurable state file path in config.toml
- **BAR-07**: Waybar/i3bar JSON output format option
- **BAR-08**: Google Calendar event count in status output

## Out of Scope

| Feature | Reason |
|---------|--------|
| Desktop notifications (D-Bus/notify-send) | Active notifications are out of scope; passive status only |
| Systemd timer/service generation | User configures i3/Polybar manually |
| Click actions in Polybar | Keep it read-only for v2.3 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| BAR-01 | Phase 36 | Complete |
| BAR-02 | Phase 36 | Complete |
| BAR-03 | Phase 36 | Complete |
| BAR-04 | Phase 37 | Complete |
| BAR-05 | Phase 37 | Complete |

**Coverage:**
- v2.3 requirements: 5 total
- Mapped to phases: 5
- Unmapped: 0

---
*Requirements defined: 2026-02-23*
*Last updated: 2026-02-23 after roadmap creation*
