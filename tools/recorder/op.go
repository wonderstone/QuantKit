package recorder

type Op struct {
	file        string
	plusMode    bool
	transaction bool
}

type WithOption func(op *Op)

func WithPlusMode() WithOption {
	return func(op *Op) {
		op.plusMode = true
	}
}

func WithTransaction() WithOption {
	return func(op *Op) {
		op.transaction = true
	}
}

func WithFilePath(file string) WithOption {
	return func(op *Op) {
		op.file = file
	}
}

func NewOp(option ...WithOption) *Op {
	op := &Op{}

	for _, o := range option {
		o(op)
	}

	return op
}
