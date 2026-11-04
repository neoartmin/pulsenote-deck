//go:build windows

package main

import (
    "encoding/xml"
    "math"
    "os"
    "strings"
    "testing"
)

func TestPerMonitorV2DPIManifest(t *testing.T) {
    content,err:=os.ReadFile("app.manifest")
    if err!=nil {t.Fatal(err)}
    var root struct{ XMLName xml.Name }
    if err:=xml.Unmarshal(content,&root);err!=nil{t.Fatalf("invalid Windows manifest XML: %v",err)}
    s:=string(content)
    for _,token:=range []string{
        "PerMonitorV2,PerMonitor",
        "http://schemas.microsoft.com/SMI/2016/WindowsSettings",
        "true/pm",
        "requestedExecutionLevel",
    } {
        if !strings.Contains(s,token){t.Errorf("missing DPI manifest setting %q",token)}
    }
}

func TestPhysicalPixelCanvasScale(t *testing.T) {
    old:=mainUI
    mainUI=&ui{fonts:map[int]uintptr{},iconFonts:map[int]uintptr{}}
    defer func(){mainUI=old}()
    for _,tt:=range []struct{w,h int32;expected float64}{
        {1440,900,1},
        {1800,1125,1.25},
        {2160,1350,1.5},
        {2880,1800,2},
        {3600,2250,2.5},
        {720,450,.5},
    } {
        d:=newDraw(0,tt.w,tt.h)
        if math.Abs(d.scale-tt.expected)>1e-6{
            t.Errorf("%dx%d: scale %.5f, want %.2f",tt.w,tt.h,d.scale,tt.expected)
        }
    }
}

func TestDPISignalsAndNoDoubleScaling(t *testing.T) {
    for _,tt:=range []struct{w uintptr;want uint32}{
        {96,96},{120,120},{144,144},{192,192},{240,240},{0,96},{1500,96},
    } {
        if got:=dpiFromMessage(tt.w);got!=tt.want{t.Errorf("WM_DPICHANGED %d => %d want %d",tt.w,got,tt.want)}
    }
    c,err:=os.ReadFile("gui_windows.go")
    if err!=nil{t.Fatal(err)}
    src:=string(c)
    for _,wanted:=range []string{
        "case WM_DPICHANGED:",
        "g.handleDPIChanged(hwnd,wp,lp)",
        "g.applyNativeEditorFont(d.scale)",
        "configureProcessDPI()",
        "d.antiAliasedRounded(x,y,w,h,r,fill,border)",
    } {
        if !strings.Contains(src,wanted){t.Errorf("DPI or smooth-shape pipeline missing: %s",wanted)}
    }
    if strings.Contains(src,"if scale > 1.65") {t.Fatal("165%% cap prevents full rendering on 200%% DPI displays")}
    if !strings.Contains(src,"g.clickable(d, x, y, w, h, \"key\", i)") {
        t.Fatal("key hit regions must use the same physical scale as their rendered shapes")
    }
}

func TestGDIColorConversionForSmoothBorders(t *testing.T) {
    if got:=gdipARGB(RGB(214,40,57));got!=0xFFD62839 {
        t.Fatalf("GDI+ ARGB conversion: %#x",got)
    }
}
