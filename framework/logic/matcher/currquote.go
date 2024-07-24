package matcher

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type CurrQuote struct {
	StockSlippage  float64 // 滑点
	FutureSlippage float64 // 滑点
}

func newCurrOp() *handler.MatcherOp {
	op := &handler.MatcherOp{}

	return op
}

func (c *CurrQuote) Init(opts ...handler.WithMatcherOption) handler.Matcher {
	op := newCurrOp()
	return &CurrQuote{
		StockSlippage:  op.StockSlippage,
		FutureSlippage: op.FutureSlippage,
	}
}

func (c *CurrQuote) MatchOrder(order handler.Order, indicate dataframe.StreamingRecord, matchTime time.Time) {
	// 如果订单已经结束，则不再处理
	if order.IsExecuted() {
		return
	}
	matchPrice := indicate.ConvertToFloat("Close")
	// 目前是使用的是本根K线的收盘作为成交价
	switch order.OrderDirection() {
	case config.OrderBuy:
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
		new(CurrQuote),
		config.HandlerTypeCurrMatch,
	)
}
