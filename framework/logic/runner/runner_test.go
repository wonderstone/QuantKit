package runner

import (
	// "testing"
	"time"

	// "github.com/stretchr/testify/require"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"

	// "github.com/wonderstone/QuantKit/framework/logic/strategy"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
)

type SimplyStrategy struct {
}

func (s SimplyStrategy) OnInitialize(framework handler.Framework) {

}

func (s SimplyStrategy) OnStart(framework handler.Framework) {
}

func (s SimplyStrategy) OnDailyOpen(
	framework handler.Framework, marketType config.MarketType, acc ...account.Account,
) {
}

func (s SimplyStrategy) OnTick(
	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) (orders []handler.Order) {
	for p := indicators.Oldest(); p != nil; p = p.Next() {
		_, err := framework.Account().NewOrder(
			"stock", p.Key, 100, account.WithOrderPrice(p.Value.ConvertToFloat("Close")),
		)
		if err != nil {
			return
		}
	}

	return
}

func (s SimplyStrategy) OnDailyClose(framework handler.Framework, acc map[string]account.Account) {
}

func (s SimplyStrategy) OnEnd(framework handler.Framework) {
}

// func TestTrain_Init(t *testing.T) {
// 	trainRunner, err := setting.NewRunner("train")
// 	require.NoError(t, err)

// 	dir := config.NewDefaultPath("train")
// 	config.NewLogger("train", dir.InfoFile, dir.ErrorFile, dir.StatusFile)

// 	confMain := config.New("train", dir)

// 	err = trainRunner.Init(confMain)
// 	require.NoError(t, err)

// 	trainRunner.SetStrategyCreator(
// 		func() handler.Strategy {
// 			return &strategy.T0{}
// 		},
// 	)

// 	err = trainRunner.Start()
// 	require.NoError(t, err)
// }

// func TestBT_Init(t *testing.T) {
// 	trainRunner, err := setting.NewRunner("bt")
// 	require.NoError(t, err)

// 	p := config.NewDefaultPath("bt")
// 	config.NewLogger("train", p.InfoFile, p.ErrorFile, p.StatusFile)

// 	confMain := config.New("bt", p)
// 	require.NoError(t, err)

// 	err = trainRunner.Init(confMain)
// 	require.NoError(t, err)

// 	trainRunner.SetStrategyCreator(
// 	func() handler.Strategy {
// 	return &strategy.T0{}
// 	},
// 	)

// 	err = trainRunner.Start()
// 	require.NoError(t, err)
// }
