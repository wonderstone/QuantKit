package account

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
)

// + Order 订单定义

type Order interface {
	// * fundamental info parts
	ID() int64                   // 订单ID
	InstID() string              // 合约ID
	Contract() contract.Contract // 合约
	// & 这个返回Account接口的操作大幅度的简化了后续的操作逻辑
	Account() Account
	IsExecuted() bool                            // 是否已成交
	OrderTime() time.Time                        // 下单时间
	OrderPrice() float64                         // 下单价格
	OrderQty() float64                           // 下单数量
	OrderAmt() float64                           // 下单金额
	OrderDirection() config.OrderDirection       // 下单方向: 买、卖
	PositionDirection() config.PositionDirection // 仓位方向: 多、空、总体(期货有用)
	TransactionType() config.TransactionType     // 交易类型开平标志：开仓、平今、平昨(期货有用)
	OrderType() config.OrderType                 // 订单委托类型：限价单(limit)、市价单(market)
	OrderStatus() config.OrderStatus             // 订单状态
	Margin() float64                             // 保证金
	MarketValue() float64                        // 市值
	TradePrice() float64                         // 成交价格
	TradeQty() float64                           // 成交数量
	TradeAmt() float64                           // 成交金额
	TradeTime() time.Time                        // 成交时间
	Commission() float64                         // 手续费
	CommissionFrozen() float64                   // ? 冻结手续费
}

