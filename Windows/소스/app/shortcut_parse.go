package main

import (
	"fmt"
	"strconv"
	"strings"
)

// parseShortcut validates the entire combination before any key is injected.
// A malformed trailing key must never leave an earlier CTRL/ALT/SHIFT pressed.
func parseShortcut(value string) ([]byte, error) {
	parts := strings.Split(value, "+")
	if len(parts) == 0 || len(parts) > 5 {
		return nil, fmt.Errorf("키 조합은 1~5개여야 합니다")
	}
	codes := make([]byte, 0, len(parts))
	seen := map[byte]bool{}
	for _, raw := range parts {
		s := strings.ToUpper(strings.TrimSpace(raw))
		if s == "" {
			return nil, fmt.Errorf("비어 있는 단축키 항목이 있습니다")
		}
		code, ok := shortcutKeyCode(s)
		if !ok {
			return nil, fmt.Errorf("지원하지 않는 키: %s", raw)
		}
		if seen[code] {
			return nil, fmt.Errorf("중복된 키: %s", raw)
		}
		seen[code] = true
		codes = append(codes, code)
	}
	return codes, nil
}

func shortcutKeyCode(s string) (byte, bool) {
	mods := map[string]byte{"CTRL": 0x11, "CONTROL": 0x11, "ALT": 0x12, "SHIFT": 0x10, "WIN": 0x5b, "WINDOWS": 0x5b, "ENTER": 0x0d, "TAB": 0x09, "SPACE": 0x20, "ESC": 0x1b, "ESCAPE": 0x1b, "BACKSPACE": 0x08, "DELETE": 0x2e, "UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27, "HOME": 0x24, "END": 0x23, "PAGEUP": 0x21, "PAGEDOWN": 0x22, "INSERT": 0x2d, "PRINTSCREEN": 0x2c}
	if v, ok := mods[s]; ok {
		return v, true
	}
	if len(s) == 1 && ((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= '0' && s[0] <= '9')) {
		return s[0], true
	}
	if strings.HasPrefix(s, "F") {
		n, e := strconv.Atoi(s[1:])
		if e == nil && n >= 1 && n <= 24 {
			return byte(0x70 + n - 1), true
		}
	}
	return 0, false
}
