package common

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
)

// Event 事件，在实盘使用，用于传递策略的运行事件
// EventType 事件类型
// 为EventTypeDayOpen时，Data为*config.DayOpen
// 为EventTypeDayClose时，Data为*config.DayClose
// 为EventTypeTick时，Data为*config.Tick
type Event struct {
	Tm   time.Time
	Type config.EventType

	Data any
}
