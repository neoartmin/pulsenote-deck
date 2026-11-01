//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var winAdv = syscall.NewLazyDLL("advapi32.dll")
var createReg = winAdv.NewProc("RegCreateKeyExW")
var regSet = winAdv.NewProc("RegSetValueExW")
var regDeleteValue = winAdv.NewProc("RegDeleteValueW")

const startupPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const startupEntry = "PulseNoteDeckUnified"

func startupEnabled() bool {
	var h syscall.Handle
	if e := syscall.RegOpenKeyEx(syscall.HKEY_CURRENT_USER, syscall.StringToUTF16Ptr(startupPath), 0, syscall.KEY_READ, &h); e != nil {
		return false
	}
	defer syscall.RegCloseKey(h)
	var typ, n uint32
	entry := syscall.StringToUTF16Ptr(startupEntry)
	if err := syscall.RegQueryValueEx(h, entry, nil, &typ, nil, &n); err != nil || typ != 1 || n < 2 || n > 8192 || n%2 != 0 {
		return false
	}
	data := make([]byte, n)
	if err := syscall.RegQueryValueEx(h, entry, nil, &typ, &data[0], &n); err != nil || typ != 1 || n < 2 {
		return false
	}
	words := make([]uint16, n/2)
	for i := range words {
		words[i] = uint16(data[i*2]) | uint16(data[i*2+1])<<8
	}
	registered := syscall.UTF16ToString(words)
	current, err := os.Executable()
	return err == nil && startupMatches(registered, current)
}

// Only invoked directly from the user-clicked settings switch, never on app startup.
func setStartup(on bool) error {
	var h syscall.Handle
	if !on {
		e := syscall.RegOpenKeyEx(syscall.HKEY_CURRENT_USER, syscall.StringToUTF16Ptr(startupPath), 0, syscall.KEY_SET_VALUE, &h)
		if e == syscall.ERROR_FILE_NOT_FOUND {
			return nil
		}
		if e != nil {
			return e
		}
		defer syscall.RegCloseKey(h)
		r, _, _ := regDeleteValue.Call(uintptr(h), ptr(startupEntry))
		if r != 0 && r != 2 {
			return fmt.Errorf("자동 시작 해제 실패: %d", r)
		}
		return nil
	}
	var disp uint32
	r, _, _ := createReg.Call(uintptr(syscall.HKEY_CURRENT_USER), ptr(startupPath), 0, 0, 0, syscall.KEY_SET_VALUE, 0, uintptr(unsafe.Pointer(&h)), uintptr(unsafe.Pointer(&disp)))
	if r != 0 {
		return fmt.Errorf("자동 시작 등록 실패: %d", r)
	}
	defer syscall.RegCloseKey(h)
	self, e := os.Executable()
	if e != nil {
		return e
	}
	value := `"` + self + `" --background`
	data, _ := syscall.UTF16FromString(value)
	r, _, _ = regSet.Call(uintptr(h), ptr(startupEntry), 0, 1, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)*2))
	if r != 0 {
		return fmt.Errorf("자동 시작 저장 실패: %d", r)
	}
	return nil
}
