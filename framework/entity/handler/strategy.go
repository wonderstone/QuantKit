package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type Strategy interface {
	// OnGlobalOnce 全局一次调用的设置功能逻辑，一般用于指定一些全局性设置, 如日志等
	OnGlobalOnce(global Global)

	// OnInitialize 初始化策略
	OnInitialize(framework Framework)

	// OnStart 策略初始化完毕开始执行前
	OnStart(framework Framework)

	// OnDailyOpen 每日开盘前
	OnDailyOpen(framework Framework, marketType config.MarketType, acc ...account.Account)

	// OnTick 每组K线
	OnTick(
		framework Framework, tm time.Time,
		indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
	) (orders []account.Order)

	// OnDailyClose 每日收盘后
	OnDailyClose(framework Framework, acc map[string]account.Account)

	// OnEnd 运行结束
	OnEnd(framework Framework)
}

// StrategyCreator 创建策略的函数
type StrategyCreator func() Strategy

// EmptyStrategy 空策略，用于隐藏不需要的策略实现
type EmptyStrategy struct{}

func (e EmptyStrategy) OnGlobalOnce(global Global) {}

func (e EmptyStrategy) OnInitialize(framework Framework) {}

func (e EmptyStrategy) OnStart(framework Framework) {}

func (e EmptyStrategy) OnDailyOpen(
	framework Framework,
	marketType config.MarketType, acc ...account.Account,
) {
}

func (e EmptyStrategy) OnTick(
	framework Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {
	return
}

func (e EmptyStrategy) OnDailyClose(framework Framework, acc map[string]account.Account) {}

func (e EmptyStrategy) OnEnd(framework Framework) {}
