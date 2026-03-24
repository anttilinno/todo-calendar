package update

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockServer(t *testing.T, tag string, assets []asset) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := releaseResponse{TagName: tag, Assets: assets}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestCheckForUpdate_NewerVersion(t *testing.T) {
	srv := newMockServer(t, "v2.5.0", nil)
	defer srv.Close()

	rel, err := CheckForUpdate("v2.4.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel == nil {
		t.Fatal("expected release, got nil")
	}
	if rel.Tag != "v2.5.0" {
		t.Errorf("Tag = %q, want %q", rel.Tag, "v2.5.0")
	}
}

func TestCheckForUpdate_SameVersion(t *testing.T) {
	srv := newMockServer(t, "v2.4.0", nil)
	defer srv.Close()

	rel, err := CheckForUpdate("v2.4.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel != nil {
		t.Errorf("expected nil release for same version, got %+v", rel)
	}
}

func TestCheckForUpdate_OlderRelease(t *testing.T) {
	srv := newMockServer(t, "v2.3.0", nil)
	defer srv.Close()

	rel, err := CheckForUpdate("v2.4.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel != nil {
		t.Errorf("expected nil release for older version, got %+v", rel)
	}
}

func TestCheckForUpdate_NetworkError(t *testing.T) {
	rel, err := CheckForUpdate("v2.4.0", "http://127.0.0.1:1")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	if rel != nil {
		t.Errorf("expected nil release on error, got %+v", rel)
	}
}

func TestCheckForUpdate_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	rel, err := CheckForUpdate("v2.4.0", srv.URL)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if rel != nil {
		t.Errorf("expected nil release on error, got %+v", rel)
	}
}

func TestCheckForUpdate_WithAssets(t *testing.T) {
	assets := []asset{
		{Name: "todo-calendar-linux-amd64", DownloadURL: "https://example.com/binary"},
		{Name: "todo-calendar-linux-amd64.sha256", DownloadURL: "https://example.com/checksum"},
	}
	srv := newMockServer(t, "v2.5.0", assets)
	defer srv.Close()

	rel, err := CheckForUpdate("v2.4.0", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel == nil {
		t.Fatal("expected release, got nil")
	}
	if rel.BinaryURL != "https://example.com/binary" {
		t.Errorf("BinaryURL = %q, want %q", rel.BinaryURL, "https://example.com/binary")
	}
	if rel.ChecksumURL != "https://example.com/checksum" {
		t.Errorf("ChecksumURL = %q, want %q", rel.ChecksumURL, "https://example.com/checksum")
	}
}

func TestDownloadAsset(t *testing.T) {
	content := []byte("binary content here")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	data, err := DownloadAsset(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("got %q, want %q", data, content)
	}
}

func TestDownloadAsset_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	data, err := DownloadAsset(srv.URL)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if data != nil {
		t.Errorf("expected nil data on error, got %d bytes", len(data))
	}
}

func TestVerifyChecksum_Valid(t *testing.T) {
	binary := []byte("hello world")
	hash := sha256.Sum256(binary)
	checksumFile := []byte(fmt.Sprintf("%x  todo-calendar-linux-amd64\n", hash))

	if err := VerifyChecksum(binary, checksumFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	binary := []byte("hello world")
	checksumFile := []byte("0000000000000000000000000000000000000000000000000000000000000000  todo-calendar-linux-amd64\n")

	if err := VerifyChecksum(binary, checksumFile); err == nil {
		t.Fatal("expected error for checksum mismatch")
	}
}

func TestVerifyChecksum_MalformedChecksum(t *testing.T) {
	binary := []byte("hello world")
	checksumFile := []byte("not a valid checksum format xyz")

	if err := VerifyChecksum(binary, checksumFile); err == nil {
		t.Fatal("expected error for malformed checksum")
	}
}

func TestVerifyChecksum_HexOnly(t *testing.T) {
	binary := []byte("hello world")
	hash := sha256.Sum256(binary)
	checksumFile := []byte(fmt.Sprintf("%x\n", hash))

	if err := VerifyChecksum(binary, checksumFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"2.4.0", "2.5.0", -1},
		{"2.5.0", "2.4.0", 1},
		{"2.4.0", "2.4.0", 0},
		{"2.4.0", "2.4.1", -1},
		{"2.10.0", "2.9.0", 1},
		{"1.0.0", "2.0.0", -1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
