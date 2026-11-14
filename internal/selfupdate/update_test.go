package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mjavint/ocli/internal/version"
)

func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func writeExec(t *testing.T, path string, content []byte) string {
	t.Helper()
	if err := os.WriteFile(path, content, 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func releaseServer(t *testing.T, tag string, assets map[string][]byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mjavint/ocli/releases/latest" || r.URL.Path == "/latest" {
			w.Header().Set("Location", "https://github.com/mjavint/ocli/releases/tag/"+tag)
			w.WriteHeader(http.StatusFound)
			return
		}
		if r.URL.Path == "/download/"+tag+"/checksums.txt" {
			for name, content := range assets {
				fmt.Fprintf(w, "%s  %s\n", sha256Hex(content), name)
			}
			return
		}
		prefix := "/download/" + tag + "/"
		if strings.HasPrefix(r.URL.Path, prefix) {
			name := strings.TrimPrefix(r.URL.Path, prefix)
			if c, ok := assets[name]; ok {
				_, _ = w.Write(c)
				return
			}
		}
		http.NotFound(w, r)
	}))
}

func TestUpdateApplies(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows not supported")
	}
	asset := "ocli-linux-amd64"
	oldBin := []byte("old")
	newBin := []byte("new-binary")
	srv := releaseServer(t, "v1.0.5", map[string][]byte{asset: newBin})
	defer srv.Close()

	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), oldBin)
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	res, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.4"},
		ExecutablePath: exe,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !res.Updated || res.Latest != "v1.0.5" {
		t.Fatalf("result = %+v, want updated to v1.0.5", res)
	}
	got, _ := os.ReadFile(exe)
	if string(got) != string(newBin) {
		t.Fatal("binary not replaced")
	}
}

func TestUpdateNoopSameHash(t *testing.T) {
	asset := "ocli-linux-amd64"
	bin := []byte("same")
	srv := releaseServer(t, "v1.0.5", map[string][]byte{asset: bin})
	defer srv.Close()
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), bin)
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	res, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.4"},
		ExecutablePath: exe,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if res.Updated {
		t.Fatalf("Updated = true, want false for same hash")
	}
}

func TestUpdateCurrentNewer(t *testing.T) {
	asset := "ocli-linux-amd64"
	srv := releaseServer(t, "v1.0.5", map[string][]byte{asset: []byte("x")})
	defer srv.Close()
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), []byte("current"))
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	res, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.6"},
		ExecutablePath: exe,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if res.Updated || res.Available {
		t.Fatalf("result = %+v, want noop newer", res)
	}
}

func TestUpdateChecksumMismatch(t *testing.T) {
	asset := "ocli-linux-amd64"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			w.Header().Set("Location", "https://github.com/mjavint/ocli/releases/tag/v1.0.5")
			w.WriteHeader(http.StatusFound)
		case "/download/v1.0.5/checksums.txt":
			fmt.Fprintf(w, "%s  %s\n", strings.Repeat("0", 64), asset)
		case "/download/v1.0.5/" + asset:
			_, _ = w.Write([]byte("corrupt"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	orig := []byte("original")
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), orig)
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	_, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.4"},
		ExecutablePath: exe,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err == nil {
		t.Fatal("Update() error = nil, want checksum mismatch")
	}
	got, _ := os.ReadFile(exe)
	if string(got) != string(orig) {
		t.Fatal("original binary modified on checksum mismatch")
	}
}

func TestUpdateMissingAsset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			w.Header().Set("Location", "https://github.com/mjavint/ocli/releases/tag/v1.0.5")
			w.WriteHeader(http.StatusFound)
		case "/download/v1.0.5/checksums.txt":
			fmt.Fprintf(w, "%s  %s\n", strings.Repeat("a", 64), "ocli-darwin-arm64")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), []byte("orig"))
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	if _, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.4"},
		ExecutablePath: exe,
		GOOS:           "linux",
		GOARCH:         "amd64",
	}); err == nil {
		t.Fatal("Update() error = nil, want missing asset error")
	}
}

func TestUpdateNetworkErrors(t *testing.T) {
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), []byte("orig"))

	t.Run("404", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		defer srv.Close()
		u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
		if _, err := u.Update(context.Background(), UpdateOptions{Current: version.Info{Version: "v1.0.4"}, ExecutablePath: exe, GOOS: "linux", GOARCH: "amd64"}); err == nil {
			t.Fatal("want error")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer srv.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
		if _, err := u.Update(ctx, UpdateOptions{Current: version.Info{Version: "v1.0.4"}, ExecutablePath: exe, GOOS: "linux", GOARCH: "amd64"}); err == nil {
			t.Fatal("want timeout/cancel error")
		}
	})
}

func TestUpdateCheckUnstable(t *testing.T) {
	asset := "ocli-linux-amd64"
	srv := releaseServer(t, "v1.0.5", map[string][]byte{asset: []byte("new")})
	defer srv.Close()
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), []byte("dev"))
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	res, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "(devel)"},
		ExecutablePath: exe,
		CheckOnly:      true,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err == nil || !errors.Is(err, ErrUpdateAvailable) {
		t.Fatalf("err = %v, want ErrUpdateAvailable", err)
	}
	if !res.Available || res.Latest != "v1.0.5" {
		t.Fatalf("result = %+v, want available v1.0.5", res)
	}
	got, _ := os.ReadFile(exe)
	if string(got) != "dev" {
		t.Fatal("check-only modified binary")
	}
}

func TestUpdateUnstableRequiresForce(t *testing.T) {
	asset := "ocli-linux-amd64"
	srv := releaseServer(t, "v1.0.5", map[string][]byte{asset: []byte("new")})
	defer srv.Close()
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), []byte("dev"))
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	_, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "(devel)"},
		ExecutablePath: exe,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err == nil {
		t.Fatal("Update() error = nil, want force required for unstable build")
	}
}

func TestUpdateForce(t *testing.T) {
	asset := "ocli-linux-amd64"
	bin := []byte("same")
	srv := releaseServer(t, "v1.0.5", map[string][]byte{asset: bin})
	defer srv.Close()
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), bin)
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	res, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.4"},
		ExecutablePath: exe,
		Force:          true,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !res.Updated {
		t.Fatalf("Updated = false, want true with --force")
	}
}

func TestUpdatePinnedTag(t *testing.T) {
	asset := "ocli-linux-amd64"
	srv := releaseServer(t, "v1.0.4", map[string][]byte{asset: []byte("old")})
	defer srv.Close()
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli"), []byte("newer"))
	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	_, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.5"},
		ExecutablePath: exe,
		TargetTag:      "not-a-tag",
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err == nil {
		t.Fatal("Update() error = nil, want invalid tag error")
	}
}

func TestUpdateWindows(t *testing.T) {
	exe := writeExec(t, filepath.Join(t.TempDir(), "ocli.exe"), []byte("x"))
	u := NewUpdater(&ReleaseClient{})
	_, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.4"},
		ExecutablePath: exe,
		GOOS:           "windows",
		GOARCH:         "amd64",
	})
	if err == nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("err = %v, want ErrUnsupported", err)
	}
}

func TestUpdateSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on windows")
	}
	asset := "ocli-linux-amd64"
	newBin := []byte("new-bin")
	srv := releaseServer(t, "v1.0.5", map[string][]byte{asset: newBin})
	defer srv.Close()

	dir := t.TempDir()
	real := writeExec(t, filepath.Join(dir, "ocli-real"), []byte("old"))
	link := filepath.Join(dir, "ocli")
	if err := os.Symlink(real, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	u := NewUpdater(&ReleaseClient{HTTP: srv.Client(), DiscoveryURL: srv.URL + "/latest", DownloadBaseURL: srv.URL + "/download"})
	_, err := u.Update(context.Background(), UpdateOptions{
		Current:        version.Info{Version: "v1.0.4"},
		ExecutablePath: link,
		GOOS:           "linux",
		GOARCH:         "amd64",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	got, _ := os.ReadFile(real)
	if string(got) != string(newBin) {
		t.Fatal("symlink target not replaced")
	}
}
