//go:build windows

// Windows per-user installer. NO Run key, scheduled task, PowerShell, or autolaunch.
package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

//go:embed app.bin
var appBytes []byte

//go:embed icon.ico
var iconBytes []byte

const productName = "PulseNote Deck"

var udll = syscall.NewLazyDLL("user32.dll")
var ole = syscall.NewLazyDLL("ole32.dll")
var shell = syscall.NewLazyDLL("shell32.dll")
var reg = syscall.NewLazyDLL("advapi32.dll")
var msgBox = udll.NewProc("MessageBoxW")
var coinit = ole.NewProc("CoInitializeEx")
var couninit = ole.NewProc("CoUninitialize")
var cocreate = ole.NewProc("CoCreateInstance")
var cofree = ole.NewProc("CoTaskMemFree")
var knownFolder = shell.NewProc("SHGetKnownFolderPath")
var regCreate = reg.NewProc("RegCreateKeyExW")
var regSet = reg.NewProc("RegSetValueExW")
var regDel = reg.NewProc("RegDeleteKeyW")
var moveFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")

func w(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func alert(s string, typ uintptr) uintptr {
	r, _, _ := msgBox.Call(0, uintptr(unsafe.Pointer(w(s))), uintptr(unsafe.Pointer(w(productName))), typ)
	return r
}

// GUID layout matches Win32 GUID exactly.
type guid struct {
	D1     uint32
	D2, D3 uint16
	D4     [8]byte
}

var clsidShellLink = guid{0x00021401, 0, 0, [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46}}
var iidShellLinkW = guid{0x000214f9, 0, 0, [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46}}
var iidPersistFile = guid{0x0000010b, 0, 0, [8]byte{0xc0, 0, 0, 0, 0, 0, 0, 0x46}}
var folderDesktop = guid{0xB4BFCC3A, 0xDB2C, 0x424C, [8]byte{0xB0, 0x29, 0x7F, 0xE9, 0x9A, 0x87, 0xC6, 0x41}}
var folderPrograms = guid{0xA77F5D77, 0x2E2B, 0x44C3, [8]byte{0xA6, 0xA2, 0xAB, 0xA6, 0x01, 0x05, 0x4A, 0x51}}

func method(obj unsafe.Pointer, index int) uintptr {
	table := *(*unsafe.Pointer)(obj)
	return *(*uintptr)(unsafe.Add(table, uintptr(index)*unsafe.Sizeof(uintptr(0))))
}
func comRelease(obj unsafe.Pointer) {
	if obj != nil {
		syscall.SyscallN(method(obj, 2), uintptr(obj))
	}
}
func hrOk(hr uintptr) bool { return uint32(hr)&0x80000000 == 0 }
func shortcut(target, icon, link string) error {
	hr, _, _ := coinit.Call(0, 2)
	if hrOk(hr) {
		defer couninit.Call()
	}
	var obj unsafe.Pointer
	h, _, _ := cocreate.Call(uintptr(unsafe.Pointer(&clsidShellLink)), 0, 1, uintptr(unsafe.Pointer(&iidShellLinkW)), uintptr(unsafe.Pointer(&obj)))
	if !hrOk(h) || obj == nil {
		return fmt.Errorf("ShellLink COM 생성 오류: 0x%x", uint32(h))
	}
	defer comRelease(obj)
	// IShellLinkW: SetPath=20; SetIconLocation=17; SetWorkingDirectory=9; SetDescription=7.
	for _, spec := range []struct {
		idx  int
		args []uintptr
	}{{20, []uintptr{uintptr(unsafe.Pointer(w(target)))}}, {17, []uintptr{uintptr(unsafe.Pointer(w(icon))), 0}}, {9, []uintptr{uintptr(unsafe.Pointer(w(filepath.Dir(target))))}}, {7, []uintptr{uintptr(unsafe.Pointer(w("PulseNote Deck PN12 / PN15 설정 및 실행")))}}} {
		ret, _, _ := syscall.SyscallN(method(obj, spec.idx), append([]uintptr{uintptr(obj)}, spec.args...)...)
		if !hrOk(ret) {
			return fmt.Errorf("ShellLink 설정 실패: 0x%x", uint32(ret))
		}
	}
	var persist unsafe.Pointer
	ret, _, _ := syscall.SyscallN(method(obj, 0), uintptr(obj), uintptr(unsafe.Pointer(&iidPersistFile)), uintptr(unsafe.Pointer(&persist)))
	if !hrOk(ret) || persist == nil {
		return fmt.Errorf("IPersistFile 오류: 0x%x", uint32(ret))
	}
	defer comRelease(persist)
	ret, _, _ = syscall.SyscallN(method(persist, 6), uintptr(persist), uintptr(unsafe.Pointer(w(link))), 1)
	if !hrOk(ret) {
		return fmt.Errorf("바로가기 저장 오류: 0x%x", uint32(ret))
	}
	if _, e := os.Stat(link); e != nil {
		return fmt.Errorf("바로가기 검증 실패: %w", e)
	}
	return nil
}
func known(g *guid) (string, error) {
	var p *uint16
	ret, _, _ := knownFolder.Call(uintptr(unsafe.Pointer(g)), 0, 0, uintptr(unsafe.Pointer(&p)))
	if !hrOk(ret) || p == nil {
		return "", fmt.Errorf("KnownFolder 실패: 0x%x", uint32(ret))
	}
	defer cofree.Call(uintptr(unsafe.Pointer(p)))
	s := []uint16{}
	for i := 0; i < 32000; i++ {
		c := *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + uintptr(i)*2))
		if c == 0 {
			break
		}
		s = append(s, c)
	}
	return syscall.UTF16ToString(s), nil
}
func installDir() (string, error) {
	a := os.Getenv("LOCALAPPDATA")
	if a == "" {
		return "", fmt.Errorf("LOCALAPPDATA를 찾지 못했습니다")
	}
	return filepath.Join(a, "Programs", "PulseNoteDeckUnified"), nil
}
func regString(key syscall.Handle, name, val string) error {
	b, _ := syscall.UTF16FromString(val)
	ret, _, _ := regSet.Call(uintptr(key), uintptr(unsafe.Pointer(w(name))), 0, 1, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)*2))
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}
func registerUninstall(dir string) error {
	var h syscall.Handle
	key := `Software\Microsoft\Windows\CurrentVersion\Uninstall\PulseNoteDeckUnified`
	var disp uint32
	ret, _, _ := regCreate.Call(uintptr(syscall.HKEY_CURRENT_USER), uintptr(unsafe.Pointer(w(key))), 0, 0, 0, uintptr(syscall.KEY_SET_VALUE), 0, uintptr(unsafe.Pointer(&h)), uintptr(unsafe.Pointer(&disp)))
	if ret != 0 {
		return syscall.Errno(ret)
	}
	defer syscall.RegCloseKey(h)
	props := map[string]string{"DisplayName": productName, "DisplayVersion": productVersion, "Publisher": "PulseNote", "InstallLocation": dir, "DisplayIcon": filepath.Join(dir, "PulseNoteDeck.ico"), "UninstallString": `"` + filepath.Join(dir, "Uninstall.exe") + `" --uninstall`}
	for n, v := range props {
		if e := regString(h, n, v); e != nil {
			return e
		}
	}
	return nil
}
func install() error {
	if alert("PulseNote Deck v7.5.8로 업데이트할까요?\n\n실행 중인 기존 버전은 자동으로 종료한 후 새 버전으로 교체합니다.\n설치가 끝나면 새 버전이 자동으로 실행됩니다.\n기존 버튼 설정과 프로필은 보존합니다.\nWindows 로그인 자동 시작은 새로 등록하지 않습니다.\n관리자 권한은 필요하지 않습니다.", 0x24) != 6 {
		return nil
	}
	dir, e := installDir()
	if e != nil {
		return e
	}
	if e = stopPreviousApp(); e != nil {
		return fmt.Errorf("자동 업데이트를 위한 기존 앱 종료 실패: %w", e)
	}
	if e = os.MkdirAll(dir, 0755); e != nil {
		return e
	}
	for _, it := range []struct {
		name string
		data []byte
	}{{appFilename(), appBytes}, {"PulseNoteDeck.ico", iconBytes},
		{filepath.Join("Firmware", "PN12", "PulseNoteDeck_C3_12Key.ino"), firmwarePN12},
		{filepath.Join("Firmware", "PN15", "RIGHT", "PulseNoteDeck_C3_PN15_RIGHT", "PulseNoteDeck_C3_PN15_RIGHT.ino"), firmwarePN15},
		{filepath.Join("Firmware", "PN15", "LEFT", "PulseNoteDeck_C3_PN15_LEFT", "PulseNoteDeck_C3_PN15_LEFT.ino"), firmwarePN15Left},
		{filepath.Join("Firmware", "ESP32_C3_배선.txt"), firmwareWiring}} {
		destination := filepath.Join(dir, it.name)
		if e = os.MkdirAll(filepath.Dir(destination), 0755); e != nil {
			return fmt.Errorf("%s 폴더 생성 실패: %w", it.name, e)
		}
		if e = writeFile(destination, it.data); e != nil {
			return fmt.Errorf("%s 설치 실패: %w", it.name, e)
		}
	}
	self, e := os.Executable()
	if e != nil {
		return e
	}
	binary, e := os.ReadFile(self)
	if e != nil {
		return e
	}
	if e = writeFile(filepath.Join(dir, "Uninstall.exe"), binary); e != nil {
		return e
	}
	target := filepath.Join(dir, appFilename())
	ico := filepath.Join(dir, "PulseNoteDeck.ico")
	desk, e := known(&folderDesktop)
	if e != nil {
		return e
	}
	programs, e := known(&folderPrograms)
	if e != nil {
		return e
	}
	if e = shortcut(target, ico, filepath.Join(desk, productName+".lnk")); e != nil {
		return e
	}
	if e = shortcut(target, ico, filepath.Join(programs, productName+".lnk")); e != nil {
		return e
	}
	if e = registerUninstall(dir); e != nil {
		return e
	}
	// Keep a user's opted-in login autostart working after the executable
	// changes its name. Never create an autostart entry without consent.
	startupWarning := ""
	if err := migrateExistingStartup(dir, target); err != nil {
		startupWarning = "\n로그인 자동 시작 경로 갱신 실패: " + err.Error()
	}
	// Remove obsolete PN8 firmware from previous installs. Do not delete
	// user profiles or other files within this directory.
	if err := os.Remove(filepath.Join(dir, "Firmware", "PN8", "PulseNoteDeck_C3_8Key.ino")); err == nil || os.IsNotExist(err) {
		_ = os.Remove(filepath.Join(dir, "Firmware", "PN8"))
	}
	// Remove previous desktop and Start Menu shortcuts after the new ones exist.
	_ = os.Remove(filepath.Join(desk, "PulseNote Deck 8-12.lnk"))
	_ = os.Remove(filepath.Join(programs, "PulseNote Deck 8-12.lnk"))
	// Remove previous version binaries only when Windows releases them.
	// Old running processes / antivirus may still lock a legacy binary;
	// cleanup must never undo an otherwise successful update.
	var leftovers []string
	for _, oldName := range installedAppFilenames(dir) {
		if strings.EqualFold(oldName, appFilename()) {
			continue
		}
		oldPath := filepath.Join(dir, oldName)
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			leftovers = append(leftovers, oldName)
		}
	}
	// Start the new version now; no extra click or tray-menu exit needed.
	if e = exec.Command(target).Start(); e != nil {
		alert("v7.5.2 설치 완료. 자동 실행은 실패했습니다.\n바탕화면의 PulseNote Deck 아이콘으로 실행하세요.\n"+e.Error(), 0x30)
		return nil
	}
	notice := "v7.5.2 설치 완료 및 실행\n\n기존 설정과 프로필을 유지했습니다.\nWindows 로그인 자동 시작을 임의로 등록하지 않았습니다." + startupWarning
	if len(leftovers) != 0 {
		notice += "\n\n이전 버전 실행 파일을 잠금 때문에 정리하지 못했습니다: " + strings.Join(leftovers, ", ") + "\n다음 업데이트 또는 제거 시 다시 정리합니다."
	}
	alert(notice, 0x40)
	return nil
}
func uninstall() error {
	if alert("PulseNote Deck PN12/PN15를 제거할까요?\n\n저장된 사용자 설정과 프로필은 유지됩니다.", 0x24) != 6 {
		return nil
	}
	dir, e := installDir()
	if e != nil {
		return e
	}
	// First release executable locks; never report success after a failed delete.
	if e = stopPreviousApp(); e != nil {
		return fmt.Errorf("제거 전 실행 중인 앱 종료 실패: %w", e)
	}
	for _, nm := range append(installedAppFilenames(dir), "PulseNoteDeck.ico",
		filepath.Join("Firmware", "PN12", "PulseNoteDeck_C3_12Key.ino"),
		filepath.Join("Firmware", "PN15", "RIGHT", "PulseNoteDeck_C3_PN15_RIGHT", "PulseNoteDeck_C3_PN15_RIGHT.ino"),
		filepath.Join("Firmware", "PN15", "LEFT", "PulseNoteDeck_C3_PN15_LEFT", "PulseNoteDeck_C3_PN15_LEFT.ino"),
		filepath.Join("Firmware", "ESP32_C3_배선.txt")) {
		if e = removeInstalledFile(filepath.Join(dir, nm)); e != nil {
			return e
		}
	}
	// Only remove the directories when empty; preserve any user-added files.
	_ = os.Remove(filepath.Join(dir, "Firmware", "PN8")) // legacy folder if empty
	_ = os.Remove(filepath.Join(dir, "Firmware", "PN15", "LEFT", "PulseNoteDeck_C3_PN15_LEFT"))
	_ = os.Remove(filepath.Join(dir, "Firmware", "PN15", "RIGHT", "PulseNoteDeck_C3_PN15_RIGHT"))
	_ = os.Remove(filepath.Join(dir, "Firmware", "PN15", "LEFT"))
	_ = os.Remove(filepath.Join(dir, "Firmware", "PN15", "RIGHT"))
	_ = os.Remove(filepath.Join(dir, "Firmware", "PN15"))
	_ = os.Remove(filepath.Join(dir, "Firmware", "PN12"))
	_ = os.Remove(filepath.Join(dir, "Firmware"))
	if desk, lookupErr := known(&folderDesktop); lookupErr == nil {
		if e = removeInstalledFile(filepath.Join(desk, productName+".lnk")); e != nil {
			return e
		}
	} else {
		return fmt.Errorf("바탕화면 바로가기 경로 확인 실패: %w", lookupErr)
	}
	if programs, lookupErr := known(&folderPrograms); lookupErr == nil {
		if e = removeInstalledFile(filepath.Join(programs, productName+".lnk")); e != nil {
			return e
		}
	} else {
		return fmt.Errorf("시작 메뉴 바로가기 경로 확인 실패: %w", lookupErr)
	}
	// Schedule removal of the running uninstaller only after the other files
	// are removed successfully. The schedule itself is checked for failure.
	self, e := os.Executable()
	if e != nil {
		return e
	}
	moved, _, moveErr := moveFileEx.Call(uintptr(unsafe.Pointer(w(self))), 0, 4)
	// A per-user installer may not be allowed to schedule a reboot-time file
	// deletion on every Windows configuration. The program can still be
	// uninstalled; report the residual uninstaller instead of falsely claiming
	// that it was deleted or preserving a broken Add/Remove Programs entry.
	selfRemovalWarning := ""
	if moved == 0 {
		selfRemovalWarning = fmt.Sprintf("\n\n제거 프로그램 자체(%s)는 자동 정리하지 못했습니다: %v\n이 파일만 수동 삭제할 수 있습니다.", self, moveErr)
	}
	// Remove opted-in Windows startup only after the app is stopped.
	var key syscall.Handle
	if e := syscall.RegOpenKeyEx(syscall.HKEY_CURRENT_USER, w(`Software\Microsoft\Windows\CurrentVersion\Run`), 0, syscall.KEY_SET_VALUE, &key); e == nil {
		deleteValue := reg.NewProc("RegDeleteValueW")
		code, _, _ := deleteValue.Call(uintptr(key), uintptr(unsafe.Pointer(w("PulseNoteDeckUnified"))))
		syscall.RegCloseKey(key)
		if code != 0 && code != 2 {
			return fmt.Errorf("로그인 자동 시작 등록 해제 실패: %w", syscall.Errno(code))
		}
	} else if e != syscall.ERROR_FILE_NOT_FOUND {
		return fmt.Errorf("로그인 자동 시작 등록 정보 접근 실패: %w", e)
	}
	result, _, _ := regDel.Call(uintptr(syscall.HKEY_CURRENT_USER), uintptr(unsafe.Pointer(w(`Software\Microsoft\Windows\CurrentVersion\Uninstall\PulseNoteDeckUnified`))))
	if result != 0 && result != 2 {
		return fmt.Errorf("제거 등록 정보 삭제 실패: %w", syscall.Errno(result))
	}

	if selfRemovalWarning != "" {
		alert("프로그램 파일과 바로가기를 제거했습니다. 사용자 설정은 보존합니다."+selfRemovalWarning, 0x30)
	} else {
		alert("프로그램 제거를 완료했습니다.\n\n사용자 설정은 보존했으며 제거 프로그램 자체는 Windows 재시작 후 정리됩니다.", 0x40)
	}
	return nil
}
func main() {
	var e error
	if len(os.Args) > 1 && os.Args[1] == "--uninstall" {
		e = uninstall()
	} else {
		e = install()
	}
	if e != nil {
		alert("작업 실패:\n"+e.Error(), 0x10)
	}
}
