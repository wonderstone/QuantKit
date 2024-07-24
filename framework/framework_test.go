package framework

import "testing"

// func TestTrain(t *testing.T) {
// 	Run(
// 		WithMode("train"),

// 		WithWorkRoot("/Users/alexxiong/GolandProjects/quant"),
// 		WithStrategyPrefix("strategy"),
// 		WithStrategyID("S20230824-090000-000"),
// 		WithRunID("T20230824-110000-000"),

// 		WithStrategyDemo("T0"),
// 		WithRecalcIndicator(),

// 		WithPathStyle(PathStyleVMT),

// 		// WithDataPrefix(),
// 		// WithDownloadPrefix(),
// 		// WithCalcPrefix(),
// 		// WithModePrefix(),
// 		// WithInputPrefix(),
// 		// WithOutPrefix(),
// 		// WithCommonPrefix(),
// 	)
// }

// func TestBacktest(t *testing.T) {
// 	Run(
// 		WithMode("bt"),

// 		WithWorkRoot("/Users/alexxiong/GolandProjects/quant"),
// 		WithStrategyID("S20230824-090000-000"),
// 		WithRunID("B20230824-110000-000"),

// 		WithStrategyDemo("T0"),
// 		WithRecalcIndicator(),

// 		WithPathStyle(PathStyleVMT),
// 	)
// }

func TestCalc(t *testing.T) {
	Run(
		WithMode("calc"),

		WithWorkRoot("/Users/alexxiong/GolandProjects/quant"),
		WithStrategyPrefix("strategy"),
		WithStrategyID("S20230824-090000-000"),

		// WithFrequency("1min"),
		WithInstID([]string{"510050.XSHG.CS"}),

		WithPathStyle(PathStyleVMT),
	)
}
