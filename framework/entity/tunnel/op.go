package tunnel

import (
	"github.com/wonderstone/QuantKit/config"
)

type Op struct {
	Runtime     *config.Runtime
	TickChannel chan Tick
	QuoteOnly   bool
	TradeOnly   bool

	TunnelUrl string
}

type WithOption func(*Op)

func NewOp(options ...WithOption) *Op {
	op := &Op{}
	for _, option := range options {
		option(op)
	}
	return op
}

func QuoteOnly() WithOption {
	return func(op *Op) {
		op.QuoteOnly = true
	}
}

func TradeOnly() WithOption {
	return func(op *Op) {
		op.TradeOnly = true
	}
}

func WithConfig(runtime *config.Runtime) WithOption {
	return func(op *Op) {
		op.Runtime = runtime
	}
}

func WithTunnelUrl(url string) WithOption {
	return func(op *Op) {
		op.TunnelUrl = url
	}
}

func WithTickChannel(channel chan Tick) WithOption {
	return func(op *Op) {
		op.TickChannel = channel
	}
}
