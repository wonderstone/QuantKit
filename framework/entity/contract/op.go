package contract

import "github.com/wonderstone/QuantKit/config"

type Op struct {
	Property *config.ContractProperty
	XrxdFile string
}

type WithOption func(*Op)

func NewOp(options ...WithOption) *Op {
	op := &Op{}
	for _, option := range options {
		option(op)
	}
	return op
}

func WithProperty(property *config.ContractProperty) WithOption {
	return func(op *Op) {
		op.Property = property
	}
}

func WithXrxdFile(file string) WithOption {
	return func(op *Op) {
		op.XrxdFile = file
	}
}
