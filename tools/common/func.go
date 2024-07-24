package common

import (
	"fmt"
	"math"

	"github.com/wonderstone/QuantKit/config"
	math2 "github.com/wonderstone/QuantKit/tools/math"
)

// CalcCommStock 计算股票手续费，方向为买卖方向
func CalcCommStock(fee config.StockFee, qty float64, price float64, direction config.OrderDirection) float64 {
	// 过户费双向收取 历史上2015年 沪深交易所 万分之0.2 现在2022-04-29 沪深万分之0.1
	// 印花税  千分之一  卖方收取
	// 万分之一~万分之五 最低5元 包含交易所收取的经手费0.00487% 证监会最终收取的证管费0.002%(合计0.00687%)
	amt := qty * price
	commission := 0.0
	commission += math2.Round(fee.TransferFeeRate*amt, 2)
	if direction == config.OrderSell {
		commission += fee.TaxRate * amt
	}

	commission += math.Max(fee.CommBrokerRate*amt, fee.MinFees)

	// 保留两位小数
	commission = math.Floor(commission*100) / 100

	return commission
}

// CalcCommFuture 计算期货手续费，方向为买卖方向
func CalcCommFuture(fee config.FutureFee, qty float64, price float64, direction config.OrderDirection) float64 {
	commission := 0.0

	// TODO 期货手续费计算

	return commission
}

// CalcMarketValueStock 计算股票市值，方向为多空方向
func CalcMarketValueStock(
	stock config.ContractBasic, qty float64, price float64, direction config.PositionDirection,
) float64 {
	return qty * price
}

// CalcMarketValueFuture 计算期货市值，方向为多空方向
func CalcMarketValueFuture(
	future config.ContractBasic, qty float64, price float64, direction config.PositionDirection,
) float64 {
	return qty * price * future.ContractSize
}

// CalcMarginFuture 计算期货保证金，方向为多空方向
func CalcMarginFuture(
	contractSize float64, marginRate config.MarginRate, qty float64, price float64, direction config.PositionDirection,
) float64 {
	if direction == config.PositionLong { // 多头
		return qty * price * contractSize * (marginRate.Long + marginRate.Broker)
	} else if direction == config.PositionShort { // 空头
		return qty * price * contractSize * (marginRate.Short + marginRate.Broker)
	}

	panic(fmt.Sprintf("未知的合约方向: %d", direction))
}

// CalcSlipPrice 计算滑点，方向为买卖方向
func CalcSlipPrice(basic config.ContractBasic, price float64, slip float64, direction config.OrderDirection) float64 {
	if direction == config.OrderBuy { // 买入
		return price + basic.TickSize*slip
	} else if direction == config.OrderSell { // 卖出
		return price - basic.TickSize*slip
	}

	panic(fmt.Sprintf("未知的合约买卖方向: %d", direction))
}
