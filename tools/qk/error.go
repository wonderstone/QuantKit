package qk

import (
	"fmt"

	"github.com/wonderstone/QuantKit/config"
)

type ErrInsufficientCash struct {
	Need float64
	Have float64
}

func (e ErrInsufficientCash) Error() string {
	return fmt.Sprintf("资金不足, 需要 %f, 可用 %f", e.Need, e.Have)
}

type ErrInsufficientOrderPriceLimit struct {
}

func (e ErrInsufficientOrderPriceLimit) Error() string {
	return fmt.Sprintf("订单价格为0，订单类型为限价单，不合法")
}

type ErrInsufficientPosition struct {
	Need float64
	Have float64
}

func (e ErrInsufficientPosition) Error() string {
	return fmt.Sprintf("持仓不足, 需要 %f, 可用 %f", e.Need, e.Have)
}

type ErrInsufficientOrderQty struct{}

func (e ErrInsufficientOrderQty) Error() string {
	return fmt.Sprintf("订单数量为0，不合法")
}

type ErrInvalidRunMode struct{ Mode config.Mode }

func (e ErrInvalidRunMode) Error() string {
	return fmt.Sprintf("无效的运行模式: %v", e.Mode)
}
