package main

import (
    "fmt"
    "net/url"
    "strconv"
    "strings"
)

const (
    maxAdvancedValue = 1600
    maxSequenceSteps = 8
)

var windowActionIDs=[]string{"maximize","minimize","left","right","desktop","desk_prev","desk_next"}
var windowActionNames=[]string{"최대화","최소화","왼쪽 정렬","오른쪽 정렬","바탕화면 보기","이전 데스크톱","다음 데스크톱"}

func windowActionIndex(value string) int{
    for i,k:=range windowActionIDs{if value==k{return i}}
    return 0
}
func windowActionName(value string)string{return windowActionNames[windowActionIndex(value)]}

func searchEngine(raw string)string{
    if before,_,found:=strings.Cut(raw,"|");found{
        for _,id:=range []string{"google","naver","youtube"}{
            if strings.EqualFold(strings.TrimSpace(before),id){return id}
        }
    }
    return "google"
}
func searchQuery(raw string)string{
    if before,after,ok:=strings.Cut(raw,"|");ok{
        for _,id:=range []string{"google","naver","youtube"}{
            if strings.EqualFold(strings.TrimSpace(before),id){return after}
        }
    }
    return raw
}
func makeSearchValue(engine,query string)string{return engine+"|"+query}

func webSearchURL(raw string)(string,error){
    text:=strings.TrimSpace(raw)
    engine:="google"
    if before,after,ok:=strings.Cut(text,"|");ok{
        engine=strings.ToLower(strings.TrimSpace(before))
        text=strings.TrimSpace(after)
    }
    if text==""||len([]rune(text))>maxAdvancedValue{return "",fmt.Errorf("검색어를 1~%d자로 입력하세요",maxAdvancedValue)}
    query:=url.QueryEscape(text)
    switch engine{
    case "google":return "https://www.google.com/search?q="+query,nil
    case "naver":return "https://search.naver.com/search.naver?query="+query,nil
    case "youtube":return "https://www.youtube.com/results?search_query="+query,nil
    default:return "",fmt.Errorf("검색 엔진은 google, naver, youtube만 지원합니다")
    }
}

type emailDraft struct{to,subject,body string}
// New structured email editor persists all three fields with escaped
// delimiters. Old legacy "to|subject|body" profiles remain readable.
func encodeEmailDraft(d emailDraft)string{
    return "mail2:"+url.QueryEscape(d.to)+"|"+url.QueryEscape(d.subject)+"|"+url.QueryEscape(d.body)
}
func parseEmailDraft(raw string)(emailDraft,error){
    var parts []string
    if strings.HasPrefix(raw,"mail2:"){
        parts=strings.SplitN(strings.TrimPrefix(raw,"mail2:"),"|",3)
        if len(parts)!=3{return emailDraft{},fmt.Errorf("이메일 편집기 데이터가 손상되었습니다")}
        for i:=range parts{
            unescaped,err:=url.QueryUnescape(parts[i])
            if err!=nil{return emailDraft{},fmt.Errorf("이메일 데이터 디코딩 실패: %w",err)}
            parts[i]=unescaped
        }
    }else{
        parts=strings.SplitN(raw,"|",3)
    }
    draft:=emailDraft{to:strings.TrimSpace(parts[0])}
    if len(parts)>1{draft.subject=parts[1]}
    if len(parts)>2{draft.body=parts[2]}
    if len(raw)==0||len(raw)>4096{return draft,fmt.Errorf("이메일 초안은 1~4096자여야 합니다")}
    for _,r:=range draft.to {
        if r=='\r'||r=='\n'||r==' '||r=='?'||r=='#'||r=='&'||r==';'{
            return draft,fmt.Errorf("받는 사람은 단일 이메일 주소만 입력하세요")
        }
    }
    if draft.to!=""&&!strings.Contains(draft.to,"@"){return draft,fmt.Errorf("이메일 주소 형식이 올바르지 않습니다")}
    return draft,nil
}
func emailDraftURI(raw string)(string,error){
    d,err:=parseEmailDraft(raw)
    if err!=nil{return "",err}
    query:=url.Values{}
    if d.subject!=""{query.Set("subject",d.subject)}
    if d.body!=""{query.Set("body",d.body)}
    // Leave the address intact; only permit safe address characters.
    uri:="mailto:"+d.to
    if len(query)>0{uri+="?"+query.Encode()}
    return uri,nil
}

type chainStep struct{key,waitMS int}
func (s chainStep) isWait()bool{return s.waitMS>=0}
func parseChain(raw string,keys,self int)([]chainStep,error){
    if strings.TrimSpace(raw)==""{return nil,fmt.Errorf("연속 실행 키를 입력하세요")}
    parts:=strings.FieldsFunc(raw,func(r rune)bool{return r==','||r=='\n'})
    if len(parts)==0||len(parts)>maxSequenceSteps{
        return nil,fmt.Errorf("연속 실행은 최대 %d단계입니다",maxSequenceSteps)
    }
    steps:=make([]chainStep,0,len(parts))
    for _,token:=range parts{
        token=strings.TrimSpace(token)
        if token==""{return nil,fmt.Errorf("빈 단계가 있습니다")}
        if strings.HasPrefix(strings.ToLower(token),"wait:"){
            ms,e:=strconv.Atoi(strings.TrimSpace(token[5:]))
            if e!=nil||ms<0||ms>1000{return nil,fmt.Errorf("대기 시간은 0~1000ms여야 합니다")}
            steps=append(steps,chainStep{key:-1,waitMS:ms})
        }else{
            num,e:=strconv.Atoi(token)
            if e!=nil||num<1||num>keys||num==self{
                return nil,fmt.Errorf("자기 자신을 제외한 키 번호 1~%d를 입력하세요",keys)
            }
            steps=append(steps,chainStep{key:num,waitMS:-1})
        }
    }
    return steps,nil
}
