package quote

import (
	"fmt"
	"strconv"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/quote"
	"github.com/wonderstone/QuantKit/framework/entity/tunnel"
	// _ "github.com/wonderstone/QuantKit/ksft"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	// "github.com/wonderstone/QuantKit/tools/container/orderedmap"
	// "github.com/wonderstone/QuantKit/tools/dataframe"
)

type DT struct {
	dt orderedmap.Pair[time.Time, *orderedmap.OrderedMap[string, dataframe.StreamingRecord]]
}

// implements tunnel.Tick all interface
func (d *DT) Time() time.Time {
	return d.dt.Key
}

func (d *DT) InstID() string {
	return d.dt.Value.Newest().Key
}

func (d *DT) High() float64 {
	tmp := d.dt.Value.Newest().Value
	index := tmp.Headers["high"]
	val := tmp.Data[index]
	res, _ := strconv.ParseFloat(val, 64)
	return res
}

func (d *DT) Low() float64 {
	tmp := d.dt.Value.Newest().Value
	index := tmp.Headers["low"]
	val := tmp.Data[index]
	res, _ := strconv.ParseFloat(val, 64)
	return res
}

func (d *DT) Open() float64 {
	tmp := d.dt.Value.Newest().Value
	index := tmp.Headers["open"]
	val := tmp.Data[index]
	res, _ := strconv.ParseFloat(val, 64)
	return res
}

func (d *DT) Close() float64 {
	tmp := d.dt.Value.Newest().Value
	index := tmp.Headers["close"]
	val := tmp.Data[index]
	res, _ := strconv.ParseFloat(val, 64)
	return res
}

func (d *DT) Volume() float64 {
	tmp := d.dt.Value.Newest().Value
	index := tmp.Headers["volume"]
	val := tmp.Data[index]
	res, _ := strconv.ParseFloat(val, 64)
	return res
}

func (d *DT) Amount() float64 {
	tmp := d.dt.Value.Newest().Value
	index := tmp.Headers["amount"]
	val := tmp.Data[index]
	res, _ := strconv.ParseFloat(val, 64)
	return res
}

// a tmp market to read data from local file and publish to channel
// like a real market
type TmpMarket struct {
	Sim     Replay
	PubDtCh chan tunnel.Tick
	inst    string
}

func (tmpmkt *TmpMarket) Init(dlp string, frq config.Frequency, inst string) error {
	tmpmkt.inst = inst
	// * 利用Replay模拟实时行情
	tmpmkt.Sim = Replay{}
	conf := config.Runtime{
		Path: &config.Path{},
	}
	// 模拟器配置
	conf.Path.Download = dlp
	conf.Framework.Frequency = frq
	conf.Framework.Instrument = append(conf.Framework.Instrument, inst)

	err := tmpmkt.Sim.Init(quote.WithConfig(conf))

	// err should be nil
	if err != nil {
		return err
	}
	// 加载数据
	tmpmkt.Sim.LoadData()
	return nil
}

func (tmpmkt *TmpMarket) Subscribe() chan tunnel.Tick {

	tmpmkt.PubDtCh = make(chan tunnel.Tick, 1000)
	return tmpmkt.PubDtCh
}

func (tmpmkt *TmpMarket) Publish() {
	tmpmkt.Sim.Run()
	datach := tmpmkt.Sim.Subscribe()

	for d := range datach.DataChan {
		fmt.Println(d.Key, d.Value.Value(tmpmkt.inst))
		// iter the value
		for item := d.Value.Oldest(); item != nil; item = item.Next() {
			for k, v := range item.Value.Data {
				fmt.Println(k, v)
			}
			tmpmkt.PubDtCh <- &DT{d}
		}
		time.Sleep(time.Millisecond)
	}
	close(tmpmkt.PubDtCh)

}
