package version

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
)

// Info describes the build identity of the ocli binary.
type Info struct {
	Version  string
	Commit   string
	Date     string
	Modified bool
}

// Version is injected at build time via -ldflags.
var Version string

var stableVersionRE = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// Resolve builds an Info from an injected version and Go build info.
func Resolve(ldVersion string, bi *debug.BuildInfo) Info {
	info := Info{Version: strings.TrimSpace(ldVersion)}

	if bi != nil {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				info.Commit = s.Value
			case "vcs.time":
				info.Date = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					info.Modified = true
				}
			}
		}
		if info.Version == "" && bi.Main.Version != "" {
			info.Version = bi.Main.Version
		}
	}

	if info.Version == "" {
		info.Version = "(devel)"
	}
	if strings.Contains(info.Version, "dirty") {
		info.Modified = true
	}

	return info
}

// Current returns the Info for the running binary.
func Current() Info {
	bi, _ := debug.ReadBuildInfo()
	return Resolve(Version, bi)
}

// Short returns the short, one-line version string.
func (i Info) Short() string {
	return fmt.Sprintf("ocli %s", i.Version)
}

// Full returns the detailed version string.
func (i Info) Full() string {
	s := i.Short()
	var details []string
	if i.Commit != "" {
		details = append(details, "commit "+i.Commit)
	}
	if i.Date != "" {
		details = append(details, "built "+i.Date)
	}
	if len(details) > 0 {
		s += " (" + strings.Join(details, ", ") + ")"
	}
	if i.Modified {
		s += " [dirty]"
	}
	return s
}

// IsStable reports whether the version is a clean release build.
func (i Info) IsStable() bool {
	if i.Modified {
		return false
	}
	return stableVersionRE.MatchString(i.Version)
}
