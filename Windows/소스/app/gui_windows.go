//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	WM_CREATE           = 1
	WM_DESTROY          = 2
	WM_SIZE             = 5
	WM_PAINT            = 15
	WM_ERASEBKGND       = 20
	WM_CLOSE            = 16
	WM_COMMAND          = 0x111
	WM_LBUTTONUP        = 0x202
	WM_RBUTTONUP        = 0x205
	WM_TIMER            = 0x113
	EDIT_SAVE_TIMER     = 0x72
	PROFILE_SAVE_TIMER  = 0x73
	WM_CTLCOLORSTATIC   = 0x138
	WM_CTLCOLOREDIT     = 0x133
	WM_APP              = 0x8000
	WM_USER_STATUS      = WM_APP + 10
	WM_SHOWWINDOW       = 0x0018
	WM_USER_APPS        = WM_APP + 13
	WM_TRAY             = WM_APP + 11
	WM_QUIT             = WM_APP + 12
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_CLIPCHILDREN     = 0x02000000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	ES_AUTOHSCROLL      = 0x80
	WS_BORDER           = 0x800000
	ES_READONLY         = 0x800
	BS_PUSHBUTTON       = 0
	SW_SHOW             = 5
	SW_HIDE             = 0
	SW_MAXIMIZE         = 3
	SW_RESTORE          = 9
	COLOR_WINDOW        = 5
	PS_SOLID            = 0
	SRCCOPY             = 0x00CC0020
	TRANSPARENT         = 1
	DT_CENTER           = 1
	DT_VCENTER          = 4
	DT_SINGLELINE       = 0x20
	DT_END_ELLIPSIS     = 0x8000
	CS_HREDRAW          = 2
	CS_VREDRAW          = 1
	EN_CHANGE           = 0x300
	BN_CLICKED          = 0
	EM_SETLIMITTEXT     = 0xC5
	WM_SETFONT          = 0x30
	WM_SETTEXT          = 12
	WM_GETTEXTLENGTH    = 14
	WM_GETTEXT          = 13
	EM_SETCUEBANNER     = 0x1501
	NIF_MESSAGE         = 1
	NIF_ICON            = 2
	NIF_TIP             = 4
	NIM_ADD             = 0
	NIM_DELETE          = 2
	NIM_MODIFY          = 1
	TPM_RETURNCMD       = 0x100
	TPM_RIGHTBUTTON     = 2
	IDC_ARROW           = 32512
	IDI_APPLICATION     = 32512
	MK_LBUTTON          = 1
	WM_LBUTTONDBLCLK    = 0x203
	WM_CONTEXTMENU      = 0x7B
)

var (
	dwm                    = syscall.NewLazyDLL("dwmapi.dll")
	setDWMAttr             = dwm.NewProc("DwmSetWindowAttribute")
	u                      = syscall.NewLazyDLL("user32.dll")
	gd                     = syscall.NewLazyDLL("gdi32.dll")
	k32                    = syscall.NewLazyDLL("kernel32.dll")
	sh32                   = syscall.NewLazyDLL("shell32.dll")
	defProc                = u.NewProc("DefWindowProcW")
	registerClass          = u.NewProc("RegisterClassExW")
	createWindow           = u.NewProc("CreateWindowExW")
	showWindow             = u.NewProc("ShowWindow")
	updateWindow           = u.NewProc("UpdateWindow")
	destroyWindow          = u.NewProc("DestroyWindow")
	postQuit               = u.NewProc("PostQuitMessage")
	postMessage            = u.NewProc("PostMessageW")
	getMessage             = u.NewProc("GetMessageW")
	registerWindowMessage  = u.NewProc("RegisterWindowMessageW")
	translateMessage       = u.NewProc("TranslateMessage")
	dispatchMessage        = u.NewProc("DispatchMessageW")
	getClientRect          = u.NewProc("GetClientRect")
	beginPaint             = u.NewProc("BeginPaint")
	endPaint               = u.NewProc("EndPaint")
	invalidateRect         = u.NewProc("InvalidateRect")
	setWindowText          = u.NewProc("SetWindowTextW")
	getWindowText          = u.NewProc("GetWindowTextW")
	sendMessage            = u.NewProc("SendMessageW")
	moveWindow             = u.NewProc("MoveWindow")
	enableWindow           = u.NewProc("EnableWindow")
	getCursorPos           = u.NewProc("GetCursorPos")
	setForeground          = u.NewProc("SetForegroundWindow")
	loadCursor             = u.NewProc("LoadCursorW")
	loadIcon               = u.NewProc("LoadIconW")
	setTimer               = u.NewProc("SetTimer")
	killTimer              = u.NewProc("KillTimer")
	isIconic               = u.NewProc("IsIconic")
	getWindowLong          = u.NewProc("GetWindowLongW")
	findWindow             = u.NewProc("FindWindowW")
	messageBox             = u.NewProc("MessageBoxW")
	createSolidBrush       = gd.NewProc("CreateSolidBrush")
	createPen              = gd.NewProc("CreatePen")
	selectObject           = gd.NewProc("SelectObject")
	deleteObject           = gd.NewProc("DeleteObject")
	roundRect              = gd.NewProc("RoundRect")
	setTextColor           = gd.NewProc("SetTextColor")
	setBkColor             = gd.NewProc("SetBkColor")
	setBkMode              = gd.NewProc("SetBkMode")
	drawText               = u.NewProc("DrawTextW")
	createFont             = gd.NewProc("CreateFontW")
	fillRect               = u.NewProc("FillRect")
	createCompatibleDC     = gd.NewProc("CreateCompatibleDC")
	createCompatibleBitmap = gd.NewProc("CreateCompatibleBitmap")
	bitBlt                 = gd.NewProc("BitBlt")
	deleteDC               = gd.NewProc("DeleteDC")
	createMenu             = u.NewProc("CreatePopupMenu")
	appendMenu             = u.NewProc("AppendMenuW")
	trackPopup             = u.NewProc("TrackPopupMenu")
	destroyMenu            = u.NewProc("DestroyMenu")
	notifyIcon             = sh32.NewProc("Shell_NotifyIconW")
	getFileInfo            = sh32.NewProc("SHGetFileInfoW")
	drawIconEx             = u.NewProc("DrawIconEx")
	shGetImageList         = sh32.NewProc("SHGetImageList")
	ole32                  = syscall.NewLazyDLL("ole32.dll")
	coInit                 = ole32.NewProc("CoInitializeEx")
	coUninit               = ole32.NewProc("CoUninitialize")
	createRoundRectRgn     = gd.NewProc("CreateRoundRectRgn")
	selectClipRgn          = gd.NewProc("SelectClipRgn")
	saveDC                 = gd.NewProc("SaveDC")
	restoreDC              = gd.NewProc("RestoreDC")
	destroyIcon            = u.NewProc("DestroyIcon")
	getModuleHandle        = k32.NewProc("GetModuleHandleW")
	createMutex            = k32.NewProc("CreateMutexW")
)

type rect struct{ Left, Top, Right, Bottom int32 }
type point struct{ X, Y int32 }
type msg struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             point
	private        uint32
}
type paintstruct struct {
	Hdc                uintptr
	Erase              int32
	Rc                 rect
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type wndclassex struct {
	Size                                                       uint32
	Style                                                      uint32
	WndProc                                                    uintptr
	ClsExtra, WndExtra                                         int32
	Instance, Icon, Cursor, Brush, MenuName, ClassName, IconSm uintptr
}
type nid struct {
	Size                       uint32
	Hwnd                       uintptr
	ID, Flags, CallbackMessage uint32
	Icon                       uintptr
	Tip                        [128]uint16
	State, StateMask           uint32
	Info                       [256]uint16
	Version                    uint32
	InfoTitle                  [64]uint16
	InfoFlags                  uint32
	Guid                       [16]byte
	BalloonIcon                uintptr
}
type fileinfo struct {
	Icon        uintptr
	IconIndex   int32
	Attr        uint32
	DisplayName [260]uint16
	TypeName    [80]uint16
}
type drawCtx struct {
	graphics  uintptr
	hdc       uintptr
	scale     float64
	ox, oy    int
	fonts     map[int]uintptr
	iconFonts map[int]uintptr
}
type click struct {
	r    rect
	name string
	id   int
}
type ui struct {
	hwnd                                                 uintptr
	labelEdit, valueEdit, appSearchEdit, profileNameEdit uintptr
	editing                                              bool
	selected                                             int
	page                                                 int
	status                                               string
	scale                                                float64
	ox, oy                                               int
	touches                                              []click
	fonts                                                map[int]uintptr
	iconFonts                                            map[int]uintptr
	icons                                                map[string]uintptr
	appBitmaps                                           map[string]uintptr
	appIconOrder                                         []string
	apps                                                 []foundApp
	appReady                                             bool
	appScanID                                            uint32
	appResults                                           chan scanResult
	appPage                                              int
	appQuery                                             string
	searchRect                                           rect
	searchPlaced                                         bool
	searchVisible                                        bool
	profileNameRect                                      rect
	profileNamePlaced, profileNameVisible                bool
	profilePending                                       bool
	profilePendingName                                   string
	profilePendingModel, profilePendingIndex             int
	modelMode                                            int
	modelSide                                            string
	pendingSide                                          string
	profilePendingSide                                   string
	mutex                                                uintptr
	editsVisible                                         bool
	labelRect, valueRect                                 rect
	editsPlaced                                          bool
	pendingEdits                                         bool
	pendingLabel                                         string
	pendingValue                                         string
	pendingModel                                         int
	pendingProfile                                       int
	pendingIndex                                         int
}

type scanResult struct {
	ID   uint32
	Apps []foundApp
}

var mainUI *ui
var state = struct {
	sync.RWMutex
	Port      string
	Info      int
	Connected bool
	LastKey   int
	Message   string
	Since     time.Time
	Progress  connectionProgress
}{}

func refresh() {
	if mainUI != nil && mainUI.hwnd != 0 {
		postMessage.Call(mainUI.hwnd, WM_USER_STATUS, 0, 0)
	}
}
func setState(port string, info int, connected bool, message string) {
	now := time.Now()
	state.Lock()
	changed := state.Port != port || state.Info != info || state.Connected != connected || state.Message != message
	state.Port = port
	state.Info = info
	state.Connected = connected
	state.Message = message
	state.Since = now
	if connected {
		state.Progress = connectionProgress{Phase: phaseConnected, Port: port, Keys: info, Detail: fmt.Sprintf("PN%d 펌웨어 응답 정상", info), CheckedAt: now}
	} else {
		state.Progress = connectionProgress{Phase: phaseSearching, Detail: message, CheckedAt: now}
	}
	state.Unlock()
	if changed {
		refresh()
	}
}

// Called from the USB worker only. An unverified COM port is never reported
// as connected. All user-visible progress shares the same state lock.
func setConnectionProgress(phase connectionPhase, port, detail string, retryAt time.Time) {
	now := time.Now()
	state.Lock()
	old := state.Progress
	state.Progress = connectionProgress{Phase: phase, Port: port, Detail: detail, CheckedAt: now, RetryAt: retryAt}
	state.Message = detail
	changed := old.Phase != phase || old.Port != port || old.Detail != detail || now.Sub(old.CheckedAt) >= time.Second
	if phase != phaseConnected {
		changed = changed || state.Connected
		state.Port = ""
		state.Info = 0
		state.Connected = false
	}
	state.Unlock()
	if changed {
		refresh()
	}
}

func connectionProgressSnap() connectionProgress {
	state.RLock()
	defer state.RUnlock()
	return state.Progress
}

func connectionStateSnap() (string, int, bool, int, string, connectionProgress) {
	state.RLock()
	defer state.RUnlock()
	return state.Port, state.Info, state.Connected, state.LastKey, state.Message, state.Progress
}

// Refresh the last successful CDC response time without causing a false
// disconnect/reconnect or generating notification balloons.
func recordConnectionReply() {
	now := time.Now()
	state.Lock()
	if !state.Connected {
		state.Unlock()
		return
	}
	state.Progress.CheckedAt = now
	state.Unlock()
	refresh()
}

func stateSnap() (string, int, bool, int, string) {
	state.RLock()
	defer state.RUnlock()
	return state.Port, state.Info, state.Connected, state.LastKey, state.Message
}
func keyInput(i int) {
	c := snap()
	key, ok := selectedKey(c, i-1)
	if !ok {
		return
	}
	var err error
	if key.Type != "none" {
		err = runAction(key)
	}
	s := fmt.Sprintf("KEY:%d · %s", i, key.Label)
	if key.Type == "none" {
		s = fmt.Sprintf("KEY:%d · 기능 없음", i)
	}
	if err != nil {
		s = fmt.Sprintf("KEY:%d · 실행 실패: %v", i, err)
	}
	state.Lock()
	state.LastKey = i
	state.Message = s
	state.Unlock()
	refresh()
}

// Route each hardware key to its own PN12, PN15-L, or PN15-R settings.
// NEVER look up a USB event using the currently displayed model.
func keyInputForDevice(e deviceKeyEvent) {
	if e.Key < 1 || e.Key > e.Model {
		return
	}
	c := snap()
	key, ok := resolveHardwareKey(c, e.Model, e.Side, e.Key)
	if !ok {
		return
	}
	var err error
	if key.Type != "none" {
		err = runAction(key)
	}
	result := fmt.Sprintf("%s KEY:%d · %s", deviceIdentityLabel(e.Model, e.Side), e.Key, key.Label)
	if key.Type == "none" {
		result = fmt.Sprintf("%s KEY:%d · 기능 없음", deviceIdentityLabel(e.Model, e.Side), e.Key)
	}
	if err != nil {
		result = fmt.Sprintf("%s KEY:%d · 실행 실패: %v", deviceIdentityLabel(e.Model, e.Side), e.Key, err)
	}
	// Device input remains independent of the offline editor selection.
	state.Lock()
	state.LastKey = e.Key
	state.Message = result
	state.Unlock()
	refresh()
}

// On supported Windows versions, match the native title bar to the navy UI.
// Ignore unsupported DWM attributes on earlier systems.
func enableDarkTitlebar(hwnd uintptr) {
	one := uint32(1)
	for _, attr := range []uintptr{20, 19} {
		hr, _, _ := setDWMAttr.Call(hwnd, attr, uintptr(unsafe.Pointer(&one)), unsafe.Sizeof(one))
		if int32(hr) >= 0 {
			break
		}
	}
}
func ptr(s string) uintptr                 { p, _ := syscall.UTF16PtrFromString(s); return uintptr(unsafe.Pointer(p)) }
func uint16copy(target []uint16, s string) { r, _ := syscall.UTF16FromString(s); copy(target, r) }
func rectI(x, y, w, h int32) rect          { return rect{x, y, x + w, y + h} }
func RGB(r, g, b byte) uint32              { return uint32(r) | uint32(g)<<8 | uint32(b)<<16 }

// Crisp graphite/navy shell with vivid focused accents, matching the approved design.
var paletteTheme = map[string]uint32{
	"bg": RGB(8, 15, 29), "sidebar": RGB(12, 22, 40), "panel": RGB(19, 33, 55),
	"panel2": RGB(13, 26, 47), "border": RGB(40, 60, 90),
	"text": RGB(246, 249, 255), "muted": RGB(163, 180, 210),
	"accent": RGB(85, 102, 255), "sel": RGB(37, 53, 115),
	"green": RGB(3, 220, 121), "blacktext": RGB(10, 15, 24),
}

func theme(s string) uint32 { return paletteTheme[s] }

var preferredFace = "맑은 고딕" // Always available on Korean Windows; Pretendard if installed.

// Check user and machine font registrations. No font file is bundled or copied.
func chooseUIFont() string {
	for _, root := range []syscall.Handle{syscall.HKEY_CURRENT_USER, syscall.HKEY_LOCAL_MACHINE} {
		var h syscall.Handle
		key, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows NT\CurrentVersion\Fonts`)
		if syscall.RegOpenKeyEx(root, key, 0, syscall.KEY_READ, &h) != nil {
			continue
		}
		for i := uint32(0); i < 4000; i++ {
			buf := make([]uint16, 300)
			length := uint32(len(buf))
			rc, _, _ := regEnumValue.Call(uintptr(h), uintptr(i), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&length)), 0, 0, 0, 0)
			if rc == 259 {
				break
			}
			if rc == 0 && int(length) < len(buf) && strings.Contains(strings.ToLower(syscall.UTF16ToString(buf[:length])), "pretendard") {
				syscall.RegCloseKey(h)
				return "Pretendard"
			}
		}
		syscall.RegCloseKey(h)
	}
	return "맑은 고딕"
}

func newDraw(hdc uintptr, w, h int32) *drawCtx {
	sx := float64(w) / 1440
	sy := float64(h) / 900
	scale := sx
	if sy < sx {
		scale = sy
	}
	// Keep the whole 1440x900 design visible on smaller restored windows.
	// Text may become small below 1000px; do not crop the settings panel.
	if scale < 0.4 {
		scale = 0.4
	}
	if scale > 1.65 {
		scale = 1.65
	}
	return &drawCtx{hdc: hdc, scale: scale, ox: int((float64(w) - 1440*scale) / 2), oy: int((float64(h) - 900*scale) / 2), fonts: mainUI.fonts, iconFonts: mainUI.iconFonts}
}
func (d *drawCtx) xy(x, y int) (int32, int32) {
	return int32(d.ox + int(float64(x)*d.scale)), int32(d.oy + int(float64(y)*d.scale))
}
func (d *drawCtx) ab(x, y, w, h int) rect {
	xx, yy := d.xy(x, y)
	return rectI(xx, yy, int32(float64(w)*d.scale), int32(float64(h)*d.scale))
}
func (d *drawCtx) rounded(x, y, w, h, r int, fill, border uint32) {
	q := d.ab(x, y, w, h)
	pen, _, _ := createPen.Call(PS_SOLID, uintptr(maxInt(1, int(d.scale))), uintptr(border))
	brush, _, _ := createSolidBrush.Call(uintptr(fill))
	oldPen, _, _ := selectObject.Call(d.hdc, pen)
	oldBrush, _, _ := selectObject.Call(d.hdc, brush)
	rr := uintptr(maxInt(3, int(float64(r)*d.scale))) * 2
	roundRect.Call(d.hdc, uintptr(q.Left), uintptr(q.Top), uintptr(q.Right), uintptr(q.Bottom), rr, rr)
	selectObject.Call(d.hdc, oldPen)
	selectObject.Call(d.hdc, oldBrush)
	deleteObject.Call(pen)
	deleteObject.Call(brush)
}

// Fonts use ClearType quality and real type weight rather than rendering every
// control at the same 500 weight. Cache by logical size/weight until shutdown.
func (d *drawCtx) cachedFont(size, weight int, icon bool) uintptr {
	px := int(float64(size) * d.scale)
	if px < 9 {
		px = 9
	}
	cache := d.fonts
	face := preferredFace
	if icon {
		cache = d.iconFonts
		face = "Segoe MDL2 Assets"
		weight = 400
	}
	key := px*1000 + weight
	if v := cache[key]; v != 0 {
		return v
	}
	v, _, _ := createFont.Call(uintptr(^uint32(px-1)), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, ptr(face))
	cache[key] = v
	return v
}
func (d *drawCtx) print(s string, x, y, w, h, size int, color uint32, center bool, weight int, icon bool) {
	if len(s) == 0 {
		return
	}
	font := d.cachedFont(size, weight, icon)
	old, _, _ := selectObject.Call(d.hdc, font)
	setTextColor.Call(d.hdc, uintptr(color))
	setBkMode.Call(d.hdc, TRANSPARENT)
	q := d.ab(x, y, w, h)
	flags := uintptr(DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS)
	if center {
		flags |= DT_CENTER
	}
	drawText.Call(d.hdc, ptr(s), uintptr(^uint32(0)), uintptr(unsafe.Pointer(&q)), flags)
	selectObject.Call(d.hdc, old)
}
func (d *drawCtx) text(s string, x, y, w, h, size int, color uint32, center bool) {
	d.print(s, x, y, w, h, size, color, center, 500, false)
}
func (d *drawCtx) textBold(s string, x, y, w, h, size int, color uint32, center bool) {
	d.print(s, x, y, w, h, size, color, center, 700, false)
}
func (d *drawCtx) textRegular(s string, x, y, w, h, size int, color uint32, center bool) {
	d.print(s, x, y, w, h, size, color, center, 400, false)
}
func (d *drawCtx) icon(glyph rune, x, y, w, h, size int, color uint32) {
	d.print(string(glyph), x, y, w, h, size, color, true, 400, true)
}

// A coherent Windows fluent/MDL2 icon vocabulary replaces placeholders
// such as squares, music-note characters, and mathematical operators.
const (
	icoKeyboard rune = '\ue765'
	icoApps     rune = '\ue71d'
	icoGlobe    rune = '\ue774'
	icoFolder   rune = '\ue8b7'
	icoSettings rune = '\ue713'
	icoVolume   rune = '\ue767'
	icoMute     rune = '\ue74f'
	icoPlay     rune = '\ue768'
	icoNext     rune = '\ue893'
	icoPrev     rune = '\ue892'
	icoDevice   rune = '\ue8b7'
	icoInfo     rune = '\ue946'
	icoCheck    rune = '\ue73e'
	icoDown     rune = '\ue70d'
	icoSave     rune = '\ue74e'
	icoAdd      rune = '\ue710'
	icoRefresh  rune = '\ue72c'
	icoClose    rune = '\ue8bb'
	icoView     rune = '\ue890'
)

func glyphForAction(t string) rune {
	switch t {
	case "shortcut":
		return icoKeyboard
	case "program":
		return icoApps
	case "website":
		return icoGlobe
	case "folder":
		return icoFolder
	case "volume_up", "volume_down":
		return icoVolume
	case "volume_mute":
		return icoMute
	case "media_play":
		return icoPlay
	case "media_next":
		return icoNext
	case "media_prev":
		return icoPrev
	default:
		return icoAdd
	}
}

func (g *ui) clickable(d *drawCtx, x, y, w, h int, name string, id int) {
	g.touches = append(g.touches, click{r: d.ab(x, y, w, h), name: name, id: id})
}
func (g *ui) pill(d *drawCtx, x, y, w, h int, label string, selected bool, handler string, id int) {
	fill := theme("panel2")
	border := theme("border")
	fg := theme("muted")
	if selected {
		fill = theme("sel")
		border = theme("accent")
		fg = theme("text")
	}
	d.rounded(x, y, w, h, 11, fill, border)
	d.textBold(label, x+5, y+2, w-10, h-4, 14, fg, true)
	g.clickable(d, x, y, w, h, handler, id)
}
func (g *ui) navPill(d *drawCtx, x, y, w, h int, glyph rune, label string, selected bool, handler string, id int) {
	fill := theme("sidebar")
	border := theme("border")
	fg := theme("muted")
	if selected {
		fill = theme("sel")
		border = theme("accent")
		fg = theme("text")
	}
	d.rounded(x, y, w, h, 12, fill, border)
	d.icon(glyph, x+15, y+7, 34, h-14, 23, fg)
	d.textBold(label, x+62, y+3, w-68, h-6, 15, fg, false)
	g.clickable(d, x, y, w, h, handler, id)
}
func (g *ui) actionPill(d *drawCtx, x, y, w, h int, glyph rune, label string, selected bool, id int) {
	fill := theme("panel2")
	border := theme("border")
	fg := theme("muted")
	if selected {
		fill = theme("sel")
		border = theme("accent")
		fg = theme("text")
	}
	d.rounded(x, y, w, h, 10, fill, border)
	// Render the same colorful premium artwork here as on the physical-key
	// preview. The monochrome Fluent glyph remains a safe fallback.
	if id < 0 || id >= len(types) || !d.actionArtwork(types[id], x+6, y+3, 39) {
		d.icon(glyph, x+10, y+7, 27, h-14, 18, fg)
	}
	d.textBold(label, x+47, y+1, w-52, h-2, 13, fg, false)
	g.clickable(d, x, y, w, h, "action", id)
}

func (d *drawCtx) keyTextColor(color string) uint32 {
	if lightKeycap(color) {
		return theme("blacktext")
	}
	return RGB(255, 255, 255)
}
func (g *ui) swatch(d *drawCtx, x, y, w, h int, color string, selected bool) {
	edge := keycapHighlight(color)
	if selected {
		d.rounded(x-3, y-3, w+6, h+6, 13, theme("panel"), theme("accent"))
		d.rounded(x-1, y-1, w+2, h+2, 12, theme("panel"), theme("accent"))
	}
	d.rounded(x, y, w, h, 11, colorValue(color), edge)
	if selected {
		d.icon(icoCheck, x, y+1, w, h-2, 23, d.keyTextColor(color))
	}
}
func (g *ui) paint() {
	var ps paintstruct
	dc, _, _ := beginPaint.Call(g.hwnd, uintptr(unsafe.Pointer(&ps)))
	if dc == 0 {
		return
	}
	defer endPaint.Call(g.hwnd, uintptr(unsafe.Pointer(&ps)))
	var bounds rect
	getClientRect.Call(g.hwnd, uintptr(unsafe.Pointer(&bounds)))
	w, h := bounds.Right, bounds.Bottom
	if w <= 0 || h <= 0 {
		return
	}

	// Render an entire frame off-screen, then transfer it in one BitBlt.
	// Former V7 painted hundreds of GDI primitives directly on the visible DC.
	// This caused observable flashing when connection status or keys updated.
	target := dc
	memory, _, _ := createCompatibleDC.Call(dc)
	if memory != 0 {
		bitmap, _, _ := createCompatibleBitmap.Call(dc, uintptr(w), uintptr(h))
		if bitmap != 0 {
			previous, _, _ := selectObject.Call(memory, bitmap)
			if previous != 0 {
				target = memory
				defer func() {
					bitBlt.Call(dc, 0, 0, uintptr(w), uintptr(h), memory, 0, 0, SRCCOPY)
					selectObject.Call(memory, previous)
					deleteObject.Call(bitmap)
					deleteDC.Call(memory)
				}()
			} else {
				deleteObject.Call(bitmap)
				deleteDC.Call(memory)
			}
		} else {
			deleteDC.Call(memory)
		}
	}
	bg, _, _ := createSolidBrush.Call(uintptr(theme("bg")))
	fillRect.Call(target, uintptr(unsafe.Pointer(&bounds)), bg)
	deleteObject.Call(bg)
	d := newDraw(target, w, h)
	g.scale = d.scale
	g.ox = d.ox
	g.oy = d.oy
	g.touches = g.touches[:0]
	d.rounded(0, 0, 244, 900, 0, theme("sidebar"), theme("sidebar"))
	d.rounded(28, 30, 54, 54, 13, theme("accent"), theme("accent"))
	for rr := 0; rr < 2; rr++ {
		for cc := 0; cc < 2; cc++ {
			d.rounded(44+cc*13, 46+rr*13, 10, 10, 3, RGB(255, 255, 255), RGB(255, 255, 255))
		}
	}
	d.textBold("PulseNote Deck", 94, 36, 155, 28, 18, theme("text"), false)
	d.text("CONTROL CENTER", 95, 64, 144, 19, 11, theme("muted"), false)
	d.text("WORKSPACE", 28, 138, 150, 24, 12, theme("muted"), false)
	g.navPill(d, 23, 183, 208, 58, icoKeyboard, "버튼 설정", g.page == 0 || g.page == 3, "page", 0)
	g.navPill(d, 23, 258, 208, 56, icoFolder, "프로필 관리", g.page == 1, "page", 1)
	d.text("SYSTEM", 28, 343, 170, 22, 12, theme("muted"), false)
	g.navPill(d, 23, 379, 208, 56, icoSettings, "장치 및 설정", g.page == 2, "page", 2)
	port, info, connected, lastKey, message, progress := connectionStateSnap()
	conn := progress.title()
	if connected {
		if info > 15 {
			conn = fmt.Sprintf("%d대 연결 · %d키", connectedDeviceCount(info), info)
		} else {
			conn = port + " 연결됨"
		}
	}
	d.rounded(24, 717, 208, 96, 13, theme("panel2"), theme("border"))
	if connected {
		d.rounded(41, 742, 10, 10, 5, theme("green"), theme("green"))
	}
	d.text(conn, 57, 732, 169, 30, 17, theme("text"), false)
	d.text(progress.subtitle(), 38, 773, 180, 23, 12, theme("muted"), false)
	g.pill(d, 24, 823, 208, 43, "설정창 숨기기", false, "hide", 0)
	d.text("v"+version+"   ·   WINDOWS", 27, 875, 193, 21, 11, theme("muted"), false)
	c := snap()
	count, m, p := active(c)
	// Display the detected model, or the last verified model when disconnected.
	title, subtitle := "버튼 설정", "장치 없이도 위 탭에서 모델을 선택해 42개 키를 설정할 수 있습니다."
	if g.page == 1 {
		title, subtitle = "프로필 관리", "PN12·PN15 모델별 프로필을 관리합니다."
	}
	if g.page == 2 {
		title, subtitle = "장치 및 설정", "USB 연결과 Windows 실행 환경을 관리합니다."
	}
	if g.page == 3 {
		title, subtitle = "프로그램 찾기", "앱 이름을 검색한 다음 클릭하면 선택한 버튼에 자동 지정됩니다."
	}
	d.textBold(title, 273, 58, 490, 64, 37, theme("text"), false)
	d.textRegular(subtitle, 275, 123, 800, 29, 16, theme("muted"), false)
	if connected {
		d.rounded(1018, 58, 175, 43, 11, theme("panel2"), theme("border"))
		d.rounded(1033, 74, 10, 10, 5, theme("green"), theme("green"))
		d.text(conn, 1050, 64, 135, 31, 15, theme("text"), false)
	} else {
		d.rounded(1018, 58, 175, 43, 11, theme("panel2"), theme("border"))
		d.text(progress.title(), 1032, 64, 155, 31, 15, theme("muted"), false)
	}
	d.rounded(1208, 58, 211, 43, 11, theme("panel2"), theme("border"))
	modelText := fmt.Sprintf("PN%d · %d키", count, count)
	if count == 15 {
		if c.SelectedSide == "L" {
			modelText = "PN15 · 왼손"
		} else {
			modelText = "PN15 · 오른손"
		}
	}
	d.textBold(modelText, 1213, 62, 201, 35, 14, theme("text"), true)

	if g.page == 0 {
		g.drawKeys(d, count, p, m)
	} else if g.page == 1 {
		g.drawProfiles(d, count, m)
	} else if g.page == 2 {
		g.drawSettings(d, c, count)
	} else {
		g.drawAppPicker(d)
	}
	d.rounded(264, 852, 1154, 38, 11, theme("panel2"), theme("border"))
	hint := "설정 자동 저장 · 키캡 색상은 화면 표시용입니다"
	if connected && info != 0 && info <= 15 && info != count {
		hint = fmt.Sprintf("현재 %d키 화면 · 연결된 하드웨어 %d키 (장치 설정 확인)", count, info)
	}
	if message != "" {
		hint = message
	}
	d.rounded(277, 866, 8, 8, 4, theme("green"), theme("green"))
	d.text(hint, 294, 860, 950, 20, 13, theme("muted"), false)
	d.text(fmt.Sprintf("KEY: %d", lastKey), 1327, 859, 81, 22, 13, theme("accent"), true)
	g.moveEdits(d)
}
func (g *ui) drawKeys(d *drawCtx, count int, p Profile, m Model) {
	d.rounded(263, 175, 630, 660, 18, theme("panel"), theme("border"))
	d.rounded(907, 175, 513, 660, 18, theme("panel"), theme("border"))
	d.textBold(fmt.Sprintf("나의 %d키 컨트롤러", count), 285, 193, 342, 43, 24, theme("text"), false)
	grid := gridForHand(count, snap().SelectedSide)
	arrangement := fmt.Sprintf("가로 %d개 × 세로 %d개 · 키별 자동 저장", grid.Columns, grid.Rows)
	if count == 15 {
		if snap().SelectedSide == "L" {
			arrangement = "5열 × 3행 · 왼쪽 하강 대각선 배열"
		} else {
			arrangement = "5열 × 3행 · 오른쪽 하강 대각선 배열"
		}
	}
	d.textRegular(arrangement, 285, 237, 410, 23, 13, theme("muted"), false)
	// Offline model tabs are editor selectors, not USB connection switches.
	// USB events always use the device's own identified action map.
	selectedID := modelKey(snap())
	for i, tab := range []struct{ id, label string }{
		{"12", "PN12"}, {"15L", "15 왼손"}, {"15", "15 오른손"},
	} {
		g.pill(d, 597+i*93, 195, 89, 39, tab.label, selectedID == tab.id, "editmodel", i)
	}
	d.rounded(282, 285, 592, 477, 16, theme("bg"), theme("border"))
	if count == 15 {
		d.rounded(303, 289, 548, 466, 17, theme("panel2"), theme("border"))
	} else {
		d.rounded(392, 289, 372, 472, 17, theme("panel2"), theme("border"))
	}
	for i := 0; i < count; i++ {
		x, y := grid.posForSide(i, snap().SelectedSide)
		w, h := grid.CellW, grid.CellH
		key := p.Keys[i]
		// All keys retain their saved keycap color, border and highlight.
		// Windows application icons preserve their original alpha and artwork.
		fg := d.keyTextColor(key.Color)
		// One clean focus ring around the selected keycap.
		if g.selected == i {
			d.rounded(x-3, y-3, w+6, h+6, 15, theme("panel2"), theme("accent"))
		}
		d.rounded(x, y, w, h, 13, colorValue(key.Color), keycapHighlight(key.Color))
		d.rounded(x+8, y+5, w-16, 2, 1, keycapHighlight(key.Color), keycapHighlight(key.Color))
		d.textBold(fmt.Sprintf("%02d", i+1), x+10, y+7, 26, 18, 12, fg, false)
		// Artwork is sampled at high quality. Windows application icons
		// retain the original Shell-provided alpha and brand artwork.
		iconSize, iconY := 53, y+21
		if count == 15 {
			iconSize, iconY = 39, y+21
		}
		iconX := x + (w-iconSize)/2
		if key.Type == "program" && key.Value != "" && g.drawAppIcon(d, key.Value, iconX, iconY, iconSize) {
			// Keep genuine app branding when available.
		} else if !d.actionArtwork(key.Type, iconX, iconY, iconSize) {
			d.icon(glyphForAction(key.Type), iconX, iconY, iconSize, iconSize, iconSize-2, fg)
		}
		caption := key.Label
		maxRunes := 10

		if len([]rune(caption)) > maxRunes {
			caption = string([]rune(caption)[:maxRunes-1]) + "…"
		}
		cy := y + 79
		by := y + 99
		captionFont := 14
		typeSize := 11
		if count == 15 {
			cy, by, captionFont, typeSize, maxRunes = y+64, y+84, 13, 10, 9
			if len([]rune(caption)) > maxRunes {
				caption = string([]rune(caption)[:maxRunes-1]) + "…"
			}
		}

		d.textBold(caption, x+5, cy, w-10, 22, captionFont, fg, true)
		d.textRegular(typeName(key.Type), x+5, by, w-10, 16, typeSize, fg, true)
		g.clickable(d, x, y, w, h, "key", i)
	}
	d.text("프로필 · "+p.Name, 292, 782, 390, 27, 14, theme("muted"), false)
	g.pill(d, 704, 778, 165, 44, "동작 테스트", false, "test", 0)
	d.icon(icoPlay, 717, 788, 19, 24, 15, theme("accent"))
	g.drawKeyEdit(d, p.Keys[g.selected])
}

func typeName(t string) string {
	switch t {
	case "none":
		return "기능 없음"
	case "shortcut":
		return "단축키"
	case "program":
		return "프로그램"
	case "website":
		return "웹사이트"
	case "folder":
		return "폴더"
	case "volume_up":
		return "볼륨 +"
	case "volume_down":
		return "볼륨 -"
	case "volume_mute":
		return "음소거"
	case "media_play":
		return "재생 / 정지"
	case "media_next":
		return "다음 곡"
	case "media_prev":
		return "이전 곡"
	default:
		return t
	}
}

var types = []string{"shortcut", "program", "website", "folder", "volume_up", "volume_down", "volume_mute", "media_play", "media_next", "media_prev", "none"}

// Only action types with parameters enable the parameter field.
func needsValue(t string) bool {
	switch t {
	case "program", "website", "folder", "shortcut":
		return true
	}
	return false
}
func valueFieldTitle(t string) string {
	switch t {
	case "program":
		return "실행 파일 · 바로가기 경로"
	case "website":
		return "웹사이트 주소"
	case "folder":
		return "폴더 경로"
	case "shortcut":
		return "키보드 단축키 (예: CTRL+S)"
	default:
		return "추가 입력 없이 바로 실행"
	}
}
func valueCue(t string) string {
	switch t {
	case "program":
		return "앱 검색에서 선택하거나 경로를 입력하세요"
	case "website":
		return "https://example.com"
	case "folder":
		return `C:\Users\...\Documents`
	case "shortcut":
		return "CTRL+SHIFT+S 또는 ALT+TAB"
	default:
		return "이 기능은 별도 값이 필요하지 않습니다"
	}
}
func (g *ui) drawKeyEdit(d *drawCtx, key Key) {
	d.textBold(fmt.Sprintf("키 %02d 설정", g.selected+1), 929, 194, 365, 45, 24, theme("text"), false)
	if key.Type == "program" && key.Value != "" {
		d.text("프로그램 로고 설정", 929, 237, 234, 25, 13, theme("muted"), false)
		g.pill(d, 1167, 238, 122, 30, "투명 PNG", false, "iconpick", 0)
		g.pill(d, 1297, 238, 101, 30, "원본 로고", false, "iconreset", 0)
	} else {
		d.text("기능·이름·색상을 원하는 대로 지정하세요.", 929, 235, 459, 24, 13, theme("muted"), false)
	}
	d.rounded(928, 273, 470, 60, 12, theme("panel2"), theme("border"))
	if key.Type == "program" && key.Value != "" && g.drawAppIcon(d, key.Value, 943, 279, 47) {
		// Resolved the actual application icon.
	} else if !d.actionArtwork(key.Type, 943, 279, 47) {
		d.icon(glyphForAction(key.Type), 946, 285, 35, 35, 24, theme("accent"))
	}
	d.text(key.Label, 989, 281, 392, 25, 17, theme("text"), false)
	d.text(typeName(key.Type), 989, 305, 390, 21, 12, theme("muted"), false)
	d.textBold("기능", 931, 343, 250, 28, 16, theme("text"), false)
	// Eleven original actions remain available, but are consistently grouped.
	for i, t := range types {
		col := i % 3
		row := i / 3
		x := 928 + col*156
		y := 377 + row*53
		g.actionPill(d, x, y, 148, 45, glyphForAction(t), typeName(t), key.Type == t, i)
	}
	d.textBold("버튼 이름", 931, 598, 326, 25, 16, theme("text"), false)
	d.textBold(valueFieldTitle(key.Type), 931, 674, 330, 25, 15, theme("text"), false)
	if key.Type == "program" {
		g.pill(d, 1247, 669, 146, 31, "프로그램 선택", false, "apps", 0)
	}
	d.textBold("키캡 색상", 931, 763, 280, 27, 16, theme("text"), false)
	for i, name := range palette {
		x := 932 + i*66
		g.swatch(d, x, 793, 52, 33, name, key.Color == name)
		g.clickable(d, x-3, 789, 59, 43, "color", i)
	}
}
func (g *ui) drawProfiles(d *drawCtx, count int, m Model) {
	// Main page title/subtitle are painted once by paint() above the card.
	d.rounded(263, 187, 1150, 645, 19, theme("panel"), theme("border"))
	d.text("저장된 프로필", 298, 216, 350, 37, 25, theme("text"), false)
	d.textBold("선택 프로필 이름", 832, 218, 480, 30, 16, theme("text"), false)
	// Two columns from the fifth profile onward: at eight profiles no row
	// can overlap the add/delete actions at y=756.
	for i, p := range m.Profiles {
		x, y, w := 293, 277+i*66, 1094
		if len(m.Profiles) > 6 {
			x, y, w = 293+(i/4)*555, 277+(i%4)*76, 535
		}
		g.pill(d, x, y, w, 56, p.Name, i == m.Active, "profile", i)
	}
	g.pill(d, 297, 756, 210, 50, "+ 새 프로필", false, "profileadd", 0)
	g.pill(d, 522, 756, 218, 50, "선택 프로필 삭제", false, "profiledel", 0)
}
func (g *ui) drawSettings(d *drawCtx, c Config, count int) {
	// No duplicate title here; it overlapped paint() header at smaller scales.
	d.rounded(263, 187, 1150, 647, 19, theme("panel"), theme("border"))
	d.text("PN12 / PN15 장치 구성", 296, 224, 900, 34, 23, theme("text"), false)
	d.rounded(306, 293, 680, 68, 11, theme("sel"), theme("accent"))
	d.textBold(fmt.Sprintf("PN%d · ESP32-C3 · %d키", count, count), 316, 299, 660, 56, 18, theme("text"), true)
	d.text("PN12 기존 설정 보존 · PN15 전용 프로필 분리 · 연결 시 자동 인식", 298, 375, 980, 32, 15, theme("muted"), false)
	port, info, connected, _, _, progress := connectionStateSnap()
	label := progress.title()
	if connected {
		if info > 15 {
			label = fmt.Sprintf("%d대 연결 · 총 %d키", connectedDeviceCount(info), info)
		} else {
			label = fmt.Sprintf("PN%d 연결: %s · 하드웨어 %d키", info, port, info)
		}
	}
	d.text(label, 300, 433, 910, 36, 21, theme("text"), false)
	d.text(progress.description(time.Now()), 300, 476, 1040, 27, 15, theme("muted"), false)
	d.text(progress.checkedLabel()+" · PND:HELLO / PND:READY:2.1", 300, 513, 1040, 25, 14, theme("muted"), false)
	g.pill(d, 306, 556, 400, 64, "COM 포트 재검색", false, "reconnect", 0)
	g.pill(d, 306, 638, 400, 62, "Windows 로그인 시 자동 실행", startupEnabled(), "startup", 0)
	d.text("사용자가 스위치를 누른 경우에만 시작 프로그램에 등록합니다.", 305, 713, 1030, 27, 15, theme("muted"), false)
}

const appsPerPage = 14

func (g *ui) drawAppPicker(d *drawCtx) {
	d.rounded(263, 175, 1156, 660, 18, theme("panel"), theme("border"))
	d.textBold("설치된 프로그램 검색", 294, 196, 560, 42, 24, theme("text"), false)
	d.textRegular("Windows 시작 메뉴와 바탕화면에 등록된 프로그램을 검색합니다.", 296, 242, 945, 24, 14, theme("muted"), false)
	g.pill(d, 1146, 199, 115, 43, "다시 검색", false, "apprescan", 0)
	g.pill(d, 1270, 199, 126, 43, "돌아가기", false, "appback", 0)
	d.textBold("앱 이름 검색", 296, 283, 205, 27, 15, theme("text"), false)
	d.rounded(293, 318, 1083, 46, 10, theme("panel2"), theme("border"))
	if !g.appReady {
		d.icon(icoRefresh, 705, 425, 40, 40, 31, theme("accent"))
		d.textBold("설치된 앱 목록을 확인하고 있습니다", 495, 480, 700, 48, 22, theme("text"), false)
		d.textRegular("검색 중에도 장치 버튼은 백그라운드에서 계속 작동합니다.", 480, 536, 780, 34, 15, theme("muted"), false)
		return
	}
	matching := filterApps(g.apps, g.appQuery)
	items, page := pageApps(matching, g.appPage, appsPerPage)
	g.appPage = page
	if len(items) == 0 {
		d.textBold("검색 결과가 없습니다", 647, 450, 500, 39, 22, theme("text"), false)
		d.textRegular("이름을 바꾸거나 오른쪽의 실행 경로 입력란에 직접 입력하세요.", 530, 508, 780, 26, 14, theme("muted"), false)
		return
	}
	for i, a := range items {
		col := i % 2
		row := i / 2
		x := 294 + col*547
		y := 382 + row*57
		d.rounded(x, y, 529, 50, 10, theme("panel2"), theme("border"))
		if !g.drawAppIcon(d, a.Path, x+13, y+9, 32) {
			d.actionArtwork("program", x+13, y+9, 32)
		}
		d.textBold(a.Name, x+55, y+6, 456, 22, 16, theme("text"), false)
		d.textRegular(a.Path, x+55, y+28, 450, 16, 11, theme("muted"), false)
		g.clickable(d, x, y, 529, 50, "appselect", page*appsPerPage+i)
	}
	pageCount := (len(matching) + appsPerPage - 1) / appsPerPage
	d.text(fmt.Sprintf("검색 결과 %d개 · %d / %d 페이지", len(matching), page+1, pageCount), 305, 801, 600, 24, 14, theme("muted"), false)
	g.pill(d, 1110, 790, 127, 36, "이전", false, "appprev", 0)
	g.pill(d, 1250, 790, 130, 36, "다음", false, "appnext", 0)
}

func (g *ui) moveEdits(d *drawCtx) {
	show := g.page == 0
	if !g.editsPlaced || g.editsVisible != show {
		if show {
			showWindow.Call(g.labelEdit, SW_SHOW)
			showWindow.Call(g.valueEdit, SW_SHOW)
		} else {
			showWindow.Call(g.labelEdit, SW_HIDE)
			showWindow.Call(g.valueEdit, SW_HIDE)
		}
		g.editsVisible = show
	}
	if show {
		label := d.ab(931, 631, 458, 38)
		value := d.ab(931, 708, 458, 40)
		if !g.editsPlaced || g.labelRect != label {
			moveWindow.Call(g.labelEdit, uintptr(label.Left), uintptr(label.Top), uintptr(label.Right-label.Left), uintptr(label.Bottom-label.Top), 0)
			g.labelRect = label
		}
		if !g.editsPlaced || g.valueRect != value {
			moveWindow.Call(g.valueEdit, uintptr(value.Left), uintptr(value.Top), uintptr(value.Right-value.Left), uintptr(value.Bottom-value.Top), 0)
			g.valueRect = value
		}
	}
	g.editsPlaced = true
	searchShow := g.page == 3
	if !g.searchPlaced || g.searchVisible != searchShow {
		if searchShow {
			showWindow.Call(g.appSearchEdit, SW_SHOW)
		} else {
			showWindow.Call(g.appSearchEdit, SW_HIDE)
		}
		g.searchVisible = searchShow
	}
	if searchShow {
		q := d.ab(307, 324, 1053, 34)
		if !g.searchPlaced || g.searchRect != q {
			moveWindow.Call(g.appSearchEdit, uintptr(q.Left), uintptr(q.Top), uintptr(q.Right-q.Left), uintptr(q.Bottom-q.Top), 0)
			g.searchRect = q
		}
	}
	g.searchPlaced = true
	profileShow := g.page == 1
	if !g.profileNamePlaced || g.profileNameVisible != profileShow {
		if profileShow {
			showWindow.Call(g.profileNameEdit, SW_SHOW)
		} else {
			showWindow.Call(g.profileNameEdit, SW_HIDE)
		}
		g.profileNameVisible = profileShow
	}
	if profileShow {
		q := d.ab(1025, 217, 351, 34)
		if !g.profileNamePlaced || g.profileNameRect != q {
			moveWindow.Call(g.profileNameEdit, uintptr(q.Left), uintptr(q.Top), uintptr(q.Right-q.Left), uintptr(q.Bottom-q.Top), 0)
			g.profileNameRect = q
		}
	}
	g.profileNamePlaced = true
}
func (g *ui) redraw() { invalidateRect.Call(g.hwnd, 0, 0) }
func (g *ui) edits() {
	c := snap()
	key, ok := selectedKey(c, g.selected)
	if !ok {
		return
	}
	g.editing = true
	setWindowText.Call(g.labelEdit, ptr(key.Label))
	setWindowText.Call(g.valueEdit, ptr(key.Value))
	enableWindow.Call(g.valueEdit, 0)
	if needsValue(key.Type) {
		enableWindow.Call(g.valueEdit, 1)
	}
	sendMessage.Call(g.valueEdit, EM_SETCUEBANNER, 1, ptr(valueCue(key.Type)))
	g.editing = false
}
func (g *ui) profileEdits() {
	c := snap()
	_, m, _ := active(c)
	g.editing = true
	setWindowText.Call(g.profileNameEdit, ptr(m.Profiles[m.Active].Name))
	g.editing = false
}
func (g *ui) writeProfileName() {
	if g.editing || g.page != 1 {
		return
	}
	c := snap()
	_, m, _ := active(c)
	g.profilePending = true
	g.profilePendingName = getText(g.profileNameEdit)
	g.profilePendingModel = c.Selected
	g.profilePendingSide = c.SelectedSide
	g.profilePendingIndex = m.Active
	killTimer.Call(g.hwnd, PROFILE_SAVE_TIMER)
	setTimer.Call(g.hwnd, PROFILE_SAVE_TIMER, 450, 0)
}
func (g *ui) flushProfileName() error {
	if !g.profilePending {
		return nil
	}
	killTimer.Call(g.hwnd, PROFILE_SAVE_TIMER)
	name := strings.TrimSpace(g.profilePendingName)
	modelID, side, idx := g.profilePendingModel, g.profilePendingSide, g.profilePendingIndex
	g.profilePending = false
	if name == "" {
		g.profileEdits()
		return nil
	}
	if len([]rune(name)) > 40 {
		name = string([]rune(name)[:40])
	}
	err := update(func(c *Config) {
		m, ok := c.Models[modelKeyFor(modelID, side)]
		if !ok || idx < 0 || idx >= len(m.Profiles) {
			return
		}
		m.Profiles[idx].Name = name
		c.Models[modelKeyFor(modelID, side)] = m
	})
	if err != nil {
		// Preserve the unsaved name so the user may retry after fixing the
		// storage problem; do not let an update silently discard this edit.
		g.profilePending = true
		g.showSaveError(err)
	}
	g.redraw()
	return err
}
func getText(hwnd uintptr) string {
	sz, _, _ := sendMessage.Call(hwnd, WM_GETTEXTLENGTH, 0, 0)
	if sz > 4096 {
		sz = 4096
	}
	buf := make([]uint16, sz+1)
	getWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}
func (g *ui) click(x, y int32) {
	for _, hit := range g.touches {
		q := hit.r
		if x >= q.Left && x <= q.Right && y >= q.Top && y <= q.Bottom {
			g.perform(hit)
			break
		}
	}
}
func (g *ui) perform(hit click) {
	g.flushEdits()
	g.flushProfileName()
	switch hit.name {
	case "key":
		g.selected = hit.id
		g.edits()
	case "editmodel":
		// Changing a settings tab must never alter live USB sessions.
		targets := []struct {
			model int
			side  string
		}{{12, "R"}, {15, "L"}, {15, "R"}}
		if hit.id >= 0 && hit.id < len(targets) {
			choice := targets[hit.id]
			g.showSaveError(update(func(c *Config) { setEditorModel(c, choice.model, choice.side) }))
			g.selected = 0
			g.edits()
			g.profileEdits()
		}

	case "color":
		g.showSaveError(update(func(c *Config) { mutateKey(c, g.selected, func(k *Key) { k.Color = palette[hit.id] }) }))
	case "action":
		t := types[hit.id]
		err := update(func(c *Config) {
			mutateKey(c, g.selected, func(k *Key) {
				if k.Type != t {
					k.Value = ""
				}
				k.Type = t
			})
		})
		if err != nil {
			g.showSaveError(err)
		}
		g.edits()
		if t == "program" {
			g.pickApp()
		}
	case "page":
		g.page = hit.id
		if g.page == 1 {
			g.profileEdits()
		}
	case "profile":
		g.showSaveError(update(func(c *Config) { _, m, _ := active(*c); m.Active = hit.id; c.Models[modelKey(*c)] = m }))
		g.selected = 0
		g.edits()
		g.profileEdits()
	case "profileadd":
		g.showSaveError(update(func(c *Config) {
			_, m, p := active(*c)
			if len(m.Profiles) >= maxProfiles {
				return
			}
			q := p
			q.Name = fmt.Sprintf("프로필 %d", len(m.Profiles)+1)
			q.Keys = append([]Key(nil), p.Keys...)
			m.Profiles = append(m.Profiles, q)
			m.Active = len(m.Profiles) - 1
			c.Models[modelKey(*c)] = m
		}))
		g.selected = 0
		g.edits()
		g.profileEdits()
	case "profiledel":
		current := snap()
		_, model, _ := active(current)
		if len(model.Profiles) <= 1 {
			return
		}
		name := model.Profiles[model.Active].Name
		if answer, _, _ := messageBox.Call(g.hwnd, ptr("프로필 '"+name+"'을(를) 삭제할까요?\n삭제하면 해당 프로필의 버튼 설정을 되돌릴 수 없습니다."), ptr("프로필 삭제 확인"), 0x24); answer != 6 {
			return
		}
		g.showSaveError(update(func(c *Config) {
			_, m, _ := active(*c)
			if len(m.Profiles) <= 1 {
				return
			}
			m.Profiles = append(m.Profiles[:m.Active], m.Profiles[m.Active+1:]...)
			m.Active = 0
			c.Models[modelKey(*c)] = m
		}))
		g.selected = 0
		g.edits()
		g.profileEdits()
	case "test":
		// Use the same sequential dispatcher as physical key events so that
		// keyboard shortcuts cannot overlap with a USB-triggered shortcut.
		select {
		case actionQueue <- g.selected + 1:
		default:
			messageBox.Call(g.hwnd, ptr("동작 대기열이 가득 찼습니다. 잠시 후 다시 실행하세요."), ptr("PulseNote Deck"), 0x30)
		}
	case "hide":
		g.flushProfileName()
		showWindow.Call(g.hwnd, SW_HIDE)
	case "apps":
		g.pickApp()
	case "appselect":
		matches := filterApps(g.apps, g.appQuery)
		if hit.id >= 0 && hit.id < len(matches) {
			a := matches[hit.id]
			if err := update(func(c *Config) {
				mutateKey(c, g.selected, func(k *Key) { k.Type = "program"; k.Label = a.Name; k.Value = a.Path })
			}); err != nil {
				g.showSaveError(err)
			} else {
				g.page = 0
				g.edits()
			}
		}
	case "appback":
		g.page = 0
		g.edits()
	case "apprescan":
		g.beginAppScan()
	case "appprev":
		if g.appPage > 0 {
			g.appPage--
		}
	case "appnext":
		all := filterApps(g.apps, g.appQuery)
		if (g.appPage+1)*appsPerPage < len(all) {
			g.appPage++
		}
	case "iconpick":
		k, ok := selectedKey(snap(), g.selected)
		if !ok || k.Type != "program" || k.Value == "" {
			break
		}
		selected := chooseTransparentPNG(g.hwnd)
		if selected == "" {
			break
		}
		// GDI+ may keep the old PNG file open until the image is disposed.
		// Release it *before* atomically replacing a managed icon on Windows.
		g.invalidateAppIcon(k.Value)
		if _, err := installTransparentIcon(k.Value, os.Getenv("APPDATA"), selected); err != nil {
			messageBox.Call(g.hwnd, ptr("로고 교체 실패: "+err.Error()), ptr("PulseNote Deck"), 0x10)
		}
	case "iconreset":
		k, ok := selectedKey(snap(), g.selected)
		if !ok || k.Type != "program" || k.Value == "" {
			break
		}
		// Release GDI+ image handles before removing the on-disk PNG.
		g.invalidateAppIcon(k.Value)
		if err := removeManagedTransparentIcon(k.Value, os.Getenv("APPDATA")); err != nil {
			messageBox.Call(g.hwnd, ptr("원본 로고 복구 실패: "+err.Error()), ptr("PulseNote Deck"), 0x10)
		} else {
			if findTransparentIcon(k.Value) != "" {
				messageBox.Call(g.hwnd, ptr("이름으로 지정한 별도 PNG가 남아 있습니다. 기본 아이콘으로 복구하려면 %APPDATA%\\PulseNoteDeck\\icons 폴더의 해당 PNG를 직접 제거하세요."), ptr("로고 복구"), 0x40)
			}
		}
	case "startup":
		if e := setStartup(!startupEnabled()); e != nil {
			messageBox.Call(g.hwnd, ptr(e.Error()), ptr("자동 시작"), 0x10)
		}
	case "reconnect":
		go requestReconnect()
	}
	g.redraw()
}
func (g *ui) menu(items []string) int {
	h, _, _ := createMenu.Call()
	for i, s := range items {
		appendMenu.Call(h, 0, uintptr(i+1), ptr(s))
	}
	var p point
	getCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	setForeground.Call(g.hwnd)
	res, _, _ := trackPopup.Call(h, TPM_RETURNCMD|TPM_RIGHTBUTTON, uintptr(p.X), uintptr(p.Y), 0, g.hwnd, 0)
	destroyMenu.Call(h)
	return int(res)
}
func (g *ui) showSaveError(err error) {
	if err != nil {
		messageBox.Call(g.hwnd, ptr("설정을 저장하지 못했습니다: "+err.Error()), ptr("PulseNote Deck"), 0x10)
	}
}

func (g *ui) beginAppScan() {
	g.appScanID++
	seq := g.appScanID
	g.appReady = false
	g.appPage = 0
	g.redraw()
	go func() {
		apps := scanInstalledPrograms()
		// Use the Win32 message loop to transfer results to the UI thread.
		// Make room for the newest scan if multiple rescans completed before
		// the UI thread drained the channel. Stale scans are ignored by ID.
		result := scanResult{ID: seq, Apps: apps}
		select {
		case g.appResults <- result:
		default:
			select {
			case <-g.appResults:
			default:
			}
			g.appResults <- result
		}
		postMessage.Call(g.hwnd, WM_USER_APPS, 0, 0)
	}()
}
func (g *ui) pickApp() {
	g.page = 3
	g.appPage = 0
	g.appQuery = ""
	setWindowText.Call(g.appSearchEdit, ptr(""))
	if !g.appReady {
		g.beginAppScan()
	}
	g.redraw()
}

// IImageList IID (Shell32); GetIcon is vtable slot 10.
type iconGUID struct {
	D1     uint32
	D2, D3 uint16
	D4     [8]byte
}

var iidImageList = iconGUID{0x46eb5926, 0x582e, 0x4017, [8]byte{0x9f, 0xdf, 0xe8, 0x99, 0x8d, 0xaa, 0x09, 0x50}}

func imageListMethod(object unsafe.Pointer, index int) uintptr {
	// Keep COM interfaces as actual pointers; converting uintptr back into
	// unsafe.Pointer is not safe for the Go compiler/GC and fails go vet.
	table := *(*unsafe.Pointer)(object)
	return *(*uintptr)(unsafe.Add(table, uintptr(index)*unsafe.Sizeof(uintptr(0))))
}

func highResShellIcon(path string) uintptr {
	var fi fileinfo
	// SHGFI_SYSICONINDEX. On a .lnk path Windows Shell resolves its visual
	// application identity (including the shortcut's custom icon).
	res, _, _ := getFileInfo.Call(ptr(path), 0, uintptr(unsafe.Pointer(&fi)), uintptr(unsafe.Sizeof(fi)), 0x4000)
	if res != 0 && fi.IconIndex >= 0 {
		for _, listSize := range []uintptr{4, 2} { // JUMBO 256px, EXTRALARGE 48px
			var list unsafe.Pointer
			hr, _, _ := shGetImageList.Call(listSize, uintptr(unsafe.Pointer(&iidImageList)), uintptr(unsafe.Pointer(&list)))
			if uint32(hr)&0x80000000 != 0 || list == nil {
				continue
			}
			var hicon uintptr
			call, _, _ := syscall.SyscallN(imageListMethod(list, 10), uintptr(list), uintptr(fi.IconIndex), 1, uintptr(unsafe.Pointer(&hicon)))
			syscall.SyscallN(imageListMethod(list, 2), uintptr(list))
			if uint32(call)&0x80000000 == 0 && hicon != 0 {
				return hicon
			}
		}
	}
	// SHGFI_ICON without SHGFI_SMALLICON: 32x32+ instead of 16x16.
	fi = fileinfo{}
	res, _, _ = getFileInfo.Call(ptr(path), 0, uintptr(unsafe.Pointer(&fi)), uintptr(unsafe.Sizeof(fi)), 0x100)
	if res != 0 {
		return fi.Icon
	}
	return 0
}

// Release a single icon after a user selects or restores an override.
// The next paint resolves the same application through the shared pipeline.
func (g *ui) invalidateAppIcon(path string) {
	if bitmap := g.appBitmaps[path]; bitmap != 0 {
		gdipDisposeImage.Call(bitmap)
	}
	if icon := g.icons[path]; icon != 0 {
		destroyIcon.Call(icon)
	}
	delete(g.appBitmaps, path)
	delete(g.icons, path)
	for i, cached := range g.appIconOrder {
		if cached == path {
			g.appIconOrder = append(g.appIconOrder[:i], g.appIconOrder[i+1:]...)
			break
		}
	}
}

// Windows Shell icons already contain their own alpha information. Keep that
// alpha *as is*: classifying large colored areas as "background" can erase
// genuine logo artwork (notably Arduino's turquoise icon). There is no safe
// fully automatic way to remove a baked-in backdrop from every brand image.
// GDI+ draws the native image with high-quality downscaling and no backplate.
// Bounded LRU: previously a permanent 160-entry limit made all later apps
// display a generic icon. Eviction disposes both GDI+ and Shell resources.
// Refresh recently used icons so browsing many installed apps does not
// repeatedly evict the icons currently visible on the selected profile.
func (g *ui) touchAppIcon(path string) {
	for i, cached := range g.appIconOrder {
		if cached == path {
			copy(g.appIconOrder[i:], g.appIconOrder[i+1:])
			g.appIconOrder[len(g.appIconOrder)-1] = path
			return
		}
	}
}

func (g *ui) trackAppIcon(path string) {
	if len(g.appIconOrder) >= 128 {
		old := g.appIconOrder[0]
		g.appIconOrder = g.appIconOrder[1:]
		if bmp := g.appBitmaps[old]; bmp != 0 {
			gdipDisposeImage.Call(bmp)
		}
		if icon := g.icons[old]; icon != 0 {
			destroyIcon.Call(icon)
		}
		delete(g.appBitmaps, old)
		delete(g.icons, old)
	}
	g.appIconOrder = append(g.appIconOrder, path)
}

// GDI+ LockBits converts the icon to 32-bit ARGB for cautious border-matte
// recovery. When any step fails or the bitmap is ambiguous, leave it intact.
// This is a generic algorithm for all programs, not a name-based exception.
type gdipPixelRect struct{ X, Y, Width, Height int32 }
type gdipPixelData struct {
	Width, Height uint32
	Stride        int32
	PixelFormat   uint32
	Scan0         unsafe.Pointer
	Reserved      uintptr
}

func recoverAppIconMatte(bitmap uintptr) {
	if bitmap == 0 {
		return
	}
	var w, h uint32
	if code, _, _ := gdipImageWidth.Call(bitmap, uintptr(unsafe.Pointer(&w))); code != 0 {
		return
	}
	if code, _, _ := gdipImageHeight.Call(bitmap, uintptr(unsafe.Pointer(&h))); code != 0 {
		return
	}
	if w < 8 || h < 8 || w > 1024 || h > 1024 {
		return
	}
	rc := gdipPixelRect{Width: int32(w), Height: int32(h)}
	var locked gdipPixelData
	// 3 = ImageLockModeRead | ImageLockModeWrite.
	if code, _, _ := gdipBitmapLockBits.Call(bitmap, uintptr(unsafe.Pointer(&rc)), 3, pixelFormat32bppARGB, uintptr(unsafe.Pointer(&locked))); code != 0 {
		return
	}
	defer gdipBitmapUnlock.Call(bitmap, uintptr(unsafe.Pointer(&locked)))
	if locked.Scan0 == nil || locked.Stride <= 0 || locked.Stride > 1024*4*4 {
		return
	}
	bytes := unsafe.Slice((*byte)(locked.Scan0), int(locked.Stride)*int(h))
	recoverOpaqueBlackMatteBGRA(bytes, int(w), int(h), int(locked.Stride))
}

func (g *ui) drawAppIcon(d *drawCtx, path string, x, y, size int) bool {
	if g.icons == nil || g.appBitmaps == nil {
		return false
	}
	bitmap, cached := g.appBitmaps[path]
	if cached {
		g.touchAppIcon(path)
	} else {
		g.trackAppIcon(path)
		// Shared rule for EVERY application: an optional real transparent PNG
		// takes precedence, otherwise use the unmodified native Windows icon.
		// Never guess which opaque colored pixels are part of the logo.
		if iconPath := findTransparentIcon(path); iconPath != "" {
			if result, _, _ := gdipBitmapFromFile.Call(ptr(iconPath), uintptr(unsafe.Pointer(&bitmap))); result != 0 {
				bitmap = 0
			}
		}
		if bitmap == 0 {
			hicon := highResShellIcon(path)
			g.icons[path] = hicon
			if hicon != 0 {
				if result, _, _ := gdipBitmapFromHICON.Call(hicon, uintptr(unsafe.Pointer(&bitmap))); result != 0 {
					bitmap = 0
				} else {
					recoverAppIconMatte(bitmap)
				}
			}
		}
		g.appBitmaps[path] = bitmap
	}
	if bitmap == 0 {
		return false
	}
	var graphics uintptr
	if result, _, _ := gdipFromHDC.Call(d.hdc, uintptr(unsafe.Pointer(&graphics))); result != 0 || graphics == 0 {
		return false
	}
	defer gdipDeleteGraphics.Call(graphics)
	gdipInterpolation.Call(graphics, 7)   // HighQualityBicubic
	gdipPixelOffsetMode.Call(graphics, 4) // HighQuality
	xx, yy := d.xy(x, y)
	pixels := maxInt(1, int(float64(size)*d.scale))
	result, _, _ := gdipImageRect.Call(graphics, bitmap, uintptr(xx), uintptr(yy), uintptr(pixels), uintptr(pixels))
	return result == 0
}

func scanInstalledPrograms() []foundApp {
	roots := []string{}
	for _, spec := range []struct {
		env string
		sub string
	}{{"APPDATA", `Microsoft\Windows\Start Menu\Programs`}, {"ProgramData", `Microsoft\Windows\Start Menu\Programs`}, {"USERPROFILE", "Desktop"}, {"PUBLIC", "Desktop"}} {
		if base := os.Getenv(spec.env); base != "" {
			roots = append(roots, filepath.Join(base, spec.sub))
		}
	}
	apps := make([]foundApp, 0, 128)
	for _, root := range roots {
		if root == "" {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, de os.DirEntry, err error) error {
			if err != nil || de == nil || de.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(path), ".lnk") {
				name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				if len(name) > 1 && len(name) < 80 {
					apps = append(apps, foundApp{Name: name, Path: path})
				}
			}
			return nil
		})
	}
	list := uniqueApps(apps)
	if len(list) > 1000 {
		list = list[:1000]
	}
	return list
}

// Text changes are committed after a short idle interval rather than on every
// keystroke, avoiding synchronous file I/O and GDI redraws during typing.
func (g *ui) writeFromEdit(which int) {
	if g.editing || g.page != 0 {
		return
	}
	c := snap()
	_, m, _ := active(c)
	g.pendingEdits = true
	g.pendingLabel = getText(g.labelEdit)
	g.pendingValue = getText(g.valueEdit)
	g.pendingModel = c.Selected
	g.pendingSide = c.SelectedSide
	g.pendingProfile = m.Active
	g.pendingIndex = g.selected
	killTimer.Call(g.hwnd, EDIT_SAVE_TIMER)
	setTimer.Call(g.hwnd, EDIT_SAVE_TIMER, 400, 0)
}
func (g *ui) flushEdits() error {
	if !g.pendingEdits {
		return nil
	}
	killTimer.Call(g.hwnd, EDIT_SAVE_TIMER)
	label, value := g.pendingLabel, g.pendingValue
	modelID, side, profileID, keyID := g.pendingModel, g.pendingSide, g.pendingProfile, g.pendingIndex
	// Do not depend on the visible selection when the autosave timer fires.
	err := update(func(c *Config) {
		model := c.Models[modelKeyFor(modelID, side)]
		if profileID < 0 || profileID >= len(model.Profiles) || keyID < 0 || keyID >= len(model.Profiles[profileID].Keys) {
			return
		}
		model.Profiles[profileID].Keys[keyID].Label = label
		model.Profiles[profileID].Keys[keyID].Value = value
		c.Models[modelKeyFor(modelID, side)] = model
	})
	if err != nil {
		// Keep the pending snapshot in memory. A close/update request will be
		// refused rather than destroying unsaved edits after an I/O failure.
		messageBox.Call(g.hwnd, ptr("설정 자동 저장 실패: "+err.Error()), ptr("PulseNote Deck"), 0x10)
	} else {
		g.pendingEdits = false
	}
	g.redraw()
	return err
}

// show reopens the existing process rather than starting a second USB worker.
// On every interactive reopen, force a live repaint of the latest synchronized
// serial state. If there is no active port, request an immediate re-scan; do
// NOT tear down an already healthy connection just to refresh the UI.
func (g *ui) show() {
	if g.hwnd == 0 {
		return
	}
	port, _, connected, _, _ := stateSnap()
	if shouldRescanOnShow(port, connected, portList()) {
		requestDiscoveryWake()
	}
	mini, _, _ := isIconic.Call(g.hwnd)
	if mini != 0 {
		showWindow.Call(g.hwnd, SW_RESTORE)
	} else {
		showWindow.Call(g.hwnd, SW_SHOW)
	}
	setForeground.Call(g.hwnd)
	g.redraw()
	updateWindow.Call(g.hwnd)
}
func (g *ui) tray(add bool) {
	if g.hwnd == 0 {
		return
	}
	var n nid
	n.Size = uint32(unsafe.Sizeof(n))
	n.Hwnd = g.hwnd
	n.ID = 1
	n.Flags = NIF_ICON | NIF_TIP | NIF_MESSAGE
	n.CallbackMessage = WM_TRAY
	n.Icon = appIcon
	n.Tip[0] = 0
	uint16copy(n.Tip[:], "PulseNote Deck · PN12/PN15")
	op := uintptr(NIM_ADD)
	if !add {
		op = NIM_DELETE
	}
	notifyIcon.Call(op, uintptr(unsafe.Pointer(&n)))
}
func (g *ui) trayMenu() {
	toggle := "장치 연결 알림 켜기"
	if notificationsEnabled() {
		toggle = "장치 연결 알림 끄기"
	}
	id := g.menu([]string{"설정창 열기", toggle, "완전 종료"})
	if id == 1 {
		g.show()
	} else if id == 2 {
		if err := saveNotificationPreference(notificationPreferencePath(), !notificationsEnabled()); err != nil {
			messageBox.Call(g.hwnd, ptr("장치 알림 설정을 저장하지 못했습니다: "+err.Error()), ptr("PulseNote Deck"), 0x10)
		}
	} else if id == 3 {
		if g.flushProfileName() != nil || g.flushEdits() != nil {
			g.show() // Let the user resolve the save error rather than losing edits.
			return
		}
		g.tray(false)
		destroyWindow.Call(g.hwnd)
	}
}
func (g *ui) quit() {
	g.flushProfileName()
	g.flushEdits()
	if editBgBrush != 0 {
		deleteObject.Call(editBgBrush)
	}
	g.tray(false)
	for _, v := range g.fonts {
		deleteObject.Call(v)
	}
	for _, v := range g.iconFonts {
		deleteObject.Call(v)
	}
	// Dispose GDI+ images before destroying their source shell icons.
	for _, bmp := range g.appBitmaps {
		if bmp != 0 {
			gdipDisposeImage.Call(bmp)
		}
	}
	for _, v := range g.icons {
		if v != 0 {
			destroyIcon.Call(v)
		}
	}
	shutdownRoundedArtwork()
	postQuit.Call(0)
}

var imageLoader = u.NewProc("LoadImageW")
var actionQueue = make(chan int, 128)
var multiActionQueue = make(chan deviceKeyEvent, 256)
var appIcon uintptr
var editBgBrush uintptr
var windowClassName = "PulseNoteDeck_8_12_v74"
var taskbarCreatedMessage uint32

func wndproc(hwnd uintptr, m uint32, wp, lp uintptr) uintptr {
	g := mainUI
	if g == nil {
		return def(hwnd, m, wp, lp)
	}
	// Explorer recreates the notification area after an Explorer restart.
	if taskbarCreatedMessage != 0 && m == taskbarCreatedMessage {
		g.tray(true)
		return 0
	}
	switch m {
	case WM_PAINT:
		g.paint()
		return 0
	case WM_SHOWWINDOW:
		// Handle third-party / legacy direct ShowWindow as well. Do not
		// initiate a new connection during the initial splash/startup.
		if wp != 0 {
			g.redraw()
		}
		return def(hwnd, m, wp, lp)
	case WM_ERASEBKGND:
		// Entire client area is painted by the double-buffered WM_PAINT.
		return 1
	case WM_LBUTTONUP:
		g.click(int32(int16(lp&0xffff)), int32(int16((lp>>16)&0xffff)))
		return 0
	case WM_SIZE:
		g.redraw()
		return 0
	case WM_CLOSE:
		// Close the SETTINGS WINDOW, not the USB listener / action worker.
		// Never hide an unsaved configuration when storage is unavailable.
		if g.flushProfileName() != nil || g.flushEdits() != nil {
			return 0
		}
		showWindow.Call(hwnd, SW_HIDE)
		return 0
	case WM_REOPEN_INSTANCE:
		// A second click on the desktop shortcut asks this running UI thread
		// to synchronize status, re-scan when needed, and reopen the window.
		g.show()
		return 0
	case WM_REPLACE_INSTANCE:
		// Save first, then permit a version upgrade. An I/O failure must not
		// replace this process and discard pending user edits.
		if g.flushProfileName() != nil || g.flushEdits() != nil {
			return 0
		}
		destroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		g.quit()
		return 0
	case WM_COMMAND:
		if g.editing {
			return 0
		}
		id := int(wp & 0xffff)
		code := int((wp >> 16) & 0xffff)
		if code == EN_CHANGE {
			if id == 1201 {
				g.writeFromEdit(1)
			} else if id == 1202 {
				g.writeFromEdit(2)
			} else if id == 1203 && g.page == 3 {
				g.appQuery = getText(g.appSearchEdit)
				g.appPage = 0
				g.redraw()
			} else if id == 1204 && g.page == 1 {
				g.writeProfileName()
			}
		}
		return 0
	case WM_CTLCOLORSTATIC, WM_CTLCOLOREDIT:
		setTextColor.Call(wp, uintptr(theme("text")))
		setBkColor.Call(wp, uintptr(theme("panel2")))
		return editBgBrush
	case WM_USER_STATUS:
		current := snap().Selected
		if g.modelMode != current || g.modelSide != snap().SelectedSide {
			g.modelMode = current
			g.modelSide = snap().SelectedSide
			g.selected = 0
			g.edits()
			g.profileEdits()
		}
		g.redraw()
		return 0
	case WM_DEVICE_NOTICE:
		for {
			select {
			case notice := <-noticeEvents:
				g.trayBalloon(notice)
			default:
				return 0
			}
		}
	case WM_USER_APPS:
		for {
			select {
			case result := <-g.appResults:
				if result.ID == g.appScanID {
					g.apps = result.Apps
					g.appReady = true
					g.appPage = 0
				}
			default:
				g.redraw()
				return 0
			}
		}
	case WM_TRAY:
		if lp == WM_LBUTTONDBLCLK {
			g.show()
		} else if lp == WM_RBUTTONUP {
			g.trayMenu()
		}
		return 0
	case WM_TIMER:
		if wp == EDIT_SAVE_TIMER {
			g.flushEdits()
		}
		if wp == PROFILE_SAVE_TIMER {
			g.flushProfileName()
		}
		return 0
	}
	return def(hwnd, m, wp, lp)
}
func def(h uintptr, m uint32, w, l uintptr) uintptr {
	r, _, _ := defProc.Call(h, uintptr(m), w, l)
	return r
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func main() {
	// Shell's jumbo icon list is COM-backed. Pin the main UI message loop to
	// one STA thread for SHGetImageList and keep it initialized until exit.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := coInit.Call(0, 2) // COINIT_APARTMENTTHREADED
	if uint32(hr)&0x80000000 == 0 {
		defer coUninit.Call()
	}

	taskbarCreated, _, _ := registerWindowMessage.Call(ptr("TaskbarCreated"))
	taskbarCreatedMessage = uint32(taskbarCreated)
	// An automatic login launch must never pop open an already running UI.
	background := false
	for _, arg := range os.Args[1:] {
		if arg == "--background" {
			background = true
		}
	}
	// A new EXE automatically hands off from an older version. The old
	// instance flushes pending edits and removes its tray icon before exit.
	guardMutex, primary := acquireSingleInstance(background)
	if !primary {
		return
	}
	defer closeHandle.Call(guardMutex)
	// Present a short, themed splash only for a real interactive startup.
	// Login autostart remains silent; second launches reuse the running UI.
	var splash *startupSplash
	if shouldShowStartupSplash(background, primary) {
		instance, _, _ := getModuleHandle.Call(0)
		splash = startSplash(instance)
		defer splash.close()
	}
	loadStore()
	if store.Err != nil {
		messageBox.Call(0, ptr("기존 설정이 손상되어 데이터를 보호하기 위해 실행을 중단했습니다.\n"+store.Err.Error()), ptr("PulseNote Deck"), 0x10)
		return
	}
	if err := loadNotificationPreference(notificationPreferencePath()); err != nil && !background {
		// Silent autostart must not block behind a hidden modal dialog.
		// Do not destroy or overwrite an unreadable preference file.
		messageBox.Call(0, ptr("장치 알림 설정 파일을 읽지 못했습니다. 기존 파일은 보존합니다.\n"+err.Error()), ptr("PulseNote Deck"), 0x30)
	}
	if splash != nil {
		splash.advance(1)
	}
	ins, _, _ := getModuleHandle.Call(0)
	cursor, _, _ := loadCursor.Call(0, IDC_ARROW)
	appIcon, _, _ = loadIcon.Call(0, IDI_APPLICATION)
	if exe, e := os.Executable(); e == nil {
		iconPath := filepath.Join(filepath.Dir(exe), "PulseNoteDeck.ico")
		if _, e := os.Stat(iconPath); e == nil {
			if ico, _, _ := imageLoader.Call(0, ptr(iconPath), 1, 64, 64, 0x10); ico != 0 {
				appIcon = ico
			}
		}
	}
	editBgBrush, _, _ = createSolidBrush.Call(uintptr(theme("panel2")))
	preferredFace = chooseUIFont()
	initRoundedArtwork()
	g := &ui{selected: 0, page: 0, modelMode: snap().Selected, modelSide: snap().SelectedSide, fonts: map[int]uintptr{}, iconFonts: map[int]uintptr{}, icons: map[string]uintptr{}, appBitmaps: map[string]uintptr{}, appResults: make(chan scanResult, 8)}
	mainUI = g
	callback := syscall.NewCallback(wndproc)
	class := wndclassex{Size: uint32(unsafe.Sizeof(wndclassex{})), Style: 0, WndProc: callback, Instance: ins, Icon: appIcon, Cursor: cursor, ClassName: ptr(windowClassName), IconSm: appIcon}
	r, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class)))
	if r == 0 {
		return
	}
	// Windows logon autostart uses --background: create the UI hidden from
	// the outset instead of briefly flashing a visible window.
	// Keep the large settings window hidden until the splash has finished.
	// This prevents a one-frame blank window appearing behind the splash.
	windowStyle := uintptr(WS_OVERLAPPEDWINDOW | WS_CLIPCHILDREN)
	hwnd, _, _ := createWindow.Call(0, ptr(windowClassName), ptr("PulseNote Deck "+version+" · PN12 / PN15 Controller"), windowStyle, 80, 55, 1450, 925, 0, 0, ins, 0)
	if hwnd == 0 {
		return
	}
	g.hwnd = hwnd
	enableDarkTitlebar(hwnd)
	label, _, _ := createWindow.Call(0, ptr("EDIT"), ptr(""), WS_VISIBLE|WS_CHILD|WS_BORDER|ES_AUTOHSCROLL, 0, 0, 50, 30, hwnd, uintptr(1201), ins, 0)
	g.labelEdit = label
	value, _, _ := createWindow.Call(0, ptr("EDIT"), ptr(""), WS_VISIBLE|WS_CHILD|WS_BORDER|ES_AUTOHSCROLL, 0, 0, 50, 30, hwnd, uintptr(1202), ins, 0)
	g.valueEdit = value
	search, _, _ := createWindow.Call(0, ptr("EDIT"), ptr(""), WS_CHILD|WS_BORDER|ES_AUTOHSCROLL, 0, 0, 50, 30, hwnd, uintptr(1203), ins, 0)
	g.appSearchEdit = search
	sendMessage.Call(g.appSearchEdit, EM_SETLIMITTEXT, 120, 0)
	profile, _, _ := createWindow.Call(0, ptr("EDIT"), ptr(""), WS_CHILD|WS_BORDER|ES_AUTOHSCROLL, 0, 0, 50, 30, hwnd, uintptr(1204), ins, 0)
	g.profileNameEdit = profile
	sendMessage.Call(g.profileNameEdit, EM_SETLIMITTEXT, 40, 0)
	editFont, _, _ := createFont.Call(uintptr(^uint32(17)), 0, 0, 0, uintptr(500), 0, 0, 0, 1, 0, 0, 0, 0, ptr(preferredFace))
	g.fonts[17*1000+500] = editFont
	sendMessage.Call(g.labelEdit, WM_SETFONT, editFont, 1)
	sendMessage.Call(g.valueEdit, WM_SETFONT, editFont, 1)
	sendMessage.Call(g.appSearchEdit, WM_SETFONT, editFont, 1)
	sendMessage.Call(g.profileNameEdit, WM_SETFONT, editFont, 1)
	g.edits()
	g.tray(true)
	if background {
		showWindow.Call(hwnd, SW_HIDE)
	} else {
		if splash != nil {
			splash.finish()
		}
		showWindow.Call(hwnd, SW_MAXIMIZE)
		updateWindow.Call(hwnd)
	}
	go func() {
		for {
			select {
			case n := <-actionQueue:
				keyInput(n)
			case event := <-multiActionQueue:
				keyInputForDevice(event)
			}
		}
	}()
	go serialLoop()
	var m msg
	for {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}
