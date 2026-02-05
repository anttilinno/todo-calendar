# Roadmap: Todo Calendar

## Overview

Deliver a terminal-based calendar+todo application in three phases: first establish the split-pane TUI scaffold with keyboard navigation, then render the calendar with holidays, then wire up todo management with persistence. Each phase delivers a runnable application with progressively more capability, building on the Bubble Tea Elm Architecture from the first commit.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: TUI Scaffold** - Runnable split-pane Bubble Tea app with keyboard navigation and resize handling
- [x] **Phase 2: Calendar + Holidays** - Monthly calendar grid with navigation, today highlight, holiday display, and country config
- [ ] **Phase 3: Todo Management** - Todo CRUD, date-bound and floating display, persistence, and help bar

## Phase Details

### Phase 1: TUI Scaffold
**Goal**: User launches the app and sees a responsive split-pane terminal layout they can navigate with keyboard
**Depends on**: Nothing (first phase)
**Requirements**: UI-01, UI-02, UI-04
**Success Criteria** (what must be TRUE):
  1. User runs the binary and sees a split-pane layout with distinct left and right panels
  2. User can switch focus between panes with Tab, and the focused pane has a visually distinct border
  3. User can resize the terminal and the layout adjusts without breaking or panicking
  4. User can quit the app with q or Ctrl+C
**Plans**: 1 plan

Plans:
- [x] 01-01-PLAN.md -- Go project init, split-pane scaffold with focus routing, keyboard nav, and resize handling

### Phase 2: Calendar + Holidays
**Goal**: User sees the current month's calendar with today highlighted and national holidays in red, and can navigate between months
**Depends on**: Phase 1
**Requirements**: CAL-01, CAL-02, CAL-03, CAL-04, CAL-05, DATA-02
**Success Criteria** (what must be TRUE):
  1. Left pane displays a monthly calendar grid with day-of-week headers resembling `cal` output
  2. Today's date is visually highlighted on the calendar
  3. User can navigate to the next or previous month and the calendar updates immediately
  4. National holidays appear in red on the calendar, sourced from a configurable country setting stored in a TOML config file
**Plans**: 2 plans

Plans:
- [x] 02-01-PLAN.md -- Config package, holiday provider/registry, calendar grid renderer with styles and key bindings
- [x] 02-02-PLAN.md -- Wire calendar model, app integration, month navigation, and visual verification

### Phase 3: Todo Management
**Goal**: User can manage todos with optional dates, see them organized by month and floating section, with all data persisted to disk
**Depends on**: Phase 2
**Requirements**: TODO-01, TODO-02, TODO-03, TODO-04, TODO-05, DATA-01, DATA-03, UI-03
**Success Criteria** (what must be TRUE):
  1. User can add a new todo with text and an optional date via inline text input
  2. User can mark a todo as complete (visual indicator) and delete a todo
  3. Right pane shows date-bound todos for the currently viewed month and a separate section for floating (undated) todos
  4. Todos persist across app restarts (stored as JSON in XDG-compliant path)
  5. A help bar at the bottom shows available keybindings for the current context
**Plans**: TBD

Plans:
- [ ] 03-01: Todo data model, JSON store with atomic writes, XDG paths
- [ ] 03-02: Todo list UI with CRUD, date filtering, floating section, and help bar

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. TUI Scaffold | 1/1 | Complete | 2026-02-05 |
| 2. Calendar + Holidays | 2/2 | Complete | 2026-02-05 |
| 3. Todo Management | 0/2 | Not started | - |
