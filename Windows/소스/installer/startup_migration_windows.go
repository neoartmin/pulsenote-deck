//go:build windows

package main

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

func migrateExistingStartup(dir, target string) error {
	const keyName = `Software\Microsoft\Windows\CurrentVersion\Run`
	const entry = "PulseNoteDeckUnified"
	var h syscall.Handle
	e := syscall.RegOpenKeyEx(syscall.HKEY_CURRENT_USER, w(keyName), 0, syscall.KEY_QUERY_VALUE|syscall.KEY_SET_VALUE, &h)
	if e == syscall.ERROR_FILE_NOT_FOUND {
		return nil
	}
	if e != nil {
		return fmt.Errorf("자동 시작 레지스트리 확인 실패: %w", e)
	}
	defer syscall.RegCloseKey(h)
	var dataType, count uint32
	name := w(entry)
	if e = syscall.RegQueryValueEx(h, name, nil, &dataType, nil, &count); e == syscall.ERROR_FILE_NOT_FOUND {
		return nil
	}
	if e != nil {
		return e
	}
	if dataType != syscall.REG_SZ || count < 2 || count > 8192 || count%2 != 0 {
		return nil
	}
	data := make([]byte, count)
	if e = syscall.RegQueryValueEx(h, name, nil, &dataType, &data[0], &count); e != nil {
		return e
	}
	words := make([]uint16, count/2)
	for i := range words {
		words[i] = uint16(data[2*i]) | uint16(data[2*i+1])<<8
	}
	previous := syscall.UTF16ToString(words)
	if !existingStartupBelongsToInstall(previous, dir) {
		return nil
	}
	current := `"` + filepath.Clean(target) + `" --background`
	if previous == current {
		return nil
	}
	encoded, e := syscall.UTF16FromString(current)
	if e != nil {
		return e
	}
	code, _, _ := regSet.Call(uintptr(h), uintptr(unsafe.Pointer(name)), 0, 1, uintptr(unsafe.Pointer(&encoded[0])), uintptr(len(encoded)*2))
	if code != 0 {
		return fmt.Errorf("자동 시작 경로 교체 실패: %w", syscall.Errno(code))
	}
	return nil
}
