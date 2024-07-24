package matcher

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type NextQuote struct {
	StockSlippage  float64 // 滑点
	FutureSlippage float64 // 滑点
}

func newNextOp(opts ...handler.WithMatcherOption) *handler.MatcherOp {
	op := &handler.MatcherOp{}

	for _, option := range opts {
		option(op)
	}

	return op
}

func (c *NextQuote) Init(opts ...handler.WithMatcherOption) handler.Matcher {
	op := newNextOp(opts...)
	return &NextQuote{
		StockSlippage:  op.StockSlippage,
		FutureSlippage: op.FutureSlippage,
	}
}

func (c *NextQuote) MatchOrder(order handler.Order, indicate dataframe.StreamingRecord, matchTime time.Time) {
	// 如果订单已经结束，则不再处理
	if order.IsExecuted() {
		return
	}

	matchPrice := indicate.ConvertToFloat("Open")
	// 目前是使用的是下一根K线的开盘价作为成交价
	switch order.OrderDirection() {
	case config.OrderBuy:
		// 判断下单价格是否大于等于成交价(暂时不用)
		// if order.OrderPrice() >= matchPrice {
		// 订单成交
		order.DoTradeUpdate(
			order.Contract().CalcSlipPrice(
				matchPrice,
				c.StockSlippage,
				order.OrderDirection()),
			order.OrderQty(), matchTime, nil,
		)
		// } else {
		// 	order.DoCancelUpdate(matchTime)
		// }
	case config.OrderSell:
		// 判断下单价格是否小于等于成交价(暂时不用)
		// if order.OrderPrice() <= matchPrice {
		// 订单成交
		order.DoTradeUpdate(
			order.Contract().CalcSlipPrice(
				matchPrice, 
				c.StockSlippage, 
				order.OrderDirection()),
			order.OrderQty(), matchTime, nil,
		)
		// } else {
		// 	order.DoCancelUpdate(matchTime)
		// }
	}
}

func init() {
	setting.RegisterMatcher(
		new(NextQuote), 
		config.HandlerTypeNextMatch)
}
