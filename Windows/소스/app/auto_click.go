package main

import (
    "fmt"
    "strconv"
    "strings"
)

type autoClickSpec struct {
    IntervalMS int
    Count int // 0 = unlimited
    Button string
}

func defaultAutoClick() autoClickSpec { return autoClickSpec{IntervalMS:100,Count:0,Button:"left"} }

func formatAutoClick(c autoClickSpec) string {
    return fmt.Sprintf("%d|%d|%s", c.IntervalMS, c.Count, c.Button)
}

func parseAutoClick(raw string) (autoClickSpec,error) {
    if strings.TrimSpace(raw)=="" { return defaultAutoClick(),nil }
    parts:=strings.Split(raw,"|")
    if len(parts)!=3 {return autoClickSpec{},fmt.Errorf("오토클릭 설정 형식이 올바르지 않습니다")}
    interval,e:=strconv.Atoi(strings.TrimSpace(parts[0]))
    if e!=nil||interval<20||interval>10000 {
        return autoClickSpec{},fmt.Errorf("클릭 간격은 20~10000ms여야 합니다")
    }
    count,e:=strconv.Atoi(strings.TrimSpace(parts[1]))
    if e!=nil||count<0||count>100000 {
        return autoClickSpec{},fmt.Errorf("반복 횟수는 0~100000회여야 합니다 (0=무제한)")
    }
    button:=strings.TrimSpace(parts[2])
    if button!="left"&&button!="right" {
        return autoClickSpec{},fmt.Errorf("클릭 버튼은 왼쪽 또는 오른쪽이어야 합니다")
    }
    return autoClickSpec{IntervalMS:interval,Count:count,Button:button},nil
}
