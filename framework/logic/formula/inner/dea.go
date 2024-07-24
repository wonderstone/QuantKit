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

type DEA struct {
	Name       string
	Base       string
	S, L, N    int64
	Dif        *DIF
	EmaDif     *EMA
}

func (d *DEA) DoInit(f config.Formula) {
	d.Name = f.Name
	d.Base = config.MustGetParamString(f.Param, "Base")
	d.S = config.MustGetParamInt(f.Param, "S")
	d.L = config.MustGetParamInt(f.Param, "L")
	d.N = config.MustGetParamInt(f.Param, "N")
	d.Dif = NewDIF("DIF", d.S, d.L, d.Base)
	d.EmaDif = NewEMA("EMADIF", d.N, "DIF")
}


func (d *DEA) DoCalculate(tm time.Time, data dataframe.RecordFunc) string{
	// d.LoadData(dataframe.ConvertToFloat(data, d.Base))
	v, err := dataframe.TryConvertToFloat(data, d.Base)
	if err != nil {
		return ""
	}
	d.LoadData(v)
	return fmt.Sprintf("%.4f", d.Eval())
}

func (d *DEA) DoReset() {
	d.Dif.DoReset()
	d.EmaDif.DoReset()
}

// LoadData loads 1 tick info datas into the indicator
func (d *DEA) LoadData(close float64) {
	d.Dif.LoadData(close)
	d.EmaDif.LoadData(d.Dif.Eval())
}

// Eval evaluates the indicator
func (d *DEA) Eval() float64 {
	return d.EmaDif.Eval()
}

func NewDEA(Name string, S,L,N int64, Base string) *DEA {
	return &DEA{
		Name:        Name,
		S:           S,
		L:           L,
		N:           N,
		Base:        Base,
		Dif:         NewDIF("DIF", S, L, Base),
		EmaDif:      NewEMA("EMADIF", N, "DIF"),
	}
}

func init() {
	formula.RegisterNewFormula(new(DEA), "DEA")
}
