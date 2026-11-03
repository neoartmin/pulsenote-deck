//go:build windows

package main

import (
    "encoding/json"
    "fmt"
    "strings"
    "testing"
)

func drainToggleEvents() []shortcutToggleNotice {
    result:=[]shortcutToggleNotice{}
    for {
        select { case n:=<-shortcutToggleEvents: result=append(result,n)
        default: return result }
    }
}

func testToggleCapture(t *testing.T) (*[]string,func()){
    t.Helper()
    releaseAllHeldShortcuts()
    drainToggleEvents()
    original:=emitToggleKey
    entries:=[]string{}
    emitToggleKey=func(code byte,up bool){
        event:="down"
        if up {event="up"}
        entries=append(entries,fmt.Sprintf("%02x/%s",code,event))
    }
    cleanup:=func(){
        releaseAllHeldShortcuts()
        emitToggleKey=original
        drainToggleEvents()
    }
    t.Cleanup(cleanup)
    return &entries,cleanup
}

func TestShortcutTogglePressOnceAndReleaseOnce(t *testing.T){
    events,_:=testToggleCapture(t)
    on,e:=toggleShortcut("CTRL+A","테스트","12","R",0,1)
    if e!=nil||!on||heldShortcutCount()!=1{t.Fatalf("toggle ON failed: %v",e)}
    if !isHeldShortcut("12",0,1){t.Fatal("UI cannot see active hold")}
    off,e:=toggleShortcut("CTRL+A","테스트","12","R",0,1)
    if e!=nil||off||heldShortcutCount()!=0{t.Fatalf("toggle OFF failed: %v",e)}
    want:="11/down,41/down,41/up,11/up"
    if got:=strings.Join(*events,",");got!=want{t.Fatalf("key transitions %s, expected %s",got,want)}
    notes:=drainToggleEvents()
    if len(notes)!=2||!notes[0].active||notes[1].active{
        t.Fatalf("one alert for each transition expected: %+v",notes)
    }
}

func TestSharedModifierRetainsPhysicalHoldAcrossToggles(t *testing.T){
    events,_:=testToggleCapture(t)
    toggleShortcut("CTRL+A","하나","12","R",0,1)
    toggleShortcut("CTRL+B","둘","12","R",0,2)
    toggleShortcut("CTRL+A","하나","12","R",0,1)
    if len(*events)!=4 || (*events)[0]!="11/down" ||
       (*events)[1]!="41/down" || (*events)[2]!="42/down" || (*events)[3]!="41/up"{
        t.Fatalf("CTRL was released early: %+v",*events)
    }
    if err:=shortcutRespectingHolds("CTRL+C");err!=nil{t.Fatal(err)}
    toggleShortcut("CTRL+B","둘","12","R",0,2)
    if got:=strings.Join(*events,",");!strings.HasSuffix(got,",43/down,43/up,42/up,11/up"){
        t.Fatalf("incorrect key handling for simultaneous holds: %s",got)
    }
    if len(drainToggleEvents())!=4{t.Fatal("expected exactly 4 notifications")}
}

func TestHoldReleasedWhenProfileOrAssignmentChanges(t *testing.T){
    events,_:=testToggleCapture(t)
    c:=defaults()
    c.Selected=12
    mutateKey(&c,0,func(k *Key){
        k.Type="shortcut";k.Value="CTRL+A";k.Toggle=true
    })
    if _,err:=toggleShortcut("CTRL+A","고정","12","R",0,1);err!=nil{t.Fatal(err)}
    m:=c.Models["12"]
    m.Profiles[m.Active].Keys[0].Value="CTRL+B"
    c.Models["12"]=m
    reconcileHeldShortcuts(c)
    if heldShortcutCount()!=0{t.Fatal("editing held key did not release injected keys")}
    if got:=strings.Join(*events,",");!strings.HasSuffix(got,"41/up,11/up"){
        t.Fatalf("held keys not released: %s",got)
    }
    notes:=drainToggleEvents()
    if len(notes)!=2||notes[1].active{t.Fatalf("expected release notice %+v",notes)}
}

func TestShortcutTogglePersistAndLegacyActionsPreserved(t *testing.T){
    var c Config=defaults()
    mutateKey(&c,0,func(k *Key){k.Type="shortcut";k.Value="CTRL";k.Toggle=true})
    blob,err:=json.Marshal(normalize(c))
    if err!=nil{t.Fatal(err)}
    var back Config
    if err=json.Unmarshal(blob,&back);err!=nil{t.Fatal(err)}
    saved,ok:=selectedKey(normalize(back),0)
    if !ok||!saved.Toggle{t.Fatal("shortcut toggle did not persist")}
    mutateKey(&back,0,func(k *Key){k.Type="copy_text"})
    changed,ok:=selectedKey(normalize(back),0)
    if !ok||changed.Toggle{t.Fatal("non-shortcut retained toggle")}
    if !validTypes["compose_email"]||!validTypes["sequence"]{
        t.Fatal("legacy email or sequence settings must remain readable")
    }
}

func TestIntuitivePickerAndEveryActionHelp(t *testing.T){
    for _,ty:=range types{
        help:=guideForAction(ty)
        if strings.TrimSpace(help.summary)==""||strings.TrimSpace(help.detail)==""{
            t.Fatalf("%s lacks functional help",ty)
        }
        if ty=="compose_email"||ty=="sequence"{
            t.Fatalf("non-intuitive action exposed as new option: %s",ty)
        }
    }
    for _,legacy:=range []string{"compose_email","sequence"}{
        h:=guideForAction(legacy)
        if h.summary==""||h.detail==""{t.Fatalf("legacy %s lost help",legacy)}
    }
}
