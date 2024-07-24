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
	"github.com/wonderstone/QuantKit/tools/container/queue"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// MA is the moving average indicator
type MA struct {
	Name string
	N    int
	Base string
	sum  float64
	DQ   *queue.Queue[float64]
}

func (m *MA) DoInit(f config.Formula) {
	m.Name = f.Name
	if len(f.Input) != 1 {
		config.ErrorF("MA指标[%s]输入参数个数不正确，MA公式只允许一个输入值", m.Name)
	}
	// get f.input key and value
	for k, v := range f.Input {
		m.Base = k
		// convert string to int
		v, err := strconv.Atoi(v)
		if err != nil {
			config.ErrorF("MA指标[%s]输入参数[%s]不是整数", m.Name, k)
		}
		m.N = v
	}
	m.DQ = queue.New[float64](m.N)
}

func (m *MA) DoCalculate(tm time.Time, row dataframe.RecordFunc) string {
	// m.LoadData(dataframe.ConvertToFloat(row, m.Base))
	v, err := dataframe.TryConvertToFloat(row, m.Base)
	if err != nil {
		return ""
	} else {
		m.LoadData(v)
	}
	// 如果数据不够，不计算，只更新累计值
	if !m.DQ.Full() {
		return ""
	}
	return fmt.Sprintf("%.4f", m.Eval())
}

func (m *MA) DoReset() {
	m.sum = 0
	m.DQ.Clear()
}

// LoadData 加载数据
func (m *MA) LoadData(close float64) {
	m.sum += close
	lv, full := m.DQ.EnqueueWithDequeue(close)
	if full {
		m.sum -= lv
	}
	
}

func (m *MA) Eval() float64 {
	// if m.DQ.Len() == 0 {
	// 	return m.sum / 1.0
	// }
	return m.sum / float64(m.DQ.Len())
}

// NewMA returns a new MA indicator
func NewMA(Name string, N int, Base string) *MA {
	return &MA{
		Name: Name,
		N:    N,
		Base: Base,
		sum:  0,
		DQ:   queue.New[float64](N),
	}
}

func init() {
	formula.RegisterNewFormula(new(MA), "MA")
}
