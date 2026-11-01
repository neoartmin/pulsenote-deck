//go:build windows

package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	winErrorAlreadyExists = 183
	processSynchronize    = 0x00100000
	smtoAbortIfHung       = 0x0002
	WM_REPLACE_INSTANCE   = WM_APP + 25
	WM_REOPEN_INSTANCE    = WM_APP + 26
)

var (
	sendMessageTimeout = u.NewProc("SendMessageTimeoutW")
	getWindowTextLen   = u.NewProc("GetWindowTextLengthW")
	getWindowPID       = u.NewProc("GetWindowThreadProcessId")
	openProcess        = k32.NewProc("OpenProcess")
	waitForSingle      = k32.NewProc("WaitForSingleObject")
	closeHandle        = k32.NewProc("CloseHandle")
)

func runningWindow() uintptr {
	hwnd, _, _ := findWindow.Call(ptr(windowClassName), 0)
	return hwnd
}

func runningWindowTitle(hwnd uintptr) string {
	length, _, _ := getWindowTextLen.Call(hwnd)
	if length > 1024 {
		length = 1024
	}
	buf := make([]uint16, int(length)+2)
	getWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func requestRunningExit(hwnd uintptr, oldTitle string) error {
	if hwnd == 0 {
		return fmt.Errorf("기존 프로그램 창을 찾을 수 없습니다")
	}
	var pid uint32
	getWindowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return fmt.Errorf("기존 프로그램 PID 확인 실패")
	}
	process, _, _ := openProcess.Call(processSynchronize, 0, uintptr(pid))
	if process != 0 {
		defer closeHandle.Call(process)
	}

	// v7.4.14+ supports explicit shutdown via a private app message. Releases
	// <=7.4.13 do not; their WM_DESTROY handler flushes edits, removes the
	// tray icon, and posts WM_QUIT, so use this only for known legacy builds.
	message := uintptr(WM_REPLACE_INSTANCE)
	if instanceAction(oldTitle, version, false) == InstanceReplace {
		// Every release before 7.4.14 has a different version title.
		// Newer future releases also accept WM_REPLACE_INSTANCE.
		if !supportsReplacementIPC(oldTitle) {
			message = WM_DESTROY
		}
	}
	var result uintptr
	ok, _, _ := sendMessageTimeout.Call(hwnd, message, 0, 0, smtoAbortIfHung, 3000, uintptr(unsafe.Pointer(&result)))
	if ok == 0 {
		return fmt.Errorf("기존 프로그램에 종료 요청을 전달하지 못했습니다")
	}
	if process != 0 {
		result, _, _ := waitForSingle.Call(process, 6000)
		if result == 0 {
			return nil
		} // WAIT_OBJECT_0
	}
	for i := 0; i < 30; i++ {
		if runningWindow() == 0 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("기존 프로그램이 종료되지 않았습니다")
}

// For the first migration, all legacy binaries precede 7.4.14. Future
// versions can use the explicit IPC handler regardless of their exact version.
func supportsReplacementIPC(title string) bool {
	running, valid := parseWindowVersion(title)
	return valid && compareVersion(running, [3]int{7, 4, 14}) >= 0
}

// acquireSingleInstance uses the last-error value returned by the SAME
// CreateMutexW call. Calling GetLastError via a separate syscall overwrites
// the thread's last-error state and can allow several primary app processes.
func acquireSingleInstance(background bool) (uintptr, bool) {
	// Keep the mutex shared with all v7.4 and v7.5 builds in this session.
	name := ptr("Local\\PulseNoteDeck_8_12_Main_v74")

	// Older releases affected by the bug can already have multiple copies
	// running. Shut down one older instance at a time, then try the mutex
	// again until no older instance still holds it.
	for replacement := 0; replacement < 10; replacement++ {
		mutex, _, createErr := createMutex.Call(0, 0, name)
		if mutex == 0 {
			return 0, false
		}
		if !mutexAlreadyExists(createErr) {
			return mutex, true
		}
		closeHandle.Call(mutex)
		if background {
			return 0, false
		}

		// The mutex may be created shortly before the window is registered.
		var hwnd uintptr
		for i := 0; i < 25; i++ {
			hwnd = runningWindow()
			if hwnd != 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if hwnd == 0 {
			messageBox.Call(0, ptr("기존 PulseNote Deck가 실행 중이지만 응답 창을 찾지 못했습니다. 잠시 후 다시 실행하세요."), ptr("PulseNote Deck"), 0x30)
			return 0, false
		}
		title := runningWindowTitle(hwnd)
		if instanceAction(title, version, false) == InstanceShow {
			// Reuse the UI thread of the already running instance.
			ok, _, _ := postMessage.Call(hwnd, WM_REOPEN_INSTANCE, 0, 0)
			if ok == 0 {
				showWindow.Call(hwnd, SW_RESTORE)
				postMessage.Call(hwnd, WM_USER_STATUS, 0, 0)
				setForeground.Call(hwnd)
			}
			return 0, false
		}
		if err := requestRunningExit(hwnd, title); err != nil {
			messageBox.Call(0, ptr("기존 버전 자동 전환 실패: "+err.Error()+"\n기존 프로그램을 종료한 뒤 다시 실행해 주세요."), ptr("PulseNote Deck"), 0x30)
			return 0, false
		}
	}
	messageBox.Call(0, ptr("기존 PulseNote Deck 프로세스가 여러 개 남아 있습니다. 종료가 완료되지 않아 새 앱 시작을 중단했습니다."), ptr("PulseNote Deck"), 0x30)
	return 0, false
}
