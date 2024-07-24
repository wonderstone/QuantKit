package indicator

import (
	"testing"
	"time"
)

// func to test ma.go
func TestConst(t *testing.T) {


	cr1 := NewConst("test", 0.9)

	// test DoCalculate
	tr := &tmpRecordFunc{}
	tr.Data = []string{"4.0"}

	ttt := time.Now()
	tmpstr := cr1.DoCalculate(ttt, tr)
	if tmpstr != "0.9000" {
		t.Error("Getter.DoCalculate() error")
	}




}
