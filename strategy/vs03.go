package strategy

import (
	// "fmt"
	// "fmt"
	"math"
	"sort"

	// "sort"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"

	// "github.com/wonderstone/QuantKit/tools/container/rank"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

// + Base cond and GEP Sort and buy
// + change max_min_ratio to 2.0

type VS03 struct {
	handler.EmptyStrategy // 继承EmptyStrategy类，可以减少实现不想处理的方法

	acc account.Account // 账户

	sInstrument []string   // 股票标的名称
	sIndicators []string   // 股票参与GEP指标名称
	tmpTargets  []SortRank // 临时目标集合

	bTriggerOnce bool // 是否已经触发过

	notTODay map[string]void // 是否为起始日集合

	fMaxMinRatio   float64 // 等差数列差值
	fCashUsedRatio float64 // 资金利用率
	iBuyNum        int     // 买入数量
	// vsBuyNum          int       // vs买入数量
	// sBuyAmtRatioSlice []float64 // 买入金额比例切片
	// fInitMoney        float64   // 初始资金

	lvMode string // 是否为LV模式 last value mode will have more frequency

	bGEPMode bool // 是否为GEP模式
	bLogOn   bool // 是否打印日志
}

func (s *VS03) OnInitialize(framework handler.Framework) {

	s.acc = framework.Account().GetAccountByID(config.AccountTypeStockSimple)
	c := framework.ConfigStrategyFromFile()
	// ifODay
	s.notTODay = make(map[string]void)
	// 获取
	s.sInstrument = framework.Config().Framework.Instrument
	s.sIndicators = framework.Config().Framework.Indicator

	// 触发交易时间
	s.fMaxMinRatio = config.GetParam(c, "max_min_ratio", 2.0) // 资金使用量最大和最小的相差比例
	s.fCashUsedRatio = config.GetParam(c, "cash_used_ratio", 0.90)
	s.iBuyNum = config.GetParam(c, "buy_num", 5)
	s.lvMode = config.GetParam(c, "lv_mode", "lv")
	// get bLogOn
	tmp := config.GetParam(c, "log_on", "false")
	switch tmp {
	case "true", "True", "TRUE", "t", "T":
		s.bLogOn = true
	default:
		s.bLogOn = false
	}
	// get bGEPMode
	tmp = config.GetParam(c, "gep_mode", "false")
	switch tmp {
	case "true", "True", "TRUE", "t", "T":
		s.bGEPMode = true
	default:
		s.bGEPMode = false
	}

}

func (s *VS03) OnDailyOpen(framework handler.Framework, marketType config.MarketType, acc ...account.Account) {
	// config.DebugF("%s 盘前时间", framework.CurrTime())
	s.bTriggerOnce = false
}

func (s *VS03) OnTick(
	// * 注意一下 这个函数返回在这里没什么用 在这里并没有尊重框架的设计
	// * 后续所有的买卖实际上都是通过buy、sell方法来给策略结构体内的虚拟账户进行插入订单的操作
	// * 而插入订单这个接口在StockSimple结构体中可以去同步实现链接交易功能 around line565

	// - 这里不输出，所以replay->nextmode around line 212 实际上被架空了
	// - 从一开始的设计来看，外层存在接收订单的接口实现然后对接柜台，这是最开始的想法
	// + 爱怎么用就怎么用吧，其实框架也只是辅助，不是必须
	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {
	// 判断股票标的切片SInstrNames是否为空，如果为空，则不操作股票数据循环
	if !s.bTriggerOnce {
		// reset s.tmpTargets
		s.tmpTargets = make([]SortRank, 0)
		defer func() { s.bTriggerOnce = true }()
		// print the time
		// fmt.Println(tm)
		// // check if the date is 2021-04-30
		// if tm.Format("2006-01-02") == "2021-04-30" {
		// 	fmt.Println("2021-04-30")
		// }
		// if tm.Format("2006-01-02") == "2021-05-06" {
		// 	fmt.Println("2021-05-06")
		// }

		// config.DebugF("%s 触发交易", framework.CurrTime())
		// && 调整到目标仓位模式
		// & Step 0: 确认标的操作顺序和数量, 这段代码位置并不固定
		// & Step 0.0 : 建立并初始化“信号集合”以产生“目标列表”，后者可能是前者的一部分
		// let signals be a SortRank slice
		signals := make([]SortRank, 0)
		// & Step 0.1 : 更新集合信息
		closes := make(map[string]float64) // 用于记录股票标的名称和close价
		for pIndicate := indicators.Oldest(); pIndicate != nil; pIndicate = pIndicate.Next() {
			indicate := pIndicate.Value
			closePrice := indicate.ConvertToFloat("Close")
			closes[pIndicate.Key] = closePrice
			// check if pIndicate.Key is in s.notTODay
			_, notT0 := s.notTODay[pIndicate.Key]
			notNaN := !ContainNaN(indicate)
			// 如果数据不包含NaN值(有效)，且不是初始天，添加到signals集合
			if notNaN && notT0 {
				// fmt.Println(tm)
				tradeSignal := make(model.OutputValues, 0)
				bp, errbp := indicate.TryConvertToFloat("bp")
				np, errnp := indicate.TryConvertToFloat("np")
				div, errdiv := indicate.TryConvertToFloat("div")

				condall := false

				if errbp != nil || errnp != nil || errdiv != nil {
					condall = false
				} else {
					cond1 := bp >= indicate.ConvertToFloat("bpc")
					cond2 := np <= indicate.ConvertToFloat("npc")
					cond3 := div >= indicate.ConvertToFloat("divc")
					condall = cond1 && cond2 && cond3
				}

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

					// // check if tradeSignal[0] != 0
					// if tradeSignal[0] != 0 {
					// 	// continue with the logic or actions here
					// 	fmt.Println("*****************")
					// }

					// gepCond := tradeSignal[0] > 0.0
					signals = append(signals, SortRank{instID: pIndicate.Key, signal: tradeSignal[0], close: closePrice})
					// condall = condall && gepCond

					// signals.Insert(SortRank{instID: pIndicate.Key, signal: tradeSignal[0], close: closePrice})
				} else {
					if condall {
						signals = append(signals, SortRank{instID: pIndicate.Key, signal: 1.0, close: closePrice})
					} else {
						signals = append(signals, SortRank{instID: pIndicate.Key, signal: 0.0, close: closePrice})
					}

				}
			}
			// 如果数据不包含NaN值(有效)，是初始天，躲避初始日。
			// 指标逻辑处理后添加到notTODay集合
			if notNaN && !notT0 {
				s.notTODay[pIndicate.Key] = void{}
			}
			// 如果不是初始天，但是有NaN值，删除
			// 有缺失值，但不排除今后还会再有，再有数值就是初始天
			// if !notNaN && notT0 {
			// 	delete(s.notTODay, pIndicate.Key)
			// }
		}

		// & Step 0.1 : 确认标的操作顺序 例如SB中 按照signal逆序计算等差买入金额
		// & 可能存在潜在的排序操作
		// sort.Slice(
		// 	signals.Values(), func(i, j int) bool {
		// 		return signals.Values()[i].signal < signals.Values()[j].signal
		// 	},
		// )
		// if signals have values
		if len(signals) != 0 {
			tmp := filterSignal(signals, func(sr SortRank) bool {
				return sr.close > 0.0
			})
			sort.Slice(tmp, func(i, j int) bool { return tmp[i].signal < tmp[j].signal })
			s.tmpTargets = tmp
		}



		// & Step 0.2 : 确认标的操作数量
		// 计算等差买入金额
		// amt := arithmeticSequence(s.acc.GetAsset().Total*s.fCashUsedRatio, len(signals.Values()), s.fMaxMinRatio)

		amt := make([]float64, len(s.tmpTargets))
		if len(signals) != 0 {
			amt = arithmeticSequence(s.acc.GetAsset().Total*s.fCashUsedRatio, len(s.tmpTargets), s.fMaxMinRatio)
		}

		// & Step 1: 如果目标列表是空的，表明空仓，遍历虚拟账户全部卖出, 返回
		// if len(s.tmpTargets) == 0 {
		// 	// 全部卖出
		// 	for _, pos := range s.acc.GetPosition() {
		// 		if closePrice, ok := closes[pos.InstID()]; ok {
		// 			if s.bLogOn {
		// 				config.WarnF("%s 卖出全部: %s", framework.CurrTime(), pos.InstID())
		// 			}
		// 			s.sell(pos.InstID(), pos.Volume(), closePrice)
		// 		} else if s.bLogOn {
		// 			config.WarnF("%s 未找到对应的收盘价: %s", framework.CurrTime(), pos.InstID())
		// 		}
		// 	}

		// 	return
		// }

		// & Step 2: 如果目标列表不为空，表明有调整，遍历虚拟账户，买入或卖出
		if len(s.tmpTargets) != 0 {
			// & Step 2.1 : 遍历虚拟账户持仓，买入或卖出
			// & Step 2.1.1 : 得到虚拟账户临时持仓信息
			totalPos := s.acc.GetPosition()
			// & Step 2.1.2 : 按照目标标的列表的顺序和数量，遍历操作，标的操作完成后，删除对应的标的
			for i := len(s.tmpTargets) - 1; i >= 0; i-- {
				expectedPosAmt := amt[i]
				if pos, ok := s.acc.GetPositionByInstID(s.tmpTargets[i].instID); ok {
					deltaAmt := expectedPosAmt - pos.Amt()
					instID := s.tmpTargets[i].instID
					buyNum := framework.Contract().GetContract(instID).CalcMaxQty(deltaAmt / s.tmpTargets[i].close)

					if buyNum == 0 {
						continue
					}

					if deltaAmt > 0 {
						if s.bLogOn {
							config.WarnF("%s 买入: %s", framework.CurrTime(), pos.InstID())
						}
						// s.buy(signals.Values()[i].instID, buyNum, signals.Values()[i].close)

						defer s.buy(s.tmpTargets[i].instID, buyNum, s.tmpTargets[i].close)
					} else {
						if s.bLogOn {
							config.WarnF("%s 卖出: %s", framework.CurrTime(), pos.InstID())
						}
						buyNum = math.Min(pos.Volume(account.WithSellAvailable(true)), buyNum)
						s.sell(s.tmpTargets[i].instID, -buyNum, s.tmpTargets[i].close)
					}
				} else {
					instID := s.tmpTargets[i].instID
					buyNum := framework.Contract().GetContract(instID).CalcMaxQty(expectedPosAmt / s.tmpTargets[i].close)

					if buyNum == 0 {
						continue
					}
					if s.bLogOn {
						config.WarnF("%s 买入: %s", framework.CurrTime(), instID)
					}
					// s.buy(signals.Values()[i].instID, buyNum, signals.Values()[i].close)

					defer s.buy(s.tmpTargets[i].instID, buyNum, s.tmpTargets[i].close)
				}

				delete(totalPos, s.tmpTargets[i].instID)
			}
			// & Step 2.1.3 : 临时持仓信息剩余标的均卖出
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

	}

	return
}

func (s *VS03) buy(instID string, qty, price float64) bool {
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

func (s *VS03) sell(instID string, qty, price float64) bool {
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
