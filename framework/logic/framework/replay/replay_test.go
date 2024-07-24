package replay

// import (
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/require"
// 	"github.com/wonderstone/QT/framework/entity/account"
// 	"github.com/wonderstone/QT/framework/entity/handler"
// 	config2 "github.com/wonderstone/QT/framework/logic/contract/config"
// 	_ "github.com/wonderstone/QT/framework/logic/indicator"
// 	"github.com/wonderstone/QT/modelgene/gep/genome"
// 	"github.com/wonderstone/QT/tools/config"
// 	"github.com/wonderstone/QT/tools/container/orderedmap"
// 	"github.com/wonderstone/QT/tools/dataframe"
// )

// type MyStrategy struct {
// 	acc account.Account
// }

// var nowDirection = config.OrderBuy
// var nextDirection = config.OrderSell

// func (m *MyStrategy) OnTick(
// 	framework handler.Framework, tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
// ) (orders []account.Order) {

// 	for p := indicators.Oldest(); p != nil; p = p.Next() {
// 		order, err := m.acc.NewOrder(
// 			p.Key, 100,
// 			account.WithOrderPrice(p.Value.ConvertToFloat("Close")),
// 			account.WithOrderDirection(nowDirection),
// 		)

// 		if err != nil {
// 			config.WarnF("创建新单失败: %s", err)
// 			continue
// 		}

// 		orders = append(orders, order)
// 	}

// 	nowDirection, nextDirection = nextDirection, nowDirection

// 	return
// }

// func (m *MyStrategy) OnInitialize(framework handler.Framework) {
// 	m.acc = framework.AccountHandler().GetAccountByID("stock")
// }

// func (m *MyStrategy) OnStart(framework handler.Framework) {
// }

// func (m *MyStrategy) OnDailyOpen(framework handler.Framework, marketType config.MarketType, acc ...account.Account) {

// }

// func (m *MyStrategy) OnDailyClose(framework handler.Framework, acc map[string]account.Account) {

// }

// func (m *MyStrategy) OnEnd(framework handler.Framework) {
// }

// func TestTrain(t *testing.T) {
// // 目录
// p := config.NewDefaultPath(
// "train",
// config.WithRoot("../../../tools/config/template"),
// )

// config.NewLogger("train", p.InfoFile, p.ErrorFile, p.StatusFile)

// confMain := config.New("train", p)

// indicatorHandler := handler.NewHandler("full-loader")
// err := indicatorHandler.Init(
// // indicator.WithIndicator(confMain.Framework.Indicator),
// // indicator.WithStockInstID(confMain.Framework.Instrument),
// // indicator.WithIndicatorDataPath("/Users/alexxiong/GolandProjects/quant/strategy/S20230824-090000-000/.vqt/train/T20230824-110000-000/result"),
// )
// 	require.NoError(t, err)

// indicatorHandler.LoadData() // 加载数据

// contractHandler := config2.NewConfigContractHandle(*confMain.Contract)

// cs := make(chan *genome.Genome, 20)

// prepareFunc := func(genome2 *genome.Genome) NextMode {
// f := NextMode{}

// f.SetStrategy(&MyStrategy{})
// err = f.Init(confMain, indicatorHandler, contractHandler, p)
// f.SubscribeData()

// return f
// }

// validFunc := func(f NextMode, c chan *genome.Genome) {
// go f.Run()
// if f.IsFinished() {
// c <- &genome.Genome{
// Score: f.GetPerformance(),
// }
// 		} else {
// config.ErrorF("回测失败")
// }
// 		return
// }

// for i := 0; i < 20; i++ {
// f := prepareFunc(&genome.Genome{})
// go validFunc(f, cs)
// }

// indicatorHandler.Replay()
// indicatorHandler.Close()

// for i := 0; i < 20; i++ {
// g := <-cs
// t.Log(g.Score)
// }

// require.NoError(t, err)
// }

// func TestBT(t *testing.T) {
// // 目录
// p := &config.Path{
// IndicatorFile:      "../../../tools/config/template/indicator.yaml",
// ContractFile:       "../../../tools/config/template/contract.yaml",
// ModelConfigFile:    "../../../tools/config/template/model.yaml",
// TrainConfigFile:    "../../../tools/config/template/train.yaml",
// BackTestConfigFile: "../../../tools/config/template/backtest.yaml",
// InfoFile:           "../../../tools/config/template/log/info.log",
// ErrorFile:          "../../../tools/config/template/log/error.log",
// OrderResultFile:    "../../../tools/config/template/result/order.csv",
// PositionResultFile: "../../../tools/config/template/result/position.csv",
// AccountResultFile:  "../../../tools/config/template/result/account.csv",
// Root:               "../../../tools/config",
// Common:             "../../../tools/config/template",
// }

// config.NewLogger("bt", p.InfoFile, p.ErrorFile, p.StatusFile)

// confMain := config.New("bt", p)

// indicatorHandler := handler.NewHandler("full-loader")

// err := indicatorHandler.Init(
// // indicator.WithIndicator(confMain.Framework.Indicator),
// // indicator.WithStockInstID(confMain.Framework.Instrument),
// // indicator.WithIndicatorDataPath("/Users/alexxiong/GolandProjects/quant/strategy/S20230824-090000-000/.vqt/train/T20230824-110000-000/result"),
// )
// 	require.NoError(t, err)

// indicatorHandler.LoadData() // 加载数据

// contractHandler := config2.NewConfigContractHandle(*confMain.Contract)

// f := NextMode{}

// f.SetStrategy(&MyStrategy{})
// err = f.Init(confMain, indicatorHandler, contractHandler, p)
// require.NoError(t, err)

// f.SubscribeData()

// go f.Run()

// f.indicator.Replay()

// f.indicator.Close()

// f.IsFinished()

// require.NoError(t, err)
// }
