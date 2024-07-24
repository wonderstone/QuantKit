package quote

import "github.com/wonderstone/QuantKit/config"



type Op struct {
	Config config.Runtime
}

type WithOption func(op *Op)

func WithConfig(config config.Runtime) WithOption {
	return func(op *Op) {
		op.Config = config
	}
}

func NewOp(option ...WithOption) *Op {
	op := &Op{}

	for _, opt := range option {
		opt(op)
	}

	return op
}
