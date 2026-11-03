package main

import (
    "fmt"
    "strings"
    "unicode/utf16"
)

const maxClipboardUTF16Units = 4096

var settingIDs = []string{"wifi","bluetooth","display","sound","battery","general"}
var settingLabels = []string{"Wi-Fi","블루투스","화면","소리","절전 모드","전체 설정"}

var settingTargets = map[string]string{
    "wifi":      "ms-settings:network-wifi",
    "bluetooth": "ms-settings:bluetooth",
    "display":   "ms-settings:display",
    "sound":     "ms-settings:sound",
    "battery":   "ms-settings:batterysaver",
    "general":   "ms-settings:",
}

func settingTarget(value string) string {
    if uri, ok := settingTargets[value]; ok { return uri }
    return settingTargets["general"]
}

func settingIndex(value string) int {
    for i, id := range settingIDs {
        if value == id { return i }
    }
    return len(settingIDs) - 1
}

func nextSetting(value string) string {
    return settingIDs[(settingIndex(value)+1)%len(settingIDs)]
}

func settingLabel(value string) string {
    return settingLabels[settingIndex(value)]
}

func clipboardMacroText(value string) (string,error) {
    if value=="" {return "",fmt.Errorf("복사할 문구가 없습니다")}
    if strings.ContainsRune(value, '\x00') {return "",fmt.Errorf("NUL 문자를 포함할 수 없습니다")}
    if len(utf16.Encode([]rune(value)))>maxClipboardUTF16Units {
        return "",fmt.Errorf("문구는 UTF-16 기준 %d자 이하여야 합니다",maxClipboardUTF16Units)
    }
    // Preserve user-authored line breaks; native Windows clipboard uses CRLF.
    return strings.ReplaceAll(
        strings.ReplaceAll(strings.ReplaceAll(value,"\r\n","\n"),"\r","\n"),
        "\n","\r\n"),nil
}
