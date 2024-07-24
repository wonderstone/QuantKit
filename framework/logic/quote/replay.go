package quote

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/entity/quote"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/container/btree"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/idgen"
)

type Replay struct {
	dataPath string

	instID []string // 合约ID

	dfs *orderedmap.OrderedMap[string, *dataframe.DataFrame] 	// + key: instID  value:dataframe

	columns map[string]int

	subs map[int64]*handler.Channel

	gen *idgen.AutoInc

	mutex sync.RWMutex

	wg sync.WaitGroup
}

type WithFilter func(name string) bool

func WithAll(name string) bool {
	return true
}

func WithPrefix(name string) bool {
	return true
}

func (f *Replay) Init(option ...quote.WithOption) error {
	op := quote.NewOp(option...)
	f.gen = idgen.New(100000, 1)
	f.subs = make(map[int64]*handler.Channel)

	for _, withOption := range option {
		withOption(op)
	}

	f.dataPath = filepath.Join(op.Config.Path.Download, string(op.Config.Framework.Frequency))
	f.instID = op.Config.Framework.Instrument

	if len(op.Config.Framework.GroupInstrument) != 0 {
		var filters []WithFilter
		for _, instGrp := range op.Config.Framework.GroupInstrument {
			switch instGrp {
			case "all":
				filters = append(filters, WithAll)
			case "600.PRX":
				filters = append(
					filters, func(name string) bool {
						return name[:3] == "600"
					},
				)
			case "300.PRX":
				filters = append(
					filters, func(name string) bool {
						return name[:3] == "300"
					},
				)
			case "00.PRX":
				filters = append(
					filters, func(name string) bool {
						return name[:2] == "00"
					},
				)
			case "688.PRX":
				filters = append(
					filters, func(name string) bool {
						return name[:3] == "688"
					},
				)
			}
		}

		err := filepath.Walk(
			f.dataPath, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					if filepath.Ext(info.Name()) == ".csv" {
						for _, filter := range filters {
							if filter(info.Name()) {
								f.instID = append(f.instID, info.Name()[0:len(info.Name())-4])
								return nil
							}
						}
					}
				}
				return nil
			},
		)

		if err != nil {
			config.ErrorF("加载指标数据失败: %v", err)
		}
	}

	f.dfs = orderedmap.New[string, *dataframe.DataFrame]()

	return nil
}

func (f *Replay) SetIndicatorDataPath(dir string) {
	f.dataPath = dir
}

func (f *Replay) LoadData() {
	dfs, err := dataframe.LoadFrames(f.dataPath, f.instID)
	if err != nil {
		config.ErrorF("加载行情数据失败: %v", err)
	}

	if dfs == nil {
		config.ErrorF("加载的行情数据为空，可能没有正确设置所需合约[framework->instrument]")
	}

	if len(dfs) != len(f.instID) {
		config.WarnF("加载行情数据长度与instrument长度不一致，可能数据不完整")
	}

	for i := range dfs {
		f.dfs.Set(f.instID[i], &dfs[i])
	}

	if f.dfs.Len() == 0 {
		config.ErrorF("没有加载数据")
	}

	f.columns = f.dfs.Value(f.instID[0]).HeaderToIndex
}

func (f *Replay) Subscribe() *handler.Channel {
	ch := &handler.Channel{
		DataChan: make(chan orderedmap.Pair[time.Time, *orderedmap.OrderedMap[string, dataframe.StreamingRecord]], 100),
		StopChan: make(chan struct{}),
	}

	id := f.gen.Id()
	f.subs[id] = ch
	return ch
}

func (f *Replay) GetRecord(instId string, tm time.Time, xrxd config.Xrxd) {
	// if df, ok := f.dfs.Get(instId); ok {
	// 	// 将tick全部进行前复权
	// 	df.Value.ForwardAdjust(xrxd)
	// }
	//
	// return nil
}

func (f *Replay) Run() {
	all := btree.NewMapG[time.Time, *orderedmap.OrderedMap[string, dataframe.StreamingRecord]](
		2,
		func(a, b time.Time) int {
			return a.Compare(b)
		},
	)

	for pair := f.dfs.Oldest(); pair != nil; pair = pair.Next() {
		for _, record := range pair.Value.FrameRecords {
			tm := record.ConvertToTime("Time", f.columns)
			if v, ok := all.Get(tm); !ok {
				all.Set(
					tm, orderedmap.New[string, dataframe.StreamingRecord](
						orderedmap.WithInitialData(
							orderedmap.Pair[string, dataframe.StreamingRecord]{
								Key:   pair.Key,
								Value: dataframe.StreamingRecord{Data: record.Data, Headers: f.columns},
							},
						),
					),
				)
			} else {
				v.Set(pair.Key, dataframe.StreamingRecord{Data: record.Data, Headers: f.columns})
			}
		}
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		iter := all.Iter()
		for ok := iter.First(); ok; ok = iter.Next() {
			p := orderedmap.Pair[time.Time, *orderedmap.OrderedMap[string, dataframe.StreamingRecord]]{
				Key:   iter.Key(),
				Value: iter.Value(),
			}

			if len(f.subs) == 0 {
				break
			}

			for i, ch := range f.subs {
				select {
				case <-ch.StopChan:
					close(ch.DataChan)
					delete(f.subs, i)
				case ch.DataChan <- p:
				}
			}
		}

		if len(f.subs) == 0 {
			return
		}

		for i, ch := range f.subs {
			ch.CloseFlag.Do(
				func() {
					close(f.subs[i].DataChan)
				},
			)
		}

		f.subs = make(map[int64]*handler.Channel)
	}()

}

func (f *Replay) WaitForShutdown() {
	f.wg.Wait()
}

func init() {
	setting.RegisterNewQuote(
		&Replay{},
		config.HandlerTypeReplay,
	)
}
