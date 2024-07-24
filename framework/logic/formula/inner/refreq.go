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

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// Ref is the moving average indicator
type Refreq struct {
	Name       string
	Freq 	   int64
	counter    int64
	Ref        *Ref
}

func (r *Refreq) DoInit(f config.Formula) {
	r.Name = f.Name
	r.Freq = config.MustGetParamInt(f.Param, "Freq")
	r.Ref = NewRef(f.Name, config.MustGetParamInt(f.Param, "N"), config.MustGetParamString(f.Param, "Base"))
}

func (r *Refreq) DoCalculate(tm time.Time, data dataframe.RecordFunc) string {
	r.counter++
	if r.counter%r.Freq == 0{
		r.counter = 0
		return r.Ref.DoCalculate(tm, data)
	}
	return fmt.Sprintf("%.4f", r.Eval())
}

func (r *Refreq) DoReset() {
	r.Ref.DoReset()
}

// LoadData 加载数据
func (r *Refreq) LoadData(data float64) {
	r.Ref.LoadData(data)
}

func (m *Refreq) Eval() float64 {
	return m.Ref.Eval()
}

// NewRef returns a new Ref indicator
func NewRefreq(Name string, Freq int64, N int64, Base string) *Refreq {
	return &Refreq{
		Name: Name,
		Freq: Freq,
		Ref: NewRef(Name, N, Base),
	}
}

func init() {
	formula.RegisterNewFormula(new(Refreq), "Refreq")
}
