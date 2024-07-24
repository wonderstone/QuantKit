// test ema.go

package indicator

import (
	"testing"
	"time"
)




func TestEMA(t *testing.T) {
	ema := NewEMA("EMA", 3, "Close")
	// ema的公式貌似不止一个，这里的公式是：EMA = (2 * close + (N - 1) * EMA') / (N + 1)
	ema.LoadData(1)
	tmp := ema.Eval()
	// （2 * 1 + (3 - 1) * 0）/ (3 + 1) = 0.5
	if tmp != 0.5 {
		t.Errorf("EMA Eval() error, got %f, expected %f", tmp, 0.5)
	}
	
	ema.LoadData(2)
	tmp = ema.Eval()
	// （2 * 2 + (3 - 1) * 0.5）/ (3 + 1) = 1.25
	if tmp != 1.25 {
		t.Errorf("EMA Eval() error, got %f, expected %f", tmp, 1.25)
	}
	// do it again should be the same
	tmp = ema.Eval()
	if tmp != 1.25 {
		t.Errorf("EMA Eval() error, got %f, expected %f", tmp, 1.25)
	}

	ema.LoadData(3)
	tmp = ema.Eval()
	// （2 * 3 + (3 - 1) * 1.25）/ (3 + 1) = 2.125
	if tmp !=  2.125 {
		t.Errorf("EMA Eval() error, got %f, expected %f", tmp, 2.125)
	}
	ema.LoadData(4)
	tmp = ema.Eval()
	// （2 * 4 + (3 - 1) * 2.125）/ (3 + 1) = 3.0625
	if tmp != 3.0625 {
		t.Errorf("EMA Eval() error, got %f, expected %f", tmp, 3.0625)
	}
	// test DoCalculate
	tr:= &tmpRecordFunc{}
	tr.Data = []string{"5"}
	tr.Header = map[string]int{"Close": 0}

	ttt := time.Now()
	// （2 * 5 + (3 - 1) * 3.0625）/ (3 + 1) = 4.03125
	// 4.03125 -> 4.03	
	tmpstr:= ema.DoCalculate(ttt, tr)
	if tmpstr != "4.0312" {
		t.Errorf("EMA DoCalculate() error, got %s, expected %s", tmpstr, "4.03")
	}


	ema2:=  NewEMA("EMA", 4, "Close")
	ema2.LoadData(1)
	tmp = ema2.Eval()
	// （2 * 1 + (4 - 1) * 0）/ (4 + 1) = 0.4
	if tmp != 0.4 {
		t.Errorf("EMA Eval() error, got %f, expected %f", tmp, 0.4)
	}

	ema2.LoadData(2)
	tmp = ema2.Eval()
	// （2 * 2 + (4 - 1) * 0.4）/ (4 + 1) = 1.04
	if tmp != 1.04 {
		t.Errorf("EMA Eval() error, got %f, expected %f", tmp, 1.04)
	}

	// test DoCalculate data is empty
	tr.Data = []string{""}
	ttt = time.Now()
	tmpstr = ema2.DoCalculate(ttt, tr)
	if tmpstr != "" {
		t.Errorf("EMA DoCalculate() error, got %s, expected %s", tmpstr, "")
	}
	
}