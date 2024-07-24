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

type MACDD struct {
	Name string
	t    time.Time
	Tag  string
	MACD *MACD
}

func (m *MACDD) DoInit(f config.Formula) {
	m.Name = f.Name
	m.Tag = config.MustGetParamString(f.Param, "Tag")
	if m.Tag != "W" && m.Tag != "D" {
		config.ErrorF("MACDD指标[%s]输入参数Tag[%s]不正确，Tag只能为 W 或 D ", m.Name, m.Tag)
	}
	// New MACD
	Base := config.MustGetParamString(f.Param, "Base")
	S := config.MustGetParamInt(f.Param, "S")
	L := config.MustGetParamInt(f.Param, "L")
	N := config.MustGetParamInt(f.Param, "N")
	m.MACD = NewMACD(m.Name, S, L, N, Base)
}

func (m *MACDD) DoCalculate(tm time.Time, data dataframe.RecordFunc) string {
	_, _, tmday := tm.Date()
	_, _, mday := m.t.Date()

	_,tmweek := tm.ISOWeek()
	_,mweek := m.t.ISOWeek()

	if m.Tag == "W" && tmweek != mweek {
		m.t = tm
		return m.MACD.DoCalculate(tm, data)
	}


	if m.Tag == "D" && tmday != mday {
		m.t = tm
		return m.MACD.DoCalculate(tm, data)
	}

	return fmt.Sprintf("%.4f", m.MACD.Eval())

}

func (m *MACDD) DoReset() {
	m.MACD.DoReset()
	m.t = time.Time{}
}

// LoadData loads 1 tick info datas into the indicator
func (m *MACDD) LoadData(close float64) {
	m.MACD.LoadData(close)
}

// Eval evaluates the indicator
func (m *MACDD) Eval() float64 {
	return m.MACD.Eval()
}

func NewMACDD(Name string, S, L, N int64, Base, Tag string) *MACDD {
	return &MACDD{
		Name: Name,
		Tag:  Tag,
		MACD: NewMACD(Name, S, L, N, Base),
	}
}

func init() {
	formula.RegisterNewFormula(new(MACDD), "MACDD")
}
