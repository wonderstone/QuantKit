package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

// & Order 没有使用注册模式，订单直接NewOrder。没必要初始化Init

type Order interface {
	// * fundamental info parts
	account.Order
	// * action parts
	// DoTradeUpdate 交易更新
	DoTradeUpdate(tradePrice float64, tradeQty float64, tradeTime time.Time, recorder recorder.Handler)
	// DoResume 恢复订单：出现异常情况时，恢复订单。严苛回测场景(立即市价成交)不需要
	DoResume(record *recorder.OrderRecord)
	// DoSettle 闭市结算更新。严苛回测场景(立即市价成交)不需要
	DoSettle(time time.Time, recorder recorder.Handler)
	// DoCancelUpdate 撤单更新。严苛回测场景(立即市价成交)不需要
	DoCancelUpdate(cancelTime time.Time)
	// DoReject 更新订单，拒单
	DoReject(rejectTime time.Time, err error)
	// DoRTInsert 实盘插入订单
	DoRTInsert() error
	// DoRTCancel 实盘取消订单
	DoRTCancel() error
	// ModifyOrder 修改订单
	ModifyOrder(tag string, value interface{})
}
