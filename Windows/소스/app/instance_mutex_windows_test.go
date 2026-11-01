//go:build windows

package main

import (
	"os"
	"strings"
	"testing"
)

func TestMutexAcquisitionDoesNotCallGetLastErrorSeparately(t *testing.T) {
	raw, err := os.ReadFile("instance_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	code := string(raw)
	if !strings.Contains(code, "mutex, _, createErr := createMutex.Call(0, 0, name)") {
		t.Fatal("named mutex must preserve CreateMutexW's returned last error")
	}
	if strings.Contains(code, "getLastError.Call(") {
		t.Fatal("separate GetLastError syscall is unreliable")
	}
	if !strings.Contains(code, "if !mutexAlreadyExists(createErr)") {
		t.Fatal("duplicate process detection does not use direct result")
	}
}
