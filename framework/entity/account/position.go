package account

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
)

type Position interface {
	// - fundamental info parts
	InstID() string
	DefaultDirection() config.PositionDirection
	AvailableDirection() []config.PositionDirection
	OpenPrice(opts ...WithOpFilterPos) float64
	OpenTime(opts ...WithOpFilterPos) time.Time
	Volume(opts ...WithOpFilterPos) float64
	Amt(opts ...WithOpFilterPos) float64
	PnL(opts ...WithOpFilterPos) PnL // 持仓盈亏
}
