package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/tunnel"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

// ? 注意！ 这里的Account接口是一个多账户概念，为何没有在entity中定义Accounts?
// ! 一方面是Resource中包含basic,在account操作中需要？故而为了避免循环引入，进而设置在这里？？？？？




type Account interface {
	Resource
	// GetAccounts 获取账户信息
	GetAccounts() map[string]account.Account
	// GetAccount 获取账户信息
	// * 注意返回了账户切片
	GetAccount(market config.MarketType) []account.Account
	// GetAccountByID 获取账户信息
	GetAccountByID(accountID string) account.Account
	// IterOrders 迭代所有order，WithFilter可以过滤
	IterOrders(instId string, filter ...account.WithOrderFilter) <-chan account.Order
	// NewOrder 新建order
	// * 注意：这里有通过什么accountID下单的问题
	NewOrder(accountID, instID string, qty float64, option ...account.WithOrderOption) (account.Order, error)
	// InsertOrder 插入order
	// * 注意：这里的下单并没有需要指定accountID的问题，因为NewOrder时已经指定了accountID
	InsertOrder(orders account.Order, option ...account.WithOrderOption) error
	// CancelOrder 取消order
	// * 注意：这里有通过什么accountID取消的问题
	CancelOrder(tm time.Time, id int64, accountId string)
	// GetCurrTime 获取当前时间
	GetCurrTime() *time.Time
}





// 扩展entity.Account接口、具备handler的Init方法
type Account2 interface {
	account.Account
	// Init 初始化账户
	Init(handler Accounts, option any) error

	// Release 释放资源
	Release()

	// AddDividend 添加分红和税
	AddDividend(dividend, tax float64) // 添加分红

	// AddDynamicPnL 添加动态盈亏
	AddDynamicPnL(pnl ...float64) // 添加动态盈亏

	// AddAvailable 添加可用资金
	AddAvailable(available float64) // 添加可用资金

	// DoOrderUpdate 更新订单
	DoOrderUpdate(order Order) // 更新账户信息

	// DoTradeUpdate 更新订单
	DoTradeUpdate(qty, price float64, order Order) // 更新账户信息

	// DoSettle 结算账户信息
	DoSettle(
		tm time.Time, indicators *orderedmap.OrderedMap[string, dataframe.StreamingRecord],
		recorder map[config.RecordType]recorder.Handler,
	)

	// DoMatch 账户撮合
	DoMatch(tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord], matcher Matcher)

	// DoResume 恢复账户信息(实盘)
	DoResume(
		assets []recorder.AssetRecord, 
		positions []recorder.PositionRecord, 
		orders []recorder.OrderRecord)

	// CalcPositionPnL 计算持仓盈亏
	CalcPositionPnL(tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord])

	// CalcSettleInfo 计算结算数据
	// 如果换股，instID会变化，因此直接清空本position的数据，添加或者更新新的position
	CalcSettleInfo(instID string, qty, price, lastPrice float64) (
		settleQty, settlePrice, settleLastPrice, dividend, tax float64,
	)

	// GenOrderId 生成订单ID
	GenOrderId() int64

	// Base 获取基础数据
	Base() Basic

	// DoRTInsert 实盘插入订单
	DoRTInsert(order tunnel.Order) (string, error)

	// DoRTCancel 实盘取消订单
	DoRTCancel(orderID string) error

	// GetRunId 获取运行ID
	GetRunId() string
}



type Accounts interface {
	Account
	// Init 初始化账户
	Init(framework any, recorder map[config.RecordType]recorder.Handler) error

	// SetResource 设置资源
	SetResource(resource Resource)

	// Release 释放资源
	Release()

	// DoSettle 结算账户信息
	DoSettle(tm time.Time, ticks *orderedmap.OrderedMap[string, dataframe.StreamingRecord])

	// DoMatch 账户撮合
	DoMatch(tm time.Time, ticks orderedmap.OrderedMap[string, dataframe.StreamingRecord])

	// DoResume 恢复账户信息(实盘)
	DoResume()

	// CalcPositionPnL 计算持仓盈亏
	CalcPositionPnL(tm time.Time, ticks orderedmap.OrderedMap[string, dataframe.StreamingRecord])

	// GetHistoryAssetRecord 获取历史收益数据
	GetHistoryAssetRecord(records *[]recorder.AssetRecord) error

	// GenOrderId 生成订单ID
	GenOrderId() int64
}
