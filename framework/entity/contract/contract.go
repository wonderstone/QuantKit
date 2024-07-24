package contract

import (
	"github.com/wonderstone/QuantKit/config"
)

// + 合约概念细化到了一个标的一个合约，国内场景下：
// + 对期货来说, 一个品种 + 月份就是一个合约
// + 对股票来说, 一个标的 + 除复权信息就是一个合约

// - Fundamental concept
type Contract interface {
	// 获取合约代码
	GetInstID() string
	// 获取账户类型
	GetAccountType() config.AccountType
	// 计算全部手续费，方向为买卖方向
	CalcComm(qty float64, price float64, direction config.OrderDirection) float64
	// 计算市值，方向为多空方向
	CalcMarketValue(qty float64, price float64, direction config.PositionDirection) float64
	// 计算保证金，方向为多空方向
	CalcMargin(qty, price float64, direction config.PositionDirection) float64
	// 计算滑点，方向为买卖方向
	CalcSlipPrice(price float64, slippage float64, direction config.OrderDirection) float64
	// 计算最大可交易数量
	CalcMaxQty(qty float64) float64
}
