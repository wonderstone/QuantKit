package position

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"

	"github.com/wonderstone/QuantKit/tools/math"
	"github.com/wonderstone/QuantKit/tools/recorder"
	// "github.com/wonderstone/QuantKit/tools/recorder"
)

// ~ totally different from the original code which is based on
// ~ the queue of positions and use settle to do the T+1 operation
// ~ for stock, it is faster

type StockValue struct {
	openPrice float64     // 开仓价格
	lastPrice float64     // 最新价格
	volume    float64     // 持仓数量
	available float64     // 可用数量
	costAmt   float64     // 持仓成本金额
	pnl       account.PnL // 盈亏
}

// 同一标的下的持仓合并
func (s *StockValue) Add(r StockValue, price float64) float64 {
	pnl := s.calcPnL(price) + r.calcPnL(price)
	if s.volume == 0 && r.volume == 0 {
		return 0
	}
	s.costAmt += r.costAmt
	s.volume += r.volume
	s.openPrice = s.costAmt / s.volume
	s.available = s.volume
	s.lastPrice = price
	s.pnl = s.pnl.Add(r.pnl)

	return pnl
}

// ? 修正持仓盈亏（除复权场景才会使用）
// ! no check, believe in Master X
func (s *StockValue) repair() float64 {
	if s.volume == 0 {
		pnl := -s.pnl.Profit
		s.pnl.Profit = 0
		return pnl
	}

	pnl := (s.lastPrice - s.openPrice) * s.volume
	delta := pnl - s.pnl.Profit
	s.pnl.Profit = pnl

	return delta
}

func (s *StockValue) AddTrade(qty, price float64) (deltaDynamicPnL float64, deltaActualPnL float64) {
	matchVol := s.volume + qty

	if matchVol == 0 {
		deltaDynamicPnL = -s.pnl.Profit
		deltaActualPnL = -qty*price - s.costAmt
		s.costAmt = 0
		s.openPrice = 0
		s.lastPrice = 0
		s.volume = 0
		s.available = 0
		s.pnl.Profit = 0
	} else if qty > 0 {
		costAmtNow := s.costAmt + qty*price
		pnlNow := matchVol*price - costAmtNow

		// 浮动盈亏增量 = (最新价格 - 上一次价格) * 持仓数量
		deltaDynamicPnL = matchVol * (price - s.lastPrice)
		deltaActualPnL = 0

		s.volume = matchVol
		s.costAmt = costAmtNow
		s.openPrice = math.Round(s.costAmt/matchVol, 6)
		s.lastPrice = price
		s.pnl.Profit = pnlNow
	} else {
		costAmtNow := s.openPrice * matchVol
		pnlNow := matchVol * (price - s.openPrice)

		deltaActualPnL = -qty*price + s.costAmt
		deltaDynamicPnL = matchVol*(price-s.lastPrice) - deltaActualPnL

		s.costAmt = costAmtNow
		s.volume = matchVol
		s.lastPrice = price
		s.pnl.Profit = pnlNow
		s.available = matchVol
	}

	return
}

// 收盘结算
func (s *StockValue) AddSettle(qty, price float64) {
	s.costAmt += price * qty
	s.volume += qty
	s.openPrice = s.costAmt / s.volume
	s.lastPrice = price
	s.pnl.Profit += (price - s.openPrice) * qty
}

func (s *StockValue) addSellOrder(qty float64) {
	if s.available+qty < 0 {
		config.WarnF("股票持仓可用不足, 当前: %f, 需要: %f", s.volume, -qty)
	}

	s.available -= qty
}

func (s *StockValue) calcPnL(price float64) float64 {
	if s.volume == 0 {
		return 0
	}

	deltaPnL := (price - s.lastPrice) * s.volume
	s.lastPrice = price
	s.pnl.Profit = price*s.volume - s.costAmt

	return deltaPnL
}

// + StockPosition
// + 实现 entity.Position 接口
// + 实现 handler.Position 接口

type StockPosition struct {
	contract  contract.Contract // 合约
	account   handler.Account2
	openTime  time.Time                // 开仓时间
	direction config.PositionDirection // 持仓方向
	lastPrice float64
	valToday  StockValue // 今仓
	valHis    StockValue // 历史仓
}

// & StockPosition 实现 entity Position 接口

// InstID() string
func (s *StockPosition) InstID() string {
	return s.contract.GetInstID()
}

// DefaultDirection() config.PositionDirection
// DefaultDirection 默认的持仓方向, 一般情况下，股票只有多仓
func (s *StockPosition) DefaultDirection() config.PositionDirection {
	return config.PositionLong
}

// AvailableDirection() []config.PositionDirection
// AvailableDirection 可用的持仓方向, 一般情况下，股票只有多仓
func (s *StockPosition) AvailableDirection() []config.PositionDirection {
	return []config.PositionDirection{config.PositionLong}
}

// OpenPrice(opts ...WithOpFilterPos) float64
// OpenPrice 开仓价格 这个价格会随着持仓的变化而变化
func (s *StockPosition) OpenPrice(opts ...account.WithOpFilterPos) float64 {
	return (s.valToday.openPrice*s.valToday.volume + s.valHis.openPrice*s.valHis.volume) / (s.valToday.volume + s.valHis.volume)
}

// OpenTime(opts ...WithOpFilterPos) time.Time
func (s *StockPosition) OpenTime(opts ...account.WithOpFilterPos) time.Time {
	return s.openTime
}

// Volume(opts ...WithOpFilterPos) float64
// 竟然还有筛选是否可卖出的参数
func (s *StockPosition) Volume(opts ...account.WithOpFilterPos) float64 {
	opt := newStockOpt(opts...)

	if opt.SellAvailable {
		return s.valHis.available
	}

	return s.valToday.volume + s.valHis.volume
}

// Amt(opts ...WithOpFilterPos) float64
func (s *StockPosition) Amt(opts ...account.WithOpFilterPos) float64 {
	opt := newStockOpt(opts...)

	if opt.SellAvailable {
		return s.valHis.available * s.valHis.lastPrice
	}

	return s.valToday.volume*s.valToday.lastPrice + s.valHis.volume*s.valHis.lastPrice
}

// PnL(opts ...WithOpFilterPos) PnL // 持仓盈亏
func (s *StockPosition) PnL(opts ...account.WithOpFilterPos) account.PnL {
	return s.valToday.pnl.Add(s.valHis.pnl)
}

// & StockPosition 实现 entity Position 接口 完毕

// & StockPosition 实现 handler Position 接口

func (s *StockPosition) Init(
	account2 handler.Account2,
	contract contract.Contract,
	qty, price float64,
	direction ...config.PositionDirection,
) {
	s.contract = contract
	s.account = account2
	s.lastPrice = price
	s.direction = direction[0]
	s.valToday = StockValue{
		openPrice: price,
		volume:    qty,
		costAmt:   price * qty,
		lastPrice: price,
		pnl:       account.PnL{},
	}

	s.valHis = StockValue{}
}

// ! 这里有问题
func (s *StockPosition) DoSettle(tm time.Time, price float64, recorder recorder.Handler) bool {
	// 今仓结算
	if price == 0 {
		// config.WarnF("股票结算价格不能为0, instID: %s, p: %f\n", s.contract.GetInstID(), s.valToday.lastPrice)
		price = s.lastPrice
	}

	if s.valToday.volume == 0 && s.valHis.volume == 0 {
		return false
	}

	// !请考虑账户盈亏 这个不用position做，由account来做
	pnl := s.valHis.Add(s.valToday, price)
	s.account.AddDynamicPnL(pnl)

	// 今仓清零
	s.valToday = StockValue{}

	// !!!!!! 除权除息
	// 暂时没有处理股权登记时间
	// 分红(税前)，目前暂时不计红利税
	// 税单独记录
	// 换股则清空历史持仓，且不计算盈亏
	dividend := 0.0
	tax := 0.0
	if s.account.Base() != nil {
		s.valHis.volume, s.valHis.openPrice, s.valHis.lastPrice, dividend, tax = s.account.CalcSettleInfo(
			s.contract.GetInstID(), s.valHis.volume, s.valHis.openPrice, s.valHis.lastPrice,
		)

		pnl = s.valHis.repair()
		s.account.AddDynamicPnL(pnl)
	}

	if s.valHis.volume == 0 {
		return true
	}

	s.lastPrice = s.valHis.lastPrice

	s.valHis.pnl.Profit -= dividend

	// 账户分红
	s.account.AddDividend(dividend, tax)
	// !!!!!! 除权除息结束

	// 记录
	if recorder != nil {
		recorder.GetChannel() <- setting.MakePositionRecord(
			tm, s.contract.GetInstID(), s.direction, s.valHis.openPrice, s.valHis.lastPrice,
			s.valHis.volume, s.valHis.pnl.Profit,
		)
	}

	return true
}

// (意外)退出后的持仓信息恢复
func (s *StockPosition) DoResume(
	acc handler.Account2,
	contract contract.Contract,
	record recorder.PositionRecord) {
	// 恢复的数据都是历史
	s.contract = contract
	s.account = acc

	s.lastPrice = record.LastPrice
	s.valToday = StockValue{
		lastPrice: record.LastPrice,
	}

	s.valHis = StockValue{
		openPrice: record.CostPrice,
		lastPrice: record.LastPrice,
		volume:    record.Volume,
		available: record.Volume,
		costAmt:   record.CostPrice * record.Volume,
		pnl: account.PnL{
			Profit: (record.LastPrice - record.CostPrice) * record.Volume,
		},
	}
}

func (s *StockPosition) AddSettle(settleQty, settlePrice float64) {
	s.valHis.AddSettle(settleQty, settlePrice)
}

func newStockOpt(opts ...account.WithOpFilterPos) *account.OpFilterPos {
	opt := &account.OpFilterPos{
		Direction:     config.PositionLong,
		SellAvailable: false,
	}
	for _, v := range opts {
		v(opt)
	}

	return opt
}

func (s *StockPosition) CalcPnL(tm time.Time, price float64) {
	if price == 0 {
		return
	}

	s.lastPrice = price
	if s.valToday.volume == 0 && s.valHis.volume == 0 {
		return
	}
	// ! 请考虑账户操作在上一级进行
	s.account.AddDynamicPnL(s.valToday.calcPnL(price), s.valHis.calcPnL(price))
}

func (s *StockPosition) DoOrderUpdate(order handler.Order) {
	if order.OrderDirection() == config.OrderBuy {
		return
	}

	switch order.OrderStatus() {
	case config.OrderStatusNew:
		s.valHis.addSellOrder(order.OrderQty())
	case config.OrderStatusCanceled:
		s.valHis.addSellOrder(-order.OrderQty())
	case config.OrderStatusPartDonePartCancel, config.OrderStatusDone:
		s.valHis.addSellOrder(-order.OrderQty() + order.TradeQty())
	}
}

func (s *StockPosition) DoTradeUpdate(qty, price float64, order handler.Order) (
	deltaMarketValue float64, deltaPnL float64,
) {
	if price == 0 {
		config.ErrorF("成交价格不能为0, order: %+v", order)
	}

	s.lastPrice = price

	if order.PositionDirection() != s.DefaultDirection() {
		config.ErrorF("普通股票只能开多仓, 不能开空仓, order: %+v", order)
		return
	}

	if order.OrderDirection() == config.OrderBuy {
		if s.valToday.volume == 0 && s.valHis.volume == 0 {
			s.openTime = order.OrderTime()
		}

		s.valToday.AddTrade(qty, price)
		deltaPnL += s.valHis.calcPnL(price)
	} else {
		deltaMarketValue, deltaPnL = s.valHis.AddTrade(-qty, price)
		deltaPnL += s.valToday.calcPnL(price)
	}
	return
}
