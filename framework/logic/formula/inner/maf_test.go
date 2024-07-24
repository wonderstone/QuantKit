package indicator

import (
	"fmt"
	"testing"
	"time"
)

// func to test masf.go
func TestMAF(t *testing.T) {
	maf := NewMAF("MAF", 3, "Close", 2)
	fmt.Println(maf)
	// test LoadData
	maf.LoadData(1.0)
	maf.LoadData(2.0)
	maf.LoadData(3.0)
	// test eval
	tmp := maf.Eval()
	if tmp != 2.0 {
		t.Error("MAF.Eval() error")
	}
	// test DoCalculate
	// 第一次计算，未达到计算频率，返回上一值
	tr := &tmpRecordFunc{}
	tr.Data = []string{"4.0", "5.0", "6.0"}
	tr.Header = map[string]int{"close": 0}
	ttt := time.Now()

	tmpstr := maf.DoCalculate(ttt, tr)
	if tmpstr != "2.0000" {
		t.Error("MAF.DoCalculate() error")
	}
	// 第二次计算，达到计算频率，返回新值,此时压入的数值为7.0，1弹出.2,3,7的均值
	ttt = time.Now()
	tr.Data = []string{"7.0", "6.0", "7.0"}
	tmpstr = maf.DoCalculate(ttt, tr)
	if tmpstr != "4.0000" {
		t.Error("MAF.DoCalculate() error")
	}
}
 