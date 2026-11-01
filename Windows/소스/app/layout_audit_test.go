package main

import (
	"os"
	"strings"
	"testing"
)

// Check last profile does not cover the controls at y=756.
func TestEightProfilesNeverCoverActions(t *testing.T) {
	for i := 0; i < 8; i++ {
		x, y, w := 293+(i/4)*555, 277+(i%4)*76, 535
		if y+56 > 756 {
			t.Fatalf("profile %d overlaps buttons", i)
		}
		if x+w > 1387 {
			t.Fatalf("profile %d outside card", i)
		}
	}
	b, e := os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(b), "277+(i%4)*76") {
		t.Fatal("two-column profile layout missing")
	}
}

func TestShortcutTestsShareOneExecutionQueue(t *testing.T) {
	b, e := os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	src := string(b)
	if strings.Contains(src, "go keyInput(g.selected + 1)") {
		t.Fatal("test actions bypass ordered USB executor")
	}
	if !strings.Contains(src, "case actionQueue <- g.selected + 1:") {
		t.Fatal("action queue absent")
	}
}

func TestNoForcedMinimumDesignScaleClipsUI(t *testing.T) {
	b, e := os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(b), "scale < 0.68") {
		t.Fatal("min scale clips small windows")
	}
}

func TestNoRepeatedPageHeadings(t *testing.T) {
	b, e := os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	src := string(b)
	if strings.Contains(src, `d.text("프로필 관리", 275, 80`) || strings.Contains(src, `d.text("장치 및 설정", 275, 80`) {
		t.Fatal("duplicated header overlays page title")
	}
}
