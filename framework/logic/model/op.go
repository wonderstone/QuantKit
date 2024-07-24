package model

import (
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
)

type Op struct {
	runner      handler.Global
	gepOps      *model.Op
	numTerminal int
}

type WithOption func(*Op)

func NewOp(options ...WithOption) *Op {
	op := &Op{}
	for _, option := range options {
		option(op)
	}

	return op
}

func WithRunner(runner handler.Global) WithOption {
	return func(op *Op) {
		op.runner = runner
	}
}

func WithModelOption(option ...model.WithOption) WithOption {
	return func(op *Op) {
		op.gepOps = model.NewOp(option...)
	}
}

func WithNumTerminal(num int) WithOption {
	return func(op *Op) {
		op.numTerminal = num
	}
}
