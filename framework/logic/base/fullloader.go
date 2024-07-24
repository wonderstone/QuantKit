package base

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/base"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/container/btree"
)

// XrxdSlice 除权除息切片
type XrxdSlice struct {
	tm   time.Time
	data map[string]*config.Xrxd
}

type FullLoader struct {
	handler.Resource

	base.Op

	xrxdData map[string]*btree.MapG[time.Time, *config.Xrxd] // 除权除息数据, tm(除权登记日) -> instID -> xrxd
}

func (f *FullLoader) NeedReload() bool {
	return false
}

func (f *FullLoader) Init(resource handler.Resource, option ...base.WithOption) (handler.Basic, error) {
	f.Resource = resource

	op := base.NewOp(option...)
	f.Op = *op

	// 读取除权除息数据
	f.loadXrxd()

	return f, nil
}

func (f *FullLoader) loadXrxd() {
	f.xrxdData = make(map[string]*btree.MapG[time.Time, *config.Xrxd])

	// 加载csv文件
	var data []config.Xrxd
	err := config.ReadCsvFile(f.Op.Path.XrxdFile, &data)
	if err != nil {
		config.WarnF("读取除权除息数据失败, err: %v", err)
	}

	if len(f.Op.InstID) == 0 {
		for i, xrxd := range data {
			if f.WithTimeRange {
				if !(xrxd.RegDate.After(f.Start) && xrxd.RegDate.Before(f.End)) {
					continue
				}
			}

			if v, ok := f.xrxdData[xrxd.InstID]; ok {
				v.Set(xrxd.RegDate.Time, &data[i])
			} else {
				f.xrxdData[xrxd.InstID] = btree.NewMapG[time.Time, *config.Xrxd](
					2,
					func(a, b time.Time) int {
						return a.Compare(b)
					},
					func() (time.Time, *config.Xrxd) {
						return xrxd.RegDate.Time, &data[i]
					},
				)
			}
		}
	} else {
		needInstID := make(map[string]bool)

		for _, instID := range f.Op.InstID {
			needInstID[instID] = true
		}

		for i, xrxd := range data {
			if !needInstID[xrxd.InstID] {
				continue
			}

			if f.WithTimeRange {
				if !(xrxd.RegDate.After(f.Start) && xrxd.RegDate.Before(f.End)) {
					continue
				}
			}

			if v, ok := f.xrxdData[xrxd.InstID]; ok {
				v.Set(xrxd.RegDate.Time, &data[i])
			} else {
				f.xrxdData[xrxd.InstID] = btree.NewMapG[time.Time, *config.Xrxd](
					2,
					func(a, b time.Time) int {
						return a.Compare(b)
					},
					func() (time.Time, *config.Xrxd) {
						return xrxd.RegDate.Time, &data[i]
					},
				)
			}
		}
	}
}

// GetXrxd 获取除权除息数据，从tm这个时间之后的数据
func (f *FullLoader) GetXrxd(instID string, tm time.Time) (
	iter btree.MapIterG[time.Time, *config.Xrxd], needXrxd bool,
) {
	if f.xrxdData == nil {
		return btree.MapIterG[time.Time, *config.Xrxd]{}, false
	}

	xrxd, ok := f.xrxdData[instID]
	if !ok || xrxd == nil {
		return btree.MapIterG[time.Time, *config.Xrxd]{}, false
	}

	iter = xrxd.Iter()

	if iter.Seek(tm) {
		return iter, true
	}

	return btree.MapIterG[time.Time, *config.Xrxd]{}, false
}

func init() {
	setting.RegisterBase(&FullLoader{}, config.HandlerTypeFullLoad)
}
