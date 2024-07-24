package account

import (
	"math"
	"time"

	"github.com/wonderstone/QuantKit/config"
)

// & OrderOp 订单操作
type OrderOp struct {
	OrderTime      *time.Time            // 订单时间
	OrderPrice     float64               // 订单价格
	OrderDirection config.OrderDirection // 订单方向
	OrderType      config.OrderType      // 订单类型
	Account        Account               // 账户

	// ! 以下两项CheckPos和CheckCash在生成订单时并没有使用！！
	CheckPos       bool                  // 是否检查持仓
	CheckCash      bool                  // 是否检查资金
	// ! 完毕 ！
	// ! 以下项Resumed在生成订单时并没有使用！！甚至没有WithResumed函数！！
	Resumed        bool                  // 是否恢复订单
	// ! 完毕 ！
}

type WithOrderOption func(opts *OrderOp)
type WithOrderFilter func(order Order) bool

func NewOrderOp(opts ...WithOrderOption) *OrderOp {
	op := &OrderOp{
		OrderDirection: config.OrderBuy,
		OrderType:      config.OrderTypeLimit,
	}
	for _, o := range opts {
		o(op)
	}
	return op
}

func WithOrderTime(orderTime time.Time) WithOrderOption {
	return func(opts *OrderOp) {
		opts.OrderTime = &orderTime
	}
}

func WithOrderPrice(orderPrice float64) WithOrderOption {
	return func(opts *OrderOp) {
		if math.IsNaN(orderPrice) || math.IsInf(orderPrice, 0) {
			config.ErrorF("价格不合法，为: %f", orderPrice)
			// panic("价格不合法")
		}
		opts.OrderPrice = orderPrice
	}
}
func WithOrderDirection(orderDirection config.OrderDirection) WithOrderOption {
	return func(opts *OrderOp) {
		opts.OrderDirection = orderDirection
	}
}

func WithOrderType(orderType config.OrderType) WithOrderOption {
	return func(opts *OrderOp) {
		opts.OrderType = orderType
	}
}

func WithAccount(account Account) WithOrderOption {
	return func(opts *OrderOp) {
		opts.Account = account
	}
}


// ! CheckPos在生成股票订单时并没有使用！！
func WithCheckPosition(flag bool) WithOrderOption {
	return func(opts *OrderOp) {
		opts.CheckPos = flag
	}
}

// ! CheckCash在生成股票订单时并没有使用！！
func WithCheckCash(flag bool) WithOrderOption {
	return func(opts *OrderOp) {
		opts.CheckCash = flag
	}
}


// & OpFilterPos position操作
// OpFilterPos 持仓过滤器
type OpFilterPos struct {
	Direction     config.PositionDirection // 持仓方向
	SellAvailable bool                     // 是否筛选出可卖
}

type WithOpFilterPos func(*OpFilterPos)

func WithDirection(direction config.PositionDirection) WithOpFilterPos {
	return func(op *OpFilterPos) {
		op.Direction = direction
	}
}

func WithSellAvailable(sellAvailable bool) WithOpFilterPos {
	return func(op *OpFilterPos) {
		op.SellAvailable = sellAvailable
	}
}