package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Side-by-side, version-qualified app binaries never overwrite a running
// legacy PulseNoteDeck.exe. Desktop/Start Menu links point to this release.
func appFilename() string { return "PulseNoteDeck_v" + productVersion + ".exe" }

var deckReleaseBinary = regexp.MustCompile(`(?i)^PulseNoteDeck_v[0-9]+\.[0-9]+\.[0-9]+\.exe$`)

func isDeckBinary(filename string) bool {
	return strings.EqualFold(filename, "PulseNoteDeck.exe") || deckReleaseBinary.MatchString(filename)
}

func installedAppFilenames(dir string) []string {
	out := []string{"PulseNoteDeck.exe"} // support pre-v7.4.22 installs
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if !e.IsDir() && deckReleaseBinary.MatchString(e.Name()) {
			out = append(out, e.Name())
		}
	}
	return out
}

// Only migrate login autostart if it already points into this very install
// directory at a known PulseNote Deck binary. Never opt anyone in implicitly.
func existingStartupBelongsToInstall(value, installDir string) bool {
	s := strings.TrimSpace(value)
	var target, args string
	if strings.HasPrefix(s, `"`) {
		end := strings.Index(s[1:], `"`)
		if end < 0 {
			return false
		}
		target = s[1 : end+1]
		args = strings.TrimSpace(s[end+2:])
	} else {
		parts := strings.Fields(s)
		if len(parts) == 0 {
			return false
		}
		target = parts[0]
		args = strings.Join(parts[1:], " ")
	}
	if args != "--background" {
		return false
	}
	if !isDeckBinary(filepath.Base(target)) {
		return false
	}
	return strings.EqualFold(filepath.Clean(filepath.Dir(target)), filepath.Clean(installDir))
}
