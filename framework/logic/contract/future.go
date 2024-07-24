package contract

import (
	"math"
	"time"

	"github.com/wonderstone/QuantKit/tools/common"
	"github.com/wonderstone/QuantKit/config"
)

type FutureContract struct {
	instId string // 合约代码
	*config.FutureContract
}

func (c FutureContract) CalcMaxQty(qty float64) float64 {
	if c.Basic.MinOrderVol > qty {
		return 0
	} else {
		return qty - math.Mod(qty-c.Basic.MinOrderVol, c.Basic.ContractSize)
	}
}

func (c FutureContract) CalcSettleInfo(
	tm time.Time, qty, price float64,
) (settleQty, settlePrice, dividend, tax float64) {
	// TODO 目前先不处理，返回原始值
	return qty, price, 0, 0
}

func (c FutureContract) GetAccountType() config.AccountType {
	return config.AccountTypeFuture
}

func (c FutureContract) GetInstID() string {
	return c.instId
}

func (c FutureContract) CalcComm(qty float64, price float64, direction config.OrderDirection) float64 {
	// TODO implement me
	panic("implement me")
}

func (c FutureContract) CalcMarketValue(qty float64, price float64, direction config.PositionDirection) float64 {
	return common.CalcMarketValueFuture(c.Basic, qty, price, direction)
}

func (c FutureContract) CalcMargin(qty float64, price float64, direction config.PositionDirection) float64 {
	return common.CalcMarginFuture(c.Basic.ContractSize, c.MarginRate, qty, price, direction)
}

func (c FutureContract) CalcSlipPrice(price float64, slippage float64, direction config.OrderDirection) float64 {
	return common.CalcSlipPrice(c.Basic, price, slippage, direction)
}
