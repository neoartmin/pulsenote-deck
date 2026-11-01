package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionedAppAvoidsLockedLegacyTarget(t *testing.T) {
	if appFilename() != "PulseNoteDeck_v7.5.8.exe" {
		t.Fatalf("wrong version: %s", appFilename())
	}
	if appFilename() == "PulseNoteDeck.exe" {
		t.Fatal("must never overwrite a running legacy exe")
	}
}
func TestAppCleanupSelectsOnlyReleaseFiles(t *testing.T) {
	d := t.TempDir()
	for _, s := range []string{"PulseNoteDeck.exe", "PulseNoteDeck_v7.4.22.exe", "PulseNoteDeck_v7.4.18.exe", "PulseNoteDeck_vFOO.exe", "unrelated.exe"} {
		if err := os.WriteFile(filepath.Join(d, s), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	found := installedAppFilenames(d)
	if len(found) != 3 {
		t.Fatalf("unexpected candidates: %v", found)
	}
	for _, f := range found {
		if f == "unrelated.exe" || f == "PulseNoteDeck_vFOO.exe" {
			t.Fatalf("unsafe delete candidate %q", f)
		}
	}
}
func TestMigrateOnlyOptedInCurrentInstall(t *testing.T) {
	d := filepath.Join("C:", "Users", "test", "AppData", "Local", "Programs", "PulseNoteDeckUnified")
	cases := []struct {
		value string
		valid bool
	}{
		{`"` + filepath.Join(d, "PulseNoteDeck.exe") + `" --background`, true},
		{`"` + filepath.Join(d, "PulseNoteDeck_v7.4.18.exe") + `" --background`, true},
		{`"` + filepath.Join(d, "PulseNoteDeck.exe") + `"`, false},
		{`"` + filepath.Join(d, "other.exe") + `" --background`, false},
		{`"` + filepath.Join("C:", "temp", "PulseNoteDeck.exe") + `" --background`, false},
	}
	for _, tc := range cases {
		if got := existingStartupBelongsToInstall(tc.value, d); got != tc.valid {
			t.Errorf("case %q: got %v want %v", tc.value, got, tc.valid)
		}
	}
}
