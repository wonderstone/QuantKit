package order

import (
	"strconv"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

// + StockOrder股票订单结构体：
// ++ 框架需要：
// + 实现了entity/account.Order接口
// + 实现了entity/handler.Order接口
// ++ 下单需要：
// + 实现了tunnel.Order接口
// StockOrder 普通股票的订单

type StockOrder struct {
	id             int64                 // 订单ID
	orderTime      time.Time             // 订单时间
	orderPrice     float64               // 订单价格
	orderQty       float64               // 订单数量
	orderDirection config.OrderDirection // 订单方向
	orderType      config.OrderType      // 订单类型
	orderStatus    config.OrderStatus    // 订单状态

	sysOrderId string // 系统订单ID
	account    handler.Account2 // 账户

	tradePrice     float64           // 成交价格
	tradeQty       float64           // 成交数量
	prevTradeQty   float64           // 上一笔成交数量
	prevTradePrice float64           // 上一笔成交价格
	tradeTime      time.Time         // 成交时间
	contract       contract.Contract // 合约

	frozenCommission float64 // 冻结手续费
	commission       float64 // 手续费

	reject error // 拒单原因
}

// & StockOrder 实现了 Order 全部接口
// * fundamental info parts
// ID() int64
func (s *StockOrder) ID() int64 {
	return s.id
}

// Contract() contract.Contract
func (s *StockOrder) Contract() contract.Contract {
	return s.contract
}

func (s *StockOrder) Account() account.Account {
	return s.account.(account.Account)
}


// // 是否是已执行
// IsExecuted() bool
func (s *StockOrder) IsExecuted() bool {
	return s.orderStatus == config.OrderStatusDone || s.orderStatus == config.OrderStatusPartDonePartCancel || s.orderStatus == config.OrderStatusCanceled
}

// OrderTime() time.Time
func (s *StockOrder) OrderTime() time.Time {
	return s.orderTime
}

// OrderPrice() float64
func (s *StockOrder) OrderPrice() float64 {
	return s.orderPrice
}

// OrderQty() float64
func (s *StockOrder) OrderQty() float64 {
	return s.orderQty
}

// OrderAmt() float64
func (s *StockOrder) OrderAmt() float64 {
	return s.orderPrice * s.orderQty
}

// OrderDirection() config.OrderDirection
func (s *StockOrder) OrderDirection() config.OrderDirection {
	return s.orderDirection
}

// PositionDirection() config.PositionDirection
func (s *StockOrder) PositionDirection() config.PositionDirection {
	return config.PositionLong
}

// TransactionType() config.TransactionType
func (s *StockOrder) TransactionType() config.TransactionType {
	if s.orderDirection == config.OrderBuy {
		return config.OffsetOpen
	}
	return config.OffsetClose
}

// OrderType() config.OrderType
func (s *StockOrder) OrderType() config.OrderType {
	return s.orderType
}

// OrderStatus() config.OrderStatus
func (s *StockOrder) OrderStatus() config.OrderStatus {
	return s.orderStatus
}

// Margin() float64
func (s *StockOrder) Margin() float64 {
	return s.contract.CalcMargin(s.orderQty, s.orderPrice, config.PositionLong)
}

// MarketValue() float64
func (s *StockOrder) MarketValue() float64 {
	return s.contract.CalcMarketValue(s.orderQty, s.orderPrice, config.PositionLong)
}

// TradePrice() float64
func (s *StockOrder) TradePrice() float64 {
	return s.tradePrice
}

// TradeQty() float64
func (s *StockOrder) TradeQty() float64 {
	return s.tradeQty
}

// TradeAmt() float64
func (s *StockOrder) TradeAmt() float64 {
	return s.tradePrice * s.tradeQty
}

// TradeTime() time.Time
func (s *StockOrder) TradeTime() time.Time {
	return s.tradeTime
}

// InstID() string
func (s *StockOrder) InstID() string {
	return s.contract.GetInstID()
}

// Commission() float64
func (s *StockOrder) Commission() float64 {
	return s.commission
}

// CommissionFrozen() float64
func (s *StockOrder) CommissionFrozen() float64 {
	return s.frozenCommission
}
// * fundamental info parts end
// & StockOrder 实现了 account.Order 全部接口 end

// & StockOrder 实现了 handler.Order 超额接口
// * manage parts
// DoTradeUpdate 更新订单
func (s *StockOrder) DoTradeUpdate(
	tradePrice float64, tradeQty float64, tradeTime time.Time, recorder recorder.Handler,
) {
	// 更新订单: 订单存在部分成交的场景
	s.prevTradeQty = s.tradeQty
	s.prevTradePrice = s.tradePrice
	s.tradeQty += tradeQty
	s.tradePrice = (s.prevTradePrice*s.prevTradeQty + tradePrice*tradeQty) / s.tradeQty
	s.tradeTime = tradeTime

	// 更新订单状态
	if s.tradeQty == s.orderQty {
		s.orderStatus = config.OrderStatusDone
		s.commission = s.contract.CalcComm(s.tradeQty, s.orderPrice, s.orderDirection)
		// ! 请考虑禁止order层面更新账户，应该在account层面更新
		s.account.DoTradeUpdate(tradeQty, tradePrice, s)
		s.account.DoOrderUpdate(s)
	} else if s.tradeQty < s.orderQty {
		s.orderStatus = config.OrderStatusPartDone
		// ! 禁止order层面更新账户，应该在account层面更新
		s.account.DoTradeUpdate(tradeQty, tradePrice, s)
	}

	// 记录订单
	if recorder != nil {
		recorder.GetChannel() <- setting.MakeOrderRecord(
			s.id, tradeTime, s.contract.GetInstID(), 
			s.orderDirection, config.PositionLong, 
			config.OffsetOpen, s.orderPrice,
			s.orderQty, s.tradePrice, s.tradeQty,
			s.Commission(), s.orderStatus, s.reject,
		)
	}
}

// 撤单更新
// DoCancelUpdate(cancelTime time.Time)
func (s *StockOrder) DoCancelUpdate(cancelTime time.Time) {
	// 更新订单状态
	switch s.orderStatus {
	case config.OrderStatusNew:
		s.orderStatus = config.OrderStatusCanceled
		s.tradeTime = cancelTime
	case config.OrderStatusPartDone:
		s.orderStatus = config.OrderStatusPartDonePartCancel
		s.commission = s.contract.CalcComm(s.tradeQty, s.orderPrice, s.orderDirection)
		s.tradeTime = cancelTime
	}

	// ! 请考虑禁止order层面更新账户，应该在account层面更新
	s.account.DoOrderUpdate(s)
}

// 拒单更新
// DoReject(rejectTime time.Time, err error)
func (s *StockOrder) DoReject(rejectTime time.Time, err error) {
	// 更新订单状态
	s.orderStatus = config.OrderStatusRejected
	s.tradeTime = rejectTime
	s.reject = err

	// ! 禁止order层面更新账户，应该在account层面更新
	// s.account.DoOrderUpdate(s)
}

// DoSettle 闭市结算更新
// DoSettle(time time.Time, recorder recorder.Handler)
func (s *StockOrder) DoSettle(tm time.Time, recorder recorder.Handler) {
	// 将未完成订单根据情况更新为完成或者取消
	switch s.orderStatus {
	case config.OrderStatusNew:
		s.orderStatus = config.OrderStatusExpired
		s.tradeTime = tm
	case config.OrderStatusPartDone:
		s.orderStatus = config.OrderStatusPartDonePartCancel
		s.tradeTime = tm
		s.commission = s.contract.CalcComm(s.tradeQty, s.orderPrice, s.orderDirection)
	}

	// 记录订单
	if recorder != nil {
		recorder.GetChannel() <- setting.MakeOrderRecord(
			s.id, 
			s.orderTime, s.contract.GetInstID(), s.orderDirection, config.PositionLong, config.OffsetOpen, s.orderPrice,
			s.orderQty, s.tradePrice, s.tradeQty, s.commission, s.orderStatus, s.reject,
		)
	}
}

// 恢复订单：出现异常情况时，恢复订单
// DoResume(record *OrderRecord)
func (s *StockOrder) DoResume(record * recorder.OrderRecord) {
	s.DoTradeUpdate(record.TradePrice, record.TradeQty, s.orderTime, nil)
}


func (s *StockOrder) DoRTInsert() error {
	orderID, err := s.account.DoRTInsert(s)
	if err != nil {
		return err
	}

	s.sysOrderId = orderID
	return nil
}

func (s *StockOrder) DoRTCancel() error {
	return s.account.DoRTCancel(s.sysOrderId)
}
// ModifyOrder
func (s *StockOrder) ModifyOrder(tag string, value interface{}) {
	switch tag {
	case "orderStatus":
		s.orderStatus = value.(config.OrderStatus)
	case "sysOrderId":
		s.sysOrderId = value.(string)
	}
}




// & StockOrder 实现了 Order 全部接口 完毕


// & StockOrder 实现了 tunnel.Order 接口

func (s *StockOrder) RequestID() string {
	return s.account.GetRunId() + "-" + strconv.FormatInt(s.id, 10)
}

func (s *StockOrder) OrderID() string {
	return s.sysOrderId
}

func (s *StockOrder) Direction() string {
	return string(s.orderDirection)
}

func (s *StockOrder) Volume() float64 {
	return s.orderQty
}

func (s *StockOrder) Price() float64 {
	return s.orderPrice
}

func (s *StockOrder) AccountID() string {
	return s.account.AccountID()
}

func (s *StockOrder) CashAccount() string {
	return s.account.CashAccount()
}


// & StockOrder 实现了 tunnel.Order 接口 完毕
