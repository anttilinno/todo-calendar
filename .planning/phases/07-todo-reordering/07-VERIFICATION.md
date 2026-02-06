---
phase: 07-todo-reordering
verified: 2026-02-06T07:50:43Z
status: passed
score: 10/10 must-haves verified
---

# Phase 7: Todo Reordering Verification Report

**Phase Goal:** Users can arrange todos in their preferred order
**Verified:** 2026-02-06T07:50:43Z
**Status:** PASSED
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can move the selected todo one position up in the list via keybinding | ✓ VERIFIED | MoveUp handler (Shift+K) in model.go:195-206 calls SwapOrder, decrements cursor |
| 2 | User can move the selected todo one position down in the list via keybinding | ✓ VERIFIED | MoveDown handler (Shift+J) in model.go:208-219 calls SwapOrder, increments cursor |
| 3 | Custom todo order survives app restart (order is persisted in JSON) | ✓ VERIFIED | SortOrder field with json tag in todo.go:17, Save() called in SwapOrder store.go:249 |
| 4 | Reorder keybindings appear in the help bar when a todo is selected | ✓ VERIFIED | MoveUp/MoveDown in HelpBindings() model.go:101, ShortHelp keys.go:23, FullHelp keys.go:29 |

**Score:** 4/4 truths verified

### Plan 07-01 Must-Haves

#### Truths from Plan 07-01

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Todo struct has a SortOrder field that persists to JSON | ✓ VERIFIED | `SortOrder int` with `json:"sort_order,omitempty"` tag at todo.go:17 |
| 2 | Legacy todos without sort_order get unique values on load | ✓ VERIFIED | EnsureSortOrder() assigns (i+1)*10 at store.go:222-232, called in NewStore at store.go:34 |
| 3 | Dated todos sort by SortOrder first, then date, then ID | ✓ VERIFIED | TodosForMonth sort logic at store.go:173-181 |
| 4 | Floating todos sort by SortOrder first, then ID | ✓ VERIFIED | FloatingTodos sort logic at store.go:211-216 |
| 5 | New todos get a SortOrder placing them at end of list | ✓ VERIFIED | Add() assigns maxOrder+10 at store.go:93-105 |
| 6 | SwapOrder exchanges sort order of two todos and persists | ✓ VERIFIED | SwapOrder method at store.go:237-251, calls Save() at store.go:249 |

**Score:** 6/6 truths verified

#### Artifacts from Plan 07-01

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/store/todo.go` | SortOrder field on Todo struct | ✓ VERIFIED | Line 17: `SortOrder int` with correct JSON tag |
| `internal/store/store.go` | EnsureSortOrder, SwapOrder methods, updated sort logic, updated Add | ✓ VERIFIED | EnsureSortOrder (222-232), SwapOrder (237-251), Add assigns SortOrder (105), sort logic updated in TodosForMonth (173-181) and FloatingTodos (211-216) |

#### Key Links from Plan 07-01

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| internal/store/store.go | internal/store/todo.go | SortOrder field used in sort comparators | ✓ WIRED | TodosForMonth and FloatingTodos both check `result[i].SortOrder != result[j].SortOrder` |
| internal/store/store.go (NewStore) | internal/store/store.go (EnsureSortOrder) | called after load to migrate legacy data | ✓ WIRED | Line 34 calls `s.EnsureSortOrder()` after successful load |
| internal/store/store.go (Add) | internal/store/todo.go (SortOrder) | new todos get maxOrder + 10 | ✓ WIRED | Lines 93-105 find maxOrder and assign `SortOrder: maxOrder + 10` |

### Plan 07-02 Must-Haves

#### Truths from Plan 07-02

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can move the selected todo one position up via Shift+K | ✓ VERIFIED | MoveUp handler at model.go:195-206, keybinding "K" at keys.go:44-47 |
| 2 | User can move the selected todo one position down via Shift+J | ✓ VERIFIED | MoveDown handler at model.go:208-219, keybinding "J" at keys.go:48-51 |
| 3 | Move is a no-op at section boundaries (dated todo cannot swap with floating) | ✓ VERIFIED | Section boundary check: `curTodo.HasDate() == prevTodo.HasDate()` at model.go:202 and model.go:215 |
| 4 | Cursor follows the moved todo after swap | ✓ VERIFIED | `m.cursor--` after MoveUp at model.go:204, `m.cursor++` after MoveDown at model.go:217 |
| 5 | Reorder keybindings K/J appear in help bar during normal mode | ✓ VERIFIED | HelpBindings includes MoveUp/MoveDown at model.go:101 |

**Score:** 5/5 truths verified

#### Artifacts from Plan 07-02

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/todolist/keys.go` | MoveUp and MoveDown key bindings | ✓ VERIFIED | MoveUp (line 9) and MoveDown (line 10) fields in KeyMap, initialized at lines 44-51, included in ShortHelp (23) and FullHelp (29) |
| `internal/todolist/model.go` | Move handlers in updateNormalMode, updated HelpBindings | ✓ VERIFIED | MoveUp handler (195-206), MoveDown handler (208-219), HelpBindings updated (101) |

#### Key Links from Plan 07-02

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| internal/todolist/model.go (updateNormalMode) | internal/store/store.go (SwapOrder) | m.store.SwapOrder call on move key press | ✓ WIRED | Line 203: `m.store.SwapOrder(curTodo.ID, prevTodo.ID)`, Line 216: `m.store.SwapOrder(curTodo.ID, nextTodo.ID)` |
| internal/todolist/model.go (updateNormalMode) | internal/store/todo.go (HasDate) | section boundary check before swap | ✓ WIRED | Lines 202 and 215 check `curTodo.HasDate() == prevTodo.HasDate()` before swapping |
| internal/todolist/model.go (HelpBindings) | internal/todolist/keys.go (MoveUp, MoveDown) | included in normal mode help bindings list | ✓ WIRED | Line 101 returns bindings including `m.keys.MoveUp, m.keys.MoveDown` |

### Requirements Coverage

| Requirement | Status | Supporting Truths |
|-------------|--------|-------------------|
| REORD-01: User can move a selected todo up in the list | ✓ SATISFIED | Truth #1 (phase-level), Truth #1 (plan 07-02) |
| REORD-02: User can move a selected todo down in the list | ✓ SATISFIED | Truth #2 (phase-level), Truth #2 (plan 07-02) |
| REORD-03: Custom order persists across app restarts | ✓ SATISFIED | Truth #3 (phase-level), Truth #1, #2, #6 (plan 07-01) |

### Anti-Patterns Found

No blocker anti-patterns detected.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | - | - | - | - |

### Build & Compilation Status

- ✓ `go build ./...` — passes without errors
- ✓ `go vet ./...` — no warnings
- ✓ All artifacts exist and are substantive
- ✓ All key links are wired correctly

### Implementation Quality

**Substantive Evidence:**
- `internal/store/todo.go`: 44 lines (well above 15-line component minimum)
- `internal/store/store.go`: 252 lines (well above 10-line route minimum)
- `internal/todolist/keys.go`: 86 lines (well above 10-line util minimum)
- `internal/todolist/model.go`: 495 lines (well above 15-line component minimum)

**No Stubs Found:**
- All methods have complete implementations
- No TODO/FIXME comments in critical paths
- No placeholder returns or console.log-only handlers
- SwapOrder properly finds todos by ID and persists changes
- Move handlers include proper bounds checking and section boundary enforcement

**Wiring Integrity:**
- SortOrder used in 2 sort functions (TodosForMonth, FloatingTodos)
- EnsureSortOrder called from NewStore (migration on load)
- SwapOrder called from 2 locations in model.go (MoveUp, MoveDown handlers)
- MoveUp/MoveDown keybindings included in help bar (ShortHelp, FullHelp, HelpBindings)

## Verification Details

### Level 1: Existence
All artifacts exist at expected paths:
- ✓ `internal/store/todo.go`
- ✓ `internal/store/store.go`
- ✓ `internal/todolist/keys.go`
- ✓ `internal/todolist/model.go`

### Level 2: Substantive
All artifacts contain real implementations:
- ✓ `todo.go`: SortOrder field with proper JSON tag
- ✓ `store.go`: Complete EnsureSortOrder migration, SwapOrder method, updated sort functions, updated Add
- ✓ `keys.go`: MoveUp/MoveDown bindings with proper help text
- ✓ `model.go`: Complete move handlers with bounds checks, section boundary enforcement, cursor adjustment

### Level 3: Wired
All critical connections verified:
- ✓ SortOrder field used in TodosForMonth and FloatingTodos sort comparators
- ✓ EnsureSortOrder called in NewStore after successful load
- ✓ Add assigns SortOrder to new todos
- ✓ SwapOrder called from MoveUp and MoveDown handlers
- ✓ HasDate used for section boundary checks
- ✓ MoveUp/MoveDown included in all help contexts

### Critical Path Verification

**Path 1: User presses Shift+K to move todo up**
1. ✓ Key "K" bound to MoveUp in DefaultKeyMap (keys.go:44-47)
2. ✓ MoveUp handler receives key event (model.go:195)
3. ✓ Handler checks cursor bounds (model.go:196)
4. ✓ Handler retrieves current and previous todo (model.go:197-200)
5. ✓ Handler validates section boundary (model.go:201-202)
6. ✓ Handler calls SwapOrder with todo IDs (model.go:203)
7. ✓ SwapOrder finds both todos by ID (store.go:239-245)
8. ✓ SwapOrder swaps SortOrder values (store.go:248)
9. ✓ SwapOrder calls Save to persist (store.go:249)
10. ✓ Handler decrements cursor to follow moved item (model.go:204)

**Path 2: User presses Shift+J to move todo down**
1. ✓ Key "J" bound to MoveDown in DefaultKeyMap (keys.go:48-51)
2. ✓ MoveDown handler receives key event (model.go:208)
3. ✓ Handler checks cursor bounds (model.go:209)
4. ✓ Handler retrieves current and next todo (model.go:210-213)
5. ✓ Handler validates section boundary (model.go:214-215)
6. ✓ Handler calls SwapOrder with todo IDs (model.go:216)
7. ✓ SwapOrder finds both todos by ID (store.go:239-245)
8. ✓ SwapOrder swaps SortOrder values (store.go:248)
9. ✓ SwapOrder calls Save to persist (store.go:249)
10. ✓ Handler increments cursor to follow moved item (model.go:217)

**Path 3: Order persists across restart**
1. ✓ SwapOrder calls Save which writes JSON (store.go:249 → store.go:54)
2. ✓ SortOrder has JSON tag "sort_order,omitempty" (todo.go:17)
3. ✓ NewStore calls load to read JSON (store.go:31)
4. ✓ NewStore calls EnsureSortOrder for legacy migration (store.go:34)
5. ✓ TodosForMonth and FloatingTodos sort by SortOrder (store.go:173-181, 211-216)
6. ✓ Sorted todos rendered in visibleItems (model.go:105-136)

**Path 4: Section boundary enforcement**
1. ✓ MoveUp checks `curTodo.HasDate() == prevTodo.HasDate()` (model.go:202)
2. ✓ MoveDown checks `curTodo.HasDate() == nextTodo.HasDate()` (model.go:215)
3. ✓ If mismatch, no SwapOrder call → no-op (model.go:201-206, 214-219)

**Path 5: Help bar integration**
1. ✓ MoveUp/MoveDown in KeyMap struct (keys.go:9-10)
2. ✓ Initialized with help text "move up"/"move down" (keys.go:46, 50)
3. ✓ Included in ShortHelp (keys.go:23)
4. ✓ Included in FullHelp (keys.go:29)
5. ✓ Returned from HelpBindings in normal mode (model.go:101)

## Human Verification Required

### 1. Visual Todo Reordering

**Test:** 
1. Run `go run .`
2. Add 3 floating todos: "Task A", "Task B", "Task C"
3. Select "Task B" (cursor on middle item)
4. Press Shift+J to move down
5. Press Shift+K twice to move to top
6. Quit (Ctrl+C) and restart app

**Expected:**
- After Shift+J: Order becomes "Task A", "Task C", "Task B"
- After Shift+K twice: Order becomes "Task B", "Task A", "Task C"
- After restart: Order remains "Task B", "Task A", "Task C"
- Cursor follows the moved item each time
- Help bar shows "K move up | J move down" in normal mode

**Why human:** Visual confirmation of UI behavior, cursor movement, and persistence

### 2. Section Boundary Enforcement

**Test:**
1. Run app
2. Add 2 dated todos for current month
3. Add 2 floating todos
4. Select last dated todo (cursor on bottom of dated section)
5. Press Shift+J to attempt move into floating section
6. Verify todo does NOT move
7. Select first floating todo
8. Press Shift+K to attempt move into dated section
9. Verify todo does NOT move

**Expected:**
- Move at section boundary is a no-op
- Cursor does not change
- Todo order unchanged

**Why human:** Need to verify boundary enforcement behavior visually

### 3. Legacy Data Migration

**Test:**
1. Create `~/.config/todo-calendar/todos.json` manually with old format (no sort_order):
```json
{
  "next_id": 3,
  "todos": [
    {"id": 1, "text": "Old task 1", "date": "", "done": false, "created_at": "2026-02-06"},
    {"id": 2, "text": "Old task 2", "date": "", "done": false, "created_at": "2026-02-06"}
  ]
}
```
2. Run app
3. Verify todos appear in order
4. Quit and check JSON file

**Expected:**
- App loads without error
- Todos displayed in original order
- JSON file now contains `"sort_order": 10` and `"sort_order": 20`

**Why human:** Need to verify migration runs and persists correctly

## Summary

Phase 7 goal **ACHIEVED**. All must-haves verified:

**Data Layer (Plan 07-01):**
- ✓ SortOrder field exists with proper JSON serialization
- ✓ EnsureSortOrder migration handles legacy data
- ✓ Sort functions use SortOrder as primary key
- ✓ New todos get SortOrder at end of list
- ✓ SwapOrder exchanges order and persists atomically

**UI Layer (Plan 07-02):**
- ✓ Shift+K moves todo up within section
- ✓ Shift+J moves todo down within section
- ✓ Section boundaries enforced (HasDate equality check)
- ✓ Cursor follows moved item
- ✓ Keybindings visible in help bar

**Requirements:**
- ✓ REORD-01: User can move todo up
- ✓ REORD-02: User can move todo down
- ✓ REORD-03: Order persists across restarts

**Code Quality:**
- ✓ Compiles without errors
- ✓ No vet warnings
- ✓ No stub patterns detected
- ✓ All critical paths fully wired
- ✓ Proper error handling and boundary checks

Phase 7 is ready for milestone v1.2. Proceed to Phase 8 (Settings Overlay).

---

_Verified: 2026-02-06T07:50:43Z_
_Verifier: Claude (gsd-verifier)_
