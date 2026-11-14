package selfupdate

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var hex64Re = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// ParseChecksums parses a checksums.txt payload in the form "<sha256>  <asset>".
// It accepts the optional '*' marker before the asset name, ignores empty lines,
// and rejects invalid hex, malformed lines, and duplicate assets.
func ParseChecksums(data []byte) (map[string]string, error) {
	sums := make(map[string]string)
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid checksum line: %q", line)
		}
		sum := strings.ToLower(fields[0])
		if !hex64Re.MatchString(sum) {
			return nil, fmt.Errorf("invalid checksum hex: %q", fields[0])
		}
		asset := strings.TrimPrefix(fields[1], "*")
		if asset == "" {
			return nil, fmt.Errorf("invalid checksum asset: %q", line)
		}
		if _, exists := sums[asset]; exists {
			return nil, fmt.Errorf("duplicate checksum asset: %s", asset)
		}
		sums[asset] = sum
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(sums) == 0 {
		return nil, fmt.Errorf("no checksums found")
	}
	return sums, nil
}

// VerifyChecksum compares expected and actual hex hashes (case-insensitive).
func VerifyChecksum(expected, actual string) error {
	expected = strings.ToLower(strings.TrimSpace(expected))
	actual = strings.ToLower(strings.TrimSpace(actual))
	if !hex64Re.MatchString(expected) {
		return fmt.Errorf("invalid expected checksum: %q", expected)
	}
	if !hex64Re.MatchString(actual) {
		return fmt.Errorf("invalid actual checksum: %q", actual)
	}
	if expected != actual {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}
