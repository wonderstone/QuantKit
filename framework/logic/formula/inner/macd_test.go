// test macd

package indicator

import (
	"testing"
	"time"
)

func TestMACD(t *testing.T) {

	// !!! i did not check the result
	// !!! i did not check the result
	// !!! i did not check the result
	macd := NewMACD("MACD", 12, 26, 9,"Close")
	macd.LoadData(1)
	tmp:= macd.Eval()
	if tmp != 0.12763532763532764 {
		t.Error("macd eval error")
	}

	// test DoCalculate
	tr:= &tmpRecordFunc{}
	tr.Data = []string{"2"}
	tr.Header = map[string]int{"Close": 0}

	ttt := time.Now()
	tmpstr:= macd.DoCalculate(ttt,tr)
	
	if tmpstr != "0.3283" {
		t.Errorf("MACD DoCalculate() error, got %s, expected %s", tmpstr, "0.59")
	}





	
	
} 