//go:build windows

package main

import (
    "reflect"
    "strings"
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
        t.Fatalf("UTF-16 count = %d, want 8 (5 Hangul code units + space + emoji pair)",got)
    }
}

func TestTextMacroNewlineUsesShiftEnter(t *testing.T) {
    // A literal newline, escaped \\n and CRLF must all emit the same
    // physical Shift+Enter chord, never a Unicode CR code unit.
    for _, input := range []string{"가\\n나", "가\n나", "가\r\n나"} {
        expanded, err := expandTextMacro(input)
        if err != nil {
            t.Fatal(err)
        }
        var events []macroInput
        for _, batch := range textMacroInputBatches(expanded) {
            events = append(events, batch...)
        }
        want := []macroInput{
            macroUnicodeUnit('가', false),
            macroUnicodeUnit('가', true),
            macroVirtualKey(virtualKeyShift, false),
            macroVirtualKey(virtualKeyReturn, false),
            macroVirtualKey(virtualKeyReturn, true),
            macroVirtualKey(virtualKeyShift, true),
            macroUnicodeUnit('나', false),
            macroUnicodeUnit('나', true),
        }
        if !reflect.DeepEqual(events, want) {
            t.Errorf("input %q emitted unexpected key events: %#v", input, events)
        }
    }
}

func TestTextMacroBatchDoesNotSplitEmojiOrNewline(t *testing.T) {
    // 31 letters use 62 events. The following emoji uses four events
    // and must move wholly into the next batch if no room remains.
    batches := textMacroInputBatches(strings.Repeat("가", 31) + "😀\n다음")
    if len(batches) < 2 {
        t.Fatalf("want multiple input batches, got %d", len(batches))
    }
    for i, batch := range batches {
        if len(batch) == 0 || len(batch) > 64 {
            t.Errorf("batch %d contains %d events", i, len(batch))
        }
    }
    // The first UTF-16 surrogate unit and its matching pair are in one batch.
    emoji := utf16.Encode([]rune("😀"))
    if len(emoji) != 2 {
        t.Fatal("emoji must have two UTF-16 units")
    }
    wantStart := []macroInput{
        macroUnicodeUnit(emoji[0], false),
        macroUnicodeUnit(emoji[0], true),
        macroUnicodeUnit(emoji[1], false),
        macroUnicodeUnit(emoji[1], true),
        macroVirtualKey(virtualKeyShift, false),
        macroVirtualKey(virtualKeyReturn, false),
        macroVirtualKey(virtualKeyReturn, true),
        macroVirtualKey(virtualKeyShift, true),
    }
    got := batches[1]
    if len(got) < len(wantStart) || !reflect.DeepEqual(got[:len(wantStart)], wantStart) {
        t.Fatalf("emoji/newline events are not contiguous: %#v", got)
    }
}
