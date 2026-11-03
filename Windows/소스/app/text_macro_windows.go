//go:build windows

package main

import (
    "fmt"
    "runtime"
    "syscall"
    "unicode/utf16"
    "unsafe"
)

// SendInput + KEYEVENTF_UNICODE injects UTF-16 (including Hangul and emoji)
// into the foreground control without overwriting the user's clipboard.
// The target must accept VK_PACKET / Unicode input and must not be elevated.
var sendUnicodeInput = syscall.NewLazyDLL("user32.dll").NewProc("SendInput")

const keyboardInputType uint32 = 1
const keyeventfUnicode uint32 = 0x0004
const keyeventfKeyUp uint32 = 0x0002

type macroKeyboardInput struct {
    VirtualKey uint16
    Scan uint16
    Flags uint32
    Time uint32
    ExtraInfo uintptr
}

type macroInput struct {
    Type uint32
    _ uint32 // Win64 INPUT union alignment (application builds for windows/amd64).
    Keyboard macroKeyboardInput
    // INPUT's union is sized by MOUSEINPUT (32 bytes on x64), whereas
    // KEYBDINPUT occupies only 24 bytes. Without these trailing 8 bytes
    // sizeof(INPUT) becomes 32 instead of the required 40, and SendInput
    // fails with ERROR_INVALID_PARAMETER (87).
    _ [8]byte
}

func sendTextMacro(raw string) error {
    text, err := expandTextMacro(raw)
    if err != nil { return err }
    if unsafe.Sizeof(uintptr(0)) != 8 || unsafe.Sizeof(macroInput{}) != 40 {
        return fmt.Errorf("이 버전의 텍스트 매크로는 Windows x64 전용입니다")
    }
    codes := utf16.Encode([]rune(text))
    if len(codes) == 0 { return fmt.Errorf("입력할 텍스트가 없습니다") }
    const codeUnitsPerBatch = 32
    for start := 0; start < len(codes); start += codeUnitsPerBatch {
        end := start+codeUnitsPerBatch
        if end>len(codes) {end=len(codes)}
        batch := make([]macroInput,0,(end-start)*2)
        for _,unit := range codes[start:end] {
            // Enter is a CR code in Windows text controls; keep other UTF-16
            // code units, including surrogate pairs, unmodified.
            if unit == '\n' { unit = '\r' }
            batch=append(batch,
                macroInput{Type:keyboardInputType,Keyboard:macroKeyboardInput{Scan:unit,Flags:keyeventfUnicode}},
                macroInput{Type:keyboardInputType,Keyboard:macroKeyboardInput{Scan:unit,Flags:keyeventfUnicode|keyeventfKeyUp}})
        }
        count,_,callErr:=sendUnicodeInput.Call(uintptr(len(batch)),
            uintptr(unsafe.Pointer(&batch[0])),uintptr(unsafe.Sizeof(batch[0])))
        runtime.KeepAlive(batch)
        if count != uintptr(len(batch)) {
            return fmt.Errorf("Windows 텍스트 입력 실패 (%d/%d): %v", count,len(batch),callErr)
        }
    }
    return nil
}
