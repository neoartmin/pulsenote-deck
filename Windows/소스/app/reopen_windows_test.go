//go:build windows

package main

import (
	"os"
	"strings"
	"testing"
)

func TestReopenDelegatesToLiveUI(t *testing.T) {
	b, e := os.ReadFile("instance_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	s := string(b)
	if !strings.Contains(s, "postMessage.Call(hwnd, WM_REOPEN_INSTANCE, 0, 0)") {
		t.Fatal("second launch must request UI-thread reopen")
	}
	b, e = os.ReadFile("gui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	s = string(b)
	for _, req := range []string{"case WM_REOPEN_INSTANCE:", "g.show()", "case WM_SHOWWINDOW:", "updateWindow.Call(g.hwnd)", "shouldRescanOnShow(port, connected, portList())"} {
		if !strings.Contains(s, req) {
			t.Fatal("missing reopen behavior: " + req)
		}
	}
}
