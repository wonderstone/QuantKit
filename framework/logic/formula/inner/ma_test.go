// All rights reserved. This is part of West Securities ltd. proprietary source code.
// No part of this file may be reproduced or transmitted in any form or by any means,
// electronic or mechanical, including photocopying, recording, or by any information
// storage and retrieval system, without the prior written permission of West Securities ltd.

// author:  Wonderstone (Digital Office Product Department #2)
// revisor:

package indicator

import (
	"fmt"
	"testing"
	"time"
)

// func to test ma.go
func TestMA(t *testing.T) {
	ma := NewMA("MA", 3, "Close")
	// new calculator

	fmt.Println(ma)
	// test LoadData
	ma.LoadData(1.0)
	ma.LoadData(2.0)
	ma.LoadData(3.0)
	// test eval
	tmp :=ma.Eval()
	if tmp != 2.0 {
		t.Error("MA.Eval() error")
	}
	// do it again should be the same
	tmp = ma.Eval()
	if tmp != 2.0 {
		t.Error("MA.Eval() error")
	}

	// test DoCalculate
	tr:= &tmpRecordFunc{}
	tr.Data = []string{"4.0"}
	tr.Header = map[string]int{"Close": 0}

	ttt := time.Now()
	tmpstr:= ma.DoCalculate(ttt, tr)
	if tmpstr != "3.0000" {
		t.Error("MA.DoCalculate() error")
	}

	ttt = time.Now()
	tr.Data = []string{""}
	tr.Header = map[string]int{"Close": 0}
	tmpstr = ma.DoCalculate(ttt, tr)
	if tmpstr != "" {
		t.Error("MA.DoCalculate() error")
	}

}