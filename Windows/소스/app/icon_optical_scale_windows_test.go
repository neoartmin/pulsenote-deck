//go:build windows

package main

import (
    "image/png"
    "os"
    "testing"
)

// Embedded PNG artwork in the function picker uses an approximately 80%
// visible plate inside a 512x512 transparent canvas. This regression test
// guards the native text icon against becoming much larger again.
func TestNativeTextIconMatchesArtSafeArea(t *testing.T) {
    for _, tc := range []struct{ x, y, size int }{
        {7, 5, 38},     // enlarged 3-column action chooser
        {24, 35, 39},   // PN15 keycap preview
        {13, 21, 53},   // PN12 keycap preview
        {943, 279, 47}, // selected key description
        {10, 10, 1},
    } {
        x, y, visible := fittedTextMacroBounds(tc.x, tc.y, tc.size)
        expected := (tc.size*4 + 2) / 5
        if visible != expected {
            t.Fatalf("size %d: text plate %d, expected %d", tc.size, visible, expected)
        }
        if x < tc.x || y < tc.y ||
            x+visible > tc.x+tc.size || y+visible > tc.y+tc.size {
            t.Errorf("text icon exceeded slot %+v: %d,%d,%d", tc,x,y,visible)
        }
        l, r := x-tc.x, tc.x+tc.size-(x+visible)
        if l-r > 1 || r-l > 1 {
            t.Errorf("text icon not horizontally centered: l=%d r=%d",l,r)
        }
        top, bottom := y-tc.y, tc.y+tc.size-(y+visible)
        if top-bottom > 1 || bottom-top > 1 {
            t.Errorf("text icon not vertically centered: t=%d b=%d",top,bottom)
        }
    }
    if x,y,z := fittedTextMacroBounds(2,3,0);x!=2 || y!=3 || z!=0 {
        t.Fatal("zero-size slot must remain empty")
    }
}

func TestExistingActionPNGPlatesRemainUniform(t *testing.T) {
    // Check representative old and new icons. No original PNG is modified:
    // only the oversized native text artwork is resized.
    names := []string{"shortcut","capture","copy_text","system_settings",
        "auto_click","audio_output","web_search","qr_code"}
    for _, name := range names {
        f,err:=os.Open("assets/"+name+".png")
        if err!=nil {t.Fatal(err)}
        img,err:=png.Decode(f)
        f.Close()
        if err!=nil{t.Fatal(err)}
        bounds:=img.Bounds()
        if bounds.Dx()!=512 || bounds.Dy()!=512 {
            t.Fatalf("%s changed native PNG dimensions",name)
        }
        xmin,ymin,xmax,ymax:=512,512,-1,-1
        for y:=bounds.Min.Y;y<bounds.Max.Y;y++{
            for x:=bounds.Min.X;x<bounds.Max.X;x++{
                _,_,_,alpha:=img.At(x,y).RGBA()
                if alpha>=0x8000 {
                    xx,yy:=x-bounds.Min.X,y-bounds.Min.Y
                    if xx<xmin{xmin=xx}
                    if yy<ymin{ymin=yy}
                    if xx>xmax{xmax=xx}
                    if yy>ymax{ymax=yy}
                }
            }
        }
        if xmax<0 {t.Fatalf("%s is invisible",name)}
        width,height:=xmax-xmin+1,ymax-ymin+1
        if height>width{width=height}
        occupied:=float64(width)/512
        if occupied<0.75 || occupied>0.84 {
            t.Errorf("%s occupies %.1f%% (expected ~80%% optical plate)",name,occupied*100)
        }
    }
}
