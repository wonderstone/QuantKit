package main

import (
	_ "Strategy/indicator"
	"Strategy/strategy"
	"west.vqt.com/vqt-model/framework"
	"west.vqt.com/vqt-model/framework/entity/handler"
)

func main() {
	framework.Run(
		framework.WithStrategyCreator(
			func() handler.Strategy {
				return &strategy.DMT{}
			},
		),
	)
}
