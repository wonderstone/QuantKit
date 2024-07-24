package order

import (
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
)

func NewOrder(id int64, contract contract.Contract, qty float64, op *account.OrderOp) handler.Order {
	switch contract.GetAccountType() {
	case config.AccountTypeStockSimple:
		newOrder := StockOrder{
			id:               id,
			contract:         contract,
			orderTime:        *op.OrderTime,
			orderPrice:       op.OrderPrice,
			orderType:        op.OrderType,
			orderQty:         qty,
			orderDirection:   op.OrderDirection,
			account:          op.Account.(handler.Account2),

			orderStatus:      config.OrderStatusNew,
			frozenCommission: contract.CalcComm(qty, op.OrderPrice, op.OrderDirection),
		}

		return &newOrder
		// 	! 期货未实现
		// case config.AccountTypeFuture:
		// 	newOrder := FutureOrder{
		// 		// TODO
		// 	}
		// 	return &newOrder
	}

	return nil
}

func ResumeOrder(orderId int64, contract contract.Contract, qty float64, op *account.OrderOp) handler.Order {
	switch contract.GetAccountType() {
	case config.AccountTypeStockSimple:
		newOrder := StockOrder{
			id:               orderId,
			contract:         contract,
			orderTime:        *op.OrderTime,
			orderPrice:       op.OrderPrice,
			orderType:        op.OrderType,
			orderQty:         qty,
			orderDirection:   op.OrderDirection,
			orderStatus:      config.OrderStatusNew,
			frozenCommission: contract.CalcComm(qty, op.OrderPrice, op.OrderDirection),
		}

		return &newOrder
		// 	! 期货未实现
		// case config.AccountTypeFuture:
		// 	newOrder := FutureOrder{
		// 		// TODO
		// 	}
		// 	return &newOrder
	}

	return nil
}
