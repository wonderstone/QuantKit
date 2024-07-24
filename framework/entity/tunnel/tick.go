package tunnel

import "time"

type Tick interface {
	// Time 获取tick时间
	Time() time.Time
	// InstID 获取合约ID
	InstID() string
	// High 获取最高价
	High() float64
	// Low 获取最低价
	Low() float64
	// Open 获取开盘价
	Open() float64
	// Close 获取收盘价
	Close() float64
	// Volume 获取成交量
	Volume() float64
	// Amount 获取成交额
	Amount() float64
}
