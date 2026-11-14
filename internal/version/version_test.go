package version

import (
	"runtime/debug"
	"testing"
)

func TestResolveLDFlags(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "deadbeef"},
			{Key: "vcs.time", Value: "2026-04-19T14:19:35Z"},
			{Key: "vcs.modified", Value: "false"},
		},
	}

	info := Resolve("v1.0.5", bi)
	if info.Version != "v1.0.5" {
		t.Fatalf("Version = %q, want %q", info.Version, "v1.0.5")
	}
	if info.Commit != "deadbeef" {
		t.Fatalf("Commit = %q, want %q", info.Commit, "deadbeef")
	}
	if info.Date != "2026-04-19T14:19:35Z" {
		t.Fatalf("Date = %q, want %q", info.Date, "2026-04-19T14:19:35Z")
	}
	if info.Modified {
		t.Fatal("Modified = true, want false")
	}
}

func TestResolveBuildInfo(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.0.4"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "23ef66b"},
			{Key: "vcs.time", Value: "2026-04-19T14:19:35Z"},
			{Key: "vcs.modified", Value: "false"},
		},
	}

	info := Resolve("", bi)
	if info.Version != "v1.0.4" {
		t.Fatalf("Version = %q, want %q", info.Version, "v1.0.4")
	}
}

func TestResolveModified(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.0.4+dirty"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "23ef66b"},
			{Key: "vcs.time", Value: "2026-04-19T14:19:35Z"},
			{Key: "vcs.modified", Value: "true"},
		},
	}

	info := Resolve("", bi)
	if !info.Modified {
		t.Fatal("Modified = false, want true")
	}
}

func TestResolveDevelFallback(t *testing.T) {
	info := Resolve("", nil)
	if info.Version != "(devel)" {
		t.Fatalf("Version = %q, want %q", info.Version, "(devel)")
	}
}

func TestInfoIsStable(t *testing.T) {
	tests := []struct {
		name string
		info Info
		want bool
	}{
		{name: "stable release", info: Info{Version: "v1.0.5"}, want: true},
		{name: "dirty", info: Info{Version: "v1.0.5", Modified: true}, want: false},
		{name: "devel", info: Info{Version: "(devel)"}, want: false},
		{name: "pseudo", info: Info{Version: "v1.0.4-3-gabc"}, want: false},
		{name: "prerelease", info: Info{Version: "v1.0.5-rc1"}, want: false},
		{name: "dirty suffix", info: Info{Version: "v1.0.4+dirty"}, want: false},
		{name: "invalid", info: Info{Version: "1.0.5"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.IsStable(); got != tt.want {
				t.Fatalf("IsStable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoShortAndFull(t *testing.T) {
	info := Info{Version: "v1.0.5", Commit: "23ef66b", Date: "2026-04-19T14:19:35Z"}
	if got := info.Short(); got != "ocli v1.0.5" {
		t.Fatalf("Short() = %q, want %q", got, "ocli v1.0.5")
	}
	want := "ocli v1.0.5 (commit 23ef66b, built 2026-04-19T14:19:35Z)"
	if got := info.Full(); got != want {
		t.Fatalf("Full() = %q, want %q", got, want)
	}

	dirty := Info{Version: "v1.0.5", Commit: "23ef66b", Date: "2026-04-19T14:19:35Z", Modified: true}
	if got := dirty.Full(); got != want+" [dirty]" {
		t.Fatalf("dirty Full() = %q, want %q", got, want+" [dirty]")
	}
}
