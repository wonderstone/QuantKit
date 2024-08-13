package strategy

import (
	// "fmt"

	"math"
	"sort"
	"strconv"
	"strings"

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

// ! WHY ADD FOLLOWING TWO VARIABLES HERE GONNA MESS UP?
// ! var TM time.Time
// ! var Indicators *orderedmap.OrderedMap[string, dataframe.StreamingRecord]

// + Base with Open price plus GEP Condition
// // todo: 1. base info read into the strategy when initialized
// // todo: 2. indicator has only basic price info and consts
// // todo: 3. use ondailyopen or ondailyclose to trigger the trade
// // todo: 4. use ontick to get the indicators and save then in the strategy
// todo: 5. use the saved indicators and VS base info to trigger the trade
// // todo: 6. reduce data required from 3 to 2
// todo: 7. GEPSlice needs to actually have 2 parts
// todo: 8. frequency cannot be daily anymore for replay/nextmode.go 208 line
type SortRankVS struct {
	instID string  // 合约ID
	close  float64 // 收盘价
	signal float64 // 信号
	vscond	bool	// 是否满足VS条件
	// posAmt float64 // 持仓金额
}



type VSS struct {
	handler.EmptyStrategy // 继承EmptyStrategy类，可以减少实现不想处理的方法

	acc account.Account // 账户

	sInstrument    []string   // 股票标的名称
	sIndicators    []string   // 股票参与GEP指标名称
	tmpTargets     []SortRankVS // 临时目标集合
	vsTargets      []SortRankVS
	notTODay       map[string]void // 是否为起始日集合
	vsHoldNumR     int
	vsHoldNumM     int
	fMaxMinRatio   float64 // 等差数列差值
	fCashUsedRatio float64 // 资金利用率
	iBuyNum        int     // 买入数量
	// vsBuyNum          int       // vs买入数量
	// sBuyAmtRatioSlice []float64 // 买入金额比例切片
	// fInitMoney        float64   // 初始资金

	lvMode   string // 是否为LV模式 last value mode will have more frequency
	bGEPMode bool   // 是否为GEP模式
	bLogOn   bool   // 是否打印日志

	// todo: VS base info and indicators saved in the strategy
	// CSVData    map[string]map[string]map[string]string
	indicators *orderedmap.OrderedMap[string, dataframe.StreamingRecord]
	tmpDate    string

	TM         time.Time
	parmp      map[string]string
	lastvalMap map[string]map[string]string // map[instr]map[indi]lastval
}

func (s *VSS) OnInitialize(framework handler.Framework) {

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
	s.lvMode = config.GetParam(c, "lv_mode", "LV")
	vspara := config.GetParam(c, "hold_num_para", "20,5")
	// sepereate the string vspara with ","
	vsparaSlice := strings.Split(vspara, ",")
	tmpR, err := strconv.Atoi(vsparaSlice[0])
	if err != nil {
		panic(err)
	}
	s.vsHoldNumR = tmpR
	tmpM, err := strconv.Atoi(vsparaSlice[1])
	if err != nil {
		panic(err)
	}
	s.vsHoldNumM = tmpM
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
	s.parmp = make(map[string]string)
	s.parmp["bp"] = "b2_factor_value"
	s.parmp["np"] = "b3_factor_value"
	s.parmp["div"] = "b4_factor_value"

	s.lastvalMap = make(map[string]map[string]string)
	// init the indicators
	s.indicators = orderedmap.New[string, dataframe.StreamingRecord]()

}

func (s *VSS) OnDailyClose(framework handler.Framework, acc map[string]account.Account) {
	// config.DebugF("%s 盘前时间", framework.CurrTime())

	// reset s.tmpTargets
	s.tmpTargets = make([]SortRankVS, 0)

	// // check if the date is 2021-04-30

	signals := make([]SortRankVS, 0)
	// & Step 0.1 : 更新集合信息
	closes := make(map[string]float64) // 用于记录股票标的名称和close价
	for pIndicate := s.indicators.Oldest(); pIndicate != nil; pIndicate = pIndicate.Next() {
		indicate := pIndicate.Value
		//
		if s.lastvalMap[pIndicate.Key] == nil {
			// has three keys bp, np, div, value is ""
			s.lastvalMap[pIndicate.Key] = map[string]string{"bp": "", "np": "", "div": ""}

		}
		//
		closePrice := indicate.ConvertToFloat("Close")
		closes[pIndicate.Key] = closePrice
		// check if pIndicate.Key is in s.notTODay
		_, notT0 := s.notTODay[pIndicate.Key]
		notNaN := !ContainNaN(indicate)
		tmpbp := s.GetCSVData(s.tmpDate, pIndicate.Key, s.parmp["bp"], s.lastvalMap[pIndicate.Key], "bp")
		bp, errbp := TryConvertToFloat(tmpbp)
		tmpnp := s.GetCSVData(s.tmpDate, pIndicate.Key, s.parmp["np"], s.lastvalMap[pIndicate.Key], "np")
		np, errnp := TryConvertToFloat(tmpnp)
		tmpdiv := s.GetCSVData(s.tmpDate, pIndicate.Key, s.parmp["div"], s.lastvalMap[pIndicate.Key], "div")
		div, errdiv := TryConvertToFloat(tmpdiv)
		
		
		
		tmpbpv:= s.GetCSVDataVal(s.tmpDate, pIndicate.Key, s.parmp["bp"], s.lastvalMap[pIndicate.Key], "bp")
		bpv , _ := TryConvertToFloat(tmpbpv)
		tmpnpv:= s.GetCSVDataVal(s.tmpDate, pIndicate.Key, s.parmp["np"], s.lastvalMap[pIndicate.Key], "np")
		npv , _ := TryConvertToFloat(tmpnpv)
		tmmpdivv:= s.GetCSVDataVal(s.tmpDate, pIndicate.Key, s.parmp["div"], s.lastvalMap[pIndicate.Key], "div")
		divv , _ := TryConvertToFloat(tmmpdivv)
		
		tmpslice := []float64{bpv, npv, divv}
		// fmt.Println(tmpslice)
		// 如果数据不包含NaN值(有效)，且不是初始天，添加到signals集合
		if notNaN && notT0 {
			// fmt.Println(tm)
			tradeSignal := make(model.OutputValues, 0)

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
				GEPSlice = append(GEPSlice, tmpslice...)
				tradeSignal = framework.Evaluate(GEPSlice)

				if math.IsNaN(tradeSignal[0]) || math.IsInf(tradeSignal[0], 0) {
					continue
				}
				// fmt.Println(tradeSignal[0],condall)
				if condall {
					signals = append(signals, SortRankVS{instID: pIndicate.Key, signal: tradeSignal[0], close: closePrice, vscond: true})
				} else {
					signals = append(signals, SortRankVS{instID: pIndicate.Key, signal: 0.0, close: closePrice, vscond: false})
				}

			} else {
				if condall {
					signals = append(signals, SortRankVS{instID: pIndicate.Key, signal: 1.0, close: closePrice, vscond: true})
				} else {
					signals = append(signals, SortRankVS{instID: pIndicate.Key, signal: 0.0, close: closePrice, vscond: false})
				}

			}
		}

		if notNaN && !notT0 {
			s.notTODay[pIndicate.Key] = void{}
		}

	}
	// fmt.Println(framework.CurrDate(),framework.CurrTime())

	if s.bGEPMode {
		if len(signals) != 0 {
			tmp := filterSignalVS(signals, func(sr SortRankVS) bool {
				// if sr.signal is not equal to NaN
				return sr.vscond
			})
			if len(tmp) != 0 {
				s.vsTargets = tmp
				// fmt.Println(framework.CurrDate(),framework.CurrTime())

			}
			// s.vsTargets = tmp
		}

		if len(s.vsTargets) > 1 {
			sort.Slice(
				s.vsTargets, func(i, j int) bool {
					return s.vsTargets[i].signal < s.vsTargets[j].signal
				},
			)
		}

		// set the tmpTargets, the length of tmpTargets is vsHoldNumR percent of the length of vsTargets
		// and should be more than vsHoldNumM, or take them all
		if len(s.vsTargets) != 0 {
			if s.vsHoldNumM > len(s.vsTargets)*s.vsHoldNumR/100 {
				s.tmpTargets = s.vsTargets
			} else {
				tmp := len(s.vsTargets) * s.vsHoldNumR / 100
				// keep the last tmp elements
				s.tmpTargets = s.vsTargets[len(s.vsTargets)-tmp:]
			}
		}
	} else {
		if len(signals) != 0 {
			s.tmpTargets = filterSignalVS(signals, func(sr SortRankVS) bool {
				return sr.signal > 0.0
			})
		}
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

func (s *VSS) OnTick(
	// * 注意一下 这个函数返回在这里没什么用 在这里并没有尊重框架的设计
	// * 后续所有的买卖实际上都是通过buy、sell方法来给策略结构体内的虚拟账户进行插入订单的操作
	// * 而插入订单这个接口在StockSimple结构体中可以去同步实现链接交易功能 around line565

	// - 这里不输出，所以replay->nextmode around line 212 实际上被架空了
	// - 从一开始的设计来看，外层存在接收订单的接口实现然后对接柜台，这是最开始的想法
	// + 爱怎么用就怎么用吧，其实框架也只是辅助，不是必须
	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []account.Order) {
	// 判断股票标的切片SInstrNames是否为空，如果为空，则不操作股票数据循环
	s.tmpDate = tm.Format("2006.01.02")

	if s.TM != tm {
		s.TM = tm
		// deep copy the indicators
		tmp := orderedmap.New[string, dataframe.StreamingRecord]()
		for pIndicate := indicators.Oldest(); pIndicate != nil; pIndicate = pIndicate.Next() {

			tmpSR := dataframe.StreamingRecord{}
			// copy the data to tmpSR.Data
			tmpSR.Data = append(tmpSR.Data, pIndicate.Value.Data...)
			// iter pIndicate.Value.Headers Map and copy the data to tmpSR.Headers

			tmpSR.Headers = make(map[string]int)
			for key, value := range pIndicate.Value.Headers {
				tmpSR.Headers[key] = value
			}
			tmp.Store(pIndicate.Key, tmpSR)
		}
		// Indicators = tmp
		s.indicators = tmp
	}

	// // check if the tm is 2021-06-18
	// if tm.Format("2006-01-02") == "2021-06-18" {
	// 	fmt.Println("2021-06-18")
	// }

	return
}

func (s *VSS) buy(instID string, qty, price float64) bool {
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

func (s *VSS) sell(instID string, qty, price float64) bool {
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

// func to give the value of the CSVData
func (s *VSS) GetCSVData(tmStr, InstID, IndiName string, lastvalmp map[string]string, key string) string {
	if s.lvMode == "LV" {
		// this is the last value mode
		// compare the value with the last value
		// if the value is equal to the last value return the ""
		// if the value is not equal to the last value return the value
		if v, ok := CSVData[tmStr]; ok {
			if v1, ok1 := v[ModifyInstID(InstID)]; ok1 {
				if v2, ok2 := v1[IndiName]; ok2 {
					if lastvalmp[key] == v2 {
						return ""
					}
					lastvalmp[key] = v2
					return v2
				}
			}
			// if the value is not in the map return the last value
		} else {
			return ""
		}
		// if the value is not in the map return the last value
		return ""

	} else {
		// this is the normal mode
		if v, ok := CSVData[tmStr]; ok {
			if v1, ok1 := v[ModifyInstID(InstID)]; ok1 {
				if v2, ok2 := v1[IndiName]; ok2 {
					lastvalmp[key] = v2
					return v2
				}
			}
			// if the value is not in the map return the last value
		} else {
			return ""
		}
		return ""
	}
}


func (s *VSS) GetCSVDataVal(tmStr, InstID, IndiName string, lastvalmp map[string]string, key string) string{

	if v, ok := CSVData[tmStr]; ok {
		if v1, ok1 := v[ModifyInstID(InstID)]; ok1 {
			if v2, ok2 := v1[IndiName]; ok2 {
				return v2
			}
		}
		// if the value is not in the map return the last value
	} else {
		return ""
	}
	return ""
}

func init() {
	srce := "/Users/wonderstone/Desktop/QuantKit/EX-VS-dir/download/1dayfactor/to_gep.csv"
	InstColName := "data_code"
	DateColName := "end_date"
	CSVData = ProcCsvBase(srce, InstColName, DateColName)
}
