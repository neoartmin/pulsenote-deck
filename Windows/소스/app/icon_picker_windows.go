//go:build windows

package main

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var commonDialog = syscall.NewLazyDLL("comdlg32.dll")
var getOpenFileName = commonDialog.NewProc("GetOpenFileNameW")

// Windows OPENFILENAMEW, used rather than adding a second app-specific UI.
// The standard file dialog supports ordinary keyboard and mouse navigation.
type openFileNameW struct {
	Size            uint32
	Owner           uintptr
	Instance        uintptr
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefaultExt      *uint16
	CustomData      uintptr
	Hook            uintptr
	Template        *uint16
	Reserved        uintptr
	ReservedDWORD   uint32
	FlagsEx         uint32
}

func chooseTransparentPNG(owner uintptr) string {
	path := make([]uint16, 32768)
	filter := utf16.Encode([]rune("PNG 이미지 (*.png)\x00*.png\x00\x00"))
	title, _ := syscall.UTF16PtrFromString("배경이 투명한 프로그램 아이콘 PNG 선택")
	ofn := openFileNameW{
		Size:        uint32(unsafe.Sizeof(openFileNameW{})),
		Owner:       owner,
		Filter:      &filter[0],
		FilterIndex: 1,
		File:        &path[0],
		MaxFile:     uint32(len(path)),
		Title:       title,
		Flags:       0x80000 | 0x1000 | 0x800 | 0x8, // EXPLORER, FILEMUSTEXIST, PATHMUSTEXIST, NOCHANGEDIR
	}
	success, _, _ := getOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if success == 0 {
		return ""
	} // Cancelled: no change.
	return syscall.UTF16ToString(path)
}
