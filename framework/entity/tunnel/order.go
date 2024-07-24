package tunnel

type Order interface {
	// RequestID 获取请求ID
	RequestID() string
	// InstID 获取合约ID
	InstID() string
	// Direction 获取订单方向
	Direction() string
	// Volume 获取订单数量
	Volume() float64
	// Price 获取订单价格
	Price() float64
	// AccountID 获取账户ID
	AccountID() string
	// CashAccount 获取资金账户ID
	CashAccount() string
}
