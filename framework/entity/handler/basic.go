package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/base"
	"github.com/wonderstone/QuantKit/tools/container/btree"
)

type Basic interface {
	Resource

	Init(resource Resource, option ...base.WithOption) (Basic, error)

	// NeedReload 判断是否需要重新加载， 对于回测模式的实现是不需要reload的
	NeedReload() bool

	// GetXrxd 获取除权除息数据
	GetXrxd(instID string, tm time.Time) (
		iter btree.MapIterG[time.Time, *config.Xrxd], needXrxd bool,
	)
}
