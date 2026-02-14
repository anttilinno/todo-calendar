---
phase: 33-oauth-offline-guard
verified: 2026-02-14T14:20:00Z
status: human_needed
score: 11/11 must-haves verified
human_verification:
  - test: "OAuth flow end-to-end (without credentials)"
    expected: "App launches normally, settings shows 'Not configured', no errors"
    why_human: "Visual verification of TUI state, no network calls"
  - test: "OAuth flow end-to-end (with credentials but no token)"
    expected: "Settings shows 'Sign in', Enter opens browser, consent page appears, callback succeeds, settings shows 'Connected', token file created with 0600 permissions"
    why_human: "Browser interaction, user consent flow, visual feedback in TUI"
  - test: "Token persistence across restarts"
    expected: "After successful auth, restart app, settings immediately shows 'Connected' without re-authentication"
    why_human: "Multi-session state verification"
  - test: "Token auto-refresh"
    expected: "Token refreshes transparently when expired (requires waiting for token expiry or manual token manipulation)"
    why_human: "Time-dependent behavior, requires token expiry simulation"
---

# Phase 33: OAuth & Offline Guard Verification Report

**Phase Goal:** Users can authenticate with Google and the app handles unconfigured/offline states gracefully

**Verified:** 2026-02-14T14:20:00Z

**Status:** human_needed

**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| **Plan 01** |
| 1 | OAuth config loads from credentials.json in config dir | ✓ VERIFIED | `loadConfig()` reads from `CredentialsPath()`, calls `google.ConfigFromJSON()` with calendar scope (auth.go:96-106) |
| 2 | Token persists as JSON at google-token.json with 0600 permissions | ✓ VERIFIED | `saveToken()` uses atomic write with `os.Chmod(tmpName, 0600)` (auth.go:110-147), verified by `TestSaveLoadToken` checking file permissions (auth_test.go:29-35) |
| 3 | Token auto-refreshes and persisted token updates on refresh | ✓ VERIFIED | `persistingTokenSource` wraps `oauth2.TokenSource`, compares token hash, saves if changed (auth.go:163-195) |
| 4 | Auth state is correctly detected (NotConfigured, NeedsLogin, Ready) | ✓ VERIFIED | `checkAuthStateAt()` checks file existence, returns correct state (auth.go:83-91), verified by 3 unit tests (auth_test.go:75-119) |
| 5 | Auth flow opens browser, receives callback on loopback, exchanges code for token | ✓ VERIFIED | `performAuthFlow()` listens on 127.0.0.1:0, opens browser via `openBrowser()`, HTTP handler on /callback extracts code, exchanges via `cfg.Exchange()` with PKCE verifier (auth.go:233-293) |
| 6 | App does not error when credentials.json is absent | ✓ VERIFIED | `CheckAuthState()` returns `AuthNotConfigured` when credentials missing (auth.go:70-80), no panics or hard errors |
| **Plan 02** |
| 7 | Settings overlay shows Google Calendar row with status (Not configured / Sign in / Connected) | ✓ VERIFIED | Settings model has Google Calendar row at index 6, display driven by `googleStatusDisplay()` function (model.go:58-69, 106) |
| 8 | Pressing Enter on Google Calendar row when status is Sign in or Reconnect triggers OAuth flow | ✓ VERIFIED | Enter key handler checks cursor == googleCalendarRow and state is NeedsLogin or Revoked, emits `StartGoogleAuthMsg` (model.go:191-198) |
| 9 | Auth flow result updates settings status to Connected on success | ✓ VERIFIED | `app.Model.Update()` handles `google.AuthResultMsg`, sets state to `AuthReady` on success, calls `m.settings.SetGoogleAuthState(google.AuthReady)` (app/model.go:147-155) |
| 10 | App launches normally when credentials.json is absent (no errors, no prompts) | ✓ VERIFIED | `main.go` calls `google.CheckAuthState()` which returns `AuthNotConfigured` when missing, app passes this to `app.New()`, no error paths triggered (main.go:45) |
| 11 | App model passes AuthResultMsg to settings and updates Google auth state | ✓ VERIFIED | `app.Model.Update()` has case for `google.AuthResultMsg`, updates `m.googleAuthState` and calls `m.settings.SetGoogleAuthState()` (app/model.go:147-155) |

**Score:** 11/11 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/google/auth.go` | OAuth config, token persistence, auth flow, AuthState, AuthResultMsg | ✓ VERIFIED | 349 lines, exports: AuthState (4 constants), CheckAuthState, TokenSource, StartAuthFlow, AuthResultMsg, CredentialsPath, TokenPath |
| `internal/google/auth_test.go` | Unit tests for token persistence, config loading, auth state detection | ✓ VERIFIED | 134 lines, 6 tests: TestSaveLoadToken, TestSaveTokenAtomicDir, TestCheckAuthState_{NotConfigured,NeedsLogin,Ready}, TestLoadConfig_InvalidJSON |
| `internal/settings/model.go` | Google Calendar option row with auth state display and Enter trigger | ✓ VERIFIED | Google Calendar row at index 6 (line 106), Enter handler (lines 191-198), `googleStatusDisplay()` function (lines 58-69), imports `internal/google` |
| `internal/app/model.go` | AuthResultMsg handling, google auth state tracking | ✓ VERIFIED | `googleAuthState google.AuthState` field (line 80), `google.AuthResultMsg` case (lines 147-155), `settings.StartGoogleAuthMsg` triggers `google.StartAuthFlow()` (line 158) |
| `main.go` | Google auth state checked at startup and passed to app.New | ✓ VERIFIED | `authState := google.CheckAuthState()` (line 45), passed to `app.New()` (line 48) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `internal/google/auth.go` | `golang.org/x/oauth2` | oauth2.Config, oauth2.Token, oauth2.TokenSource | ✓ WIRED | Import on line 16, used throughout: oauth2.Config (line 96), oauth2.Token (lines 111, 150), oauth2.TokenSource (lines 166, 199), oauth2.GenerateVerifier (line 241), oauth2.S256ChallengeOption (line 244) |
| `internal/google/auth.go` | `~/.config/todo-calendar/credentials.json` | google.ConfigFromJSON | ✓ WIRED | `CredentialsPath()` returns path (lines 34-41), `loadConfig()` reads file and calls `google.ConfigFromJSON(b, calendarScope)` (lines 97-101) |
| `internal/google/auth.go` | `~/.config/todo-calendar/google-token.json` | saveToken/loadToken | ✓ WIRED | `TokenPath()` returns path (lines 44-50), `saveToken()` writes atomically (lines 111-147), `loadToken()` reads (lines 150-160), used in `TokenSource()` and `StartAuthFlow()` |
| `internal/settings/model.go` | `internal/google/auth.go` | google.AuthState, google.StartAuthFlow | ✓ WIRED | Import on line 12, `google.AuthState` type used (lines 40, 53, 72, 128), `google.AuthNeedsLogin/AuthReady/AuthRevoked` constants used (lines 60-65, 193) |
| `internal/app/model.go` | `internal/google/auth.go` | google.AuthResultMsg handling | ✓ WIRED | Import on line 12, `google.AuthState` field (line 80), `google.AuthResultMsg` case (line 147), `google.AuthReady/AuthNeedsLogin` constants (lines 149, 152), `google.StartAuthFlow()` call (line 158) |
| `main.go` | `internal/google/auth.go` | google.CheckAuthState at startup | ✓ WIRED | Import on line 9, `google.CheckAuthState()` called (line 45), result passed to `app.New()` (line 48) |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
|-------------|--------|----------------|
| AUTH-01: User can authenticate with Google via OAuth 2.0 loopback redirect flow | ✓ SATISFIED | Truth 5 verified: `performAuthFlow()` implements loopback on 127.0.0.1:0 with PKCE, opens browser, receives callback, exchanges code for token |
| AUTH-02: OAuth refresh token persisted to disk (0600 permissions) | ✓ SATISFIED | Truth 2 verified: `saveToken()` uses atomic write with 0600 permissions, unit test verifies permissions |
| AUTH-03: Token auto-refreshes transparently | ✓ SATISFIED | Truth 3 verified: `persistingTokenSource` wraps oauth2.TokenSource, auto-saves on refresh via hash comparison |
| AUTH-04: App works fully offline when Google not configured | ✓ SATISFIED | Truths 6 & 10 verified: `CheckAuthState()` returns NotConfigured when credentials absent, app launches cleanly without errors |
| CONF-02: OAuth setup flow triggered from settings | ✓ SATISFIED | Truths 7 & 8 verified: Google Calendar row in settings, Enter key triggers `StartGoogleAuthMsg` which dispatches `google.StartAuthFlow()` |

### Anti-Patterns Found

No anti-patterns detected. Scanned:
- `internal/google/auth.go` - No TODO/FIXME/placeholder comments, no empty implementations, no console.log stubs
- `internal/google/auth_test.go` - 6 substantive tests, all passing
- `internal/settings/model.go` - No placeholders, complete auth state handling
- `internal/app/model.go` - No placeholders, complete message routing
- `main.go` - Clean startup sequence

### Human Verification Required

#### 1. OAuth flow end-to-end (unconfigured state)

**Test:** Launch app without credentials.json at `~/.config/todo-calendar/credentials.json`. Press `s` to open settings. Observe Google Calendar row.

**Expected:** 
- App launches normally without errors or prompts
- Settings overlay opens
- Google Calendar row shows "Not configured" (dimmed, non-interactive)
- Left/Right arrow keys do nothing on this row
- Esc closes settings

**Why human:** Visual verification of TUI rendering and interaction behavior. CheckAuthState is file-based, no network calls to verify programmatically.

#### 2. OAuth flow end-to-end (configured state, first auth)

**Test:** 
1. Create GCP project, enable Calendar API, create OAuth 2.0 Desktop credentials
2. Download credentials.json to `~/.config/todo-calendar/credentials.json`
3. Launch app, press `s`, observe Google Calendar row shows "Sign in"
4. Press Enter on Google Calendar row
5. Observe browser opens to Google consent page
6. Complete sign-in in browser
7. Check settings updates to "Connected"
8. Verify `~/.config/todo-calendar/google-token.json` exists with 0600 permissions

**Expected:**
- Settings shows "Sign in" when credentials exist but no token
- Enter press changes display to "Waiting for browser..."
- Browser opens to Google OAuth consent page
- After consent, browser shows "Authorization successful! You can close this tab."
- Settings updates to "Connected"
- Token file created at `~/.config/todo-calendar/google-token.json` with `-rw-------` (0600) permissions

**Why human:** Requires external browser interaction, Google OAuth consent UI, visual feedback in TUI. Cannot simulate browser interaction programmatically.

#### 3. Token persistence across restarts

**Test:**
1. Complete OAuth flow (test 2)
2. Quit app (Ctrl+C)
3. Restart app
4. Press `s` to open settings
5. Observe Google Calendar row

**Expected:**
- App launches without re-authentication
- Settings shows "Connected" immediately (no "Sign in" state)
- No browser opens, no network delay

**Why human:** Multi-session verification. Requires app restart and state observation across process boundaries.

#### 4. Token auto-refresh transparency

**Test:**
1. Complete OAuth flow (test 2)
2. Manually edit `~/.config/todo-calendar/google-token.json`, set expiry to past date
3. Launch app (or trigger token use if app already running)
4. Observe no re-authentication prompt

**Expected:**
- Token refreshes transparently via oauth2.TokenSource
- Settings remains "Connected"
- No user prompts, no errors
- Token file updated with new expiry (persistingTokenSource saves on hash change)

**Why human:** Requires token manipulation and time-dependent behavior. OAuth refresh flow involves network calls to Google that can't be easily mocked without running app.

---

## Summary

**All automated checks passed.** Phase 33 goal is ACHIEVED pending human verification of the OAuth flow UX.

- **11/11 observable truths verified**
- **All artifacts exist and are substantive** (no stubs, no placeholders)
- **All key links wired** (imports present, functions called, data flows)
- **All 5 requirements satisfied** (AUTH-01 through AUTH-04, CONF-02)
- **No anti-patterns found**
- **Commits verified:** cdfe03b, fa40548, 4108e5d

The implementation is complete and functional. The only remaining verification is **human testing** of the OAuth flow UX and browser interaction, which cannot be automated. All success criteria from the ROADMAP are technically satisfied:

1. ✓ User can complete OAuth flow (browser opens, user signs in, app receives token)
2. ✓ App remembers authentication across restarts (token persisted securely with 0600)
3. ✓ Token refreshes transparently (persistingTokenSource wrapper)
4. ✓ App launches fully when Google not configured (no errors, no prompts)
5. ✓ OAuth setup triggered from settings overlay (Google Calendar row + Enter key)

---

_Verified: 2026-02-14T14:20:00Z_
_Verifier: Claude (gsd-verifier)_
