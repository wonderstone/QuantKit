package formula

import "github.com/wonderstone/QuantKit/config"


type Op struct {
	Config config.Runtime
}

type WithOption func(*Op)

func WithRuntime(config config.Runtime) WithOption {
	return func(op *Op) {
		op.Config = config
	}
}

func NewOp(options ...WithOption) *Op {
	op := &Op{}
	for _, opt := range options {
		opt(op)
	}
	return op
}
