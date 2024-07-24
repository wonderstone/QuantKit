package strategy

import (
	"fmt"
	"math"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type T0State int // ~ 交易状态(内部) 0:未买入 1:已买入 -1:已卖出
const (
	NotBuy T0State = iota
	Bought         = 1
	Sold           = -1
)

type singleStock struct {
	initBTstate     bool     // ~ 是否已经初始化了标的回测买入状态(内部)
	holdNumMap      int      // ~ 股票标的持仓数量, key is the stock name,get from SInstNames and SHoldNums(内部)
	availableNum    int      // ~ 股票标的可用数量(内部)
	tCounter        int      // 交易次数计数器(内部)
	tState          T0State  // ~ 交易状态(内部) 0:未买入 1:已买入 -1:已卖出
	ifReCoverMap    bool     // 是否恢复了股票持仓(内部)
	lastTradeSignal *float64 // 公式上一次买入/卖出信号数值
	IfBT            bool     // ~ 是否为回测/实盘状态 T0特有问题 回测第一天为资金状态 需要买入标的
}

type T0 struct {
	handler.EmptyStrategy
	acc       account.Account
	lastYear  int        // 上一交易日日期(内部)
	lastMonth time.Month // 上一交易日日期(内部)
	lastDay   int        // 上一交易日日期(内部)

	InstNames []string // 股票标的名称
	IndiNames []string // 股票参与GEP指标名称，注意其数量不大于BarDE内信息数量，且strategy内可见BarDE的数据

	InitBuyTime string // ~ 初始化买入时间
	StartTime   string // ~ 交易开始时间
	StopTime    string // ~ 交易结束时间
	Tlimit      int    // 每日交易次数限制

	HoldNums []int // ~ 股票标的持仓数量, 外部获取

	stocks map[string]*singleStock // 股票标的持仓信息(内部)
}

func (t0 *T0) OnInitialize(framework handler.Framework) {
	t0.acc = framework.Account().GetAccountByID(config.AccountTypeStockSimple)
	t0.InstNames = framework.Config().Framework.Instrument
	t0.IndiNames = framework.Config().Framework.Indicator

	t0.HoldNums = []int{800}
	t0.StartTime = "093000"
	t0.StopTime = "145000"
	t0.InitBuyTime = "143000"
	t0.Tlimit = 4

	t0.stocks = make(map[string]*singleStock)

	for _, name := range t0.InstNames {
		t0.stocks[name] = &singleStock{
			initBTstate:  false,
			holdNumMap:   t0.HoldNums[0],
			availableNum: 0,
			tCounter:     0,
			tState:       0,
			ifReCoverMap: false,
			IfBT:         true,
		}
	}
}

func isSameTradingDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()

	return y1 == y2 && m1 == m2 && d1 == d2
}

// ContainNaN 此处是为了停盘数据处理设定的规则相检查用的
func ContainNaN(m dataframe.StreamingRecord) bool {
	for _, x := range m.Data {
		if len(x) == 0 {
			return true
		}
	}
	return false
}

func (t0 *T0) OnDailyOpen(framework handler.Framework, marketType config.MarketType, acc ...account.Account) {
	fmt.Println(*framework.CurrTime())
}

func (t0 *T0) buy(instID string, qty, price float64, stockInfo *singleStock) bool {
	// 买入
	o, err := t0.acc.NewOrder(
		instID, qty,
		account.WithOrderPrice(price),
	)

	if err != nil {
		config.ErrorF("创建买入失败: %v", err)
	}

	err = t0.acc.InsertOrder(
		o,
		account.WithCheckCash(true),
	)

	if err != nil {
		config.DebugF(
			"买入失败: %f, %f, %s, %s, %v", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
			o.OrderTime().String(), err,
		)
		return false
	} else {
		config.DebugF(
			"买入: %f, %f, %s, %s", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
			o.OrderTime().String(),
		)
		return true
	}
}

func (t0 *T0) sell(instID string, qty, price float64, stockInfo *singleStock) bool {
	// 卖出
	o, err := t0.acc.NewOrder(
		instID, qty,
		account.WithOrderPrice(price),
		account.WithOrderDirection(config.OrderSell),
	)

	if err != nil {
		config.ErrorF("创建卖出失败: %v", err)
	}

	err = t0.acc.InsertOrder(
		o,
		account.WithCheckPosition(true),
	)

	if err != nil {
		config.DebugF(
			"卖出失败: %f, %f, %s, %s, %v", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
			o.OrderTime().String(), err,
		)
		return false
	} else {
		config.DebugF(
			"卖出: %f, %f, %s, %s", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
			o.OrderTime().String(),
		)

		return true
	}
}

func (t0 *T0) OnTick(
	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {
	// 1. check if a new day, then change all ifReCoverMap to false
	// # 1.1 使用字符串形式比较日期，一旦VDS更改格式需要调整
	y, m, d := tm.Date()
	if t0.lastDay != d || t0.lastMonth != m || t0.lastYear != y {
		for i, _ := range t0.stocks {
			// counter 重置
			t0.stocks[i].ifReCoverMap = false
			t0.stocks[i].tCounter = 0

			t0.stocks[i].lastTradeSignal = nil
		}
	}

	timeDay := tm.Format("150405")

	// 2. 获取当前时间并判定是否介于StartTime与StopTime，是则进行常规操作
	// 否 且大于StopTime,则查看恢复操作状态并进行恢复持仓操作。
	for pIndicator := indicators.Oldest(); pIndicator != nil; pIndicator = pIndicator.Next() {
		instID := pIndicator.Key

		stockInfo, ok := t0.stocks[instID]
		if !ok {
			t0.stocks[instID] = &singleStock{
				initBTstate:  false,
				holdNumMap:   t0.HoldNums[0],
				availableNum: 0,
				tCounter:     0,
				tState:       0,
				ifReCoverMap: false,
				IfBT:         true,
			}

			stockInfo = t0.stocks[instID]
		}

		if !ContainNaN(pIndicator.Value) {
			closePrice := pIndicator.Value.ConvertToFloat("Close")

			qty := 0

			// # 2.1 在ifBT状态下，需要特判第一天的买入操作
			if stockInfo.IfBT {
				// / 因为T+1,所以不用再担心第一天股票交易时间段下单问题。
				if !stockInfo.initBTstate {
					if timeDay >= t0.InitBuyTime {
						if t0.buy(instID, float64(stockInfo.holdNumMap), closePrice, stockInfo) {
							stockInfo.initBTstate = true

							// tCounter计数直接置满
							stockInfo.tCounter = t0.Tlimit * 2
							// tState 赋值
							stockInfo.tState = NotBuy

							stockInfo.IfBT = false
						}
					}
				}
			} else if stockInfo.tCounter < t0.Tlimit*2 && timeDay >= t0.StartTime && timeDay <= t0.StopTime {
				// ~ 计算策略依托的有效数值
				// 常规手动操作
				// % GEP 引入
				var GEPSlice = make(model.InputValues, len(t0.IndiNames))
				for i := 0; i < len(t0.IndiNames); i++ {
					GEPSlice[i] = pIndicator.Value.ConvertToFloat(t0.IndiNames[i])
				}

				// % GEP 应用 注意这是genome的应用
				// tradeSignal := framework.Evaluate(GEPSlice)[0]
				tradeSignal := pIndicator.Value.ConvertToFloat("Close") - pIndicator.Value.ConvertToFloat("Amount")/pIndicator.Value.ConvertToFloat("Volume")

				if math.IsNaN(tradeSignal) || math.IsInf(tradeSignal, 0) {
					return
				}

				// ~ 策略买入逻辑
				// 指标计算出的交易信号 > 0
				// 上一次交易信号 < 0 或者 无交易信号
				// 交易状态为非买入
				// 交易次数计数器 < 交易次数限制 * 2
				if tradeSignal > 0 {
					if (stockInfo.lastTradeSignal == nil || *stockInfo.lastTradeSignal < 0) &&
						stockInfo.tState != Bought {
						qty += stockInfo.holdNumMap / t0.Tlimit

						// ~lastBuyValue 赋值 lastSellValue 赋值
						stockInfo.lastTradeSignal = &tradeSignal
						fmt.Printf("交易信号正: %v\n", *framework.CurrTime())
					}
				} else
				// ~ 策略卖出逻辑
				// 指标计算出的交易信号 < 0
				// 上一次交易信号 > 0 或者 无交易信号
				// 交易状态为非卖出
				// 交易次数计数器 < 交易次数限制 * 2
				if tradeSignal < 0 {
					if (stockInfo.lastTradeSignal == nil || *stockInfo.lastTradeSignal > 0) &&
						stockInfo.tState != Sold {
						qty -= stockInfo.holdNumMap / t0.Tlimit

						// 上一次交易信号赋值
						stockInfo.lastTradeSignal = &tradeSignal
						fmt.Printf("交易信号负: %v\n", *framework.CurrTime())
					}
				}
			} else if stockInfo.initBTstate && timeDay > t0.StopTime {
				if !stockInfo.ifReCoverMap {
					// 恢复持仓
					if pos, ok := t0.acc.GetPositionByInstID(instID); ok {
						qty = stockInfo.holdNumMap - int(pos.Volume())
					} else {
						qty = stockInfo.holdNumMap
					}
					stockInfo.ifReCoverMap = true
				}

			}

			if qty > 0 {
				t0.buy(instID, float64(qty), closePrice, stockInfo)
				// tState 赋值
				stockInfo.tState = Bought
				stockInfo.tCounter += 1

			} else if qty < 0 {
				t0.sell(instID, float64(-qty), closePrice, stockInfo)
				// tState 赋值
				stockInfo.tState = Sold
				stockInfo.tCounter += 1
			}

		}
	}

	t0.lastYear, t0.lastMonth, t0.lastDay = y, m, d

	return
}
