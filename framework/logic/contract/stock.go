package contract

import (
	"math"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/common"
	"github.com/wonderstone/QuantKit/tools/container/btree"
)

type StockContract struct {
	instId string // 合约代码
	*config.StockContract

	btree.MapG[time.Time, *config.Xrxd] // 除权除息数据, tm(除权登记日) -> xrxd
}

// CalcMaxQty 计算最大可交易数量
func (c StockContract) CalcMaxQty(qty float64) float64 {
	if c.Basic.MinOrderVol > qty {
		return 0
	} else {
		return qty - math.Mod(qty-c.Basic.MinOrderVol, c.Basic.ContractSize)
	}
}

func (c StockContract) GetAccountType() config.AccountType {
	return config.AccountTypeStockSimple
}

func (c StockContract) GetInstID() string {
	return c.instId
}

func (c StockContract) CalcComm(qty float64, price float64, direction config.OrderDirection) float64 {
	return common.CalcCommStock(c.StockFee, qty, price, direction)
}

func (c StockContract) CalcMarketValue(qty float64, price float64, direction config.PositionDirection) float64 {
	return common.CalcMarketValueStock(c.Basic, qty, price, direction)
}

func (c StockContract) CalcMargin(qty float64, price float64, direction config.PositionDirection) float64 {
	return common.CalcMarketValueStock(c.Basic, qty, price, direction)
}

func (c StockContract) CalcSlipPrice(price float64, slippage float64, direction config.OrderDirection) float64 {
	return common.CalcSlipPrice(c.Basic, price, slippage, direction)
}
