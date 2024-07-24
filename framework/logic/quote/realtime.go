package quote

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/entity/quote"
	"github.com/wonderstone/QuantKit/framework/entity/tunnel"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/container/queue"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type InstTicks struct {
	InstID  string
	Columns map[string]int
	Ticks   *queue.Queue[tunnel.Tick]
}

func NewInstTicks(instID string, columns map[string]int) *InstTicks {
	return &InstTicks{
		InstID:  instID,
		Columns: columns,
		// !!!! possibly bug here 
		Ticks:   queue.New[tunnel.Tick](100),
	}
}

func (i *InstTicks) Push(t tunnel.Tick) {
	// fmt.Printf(
	// 	"%s %s %f %f %f %f %f %f\n", t.Time(), t.InstID(), t.High(), t.Low(), t.Open(), t.Close(), t.Volume(),
	// 	t.Amount(),
	// )
	i.Ticks.Enqueue(t)
}

func (i *InstTicks) Sum(tm time.Time) *dataframe.StreamingRecord {
	if i.Ticks.Len() == 0 {
		return nil
	}
	bar := dataframe.StreamingRecord{
		Headers: i.Columns,
		Data:    make([]string, len(i.Columns)),
	}

	var high, low, open, close_, volume, amount float64
	if e, ok := i.Ticks.Dequeue(); ok {
		high = e.High()
		low = e.Low()
		open = e.Open()
		close_ = e.Close()
		volume = e.Volume()
		amount = e.Amount()
	}
	for e, ok := i.Ticks.Dequeue(); ok; e, ok = i.Ticks.Dequeue() {
		high = math.Max(high, e.High())
		low = math.Min(low, e.Low())
		close_ = e.Close()
		volume += e.Volume()
		amount += e.Amount()
	}

	bar.Update("Date", tm.Format(config.TimeFormatDate))
	bar.Update("Time", tm.Format(config.TimeFormatDefault))
	bar.Update("High", fmt.Sprintf("%f", high))
	bar.Update("Low", fmt.Sprintf("%f", low))
	bar.Update("Open", fmt.Sprintf("%f", open))
	bar.Update("Close", fmt.Sprintf("%f", close_))
	bar.Update("Volume", fmt.Sprintf("%f", volume))
	bar.Update("Amount", fmt.Sprintf("%f", amount))

	// fmt.Printf(
	// 	"Bar: %s %s %f %f %f %f %f %f\n", tm.Format(config.TimeFormatDefault), i.InstID, high, low, open, close_,
	// 	volume,
	// 	amount,
	// )

	return &bar
}

func (i *InstTicks) Trim(tm time.Time) {
	if i.Ticks.Len() == 0 {
		return
	}

	i.Ticks.Clear()
}

type RTEvent struct {
	tm time.Time
	ev config.EventType
}

type Realtime struct {
	config config.Runtime

	dataPath string

	instID []string // 合约ID

	tm time.Time

	nextTriggerTime       time.Time
	nextTriggerTimeNext   time.Time
	nextTriggerTimeSettle time.Time

	scheduler *config.Scheduler

	df *orderedmap.OrderedMap[string, dataframe.StreamingRecord]

	columns map[string]int

	sub *handler.Channel

	mutex sync.RWMutex

	wg sync.WaitGroup

	tunnel tunnel.Tunnel

	ticks map[string]*InstTicks

	// 收到的数据
	recvChan chan tunnel.Tick

	eventChan chan RTEvent

	quit chan struct{}
}

func (f *Realtime) Init(option ...quote.WithOption) error {
	op := quote.NewOp(option...)

	for _, withOption := range option {
		withOption(op)
	}

	f.config = op.Config
	f.quit = make(chan struct{})

	f.columns = map[string]int{
		"Date":   0,
		"Time":   1,
		"High":   2,
		"Low":    3,
		"Open":   4,
		"Close":  5,
		"Volume": 6,
		"Amount": 7,
	}

	f.scheduler = config.NewScheduler()

	// 设置监听操作系统信号并正确停止调度
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		config.InfoF("收到退出信号，正在退出...")
		f.scheduler.Stop()
		f.quit <- struct{}{}
	}()

	f.eventChan = make(chan RTEvent, 1)

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

	f.ticks = make(map[string]*InstTicks, len(f.instID))
	for _, instID := range f.instID {
		f.ticks[instID] = NewInstTicks(instID, f.columns)
	}

	return nil
}

func (f *Realtime) LoadData() {
	// 建立与实盘行情通道的连接
	f.recvChan = make(chan tunnel.Tick)

	f.tunnel = setting.MustNewTunnelHandler(
		config.HandlerTypeDefault,
		tunnel.QuoteOnly(),
		tunnel.WithConfig(&f.config),
		tunnel.WithTickChannel(f.recvChan),
	)

	err := f.tunnel.RegTick(f.instID)
	if err != nil {
		config.ErrorF("注册行情失败: %v", err)
	}
}

func (f *Realtime) Subscribe() *handler.Channel {
	f.sub = &handler.Channel{
		DataChan: make(chan orderedmap.Pair[time.Time, *orderedmap.OrderedMap[string, dataframe.StreamingRecord]], 100),
		StopChan: make(chan struct{}),
	}

	return f.sub
}

func (f *Realtime) GetRecord(instId string, tm time.Time, xrxd config.Xrxd) {
	// if df, ok := f.dfs.Get(instId); ok {
	// 	// 将tick全部进行前复权
	// 	df.Value.ForwardAdjust(xrxd)
	// }
	//
	// return nil
}

func (f *Realtime) doCalc() {

}

func (f *Realtime) updateTimer() {
	var times time.Duration
	switch f.config.Framework.Frequency {
	case config.Frequency1Min, config.Frequency1Day:
		times = 1 * time.Minute
	case config.Frequency5Min:
		times = 5 * time.Minute
	case config.Frequency15Min:
		times = 15 * time.Minute
	case config.Frequency30Min:
		times = 30 * time.Minute
	}

	t := time.Now().Truncate(times).Add(times)
	// fmt.Printf("下一次触发时间: %s\n", t)
	_, err := f.scheduler.Add(
		&config.Task{
			RunOnce:    true,
			StartAfter: t,
			TaskFunc: func() error {
				f.eventChan <- RTEvent{
					tm: t,
					ev: config.EventTypeTick,
				}
				return nil
			},
		},
	)
	if err != nil {
		config.ErrorF("添加定时任务失败: %v", err)
		return
	}
}

func (f *Realtime) sendData(tm time.Time) {
	records := orderedmap.New[string, dataframe.StreamingRecord]()
	for _, tick := range f.ticks {
		if bar := tick.Sum(tm); bar != nil {
			records.Set(tick.InstID, *bar)
		}
	}

	if records.Len() != 0 {
		f.sub.DataChan <- orderedmap.Pair[time.Time, *orderedmap.OrderedMap[string, dataframe.StreamingRecord]]{
			Key:   tm,
			Value: records,
		}
	}
}

func (f *Realtime) trimData(tm time.Time) {
	for _, tick := range f.ticks {
		tick.Trim(tm)
	}
}

func (f *Realtime) runOnce() bool {
	select {
	case <-f.sub.StopChan:
		fmt.Println("停止")
		close(f.sub.DataChan)
		f.sub = nil
		return false
	case p := <-f.recvChan:
		if tick, ok := f.ticks[p.InstID()]; ok {
			tick.Push(p)
		}

	case t := <-f.eventChan:
		switch f.config.Framework.Frequency {
		case config.Frequency1Min, config.Frequency5Min, config.Frequency15Min, config.Frequency30Min:
			config.InfoF("[%s]发送数据，更新下一次触发时间", t.tm)
			f.sendData(t.tm)
		case config.Frequency1Day:
			if t.tm.Equal(f.nextTriggerTime) {
				config.InfoF("[%s]发送数据，更新下一次下单时间", t.tm)
				f.sendData(t.tm)
				f.nextTriggerTime.AddDate(0, 0, 1)
			} else if t.tm.Equal(f.nextTriggerTimeNext) {
				config.InfoF("[%s]发送数据，更新下一次撮合时间", t.tm)
				f.sendData(t.tm)
				f.nextTriggerTimeNext.AddDate(0, 0, 1)
			} else if t.tm.Equal(f.nextTriggerTimeSettle) {
				config.InfoF("[%s]发送数据，更新下一次结算时间", t.tm)
				f.sendData(t.tm)
				f.nextTriggerTimeNext.AddDate(0, 0, 1)
			} else {
				f.trimData(t.tm)
			}
		}

		f.updateTimer()
	case <-f.quit:
		fmt.Println("退出")
		return false
	}

	return true
}

func (f *Realtime) Run() {
	f.wg.Add(1)

	f.updateTimer()

	if f.config.Framework.Frequency == config.Frequency1Day {
		now := time.Now()
		f.nextTriggerTime = time.Date(
			now.Year(),
			now.Month(),
			now.Day(), 0, 0, 0, 0,
			now.Location(),
		).Add(f.config.Framework.DailyTriggerTime)

		if f.nextTriggerTime.Before(time.Now()) {
			f.nextTriggerTime = f.nextTriggerTime.AddDate(0, 0, 1)
		}

		f.nextTriggerTimeNext = f.nextTriggerTime.Add(time.Minute)
		f.nextTriggerTimeSettle = time.Date(
			now.Year(),
			now.Month(),
			now.Day(), 15, 0, 0, 0,
			now.Location(),
		)

		if f.nextTriggerTime.Before(time.Now()) {
			f.nextTriggerTime = f.nextTriggerTime.AddDate(0, 0, 1)
			f.nextTriggerTimeNext = f.nextTriggerTimeNext.AddDate(0, 0, 1)
			f.nextTriggerTimeSettle = f.nextTriggerTimeSettle.AddDate(0, 0, 1)
		}
	}
	go func() {
		defer f.wg.Done()
		for {
			if !f.runOnce() {
				break
			}
		}

		if f.sub == nil {
			fmt.Printf("无订阅退出\n")
			return
		}

		f.sub.CloseFlag.Do(
			func() {
				close(f.sub.DataChan)
			},
		)

		f.sub = nil
	}()

}

func (f *Realtime) WaitForShutdown() {
	f.wg.Wait()
}

func init() {
	// ! maybe wrong
	setting.RegisterNewQuote(
		&Realtime{},
		config.HandlerTypeRealtime,
	)

}
