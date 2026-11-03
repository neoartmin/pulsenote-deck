package main

import (
    "strings"
    "testing"
)

func TestWindowsPermissionFreeActionSettings(t *testing.T){
    want:=[]string{
        "ms-settings:network-wifi",
        "ms-settings:bluetooth",
        "ms-settings:display",
        "ms-settings:sound",
        "ms-settings:batterysaver",
        "ms-settings:",
    }
    if len(settingIDs)!=len(want)||len(settingLabels)!=len(want){
        t.Fatal("setting picker and allowed URI list differ")
    }
    for i,v:=range want {
        if got:=settingTarget(settingIDs[i]);got!=v {
            t.Fatalf("setting %d URI %s expected %s",i,got,v)
        }
        if got:=settingIndex(settingIDs[i]);got!=i{t.Errorf("index %d %d",i,got)}
        if got:=nextSetting(settingIDs[i]);got!=settingIDs[(i+1)%len(settingIDs)]{
            t.Errorf("selection cycle %d: %s",i,got)
        }
        if settingLabel(settingIDs[i])==""{t.Errorf("empty label")}
    }
    if settingTarget("ms-settings:privacy-webcam")!="ms-settings:"{
        t.Fatal("arbitrary settings protocol must not pass through")
    }
    if settingIndex("wrong")!=len(settingIDs)-1{t.Fatal("unsafe picker default")}
}

func TestWindowsClipboardMacroUTF16AndNewlines(t *testing.T){
    cases:=[]struct{raw,want string}{
        {"안녕\n두 번째 줄 😀","안녕\r\n두 번째 줄 😀"},
        {"line1\r\nline2","line1\r\nline2"},
        {"line1\rline2","line1\r\nline2"},
        {"  앞뒤 공백  ","  앞뒤 공백  "},
    }
    for _,tc:=range cases{
        got,err:=clipboardMacroText(tc.raw)
        if err!=nil||got!=tc.want{t.Fatalf("clipboard: %q => %q, %v",tc.raw,got,err)}
    }
    for _,bad:=range []string{"","NUL\x00text",strings.Repeat("A",maxClipboardUTF16Units+1),strings.Repeat("😀",maxClipboardUTF16Units/2+1)}{
        if _,err:=clipboardMacroText(bad);err==nil{
            t.Fatalf("accepted invalid clipboard macro length=%d",len(bad))
        }
    }
    if _,err:=clipboardMacroText(strings.Repeat("A",maxClipboardUTF16Units));err!=nil{
        t.Fatal("rejects exact limit",err)
    }
}

func TestWindowsPermissionFreeActionsRemainEditable(t *testing.T){
    actions:=[]struct{kind,caption string;param bool}{
        {"copy_text","문구 복사",true},
        {"system_settings","설정 바로가기",true},
        {"open_timers","타이머 열기",false},
        {"open_alarms","알람 열기",false},
    }
    for _,tc:=range actions{
        if !validTypes[tc.kind]{t.Fatalf("missing action %s",tc.kind)}
        if typeName(tc.kind)!=tc.caption{t.Fatalf("label %s",tc.kind)}
        if needsValue(tc.kind)!=tc.param{t.Fatalf("input mismatch %s",tc.kind)}
        found:=false
        for _,typ:=range types{if typ==tc.kind{found=true}}
        if !found{t.Fatalf("%s missing from GUI picker",tc.kind)}
        for _,size:=range []int{12,15}{
            c:=defaults()
            c.Selected=size
            if size==15{c.SelectedSide="L"}
            ok:=mutateKey(&c,0,func(k *Key){
                k.Type=tc.kind
                k.Value="두 줄\n두 번째 줄"
                k.Label="작업"
            })
            if !ok{t.Fatalf("cannot edit %s on %d",tc.kind,size)}
            normalized:=normalize(c)
            got,selected:=selectedKey(normalized,0)
            if !selected||got.Type!=tc.kind||got.Label!="작업"{
                t.Fatalf("%s action lost on %d: %+v",tc.kind,size,got)
            }
        }
    }
}
