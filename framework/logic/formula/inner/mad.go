// All rights reserved. This is part of West Securities ltd. proprietary source code.
// No part of this file may be reproduced or transmitted in any form or by any means,
// electronic or mechanical, including photocopying, recording, or by any information
// storage and retrieval system, without the prior written permission of West Securities ltd.

// author:  Wonderstone (Digital Office Product Department #2)
// revisor:

package indicator

import (
	"fmt"
	"strconv"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// MA is the moving average indicator
type MAD struct {
	Ma   *MA
	Name string
	Tag  string

	t time.Time
}

func (m *MAD) DoInit(f config.Formula) {
	m.Name = f.Name
	Base := ""
	N := 0
	m.Tag = config.MustGetParamString(f.Param, "Tag")

	// if Tag is not "W" or "D", return error
	if m.Tag != "W" && m.Tag != "D" {
		config.ErrorF("MA指标[%s]输入参数Tag[%s]不正确，Tag只能为 W 或 D ", m.Name, m.Tag)
	}

	if len(f.Input) != 1 {
		config.ErrorF("MA指标[%s]输入参数个数不正确，MA公式只允许一个输入值", m.Name)
	}
	// get f.input key and value
	for k, v := range f.Input {
		Base = k
		// convert string to int
		v, err := strconv.Atoi(v)
		if err != nil {
			config.ErrorF("MA指标[%s]输入参数[%s]不是整数", m.Name, k)
		}
		N = v
	}
	m.Ma = NewMA(m.Name, N, Base)
}

func (m *MAD) DoCalculate(tm time.Time, row dataframe.RecordFunc) string {
	_, _, tmday := tm.Date()
	_, _, mday := m.t.Date()

	// get tm.Date week number
	_, tmweek := tm.ISOWeek()
	_, mweek := m.t.ISOWeek()

	if m.Tag == "W"  && tmweek != mweek {
		m.t = tm
		return m.Ma.DoCalculate(tm, row)
	}

	if m.Tag == "D" && tmday != mday {
		m.t = tm
		return m.Ma.DoCalculate(tm, row)
	}

	if !m.Ma.DQ.Full() {
		return ""
	}
	return fmt.Sprintf("%.4f", m.Eval())
}

func (m *MAD) DoReset() {
	m.Ma.DoReset()
	m.t = time.Time{}
}

// LoadData 加载数据
func (m *MAD) LoadData(close float64) {
	m.Ma.LoadData(close)
}

func (m *MAD) Eval() float64 {
	return m.Ma.Eval()
}

// NewMA returns a new MA indicator
func NewMAD(Name string, N int, Base, Tag string) *MAD {
	return &MAD{
		Name: Name,
		Tag:  Tag,
		Ma:   NewMA(Name, int(N), Base),
	}
}

func init() {
	formula.RegisterNewFormula(new(MAD), "MAD")
}
