//go:build windows

package main

import (
    "fmt"
    "time"
)

// Reuses existing configured key actions for this device's current profile.
// The serial action worker already enforces ordered key execution. We prohibit
// nested sequences and cap delays to avoid unbounded loops or queue starvation.
func runActionForDevice(k Key, c Config, model int, side string, self int) error {
    if k.Type!="sequence" { return runAction(k) }
    steps,err:=parseChain(k.Value,model,self)
    if err!=nil{return err}
    keys:=make([]Key,len(steps))
    // Prevalidate every target BEFORE running any side-effectful step.
    for i,step:=range steps {
        if step.isWait(){continue}
        item,ok:=resolveHardwareKey(c,model,side,step.key)
        if !ok{return fmt.Errorf("연속 동작 %d: 키 %d가 없습니다",i+1,step.key)}
        if item.Type=="sequence"{
            return fmt.Errorf("연속 동작에는 다른 연속 동작 키를 넣을 수 없습니다")
        }
        keys[i]=item
    }
    for i,step:=range steps {
        if step.isWait(){time.Sleep(time.Duration(step.waitMS)*time.Millisecond);continue}
        if err:=runAction(keys[i]);err!=nil{
            return fmt.Errorf("연속 동작 %d단계 (키 %d): %w",i+1,step.key,err)
        }
    }
    return nil
}
