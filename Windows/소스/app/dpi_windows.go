//go:build windows

package main

import (
    "syscall"
    "unsafe"
)

const (
    WM_DPICHANGED = 0x02E0
    SWP_NOZORDER = 0x0004
    SWP_NOACTIVATE = 0x0010
    defaultWindowDPI uint32 = 96
)

var (
    winSetProcessDPIContext = u.NewProc("SetProcessDpiAwarenessContext")
    winGetDpiForWindow = u.NewProc("GetDpiForWindow")
    winSetWindowPos = u.NewProc("SetWindowPos")
    shcoreDPI = syscall.NewLazyDLL("shcore.dll")
    shcoreSetProcessDPI = shcoreDPI.NewProc("SetProcessDpiAwareness")
)

// The embedded EXE manifest sets PerMonitorV2 before any HWND exists.
// This is a fallback for source builds that omit the .syso resource.
// A failed call is expected when the manifest has already selected DPI mode.
func configureProcessDPI() {
    if winSetProcessDPIContext.Find() == nil {
        // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
        if ok,_,_:=winSetProcessDPIContext.Call(^uintptr(3)); ok!=0 {
            return
        }
        // If Windows has already set the mode through our manifest, leave it.
        return
    }
    if shcoreSetProcessDPI.Find() == nil {
        // PROCESS_PER_MONITOR_DPI_AWARE = 2 (Windows 8.1 fallback)
        shcoreSetProcessDPI.Call(2)
    }
}

func currentWindowDPI(hwnd uintptr) uint32 {
    if hwnd!=0 && winGetDpiForWindow.Find()==nil {
        dpi,_,_:=winGetDpiForWindow.Call(hwnd)
        if dpi>=48 && dpi<=960 {return uint32(dpi)}
    }
    return defaultWindowDPI
}

func dpiFromMessage(wParam uintptr) uint32 {
    dpi:=uint32(wParam & 0xffff)
    if dpi<48 || dpi>960 {return defaultWindowDPI}
    return dpi
}

func (g *ui) handleDPIChanged(hwnd,wp,lp uintptr) {
    g.windowDPI = dpiFromMessage(wp)
    // lParam points to the Windows-recommended *physical pixel* rectangle.
    // Following it preserves the visual size when crossing monitors.
    if lp!=0 {
        suggested:=*(*rect)(unsafe.Pointer(lp))
        width,height:=int64(suggested.Right)-int64(suggested.Left),int64(suggested.Bottom)-int64(suggested.Top)
        if width>0 && height>0 {
            winSetWindowPos.Call(hwnd,0,uintptr(int64(suggested.Left)),uintptr(int64(suggested.Top)),
                uintptr(width),uintptr(height),SWP_NOZORDER|SWP_NOACTIVATE)
        }
    }
    // Old handles can be tied to the former physical pixel font sizes.
    // The next paint creates correct-size GDI fonts and positions all HWND edits.
    g.resetDPIRenderCache()
    g.redraw()
}

func (g *ui) resetDPIRenderCache() {
    for _,font:=range g.fonts {if font!=0{deleteObject.Call(font)}}
    for _,font:=range g.iconFonts {if font!=0{deleteObject.Call(font)}}
    g.fonts=map[int]uintptr{}
    g.iconFonts=map[int]uintptr{}
    g.editsPlaced=false
    g.profileNamePlaced=false
    g.searchPlaced=false
}

// Native EDIT controls are separate Win32 HWNDs and cannot inherit the GDI
// text sizing of the parent backbuffer. Recompute when the physical pixel
// ratio changes, and do not tear down GDI objects before replacing them.
func (g *ui) applyNativeEditorFont(scale float64) {
    px:=int(17*scale+0.5)
    if px<9 {px=9}
    if px==g.nativeEditPx && g.nativeEditFont!=0 {return}
    font,_,_:=createFont.Call(uintptr(^uint32(px-1)),0,0,0,uintptr(500),
        0,0,0,1,0,0,5,0,ptr(preferredFace))
    if font==0 {return}
    for _,edit:=range []uintptr{
        g.labelEdit,g.valueEdit,g.textValueEdit,g.appSearchEdit,
        g.profileNameEdit,g.emailToEdit,g.emailSubjectEdit,g.emailBodyEdit,
    } {
        if edit!=0 {sendMessage.Call(edit,WM_SETFONT,font,0)}
    }
    old:=g.nativeEditFont
    g.nativeEditFont=font
    g.nativeEditPx=px
    if old!=0 {deleteObject.Call(old)}
}
