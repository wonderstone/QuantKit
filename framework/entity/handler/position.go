package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

type Position interface {
	account.Position

	Init(account Account2, contract contract.Contract, qty, price float64, direction ...config.PositionDirection)

	DoOrderUpdate(order Order)

	DoTradeUpdate(qty, price float64, order Order) (deltaActualPnL float64, deltaFloatingPnL float64)

	DoSettle(tm time.Time, price float64, recorder recorder.Handler) bool

	DoResume(acc Account2, contract contract.Contract, record recorder.PositionRecord)

	CalcPnL(tm time.Time, price float64)

	AddSettle(settleQty, settlePrice float64)
}

