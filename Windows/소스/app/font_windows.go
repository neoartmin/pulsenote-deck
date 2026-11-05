//go:build windows

package main

import "unsafe"

// Fonts are read from artwork's embedded assets/* collection. CI pins the
// original, unmodified Pretendard files by upstream commit and Git blob SHA.
// No machine-wide installation, registry edits, or font file write is needed.
var (
    addFontMemResource = gd.NewProc("AddFontMemResourceEx")
    removeFontMemResource = gd.NewProc("RemoveFontMemResourceEx")
    privateFontHandles []uintptr
    privateFontBuffers [][]byte
)

func loadBundledPretendard() bool {
    filenames := []string{
        "assets/Pretendard-Regular.otf",
        "assets/Pretendard-Medium.otf",
        "assets/Pretendard-Bold.otf",
    }
    // Read the entire bundle before registering any member. A source build
    // without the preparation step safely falls back to installed fonts.
    buffers := make([][]byte, 0, len(filenames))
    for _, name := range filenames {
        b, err := artwork.ReadFile(name)
        if err != nil || len(b) < 1024 {
            return false
        }
        buffers = append(buffers, b)
    }
    for _, b := range buffers {
        var installed uint32
        handle, _, _ := addFontMemResource.Call(
            uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), 0,
            uintptr(unsafe.Pointer(&installed)))
        if handle == 0 || installed == 0 {
            unloadBundledPretendard()
            return false
        }
        privateFontHandles = append(privateFontHandles, handle)
        privateFontBuffers = append(privateFontBuffers, b) // retain for GDI lifetime
    }
    return true
}

func unloadBundledPretendard() {
    for i := len(privateFontHandles) - 1; i >= 0; i-- {
        removeFontMemResource.Call(privateFontHandles[i])
    }
    privateFontHandles = nil
    privateFontBuffers = nil
}
