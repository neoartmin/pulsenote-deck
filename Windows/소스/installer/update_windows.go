//go:build windows

package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	wmDestroyLegacy    = 2
	wmReplaceInstance  = 0x8000 + 25
	processSynchronize = 0x00100000
	smtoAbortIfHung    = 0x0002
)

var (
	findDeckWindow         = udll.NewProc("FindWindowW")
	deckWindowPID          = udll.NewProc("GetWindowThreadProcessId")
	deckWindowTextLen      = udll.NewProc("GetWindowTextLengthW")
	deckWindowText         = udll.NewProc("GetWindowTextW")
	deckSendMessageTimeout = udll.NewProc("SendMessageTimeoutW")
	deckOpenProcess        = syscall.NewLazyDLL("kernel32.dll").NewProc("OpenProcess")
	deckWaitForSingle      = syscall.NewLazyDLL("kernel32.dll").NewProc("WaitForSingleObject")
	deckCloseHandle        = syscall.NewLazyDLL("kernel32.dll").NewProc("CloseHandle")
	deckOpenMutex          = syscall.NewLazyDLL("kernel32.dll").NewProc("OpenMutexW")
)

// Ensure the previous executable releases file locks before replacing the
// installed binary. A known legacy build's WM_DESTROY callback saves its
// pending edits, removes the tray icon and leaves its message loop.
func stopPreviousApp() error {
	// A pre-v7.5.2 build could accidentally start multiple processes with
	// the same window class and shared mutex. Close all of them, not just
	// the first FindWindowW result, before updating desktop shortcuts.
	for stopped := 0; stopped < 12; stopped++ {
		hwnd, _, _ := findDeckWindow.Call(uintptr(unsafe.Pointer(w("PulseNoteDeck_8_12_v74"))), 0)
		if hwnd == 0 {
			for attempt := 0; attempt < 30; attempt++ {
				mutex, _, mutexErr := deckOpenMutex.Call(processSynchronize, 0, uintptr(unsafe.Pointer(w("Local\\PulseNoteDeck_8_12_Main_v74"))))
				if mutex == 0 {
					if mutexErr == syscall.ERROR_FILE_NOT_FOUND {
						return nil
					}
					return fmt.Errorf("기존 앱 실행 상태 확인 실패: %v", mutexErr)
				}
				deckCloseHandle.Call(mutex)
				time.Sleep(100 * time.Millisecond)
				hwnd, _, _ = findDeckWindow.Call(uintptr(unsafe.Pointer(w("PulseNoteDeck_8_12_v74"))), 0)
				if hwnd != 0 {
					break
				}
			}
			if hwnd == 0 {
				return fmt.Errorf("이전 앱이 백그라운드에서 실행 중이지만 종료 요청 창을 찾지 못했습니다. 작업 관리자에서 PulseNoteDeck를 확인해 주세요")
			}
		}
		if err := stopOnePreviousApp(hwnd); err != nil {
			return err
		}
	}
	return fmt.Errorf("기존 PulseNote Deck 프로세스가 너무 많이 남아 있어 업데이트를 안전하게 중단했습니다")
}

// stopOnePreviousApp requests a graceful shutdown and waits for that
// specific process to exit. It must never force-kill unsaved settings.
func stopOnePreviousApp(hwnd uintptr) error {
	var pid uint32
	deckWindowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return fmt.Errorf("기존 앱 프로세스를 식별하지 못했습니다")
	}
	process, _, _ := deckOpenProcess.Call(processSynchronize, 0, uintptr(pid))
	if process == 0 {
		return fmt.Errorf("기존 앱의 종료 상태를 확인할 권한이 없습니다")
	}
	defer deckCloseHandle.Call(process)
	length, _, _ := deckWindowTextLen.Call(hwnd)
	if length > 1024 {
		length = 1024
	}
	text := make([]uint16, int(length)+2)
	deckWindowText.Call(hwnd, uintptr(unsafe.Pointer(&text[0])), uintptr(len(text)))
	title := syscall.UTF16ToString(text)
	msg := uintptr(wmDestroyLegacy)
	var major, minor, patch int
	if n, _ := fmt.Sscanf(title, "PulseNote Deck %d.%d.%d", &major, &minor, &patch); n == 3 && (major > 7 || (major == 7 && (minor > 4 || (minor == 4 && patch >= 14)))) {
		msg = wmReplaceInstance
	}
	var result uintptr
	sent, _, _ := deckSendMessageTimeout.Call(hwnd, msg, 0, 0, smtoAbortIfHung, 3000, uintptr(unsafe.Pointer(&result)))
	if sent == 0 {
		return fmt.Errorf("기존 앱이 종료 요청에 응답하지 않았습니다")
	}
	waited, _, _ := deckWaitForSingle.Call(process, 8000)
	if waited != 0 {
		return fmt.Errorf("기존 앱 종료를 확인하지 못했습니다")
	}
	return nil
}
