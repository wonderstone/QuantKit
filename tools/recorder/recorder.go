package recorder

import "github.com/wonderstone/QuantKit/config"

type Handler interface {
	RecordChan() error

	GetChannel() chan any

	Read(data any) error

	QueryRecord(query ...WithQuery) []any

	Release()
}

func NewRecorder[T any](handlerType config.HandlerType, option ...WithOption) Handler {
	switch handlerType {
	case config.HandlerTypeCsv:
		return NewCsvRecorder(option...)
	case config.HandlerTypeMemory:
		return NewMemoryRecorder[T](option...)
	case config.HandlerTypeSqlite, config.HandlerTypeDefault:
		return NewSqliteRecorder[T](option...)
	default:
		config.ErrorF("不支持的记录处理器类型: %s", handlerType)
		return nil
	}
}


// + OrderRecord 订单记录
type OrderRecord struct {
	// gorm.Model `csv:"-"`

	OrderDate         string             `csv:"date" gorm:"primaryKey"`   
	Account           string             `csv:"account" gorm:"primaryKey"`                // 账户id
	OrderId           int64              `csv:"id" gorm:"primaryKey;autoIncrement:false"` // 订单id
	OrderTime         string             `csv:"time"`                                     // 订单时间
	InstId            string             `csv:"inst-id"`                                  // 订单合约
	OrderDirection    string             `csv:"side"`                                     // 订单方向 买卖
	PositionDirection string             `csv:"pos-side"`                                 // 仓位方向 多空
	TransactionType   string             `csv:"trans-type"`                               // 交易类型 开平
	OrderPrice        float64            `csv:"order-price"`                              // 订单价格
	OrderQty          float64            `csv:"order-qty"`                                // 订单数量
	TradePrice        float64            `csv:"trade-price"`                              // 成交价格
	TradeQty          float64            `csv:"trade-qty"`                                // 成交数量
	Commission        float64            `csv:"commission"`                               // 手续费
	Status            config.OrderStatus `csv:"status"`                                   // 订单状态
	RejectReason      string             `csv:"RejectReason"`                             // 拒单原因
}

// + PositionRecord 持仓记录
type PositionRecord struct {
	// gorm.Model `csv:"-"`
	Date        string  `csv:"date" gorm:"primaryKey"`
	Time        string  `csv:"time"`
	InstID      string  `csv:"inst-id" gorm:"primaryKey"`
	Direction   string  `csv:"direction" gorm:"primaryKey"`
	CostPrice   float64 `csv:"cost-price" gorm:"type:float"`
	LastPrice   float64 `csv:"last-price" gorm:"type:float"`
	Volume      float64 `csv:"volume" gorm:"type:float"`
	AvailVolume float64 `csv:"avail-vol" gorm:"type:float"`
	Amt         float64 `csv:"amt" gorm:"type:float"`
	PnL         float64 `csv:"pnl" gorm:"column:pnl;type:float"`
	Account     string  `csv:"account" gorm:"primaryKey"`
}

// + AssetRecord 资产记录
type AssetRecord struct {
	// gorm.Model  `csv:"-"`
	Date        string  `csv:"date" gorm:"primaryKey"`
	Time        string  `csv:"time"`
	MarketValue float64 `csv:"market-val"`
	Margin      float64 `csv:"margin"`
	FundAvail   float64 `csv:"fund-avail"`
	TotalAsset  float64 `csv:"total-asset"`
	Profit      float64 `csv:"profit"`
	Commission  float64 `csv:"commission"`
	Account     string  `csv:"account"  gorm:"primaryKey"`
	Mode        string  `csv:"mode"`
}

