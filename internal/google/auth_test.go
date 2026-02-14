package google

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestSaveLoadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	expiry := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	tok := &oauth2.Token{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		Expiry:       expiry,
	}

	if err := saveToken(path, tok); err != nil {
		t.Fatalf("saveToken: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permissions = %o, want 0600", perm)
	}

	// Load and verify
	got, err := loadToken(path)
	if err != nil {
		t.Fatalf("loadToken: %v", err)
	}
	if got.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, tok.AccessToken)
	}
	if got.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, tok.RefreshToken)
	}
	if got.TokenType != tok.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, tok.TokenType)
	}
	if !got.Expiry.Equal(tok.Expiry) {
		t.Errorf("Expiry = %v, want %v", got.Expiry, tok.Expiry)
	}
}

func TestSaveTokenAtomicDir(t *testing.T) {
	dir := t.TempDir()
	// nested path that doesn't exist yet
	path := filepath.Join(dir, "sub", "deep", "token.json")

	tok := &oauth2.Token{
		AccessToken: "test",
		TokenType:   "Bearer",
	}

	if err := saveToken(path, tok); err != nil {
		t.Fatalf("saveToken to nested dir: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("token file not created: %v", err)
	}
}

func TestCheckAuthState_NotConfigured(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	tokPath := filepath.Join(dir, "google-token.json")

	state := checkAuthStateAt(credPath, tokPath)
	if state != AuthNotConfigured {
		t.Errorf("state = %d, want AuthNotConfigured (%d)", state, AuthNotConfigured)
	}
}

func TestCheckAuthState_NeedsLogin(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	tokPath := filepath.Join(dir, "google-token.json")

	// Create credentials file
	if err := os.WriteFile(credPath, []byte(`{"installed":{}}`), 0600); err != nil {
		t.Fatal(err)
	}

	state := checkAuthStateAt(credPath, tokPath)
	if state != AuthNeedsLogin {
		t.Errorf("state = %d, want AuthNeedsLogin (%d)", state, AuthNeedsLogin)
	}
}

func TestCheckAuthState_Ready(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	tokPath := filepath.Join(dir, "google-token.json")

	// Create both files
	if err := os.WriteFile(credPath, []byte(`{"installed":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tokPath, []byte(`{"access_token":"x"}`), 0600); err != nil {
		t.Fatal(err)
	}

	state := checkAuthStateAt(credPath, tokPath)
	if state != AuthReady {
		t.Errorf("state = %d, want AuthReady (%d)", state, AuthReady)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(path, []byte(`not json at all`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := loadConfig(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
