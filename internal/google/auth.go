package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// calendarScope is the read-only scope for Google Calendar events.
const calendarScope = "https://www.googleapis.com/auth/calendar.events.readonly"

// --- Path helpers ---

// configDir returns the todo-calendar config directory path.
func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "todo-calendar"), nil
}

// CredentialsPath returns the path to the Google OAuth credentials file.
func CredentialsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// TokenPath returns the path to the persisted Google OAuth token.
func TokenPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "google-token.json"), nil
}

// --- AuthState ---

// AuthState represents the current Google authentication state.
type AuthState int

const (
	// AuthNotConfigured means credentials.json does not exist.
	AuthNotConfigured AuthState = iota
	// AuthNeedsLogin means credentials exist but no token is saved.
	AuthNeedsLogin
	// AuthReady means both credentials and token exist.
	AuthReady
	// AuthRevoked means the token exists but refresh has failed.
	AuthRevoked
)

// CheckAuthState checks the current authentication state by inspecting
// the filesystem. No network calls are made.
func CheckAuthState() AuthState {
	credPath, err := CredentialsPath()
	if err != nil {
		return AuthNotConfigured
	}
	tokPath, err := TokenPath()
	if err != nil {
		return AuthNotConfigured
	}
	return checkAuthStateAt(credPath, tokPath)
}

// checkAuthStateAt checks auth state using explicit paths (for testing).
func checkAuthStateAt(credPath, tokPath string) AuthState {
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		return AuthNotConfigured
	}
	if _, err := os.Stat(tokPath); os.IsNotExist(err) {
		return AuthNeedsLogin
	}
	return AuthReady
}

// --- Config loading ---

// loadConfig reads a Google OAuth credentials.json and returns an oauth2.Config.
func loadConfig(credPath string) (*oauth2.Config, error) {
	b, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}
	cfg, err := google.ConfigFromJSON(b, calendarScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	return cfg, nil
}

// --- Token persistence ---

// saveToken writes an OAuth token to disk atomically with 0600 permissions.
func saveToken(path string, tok *oauth2.Token) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".token-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0600); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// loadToken reads a persisted OAuth token from disk.
func loadToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

// --- Persisting token source ---

// persistingTokenSource wraps an oauth2.TokenSource and persists the token
// to disk whenever it changes (e.g., after a refresh).
type persistingTokenSource struct {
	inner    oauth2.TokenSource
	path     string
	mu       sync.Mutex
	lastHash string
}

func tokenHash(tok *oauth2.Token) string {
	return tok.AccessToken + "|" + tok.Expiry.String()
}

// Token returns a token, saving it to disk if it has changed.
func (s *persistingTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tok, err := s.inner.Token()
	if err != nil {
		return nil, err
	}

	h := tokenHash(tok)
	if h != s.lastHash {
		if err := saveToken(s.path, tok); err != nil {
			return nil, fmt.Errorf("persist token: %w", err)
		}
		s.lastHash = h
	}
	return tok, nil
}

// --- Public TokenSource ---

// TokenSource returns an oauth2.TokenSource that automatically refreshes
// and persists the token. Returns an error if credentials or token are missing.
func TokenSource() (oauth2.TokenSource, error) {
	credPath, err := CredentialsPath()
	if err != nil {
		return nil, fmt.Errorf("credentials path: %w", err)
	}
	tokPath, err := TokenPath()
	if err != nil {
		return nil, fmt.Errorf("token path: %w", err)
	}

	cfg, err := loadConfig(credPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	tok, err := loadToken(tokPath)
	if err != nil {
		return nil, fmt.Errorf("load token: %w", err)
	}

	inner := cfg.TokenSource(context.Background(), tok)
	return &persistingTokenSource{
		inner:    inner,
		path:     tokPath,
		lastHash: tokenHash(tok),
	}, nil
}

// --- Auth flow ---

// performAuthFlow runs the OAuth 2.0 authorization code flow with PKCE
// using a loopback redirect on an ephemeral port.
func performAuthFlow(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	cfg.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	verifier := oauth2.GenerateVerifier()
	authURL := cfg.AuthCodeURL("state",
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Authorization successful!</h1><p>You can close this tab.</p></body></html>")
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	defer srv.Shutdown(context.Background())

	// Try to open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Open this URL in your browser:\n%s\n", authURL)
	}

	// Wait for callback with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	select {
	case code := <-codeCh:
		tok, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			return nil, fmt.Errorf("exchange: %w", err)
		}
		return tok, nil
	case err := <-errCh:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("auth flow timed out after 2 minutes")
	}
}

// openBrowser attempts to open a URL in the default browser.
func openBrowser(url string) error {
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return fmt.Errorf("no display available")
	}
	path, err := exec.LookPath("xdg-open")
	if err != nil {
		return fmt.Errorf("xdg-open not found: %w", err)
	}
	return exec.Command(path, url).Start()
}

// --- Bubble Tea integration ---

// AuthResultMsg is sent when the auth flow completes.
type AuthResultMsg struct {
	Success bool
	Err     error
}

// StartAuthFlow returns a tea.Cmd that runs the OAuth flow and returns
// an AuthResultMsg when complete.
func StartAuthFlow() tea.Cmd {
	return func() tea.Msg {
		credPath, err := CredentialsPath()
		if err != nil {
			return AuthResultMsg{Err: fmt.Errorf("credentials path: %w", err)}
		}

		cfg, err := loadConfig(credPath)
		if err != nil {
			return AuthResultMsg{Err: fmt.Errorf("load config: %w", err)}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		tok, err := performAuthFlow(ctx, cfg)
		if err != nil {
			return AuthResultMsg{Err: err}
		}

		tokPath, err := TokenPath()
		if err != nil {
			return AuthResultMsg{Err: fmt.Errorf("token path: %w", err)}
		}

		if err := saveToken(tokPath, tok); err != nil {
			return AuthResultMsg{Err: fmt.Errorf("save token: %w", err)}
		}

		return AuthResultMsg{Success: true}
	}
}
