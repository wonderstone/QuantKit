package position

import (
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"
)

// NewPosition 持仓
func NewPosition(
	account handler.Account2,
	contract contract.Contract,
	qty,price float64,
	posDirection ...config.PositionDirection,
) handler.Position {
	accType := string(contract.GetAccountType())
	elem, ok := setting.TypeRegistry[accType]
	if !ok {
		config.ErrorF("未知的账户类型: " + accType)
	}

	position := reflect.New(elem).Interface().(handler.Position)

	if len(posDirection) == 0 {
		position.Init(account, contract, qty, price, config.PositionLong)
	} else {
		position.Init(account, contract, qty, price, posDirection[0])
	}

	return position
}

func init() {
	setting.Register(new(StockPosition), config.AccountTypeStockSimple)
}