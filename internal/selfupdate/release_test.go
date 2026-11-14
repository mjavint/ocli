package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatestTagFromRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mjavint/ocli/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Location", "https://github.com/mjavint/ocli/releases/tag/v1.0.5")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	c := &ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/mjavint/ocli/releases/latest"}
	tag, err := c.LatestTag(context.Background())
	if err != nil {
		t.Fatalf("LatestTag() error = %v", err)
	}
	if tag != "v1.0.5" {
		t.Fatalf("tag = %q, want v1.0.5", tag)
	}
}

func TestLatestTagRejects(t *testing.T) {
	tests := []struct {
		name     string
		location string
		status   int
	}{
		{name: "wrong host", location: "https://evil.example/releases/tag/v1.0.5", status: http.StatusFound},
		{name: "invalid tag", location: "https://github.com/mjavint/ocli/releases/tag/latest", status: http.StatusFound},
		{name: "bad path", location: "https://github.com/mjavint/ocli/releases/download/v1.0.5/x", status: http.StatusFound},
		{name: "no redirect", location: "", status: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.location != "" {
					w.Header().Set("Location", tt.location)
				}
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()
			c := &ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest"}
			if _, err := c.LatestTag(context.Background()); err == nil {
				t.Fatal("LatestTag() error = nil, want error")
			}
		})
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos, goarch, want string
		wantErr            bool
	}{
		{"darwin", "amd64", "ocli-darwin-amd64", false},
		{"darwin", "arm64", "ocli-darwin-arm64", false},
		{"linux", "amd64", "ocli-linux-amd64", false},
		{"linux", "arm64", "ocli-linux-arm64", false},
		{"windows", "amd64", "ocli-windows-amd64.exe", false},
		{"plan9", "amd64", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.goos+"-"+tt.goarch, func(t *testing.T) {
			got, err := AssetName(tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Fatal("AssetName() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("AssetName() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("AssetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDownloadVerified(t *testing.T) {
	content := []byte("binary-content")
	sum := sha256.Sum256(content)
	hexSum := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/download/v1.0.5/checksums.txt":
			fmt.Fprintf(w, "%s  ocli-linux-amd64\n", hexSum)
		case "/download/v1.0.5/ocli-linux-amd64":
			_, _ = w.Write(content)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := &ReleaseClient{HTTP: srv.Client(), DownloadBaseURL: srv.URL + "/download"}
	dir := t.TempDir()
	path, gotSum, err := c.DownloadAsset(context.Background(), "v1.0.5", "ocli-linux-amd64", hexSum, dir)
	if err != nil {
		t.Fatalf("DownloadAsset() error = %v", err)
	}
	if gotSum != hexSum {
		t.Fatalf("sum = %q, want %q", gotSum, hexSum)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != string(content) {
		t.Fatalf("content mismatch")
	}
}

func TestDefaultReleaseClient(t *testing.T) {
	c := NewDefaultReleaseClient()
	if c == nil {
		t.Fatal("NewDefaultReleaseClient() returned nil")
	}
	if c.DiscoveryURL != "https://github.com/mjavint/ocli/releases/latest" {
		t.Fatalf("DiscoveryURL = %q, want default", c.DiscoveryURL)
	}
	if c.DownloadBaseURL != "https://github.com/mjavint/ocli/releases/download" {
		t.Fatalf("DownloadBaseURL = %q, want default", c.DownloadBaseURL)
	}
}

func TestDownloadTooLarge(t *testing.T) {
	content := []byte(strings.Repeat("x", 1024))
	sum := sha256.Sum256(content)
	hexSum := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/v1.0.5/ocli-linux-amd64" {
			_, _ = w.Write(content)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := &ReleaseClient{HTTP: srv.Client(), DownloadBaseURL: srv.URL + "/download", MaxBytes: 64}
	dir := t.TempDir()
	if _, _, err := c.DownloadAsset(context.Background(), "v1.0.5", "ocli-linux-amd64", hexSum, dir); err == nil {
		t.Fatal("DownloadAsset() error = nil, want size limit error")
	}
	// ensure no leftover binary placed as final result except maybe temp cleaned
	matches, _ := filepath.Glob(filepath.Join(dir, "ocli-linux-amd64"))
	if len(matches) != 0 {
		t.Fatalf("unexpected final file left: %v", matches)
	}
}
