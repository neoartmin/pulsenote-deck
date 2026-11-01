//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var keybdEvent = u.NewProc("keybd_event")
var shellExecute = sh32.NewProc("ShellExecuteW")

const (
	KEYEVENTF_KEYUP = 2
)

func shellOpen(target string) error {
	target = os.ExpandEnv(strings.TrimSpace(target))
	if target == "" {
		return fmt.Errorf("대상이 없습니다")
	}
	v, _, e := shellExecute.Call(0, ptr("open"), ptr(target), 0, 0, 1)
	if v <= 32 {
		return fmt.Errorf("ShellExecute 실패: %v (%d)", e, v)
	}
	return nil
}
func press(v byte, up bool) {
	flags := uintptr(0)
	if up {
		flags = KEYEVENTF_KEYUP
	}
	keybdEvent.Call(uintptr(v), 0, flags, 0)
}
func shortcut(value string) error {
	// Validate first; an unsupported trailing key must not inject any keys.
	codes, err := parseShortcut(value)
	if err != nil {
		return err
	}
	for _, code := range codes {
		press(code, false)
	}
	for i := len(codes) - 1; i >= 0; i-- {
		press(codes[i], true)
	}
	return nil
}
func runAction(k Key) error {
	switch k.Type {
	case "none", "":
		return nil
	case "shortcut":
		return shortcut(k.Value)
	case "program":
		path := os.ExpandEnv(strings.TrimSpace(k.Value))
		if path == "" {
			return fmt.Errorf("프로그램이 선택되지 않았습니다")
		}
		if strings.EqualFold(filepath.Ext(path), ".lnk") {
			return shellOpen(path)
		}
		if strings.EqualFold(filepath.Ext(path), ".exe") {
			cmd := exec.Command(path)
			cmd.Dir = filepath.Dir(path)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			return cmd.Start()
		}
		return shellOpen(path)
	case "website":
		if !strings.HasPrefix(k.Value, "https://") && !strings.HasPrefix(k.Value, "http://") {
			return fmt.Errorf("웹 주소는 https:// 또는 http:// 로 시작해야 합니다")
		}
		return shellOpen(k.Value)
	case "folder":
		return shellOpen(k.Value)
	case "volume_up":
		press(0xaf, false)
		press(0xaf, true)
		return nil
	case "volume_down":
		press(0xae, false)
		press(0xae, true)
		return nil
	case "volume_mute":
		press(0xad, false)
		press(0xad, true)
		return nil
	case "media_play":
		press(0xb3, false)
		press(0xb3, true)
		return nil
	case "media_next":
		press(0xb0, false)
		press(0xb0, true)
		return nil
	case "media_prev":
		press(0xb1, false)
		press(0xb1, true)
		return nil
	default:
		return fmt.Errorf("지원하지 않는 기능")
	}
}

var _ = unsafe.Pointer(nil)
