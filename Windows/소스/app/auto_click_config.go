package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Auto-click configuration is stable across PN12, PN15-L and PN15-R.
// The last field is the click limit: zero means repeat until stopped.
type autoClickSettings struct {
	Button string
	IntervalMS int
	Limit int
}

func defaultAutoClickSettings() autoClickSettings {
	return autoClickSettings{Button:"left",IntervalMS:100,Limit:0}
}

func (s autoClickSettings) encode() string {
	return fmt.Sprintf("%s|%d|%d",s.Button,s.IntervalMS,s.Limit)
}

func parseAutoClickSettings(raw string) (autoClickSettings,error) {
	if strings.TrimSpace(raw)=="" { return defaultAutoClickSettings(),nil }
	parts:=strings.Split(raw,"|")
	if len(parts)!=3 {return autoClickSettings{},fmt.Errorf("오토클릭 설정이 올바르지 않습니다")}
	button:=strings.ToLower(strings.TrimSpace(parts[0]))
	if button!="left" && button!="right" {return autoClickSettings{},fmt.Errorf("왼쪽/오른쪽 클릭만 지원합니다")}
	ms,e:=strconv.Atoi(strings.TrimSpace(parts[1]))
	if e!=nil||ms<20||ms>10000 {return autoClickSettings{},fmt.Errorf("클릭 간격은 20~10000ms여야 합니다")}
	limit,e:=strconv.Atoi(strings.TrimSpace(parts[2]))
	if e!=nil||limit<0||limit>100000 {return autoClickSettings{},fmt.Errorf("반복 횟수는 0~100000회여야 합니다")}
	return autoClickSettings{Button:button,IntervalMS:ms,Limit:limit},nil
}

func nextOutputIndex(count,current int) int {
	if count<2 {return -1}
	if current<0||current>=count{return 0}
	return (current+1)%count
}
