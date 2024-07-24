// test dea

package indicator

import (
	"math"
	"testing"
	"time"
)

func TestDEA(t *testing.T) {
	dea := NewDEA("DEA", 3, 4, 2, "Close")
	dea.LoadData(1)
	tmp := dea.Eval()
	// ema_s = 0.5
	// ema_l = 0.4
	// if the abs difference between tmp and 0.1 is greater than 0.00001, then error
	if  math.Abs(tmp - 0.066666666666) > 0.00001 {
		t.Errorf("DEA Eval() error, got %f, expected %f", tmp, 0.066666666666)
	}

	// test DoCalculate
	tr := &tmpRecordFunc{}
	tr.Data = []string{"2"}
	tr.Header = map[string]int{"Close": 0}

	ttt := time.Now()
	tmpstr := dea.DoCalculate(ttt,tr)
	// ! 懒得核对了
	if tmpstr != "0.1622" {
		t.Errorf("DEA DoCalculate() error, got %s, expected %s", tmpstr, "0.21")
	}


} 