package base

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
)

type Op struct {
	Path   *config.Path
	InstID []string

	WithTimeRange bool
	Start         time.Time
	End           time.Time

	XrxdHandlerType config.HandlerType
}

type WithOption func(*Op)

func NewOp(options ...WithOption) *Op {
	op := &Op{
		XrxdHandlerType: config.HandlerTypeCsv,
	}
	for _, option := range options {
		option(op)
	}
	return op
}

// WithPath 本地基础数据文件路径
func WithPath(path *config.Path) WithOption {
	return func(op *Op) {
		op.Path = path
	}
}

// WithXrxdHandler 除权除息处理器类型
func WithXrxdHandler(handlerType config.HandlerType) WithOption {
	return func(op *Op) {
		op.XrxdHandlerType = handlerType
	}
}

func WithInstID(instID []string) WithOption {
	return func(op *Op) {
		op.InstID = instID
	}
}

// WithTimeRange 指定时间范围，只关注日期在该范围内的数据
func WithTimeRange(start, end time.Time) WithOption {
	return func(op *Op) {
		op.Start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)

		op.End = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.Local)

		op.WithTimeRange = true
	}
}
