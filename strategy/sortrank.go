package strategy

import (
	"math"
	"sort"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/container/rank"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type SortRank struct {
	instID string  // 合约ID
	close  float64 // 收盘价
	signal float64 // 信号
	// posAmt float64 // 持仓金额
}

type SortBuy struct {
	handler.EmptyStrategy // 继承EmptyStrategy类，可以减少实现不想处理的方法

	acc account.Account // 账户

	sInstrument []string      // 股票标的名称
	sIndicators []string      // 股票参与GEP指标名称
	timeTrigger time.Duration // 时间关键字，用于判断是否需要进行交易

	timeNowTrigger time.Time // 当前触发时间
	bTriggerOnce   bool      // 是否已经触发过

	fMaxMinRatio      float64   // 等差数列差值
	fCashUsedRatio    float64   // 资金利用率
	iBuyNum           int       // 买入数量
	// sBuyAmtRatioSlice []float64 // 买入金额比例切片
	// fInitMoney        float64   // 初始资金

	bGEPMode bool // 是否为GEP模式
	bLogOn   bool // 是否打印日志
}

func (s *SortBuy) OnInitialize(framework handler.Framework) {
	s.bLogOn = false
	s.bGEPMode = true

	s.acc = framework.Account().GetAccountByID(config.AccountTypeStockSimple)

	c := framework.ConfigStrategyFromFile()

	// 获取
	s.sInstrument = framework.Config().Framework.Instrument
	s.sIndicators = framework.Config().Framework.Indicator

	// 触发交易时间
	s.fMaxMinRatio = config.GetParam(c, "max_min_ratio", 2.0) // 资金使用量最大和最小的相差比例
	s.fCashUsedRatio = config.GetParam(c, "cash_used_ratio", 0.90)
	s.iBuyNum = config.GetParam(c, "buy_num", 5)
}

func (s *SortBuy) OnDailyOpen(framework handler.Framework, marketType config.MarketType, acc ...account.Account) {
	// config.DebugF("%s 盘前时间", framework.CurrTime())
	cur := framework.CurrTime()
	date := time.Date(cur.Year(), cur.Month(), cur.Day(), 0, 0, 0, 0, cur.Location())
	s.timeNowTrigger = date.Add(s.timeTrigger)
	s.bTriggerOnce = false
}

func (s *SortBuy) OnTick(
	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {
	// 判断股票标的切片SInstrNames是否为空，如果为空，则不操作股票数据循环
	if !s.bTriggerOnce {
		defer func() { s.bTriggerOnce = true }()

		// config.DebugF("%s 触发交易", framework.CurrTime())

		signals := rank.NewSortedHeap(
			s.iBuyNum, func(i, j SortRank) bool { return i.signal < j.signal },
		) // 用于记录股票标的名称和信号值，逆序排列

		closes := make(map[string]float64) // 用于记录股票标的名称和close价
		for pIndicate := indicators.Oldest(); pIndicate != nil; pIndicate = pIndicate.Next() {
			indicate := pIndicate.Value
			closePrice := indicate.ConvertToFloat("Close")
			closes[pIndicate.Key] = closePrice
			if !ContainNaN(indicate) {
				tradeSignal := make(model.OutputValues, 0)
				if s.bGEPMode {
					// % GEP 引入
					var GEPSlice = make(model.InputValues, len(s.sIndicators))
					for i := 0; i < len(s.sIndicators); i++ {
						GEPSlice[i] = indicate.ConvertToFloat(s.sIndicators[i])
					}

					tradeSignal = framework.Evaluate(GEPSlice)

					if math.IsNaN(tradeSignal[0]) || math.IsInf(tradeSignal[0], 0) {
						continue
					}

					signals.Insert(SortRank{instID: pIndicate.Key, signal: tradeSignal[0], close: closePrice})
				} else {
					signals.Insert(
						SortRank{
							instID: pIndicate.Key, signal: closePrice - indicate.ConvertToFloat("MA3"),
							close: closePrice,
						},
					)
				}
			}
		}

		if len(signals.Values()) == 0 {
			// 全部卖出
			for _, pos := range s.acc.GetPosition() {
				if closePrice, ok := closes[pos.InstID()]; ok {
					if s.bLogOn {
						config.WarnF("%s 卖出全部: %s", framework.CurrTime(), pos.InstID())
					}
					s.sell(pos.InstID(), pos.Volume(), closePrice)
				} else if s.bLogOn {
					config.WarnF("%s 未找到对应的收盘价: %s", framework.CurrTime(), pos.InstID())
				}
			}

			return
		}

		// 按照signal逆序计算等差买入金额
		sort.Slice(
			signals.Values(), func(i, j int) bool {
				return signals.Values()[i].signal < signals.Values()[j].signal
			},
		)

		// 计算等差买入金额
		amt := arithmeticSequence(s.acc.GetAsset().Total*s.fCashUsedRatio, len(signals.Values()), s.fMaxMinRatio)

		totalPos := s.acc.GetPosition()

		for i := len(signals.Values())-1 ; i >= 0; i--{
			expectedPosAmt := amt[i]
			if pos, ok := s.acc.GetPositionByInstID(signals.Values()[i].instID); ok {
				deltaAmt := expectedPosAmt - pos.Amt()
				instID := signals.Values()[i].instID
				buyNum := framework.Contract().GetContract(instID).CalcMaxQty(deltaAmt / signals.Values()[i].close)

				if buyNum == 0 {
					continue
				}

				if deltaAmt > 0 {
					if s.bLogOn {
						config.WarnF("%s 买入: %s", framework.CurrTime(), pos.InstID())
					}
					// s.buy(signals.Values()[i].instID, buyNum, signals.Values()[i].close)

					defer s.buy(signals.Values()[i].instID, buyNum, signals.Values()[i].close)
				} else {
					if s.bLogOn {
						config.WarnF("%s 卖出: %s", framework.CurrTime(), pos.InstID())
					}
					buyNum = math.Min(pos.Volume(account.WithSellAvailable(true)), buyNum)
					s.sell(signals.Values()[i].instID, -buyNum, signals.Values()[i].close)
				}
			} else {
				instID := signals.Values()[i].instID
				buyNum := framework.Contract().GetContract(instID).CalcMaxQty(expectedPosAmt / signals.Values()[i].close)

				if buyNum == 0 {
					continue
				}
				if s.bLogOn {
					config.WarnF("%s 买入: %s", framework.CurrTime(), instID)
				}
				// s.buy(signals.Values()[i].instID, buyNum, signals.Values()[i].close)

				defer s.buy(signals.Values()[i].instID, buyNum, signals.Values()[i].close)
			}

			delete(totalPos, signals.Values()[i].instID)
		}

		for _, pos := range totalPos {
			if closePrice, ok := closes[pos.InstID()]; ok {
				s.sell(pos.InstID(), pos.Volume(account.WithSellAvailable(true)), closePrice)
				if s.bLogOn {
					config.WarnF("%s 卖出全部: %s", framework.CurrTime(), pos.InstID())
				}
			} else {
				if s.bLogOn {
					config.WarnF("%s 未找到对应的收盘价: %s", framework.CurrTime(), pos.InstID())
				}
			}
		}
	}

	return
}

func (s *SortBuy) buy(instID string, qty, price float64) bool {
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
		// account.WithCheckCash(true),
	)

	if err != nil {
		if s.bLogOn {
			config.DebugF(
				"买入失败: %f, %f, %s, %s, %v", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(), err,
			)
		}
		return false
	} else {
		if s.bLogOn {

			config.DebugF(
				"买入: %f, %f, %s, %s", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(),
			)
		}
		return true
	}
}

func (s *SortBuy) sell(instID string, qty, price float64) bool {
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
		if s.bLogOn {
			config.DebugF(
				"卖出失败: %f, %f, %s, %s, %v", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(), err,
			)
		}
		return false
	} else {
		if s.bLogOn {

			config.DebugF(
				"卖出: %f, %f, %s, %s", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
				o.OrderTime().String(),
			)
		}

		return true
	}
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

