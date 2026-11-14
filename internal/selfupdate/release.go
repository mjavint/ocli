package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

var tagRe = regexp.MustCompile(`^v\d+\.\d+\.\d+([.-][0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// ReleaseClient discovers and downloads release artifacts.
type ReleaseClient struct {
	HTTP            *http.Client
	DiscoveryURL    string
	DownloadBaseURL string
	MaxBytes        int64
}

const (
	defaultDiscoveryURL    = "https://github.com/mjavint/ocli/releases/latest"
	defaultDownloadBaseURL = "https://github.com/mjavint/ocli/releases/download"
)

// NewDefaultReleaseClient returns a client configured for production releases.
func NewDefaultReleaseClient() *ReleaseClient {
	return &ReleaseClient{
		HTTP:            http.DefaultClient,
		DiscoveryURL:    defaultDiscoveryURL,
		DownloadBaseURL: defaultDownloadBaseURL,
	}
}

func (c *ReleaseClient) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

// AssetName returns the expected asset filename for the given platform.
func AssetName(goos, goarch string) (string, error) {
	switch goos + "/" + goarch {
	case "darwin/amd64":
		return "ocli-darwin-amd64", nil
	case "darwin/arm64":
		return "ocli-darwin-arm64", nil
	case "linux/amd64":
		return "ocli-linux-amd64", nil
	case "linux/arm64":
		return "ocli-linux-arm64", nil
	case "windows/amd64":
		return "ocli-windows-amd64.exe", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}
}

// LatestTag discovers the latest release tag via redirect.
func (c *ReleaseClient) LatestTag(ctx context.Context) (string, error) {
	if c.DiscoveryURL == "" {
		return "", fmt.Errorf("discovery URL is required")
	}
	client := c.httpClient()
	// copy client to avoid mutating caller's client, but preserve transport via base
	hc := *client
	hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.DiscoveryURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently && resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusTemporaryRedirect && resp.StatusCode != http.StatusPermanentRedirect {
		return "", fmt.Errorf("unexpected status discovering latest release: %s", resp.Status)
	}
	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("missing Location header")
	}
	tag, err := parseTagFromLocation(loc)
	if err != nil {
		return "", err
	}
	return tag, nil
}

func parseTagFromLocation(loc string) (string, error) {
	u, err := url.Parse(loc)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(u.Host, "github.com") {
		return "", fmt.Errorf("unexpected release host: %s", u.Host)
	}
	// expect /mjavint/ocli/releases/tag/<tag>
	clean := path.Clean(u.Path)
	parts := strings.Split(strings.TrimPrefix(clean, "/"), "/")
	if len(parts) != 5 || parts[0] != "mjavint" || parts[1] != "ocli" || parts[2] != "releases" || parts[3] != "tag" {
		return "", fmt.Errorf("unexpected release path: %s", u.Path)
	}
	tag := parts[4]
	if !tagRe.MatchString(tag) || !semver.IsValid(tag) {
		return "", fmt.Errorf("invalid release tag: %s", tag)
	}
	return tag, nil
}

// FetchChecksums downloads and returns checksums.txt for a tag.
func (c *ReleaseClient) FetchChecksums(ctx context.Context, tag string) ([]byte, error) {
	urlStr, err := c.downloadURL(tag, "checksums.txt")
	if err != nil {
		return nil, err
	}
	return c.fetchLimited(ctx, urlStr, 64*1024)
}

func (c *ReleaseClient) downloadURL(tag, asset string) (string, error) {
	if c.DownloadBaseURL == "" {
		return "", fmt.Errorf("download base URL is required")
	}
	base := strings.TrimRight(c.DownloadBaseURL, "/")
	return base + "/" + tag + "/" + asset, nil
}

func (c *ReleaseClient) fetchLimited(ctx context.Context, urlStr string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s for %s", resp.Status, urlStr)
	}
	if limit <= 0 {
		limit = 64 * 1024
	}
	lr := &io.LimitedReader{R: resp.Body, N: limit + 1}
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response too large")
	}
	return data, nil
}

// DownloadAsset downloads an asset to dir and returns its path and SHA256.
// It verifies the hash during streaming and fails closed on mismatch/limit.
func (c *ReleaseClient) DownloadAsset(ctx context.Context, tag, asset, expectedSHA string, dir string) (string, string, error) {
	expectedSHA = strings.ToLower(strings.TrimSpace(expectedSHA))
	if !hex64Re.MatchString(expectedSHA) {
		return "", "", fmt.Errorf("invalid expected checksum")
	}
	urlStr, err := c.downloadURL(tag, asset)
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", "", err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status %s for %s", resp.Status, urlStr)
	}

	maxBytes := c.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 128 * 1024 * 1024
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	tmp, err := os.CreateTemp(dir, ".ocli-download-*")
	if err != nil {
		return "", "", err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	h := sha256.New()
	lr := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
	n, err := io.Copy(io.MultiWriter(tmp, h), lr)
	if err != nil {
		return "", "", err
	}
	if n > maxBytes {
		return "", "", fmt.Errorf("download too large")
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if err := VerifyChecksum(expectedSHA, sum); err != nil {
		return "", "", err
	}
	if err := tmp.Chmod(0o755); err != nil {
		return "", "", err
	}
	if err := tmp.Sync(); err != nil {
		return "", "", err
	}
	if err := tmp.Close(); err != nil {
		return "", "", err
	}
	// keep temp path as returned path (updater will move/replace it)
	ok = true
	return tmpName, sum, nil
}
