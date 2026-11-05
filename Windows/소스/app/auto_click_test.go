package main

import "testing"

func TestAutoClickConfig(t *testing.T) {
    def,e:=parseAutoClick("")
    if e!=nil||def.IntervalMS!=100||def.Count!=0||def.Button!="left"{
        t.Fatalf("bad defaults: %+v %v",def,e)
    }
    for _,tc:=range []struct{input string;ms,count int;button string}{
        {"20|0|left",20,0,"left"},
        {"100|10|right",100,10,"right"},
        {"10000|100000|left",10000,100000,"left"},
    }{
        got,e:=parseAutoClick(tc.input)
        if e!=nil||got.IntervalMS!=tc.ms||got.Count!=tc.count||got.Button!=tc.button||formatAutoClick(got)!=tc.input{
            t.Fatalf("%q => %+v %v",tc.input,got,e)
        }
    }
    for _,v:=range []string{"2","0|0|left","19|1|left","10001|1|left","100|-1|left","100|100001|left","100|10|middle","1|2|left|bad"}{
        if _,e:=parseAutoClick(v);e==nil{t.Errorf("accepted invalid %q",v)}
    }
}

func TestAutoClickProfilePersistence(t *testing.T){
    for _,m:=range []struct{count int;side string}{{12,"R"},{15,"R"},{15,"L"}}{
        c:=defaults();c.Selected=m.count;c.SelectedSide=m.side
        mutateKey(&c,0,func(k *Key){k.Type="auto_click";k.Value="150|100|right"})
        c=normalize(c)
        k,_:=selectedKey(c,0)
        if k.Type!="auto_click"||k.Value!="150|100|right"{
            t.Fatalf("lost auto-click for %d/%s: %+v",m.count,m.side,k)
        }
    }
}
