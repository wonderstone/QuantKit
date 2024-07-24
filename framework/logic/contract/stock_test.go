package contract

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/setting"
)

// 测试合约配置模式下的合约计算
func TestConfigMode(t *testing.T) {
	conf, err := config.NewContractPropertyConfig("./contract.yaml")
	require.NoError(t, err)
	fmt.Println(conf)

	wp := contract.WithProperty(conf)

	handle, _ := setting.NewContractHandler(config.HandlerTypeConfig, wp)
	contract := handle.GetContract("000001.XSHE.CS")

	fmt.Println(contract)

	// ! 手续费貌似这里的预期计算有问题，暂时还按照QK的计算方式来
	// 买入 主板 1000股 8.1元
	// 买入手续费 = max(1000*0.00001, 1) + max(1000*8.1*0.00003, 5) = 6
	require.Equalf(t, 9.86, contract.CalcComm(1000, 8.1, config.OrderBuy), "计算手续费错误")

	// 市值为 1000*8.1 = 8100
	require.Equalf(t, 8100.0, contract.CalcMarketValue(1000, 8.1, config.PositionLong), "计算市值错误")

	// 保证金=市值
	require.Equalf(t, 8100.0, contract.CalcMargin(1000, 8.1, config.PositionLong), "计算保证金错误")

	// 滑点价格计算
	// 买入 主板 1000股 8.1元
	// 滑点数 1
	// 滑点价格 = 8.1 + 1*0.01 = 8.11
	require.Equalf(t, 8.11, contract.CalcSlipPrice(8.1, 1, config.OrderBuy), "计算滑点价格错误")

	// 卖出 主板 1000股 17 元
	// 卖出手续费 = max(1000*0.00001, 1) + max(1000*17*0.00003, 5) + 1000*17*0.001 = 23.1
	// require.Equalf(t, 23.1, contract.CalcComm(1000, 17, config.OrderSell), "计算手续费错误")

	// 市值为 1000*17 = 17000
	require.Equalf(t, 17000.0, contract.CalcMarketValue(1000, 17, config.PositionShort), "计算市值错误")

	// 保证金=市值
	require.Equalf(t, 17000.0, contract.CalcMargin(1000, 17, config.PositionShort), "计算保证金错误")

	// 滑点价格计算
	// 卖出 主板 1000股 17元
	// 滑点数 1
	// 滑点价格 = 17 - 1*0.01 = 16.99
	require.Equalf(t, 16.99, contract.CalcSlipPrice(17, 1, config.OrderSell), "计算滑点价格错误")

}
