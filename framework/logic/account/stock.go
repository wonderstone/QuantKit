package account

import (
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/entity/tunnel"
	"github.com/wonderstone/QuantKit/framework/logic/order"
	"github.com/wonderstone/QuantKit/framework/logic/position"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/container/btree"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/qk"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

type XrxdHandler struct {
	instID2SettleInfo map[string]*btree.MapIterG[time.Time, *config.Xrxd]
	basic             handler.Basic
}

func (x *XrxdHandler) Get(instID string, tm time.Time) (*config.Xrxd, bool) {
	var pXrxd *btree.MapIterG[time.Time, *config.Xrxd]
	if xrxd, ok := x.instID2SettleInfo[instID]; ok {
		pXrxd = xrxd
	} else {
		xrxd, need := x.basic.GetXrxd(instID, tm)
		if !need {
			x.instID2SettleInfo[instID] = nil
			return nil, false
		}

		x.instID2SettleInfo[instID] = &xrxd
		pXrxd = &xrxd
	}

	// 比较除权除息日是否大于当前时间，如果大于当前时间，则不进行除权除息
	if pXrxd == nil || pXrxd.Key().After(tm) {
		return nil, false
	}

	xrxd := pXrxd.Value()

	if !x.instID2SettleInfo[instID].Next() {
		x.instID2SettleInfo[instID] = nil
	}

	return xrxd, true
}

func (x *XrxdHandler) DoSettle(tm time.Time) {
	for instID, xrxd := range x.instID2SettleInfo {
		if xrxd == nil || xrxd.Value().RegDate.After(tm) {
			continue
		}

		if !xrxd.Next() {
			x.instID2SettleInfo[instID] = nil
		}
	}
}

type StockSimple struct {
	handler.Resource
	// 账户ID
	accountID string
	// 资金账户
	cashAccount *config.TradeAcc
	// 账户类型
	marketType string
	// 账户资产
	Asset account.Asset
	// 账户盈亏
	PnL account.PnL
	// 账户持仓
	Position map[string]handler.Position

	orders          []handler.Order
	inst2buyOrders  map[string]*orderedmap.OrderedMap[int64, handler.Order]
	inst2sellOrders map[string]*orderedmap.OrderedMap[int64, handler.Order]

	// 除权除息数据集合
	xrxd   *XrxdHandler
	tunnel tunnel.Tunnel
}

// + 获得运行ID
func (s *StockSimple) GetRunId() string {
	return s.Config().ID
}

// + 获得真实账户用户名，实盘下单用
func (s *StockSimple) CashAccount() string {
	return s.cashAccount.Username
}

// + 插入真实订单，通过tunnel的TradeTunnel
func (s *StockSimple) DoRTInsert(order tunnel.Order) (string, error) {
	return s.tunnel.PlaceOrder(order)
}

// - 撤销真实订单，通过tunnel的TradeTunnel，但我觉得这个功能不必须
func (s *StockSimple) DoRTCancel(orderId string) error {
	return s.tunnel.CancelOrder(orderId)
}

// - 账户恢复，通过asset、order、position记录，但我觉得这个功能不必须
func (s *StockSimple) DoResume(
	assets []recorder.AssetRecord,
	positions []recorder.PositionRecord,
	orders []recorder.OrderRecord,
) {
	// 恢复持仓
	for _, record := range positions {
		if record.Account != s.accountID {
			continue
		}

		contractInfo := s.Contract().GetContract(record.InstID)
		pos := position.NewPosition(s, contractInfo, 0, 0)
		pos.DoResume(s, contractInfo, record)
		s.Position[record.InstID] = pos
	}

	// 恢复资金
	for _, record := range assets {
		if record.Account != s.accountID {
			continue
		}

		s.Asset.Initial = s.Config().Framework.Stock.Cash
		s.Asset.Available = record.FundAvail
		s.Asset.Commission = record.Commission
		s.Asset.MarketValue = record.MarketValue
		s.Asset.Margin = record.Margin
		s.Asset.Total = record.TotalAsset
		s.PnL.Profit = record.Profit
	}

	// 计算并恢复订单
	for _, record := range orders {
		if record.Account != s.accountID {
			continue
		}

		contractInfo := s.Contract().GetContract(record.InstId)
		tm, err := time.Parse(config.TimeFormatDate2+" "+config.TimeFormatTime2, record.OrderDate+" "+record.OrderTime)
		if err != nil {
			config.ErrorF("恢复订单失败: %+v\n", record)
			return
		}

		var orderType config.OrderType
		if record.OrderPrice == 0.0 {
			orderType = config.OrderTypeMarket
		} else {
			orderType = config.OrderTypeLimit
		}

		o := order.ResumeOrder(
			record.OrderId,
			contractInfo, record.OrderQty, &account.OrderOp{
				OrderTime:      &tm,
				OrderPrice:     record.OrderPrice,
				OrderDirection: config.OrderDirection(record.OrderDirection),
				OrderType:      orderType,
			},
		)

		_ = s.InsertOrder(o, ResumeOrder())
		o.DoResume(&record)
	}

	config.InfoF("上场账户%s资金: %+v\n", s.accountID, s.Asset)
}

// - 账户释放资源，但连熊都想不到目前该释放啥
func (s *StockSimple) Release() {

}

// + 获得账户持仓标的切片
func (s *StockSimple) GetPosInstIDs() []string {
	var instIDs []string
	for instID := range s.Position {
		instIDs = append(instIDs, instID)
	}

	return instIDs
}

// - 增加账户可用资金，目前没有使用场景
func (s *StockSimple) AddAvailable(available float64) {
	s.Asset.Available += available
}

// + 通过获得账户持仓
func (s *StockSimple) GetPositionByInstID(instID string) (account.Position, bool) {
	p, ok := s.Position[instID]
	return p, ok
}

func (s *StockSimple) GenOrderId() int64 {
	return s.Account().(*DefaultHandler).GenOrderId()
}

func (s *StockSimple) AddDynamicPnL(pnl ...float64) {
	for _, v := range pnl {
		s.PnL.Profit += v
		s.Asset.Total += v
		s.Asset.Margin += v
		s.Asset.MarketValue += v
	}
}

func (s *StockSimple) CalcPositionPnL(
	tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) {
	for instID, p := range s.Position {
		if v, ok := indicators.Get(instID); ok {
			p.CalcPnL(tm, v.ConvertToFloat("Close"))
		}
	}
}

func (s *StockSimple) Init(account handler.Accounts, option any) error {
	s.xrxd = &XrxdHandler{
		instID2SettleInfo: make(map[string]*btree.MapIterG[time.Time, *config.Xrxd]),
		basic:             account.Base(),
	}

	s.Resource = account.(handler.Resource)

	// s.handler = handler
	s.accountID = config.AccountTypeStockSimple
	s.marketType = config.MarketTypeStock
	s.Asset.Initial = option.(config.Account).Cash
	s.Asset.Available = option.(config.Account).Cash
	s.Asset.Total = option.(config.Account).Cash

	s.Position = make(map[string]handler.Position)

	s.orders = make([]handler.Order, 0)
	s.inst2buyOrders = make(map[string]*orderedmap.OrderedMap[int64, handler.Order])
	s.inst2sellOrders = make(map[string]*orderedmap.OrderedMap[int64, handler.Order])

	// 实盘账户，需要加载交易通道
	if s.Config().Framework.Realtime {
		s.cashAccount = s.Config().Framework.Stock.Account
		s.tunnel = setting.MustNewTunnelHandler(
			config.HandlerType(s.cashAccount.Tunnel),
			tunnel.TradeOnly(),
			tunnel.WithConfig(s.Config()),
			tunnel.WithTunnelUrl(s.cashAccount.URL),
		)
	}

	return nil
}

func (s *StockSimple) GetPosition() (pos map[string]account.Position) {
	pos = make(map[string]account.Position)
	for instID, p := range s.Position {
		pos[instID] = p
	}

	return
}

func (s *StockSimple) GetType() string {
	return config.AccountTypeStockSimple
}

func (s *StockSimple) AccountID() string {
	return s.accountID
}

func (s *StockSimple) GetAsset() account.Asset {
	return s.Asset
}

func (s *StockSimple) GetPnL() account.PnL {
	return s.PnL
}

func (s *StockSimple) DoOrderUpdate(order handler.Order) {
	if order.OrderDirection() == config.OrderBuy {
		// 买入
		switch order.OrderStatus() {
		case config.OrderStatusNew:
			orderAmt := order.OrderAmt()
			commission := order.CommissionFrozen()
			s.Asset.Frozen += orderAmt + commission
			s.Asset.Available -= orderAmt + commission
		case config.OrderStatusCanceled:
			orderAmt := order.OrderAmt()
			s.Asset.Frozen -= orderAmt
			s.Asset.Available += orderAmt
		case config.OrderStatusDone:
			orderAmt := order.OrderAmt()
			tradeAmt := order.TradeAmt()
			commissionFrozen := order.CommissionFrozen()
			commission := order.Commission()
			s.Asset.Commission += commission
			s.Asset.Available += orderAmt - tradeAmt + commissionFrozen - commission
			s.Asset.Total -= commission
			s.PnL.Profit -= commission
			s.Asset.Frozen -= orderAmt - tradeAmt + commissionFrozen
		}
	} else {
		// 卖出
		switch order.OrderStatus() {
		case config.OrderStatusNew:
			commission := order.CommissionFrozen()

			// 卖出订单，冻结手续费资金
			s.Asset.Available -= commission
			s.Asset.Frozen += commission
		// 完结订单，解冻手续费资金，扣除手续费
		case config.OrderStatusCanceled, config.OrderStatusPartDonePartCancel, config.OrderStatusDone:
			commissionFrozen := order.CommissionFrozen()
			commission := order.Commission()
			// 解冻手续费资金
			s.Asset.Available += commissionFrozen - commission
			s.Asset.Frozen -= commissionFrozen
			// 扣除手续费
			s.Asset.Commission += commission
			s.Asset.Total -= commission
			s.PnL.Profit -= commission
		}
	}
}

func (s *StockSimple) DoTradeUpdate(qty, price float64, order handler.Order) {
	deltaMarketValue := 0.0
	deltaPnL := 0.0
	if s.Asset.Frozen < -0.1 {
		config.DebugF("账户冻结资金小于0: %+v\n", s.Asset)
	}
	if s.Position[order.InstID()] != nil {
		deltaMarketValue, deltaPnL = s.Position[order.InstID()].DoTradeUpdate(qty, price, order)
	} else {
		s.Position[order.InstID()] = position.NewPosition(
			order.Account().(handler.Account2),order.Contract(), qty, price,
		)
		deltaMarketValue = 0
		deltaPnL = 0
	}

	s.Asset.MarketValue += deltaPnL + deltaMarketValue
	s.Asset.Total += deltaPnL + deltaMarketValue
	s.PnL.Profit += deltaPnL + deltaMarketValue

	if order.OrderDirection() == config.OrderBuy {
		s.Asset.MarketValue += qty * price
		s.Asset.Frozen -= qty * price
	} else {
		s.Asset.MarketValue -= qty * price
		s.Asset.Available += qty * price
	}

	s.Asset.Margin = s.Asset.MarketValue
}

func getSettlePrice(instID string, indicators *orderedmap.OrderedMap[string, dataframe.StreamingRecord]) float64 {
	if indicators == nil {
		return 0
	}

	if v, ok := indicators.Get(instID); ok {
		settlePrice, err := v.TryConvertToFloat("Close")
		if err == nil {
			return settlePrice
		}
	}

	return 0
}

func (s *StockSimple) DoSettle(
	tm time.Time, indicators *orderedmap.OrderedMap[string, dataframe.StreamingRecord],
	recorder map[config.RecordType]recorder.Handler,
) {
	// 处理订单
	for _, o := range s.orders {
		o.DoSettle(tm, recorder[config.RecordTypeOrder])
	}

	// 处理持仓
	// 如果持仓为空，不处理，且清理持仓
	emptyPos := make([]string, 0)

	for instID, p := range s.Position {
		if !p.DoSettle(
			tm, getSettlePrice(instID, indicators), recorder[config.RecordTypePosition],
		) {
			emptyPos = append(emptyPos, instID)
		}
	}

	for _, settleInfo := range s.xrxd.instID2SettleInfo {
		if settleInfo == nil || settleInfo.Value().RegDate.After(*s.Framework().CurrTime()) {
			continue
		}

		if !settleInfo.Next() {
			settleInfo = nil
		}
	}

	// 删除空仓
	for _, instID := range emptyPos {
		delete(s.Position, instID)
	}

	// 处理资产，冻结资金返回可用并清零
	s.Asset.Available += s.Asset.Frozen
	s.Asset.Frozen = 0

	if r := recorder[config.RecordTypeAsset]; r != nil {
		record := MakeRecord(tm, s.accountID, s.Asset, s.PnL)
		r.GetChannel() <- &record
	}

	s.orders = make([]handler.Order, 0)
	s.inst2buyOrders = make(map[string]*orderedmap.OrderedMap[int64, handler.Order])
	s.inst2sellOrders = make(map[string]*orderedmap.OrderedMap[int64, handler.Order])

	// 将除权除息数据处理
	s.xrxd.DoSettle(tm)
}

func (s *StockSimple) AddDividend(dividend, tax float64) {
	s.Asset.Available += dividend - tax
	s.Asset.MarketValue -= dividend
	s.Asset.Commission += tax // 分红税加入手续费
	s.Asset.Margin = s.Asset.MarketValue
	s.Asset.Total -= tax

	s.PnL.Profit -= tax
}

func (s *StockSimple) GetOrders(instID string) []account.Order {
	var orders []account.Order

	if v, ok := s.inst2buyOrders[instID]; ok {
		for pOrder := v.Oldest(); pOrder != nil; pOrder = pOrder.Next() {
			orders = append(orders, pOrder.Value)
		}
	}

	if v, ok := s.inst2sellOrders[instID]; ok {
		for pOrder := v.Oldest(); pOrder != nil; pOrder = pOrder.Next() {
			orders = append(orders, pOrder.Value)
		}
	}
	return orders
}

func (s *StockSimple) DoMatch(
	tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord], matcher handler.Matcher,
) {
	for pInst := indicators.Oldest(); pInst != nil; pInst = pInst.Next() {
		orders := s.inst2sellOrders[pInst.Key]
		if orders != nil {
			for pOrder := orders.Oldest(); pOrder != nil; pOrder = pOrder.Next() {
				matcher.MatchOrder(pOrder.Value, pInst.Value, tm)
			}
		}

		orders = s.inst2buyOrders[pInst.Key]
		if orders != nil {
			for pOrder := orders.Oldest(); pOrder != nil; pOrder = pOrder.Next() {
				matcher.MatchOrder(pOrder.Value, pInst.Value, tm)
			}
		}

	}
}

func (s *StockSimple) NewOrder(instId string, qty float64, options ...account.WithOrderOption) (account.Order, error) {
	if qty == 0 {
		return nil, qk.ErrInsufficientOrderQty{}
	}

	op := account.NewOrderOp(options...)

	if op.Account == nil {
		op.Account = s
	}

	if op.OrderTime == nil {
		op.OrderTime = s.Account().GetCurrTime()
	}

	if op.OrderPrice == 0 && op.OrderType == config.OrderTypeLimit {
		return nil, qk.ErrInsufficientOrderPriceLimit{}
	}

	contract := s.Contract().GetContract(instId)
	o := order.NewOrder(s.GenOrderId(), contract, qty, op)

	return o, nil
}

func (s *StockSimple) InsertOrder(o account.Order, options ...account.WithOrderOption) error {
	orderOp := o.(handler.Order)
	s.orders = append(s.orders, orderOp)
	op := account.NewOrderOp(options...)
	if op.CheckCash && o.OrderDirection() == config.OrderBuy {
		if s.Asset.Available < o.OrderAmt() {
			err := qk.ErrInsufficientCash{Need: o.OrderAmt(), Have: s.Asset.Available}
			orderOp.DoReject(*s.Account().GetCurrTime(), err)
			return err
		}
	}

	if op.CheckPos && o.OrderDirection() == config.OrderSell {
		pos := s.Position[o.InstID()]
		if pos == nil {
			err := qk.ErrInsufficientPosition{Need: o.OrderQty(), Have: 0}
			orderOp.DoReject(*s.Account().GetCurrTime(), err)
			return err
		} else if pos.Volume(account.WithSellAvailable(true)) < o.OrderQty() {
			err := qk.ErrInsufficientPosition{Need: o.OrderQty(), Have: pos.Volume()}
			orderOp.DoReject(*s.Account().GetCurrTime(), err)
			return err
		}
	}

	if o.OrderDirection() == config.OrderSell {
		if s.inst2sellOrders[o.InstID()] == nil {
			s.inst2sellOrders[o.InstID()] = orderedmap.New[int64, handler.Order](
				orderedmap.WithInitialData(
					orderedmap.Pair[int64, handler.Order]{
						Key: o.ID(), Value: orderOp,
					},
				),
			)
		} else {
			s.inst2sellOrders[o.InstID()].Set(o.ID(), orderOp)
		}
	} else {
		if s.inst2buyOrders[o.InstID()] == nil {
			s.inst2buyOrders[o.InstID()] = orderedmap.New[int64, handler.Order](
				orderedmap.WithInitialData(
					orderedmap.Pair[int64, handler.Order]{
						Key: o.ID(), Value: orderOp,
					},
				),
			)
		} else {
			s.inst2buyOrders[o.InstID()].Set(o.ID(), orderOp)
		}
	}

	s.DoOrderUpdate(orderOp)

	if s.Config().Framework.Realtime && !op.Resumed {

		// !!!!! make orderOp to tunnel.Orde
		// !!!!! order 结构体必须实现tunnekl.Order接口
		_, err := s.DoRTInsert(orderOp.(tunnel.Order))
		// err := orderOp.DoRTInsert()
		if err != nil {
			config.WarnF("实盘下单失败: %+v\n", err)
		} else {
			config.InfoF("实盘下单成功: %+v\n", orderOp)

		}
	}

	return nil
}

func (s *StockSimple) CancelOrder(tm time.Time, id int64) {
	o := s.orders[id]
	o.DoCancelUpdate(tm)

	// 将订单从列表中删除
	if o.OrderDirection() == config.OrderSell {
		orders := *s.inst2sellOrders[o.InstID()]
		orders.Delete(id)
	} else {
		orders := *s.inst2buyOrders[o.InstID()]
		orders.Delete(id)
	}

}

func init() {
	setting.RegisterAccount(&StockSimple{}, config.AccountTypeStockSimple)
}

func (s *StockSimple) CalcSettleInfo(
	instID string, qty, price, lastPrice float64,
) (settleQty, settlePrice, settleLastPrice, dividend, tax float64) {
	if s.Base() == nil {
		return qty, price, lastPrice, 0, 0
	}

	xrxd, ok := s.xrxd.Get(instID, *s.Resource.Framework().CurrTime())
	if !ok {
		return qty, price, lastPrice, 0, 0
	}

	// 计算除权除息
	settleInstID := instID
	settleQty, settlePrice, settleLastPrice, dividend, tax, settleInstID = xrxd.CalcPos(qty, price, lastPrice)

	// 如果换股，instID会变化，因此直接清空本position的数据，添加或者更新新的position
	if settleInstID != "" {
		if pos, ok := s.Position[settleInstID]; ok {
			pos.AddSettle(settleQty, settlePrice)
		} else {
			contractInfo := s.Contract().GetContract(settleInstID)
			s.Position[settleInstID] = position.NewPosition(s,contractInfo, settleQty, settlePrice)
		}

		s.Asset.Available += dividend - tax
		s.Asset.MarketValue -= dividend
		s.Asset.Margin = s.Asset.MarketValue
		s.Asset.Commission += tax // 分红税加入手续费
		s.Asset.Total -= tax

		s.PnL.Profit -= tax

		return qty, price, lastPrice, 0, 0
	}

	// 其他情况，直接返回
	return
}
