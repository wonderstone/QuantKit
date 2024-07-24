package indicator

// todo

import (
	"fmt"
	"testing"
	"time"
)

// func to test macdf.go
func TestMACDD(t *testing.T) {
	// !!! i did not check the result
	// !!! i did not check the result
	// !!! i did not check the result
	macdd := NewMACDD("MACDD", 12, 26, 9,  "Close", "D")
	fmt.Println(macdd)
	// test LoadData
	macdd.LoadData(1.0)
	macdd.LoadData(2.0)
	macdd.LoadData(3.0)
	// test eval
	tmp := macdd.Eval()
	if tmp != 0.6546250931150404 {
		t.Error("MACDF.Eval() error")
	}
	// test DoCalculate
	tr := &tmpRecordFunc{}
	tr.Data = []string{"4.0", "5.0", "6.0"}
	tr.Header = map[string]int{"close": 0}
	ttt := time.Now()

	tmpstr := macdd.DoCalculate(ttt, tr)
	if tmpstr != "0.8796" {
		t.Error("MACDF.DoCalculate() error")
	}
	ttt = time.Now()
	tr.Data = []string{"7.0", "6.0", "7.0"}
	tmpstr = macdd.DoCalculate(ttt, tr)
	if tmpstr != "0.8796" {
		t.Error("MACDF.DoCalculate() error")
	}
} 