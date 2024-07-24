package account

import (
	"time"
)

// - Fundamental concept
// + 它的实现是StockSimple结构体
type Account interface {
	GetType() string // GetType 获取账户信息, A股stock，期货future
	AccountID() string
	GetAsset() Asset                                    // 获取账户资产信息
	GetPnL() PnL                                        // 获取账户盈亏信息
	GetPosition() map[string]Position                   // 获取账户持仓信息
	GetPositionByInstID(instID string) (Position, bool) // 通过instID获取持仓信息
	GetPosInstIDs() []string                            // 获取账户持仓股票信息
	GetOrders(instID string) []Order                    // 获取账户订单信息
	NewOrder(instId string, qty float64, option ...WithOrderOption) (Order, error)
	InsertOrder(order Order, option ...WithOrderOption) error
	CancelOrder(tm time.Time, id int64)
	// @ 实盘下单需要，可以返回空，但功能完备就应该需要
	CashAccount() string // CashAccount 获取资金账户ID
}



