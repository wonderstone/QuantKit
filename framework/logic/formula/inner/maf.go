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
type MAF struct {
	Ma      *MA
	Name   string
	Freq    int
	counter int
}

func (m *MAF) DoInit(f config.Formula) {
	m.Name = f.Name
	Base := ""
	N := 0

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
	m.Freq = int(config.MustGetParamInt(f.Param, "Freq"))
}

func (m *MAF) DoCalculate(tm time.Time, row dataframe.RecordFunc) string {
	m.counter++
	if m.counter%m.Freq == 0 {
		m.counter = 0
		// 加载数据
		return m.Ma.DoCalculate(tm, row)
	}
	if !m.Ma.DQ.Full() {
		return ""
	}
	return fmt.Sprintf("%.4f", m.Eval())
}

func (m *MAF) DoReset() {
	m.Ma.DoReset()
	m.counter = 0
}

// LoadData 加载数据
func (m *MAF) LoadData(close float64) {
	m.Ma.LoadData(close)
}

func (m *MAF) Eval() float64 {
	return m.Ma.Eval()
}

// NewMA returns a new MA indicator
func NewMAF(Name string, N int, Base string, Freq int) *MAF {
	return &MAF{
		Name: Name,
		Ma:   NewMA(Name, int(N), Base),
		Freq: Freq,
	}
}

func init() {
	formula.RegisterNewFormula(new(MAF), "MAF")
}
