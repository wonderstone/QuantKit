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

type MACD struct {
	Name    string
	Base    string
	S, L, N int64
	Dea     *DEA
}


func (m *MACD) DoInit(f config.Formula) {
	m.Name = f.Name
	m.Base = config.MustGetParamString(f.Param, "Base")
	m.S = config.MustGetParamInt(f.Param, "S")
	m.L = config.MustGetParamInt(f.Param, "L")
	m.N = config.MustGetParamInt(f.Param, "N")
	m.Dea = NewDEA("DEA", m.S, m.L, m.N, m.Base)
}

func (m *MACD) DoCalculate(tm time.Time, data dataframe.RecordFunc) string {
	// m.LoadData(dataframe.ConvertToFloat(data, m.Base))
	v, err := dataframe.TryConvertToFloat(data, m.Base)
	if err != nil {
		return ""
	}
	m.LoadData(v)
	return fmt.Sprintf("%.4f", m.Eval())
}

func (m *MACD) DoReset() {
	m.Dea.DoReset()
}
// LoadData loads 1 tick info datas into the indicator
func (m *MACD) LoadData(close float64) {
	m.Dea.LoadData(close)
}

// Eval evaluates the indicator
func (m *MACD) Eval() float64 {
	return 2 * (m.Dea.Dif.Eval() - m.Dea.Eval())
}

func NewMACD(Name string, S,L,N int64, Base string) *MACD {
	return &MACD{
		Name:        Name,
		S:           S,
		L:           L,
		N:           N,
		Base:        Base,
		Dea:         NewDEA("DEA", S, L, N, Base),
	}
}

func init() {
	formula.RegisterNewFormula(new(MACD), "MACD")
}



