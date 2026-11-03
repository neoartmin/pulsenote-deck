package main

import "testing"

// Screen capture is an additional action, not a redefinition of existing
// Windows text macros or 12/15-key configuration layouts.
func TestCaptureAvailableInAllModels(t *testing.T) {
    if !validTypes["capture"] || !validTypes["text"] {
        t.Fatal("Windows must support both capture and legacy text")
    }
    if typeName("capture") != "화면 캡처" {
        t.Fatalf("capture action label: %s", typeName("capture"))
    }
    if needsValue("capture") {
        t.Fatal("capture must not require a free-text input value")
    }
    found := false
    for _, kind := range types {
        if kind == "capture" { found = true }
    }
    if !found {
        t.Fatal("capture action must appear in the Windows editor")
    }
    keypress, err := parseShortcut("WIN+SHIFT+S")
    if err != nil || len(keypress) != 3 ||
        keypress[0] != 0x5b || keypress[1] != 0x10 || keypress[2] != 'S' {
        t.Fatalf("Windows snip shortcut invalid: %v, %v", keypress, err)
    }
    for _, tc := range []struct{count int; side string}{
        {12, "R"}, {15, "L"}, {15, "R"},
    } {
        config := defaults()
        config.Selected, config.SelectedSide = tc.count, tc.side
        if !mutateKey(&config, 0, func(k *Key){
            k.Type = "capture"
            k.Label = "캡처"
            k.Color = "보라"
        }) { t.Fatal("capture not editable") }
        config = normalize(config)
        key, ok := selectedKey(config, 0)
        if !ok || key.Type != "capture" || key.Color != "보라" {
            t.Fatalf("model %d %s lost capture mapping: %#v", tc.count, tc.side, key)
        }
    }
}
