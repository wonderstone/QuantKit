package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)



type MatcherOp struct {
	StockSlippage  float64          // 滑点
	FutureSlippage float64          // 滑点
	config         config.Framework // 配置
}

type WithMatcherOption func(*MatcherOp)

func WithStockSlippage(slippage float64) WithMatcherOption {
	return func(op *MatcherOp) {
		op.StockSlippage = slippage
	}
}

func WithFutureSlippage(slippage float64) WithMatcherOption {
	return func(op *MatcherOp) {
		op.FutureSlippage = slippage
	}
}

func WithConfig(config config.Framework) WithMatcherOption {
	return func(op *MatcherOp) {
		op.config = config
	}
}

// + Matcher 接口动作太少 而且直接基于handler.Order操作 
// + 为避免循环导入只适合放在handler包下
// + 关键：撮合器虽然操作Order，但是Order会立刻调用handler级别接口调整account
// + 所以撮合器需要建立在handler之上
// + 如果建立在account之上，就非常违背直觉，而且如果依赖order索引到account就非常麻烦

type Matcher interface {
	Init(opt ...WithMatcherOption) Matcher
	MatchOrder(order Order, indicate dataframe.StreamingRecord, matchTime time.Time)
}

