//go:build windows

package main

import "testing"

func TestAudioOutputChoice(t *testing.T){
    devices:=[]audioEndpoint{{"a","스피커"},{"b","헤드셋"},{"c","모니터"}}
    got,e:=chooseAudioOutput(devices,"a","cycle")
    if e!=nil||got.ID!="b"{t.Fatalf("first cycle: %+v %v",got,e)}
    got,e=chooseAudioOutput(devices,"c","")
    if e!=nil||got.ID!="a"{t.Fatalf("wrap around: %+v %v",got,e)}
    got,e=chooseAudioOutput(devices,"a","c")
    if e!=nil||got.ID!="c"{t.Fatalf("specific device: %+v %v",got,e)}
    if _,e=chooseAudioOutput(devices,"a","unknown");e==nil{t.Fatal("accepted offline target")}
    if _,e=chooseAudioOutput(devices[:1],"a","cycle");e==nil{t.Fatal("accepted single device cycling")}
}
