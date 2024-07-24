package tunnel



type QuoteTunnel interface {
	// RegTick 注册tick数据
	RegTick(inst []string) error
}

type TradeTunnel interface {
	// PlaceOrder 下单
	PlaceOrder(order Order) (orderID string, err error)

	// CancelOrder 撤单
	CancelOrder(orderID string) error
}

type Tunnel interface {
	QuoteTunnel
	TradeTunnel

	Init(options ...WithOption) error
}

