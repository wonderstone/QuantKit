package strategy

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/recorder"
	// "github.com/wonderstone/QuantKit/tools/times"
)

type CQStockInfo struct {
	fa          float64   // 各支股票对应可用资金  fund available
	lastVSignal []float64 // 上一次可变指标信号数值
	lastFSignal []float64 // 上一次固定指标信号数值
	indi        model.InputValues
}

type CQOPinfo struct {
	DT              time.Time `csv:"time"`
	InstID          string    `csv:"instID"`
	MA              float64   `csv:"MA"`
	MAR             float64   `csv:"MAR"`
	MACD            float64   `csv:"MACD"`
	MACDR           float64   `csv:"MACDR"`
	Close1d         float64   `csv:"Close1d"`
	Close1dR        float64   `csv:"Close1dR"`
	Close1w         float64   `csv:"Close1w"`
	Close1wR        float64   `csv:"Close1wR"`
	MAClose1dR3     float64   `csv:"MAClose1dR3"`
	MAClose1dR3R1   float64   `csv:"MAClose1dR3R1"`
	MACDClose1dR3   float64   `csv:"MACDClose1dR3"`
	MACDClose1dR3R1 float64   `csv:"MACDClose1dR3R1"`
	MAClose1wR3     float64   `csv:"MAClose1wR3"`
	MAClose1wR3R1   float64   `csv:"MAClose1wR3R1"`
	MACDClose1wR3   float64   `csv:"MACDClose1wR3"`
	MACDClose1wR3R1 float64   `csv:"MACDClose1wR3R1"`
	State           string    `csv:"state"`
}

type CQ struct {
	handler.EmptyStrategy
	acc       account.Account
	initMoney float64 // 初始资金

	InstNames     []string // 股票标的名称
	IndiNames     []string // 股票参与GEP指标名称
	minCashRatio  float64
	timeintervals []string

	// HoldAmt float64 // ~ 股票标的持仓金额

	stockInfos map[string]*CQStockInfo // 股票标的持仓信息(内部)
	cqMap      map[string]float64      // 用户自定义持仓规则map表

	logOn           bool // 是否打印日志
	GEPMode         bool // 是否为GEP模式
	IndicatorOutput bool // 是否输出指标
	slip            float64

	tradeSignals []TradeSignal
	recorder     *recorder.CsvRecorder
}

func (s *CQ) OnGlobalOnce(global handler.Global) {
	global.SetGEPInputParams(s.IndiNames)
}

func (s *CQ) OnInitialize(framework handler.Framework) {
	s.acc = framework.Account().GetAccountByID(config.AccountTypeStockSimple)
	s.stockInfos = make(map[string]*CQStockInfo)
	// 获取配置
	c := framework.ConfigStrategyFromFile()
	// 获取股票标的名称
	s.InstNames = config.GetParam(c, "instruments", framework.Config().Framework.Instrument)
	// 获取股票参与GEP指标名称
	s.IndiNames = config.GetParam(c, "indicators", framework.Config().Framework.Indicator)
	// s.SIndiNames = framework.Config().Framework.Indicator

	// 获取最小资金比例
	s.minCashRatio = config.GetParam(c, "mincashratio", 0.01)
	// 获取时间间隔
	s.timeintervals = config.GetParam(c, "timeintervals", []string{"1W", "1D", "15"})

	// 获取用户自定义持仓规则
	s.cqMap = readCSV(framework.Base().Dir().Root + "/状态持仓.csv")

	// 获取初始资金
	s.initMoney = s.acc.GetAsset().Initial

	// init the faMap
	average, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", s.initMoney/float64(len(s.InstNames))), 64)
	for _, name := range s.InstNames {
		s.stockInfos[name] = new(CQStockInfo)
		// 账户初始资金充当市值情况
		s.stockInfos[name].fa = average
		for i := 0; i < len(s.timeintervals); i++ {
			s.stockInfos[name].lastVSignal = append(s.stockInfos[name].lastVSignal, 0)
			s.stockInfos[name].lastFSignal = append(s.stockInfos[name].lastFSignal, 0)
		}
		// s.stockInfos[name].lastVSignal = -1
		// ! 日内是否需要设置触发时间
		// s.stockInfos[name].triggerState = 0 // 目前触发时间是标准参数，不需要单独触发状态
	}
	// ! change!
	s.slip = framework.Framework().Config().Framework.Stock.Slippage
	s.logOn = false
	s.GEPMode = false
	s.IndicatorOutput = true

	if s.IndicatorOutput {
		opt := []recorder.WithOption{
			recorder.WithFilePath("test01.csv"),
			recorder.WithPlusMode(),
			recorder.WithTransaction(),
		}

		s.recorder = recorder.NewCsvRecorder(opt...)

		go func() {
			err := s.recorder.RecordChan()
			if err != nil {
				panic(err)
			}
		}()

	} // // ! set csv recorder for indicators

}

// ContainNaN 此处是为了停盘数据处理设定的规则相检查用的
func (s *CQ) ContainNaN(m dataframe.StreamingRecord) bool {
	for _, x := range m.Data {
		if len(x) == 0 {
			return true
		}
	}
	return false
}

func (s *CQ) OnDailyOpen(framework handler.Framework, marketType config.MarketType, acc ...account.Account) {
	// date := framework.CurrDate()
	// 下隔夜单
}

func (s *CQ) buy(instID string, qty, price float64) bool {
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

func (s *CQ) sell(instID string, qty, price float64) bool {
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

func (s *CQ) OnTick(
	framework handler.Framework, tm time.Time, indicates orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {
	dealingFunc := func(
		instID string, stockInfo *CQStockInfo, record ...*dataframe.StreamingRecord,
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
			var GEPSlice = make(model.InputValues, len(s.IndiNames))

			if s.GEPMode {
				// @ partial GEP import
				for i, v := range s.timeintervals {
					// if i is equal to the last index of s.timeintervals
					if i != len(s.timeintervals)-1 {
						tradeSignal = append(
							tradeSignal, indicate.ConvertToFloat("MA3_"+v), indicate.ConvertToFloat("MA5_"+v),
						)
					} else {
						for i := 0; i < len(s.IndiNames); i++ {
							GEPSlice[i] = indicate.ConvertToFloat(s.IndiNames[i])
						}

						tmp := framework.Evaluate(GEPSlice)
						tradeSignal = append(
							// @ partial GEP
							// todo replace MA5 with MACD
							tradeSignal, tmp...,
						)
						tradeSignal = append(
							// @ partial GEP
							// todo replace MA5 with MACD
							tradeSignal, indicate.ConvertToFloat("MA5_"+v),
						)
					}

				}

			} else {
				// todo todo replace MA5 with MACD
				for _, _ = range s.timeintervals {
					// fmt.Println("MA" + v)
					tradeSignal = append(
						tradeSignal, indicate.ConvertToFloat("MA"), indicate.ConvertToFloat("MAR"),
					)
				}
			}

			for _, v := range tradeSignal {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return
				}
			}

			// 根据规则 确定状态
			// 遍历s.timeintervals 得到各个状态
			// nowVSignal := []float64{}
			// nowFSignal := []float64{}
			// for i, _ := range s.timeintervals {
				// nowVSignal = append(nowVSignal, tradeSignal[2*i])
				// nowFSignal = append(nowFSignal, tradeSignal[2*i+1])
			// }
			// 如果tm是2017-06-26
			// if tm.Format("2006-01-02") == "2022-04-08" {
			// 	fmt.Println(tm)
			// }
			// 获得state  应该是能合并
			state := ""
			// 日内分钟级别状态
			ma := indicate.ConvertToFloat("MA")
			mar := indicate.ConvertToFloat("MAR")
			macd := indicate.ConvertToFloat("MACD")
			macdr := indicate.ConvertToFloat("MACDR")
			tmpst := statelogic(ma, mar, macd, macdr)
			state += tmpst
			// 日级别状态
			ma1d := indicate.ConvertToFloat("MAClose1dR3")
			mar1d := indicate.ConvertToFloat("MAClose1dR3R1")
			macd1d := indicate.ConvertToFloat("MACDClose1dR3")
			macdr1d := indicate.ConvertToFloat("MACDClose1dR3R1")
			tmpst = statelogic(ma1d, mar1d, macd1d, macdr1d)
			state += tmpst
			// 周级别状态
			ma1w := indicate.ConvertToFloat("MAClose1wR3")
			mar1w := indicate.ConvertToFloat("MAClose1wR3R1")
			macd1w := indicate.ConvertToFloat("MACDClose1wR3")
			macdr1w := indicate.ConvertToFloat("MACDClose1wR3R1")
			tmpst = statelogic(ma1w, mar1w, macd1w, macdr1w)
			state += tmpst
			// ! 用于记录指标运算信息
			if s.IndicatorOutput {
				tmp := CQOPinfo{
					DT:              tm,
					InstID:          instID,
					MA:              indicate.ConvertToFloat("MA"),
					MAR:             indicate.ConvertToFloat("MAR"),
					MACD:            indicate.ConvertToFloat("MACD"),
					MACDR:           indicate.ConvertToFloat("MACDR"),
					Close1d:         indicate.ConvertToFloat("Close1d"),
					Close1dR:        indicate.ConvertToFloat("Close1dR"),
					Close1w:         indicate.ConvertToFloat("Close1w"),
					Close1wR:        indicate.ConvertToFloat("Close1wR"),
					MAClose1dR3:     indicate.ConvertToFloat("MAClose1dR3"),
					MAClose1dR3R1:   indicate.ConvertToFloat("MAClose1dR3R1"),
					MACDClose1dR3:   indicate.ConvertToFloat("MACDClose1dR3"),
					MACDClose1dR3R1: indicate.ConvertToFloat("MACDClose1dR3R1"),
					MAClose1wR3:     indicate.ConvertToFloat("MAClose1wR3"),
					MAClose1wR3R1:   indicate.ConvertToFloat("MAClose1wR3R1"),
					MACDClose1wR3:   indicate.ConvertToFloat("MACDClose1wR3"),
					MACDClose1wR3R1: indicate.ConvertToFloat("MACDClose1wR3R1"),
					State:           state,
				}
				s.recorder.GetChannel() <- &tmp
			}
			// ! 用于记录指标运算信息结束

			// 获得positionlevel
			positionlevel, e := s.cqMap[state]
			if !e {
				// fmt.Println("no state : ", state)
				positionlevel = 0.5
			}
			// fmt.Println(positionlevel)
			// 根据positionlevel确定是否买入
			// 买入情况：positionlevel 大于 该标的当前持仓市值比例
			pos, ok := s.acc.GetPositionByInstID(instID)
			ratio := 0.0
			if ok {
				ratio = pos.Amt() / (pos.Amt() + stockInfo.fa)
			}
			if positionlevel > ratio {
				// 买入逻辑
				closePrice := indicate.ConvertToFloat("Close")
				deltamt := 0.0
				if ok {
					deltamt = (pos.Amt() + stockInfo.fa) * (positionlevel - ratio)
				} else {
					deltamt = stockInfo.fa * positionlevel
				}

				buyNum := framework.Contract().GetContract(instID).CalcMaxQty(deltamt / closePrice)
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
				}

			} else if positionlevel < ratio {
				// 卖出逻辑
				closePrice := indicate.ConvertToFloat("Close")
				deltamt := (pos.Amt() + stockInfo.fa) * (ratio - positionlevel)

				sellnum := framework.Contract().GetContract(instID).CalcMaxQty(deltamt / closePrice)

				if sellnum != 0 {
					stockavailable := pos.Volume(account.WithSellAvailable(true))
					if stockavailable == 0 {
						return
					}
					if stockavailable >= sellnum {
						// 卖掉sellnum
						s.sell(instID, sellnum, indicate.ConvertToFloat("Close"))
						tmpOrder, _ := s.acc.NewOrder(
							instID, sellnum,
							account.WithOrderPrice(indicate.ConvertToFloat("Close")),
						)
						stockInfo.fa = stockInfo.fa + tmpOrder.MarketValue() - tmpOrder.Commission()
					} else if stockavailable < sellnum {
						// 卖掉stockavailable
						s.sell(instID, stockavailable, indicate.ConvertToFloat("Close"))
						tmpOrder, _ := s.acc.NewOrder(
							instID, stockavailable,
							account.WithOrderPrice(indicate.ConvertToFloat("Close")),
						)
						stockInfo.fa = stockInfo.fa + tmpOrder.MarketValue() - tmpOrder.Commission()
					}
				}
			}

			stockInfo.indi = GEPSlice
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
		v.fa =  v.fa/simAllFundAvail*tmpFA
		// fmt.Println("新的fa是： ",v.fa)
	}
	if len(s.InstNames) != 0 {
		for _, instID := range s.InstNames {
			stockInfo := s.stockInfos[instID]
			dealingFunc(instID, stockInfo)
		}
	}
	return
}

func (s *CQ) OnEnd(framework handler.Framework) {
	close(s.recorder.GetChannel())
}

func readCSV(filename string) map[string]float64 {
	// open the file
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	// read the csv file
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println(err)
	}

	// convert the csv file to map whose key is the first 4 record elements and value is the last record element
	m := make(map[string]float64)
	for _, record := range records[1:] {
		l := len(record)
		k := ""
		for i := 0; i < l-1; i++ {
			k += record[i]
		}
		m[k], _ = strconv.ParseFloat(record[l-1], 64)
	}

	return m
}

func statelogic(ma, mar, macd, macdr float64) (res string) {
	// 如果ma>mar 且 macd > 0 && macd>macdr  为0
	// 如果ma>mar 且 macd > 0 && macd<macdr  为1
	// 如果ma>mar 且 macd < 0 && macd>macdr  为2
	// 如果ma>mar 且 macd < 0 && macd<macdr  为3
	// 如果ma<mar 且 macd > 0 && macd>macdr  为4
	// 如果ma<mar 且 macd > 0 && macd<macdr  为5
	// 如果ma<mar 且 macd < 0 && macd>macdr  为6
	// 如果ma<mar 且 macd < 0 && macd<macdr  为7
	if ma > mar {
		if macd > 0 {
			if macd > macdr {
				res = "0"
			} else {
				res = "1"
			}
		} else {
			if macd > macdr {
				res = "2"
			} else {
				res = "3"
			}
		}
	} else {
		if macd > 0 {
			if macd > macdr {
				res = "4"
			} else {
				res = "5"
			}
		} else {
			if macd > macdr {
				res = "6"
			} else {
				res = "7"
			}
		}
	}
	if res == "" {
		res = "8"
	}
	return res
}
