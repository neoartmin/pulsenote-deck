//go:build windows

package main

import (
    "encoding/json"
    "testing"
)

type toggleKeyEvent struct {code byte; up bool}

func recordToggleKeyboard(t *testing.T) *[]toggleKeyEvent {
    t.Helper()
    releaseAllHeldShortcuts()
    events := make([]toggleKeyEvent,0,20)
    original := emitToggleKey
    emitToggleKey = func(code byte,up bool){
        events = append(events,toggleKeyEvent{code,up})
    }
    t.Cleanup(func(){
        releaseAllHeldShortcuts()
        emitToggleKey = original
    })
    return &events
}

func TestShortcutTogglePressRelease(t *testing.T){
    events:=recordToggleKeyboard(t)
    // Drain earlier notice events; this test checks the exact count.
    for len(shortcutToggleEvents)>0{<-shortcutToggleEvents}
    active,err:=toggleShortcut("CTRL+SHIFT+S","저장","15","R",0,1)
    if err!=nil||!active||heldShortcutCount()!=1{t.Fatalf("on %v err=%v",active,err)}
    if len(*events)!=3||(*events)[0].up||(*events)[1].up||(*events)[2].up{
        t.Fatalf("expected 3 KEYDOWNs: %+v",*events)
    }
    active,err=toggleShortcut("CTRL+SHIFT+S","저장","15","R",0,1)
    if err!=nil||active||heldShortcutCount()!=0{t.Fatalf("off %v err=%v",active,err)}
    if len(*events)!=6||!(*events)[3].up||!(*events)[4].up||!(*events)[5].up{
        t.Fatalf("expected 3 KEYUPs: %+v",*events)
    }
    if len(shortcutToggleEvents)!=2{
        t.Fatalf("expected 1 notice per transition, got %d",len(shortcutToggleEvents))
    }
    if ev:=<-shortcutToggleEvents;!ev.active{t.Fatal("first notice must be ON")}
    if ev:=<-shortcutToggleEvents;ev.active{t.Fatal("second notice must be OFF")}
}

func TestShortcutToggleSharedModifiers(t *testing.T){
    events:=recordToggleKeyboard(t)
    _,err:=toggleShortcut("CTRL+S","save","12","R",0,1)
    if err!=nil{t.Fatal(err)}
    _,err=toggleShortcut("CTRL+C","copy","12","R",0,2)
    if err!=nil{t.Fatal(err)}
    if len(*events)!=3{t.Fatalf("shared Ctrl must have only one down: %+v",*events)}
    if err=shortcutRespectingHolds("CTRL+V");err!=nil{t.Fatal(err)}
    if len(*events)!=5{t.Fatalf("normal shortcut released held key: %+v",*events)}
    _,err=toggleShortcut("CTRL+S","save","12","R",0,1)
    if err!=nil{t.Fatal(err)}
    if len(*events)!=6||(*events)[5].code!='S'||!(*events)[5].up{
        t.Fatalf("shared modifier released too soon: %+v",*events)
    }
    _,err=toggleShortcut("CTRL+C","copy","12","R",0,2)
    if err!=nil{t.Fatal(err)}
    if len(*events)!=8||(*events)[7].code!=0x11||!(*events)[7].up{
        t.Fatalf("last CTRL release missing: %+v",*events)
    }
}

func TestShortcutToggleProfileChangeAndConfigPersistence(t *testing.T){
    events:=recordToggleKeyboard(t)
    c:=defaults()
    m:=c.Models["15"]
    m.Profiles[0].Keys[0]=Key{Label:"hold",Type:"shortcut",Value:"CTRL+S",Toggle:true,Color:"회색"}
    c.Models["15"]=m
    c=normalize(c)
    data,err:=json.Marshal(c)
    if err!=nil{t.Fatal(err)}
    var restored Config
    if err=json.Unmarshal(data,&restored);err!=nil{t.Fatal(err)}
    if !restored.Models["15"].Profiles[0].Keys[0].Toggle{t.Fatal("toggle setting lost on save/load")}
    _,err=toggleShortcut("CTRL+S","hold","15","R",0,1)
    if err!=nil{t.Fatal(err)}
    reconcileHeldShortcuts(c)
    if heldShortcutCount()!=1{t.Fatal("reconciliation unexpectedly released valid toggle")}
    m=c.Models["15"]
    m.Profiles[0].Keys[0].Toggle=false
    c.Models["15"]=m
    reconcileHeldShortcuts(c)
    if heldShortcutCount()!=0||len(*events)!=4{
        t.Fatalf("not released on toggle setting change: %+v",*events)
    }
    m=c.Models["15"]
    m.Profiles[0].Keys[0].Type="website"
    m.Profiles[0].Keys[0].Toggle=true
    c.Models["15"]=m
    clean:=normalize(c)
    if clean.Models["15"].Profiles[0].Keys[0].Toggle{
        t.Fatal("non-shortcut actions must never retain a toggle")
    }
}

func TestShortcutToggleBadInputDoesNotPressKeys(t *testing.T){
    events:=recordToggleKeyboard(t)
    if _,err:=toggleShortcut("CTRL+INVALID","error","15","R",0,1);err==nil{
        t.Fatal("accepted invalid chord")
    }
    if heldShortcutCount()!=0||len(*events)!=0{
        t.Fatal("failed parse must not inject key events")
    }
}
