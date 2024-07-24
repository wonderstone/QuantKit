package strategy

import (
	"math"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/times"
)


type T0State int // ~ 交易状态(内部) 0:未买入 1:已买入 -1:已卖出
const (
	NotBuy T0State = iota
	Bought         = 1
	Sold           = -1
)

type singleStock struct {
	initBTstate     bool     // ~ 是否已经初始化了标的回测买入状态(内部)
	holdAmt         float64  // ~ 股票标的持仓金额
	holdNum         int      // ~ 股票标的持仓数量
	availableNum    int      // ~ 股票标的可用数量(内部)
	tCounter        int      // 交易次数计数器(内部)
	tState          T0State  // ~ 交易状态(内部) 0:未买入 1:已买入 -1:已卖出
	ifReCoverMap    bool     // 是否恢复了股票持仓(内部)
	lastTradeSignal *float64 // 公式上一次买入/卖出信号数值
	IfBT            bool     // ~ 是否为回测/实盘状态 T0特有问题 回测第一天为资金状态 需要买入标的
}

type T0 struct {
	handler.EmptyStrategy
	acc account.Account

	InstNames []string // 股票标的名称
	IndiNames []string // 股票参与GEP指标名称，注意其数量不大于BarDE内信息数量，且strategy内可见BarDE的数据

	InitBuyTime time.Duration // ~ 初始化买入时间
	StartTime   time.Duration // ~ 交易开始时间
	StopTime    time.Duration // ~ 交易结束时间

	CurrInitBuyTime time.Time
	CurrStartTime   time.Time
	CurrStopTime    time.Time

	Tlimit int // 每日交易次数限制

	HoldAmt float64 // ~ 股票标的持仓金额

	stocks map[string]*singleStock // 股票标的持仓信息(内部)

	logOn   bool // 是否打印日志
	GEPMode bool // 是否为GEP模式
}

func (s *T0) OnGlobalOnce(global handler.Global) {
	global.SetGEPInputParams(s.IndiNames)
}

func (s *T0) OnInitialize(framework handler.Framework) {
	c := framework.ConfigStrategyFromFile()
	s.acc = framework.Account().GetAccountByID(config.AccountTypeStockSimple)

	s.InstNames = config.GetParam(c, "instruments", framework.Config().Framework.Instrument)
	s.IndiNames = config.GetParam(c, "indicators", framework.Config().Framework.Indicator)

	s.HoldAmt = config.GetParam(c, "cash_used_ratio", 0.8) * s.acc.GetAsset().Total
	StartTime := config.GetParam(c, "StartTime", "09:30")

	s.StartTime = times.MustDuration("15:04", StartTime)

	StopTime := config.GetParam(c, "StopTime", "14:50")
	s.StopTime = times.MustDuration("15:04", StopTime)

	InitBuyTime := config.GetParam(c, "InitBuyTime", "14:30")
	s.InitBuyTime = times.MustDuration("15:04", InitBuyTime)

	s.Tlimit = config.GetParam(c, "Tlimit", 4)

	s.stocks = make(map[string]*singleStock)

	s.logOn = true
	s.GEPMode = true

	for _, name := range s.InstNames {
		s.stocks[name] = &singleStock{
			initBTstate:  false,
			holdAmt:      s.HoldAmt,
			holdNum:      0,
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
func (s *T0) ContainNaN(m dataframe.StreamingRecord) bool {
	for _, x := range m.Data {
		if len(x) == 0 {
			return true
		}
	}
	return false
}

func (s *T0) OnDailyOpen(framework handler.Framework, marketType config.MarketType, acc ...account.Account) {
	date := framework.CurrDate()
	s.CurrInitBuyTime = date.Add(s.InitBuyTime)
	s.CurrStartTime = date.Add(s.StartTime)
	s.CurrStopTime = date.Add(s.StopTime)

	for i, _ := range s.stocks {
		// counter 重置
		s.stocks[i].ifReCoverMap = false
		s.stocks[i].tCounter = 0

		s.stocks[i].lastTradeSignal = nil
	}
}

func (s *T0) buy(instID string, qty, price float64, stockInfo *singleStock) bool {
	// 买入
	o, err := s.acc.NewOrder(
		instID, qty,
		account.WithOrderPrice(price),
	)

	if err != nil {
		config.ErrorF("创建买入失败: %v", err)
	}

	err = s.acc.InsertOrder(
		o,
		account.WithCheckCash(true),
	)

	if err != nil {
		if s.logOn {
			config.DebugF(
				"买入失败: %f, %f, %s, %s, %v", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(), err,
			)
		}

		return false
	} else {
		if s.logOn {
			config.DebugF(
				"买入: %f, %f, %s, %s", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(),
			)
		}
		return true
	}
}

func (s *T0) sell(instID string, qty, price float64, stockInfo *singleStock) bool {
	// 卖出
	o, err := s.acc.NewOrder(
		instID, qty,
		account.WithOrderPrice(price),
		account.WithOrderDirection(config.OrderSell),
	)

	if err != nil {
		config.ErrorF("创建卖出失败: %v", err)
	}

	err = s.acc.InsertOrder(
		o,
		account.WithCheckPosition(true),
	)

	if err != nil {
		if s.logOn {
			config.DebugF(
				"卖出失败: %f, %f, %s, %s, %v", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(), err,
			)
		}
		return false
	} else {
		if s.logOn {
			config.DebugF(
				"卖出: %f, %f, %s, %s", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(),
			)
		}
		return true
	}
}

func (s *T0) OnTick(
	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {
	// timeDay := tm.Format("15:04")

	// 2. 获取当前时间并判定是否介于StartTime与StopTime，是则进行常规操作
	// 否 且大于StopTime,则查看恢复操作状态并进行恢复持仓操作。
	for pIndicator := indicators.Oldest(); pIndicator != nil; pIndicator = pIndicator.Next() {
		instID := pIndicator.Key

		stockInfo, ok := s.stocks[instID]
		if !ok {
			s.stocks[instID] = &singleStock{
				initBTstate:  false,
				holdAmt:      s.HoldAmt,
				availableNum: 0,
				tCounter:     0,
				tState:       0,
				ifReCoverMap: false,
				IfBT:         true,
			}

			stockInfo = s.stocks[instID]
		}

		if !s.ContainNaN(pIndicator.Value) {
			closePrice := pIndicator.Value.ConvertToFloat("Close")

			qty := 0
			// # 2.1 在ifBT状态下，需要特判第一天的买入操作
			if stockInfo.IfBT {
				// / 因为T+1,所以不用再担心第一天股票交易时间段下单问题。
				if !stockInfo.initBTstate {
					if tm.After(s.CurrInitBuyTime) {
						qty := framework.Contract().GetContract(instID).CalcMaxQty(stockInfo.holdAmt / closePrice)
						if s.buy(instID, qty, closePrice, stockInfo) {
							stockInfo.holdNum = int(qty)
							stockInfo.initBTstate = true

							// tCounter计数直接置满
							stockInfo.tCounter = s.Tlimit * 2
							// tState 赋值
							stockInfo.tState = NotBuy

							stockInfo.IfBT = false
						}
					}
				}
			} else if stockInfo.tCounter < s.Tlimit*2 && tm.After(s.CurrStartTime) && tm.Before(s.CurrStopTime) {
				// ~ 计算策略依托的有效数值
				// 常规手动操作
				// % GEP 引入
				var GEPSlice = make(model.InputValues, len(s.IndiNames))
				for i := 0; i < len(s.IndiNames); i++ {
					GEPSlice[i] = pIndicator.Value.ConvertToFloat(s.IndiNames[i])
				}

				// % GEP 应用 注意这是genome的应用
				tradeSignal := 0.0
				if s.GEPMode {
					tradeSignal = framework.Evaluate(GEPSlice)[0]
				} else {
					tradeSignal = pIndicator.Value.ConvertToFloat("Close") - pIndicator.Value.ConvertToFloat("Amount")/pIndicator.Value.ConvertToFloat("Volume")
				}

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
						qty += stockInfo.holdNum / s.Tlimit

						// ~lastBuyValue 赋值 lastSellValue 赋值
						stockInfo.lastTradeSignal = &tradeSignal
						// fmt.Printf("交易信号正: %v\n", *framework.CurrTime())
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
						qty -= stockInfo.holdNum / s.Tlimit

						// 上一次交易信号赋值
						stockInfo.lastTradeSignal = &tradeSignal
						// fmt.Printf("交易信号负: %v\n", *framework.CurrTime())
					}
				}
			} else if stockInfo.initBTstate && tm.Compare(s.CurrStopTime) >= 0 {
				if !stockInfo.ifReCoverMap {
					// 恢复持仓
					if pos, ok := s.acc.GetPositionByInstID(instID); ok {
						qty = stockInfo.holdNum - int(pos.Volume())
					} else {
						qty = stockInfo.holdNum
					}
					stockInfo.ifReCoverMap = true
				}

			}

			if qty > 0 {
				s.buy(instID, float64(qty), closePrice, stockInfo)
				// tState 赋值
				stockInfo.tState = Bought
				stockInfo.tCounter += 1

			} else if qty < 0 {
				s.sell(instID, float64(-qty), closePrice, stockInfo)
				// tState 赋值
				stockInfo.tState = Sold
				stockInfo.tCounter += 1
			}

		}
	}

	return
}
