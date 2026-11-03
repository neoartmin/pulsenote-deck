package main

import (
    "strings"
    "testing"
)

func TestAdvancedWebAndEmailCompose(t *testing.T){
    cases:=[]struct{input,part string}{
        {"google|전자기학","https://www.google.com/search?q="},
        {"naver|회로","https://search.naver.com/search.naver?query="},
        {"youtube|실험","https://www.youtube.com/results?search_query="},
    }
    for _,tc:=range cases{
        uri,e:=webSearchURL(tc.input)
        if e!=nil||!strings.HasPrefix(uri,tc.part){t.Fatalf("%s: %q %v",tc.input,uri,e)}
    }
    for _,bad:=range []string{"","invalid|term"}{
        if _,e:=webSearchURL(bad);e==nil{t.Fatalf("accepted invalid %q",bad)}
    }
    draft,e:=parseEmailDraft("test@example.com|제목|본문\n두 줄")
    if e!=nil||draft.to!="test@example.com"||draft.subject!="제목"||!strings.Contains(draft.body,"\n"){
        t.Fatalf("email parse failed: %+v %v",draft,e)
    }
    uri,e:=emailDraftURI("test@example.com|제목|본문")
    if e!=nil||!strings.HasPrefix(uri,"mailto:test@example.com?")||!strings.Contains(uri,"subject="){
        t.Fatalf("compose mail URI invalid %q %v",uri,e)
    }
    for _,bad:=range []string{"bad-address|hello","test@example.com\nBCC:evil|hello"}{
        if _,e:=emailDraftURI(bad);e==nil{t.Fatalf("accepted unsafe email %q",bad)}
    }
}

func TestAdvancedChainLimits(t *testing.T){
    chain,e:=parseChain("2,wait:250,4",15,1)
    if e!=nil||len(chain)!=3||!chain[1].isWait()||chain[1].waitMS!=250||chain[2].key!=4{
        t.Fatalf("unexpected chain %+v err=%v",chain,e)
    }
    for _,bad:=range []string{"","1","16","wait:-1","wait:1001","2,3,4,5,6,7,8,9,10"}{
        if _,e:=parseChain(bad,15,1);e==nil{t.Fatalf("accepted invalid chain %q",bad)}
    }
    chain,e=parseChain("2,3,4",12,1)
    if e!=nil||len(chain)!=3{t.Fatalf("PN12 chain invalid: %v",e)}
}

func TestAdvancedActionsAndSettingsPersistence(t *testing.T){
    kinds:=[]struct{name,label string;field bool}{
        {"web_search","웹 검색",true},
        {"compose_email","이메일 작성",true},
        {"qr_code","QR코드 생성",true},
        {"sequence","연속 동작",true},
        {"clipboard_history","클립보드 기록",false},
        {"window_control","창 제어",true},
    }
    for _,tc:=range kinds{
        if !validTypes[tc.name]||typeName(tc.name)!=tc.label{
            t.Fatalf("action registration missing: %s",tc.name)
        }
        if needsValue(tc.name)!=tc.field{t.Errorf("action input mismatch: %s",tc.name)}
        index:=-1
        for i,v:=range types{if v==tc.name{index=i;break}}
        if index<0{t.Fatalf("GUI picker lacks %s",tc.name)}
        if index/16>=((len(types)+15)/16){t.Fatalf("pager unreachable %s",tc.name)}
        for _,model:=range []int{12,15}{
            c:=defaults()
            c.Selected=model
            if model==15{c.SelectedSide="L"}
            if !mutateKey(&c,0,func(k *Key){k.Type=tc.name;k.Label="테스트";k.Value="2,3"}){
                t.Fatalf("key mutation rejected %s",tc.name)
            }
            normalized:=normalize(c)
            saved,ok:=selectedKey(normalized,0)
            if !ok||saved.Type!=tc.name||saved.Value!="2,3"{
                t.Fatalf("profile lost %s on PN%d: %+v",tc.name,model,saved)
            }
        }
    }
    if windowActionName("desk_next")!="다음 데스크톱"{
        t.Fatal("window control selection missing")
    }
    if len(types)<20{t.Fatalf("new actions were not appended: %d",len(types))}
}
