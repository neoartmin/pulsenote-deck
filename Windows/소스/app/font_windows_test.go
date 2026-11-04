//go:build windows

package main

import "testing"

func TestBundledPretendardWeightsAndRegistration(t *testing.T) {
    names := []string{
        "assets/Pretendard-Regular.otf",
        "assets/Pretendard-Medium.otf",
        "assets/Pretendard-Bold.otf",
    }
    for _, name := range names {
        b, err := artwork.ReadFile(name)
        if err != nil {t.Fatalf("bundled font missing %s: %v (run tools/prepare_pretendard.py)", name, err)}
        if len(b) < 1000000 {t.Fatalf("font content appears truncated: %s", name)}
    }
    if !loadBundledPretendard() {t.Fatal("unable to register bundled Pretendard in current GDI process")}
    if len(privateFontHandles) != len(names) {t.Fatal("private font registration count differs")}
    unloadBundledPretendard()
    if len(privateFontHandles) != 0 || len(privateFontBuffers) != 0 {t.Fatal("private fonts were not released")}
}
