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

type DIF struct {
	Name         string
	Base         string
	S, L         int64
	EMA_S, EMA_L *EMA
}

func (d *DIF) DoInit(f config.Formula) {
	d.Name = f.Name
	d.S = config.MustGetParamInt(f.Param, "S")
	d.L = config.MustGetParamInt(f.Param, "L")
	d.Base = config.MustGetParamString(f.Param, "Base")
	d.EMA_S = NewEMA("EMA_S", d.S, d.Base)
	d.EMA_L = NewEMA("EMA_L", d.L, d.Base)
}
func (d *DIF) DoCalculate(tm time.Time, data dataframe.RecordFunc) string{
	// d.LoadData(dataframe.ConvertToFloat(data, d.Base))
	v, err := dataframe.TryConvertToFloat(data, d.Base)
	if err != nil {
		return ""
	}
	d.LoadData(v)	
	return fmt.Sprintf("%.4f", d.Eval())
	
}

func (d *DIF) DoReset() {
	d.EMA_S.DoReset()
	d.EMA_L.DoReset()
}

// LoadData loads 1 tick info datas into the indicator
func (d *DIF) LoadData(close float64) {
	d.EMA_S.LoadData(close)
	d.EMA_L.LoadData(close)
}

// Eval evaluates the indicator
func (d *DIF) Eval() float64 {
	return d.EMA_S.Eval() - d.EMA_L.Eval()
}

func NewDIF(Name string, S ,L int64, Base string) *DIF {
	// 判断inputcounts是否2个数值  不是就报错
	if S == 0 || L == 0 || Base == "" {
		panic("DIF 参数错误")
	}
	return &DIF{
		Name:        Name,
		S:           S,
		L:           L,
		EMA_S:       NewEMA("EMA_S", S, Base),
		EMA_L:       NewEMA("EMA_L", L, Base),
	}
}

func init() {
	formula.RegisterNewFormula(new(DIF), "DIF")
}
