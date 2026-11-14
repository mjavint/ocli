package selfupdate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mjavint/ocli/internal/version"
	"golang.org/x/mod/semver"
)

var ErrUpdateAvailable = fmt.Errorf("update available")

// UpdateOptions configures a self-update.
type UpdateOptions struct {
	Current        version.Info
	ExecutablePath string
	TargetTag      string
	PinnedSHA256   string
	CheckOnly      bool
	Force          bool
	GOOS           string
	GOARCH         string
}

// UpdateResult reports the outcome of an update.
type UpdateResult struct {
	Current   string
	Latest    string
	SHA256    string
	Updated   bool
	Available bool
}

// Updater applies self-updates from GitHub releases.
type Updater struct {
	Releases *ReleaseClient
}

// NewUpdater creates an Updater.
func NewUpdater(client *ReleaseClient) *Updater {
	return &Updater{Releases: client}
}

// Update performs the self-update flow.
func (u *Updater) Update(ctx context.Context, opts UpdateOptions) (UpdateResult, error) {
	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	// Platform guard (windows replacement unsupported except check handled later for reporting?).
	if goos == "windows" && !opts.CheckOnly {
		return UpdateResult{Current: opts.Current.Version}, fmt.Errorf("%w: self-update is not supported on windows", ErrUnsupported)
	}
	if goos == "windows" && opts.CheckOnly {
		// check-only is allowed to report availability; still need to resolve release/asset below.
	}

	targetExplicit := strings.TrimSpace(opts.TargetTag) != ""
	tag := strings.TrimSpace(opts.TargetTag)
	if targetExplicit {
		if !tagRe.MatchString(tag) || !semver.IsValid(tag) {
			return UpdateResult{Current: opts.Current.Version}, fmt.Errorf("invalid target tag: %s", tag)
		}
	} else {
		if u.Releases == nil {
			return UpdateResult{Current: opts.Current.Version}, fmt.Errorf("release client is required")
		}
		latest, err := u.Releases.LatestTag(ctx)
		if err != nil {
			return UpdateResult{Current: opts.Current.Version}, err
		}
		tag = latest
	}

	asset, err := AssetName(goos, goarch)
	if err != nil {
		return UpdateResult{Current: opts.Current.Version, Latest: tag}, err
	}

	if u.Releases == nil {
		return UpdateResult{Current: opts.Current.Version, Latest: tag}, fmt.Errorf("release client is required")
	}
	checksumsData, err := u.Releases.FetchChecksums(ctx, tag)
	if err != nil {
		return UpdateResult{Current: opts.Current.Version, Latest: tag}, err
	}
	sums, err := ParseChecksums(checksumsData)
	if err != nil {
		return UpdateResult{Current: opts.Current.Version, Latest: tag}, err
	}
	expected, ok := sums[asset]
	if !ok {
		return UpdateResult{Current: opts.Current.Version, Latest: tag}, fmt.Errorf("asset not found in checksums: %s", asset)
	}
	if strings.TrimSpace(opts.PinnedSHA256) != "" {
		pinned := strings.ToLower(strings.TrimSpace(opts.PinnedSHA256))
		if err := VerifyChecksum(pinned, pinned); err != nil {
			return UpdateResult{Current: opts.Current.Version, Latest: tag}, fmt.Errorf("invalid pinned checksum: %w", err)
		}
		if pinned != strings.ToLower(expected) {
			return UpdateResult{Current: opts.Current.Version, Latest: tag}, fmt.Errorf("pinned checksum does not match release checksums")
		}
		expected = pinned
	}

	exePath := opts.ExecutablePath
	if exePath == "" {
		return UpdateResult{Current: opts.Current.Version, Latest: tag}, fmt.Errorf("executable path is required")
	}
	resolvedExe, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		resolvedExe = filepath.Clean(exePath)
	}
	curHash, _ := fileSHA256(resolvedExe)

	// Version guards
	currentStable := opts.Current.IsStable()
	currentVersion := opts.Current.Version
	if currentVersion == "" {
		currentVersion = "(devel)"
	}

	// If current is newer than tag and target not explicit -> noop
	if !targetExplicit && semver.IsValid(currentVersion) && semver.IsValid(tag) && semver.Compare(currentVersion, tag) > 0 {
		return UpdateResult{Current: currentVersion, Latest: tag, SHA256: curHash, Updated: false, Available: false}, nil
	}

	// Determine if replacement is available (different hash or forced/tag semantics)
	available := curHash != strings.ToLower(expected) || opts.Force
	if semver.IsValid(currentVersion) && semver.IsValid(tag) && semver.Compare(currentVersion, tag) < 0 {
		available = true
	}
	if targetExplicit && semver.IsValid(currentVersion) && semver.IsValid(tag) && semver.Compare(tag, currentVersion) != 0 {
		available = true
	}
	if !targetExplicit && semver.IsValid(currentVersion) && semver.IsValid(tag) && semver.Compare(currentVersion, tag) == 0 && curHash == strings.ToLower(expected) && !opts.Force {
		available = false
	}

	if opts.CheckOnly {
		res := UpdateResult{Current: currentVersion, Latest: tag, SHA256: curHash, Available: available, Updated: false}
		if available {
			return res, ErrUpdateAvailable
		}
		return res, nil
	}

	// Unstable builds require --force to replace unless target explicitly specified.
	if !opts.Force && !targetExplicit && !currentStable {
		return UpdateResult{Current: currentVersion, Latest: tag, SHA256: curHash}, fmt.Errorf("refusing to replace unstable build %q without --force", currentVersion)
	}

	// If hashes match and not forced, no-op (idempotent) without downloading.
	if !opts.Force && curHash != "" && strings.EqualFold(curHash, expected) {
		return UpdateResult{Current: currentVersion, Latest: tag, SHA256: curHash, Updated: false, Available: false}, nil
	}

	// Windows unsupported for actual replacement.
	if goos == "windows" {
		return UpdateResult{Current: currentVersion, Latest: tag}, fmt.Errorf("%w: self-update is not supported on windows", ErrUnsupported)
	}

	// Download and verify
	dir := filepath.Dir(resolvedExe)
	downloaded, gotSum, err := u.Releases.DownloadAsset(ctx, tag, asset, expected, dir)
	if err != nil {
		return UpdateResult{Current: currentVersion, Latest: tag, SHA256: curHash}, err
	}

	// Replace atomically
	if err := replaceFile(downloaded, resolvedExe); err != nil {
		_ = os.Remove(downloaded)
		return UpdateResult{Current: currentVersion, Latest: tag, SHA256: curHash}, err
	}
	// verify final
	finalSum, err := fileSHA256(resolvedExe)
	if err != nil || !strings.EqualFold(finalSum, expected) {
		return UpdateResult{Current: currentVersion, Latest: tag, SHA256: curHash}, fmt.Errorf("verify replaced executable")
	}

	return UpdateResult{Current: currentVersion, Latest: tag, SHA256: gotSum, Updated: true, Available: false}, nil
}

func replaceFile(srcTmp, dst string) error {
	// srcTmp is a temp file in same dir (from DownloadAsset); rename over dst.
	if err := os.Rename(srcTmp, dst); err != nil {
		// fallback: copy then rename
		return copyAtomic(srcTmp, dst)
	}
	return nil
}
