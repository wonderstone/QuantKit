// All rights reserved. This is part of West Securities ltd. proprietary source code.
// No part of this file may be reproduced or transmitted in any form or by any means,
// electronic or mechanical, including photocopying, recording, or by any information
// storage and retrieval system, without the prior written permission of West Securities ltd.

// author:  Wonderstone (Digital Office Product Department #2)
// revisor:

package indicator

import (
	"fmt"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/container/queue"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// math.nan

// Ref is the moving average indicator
type Ref struct {
	Name string
	Base string
	N    int64
	t    time.Time
	DQ   *queue.Queue[float64]
}

func (r *Ref) DoInit(f config.Formula) {
	r.Name = f.Name
	r.N = config.MustGetParamInt(f.Param, "N")
	r.Base = config.MustGetParamString(f.Param, "Base")
	r.DQ = queue.New[float64](int(r.N+1))
}

func (r *Ref) DoCalculate(tm time.Time, data dataframe.RecordFunc) string {
	if tm != r.t {
		r.t = tm
		// 原则上一般数据指标计算都有值。空值问题来自于ref指标。
		// 此处留下ref作为ref输入的可能性通畅吧
		v, err := dataframe.TryConvertToFloat(data, r.Base)
		if err != nil {
			return ""
		} else {
			r.LoadData(v)
		}
	}
	// 如果队列是满的，那么就返回队列的头部
	if r.DQ.Full() {
		return fmt.Sprintf("%.4f", r.Eval())
	} else {
		return ""
	}
}

func (m *Ref) DoReset() {
	m.DQ.Clear()
}

// LoadData 加载数据
func (m *Ref) LoadData(data float64) {
	m.DQ.EnqueueWithDequeue(data)
}

func (m *Ref) Eval() float64 {
	v, _ := m.DQ.Peek()
	return v
}

// NewRef returns a new Ref indicator
func NewRef(Name string, N int64, Base string) *Ref {
	return &Ref{
		Name: Name,
		Base: Base,
		N:    N,
		DQ:   queue.New[float64](int(N+1)),
	}
}

func init() {
	formula.RegisterNewFormula(new(Ref), "Ref")
}
