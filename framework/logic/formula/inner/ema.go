// All rights reserved. This is part of West Securities ltd. proprietary source code.
// No part of this file may be reproduced or transmitted in any form or by any means,
// electronic or mechanical, including photocopying, recording, or by any information
// storage and retrieval system, without the prior written permission of West Securities ltd.

// author:  Maminghui (Digital Office Product Department #2)
// revisor: Wonderstone (Digital Office Product Department #2)

package indicator

import (
	"fmt"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// EMA is the moving average indicator
type EMA struct {
	Name       string
	Base       string
	x, ly, tmp float64
	N          int64
}

func (e *EMA) DoInit(f config.Formula) {
	e.Name = f.Name
	e.Base = config.MustGetParamString(f.Param, "Base")
	e.N = config.MustGetParamInt(f.Param, "N")
}
func (e *EMA) DoCalculate(tm time.Time, data dataframe.RecordFunc) string {
	v, err := dataframe.TryConvertToFloat(data, e.Base)
	if err != nil {
		return ""
	}
	e.LoadData(v)
	return fmt.Sprintf("%.4f", e.Eval())
}

func (e *EMA) DoReset() {
	e.x = 0
	e.ly = 0
	e.tmp = 0
}

// LoadData loads 1 tick info datas into the indicator
func (e *EMA) LoadData(close float64) {
	e.x = close
	e.ly = e.tmp
}

// Eval evaluates the indicator
func (e *EMA) Eval() float64 {
	// !!! EMA = (2 * close + (N - 1) * EMA') / (N + 1)
	e.tmp = (2*e.x + (float64(e.N-1) * e.ly)) / float64(e.N+1)
	return e.tmp
}

func init() {
	formula.RegisterNewFormula(new(EMA), "EMA")
}

func NewEMA(Name string, N int64, Base string) *EMA {
	return &EMA{
		Name: Name,
		x:    0,
		ly:   0,
		Base: Base,
		N:    N,
	}
}
