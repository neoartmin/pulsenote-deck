//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// A bounded best-effort log allows a user to distinguish a COM disappearance
// from an unresponsive firmware without exposing unrelated device information.
func logUSBEvent(port, event string) {
	dir := filepath.Dir(getConfigPath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	path := filepath.Join(dir, "usb_diagnostics.log")
	if st, err := os.Stat(path); err == nil && st.Size() > 128*1024 {
		_ = os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	event = strings.ReplaceAll(strings.ReplaceAll(event, "\n", " "), "\r", " ")
	fmt.Fprintf(f, "%s %s %s\n", time.Now().Format(time.RFC3339), port, event)
}
