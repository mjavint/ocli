package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mjavint/ocli/internal/version"
)

// InstallOptions configures a self-install.
type InstallOptions struct {
	Source string
	Dir    string
	Force  bool
}

// InstallResult reports the outcome of a self-install.
type InstallResult struct {
	Path        string
	Version     string
	SHA256      string
	Changed     bool
	PathWarning string
}

// Installer copies the running binary into a directory on the PATH.
type Installer struct {
	version version.Info
}

// NewInstaller creates an Installer.
func NewInstaller(v version.Info) *Installer {
	return &Installer{version: v}
}

// Install performs the self-install.
func (i *Installer) Install(opts InstallOptions) (InstallResult, error) {
	src, err := filepath.EvalSymlinks(opts.Source)
	if err != nil {
		src = filepath.Clean(opts.Source)
	}
	st, err := os.Stat(src)
	if err != nil {
		return InstallResult{}, err
	}
	if !st.Mode().IsRegular() {
		return InstallResult{}, fmt.Errorf("source is not a regular file: %s", src)
	}

	dir := opts.Dir
	if dir == "" {
		dir = defaultInstallDir()
	} else {
		dir = expandHome(dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return InstallResult{}, err
	}

	name := "ocli"
	if runtime.GOOS == "windows" {
		name = "ocli.exe"
	}
	dst := filepath.Join(dir, name)

	resolvedDst, err := filepath.EvalSymlinks(dst)
	if err != nil {
		resolvedDst = filepath.Clean(dst)
	}

	if sameFile(src, resolvedDst) {
		h, _ := fileSHA256(src)
		return InstallResult{Path: resolvedDst, Version: i.version.Version, SHA256: h, Changed: false, PathWarning: pathWarning(dir)}, nil
	}

	srcHash, err := fileSHA256(src)
	if err != nil {
		return InstallResult{}, err
	}

	if !opts.Force {
		if info, err := os.Stat(resolvedDst); err == nil && info.Mode().IsRegular() {
			dstHash, err := fileSHA256(resolvedDst)
			if err == nil && dstHash == srcHash {
				return InstallResult{Path: resolvedDst, Version: i.version.Version, SHA256: srcHash, Changed: false, PathWarning: pathWarning(dir)}, nil
			}
		}
	}

	if info, err := os.Stat(resolvedDst); err == nil && info.IsDir() {
		return InstallResult{}, fmt.Errorf("destination is a directory: %s", resolvedDst)
	}

	if err := copyAtomic(src, resolvedDst); err != nil {
		return InstallResult{}, err
	}
	finalHash, err := fileSHA256(resolvedDst)
	if err != nil || finalHash != srcHash {
		return InstallResult{}, fmt.Errorf("verify installed binary: %w", err)
	}

	return InstallResult{Path: resolvedDst, Version: i.version.Version, SHA256: finalHash, Changed: true, PathWarning: pathWarning(dir)}, nil
}

func defaultInstallDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".local/bin"
	}
	return filepath.Join(home, ".local", "bin")
}

func expandHome(p string) string {
	if p == "~" {
		if h, err := os.UserHomeDir(); err == nil {
			return h
		}
	}
	if strings.HasPrefix(p, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, p[2:])
		}
	}
	return p
}

func pathWarning(dir string) string {
	cleanDir := filepath.Clean(dir)
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.Clean(p) == cleanDir {
			return ""
		}
	}
	return fmt.Sprintf("warning: %s is not in PATH; add it with: export PATH=\"%s:$PATH\"", dir, dir)
}

func sameFile(a, b string) bool {
	ia, err1 := os.Stat(a)
	ib, err2 := os.Stat(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return os.SameFile(ia, ib)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyAtomic(src, dst string) error {
	dir := filepath.Dir(dst)
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp, err := os.CreateTemp(dir, ".ocli-install-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(tmp, in); err != nil {
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return err
	}
	ok = true
	return nil
}

// ErrUnsupported is returned for unsupported operations (placeholder for PR2).
var ErrUnsupported = errors.New("unsupported")
