//go:build windows

package main

import (
    "fmt"
    "runtime"
    "syscall"
    "unicode/utf16"
    "unsafe"
)

// KEYEVENTF_UNICODE injects characters through VK_PACKET/WM_CHAR, not physical
// keys. Chat editors may discard a Unicode carriage return instead of making
// a new line. Inject Shift+Enter for LF, keeping the message unsent.
var sendUnicodeInput = syscall.NewLazyDLL("user32.dll").NewProc("SendInput")

const keyboardInputType uint32 = 1
const keyeventfUnicode uint32 = 0x0004
const keyeventfKeyUp uint32 = 0x0002
const virtualKeyShift uint16 = 0x10
const virtualKeyReturn uint16 = 0x0D

type macroKeyboardInput struct {
    VirtualKey uint16
    Scan uint16
    Flags uint32
    Time uint32
    ExtraInfo uintptr
}

type macroInput struct {
    Type uint32
    _ uint32 // Win64 INPUT union alignment.
    Keyboard macroKeyboardInput
    // The Win64 INPUT union takes 32 bytes for MOUSEINPUT, while KEYBDINPUT
    // takes 24 bytes. sizeof(INPUT) must remain 40, not 32 (error 87).
    _ [8]byte
}

func macroVirtualKey(vk uint16, keyUp bool) macroInput {
    var flags uint32
    if keyUp {
        flags = keyeventfKeyUp
    }
    return macroInput{
        Type: keyboardInputType,
        Keyboard: macroKeyboardInput{VirtualKey: vk, Flags: flags},
    }
}

func macroUnicodeUnit(unit uint16, keyUp bool) macroInput {
    flags := keyeventfUnicode
    if keyUp {
        flags |= keyeventfKeyUp
    }
    return macroInput{
        Type: keyboardInputType,
        Keyboard: macroKeyboardInput{Scan: unit, Flags: flags},
    }
}

// Group whole characters and newline key chords into bounded SendInput calls.
// Never split a UTF-16 surrogate pair or Shift+Enter across two batches.
func textMacroInputBatches(text string) [][]macroInput {
    const maxEventsPerBatch = 64
    var batches [][]macroInput
    batch := make([]macroInput, 0, maxEventsPerBatch)
    for _, r := range text {
        events := make([]macroInput, 0, 4)
        if r == '\n' {
            events = append(events,
                macroVirtualKey(virtualKeyShift, false),
                macroVirtualKey(virtualKeyReturn, false),
                macroVirtualKey(virtualKeyReturn, true),
                macroVirtualKey(virtualKeyShift, true),
            )
        } else {
            for _, unit := range utf16.Encode([]rune{r}) {
                events = append(events,
                    macroUnicodeUnit(unit, false),
                    macroUnicodeUnit(unit, true),
                )
            }
        }
        if len(batch)+len(events) > maxEventsPerBatch {
            batches = append(batches, batch)
            batch = make([]macroInput, 0, maxEventsPerBatch)
        }
        batch = append(batch, events...)
    }
    if len(batch) != 0 {
        batches = append(batches, batch)
    }
    return batches
}

func sendTextMacro(raw string) error {
    text, err := expandTextMacro(raw)
    if err != nil {
        return err
    }
    if unsafe.Sizeof(uintptr(0)) != 8 || unsafe.Sizeof(macroInput{}) != 40 {
        return fmt.Errorf("이 버전의 텍스트 매크로는 Windows x64 전용입니다")
    }
    for _, batch := range textMacroInputBatches(text) {
        count, _, callErr := sendUnicodeInput.Call(
            uintptr(len(batch)),
            uintptr(unsafe.Pointer(&batch[0])),
            uintptr(unsafe.Sizeof(batch[0])),
        )
        runtime.KeepAlive(batch)
        if count != uintptr(len(batch)) {
            return fmt.Errorf("Windows 텍스트 입력 실패 (%d/%d): %v", count, len(batch), callErr)
        }
    }
    return nil
}
