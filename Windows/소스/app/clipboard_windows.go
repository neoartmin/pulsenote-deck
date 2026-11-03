//go:build windows

package main

import (
    "fmt"
    "runtime"
    "syscall"
    "time"
    "unsafe"
)

const (
    clipboardCFUnicodeText = 13
    clipboardGMEMMoveable  = 0x0002
    clipboardGMEMZeroInit  = 0x0040
)

var (
    clipboardUserDLL    = syscall.NewLazyDLL("user32.dll")
    clipboardKernelDLL  = syscall.NewLazyDLL("kernel32.dll")
    clipboardOpen       = clipboardUserDLL.NewProc("OpenClipboard")
    clipboardEmpty      = clipboardUserDLL.NewProc("EmptyClipboard")
    clipboardSetData    = clipboardUserDLL.NewProc("SetClipboardData")
    clipboardClose      = clipboardUserDLL.NewProc("CloseClipboard")
    clipboardGlobalAlloc  = clipboardKernelDLL.NewProc("GlobalAlloc")
    clipboardGlobalLock   = clipboardKernelDLL.NewProc("GlobalLock")
    clipboardGlobalUnlock = clipboardKernelDLL.NewProc("GlobalUnlock")
    clipboardGlobalFree   = clipboardKernelDLL.NewProc("GlobalFree")
    clipboardMoveMemory   = clipboardKernelDLL.NewProc("RtlMoveMemory")
)

// The clipboard receives a COPY of the user's text. No synthetic keystrokes
// or accessibility-style injection are used. Clipboard write replaces the
// prior content, in the normal Windows Ctrl+C manner.
func copyMacroToClipboard(text string) error {
    safe,err:=clipboardMacroText(text)
    if err!=nil{return err}
    units,err:=syscall.UTF16FromString(safe)
    if err!=nil{return err}
    byteCount:=uintptr(len(units)*2)
    block,_,callErr:=clipboardGlobalAlloc.Call(
        clipboardGMEMMoveable|clipboardGMEMZeroInit,byteCount)
    if block==0{return fmt.Errorf("클립보드 메모리 할당 실패: %v",callErr)}
    // If ownership transfers to Windows after SetClipboardData, do NOT free.
    owned:=true
    defer func(){if owned{clipboardGlobalFree.Call(block)}}()
    mem,_,lockErr:=clipboardGlobalLock.Call(block)
    if mem==0{return fmt.Errorf("클립보드 메모리 잠금 실패: %v",lockErr)}
    // Copy into the GlobalLock allocation through the Win32 API. Avoid
    // uintptr -> unsafe.Pointer conversion, which Go vet rightly rejects.
    clipboardMoveMemory.Call(mem,uintptr(unsafe.Pointer(&units[0])),byteCount)
    clipboardGlobalUnlock.Call(block)
    runtime.KeepAlive(units)

    // Other desktop apps sometimes hold the clipboard briefly.
    opened:=false
    for try:=0;try<5;try++ {
        if ok,_,_:=clipboardOpen.Call(0);ok!=0 {opened=true;break}
        time.Sleep(25*time.Millisecond)
    }
    if !opened{return fmt.Errorf("클립보드가 다른 앱에서 사용 중입니다")}
    defer clipboardClose.Call()
    if ok,_,_:=clipboardEmpty.Call();ok==0 {
        return fmt.Errorf("클립보드를 초기화하지 못했습니다")
    }
    if saved,_,_:=clipboardSetData.Call(clipboardCFUnicodeText,block);saved==0 {
        return fmt.Errorf("클립보드에 문구를 저장하지 못했습니다")
    }
    owned=false
    return nil
}
