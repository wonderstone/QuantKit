// test dif

package indicator

import (
	"math"
	"testing"
	"time"
)

func TestDIF(t *testing.T) {
	dif := NewDIF("DIF", 3, 4, "Close")
	dif.LoadData(1)
	tmp := dif.Eval()
	// ema_s = 0.5
	// ema_l = 0.4
	// if the abs difference between tmp and 0.1 is greater than 0.00001, then error
	if  math.Abs(tmp - 0.1) > 0.00001 {
		t.Errorf("DIF Eval() error, got %f, expected %f", tmp, 0.1)
	}

	// test DoCalculate
	tr := &tmpRecordFunc{}
	tr.Data = []string{"2"}
	tr.Header = map[string]int{"Close": 0}

	ttt := time.Now()
	tmpstr := dif.DoCalculate(ttt,tr)
	// ema_s = 1.25
	// ema_l = 1.04
	if tmpstr != "0.2100" {
		t.Errorf("DIF DoCalculate() error, got %s, expected %s", tmpstr, "0.21")
	}
	
	






}