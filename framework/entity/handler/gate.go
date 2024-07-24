package handler

import (
	"sync"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/quote"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type Quote interface {
	// Resource

	Init(option ...quote.WithOption) error
	// LoadData 加载数据
	LoadData()
	// Subscribe 订阅数据
	Subscribe() *Channel
	// Run 发布数据
	Run()
	// Close 关闭
	WaitForShutdown()
}

type Channel struct {
	DataChan  chan orderedmap.Pair[time.Time, *orderedmap.OrderedMap[string, dataframe.StreamingRecord]]
	StopChan  chan struct{}
	CloseFlag sync.Once
}
