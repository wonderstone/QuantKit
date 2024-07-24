package indicator

import (
	"fmt"
	"testing"
	"time"
)

// func to test ma.go
func TestCsvGetter(t *testing.T) {
	//
	// gt := NewGetter("test",
	// 	"../../../../CZM-dir/download/1dayfactor/standard-aaa.csv",
	// 	"Date",
	// 	"data_code",
	// 	"b1_factor_value",
	// 	"000501.SZ.XSHE.CS",)

	cr1 := NewCSVGetter("test",
		"end_date",
		"data_code",
		"b2_factor_value",
		"000019.XSHE.CS",
		"LV",
	)

	// test DoCalculate
	tr := &tmpRecordFunc{}
	tr.Data = []string{"4.0"}
	// let ttt be 2021.04.02
	ttt := time.Date(2022, time.April, 28, 0, 0, 0, 0, time.UTC)
	tmpstr := cr1.DoCalculate(ttt, tr)
	if tmpstr !=  "0.4639175257731959"  {
		t.Error("Getter.DoCalculate() error")
	}

	ttt = ttt.AddDate(0, 0, 1)
	tmpstr = cr1.DoCalculate(ttt, tr)
	if tmpstr !=  ""  {
		t.Error("Getter.DoCalculate() error")
	}



	// ttt += 1 day
	ttt = ttt.AddDate(0, 0, 6)
	tmpstr = cr1.DoCalculate(ttt, tr)
	if tmpstr !=  "0.77"  {
		t.Error("Getter.DoCalculate() error")
	}


	// test NoLV mode
	cr2 := NewCSVGetter("test",
		"end_date",
		"data_code",
		"b2_factor_value",
		"000037.XSHE.CS",
		"NoLV",
	)

	// test DoCalculate
	tr = &tmpRecordFunc{}
	tr.Data = []string{"4.0"}
	// let ttt be 2021.04.02
	ttt = time.Date(2021, time.April, 2, 0, 0, 0, 0, time.UTC)
	tmpstr = cr2.DoCalculate(ttt, tr)
	if tmpstr !=  "0.13095238095238096"  {
		t.Error("Getter.DoCalculate() error")
	}	

	// ttt += 1 day
	ttt = ttt.AddDate(0, 0, 1)
	tmpstr = cr2.DoCalculate(ttt, tr)
	if tmpstr !=  ""  {
		t.Error("Getter.DoCalculate() error")
	}

	






	// test get CSVData
	tmp := CSVData["2021.04.02"]["000037.SZ"]
	// tmp["b1_factor_value"] = "0.0"
	fmt.Println(tmp["b1_factor_value"])
	if CSVData["2021.04.02"] == nil {
		t.Error("CSVData is nil")
	}
	if CSVData["2021.04.02"]["000037.SZ"]["b2_factor_value"] != "0.13095238095238096" {
		t.Error("CSVData is not correct")
	}



}
