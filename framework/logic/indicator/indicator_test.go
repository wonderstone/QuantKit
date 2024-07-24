package indicator

import (
	"fmt"
	"sync"
	"testing"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/formula"
	q "github.com/wonderstone/QuantKit/framework/entity/quote"
	"github.com/wonderstone/QuantKit/framework/logic/quote"
)

// 测试全量读取指标，并且回放
func TestFullLoader(t *testing.T) {

	var wg sync.WaitGroup
	// use replay mode to load all data
	handleMod := quote.Replay{}
	conf := config.Runtime{
		Path: &config.Path{},
	}
	
	conf.Path.Download = "./testdata"
	conf.Framework.Frequency = "30min"
	conf.Framework.Instrument = append(conf.Framework.Instrument, "510300.XSHG.CS")

	err := handleMod.Init(q.WithConfig(conf))
	
	// err should be nil 
	if err != nil {
		panic(err)
	}

	handleMod.LoadData()

	sub := handleMod.Subscribe()

	handleMod.Run()

	indicator, err := config.NewIndicatorConfig("./indicator.yaml")
	if err != nil {
		panic(err)
	}
	conf.Indicator = indicator

	// new calculator
	// calct := formula.MustNewCalculator(config.HandlerTypeStream)

	calct := StreamLoadCalculator{}

	err1 := calct.Init(formula.WithRuntime(conf))

	// give StreamLoadCalculator

	// ctm := calct.(*StreamLoadCalculator)

	fmt.Println(err1)

	// wait the goroutine to finish
	wg.Add(1)
	go func() {
		for d := range sub.DataChan {
			fmt.Println(d.Key, d.Value.Value("510300.XSHG.CS"))
			val := calct.Calculate(d.Key, *d.Value)
			fmt.Println(val)

			// iter the value
			for item := val.Oldest(); item != nil; item = item.Next() {
				for k, v := range item.Value.Data {
					fmt.Println(k, v)
				}
				fmt.Println(item.Value)
			}

		}
		wg.Done()
	}()
	
	wg.Wait()



	
}
