package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// An alpha PNG is the only lossless way to remove a background that has been
// permanently baked into an application's icon. This convention works with
// any program, requires no app-specific exceptions and does not change UI.
// Supported names: icons/<shortcut or exe filename>.png and a path-hash PNG.
func transparentIconPaths(programPath, roamingData string) []string {
	if roamingData == "" || programPath == "" {
		return nil
	}
	directory := filepath.Join(roamingData, "PulseNoteDeck", "icons")
	name := strings.TrimSuffix(filepath.Base(programPath), filepath.Ext(programPath))
	hash := sha256.Sum256([]byte(strings.ToLower(programPath)))
	return []string{
		filepath.Join(directory, hex.EncodeToString(hash[:8])+".png"),
		filepath.Join(directory, name+".png"),
	}
}

func pngHasTransparency(filename string) bool {
	stat, err := os.Stat(filename)
	if err != nil || stat.IsDir() || stat.Size() > 6*1024*1024 {
		return false
	}
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()
	// Inspect the PNG dimensions before decoding pixels. A tiny compressed
	// file may claim enormous dimensions and exhaust RAM during png.Decode.
	config, err := png.DecodeConfig(f)
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 2048 || config.Height > 2048 {
		return false
	}
	if _, err = f.Seek(0, 0); err != nil {
		return false
	}
	img, err := png.Decode(f)
	if err != nil {
		return false
	}
	bounds := img.Bounds()
	if bounds.Dx() != config.Width || bounds.Dy() != config.Height {
		return false
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 65535 {
				return true
			}
		}
	}
	return false
}

func findTransparentIcon(programPath string) string {
	for _, path := range transparentIconPaths(programPath, os.Getenv("APPDATA")) {
		if pngHasTransparency(path) {
			return path
		}
	}
	return ""
}

// installTransparentIcon stores a validated alpha PNG under a path-specific
// filename. This avoids collisions when applications have the same shortcut name.
func installTransparentIcon(programPath, roamingData, chosenPNG string) (string, error) {
	paths := transparentIconPaths(programPath, roamingData)
	if len(paths) == 0 {
		return "", fmt.Errorf("프로그램 경로 또는 APPDATA를 찾지 못했습니다")
	}
	if !pngHasTransparency(chosenPNG) {
		return "", fmt.Errorf("투명한 픽셀이 있는 PNG 파일만 사용할 수 있습니다 (최대 6MB, 2048픽셀)")
	}
	content, err := os.ReadFile(chosenPNG)
	if err != nil {
		return "", err
	}
	dest := paths[0]
	if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return "", err
	}
	// Use a new temp file to avoid truncating an existing working override.
	f, err := os.CreateTemp(filepath.Dir(dest), ".pnd-icon-*.tmp")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return "", err
	}
	if _, err = f.Write(content); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = os.Rename(tmp, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// removeManagedTransparentIcon removes only the override created by the UI.
// Manually supplied name-based PNGs are not silently deleted.
func removeManagedTransparentIcon(programPath, roamingData string) error {
	paths := transparentIconPaths(programPath, roamingData)
	if len(paths) == 0 {
		return fmt.Errorf("프로그램 경로 또는 APPDATA를 찾지 못했습니다")
	}
	err := os.Remove(paths[0])
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
