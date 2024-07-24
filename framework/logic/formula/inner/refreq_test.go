package indicator

import (
	"testing"
	"time"
)

// func to test ma.go
func TestRefreq(t *testing.T) {
	// new ref
	refreq := NewRefreq("Ref", 2, 1, "close")
	// test LoadData
	refreq.LoadData(1.0)
	refreq.LoadData(2.0)
	refreq.LoadData(3.0)
	refreq.LoadData(4.0)
	// test eval
	tmp :=refreq.Eval()
	if tmp != 3.0 {
		t.Error("ref.Eval() error")
	}
	// test DoCalculate
	tr:= &tmpRecordFunc{}
	tr.Data = []string{"5.0"}
	tr.Header = map[string]int{"close": 0}

	ttt := time.Now()
	tmpstr:= refreq.DoCalculate(ttt, tr)
	
	if tmpstr != "3.0000" {
		t.Error("Ref.DoCalculate() error")
	}
	// 同一天不同时间 会更新
	ttt = time.Now()
	tmpstr= refreq.DoCalculate(ttt, tr)
	if tmpstr != "4.0000" {
		t.Error("Ref.DoCalculate() error")
	}
	// 不同天 会更新
	ttt = ttt.AddDate(0,0,1)
	tmpstr= refreq.DoCalculate(ttt, tr)
	if tmpstr != "4.0000" {
		t.Error("Ref.DoCalculate() error")
	}

	ttt = time.Now()
	tmpstr= refreq.DoCalculate(ttt, tr)
	if tmpstr != "5.0000" {
		t.Error("Ref.DoCalculate() error")
	}

} 