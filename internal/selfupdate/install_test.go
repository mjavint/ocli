package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mjavint/ocli/internal/version"
)

func writeBinary(t *testing.T, path string, content []byte, mode os.FileMode) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func hashFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func TestInstallCopies(t *testing.T) {
	dir := t.TempDir()
	src := writeBinary(t, filepath.Join(t.TempDir(), "ocli"), []byte("bin-v1"), 0o755)
	inst := NewInstaller(version.Info{Version: "v1.0.5"})
	res, err := inst.Install(InstallOptions{Source: src, Dir: dir})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !res.Changed {
		t.Fatal("Changed = false, want true")
	}
	dst := filepath.Join(dir, "ocli")
	if runtime.GOOS == "windows" {
		dst = filepath.Join(dir, "ocli.exe")
	}
	if res.Path != dst {
		t.Fatalf("Path = %q, want %q", res.Path, dst)
	}
	if got := hashFile(t, dst); got != hashFile(t, src) {
		t.Fatalf("destination hash mismatch")
	}
	if res.SHA256 == "" || res.SHA256 != hashFile(t, src) {
		t.Fatalf("SHA256 = %q, want source hash", res.SHA256)
	}
}

func TestInstallNoopSameHash(t *testing.T) {
	dir := t.TempDir()
	content := []byte("same-bin")
	src := writeBinary(t, filepath.Join(t.TempDir(), "ocli"), content, 0o755)
	dstName := "ocli"
	if runtime.GOOS == "windows" {
		dstName = "ocli.exe"
	}
	dst := writeBinary(t, filepath.Join(dir, dstName), content, 0o755)
	infoBefore, _ := os.Stat(dst)
	inst := NewInstaller(version.Info{Version: "v1.0.5"})
	res, err := inst.Install(InstallOptions{Source: src, Dir: dir})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if res.Changed {
		t.Fatal("Changed = true, want false for identical hash")
	}
	infoAfter, _ := os.Stat(dst)
	if !infoAfter.ModTime().Equal(infoBefore.ModTime()) {
		t.Fatal("destination was rewritten on noop install")
	}
}

func TestInstallReplacesDifferent(t *testing.T) {
	dir := t.TempDir()
	src := writeBinary(t, filepath.Join(t.TempDir(), "ocli"), []byte("new-bin"), 0o755)
	dstName := "ocli"
	if runtime.GOOS == "windows" {
		dstName = "ocli.exe"
	}
	writeBinary(t, filepath.Join(dir, dstName), []byte("old-bin"), 0o755)
	inst := NewInstaller(version.Info{Version: "v1.0.5"})
	res, err := inst.Install(InstallOptions{Source: src, Dir: dir})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !res.Changed {
		t.Fatal("Changed = false, want true")
	}
	dst := filepath.Join(dir, dstName)
	if got := hashFile(t, dst); got != hashFile(t, src) {
		t.Fatal("destination not replaced")
	}
}

func TestInstallSameFile(t *testing.T) {
	dir := t.TempDir()
	dstName := "ocli"
	if runtime.GOOS == "windows" {
		dstName = "ocli.exe"
	}
	src := writeBinary(t, filepath.Join(dir, dstName), []byte("bin"), 0o755)
	inst := NewInstaller(version.Info{Version: "v1.0.5"})
	res, err := inst.Install(InstallOptions{Source: src, Dir: dir})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if res.Changed {
		t.Fatal("Changed = true, want false when source and destination are same file")
	}
}

func TestInstallReadOnlyDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on windows")
	}
	parent := t.TempDir()
	dir := filepath.Join(parent, "bin")
	if err := os.Mkdir(dir, 0o555); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	src := writeBinary(t, filepath.Join(t.TempDir(), "ocli"), []byte("bin"), 0o755)
	inst := NewInstaller(version.Info{Version: "v1.0.5"})
	_, err := inst.Install(InstallOptions{Source: src, Dir: dir})
	if err == nil {
		t.Fatal("Install() error = nil, want permission/write error")
	}
}

func TestInstallPathWarning(t *testing.T) {
	dir := t.TempDir()
	src := writeBinary(t, filepath.Join(t.TempDir(), "ocli"), []byte("bin"), 0o755)
	t.Setenv("PATH", "/usr/bin:/bin")
	inst := NewInstaller(version.Info{Version: "v1.0.5"})
	res, err := inst.Install(InstallOptions{Source: src, Dir: dir})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !res.Changed {
		t.Fatal("Changed = false, want true")
	}
	if res.PathWarning == "" {
		t.Fatal("PathWarning empty, want warning when dir not in PATH")
	}
	if !strings.Contains(res.PathWarning, dir) {
		t.Fatalf("PathWarning = %q, want to mention dir %q", res.PathWarning, dir)
	}
}
