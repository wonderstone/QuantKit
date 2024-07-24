package indicator

import (
	"testing"
	"time"
)

// func to test ma.go
func TestRef(t *testing.T) {
	// new ref
	ref := NewRef("Ref", 1, "close")
	// test LoadData
	ref.LoadData(1.0)
	ref.LoadData(2.0)
	ref.LoadData(3.0)
	ref.LoadData(4.0)
	// test eval
	tmp :=ref.Eval()
	if tmp != 3.0 {
		t.Error("ref.Eval() error")
	}
	// test DoCalculate
	tr:= &tmpRecordFunc{}
	tr.Data = []string{"5.0"}
	tr.Header = map[string]int{"close": 0}

	ttt := time.Now()
	tmpstr:= ref.DoCalculate(ttt, tr)
	
	if tmpstr != "4.0000" {
		t.Error("Ref.DoCalculate() error")
	}
	// 同一时间 不会更新
	tmpstr= ref.DoCalculate(ttt, tr)
	if tmpstr != "4.0000" {
		t.Error("Ref.DoCalculate() error")
	}
	// 不同天 会更新
	ttt = ttt.AddDate(0,0,1)
	tmpstr= ref.DoCalculate(ttt, tr)
	if tmpstr != "5.0000" {
		t.Error("Ref.DoCalculate() error")
	}

	// 空数据  比如输入是ref的第一天情况  返回 空字符串 但是数据仍然保留

	tr.Data = []string{""}
	ttt = time.Now()
	tmpstr = ref.DoCalculate(ttt, tr)
	if tmpstr != "" {
		t.Error("Ref.DoCalculate() error")
	}	

} 