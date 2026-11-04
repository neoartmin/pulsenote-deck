package main

import "testing"

func TestAutoClickSettings(t *testing.T) {
	for _,test:=range []autoClickSettings{
		defaultAutoClickSettings(),
		{Button:"right",IntervalMS:20,Limit:1},
		{Button:"left",IntervalMS:10000,Limit:100000},
	} {
		got,err:=parseAutoClickSettings(test.encode())
		if err!=nil||got!=test {t.Fatalf("roundtrip got %+v error %v",got,err)}
	}
	for _,bad:=range []string{"L|100|0","left|0|0","left|19|0","left|10001|0","left|100|-1","left|100|100001","left|x|0","left|100","left|100|0|other"} {
		if _,err:=parseAutoClickSettings(bad);err==nil{t.Fatalf("accepted %q",bad)}
	}
	if got,err:=parseAutoClickSettings("");err!=nil||got!=defaultAutoClickSettings(){t.Fatal("default invalid")}
}
func TestOutputCycle(t *testing.T) {
	for _,tc:=range []struct{count,current,want int}{
		{0,-1,-1},{1,0,-1},{2,0,1},{2,1,0},{3,2,0},{3,-1,0},
	}{
		if got:=nextOutputIndex(tc.count,tc.current);got!=tc.want{t.Fatalf("%+v got %d",tc,got)}
	}
}
