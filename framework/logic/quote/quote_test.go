package quote

import (
	"fmt"
	"testing"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/quote"
	// _ "github.com/wonderstone/QuantKit/ksft"
	// "github.com/wonderstone/QuantKit/tools/container/orderedmap"
	// "github.com/wonderstone/QuantKit/tools/dataframe"
)

// test tmpmarket Subscribe
func TestMarket(t *testing.T) {
	// build a market
	var market TmpMarket
	// initialize the market with data already
	market.Init("./testdata","30min","510300.XSHG.CS")

	// subscribe to the market
	pub := market.Subscribe()
	// tell the market to publish the data
	go market.Publish()
	
	for d := range pub {
		fmt.Println(d.Time(), d.Close())
	}
	// wait for the market to shutdown
	market.Sim.WaitForShutdown()
}


// test Realtime
func TestRealtime(t *testing.T) {

	// build a market
	var market TmpMarket
	// initialize the market with data already
	market.Init("./testdata","30min","510300.XSHG.CS")

	// subscribe to the market

	// pub := market.Subscribe("510300.XSHG.CS")
	// for d := range pub {
	// 	fmt.Println(d.Time(), d.Close())
	// }


	handlerMod := Realtime{}
	conf := config.Runtime{
		Path: &config.Path{},
	}


	conf.Path.Download = "./testdata"
	err := handlerMod.Init(quote.WithConfig(conf))

	// err should be nil
	if err != nil {
		panic(err)
	}

	handlerMod.LoadData()


}





func TestReplay(t *testing.T) {
	handleMod := Replay{}
	conf := config.Runtime{
		Path: &config.Path{},
	}

	conf.Path.Download = "./testdata"
	conf.Framework.Frequency = "30min"
	conf.Framework.Instrument = append(conf.Framework.Instrument, "510300.XSHG.CS")

	err := handleMod.Init(quote.WithConfig(conf))

	// err should be nil
	if err != nil {
		panic(err)
	}

	handleMod.LoadData()

	sub := handleMod.Subscribe()

	handleMod.Run()

	go func() {
		for d := range sub.DataChan {
			fmt.Println(d.Key, d.Value.Value("510300.XSHG.CS"))
			// iter the value
			for item := d.Value.Oldest(); item != nil; item = item.Next() {
				for k, v := range item.Value.Data {
					fmt.Println(k, v)

				}
				fmt.Println(item.Value)
			}
		}
	}()

	handleMod.WaitForShutdown()

	fmt.Println(err)

}

// test config.scheduler

func TestScheduler(t *testing.T) {

	//var scheduler *config.Scheduler
	scheduler := config.NewScheduler()
	times := 8 * time.Second
	tmp := time.Now().Truncate(times).Add(times)
	// fmt.Printf("下一次触发时间: %s\n", t)
	_, err := scheduler.Add(
		&config.Task{
			Interval:   time.Second,
			RunOnce:    false,
			StartAfter: tmp,
			TaskFunc: func() error {
				tmp = time.Now()
				fmt.Println("hello", tmp)
				return nil
			},
		},
	)
	if err != nil {
		panic(err)
	}

	time.Sleep(20 * time.Second)

}
