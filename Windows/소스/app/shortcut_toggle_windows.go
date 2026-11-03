//go:build windows

package main

import (
    "fmt"
    "sync"
)

// One keyboard hold is scoped to a concrete PN12 / PN15-side/profile/key.
// Reference counts prevent releasing a modifier held by another toggle.
type heldShortcut struct {
    modelKey string
    profile, key int
    value, label string
    codes []byte
}
var shortcutHolds=struct {
    sync.Mutex
    holds map[string]heldShortcut
    refs map[byte]int
}{
    holds:make(map[string]heldShortcut),
    refs:make(map[byte]int),
}

func shortcutHoldID(model,side string,profile,key int)string{
    return fmt.Sprintf("%s/%d/%d",model,profile,key)
}

// Toggle a physical key hold: first press sends KEYDOWN exactly once, second
// press sends KEYUP exactly once, and each transition enqueues ONE notification.
func toggleShortcut(value,label string,model,side string,profile,key int)(bool,error){
    codes,err:=parseShortcut(value)
    if err!=nil{return false,err}
    id:=shortcutHoldID(model,side,profile,key)
    shortcutHolds.Lock()
    if old,ok:=shortcutHolds.holds[id];ok{
        releaseHeldCodesLocked(old.codes)
        delete(shortcutHolds.holds,id)
        shortcutHolds.Unlock()
        enqueueShortcutToggleNotice(label,false)
        return false,nil
    }
    for _,code:=range codes{
        if shortcutHolds.refs[code]==0{press(code,false)}
        shortcutHolds.refs[code]++
    }
    shortcutHolds.holds[id]=heldShortcut{
        modelKey:model,profile:profile,key:key,value:value,label:label,
        codes:append([]byte(nil),codes...),
    }
    shortcutHolds.Unlock()
    enqueueShortcutToggleNotice(label,true)
    return true,nil
}

func releaseHeldCodesLocked(codes []byte){
    for i:=len(codes)-1;i>=0;i--{
        code:=codes[i]
        if shortcutHolds.refs[code]<=1{
            delete(shortcutHolds.refs,code)
            press(code,true)
        }else{
            shortcutHolds.refs[code]--
        }
    }
}

// Standard momentary shortcuts must not accidentally release a modifier
// currently held by a separate toggle, e.g. held CTRL + ordinary CTRL+C.
func shortcutRespectingHolds(value string)error{
    codes,err:=parseShortcut(value)
    if err!=nil{return err}
    shortcutHolds.Lock()
    defer shortcutHolds.Unlock()
    for _,code:=range codes{
        if shortcutHolds.refs[code]==0{press(code,false)}
    }
    for i:=len(codes)-1;i>=0;i--{
        if shortcutHolds.refs[codes[i]]==0{press(codes[i],true)}
    }
    return nil
}

// Called only after a successful, persisted config update. Editing a held
// shortcut, switching profiles or changing its type releases the old hold.
func reconcileHeldShortcuts(c Config){
    shortcutHolds.Lock()
    released:=make([]string,0)
    for id,old:=range shortcutHolds.holds{
        m,ok:=c.Models[old.modelKey]
        valid:=ok&&m.Active==old.profile&&old.profile>=0&&old.profile<len(m.Profiles)
        if valid{
            keys:=m.Profiles[old.profile].Keys
            valid=old.key>=1&&old.key<=len(keys)
            if valid{
                now:=keys[old.key-1]
                valid=now.Type=="shortcut"&&now.Toggle&&now.Value==old.value
            }
        }
        if !valid{
            releaseHeldCodesLocked(old.codes)
            delete(shortcutHolds.holds,id)
            released=append(released,old.label)
        }
    }
    shortcutHolds.Unlock()
    for _,label:=range released{enqueueShortcutToggleNotice(label,false)}
}

// Best-effort safety cleanup: do not leave injected keys down across USB
// detach, application shutdown or emergency stop.
func releaseAllHeldShortcuts(){
    shortcutHolds.Lock()
    names:=make([]string,0,len(shortcutHolds.holds))
    for id,old:=range shortcutHolds.holds{
        releaseHeldCodesLocked(old.codes)
        delete(shortcutHolds.holds,id)
        names=append(names,old.label)
    }
    shortcutHolds.Unlock()
    for _,label:=range names{enqueueShortcutToggleNotice(label,false)}
}

func heldShortcutCount()int{
    shortcutHolds.Lock()
    defer shortcutHolds.Unlock()
    return len(shortcutHolds.holds)
}
