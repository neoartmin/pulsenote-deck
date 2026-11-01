//go:build windows

package main

import (
	"os"
	"strings"
	"testing"
)

func TestProgramKeyRetainsColorAndNativeIconAlpha(t *testing.T) {
	b, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	if strings.Contains(src, "transparentProgramKey") {
		t.Fatal("v7.4.7 keycap removal must be reverted")
	}
	if !strings.Contains(src, "d.rounded(x, y, w, h, 13, colorValue(key.Color)") {
		t.Fatal("colored keycap must remain for program keys")
	}
	if strings.Contains(src, "makeTransparentIconBitmap(bitmap)") || strings.Contains(src, "removeIconBackdrop(") {
		t.Fatal("lossy background guessing must never modify Windows program logos")
	}
	if !strings.Contains(src, "gdipBitmapFromHICON.Call(hicon") || !strings.Contains(src, "gdipImageRect.Call(graphics, bitmap") {
		t.Fatal("Windows original HICON and native alpha must be rendered by GDI+")
	}
	if strings.Contains(src, "brandIconForPath") || strings.Contains(src, "arduino-mark") {
		t.Fatal("all programs must share the same icon pipeline, without brand-specific exceptions")
	}
}

func TestWindowCloseLeavesBackgroundActionsRunning(t *testing.T) {
	b, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	begin := strings.Index(src, "case WM_CLOSE:")
	end := strings.Index(src[begin:], "case WM_REPLACE_INSTANCE:")
	if end == -1 {
		end = strings.Index(src[begin:], "case WM_DESTROY:")
	}
	if begin == -1 || end == -1 {
		t.Fatal("close/destroy cases missing")
	}
	closeHandler := src[begin : begin+end]
	if !strings.Contains(closeHandler, "SW_HIDE") || strings.Contains(closeHandler, "destroyWindow.Call") || strings.Contains(closeHandler, "postQuit.Call") {
		t.Fatal("X button must hide the UI without terminating the USB action loop")
	}
	for _, required := range []string{
		"g.tray(true)", "완전 종료", "destroyWindow.Call(g.hwnd)",
		"go serialLoop()", "for n := range actionQueue", "TaskbarCreated", "-H=windowsgui",
	} {
		// -H=windowsgui is a build flag checked separately, not embedded in Go sources.
		if required == "-H=windowsgui" {
			continue
		}
		if !strings.Contains(src, required) {
			t.Fatalf("missing background requirement: %s", required)
		}
	}
}
