package setting

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

// & Order 没有采用注册模式！！！
// & NewOrder 创建新订单动作在logic中完成了！

func MakeOrderRecord(
	ID int64, t time.Time, instId string,
	orderDirection config.OrderDirection,
	positionDirection config.PositionDirection,
	TransactionType config.TransactionType,
	orderPrice float64, orderQty float64,
	tradePrice float64, tradeQty float64,
	commission float64, status config.OrderStatus, reason error,
) *recorder.OrderRecord {
	return &recorder.OrderRecord{
		OrderId:           ID,
		OrderDate:         t.Format("2006-01-02"),
		OrderTime:         t.Format("15:04:05"),
		InstId:            instId,
		OrderDirection:    string(orderDirection),
		PositionDirection: string(positionDirection),
		TransactionType:   string(TransactionType),

		OrderPrice: orderPrice,
		OrderQty:   orderQty,
		TradePrice: tradePrice,
		TradeQty:   tradeQty,

		Commission: commission,
		Status:     status,
		RejectReason: func() string {
			if reason != nil {
				return reason.Error()
			}
			return ""
		}(),
	}
}
