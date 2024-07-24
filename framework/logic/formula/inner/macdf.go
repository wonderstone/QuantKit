// All rights reserved. This is part of West Securities ltd. proprietary source code.
// No part of this file may be reproduced or transmitted in any form or by any means,
// electronic or mechanical, including photocopying, recording, or by any information
// storage and retrieval system, without the prior written permission of West Securities ltd.

// author:  Zhangweixuan (Digital Office Product Department #2)
// revisor:
package indicator

import (
	"fmt"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type MACDF struct {
	Name    string
	Freq    int
	counter int
	MACD    *MACD
}

func (m *MACDF) DoInit(f config.Formula) {
	m.Name = f.Name
	m.Freq = int(config.MustGetParamInt(f.Param, "Freq"))
	// New MACD
	Base := config.MustGetParamString(f.Param, "Base")
	S := config.MustGetParamInt(f.Param, "S")
	L := config.MustGetParamInt(f.Param, "L")
	N := config.MustGetParamInt(f.Param, "N")
	m.MACD = NewMACD(m.Name, S, L, N, Base)
}

func (m *MACDF) DoCalculate(tm time.Time, data dataframe.RecordFunc) string {
	m.counter++
	if m.counter%m.Freq == 0 {
		m.counter = 0
		return m.MACD.DoCalculate(tm, data)
	}
	return fmt.Sprintf("%.4f", m.Eval())
}

func (m *MACDF) DoReset() {
	m.MACD.DoReset()
	m.counter = 0
}

// LoadData loads 1 tick info datas into the indicator
func (m *MACDF) LoadData(close float64) {
	m.MACD.LoadData(close)
}

// Eval evaluates the indicator
func (m *MACDF) Eval() float64 {
	return m.MACD.Eval()
}

func NewMACDF(Name string, S, L, N,Freq int64,  Base string) *MACDF {
	return &MACDF{
		Name: Name,
		Freq: int(Freq),
		MACD: NewMACD(Name, S, L, N, Base),
	}
}

func init() {
	formula.RegisterNewFormula(new(MACDF), "MACDF")
}
