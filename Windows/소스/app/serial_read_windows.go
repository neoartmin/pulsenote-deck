//go:build windows

package main

import (
	"os"
	"syscall"
)

// readCOM intentionally bypasses os.File.Read. Windows COMMTIMEOUTS can
// return 0 bytes, nil error when no serial data is available. os.File.Read
// converts that legitimate idle timeout to io.EOF; treating EOF as device
// removal caused immediate connect/disconnect loops while COM still existed.
// syscall.Read preserves (0, nil). Real device/driver errors still propagate.
func readCOM(f *os.File, buffer []byte) (int, error) {
	return syscall.Read(syscall.Handle(f.Fd()), buffer)
}
