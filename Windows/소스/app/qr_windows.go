//go:build windows

package main

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
    "unicode/utf8"

    qrcode "github.com/skip2/go-qrcode"
)

// Generates QR in-process: no web service sees the configured text.
// The resulting PNG is saved in the user's Pictures folder and opened in
// the default image viewer. Explicitly selected content may be confidential.
func showOfflineQRCode(value string) error {
    if value==""||!utf8.ValidString(value)||len([]byte(value))>900 {
        return fmt.Errorf("QR 내용은 UTF-8 1~900바이트로 입력하세요")
    }
    home,err:=os.UserHomeDir()
    if err!=nil{return err}
    directory:=filepath.Join(home,"Pictures","PulseNote Deck")
    if err=os.MkdirAll(directory,0700);err!=nil{return err}
    filename:=filepath.Join(directory,fmt.Sprintf(
        "PulseNote_QR_%s.png",time.Now().Format("20060102_150405.000000000")))
    if err=qrcode.WriteFile(value,qrcode.Medium,512,filename);err!=nil{return err}
    return shellOpen(filename)
}
