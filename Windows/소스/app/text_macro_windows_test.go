//go:build windows

package main

import (
    "testing"
    "unicode/utf16"
    "unsafe"
)

// Win64 sizeof(INPUT) is 40 because its union is sized for the 32-byte
// MOUSEINPUT, not the 24-byte KEYBDINPUT. Passing 32 to SendInput causes
// ERROR_INVALID_PARAMETER (87) and injects 0 of N keyboard events.
func TestUnicodeSendInputABIWin64(t *testing.T) {
    if got:=unsafe.Sizeof(uintptr(0));got!=8 {t.Skip("win64-only layout test")}
    if got:=unsafe.Sizeof(macroKeyboardInput{});got!=24 {
        t.Errorf("sizeof(KEYBDINPUT) = %d; want 24",got)
    }
    if got:=unsafe.Offsetof(macroKeyboardInput{}.ExtraInfo);got!=16 {
        t.Errorf("dwExtraInfo offset = %d; want 16",got)
    }
    if got:=unsafe.Offsetof(macroInput{}.Keyboard);got!=8 {
        t.Errorf("INPUT union offset = %d; want 8",got)
    }
    if got:=unsafe.Sizeof(macroInput{});got!=40 {
        t.Errorf("sizeof(INPUT) = %d; want 40 (the previous 32 caused error 87)",got)
    }
}

func TestUnicodeMacroSurrogatePairs(t *testing.T) {
    content,err:=expandTextMacro("안녕하세요 😀")
    if err!=nil {t.Fatal(err)}
    codes:=utf16.Encode([]rune(content))
    if got:=len(codes);got!=8 {
        t.Fatalf("UTF-16 count = %d, want 8 (6 Hangul code units + space + emoji pair)",got)
    }
}
