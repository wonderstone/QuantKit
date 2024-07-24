package indicator

import (
	"fmt"
	"testing"
	"time"
)

func TestMad(t *testing.T) {
	mad := NewMAD("MAD", 3, "Close", "D")
	fmt.Println(mad)
	// test LoadData
	mad.LoadData(1.0)
	mad.LoadData(2.0)
	mad.LoadData(3.0)
	// test eval
	tmp := mad.Eval()
	if tmp != 2.0 {
		t.Error("MAD.Eval() error")
	}
	// test DoCalculate
	// 第一次计算，日期不同，压入并计算新值
	tr := &tmpRecordFunc{}
	tr.Data = []string{"4.0", "5.0", "6.0"}
	tr.Header = map[string]int{"close": 0}
	ttt := time.Now()

	tmpstr := mad.DoCalculate(ttt, tr)
	if tmpstr != "3.0000" {
		t.Error("MAD.DoCalculate() error")
	}

	// 第二次计算，日期相同，不压入新值
	ttt = time.Now()
	tr.Data = []string{"7.0", "6.0", "7.0"}
	tmpstr = mad.DoCalculate(ttt, tr)
	if tmpstr != "3.0000" {
		t.Error("MAD.DoCalculate() error")
	}

	// 第三次计算，日期不同，压入并计算新值
	ttt = time.Now().AddDate(0, 0, 1)
	tr.Data = []string{"8.0", "9.0", "10.0"}
	tmpstr = mad.DoCalculate(ttt, tr)
	if tmpstr != "5.0000" {
		t.Error("MAD.DoCalculate() error")
	}



} 