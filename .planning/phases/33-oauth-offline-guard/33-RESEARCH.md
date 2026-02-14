# Phase 33: OAuth & Offline Guard - Research

**Researched:** 2026-02-14
**Domain:** Google OAuth 2.0 for Go desktop/TUI apps, token persistence, graceful degradation
**Confidence:** HIGH

## Summary

This phase implements Google OAuth 2.0 authentication using the loopback redirect flow (RFC 8252) in a Go Bubble Tea TUI app. The standard approach uses `golang.org/x/oauth2` with `golang.org/x/oauth2/google` for endpoint configuration, a temporary local HTTP server to receive the auth callback, and JSON file persistence for the refresh token. The app already uses `os.UserConfigDir()` for paths and TOML config, so the token file fits naturally alongside existing data.

The key architectural challenge is integrating browser-based OAuth into a TUI that runs in alt-screen mode. The app already handles this pattern for external editors (using `tea.ExecProcess`), but OAuth is different: the browser opens in the background while the TUI waits for a callback on a local HTTP server. This means the TUI can remain running (no need to suspend) while the browser handles authentication, then a `tea.Cmd` delivers the result.

**Primary recommendation:** Use `golang.org/x/oauth2` + `golang.org/x/oauth2/google` with PKCE, loopback redirect on an ephemeral port, token persisted as JSON with 0600 permissions at `~/.config/todo-calendar/google-token.json`, and a new `internal/google/auth` package that exposes a clean interface consumed by the app model.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| golang.org/x/oauth2 | v0.35.0 | OAuth 2.0 client (Config, Token, TokenSource) | Official Go OAuth2 library, 46K+ importers |
| golang.org/x/oauth2/google | (same module) | Google endpoint URLs, ConfigFromJSON | Official Google-specific OAuth2 helpers |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | Token serialization to/from file | Token persistence |
| net/http | stdlib | Local loopback HTTP server for OAuth callback | Receive auth code |
| net | stdlib | Listen on ephemeral port (`:0`) | Dynamic port allocation |
| os/exec | stdlib | Open browser via `xdg-open` | Launch auth URL |
| crypto/rand, crypto/sha256 | stdlib | PKCE verifier/challenge generation | Security (built into x/oauth2) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| golang.org/x/oauth2 | Manual HTTP token exchange | No reason to hand-roll; x/oauth2 handles refresh, PKCE, token reuse |
| JSON token file | Encrypted token file | Overkill for single-user desktop app; 0600 permissions are standard practice (Google's own quickstart uses this) |
| google.golang.org/api/calendar/v3 | Raw REST API calls | Not needed this phase; Phase 34 will add the Calendar API client |

**Installation:**
```bash
go get golang.org/x/oauth2
```

Note: `golang.org/x/oauth2/google` is part of the same module. No separate `go get` needed.

## Architecture Patterns

### Recommended Project Structure
```
internal/
  google/
    auth.go           # OAuth config, token load/save, auth flow orchestration
    auth_test.go       # Unit tests for token persistence, config loading
```

This follows the project's existing pattern of small, focused packages under `internal/`. A single `google` package (not `google/auth`) keeps it simple for Phase 33 while leaving room for Phase 34 to add `client.go` (Calendar API) in the same package.

### Pattern 1: Credential Embedding
**What:** Embed OAuth client credentials (client ID + secret) in the binary or load from a JSON file in the config directory.
**When to use:** Desktop apps where the "secret" is not truly secret (Google acknowledges this for installed apps).
**Approach:**

Option A - Config file (recommended): User downloads `credentials.json` from Google Cloud Console and places it at `~/.config/todo-calendar/credentials.json`. The app reads it with `google.ConfigFromJSON()`.

Option B - Embedded: Compile client ID/secret into the binary. Simpler for end users but requires the developer to manage a GCP project.

**Recommendation:** Use Option A (config file). This is a personal-use app. The user creates their own GCP project and downloads credentials. This avoids distributing client secrets and avoids Google's OAuth app verification process.

```go
// Load OAuth config from credentials file
func loadConfig(credPath string) (*oauth2.Config, error) {
    b, err := os.ReadFile(credPath)
    if err != nil {
        return nil, fmt.Errorf("read credentials: %w", err)
    }
    config, err := google.ConfigFromJSON(b,
        "https://www.googleapis.com/auth/calendar.events.readonly",
    )
    if err != nil {
        return nil, fmt.Errorf("parse credentials: %w", err)
    }
    // Override redirect to loopback (ConfigFromJSON may set web redirect)
    config.RedirectURL = "" // Set dynamically when we know the port
    return config, nil
}
```

### Pattern 2: Loopback OAuth Flow with Ephemeral Port
**What:** Start a temporary HTTP server on 127.0.0.1 with an OS-assigned port, open the browser to Google's consent page, receive the auth code on the callback endpoint, exchange for tokens, shut down the server.
**When to use:** Desktop/CLI apps per RFC 8252.

```go
func performAuthFlow(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
    // 1. Listen on ephemeral port
    listener, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        return nil, err
    }
    port := listener.Addr().(*net.TCPAddr).Port
    config.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", port)

    // 2. Generate PKCE verifier
    verifier := oauth2.GenerateVerifier()

    // 3. Build auth URL
    authURL := config.AuthCodeURL("state",
        oauth2.AccessTypeOffline,
        oauth2.S256ChallengeOption(verifier),
    )

    // 4. Open browser
    exec.Command("xdg-open", authURL).Start()

    // 5. Wait for callback
    codeCh := make(chan string, 1)
    errCh := make(chan error, 1)

    mux := http.NewServeMux()
    mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        if code == "" {
            errCh <- fmt.Errorf("no code in callback")
            return
        }
        fmt.Fprintln(w, "Authentication successful! You can close this tab.")
        codeCh <- code
    })

    srv := &http.Server{Handler: mux}
    go srv.Serve(listener)
    defer srv.Shutdown(ctx)

    // 6. Exchange code for token
    select {
    case code := <-codeCh:
        return config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
    case err := <-errCh:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Pattern 3: Token Persistence with 0600 Permissions
**What:** Save/load `oauth2.Token` as JSON to a dedicated file.
**When to use:** Always, for AUTH-02.

```go
const tokenFileName = "google-token.json"

func tokenPath() (string, error) {
    dir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, "todo-calendar", tokenFileName), nil
}

func saveToken(path string, token *oauth2.Token) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        return err
    }
    defer f.Close()
    return json.NewEncoder(f).Encode(token)
}

func loadToken(path string) (*oauth2.Token, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    tok := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(tok)
    return tok, err
}
```

### Pattern 4: Transparent Token Refresh with Persistence
**What:** Wrap `oauth2.TokenSource` to persist refreshed tokens automatically.
**When to use:** For AUTH-03 (transparent refresh).

```go
// persistingTokenSource wraps a TokenSource and saves new tokens to disk.
type persistingTokenSource struct {
    src      oauth2.TokenSource
    path     string
    mu       sync.Mutex
    lastHash string
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
    tok, err := p.src.Token()
    if err != nil {
        return nil, err
    }
    // Save if token changed (i.e., was refreshed)
    hash := tok.AccessToken + tok.Expiry.String()
    p.mu.Lock()
    defer p.mu.Unlock()
    if hash != p.lastHash {
        p.lastHash = hash
        _ = saveToken(p.path, tok) // best-effort persist
    }
    return tok, err
}
```

### Pattern 5: Bubble Tea Integration
**What:** OAuth flow runs as a `tea.Cmd` that returns a message on completion.
**When to use:** Triggering OAuth from settings overlay or first-run.

```go
// AuthResultMsg carries the outcome of an OAuth flow.
type AuthResultMsg struct {
    Success bool
    Err     error
}

// StartAuthFlow returns a tea.Cmd that runs the OAuth flow.
func StartAuthFlow() tea.Cmd {
    return func() tea.Msg {
        // ... load config, run auth flow, save token ...
        tok, err := performAuthFlow(context.Background(), config)
        if err != nil {
            return AuthResultMsg{Err: err}
        }
        if err := saveToken(path, tok); err != nil {
            return AuthResultMsg{Err: err}
        }
        return AuthResultMsg{Success: true}
    }
}
```

### Pattern 6: Graceful Offline Guard (AUTH-04)
**What:** At startup, check if credentials and token exist. If not, the app runs normally with Google Calendar features disabled (no-op).
**When to use:** Always. This is the default state.

```go
// AuthState represents the current Google authentication status.
type AuthState int

const (
    AuthNotConfigured AuthState = iota  // No credentials file
    AuthNeedsLogin                       // Credentials exist but no token
    AuthReady                            // Token exists and valid/refreshable
    AuthRevoked                          // Token exists but refresh fails
)

func CheckAuthState() AuthState {
    credPath, _ := credentialsPath()
    if _, err := os.Stat(credPath); os.IsNotExist(err) {
        return AuthNotConfigured
    }
    tokPath, _ := tokenPath()
    if _, err := os.Stat(tokPath); os.IsNotExist(err) {
        return AuthNeedsLogin
    }
    return AuthReady
}
```

### Anti-Patterns to Avoid
- **Blocking the TUI during OAuth:** Never block the Bubble Tea event loop. The auth flow must run as a `tea.Cmd` (goroutine) with result delivered via message.
- **Storing tokens in config.toml:** Keep tokens in a separate file with restrictive permissions. Config is user-readable TOML; tokens are machine-managed JSON.
- **Requiring credentials to launch:** The app must start and work fully without any Google configuration (AUTH-04).
- **Re-prompting on expired access token:** `oauth2.TokenSource` handles refresh transparently. Only prompt when the refresh token itself is revoked.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| OAuth 2.0 token exchange | Manual HTTP POST to token endpoint | `oauth2.Config.Exchange()` | Handles error responses, token parsing, PKCE verification |
| Token refresh | Manual refresh token POST | `oauth2.TokenSource` / `ReuseTokenSource` | Automatic expiry check, thread-safe, handles error cases |
| PKCE code challenge | Manual SHA256 + base64url | `oauth2.GenerateVerifier()` + `oauth2.S256ChallengeOption()` | Correct encoding guaranteed |
| Google endpoint URLs | Hardcoded URLs | `google.Endpoint` | Maintained by Google, includes auth style |
| Credential JSON parsing | Manual JSON struct | `google.ConfigFromJSON()` | Handles installed vs web app types |

**Key insight:** `golang.org/x/oauth2` was designed exactly for this use case. It handles token refresh, PKCE, and endpoint configuration. The only custom code needed is: (1) the loopback HTTP server, (2) token file persistence, (3) browser opening, and (4) Bubble Tea integration.

## Common Pitfalls

### Pitfall 1: Missing `AccessTypeOffline` in Auth URL
**What goes wrong:** Google returns an access token but no refresh token. App works once, then fails on restart.
**Why it happens:** `access_type=offline` is required to get a refresh token. Default is `online`.
**How to avoid:** Always include `oauth2.AccessTypeOffline` in `AuthCodeURL()`.
**Warning signs:** Token file has empty `RefreshToken` field.

### Pitfall 2: Refresh Token Only Issued on First Consent
**What goes wrong:** User re-authenticates but gets no refresh token. Previous token is overwritten with one that has an empty refresh token.
**Why it happens:** Google only issues a refresh token on the first authorization. Subsequent authorizations return access tokens only.
**How to avoid:** Include `oauth2.SetAuthURLParam("prompt", "consent")` to force consent screen, which always returns a refresh token. Alternatively, preserve the existing refresh token if the new one is empty.
**Warning signs:** After re-auth, token file has empty `RefreshToken`.

### Pitfall 3: Port Conflict on Loopback Server
**What goes wrong:** Auth flow fails because the port is already in use.
**Why it happens:** Hardcoded port number.
**How to avoid:** Use ephemeral port (`:0`) and read the assigned port from the listener. RFC 8252 requires OAuth servers to allow any port for loopback redirects, and Google supports this.
**Warning signs:** "address already in use" error.

### Pitfall 4: Browser Opening in Headless/SSH Environment
**What goes wrong:** `xdg-open` fails or no browser is available.
**Why it happens:** User is on SSH or a headless server.
**How to avoid:** Fall back to printing the URL to stderr/stdout and let user copy-paste. Check `os.Getenv("DISPLAY")` or `os.Getenv("WAYLAND_DISPLAY")` before attempting `xdg-open`. If neither is set, print the URL.
**Warning signs:** `xdg-open` returns non-zero exit code.

### Pitfall 5: OAuth Callback Server Never Shuts Down
**What goes wrong:** Server lingers if user closes browser without completing flow.
**Why it happens:** No timeout on the callback wait.
**How to avoid:** Use `context.WithTimeout` (e.g., 2 minutes). On timeout, shut down server and return error.
**Warning signs:** App hangs waiting for auth, goroutine leak.

### Pitfall 6: Token File Race Condition
**What goes wrong:** Concurrent reads/writes corrupt the token file.
**Why it happens:** Background refresh writes while startup reads.
**How to avoid:** Use mutex or atomic write (temp file + rename), following the project's existing pattern in `config.Save()`.
**Warning signs:** Truncated JSON in token file.

## Code Examples

### Complete OAuth Config Setup
```go
// Source: Google Calendar API Go Quickstart + golang.org/x/oauth2 docs
import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

func newOAuthConfig(credentialsJSON []byte) (*oauth2.Config, error) {
    return google.ConfigFromJSON(credentialsJSON,
        "https://www.googleapis.com/auth/calendar.events.readonly",
    )
}
```

### Token File Path (Following Project Convention)
```go
// Source: Matches internal/config/paths.go pattern
func TokenPath() (string, error) {
    dir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, "todo-calendar", "google-token.json"), nil
}

func CredentialsPath() (string, error) {
    dir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, "todo-calendar", "credentials.json"), nil
}
```

### Settings Integration
```go
// Add to settings overlay: a "Google Calendar" option that shows status
// and triggers auth flow when selected.
//
// States displayed:
// - "Not configured" (no credentials.json)
// - "Sign in" (credentials exist, no token) -> triggers auth on Enter
// - "Connected" (token exists)
// - "Reconnect" (token revoked) -> triggers auth on Enter
```

### Opening Browser Cross-Platform
```go
// Source: Go stdlib cmd/internal/browser pattern
func openBrowser(url string) error {
    // Check for display server (Linux-specific)
    if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
        return fmt.Errorf("no display available")
    }
    return exec.Command("xdg-open", url).Start()
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| OOB (Out-of-Band) flow | Loopback redirect | Google deprecated OOB Feb 2022 | Must use loopback for desktop apps |
| No PKCE | PKCE recommended | RFC 8252 / Google recommendation | Use `S256ChallengeOption` |
| App passwords for Google | OAuth 2.0 only | Google disabled app passwords Sept 2024 | OAuth is the only option |
| Manual token refresh | `TokenSource` auto-refresh | Always available in x/oauth2 | Use `ReuseTokenSource` |
| Custom URI schemes | Loopback only for desktop | Google deprecated custom schemes | Use `http://127.0.0.1:{port}` |

**Deprecated/outdated:**
- OOB flow: Fully removed by Google. Cannot use "urn:ietf:wg:oauth:2.0:oob" anymore.
- App passwords: Disabled for Google Calendar as of September 2024.
- Custom URI schemes: No longer supported by Google for desktop OAuth clients.

## Scope Recommendation

For this phase (read-only events), use the most restrictive scope:
- `https://www.googleapis.com/auth/calendar.events.readonly` - View events only

Do NOT use `calendar.readonly` (which includes calendar list access, settings, etc.) unless Phase 34/35 specifically needs it. Least privilege is better for user trust during the consent screen.

## Open Questions

1. **Credentials distribution model**
   - What we know: Google requires a GCP project with Calendar API enabled and OAuth consent screen configured. For personal-use apps, user can create their own project.
   - What's unclear: Should the app ship with embedded credentials (developer manages GCP project) or require each user to create their own?
   - Recommendation: Config file approach (`credentials.json` in config dir). Document setup steps. This avoids Google's OAuth app verification process and is standard for personal-use CLI tools.

2. **First-run detection vs settings-only trigger**
   - What we know: CONF-02 says "settings or first-run detection".
   - What's unclear: Should the app prompt on first run if credentials.json exists but no token? Or should it be purely opt-in from settings?
   - Recommendation: Check on startup. If credentials exist but no token, show a non-blocking status message (not a modal). User triggers auth from settings. No auto-popup.

3. **Auth flow UX in alt-screen TUI**
   - What we know: The TUI runs in alt-screen mode. Browser opens separately. The loopback server runs in a goroutine.
   - What's unclear: Should the TUI show a "waiting for authentication..." state while the browser flow is active?
   - Recommendation: Yes. Show a simple overlay/status message: "Complete sign-in in your browser..." with an Esc to cancel. On success, update status to "Connected".

## Sources

### Primary (HIGH confidence)
- [golang.org/x/oauth2 package docs](https://pkg.go.dev/golang.org/x/oauth2) - Config, Token, TokenSource, PKCE functions, v0.35.0
- [golang.org/x/oauth2/google](https://pkg.go.dev/golang.org/x/oauth2/google) - Google endpoint, ConfigFromJSON
- [Google OAuth 2.0 for Native Apps](https://developers.google.com/identity/protocols/oauth2/native-app) - Official loopback flow documentation
- [Google Calendar API scopes](https://developers.google.com/workspace/calendar/api/auth) - Available OAuth scopes
- [Google Calendar API Go Quickstart](https://developers.google.com/workspace/calendar/api/quickstart/go) - Token save/load pattern

### Secondary (MEDIUM confidence)
- [RFC 8252: OAuth 2.0 for Native Apps](https://www.rfc-editor.org/rfc/rfc8252.html) - Loopback redirect specification
- [Google Loopback Migration Guide](https://developers.google.com/identity/protocols/oauth2/resources/loopback-migration) - Confirms desktop apps still supported

### Tertiary (LOW confidence)
- None - all findings verified with official sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - `golang.org/x/oauth2` is the de facto standard, verified via pkg.go.dev
- Architecture: HIGH - Pattern matches Google's own quickstart and project's existing conventions
- Pitfalls: HIGH - All documented in official Google OAuth docs or RFC 8252
- Bubble Tea integration: MEDIUM - Pattern follows project's existing editor exec pattern but OAuth flow is novel for this codebase

**Research date:** 2026-02-14
**Valid until:** 2026-03-14 (stable domain, OAuth 2.0 protocol is mature)
