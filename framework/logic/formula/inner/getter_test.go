package indicator

import (
	"testing"
	"time"
)

// func to test ma.go
func TestGetter(t *testing.T) {
	//
	// gt := NewGetter("test",
	// 	"../../../../CZM-dir/download/1dayfactor/standard-aaa.csv",
	// 	"Date",
	// 	"data_code",
	// 	"b1_factor_value",
	// 	"000501.SZ.XSHE.CS",)

	gt1 := NewGetter("test",
		"../../../../CZM-dir/download/1dayfactor/aaa.csv",
		"end_date",
		"data_code",
		"b1_factor_value",
		"000501.SZ.XSHE.CS",)

	// test DoCalculate
	tr := &tmpRecordFunc{}
	tr.Data = []string{"4.0"}

	ttt := time.Now()
	tmpstr := gt1.DoCalculate(ttt, tr)
	if tmpstr != "" {
		t.Error("Getter.DoCalculate() error")
	}

	// make a time at 2010-01-03
	ttt = time.Date(2010, 1, 3, 0, 0, 0, 0, time.UTC)
	tmpstr = gt1.DoCalculate(ttt, tr)
	if tmpstr != "0.19" {
		t.Error("Getter.DoCalculate() error")
	}
}
