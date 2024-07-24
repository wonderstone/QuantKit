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
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// Ref is the moving average indicator
type Refd struct {
	Name string
	t    time.Time
	Tag  string
	Ref  *Ref
}

func (r *Refd) DoInit(f config.Formula) {
	r.Name = f.Name
	r.Tag = config.MustGetParamString(f.Param, "Tag")
	if r.Tag != "W" && r.Tag != "D" {
		config.ErrorF("MA指标[%s]输入参数Tag[%s]不正确，Tag只能为 W 或 D ", r.Name, r.Tag)
	}


	r.Ref = NewRef(f.Name, config.MustGetParamInt(f.Param, "N"), config.MustGetParamString(f.Param, "Base"))
}

func (r *Refd) DoCalculate(tm time.Time, data dataframe.RecordFunc) string {
	// 获取tm的date  验证日线级时间标签更新
	_, _, tmday := tm.Date()
	_, _, rday := r.t.Date()

	_, tmweek := tm.ISOWeek()
	_, rweek := r.t.ISOWeek()

	if r.Tag == "W" && tmweek != rweek {
		r.t = tm
		return r.Ref.DoCalculate(tm, data)
	} 

	if r.Tag == "D" && tmday != rday {
		r.t = tm
		return r.Ref.DoCalculate(tm, data)
	}
	return fmt.Sprintf("%.4f", r.Eval())

}

func (r *Refd) DoReset() {
	r.Ref.DoReset()
}

// LoadData 加载数据
func (r *Refd) LoadData(data float64) {
	r.Ref.LoadData(data)
}

func (r *Refd) Eval() float64 {
	return r.Ref.Eval()
}

// NewRef returns a new Ref indicator
func NewRefd(Name string, N int, Base,Tag string) *Refd {
	return &Refd{
		Name: Name,
		Tag:  Tag,
		Ref:  NewRef(Name, int64(N), Base),
	}
}

func init() {
	formula.RegisterNewFormula(new(Refd), "Refd")
}
