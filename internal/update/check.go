package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	releaseURL  = "https://api.github.com/repos/antti/todo-calendar/releases/latest"
	assetBase   = "todo-calendar-linux-amd64"
	httpTimeout = 5 * time.Second
)

// releaseResponse is the subset of the GitHub Releases API response we need.
type releaseResponse struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

// asset represents a single release asset from the GitHub Releases API.
type asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// Release holds information about a newer release.
type Release struct {
	Tag         string
	BinaryURL   string
	ChecksumURL string
}

// CheckForUpdate checks the GitHub Releases API for a newer version.
// If apiURL is empty, the default GitHub API URL is used.
// Returns nil, nil when the current version is up-to-date.
func CheckForUpdate(currentVersion, apiURL string) (*Release, error) {
	if apiURL == "" {
		apiURL = releaseURL
	}

	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing release response: %w", err)
	}

	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(release.TagName, "v")

	if compareVersions(current, latest) >= 0 {
		return nil, nil
	}

	rel := &Release{Tag: release.TagName}
	for _, a := range release.Assets {
		switch a.Name {
		case assetBase:
			rel.BinaryURL = a.DownloadURL
		case assetBase + ".sha256":
			rel.ChecksumURL = a.DownloadURL
		}
	}

	return rel, nil
}

// DownloadAsset fetches the content at the given URL with a 5-second timeout.
func DownloadAsset(url string) ([]byte, error) {
	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// VerifyChecksum verifies that the SHA256 hash of binary matches the expected
// hash in checksumFile. The checksum file may contain just the hex hash, or
// the standard "hash  filename" format produced by sha256sum.
func VerifyChecksum(binary, checksumFile []byte) error {
	expectedHex, err := parseChecksumFile(checksumFile)
	if err != nil {
		return err
	}

	actual := sha256.Sum256(binary)
	actualHex := hex.EncodeToString(actual[:])

	if !strings.EqualFold(actualHex, expectedHex) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHex, actualHex)
	}

	return nil
}

// parseChecksumFile extracts the hex hash from a checksum file.
// Accepts "hexhash  filename\n" or just "hexhash\n".
func parseChecksumFile(data []byte) (string, error) {
	line := strings.TrimSpace(string(data))
	if line == "" {
		return "", fmt.Errorf("empty checksum file")
	}

	// "hash  filename" format (two spaces between hash and filename)
	if parts := strings.SplitN(line, "  ", 2); len(parts) == 2 {
		line = parts[0]
	}

	// Validate it looks like a hex SHA256 hash (64 hex characters)
	if len(line) != 64 {
		return "", fmt.Errorf("invalid checksum length: %d (expected 64)", len(line))
	}
	if _, err := hex.DecodeString(line); err != nil {
		return "", fmt.Errorf("invalid hex in checksum: %w", err)
	}

	return line, nil
}

// compareVersions compares two semver strings (without "v" prefix).
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aNum, bNum int
		if i < len(aParts) {
			aNum, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bNum, _ = strconv.Atoi(bParts[i])
		}
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
	}
	return 0
}
