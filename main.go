package main

import (
	"github.com/wonderstone/QuantKit/framework"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/strategy"
	// "fmt"
	// "log"
	// "time"
	// "github.com/wonderstone/QuantKit/ctp"
	// "github.com/wonderstone/QuantKit/ctp/thost"
)


func main() {

	framework.Run(
		framework.WithStrategyCreator(
			func() handler.Strategy {
				// return &strategy.SortBuy{}
				// return &strategy.DMT{}
				// return &strategy.DMTGS{}
				// return &strategy.CQ{}
				// return &strategy.VS{}
				// return &strategy.VS02{}
				// return &strategy.VS03{}
				// return &strategy.VS04{}
				// return &strategy.VSO{}
				return &strategy.VSS{}
			},
		),
	)
}

