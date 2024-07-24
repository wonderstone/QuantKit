package indicator

import (
	"fmt"
	"testing"
	"time"
	// "github.com/wonderstone/QT/tools/config"
	// "github.com/wonderstone/QT/framework/entity/formula"
)


func TestNewDailyFactor(t *testing.T) {
	// Create mock parameters
	name := "pe10"
	// Call the NewDailyFactor function
	df := NewDailyFactor(name,  "600000.XSHG.CS", "train", "/Users/wonderstone/Desktop/QT/DMTF-dir/download/factor","pe")

	if df.Name != name {
		t.Errorf("NewDailyFactor() = %v, want %v", df.Name, name)
	}

	// test DoCalculate
	tr:= &tmpRecordFunc{}
	tr.Data = []string{"5.0"}
	tr.Header = map[string]int{"close": 0}
	// fake a time from string like "2021-01-01 00:00:00 +0000 UTC"
	ttt, _ := time.Parse("2006-01-02 15:04:05 +0000 UTC", "2016-08-23 23:30:00 +0000 UTC")
	// 2016-08-19 7.4681
	// 2016-08-22 7.3817
	// 2016-08-23 7.3817
	// 2016-08-24 7.4053
	// 2016-08-25 7.4132
	
	 
	tmpstr:= df.DoCalculate(ttt, tr)
	fmt.Println(ttt.Format("2006-01-02"))
	if tmpstr != "7.3817" {
		t.Error("DailyFactor.DoCalculate() error")
	}

	// @ -1 day  should be 2016-08-22
	ttt = ttt.AddDate(0, 0, -1)
	fmt.Println(ttt.Format("2006-01-02"))

	tmpstr = df.DoCalculate(ttt, tr)
	if tmpstr != "7.3817" {
		t.Error("DailyFactor.DoCalculate() error")
	}

	// @ -1 day again  should be 2016-08-21 but no data 
	// @ then use 2016-08-19 's value  which is 7.4681 
	// @ so the result should be 7.47 which is the rounded value 
	ttt = ttt.AddDate(0, 0, -1)
	fmt.Println(ttt.Format("2006-01-02"))

	tmpstr = df.DoCalculate(ttt, tr)
	if tmpstr != "7.4681" {
		t.Error("DailyFactor.DoCalculate() error")
	}



}

