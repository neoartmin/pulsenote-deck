package main

import (
    "errors"
    "strings"
    "unicode/utf8"
)

const maxTextMacroBytes = 4096

// expandTextMacro decodes literal \\n / \\t / \\\\ typed in the one-line editor.
// Unknown backslash sequences remain unchanged (e.g. Windows paths).
// No clipboard contents or environment expansion are used for text macros.
func expandTextMacro(raw string) (string, error) {
    if raw == "" {
        return "", errors.New("입력할 텍스트가 비어 있습니다")
    }
    if len(raw) > maxTextMacroBytes || !utf8.ValidString(raw) {
        return "", errors.New("텍스트는 UTF-8 4096바이트 이하이어야 합니다")
    }
    runes := []rune(raw)
    var out strings.Builder
    out.Grow(len(raw))
    for i := 0; i < len(runes); i++ {
        ch := runes[i]
        if ch == '\\' && i+1 < len(runes) {
            switch runes[i+1] {
            case 'n':
                out.WriteRune('\n')
                i++
                continue
            case 't':
                out.WriteRune('\t')
                i++
                continue
            case '\\':
                out.WriteRune('\\')
                i++
                continue
            }
        }
        if ch == 0 {
            return "", errors.New("NUL 문자는 입력할 수 없습니다")
        }
        out.WriteRune(ch)
    }
    result := strings.ReplaceAll(out.String(), "\r\n", "\n")
    if result == "" {
        return "", errors.New("입력할 텍스트가 비어 있습니다")
    }
    return result, nil
}
