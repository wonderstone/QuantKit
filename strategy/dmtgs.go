package strategy

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type DMTGSStockInfo struct {
	lastTradeSignal1 float64 // 公式上一次买入信号数值
	lastTradeSignal2 float64 // 公式上一次卖出信号数值
	fa              float64 // 各支股票对应可用资金  fund available
	indi            model.InputValues
	// triggerState    int     // 0: 未触发 1: 待触发 2: 已触发
}

type DMTGS struct {
	handler.EmptyStrategy
	acc account.Account

	initMoney       float64       // 初始资金
	SInstNames      []string      // 股票标的名称
	SIndiNames      []string      // 股票参与GEP指标名称，注意其数量不大于BarDE内信息数量，且strategy内可见BarDE的数据
	STimeCritic     time.Duration // 时间关键字，用于判断是否需要进行交易
	CurrTriggerTime time.Time     // 时间关键字，用于判断是否需要进行交易
	minCashRatio    float64

	stockInfos map[string]*DMTGSStockInfo // 股票信息

	logOn   bool // 是否打印日志
	GEPMode bool // 是否为GEP模式

	slip float64

	tradeSignals []TradeSignal
}

func (s *DMTGS) OnInitialize(framework handler.Framework) {
	s.acc = framework.Account().GetAccountByID(config.AccountTypeStockSimple)
	s.stockInfos = make(map[string]*DMTGSStockInfo, len(s.SInstNames))

	// 获取配置
	c := framework.ConfigStrategyFromFile()

	// 获取股票标的名称
	s.SInstNames = config.GetParam(c, "sinstrnames", framework.Config().Framework.Instrument)

	// 获取股票参与GEP指标名称
	s.SIndiNames = framework.Config().Framework.Indicator

	// 获取最小资金比例
	s.minCashRatio = config.GetParam(c, "mincashratio", 0.01)

	// 获取初始资金
	s.initMoney = s.acc.GetAsset().Initial

	// init the faMap
	average, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", s.initMoney/float64(len(s.SInstNames))), 64)
	for _, name := range s.SInstNames {
		s.stockInfos[name] = new(DMTGSStockInfo)
		// 账户初始资金充当市值情况
		s.stockInfos[name].fa = average
		s.stockInfos[name].lastTradeSignal1 = -1
		s.stockInfos[name].lastTradeSignal2 = -1
		// s.stockInfos[name].triggerState = 0 // 目前触发时间是标准参数，不需要单独触发状态
	}
	// ! change!
	s.slip = framework.Framework().Config().Framework.Stock.Slippage
	s.logOn = false
	s.GEPMode = true

}

func (s *DMTGS) OnDailyOpen(framework handler.Framework, marketType config.MarketType, acc ...account.Account) {
	//  框架已经保证，对于日频策略一天只会触发一次用户在盘中的操作，不需要单独设置触发时间
	// s.CurrTriggerTime = framework.CurrDate().Add(s.STimeCritic)
	// if marketType == config.MarketTypeStock {
	// 	for _, v := range s.stockInfos {
	// 		v.triggerState = 0
	// 	}
	// }
}

func (s *DMTGS) buy(instID string, qty, price float64) bool {
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

func (s *DMTGS) sell(instID string, qty, price float64) bool {
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

// ContainNaN 此处是为了停盘数据处理设定的规则相检查用的
func (s *DMTGS) ContainNaN(m dataframe.StreamingRecord) bool {
	for _, x := range m.Data {
		if len(x) == 0 {
			return true
		}
	}
	return false
}

func (s *DMTGS) OnTick(
	framework handler.Framework, tm time.Time, indicates orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {

	dealingFunc := func(
		instID string, stockInfo *DMTGSStockInfo, record ...*dataframe.StreamingRecord,
	) {


		var indicate *dataframe.StreamingRecord

		if len(record) != 0 {
			indicate = record[0]
		} else {
			if indi, ok := indicates.Get(instID); ok {
				indicate = &indi
			} else {
				return
			}
		}

		if !s.ContainNaN(*indicate) {
			tradeSignal := make(model.OutputValues, 0)
			var GEPSlice = make(model.InputValues, len(s.SIndiNames))
			if s.GEPMode {
				// % GEP 引入
				for i := 0; i < len(s.SIndiNames); i++ {
					GEPSlice[i] = indicate.ConvertToFloat(s.SIndiNames[i])
				}

				tradeSignal = framework.Evaluate(GEPSlice)
			} else {
				tradeSignal = append(
					tradeSignal, indicate.ConvertToFloat("Close")-indicate.ConvertToFloat("MA3"),
				)
			}

			if math.IsNaN(tradeSignal[0]) || math.IsInf(tradeSignal[0], 0) {
				return
			}

			nowSignal1 := tradeSignal[0]
			lastSignal1 := stockInfo.lastTradeSignal1

			nowSignal2 := tradeSignal[1]
			lastSignal2 := stockInfo.lastTradeSignal2

			// if stockInfo.triggerState == 1 {
			if nowSignal1 > 0 && lastSignal1 < 0 {
				closePrice := indicate.ConvertToFloat("Close")
				buyNum := framework.Contract().GetContract(instID).CalcMaxQty(stockInfo.fa / closePrice * (1 - s.minCashRatio))

				if buyNum != 0 {
					s.buy(instID, buyNum, indicate.ConvertToFloat("Close"))

					tmpOrder, _ := s.acc.NewOrder(
						instID, buyNum,
						account.WithOrderPrice(
							framework.Contract().GetContract(instID).CalcSlipPrice(
								closePrice, s.slip, config.OrderBuy,
							),
						),
					)
					stockInfo.fa = stockInfo.fa - tmpOrder.MarketValue() - tmpOrder.Commission()
					// fmt.Println(tmpOrder.Commission())

				}

			} else if nowSignal2 < 0 && lastSignal2 > 0 {
				if pos, ok := s.acc.GetPositionByInstID(instID); ok {
					vol := pos.Volume(account.WithSellAvailable(true))
					if vol > 0 {
						s.sell(
							instID, vol,
							indicate.ConvertToFloat("Close"),
						)

						tmpOrder, _ := s.acc.NewOrder(
							instID, vol,
							account.WithOrderPrice(indicate.ConvertToFloat("Close")),
						)
						stockInfo.fa = stockInfo.fa + tmpOrder.MarketValue() - tmpOrder.Commission()
					}
				}
			}

			// stockInfo.triggerState = 2
			// }

			stockInfo.indi = GEPSlice
			stockInfo.lastTradeSignal1 = nowSignal1
			stockInfo.lastTradeSignal2 = nowSignal2
		}

	}

	simAllFundAvail := 0.0
	// iter the map and sum all the fa
	for _, v := range s.stockInfos {
		simAllFundAvail += v.fa
	}
	tmpFA := s.acc.GetAsset().Available
	// fmt.Println(simAllFundAvail, tmpFA)
	// iter the map again and change the fa

	for _, v := range s.stockInfos {
		// fmt.Println("v.fa and 理论分配值分别为：",v.fa, v.fa/ simAllFundAvail * tmpFA)
		v.fa = math.Min(v.fa, v.fa/simAllFundAvail*tmpFA)
		// fmt.Println("新的fa是： ",v.fa)
	}
	if len(s.SInstNames) != 0 {
		for _, instID := range s.SInstNames {
			stockInfo := s.stockInfos[instID]
			dealingFunc(instID, stockInfo)
		}
	}
	return
}

func (s *DMTGS) OnDailyClose(framework handler.Framework, acc map[string]account.Account) {
	// TODO  获取当日有复权的标的
	// TODO  重新记录信号值
	// if len(s.SInstNames) != 0 {
	// 	for _, instID := range s.SInstNames {
	// 		stockInfo := s.stockInfos[instID]
	// 		if len(stockInfo.indi) == 0 {
	// 			continue
	// 		}
	//
	// 		config.InfoF(
	// 			"tm: %s, instID: %s, fa: %f, signal: %f, indi: %v", framework.CurrDate().Format(config.TimeFormatDate2),
	// 			instID,
	// 			stockInfo.fa,
	// 			stockInfo.lastTradeSignal,
	// 			stockInfo.indi,
	// 		)
	// 	}
	// }
}
