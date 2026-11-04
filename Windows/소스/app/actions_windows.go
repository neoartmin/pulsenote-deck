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
	return shortcutRespectingHolds(value)
}
func runAction(k Key) error {
	switch k.Type {
	case "none", "":
		return nil
	case "shortcut":
		return shortcut(k.Value)
	case "text":
		return sendTextMacro(k.Value)
	case "capture":
		// Windows snipping UI: user selects the region, then it is copied to clipboard.
		return shortcut("WIN+SHIFT+S")
	case "copy_text":
		return copyMacroToClipboard(k.Value)
	case "system_settings":
		return shellOpen(settingTarget(k.Value))
	case "open_alarms":
		// Microsoft's Clock application offers both Timer and Alarm tabs.
		// There is no documented stable launch URI for a specific Clock tab.
		return shellOpen(`shell:AppsFolder\Microsoft.WindowsAlarms_8wekyb3d8bbwe!App`)
	case "auto_click":
		return toggleAutoClick(k.Value, k.Label)
	case "audio_output":
		return switchAudioOutput(k.Value)
	case "open_timers":
		return fmt.Errorf("타이머 기능은 제거되었습니다. 다른 기능을 선택하세요")
	case "web_search":
		uri, err := webSearchURL(k.Value)
		if err != nil { return err }
		return shellOpen(uri)
	case "compose_email":
		uri, err := emailDraftURI(k.Value)
		if err != nil { return err }
		return shellOpen(uri)
	case "qr_code":
		return showOfflineQRCode(k.Value)
	case "clipboard_history":
		return shortcut("WIN+V")
	case "window_control":
		switch k.Value {
		case "maximize": return shortcut("WIN+UP")
		case "minimize":
			// Minimize active window directly; Win+Down merely restores maximized windows.
			hwnd, _, _ := u.NewProc("GetForegroundWindow").Call()
			if hwnd == 0 { return fmt.Errorf("활성 창이 없습니다") }
			result, _, _ := u.NewProc("ShowWindow").Call(hwnd, 6) // SW_MINIMIZE
			_ = result
			return nil
		case "left": return shortcut("WIN+LEFT")
		case "right": return shortcut("WIN+RIGHT")
		case "desktop": return shortcut("WIN+D")
		case "desk_prev": return shortcut("WIN+CTRL+LEFT")
		case "desk_next": return shortcut("WIN+CTRL+RIGHT")
		default: return fmt.Errorf("창 제어 동작을 선택하세요")
		}
	case "sequence":
		return fmt.Errorf("연속 동작은 장치/프로필 문맥에서 실행해야 합니다")
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
