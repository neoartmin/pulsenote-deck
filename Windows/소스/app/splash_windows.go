//go:build windows

package main

import (
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

const (
	startupSplashWidth   = 540
	startupSplashHeight  = 268
	startupSplashMinTime = 550 * time.Millisecond
	WS_POPUP             = 0x80000000
	WS_EX_TOOLWINDOW     = 0x00000080
	PM_REMOVE            = 0x0001
)

var (
	getSystemMetrics      = u.NewProc("GetSystemMetrics")
	setWindowRgn          = u.NewProc("SetWindowRgn")
	peekMessage           = u.NewProc("PeekMessageW")
	activeSplash          *startupSplash
	startupSplashCallback uintptr // keep the callback registered for the lifetime of the class
)

type startupSplash struct {
	hwnd    uintptr
	started time.Time
	phase   int
	fonts   map[int]uintptr
	icons   map[int]uintptr
}

func startSplash(instance uintptr) *startupSplash {
	s := &startupSplash{started: time.Now(), fonts: make(map[int]uintptr), icons: make(map[int]uintptr)}
	activeSplash = s
	startupSplashCallback = syscall.NewCallback(startupSplashProc)
	class := wndclassex{
		Size:      uint32(unsafe.Sizeof(wndclassex{})),
		WndProc:   startupSplashCallback,
		Instance:  instance,
		ClassName: ptr("PulseNoteDeck_StartupSplash_7415"),
	}
	if registered, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class))); registered == 0 {
		activeSplash = nil
		return nil
	}
	width, _, _ := getSystemMetrics.Call(0)
	height, _, _ := getSystemMetrics.Call(1)
	x := (int(width) - startupSplashWidth) / 2
	y := (int(height) - startupSplashHeight) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	hwnd, _, _ := createWindow.Call(
		WS_EX_TOOLWINDOW,
		ptr("PulseNoteDeck_StartupSplash_7415"), ptr("PulseNote Deck 시작 중"),
		WS_POPUP,
		uintptr(x), uintptr(y), startupSplashWidth, startupSplashHeight,
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		activeSplash = nil
		return nil
	}
	s.hwnd = hwnd
	// A round, borderless tool window never creates an extra taskbar button.
	region, _, _ := createRoundRectRgn.Call(0, 0, startupSplashWidth+1, startupSplashHeight+1, 32, 32)
	if region != 0 {
		if result, _, _ := setWindowRgn.Call(hwnd, region, 1); result == 0 {
			deleteObject.Call(region) // Windows owns the region on success.
		}
	}
	showWindow.Call(hwnd, SW_SHOW)
	updateWindow.Call(hwnd)
	return s
}

func startupSplashProc(hwnd uintptr, message uint32, wp, lp uintptr) uintptr {
	switch message {
	case WM_PAINT:
		if activeSplash != nil && activeSplash.hwnd == hwnd {
			activeSplash.paint()
			return 0
		}
	case WM_ERASEBKGND:
		return 1
	}
	return def(hwnd, message, wp, lp)
}

func (s *startupSplash) paint() {
	var ps paintstruct
	hdc, _, _ := beginPaint.Call(s.hwnd, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer endPaint.Call(s.hwnd, uintptr(unsafe.Pointer(&ps)))

	d := &drawCtx{hdc: hdc, scale: 1, fonts: s.fonts, iconFonts: s.icons}
	background, _, _ := createSolidBrush.Call(uintptr(theme("bg")))
	bounds := rectI(0, 0, startupSplashWidth, startupSplashHeight)
	fillRect.Call(hdc, uintptr(unsafe.Pointer(&bounds)), background)
	deleteObject.Call(background)

	// Subtle graphite frame, violet accent and a miniature 2 x 3 key matrix.
	d.rounded(1, 1, 538, 266, 16, theme("bg"), theme("border"))
	d.rounded(32, 30, 146, 178, 18, theme("panel"), theme("border"))
	for row := 0; row < 3; row++ {
		for col := 0; col < 2; col++ {
			fill := theme("sel")
			if row == 0 && col == 0 {
				fill = theme("accent")
			}
			if row == 1 && col == 1 {
				fill = RGB(3, 180, 112)
			}
			d.rounded(48+col*59, 46+row*51, 49, 41, 8, fill, theme("border"))
		}
	}
	d.textBold("PulseNote", 202, 39, 290, 43, 29, theme("text"), false)
	d.textBold("DECK", 202, 79, 280, 38, 30, theme("text"), false)
	d.textRegular("PN12 / PN15 · CONTROLLER", 205, 125, 280, 25, 13, theme("muted"), false)
	d.textRegular("빠르고 편리한 단축키 컨트롤러", 204, 155, 300, 28, 14, theme("text"), false)

	label := "환경 확인 중"
	if s.phase >= 1 {
		label = "설정과 화면 준비 중"
	}
	if s.phase >= 2 {
		label = "준비 완료"
	}
	d.text(label, 34, 220, 390, 24, 13, theme("muted"), false)
	d.textRegular("v"+version, 438, 220, 65, 24, 12, theme("muted"), false)
	d.rounded(34, 251, 472, 4, 3, theme("panel"), theme("panel"))
	barWidth := []int{130, 315, 472}[minInt(s.phase, 2)]
	d.rounded(34, 251, barWidth, 4, 3, theme("accent"), theme("accent"))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *startupSplash) advance(phase int) {
	if s == nil || s.hwnd == 0 {
		return
	}
	if phase < 0 {
		phase = 0
	}
	if phase > 2 {
		phase = 2
	}
	s.phase = phase
	invalidateRect.Call(s.hwnd, 0, 0)
	updateWindow.Call(s.hwnd)
}

// Pump queued paint events during only the remaining minimum display interval.
// Avoid a fake fixed startup delay if loading has genuinely taken longer.
func (s *startupSplash) finish() {
	if s == nil || s.hwnd == 0 {
		return
	}
	s.advance(2)
	for remainingSplashTime(s.started, time.Now(), startupSplashMinTime) > 0 {
		var message msg
		for {
			available, _, _ := peekMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0, PM_REMOVE)
			if available == 0 {
				break
			}
			if message.Message == 0x12 { // Preserve WM_QUIT for the main loop.
				postQuit.Call(message.WParam)
				s.close()
				return
			}
			translateMessage.Call(uintptr(unsafe.Pointer(&message)))
			dispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
		}
		time.Sleep(12 * time.Millisecond)
	}
	s.close()
}

func (s *startupSplash) close() {
	if s == nil || s.hwnd == 0 {
		return
	}
	destroyWindow.Call(s.hwnd)
	s.hwnd = 0
	for _, font := range s.fonts {
		if font != 0 {
			deleteObject.Call(font)
		}
	}
	for _, font := range s.icons {
		if font != 0 {
			deleteObject.Call(font)
		}
	}
	if activeSplash == s {
		activeSplash = nil
	}
	// Keep the Go callback rooted until the splash window is actually destroyed.
	runtime.KeepAlive(startupSplashCallback)
}
