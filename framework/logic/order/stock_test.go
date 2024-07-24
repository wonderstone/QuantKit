package order

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	_ "github.com/wonderstone/QuantKit/framework/logic/contract"
	"github.com/wonderstone/QuantKit/framework/setting"
)

// 测试生成新的订单
func TestNewOrder(t *testing.T) {

	// id int64, contract contract.Contract, qty float64, op *order.OrderOp

	conf, err := config.NewContractPropertyConfig("./contract.yaml")
	require.NoError(t, err)
	fmt.Println(conf)
	wp := contract.WithProperty(conf)

	handle, _ := setting.NewContractHandler(config.HandlerTypeConfig, wp)
	contr := handle.GetContract("000001.XSHE.CS")

	oopts := []account.WithOrderOption{
		account.WithOrderTime(time.Now()),
		account.WithOrderPrice(10.0),
		account.WithOrderDirection(config.OrderBuy),
		account.WithOrderType(config.OrderTypeLimit),
		// ! 以下两项CheckPos和CheckCash在生成订单时并没有使用！！
		account.WithCheckPosition(true),
		account.WithCheckCash(true),
		// ! 完毕 ！
	}

	oop := account.NewOrderOp(oopts...)

	order := NewOrder(10001, contr, 100, oop)
	// 如果order不为空，说明生成成功
	require.NotNil(t, order)
}
