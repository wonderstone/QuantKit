package strategy

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type SimpleStrategy struct {
	handler.EmptyStrategy
	acc account.Account

	SInstNames []string // 股票标的名称
	SIndiNames []string // 股票参与GEP指标名称，注意其数量不大于BarDE内信息数量，且strategy内可见BarDE的数据
	SPosNum    float64  // 股票标的持仓数量

	FInstNames []string // 期货标的名称
	FIndiNames []string // 期货参与GEP指标名称， there should be a rollover field in the futures IndiDataNames slice
	FPosNum    float64  // 期货标的持仓数量
}

func (s *SimpleStrategy) OnInitialize(framework handler.Framework) {
	s.acc = framework.Account().GetAccountByID(config.AccountTypeStockSimple)
	s.SInstNames = framework.Config().Framework.Instrument
	s.SIndiNames = framework.Config().Framework.Indicator

	s.SPosNum = 800
}

func (s *SimpleStrategy) buy(instID string, qty, price float64) bool {
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

func (s *SimpleStrategy) sell(instID string, qty, price float64) bool {
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

func (s *SimpleStrategy) OnTick(
	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []handler.Order) {
	// 判断股票标的切片SInstrNames是否为空，如果为空，则不操作股票数据循环
	if len(s.SInstNames) != 0 {
		for _, instID := range s.SInstNames {
			if indi, ok := indicators.Get(instID); ok {
				if !ContainNaN(indi) {
					closePrice := indi.ConvertToFloat("Close")

					// % GEP 引入
					var GEPSlice = make(model.InputValues, len(s.SIndiNames))
					for i := 0; i < len(s.SIndiNames); i++ {
						GEPSlice[i] = indi.ConvertToFloat(s.SIndiNames[i])
					}

					tradeSignal := framework.Evaluate(GEPSlice)

					if tradeSignal[0] >= 0 {
						if pos, ok := s.acc.GetPosition()[instID]; ok {
							if pos.Volume() == 0 {
								s.buy(instID, s.SPosNum, closePrice)
							}
						} else {
							s.buy(instID, s.SPosNum, closePrice)
						}
					} else {
						if pos, ok := s.acc.GetPosition()[instID]; ok {
							if pos.Volume(account.WithSellAvailable(true)) != 0 {
								s.sell(instID, pos.Volume(), closePrice)
							}
						} else {
							// do nothing
						}
					}
				}
			}
		}
	}

	return
}
