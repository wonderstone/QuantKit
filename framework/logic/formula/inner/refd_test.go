package indicator

import (
	"testing"
	"time"
)

// func to test ma.go
func TestRefd(t *testing.T) {
	// new ref
	refd := NewRefd("Ref", 1, "close","D")
	// test LoadData
	refd.LoadData(1.0)
	refd.LoadData(2.0)
	refd.LoadData(3.0)
	refd.LoadData(4.0)
	// test eval
	tmp :=refd.Eval()
	if tmp != 3.0 {
		t.Error("ref.Eval() error")
	}
	// test DoCalculate
	tr:= &tmpRecordFunc{}
	tr.Data = []string{"5.0"}
	tr.Header = map[string]int{"close": 0}

	ttt := time.Now()
	tmpstr:= refd.DoCalculate(ttt, tr)
	
	if tmpstr != "4.0000" {
		t.Error("Ref.DoCalculate() error")
	}
	// 同一天不同时间 不会更新
	ttt = time.Now()
	tmpstr= refd.DoCalculate(ttt, tr)
	if tmpstr != "4.0000" {
		t.Error("Ref.DoCalculate() error")
	}
	// 不同天 会更新
	ttt = ttt.AddDate(0,0,1)
	tmpstr= refd.DoCalculate(ttt, tr)
	if tmpstr != "5.0000" {
		t.Error("Ref.DoCalculate() error")
	}

} 