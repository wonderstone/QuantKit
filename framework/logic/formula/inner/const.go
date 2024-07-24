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

// MA is the moving average indicator
type Const struct {
	Name string
	Num  float64
}

func (c *Const) DoInit(f config.Formula) {
	c.Name = f.Name
	c.Num = config.MustGetParamFloat64(f.Param, "Num")
}

func (c *Const) DoCalculate(tm time.Time, row dataframe.RecordFunc) string {

	return fmt.Sprintf("%.4f", c.Num)
}

func (c *Const) DoReset() {

}

// NewMA returns a new MA indicator
func NewConst(Name string, Num float64) *Const {
	return &Const{
		Name: Name,
		Num:  Num,
	}
}

func init() {
	formula.RegisterNewFormula(new(Const), "Const")
}
