package indicator

import (
	"fmt"
	"testing"
	"time"
)

// func to test macdf.go
func TestMACDF(t *testing.T) {
	// !!! i did not check the result
	// !!! i did not check the result
	// !!! i did not check the result
	macdf := NewMACDF("MACDF", 12, 26, 9, 2, "Close")
	fmt.Println(macdf)
	// test LoadData
	macdf.LoadData(1.0)
	macdf.LoadData(2.0)
	macdf.LoadData(3.0)
	// test eval
	tmp := macdf.Eval()
	if tmp != 0.6546250931150404 {
		t.Error("MACDF.Eval() error")
	}
	// test DoCalculate
	// 第一次计算，未达到计算频率，返回上一值
	tr := &tmpRecordFunc{}
	tr.Data = []string{"4.0", "5.0", "6.0"}
	tr.Header = map[string]int{"close": 0}
	ttt := time.Now()

	tmpstr := macdf.DoCalculate(ttt, tr)
	if tmpstr != "0.6546" {
		t.Error("MACDF.DoCalculate() error")
	}
	// 第二次计算，达到计算频率，返回新值,此时压入的数值为7.0，1弹出.2,3,7的均值
	ttt = time.Now()
	tr.Data = []string{"7.0", "6.0", "7.0"}
	tmpstr = macdf.DoCalculate(ttt, tr)
	if tmpstr != "1.2625" {
		t.Error("MACDF.DoCalculate() error")
	}
} 