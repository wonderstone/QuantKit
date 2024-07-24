package config

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// ModelType 模型类型
type ModelType string

const (
	ModelTypeGenome    ModelType = "Genome"
	ModelTypeGenomeSet ModelType = "GenomeSet"
)

// Frequency 频率类型
type Frequency string

const (
	Frequency1Min  Frequency = "1min"
	Frequency5Min  Frequency = "5min"
	Frequency15Min Frequency = "15min"
	Frequency30Min Frequency = "30min"
	Frequency120Min Frequency = "120min"
	Frequency1Day  Frequency = "1day"
)

type HandlerType string

const (
	HandlerTypeDefault      HandlerType = "default"    // 默认模式
	HandlerTypeConfig       HandlerType = "config"     // 配置模式 (用于所有能够从配置文件读取的模式, 例如: 交易日历, 合约)
	HandlerTypeCsv          HandlerType = "csv"        // csv加载模式 (用于所有需要全量加载的模式, 例如: 指标计算, 指标回放)
	HandlerTypeMemory       HandlerType = "memory"     // csv加载模式 (用于所有需要全量加载的模式, 例如: 指标计算, 指标回放)
	HandlerTypeSqlite       HandlerType = "sqlite"     // sqlite加载模式 (用于所有需要全量加载的模式, 例如: 指标计算, 指标回放)
	HandlerTypeFullLoad     HandlerType = "full"       // 全量加载模式 (用于所有需要全量加载的模式, 例如: 指标计算, 指标回放)
	HandlerTypeStream       HandlerType = "stream"     // 流式加载模式 (用于所有需要流式加载的模式, 例如: 回测, 实盘)
	HandlerTypeNextMatch    HandlerType = "next-match" // 下一次匹配模式 (用于所有需要下一个周期进行处理的模式, 例如: 撮合，回测)
	HandlerTypeCurrMatch    HandlerType = "curr-match" // 当前周期匹配模式 (用于所有需要当前周期进行处理的模式, 例如: 撮合，回测)
	HandlerTypeDailyMode    HandlerType = "daily"      // 每日模式 (专用与创建回测框架)
	HandlerTypeGepMode      HandlerType = "gep"        // 训练和回测GEP模式
	HandlerTypeFullXrxdMode HandlerType = "full-xrxd"  // 使用全复权模式(默认)
	HandlerTypePreXrxdMode  HandlerType = "pre-xrxd"   // 使用前复权模式
	HandlerTypeReplay       HandlerType = "replay"     // 使用回放
	HandlerTypeRealtime     HandlerType = "realtime"   // 使用实盘
	HandlerTypeKsft         HandlerType = "ksft"       // 隧道模式(ksft)
)

// Mode 运行模式
type Mode string // 模式

const (
	CalcMode  Mode = "calc"    // 训练模式
	TrainMode Mode = "train"   // 训练模式
	BTMode    Mode = "bt"      // 回测模式
	RunMode   Mode = "runtime" // 运行模式
)

// MarketType 市场类型
type MarketType string

const (
	MarketTypeStock  = "stock"
	MarketTypeFuture = "future"
)

// AccountType 账户类型
type AccountType string

const (
	AccountTypeStockSimple = "stock"        // 股票账户(简单版本)
	AccountTypeFuture      = "future"       // 期货账户
	AccountTypeMarginStock = "margin-stock" // 融资融券账户
)

// StockSpecType 股票类型特化功能
type StockSpecType string

const (
	StockSpecTypeSimple StockSpecType = "simple" // 股票账户(简单版本)
)

// RecordType 记录类型
type RecordType string

const (
	RecordTypeOrder    = "order"
	RecordTypeTrade    = "trade"
	RecordTypePosition = "position"
	RecordTypeAsset    = "asset"
)

// OrderDirection 委托方向
type OrderDirection string

const (
	OrderBuy  OrderDirection = "B"
	OrderSell OrderDirection = "S"
)

// PositionDirection 持仓方向
type PositionDirection string

const (
	PositionOverall = "O"
	PositionLong    = "L"
	PositionShort   = "S"
)

// TransactionType 开平标志
type TransactionType string

const (
	OffsetOpen     = "O"
	OffsetClose    = "C"
	OffsetCloseHis = "H"
)

// 买=开多 ｜ 平空
// 卖=开空 ｜ 平多

// OrderType 委托类型
type OrderType string

const (
	OrderTypeLimit  = "L"
	OrderTypeMarket = "M"
)

// OrderStatus 委托状态
type OrderStatus int

const (
	OrderStatusNew                = 1 // 新单
	OrderStatusPartDone           = 2 // 部分成交
	OrderStatusPartDonePartCancel = 3 // 部分成交部分撤单
	OrderStatusDone               = 4 // 全部成交
	OrderStatusCanceled           = 5 // 撤单
	OrderStatusExpired            = 6 // 过期
	OrderStatusRejected           = 7 // 拒单
)

type StrategyParamType string

const (
	StrategyParamTypeInt         StrategyParamType = "int"
	StrategyParamTypeFloat       StrategyParamType = "float"
	StrategyParamTypeString      StrategyParamType = "string"
	StrategyParamTypeArrayInt    StrategyParamType = "array<int>"
	StrategyParamTypeArrayFloat  StrategyParamType = "array<float>"
	StrategyParamTypeArrayString StrategyParamType = "array<string>"
)

type EventType string

const (
	EventTypeDayOpen   EventType = "day-open"
	EventTypeDayClose  EventType = "day-close"
	EventTypeSettle    EventType = "day-settle"
	EventTypeTick      EventType = "tick"
	EventTypeTickNext  EventType = "tick-next"
	EventTypeTickClose EventType = "tick-close"
)

const (
	TimeFormatDefault string = "2006.01.02T15:04:05.000"
	TimeFormatDate    string = "20060102"
	TimeFormatDate2   string = "2006-01-02"
	TimeFormatRange   string = "20060102-150405.000"
	TimeFormatHHMM    string = "15:04"
	TimeFormatTime    string = "15:04:05.000"
	TimeFormatTime2   string = "15:04:05"
)
