package selfupdate

import (
	"strings"
	"testing"
)

func TestParseChecksums(t *testing.T) {
	darwin := strings.Repeat("a", 64)
	linux := strings.Repeat("b", 64)
	windows := strings.Repeat("c", 64)
	data := []byte(darwin + "  ocli-darwin-arm64\n" +
		linux + "  ocli-linux-amd64\n" +
		windows + "  *ocli-windows-amd64.exe\n")
	sums, err := ParseChecksums(data)
	if err != nil {
		t.Fatalf("ParseChecksums() error = %v", err)
	}
	if got := sums["ocli-darwin-arm64"]; got != darwin {
		t.Fatalf("darwin-arm64 = %q, want %q", got, darwin)
	}
	if got := sums["ocli-linux-amd64"]; got != linux {
		t.Fatalf("linux-amd64 = %q, want %q", got, linux)
	}
	if got := sums["ocli-windows-amd64.exe"]; got != windows {
		t.Fatalf("windows = %q, want %q", got, windows)
	}
}

func TestParseChecksumsRejects(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "invalid hex", data: "zzzz  ocli-linux-amd64\n"},
		{name: "duplicate asset", data: strings.Repeat("a", 64) + "  ocli-linux-amd64\n" + strings.Repeat("b", 64) + "  ocli-linux-amd64\n"},
		{name: "bad format", data: "abc123\n"},
		{name: "short hex", data: "abc123  ocli-linux-amd64\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseChecksums([]byte(tt.data)); err == nil {
				t.Fatal("ParseChecksums() error = nil, want error")
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	sum := strings.Repeat("a", 64)
	if err := VerifyChecksum(sum, sum); err != nil {
		t.Fatalf("VerifyChecksum() error = %v, want nil", err)
	}
	if err := VerifyChecksum(sum, strings.Repeat("b", 64)); err == nil {
		t.Fatal("VerifyChecksum() error = nil, want mismatch error")
	}
	if err := VerifyChecksum("not-hex", sum); err == nil {
		t.Fatal("VerifyChecksum() error = nil for invalid expected hex")
	}
}
