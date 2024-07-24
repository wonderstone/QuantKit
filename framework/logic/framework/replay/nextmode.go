package replay

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	matcher2 "github.com/wonderstone/QuantKit/framework/logic/matcher"

	account2 "github.com/wonderstone/QuantKit/framework/logic/account"
	"github.com/wonderstone/QuantKit/framework/logic/indicator"
	"github.com/wonderstone/QuantKit/framework/logic/perfeval"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/perf"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

type NextMode struct {
	setting.Resource

	trainMode bool // 是否记录

	currCloseTick orderedmap.OrderedMap[string, dataframe.StreamingRecord]

	trainRecorder *recorder.MemoryRecorder[recorder.AssetRecord] // 记录训练结果，用于计算最终的结果

	account handler.Accounts
	matcher handler.Matcher

	strategy handler.Strategy

	gn    *genome.Genome       // 当前基因组, 与gnSet二选一
	gnSet *genomeset.GenomeSet // 当前基因组集合

	evalFunc model.EvaluateFunc

	currTime           time.Time // 当前时间
	nextMarketOpenTime time.Time // 下次开盘前时间
	nextSettleTime     time.Time // 下次结算时间
	beginTime          time.Time // 开始时间
	endTime            time.Time // 结束时间

	ch *handler.Channel

	calc indicator.StreamLoadCalculator

	finish      chan bool
	processChan chan float64
}

func (b *NextMode) SetEvaluateFunc(f func(values model.InputValues) model.OutputValues) {
	b.evalFunc = f
}

func (b *NextMode) SetGenome(genome *genome.Genome) {
	b.gn = genome
	b.evalFunc = model.GetEvaluateFunc(genome)
}

func (b *NextMode) SetGenomeSet(genomeSet *genomeset.GenomeSet) {
	b.gnSet = genomeSet
	b.evalFunc = model.GetEvaluateFunc2(genomeSet)
}

func (b *NextMode) ConfigStrategyFromFile(file ...string) config.StrategyConfig {
	f := b.Dir().StrategyFile
	if len(file) != 0 {
		f = file[0]
	}

	strategyConfig, err := config.NewStrategyConfig(f)
	if err != nil {
		config.ErrorF("读取策略配置文件失败: %v", err)
	}

	return strategyConfig
}

func (b *NextMode) TrainInfo() (*model.TrainInfo, bool) {
	if !b.trainMode {
		return nil, false
	}

	return &model.TrainInfo{
		Genome:    b.gn,
		GenomeSet: b.gnSet,
	}, true
}

func (b *NextMode) GetPerformance() (result float64) {
	if !b.trainMode {
		config.ErrorF("非训练的模式下，无法计算回测结果")
	}

	pe := perfeval.NewPerfEval(b.trainRecorder.GetRecord(), true)
	result = pe.CalcPerfEvalResult(
		perfeval.WithPerformanceIndicateType(perf.IndicateType(b.Config().Performance.PerformanceType)),
		perfeval.WithRiskFreeRate(b.Config().Performance.RiskFreeRate),
	)
	return
}

func (b *NextMode) CurrTime() *time.Time {
	return &b.currTime
}

func (b *NextMode) CurrDate() time.Time {
	return time.Date(
		b.currTime.Year(),
		b.currTime.Month(), b.currTime.Day(), 0, 0, 0, 0, b.currTime.Location(),
	)
}

func (b *NextMode) SubscribeData() {
	b.ch = b.Quote().Subscribe()
}

func (b *NextMode) Run() {
	// defer config.CatchPanic()

	// 初始化
	b.strategy.OnInitialize(b)

	endFlag := false
	for {
		select {
		case d, ok := <-b.ch.DataChan:
			if !ok || endFlag {
				b.currTime = b.nextSettleTime

				b.account.DoSettle(b.currTime, &b.currCloseTick)
				b.strategy.OnDailyClose(b, b.Account().GetAccounts())

				// 结束释放资源
				b.account.Release()

				b.strategy.OnEnd(b)

				// config.InfoF("回测结束")
				b.finish <- true
				close(b.finish)
				return
			}

			// 跳过开始日期以前的数据
			if d.Key.Before(b.beginTime) {
				continue
			}

			// 判断时间，如果时间大于开盘时间了，就认为需要开盘了
			if d.Key.After(b.nextMarketOpenTime) {
				b.currTime = time.Date(
					d.Key.Year(),
					d.Key.Month(),
					d.Key.Day(), 8, 0, 0, 0,
					d.Key.Location(),
				)

				b.strategy.OnDailyOpen(b, config.MarketTypeStock, b.Account().GetAccount(config.MarketTypeStock)...)

				b.nextMarketOpenTime = time.Date(
					d.Key.Year(),
					d.Key.Month(),
					d.Key.Day(), 8, 0, 0, 0,
					d.Key.Location(),
				).AddDate(0, 0, 1)
			}

			// 判断时间，如果时间大于下午5点，就认为需要结算了
			endFlag = d.Key.After(b.endTime)
			if d.Key.After(b.nextSettleTime) || endFlag {
				b.currTime = b.nextSettleTime

				b.account.DoSettle(b.currTime, &b.currCloseTick)
				// @ 进行指标计算结算
				b.calc.DoSettle(b.currTime, b.Resource.Base())
				b.strategy.OnDailyClose(b, b.Account().GetAccounts())

				b.nextSettleTime = time.Date(
					d.Key.Year(),
					d.Key.Month(),
					d.Key.Day(), 16, 0, 0, 0,
					d.Key.Location(),
				)

				if endFlag {
					close(b.ch.StopChan)
					continue
				}
			}
			// 先撮合
			// start := time.Now()
			b.account.DoMatch(d.Key, *d.Value)
			// fmt.Printf("撮合耗时：%v\n", time.Since(start))

			// 设置当前时间
			b.currTime = d.Key

			// 计算当前持仓的指标
			b.account.CalcPositionPnL(d.Key, *d.Value)

			if b.Config().Framework.Frequency != config.Frequency1Day || b.CurrDate().Add(b.Config().Framework.DailyTriggerTime).Equal(d.Key) {
				// 计算指标
				indicate := b.calc.Calculate(d.Key, *d.Value)

				orders := b.strategy.OnTick(b, d.Key, indicate)
				for _, o := range orders {
					err := b.Account().InsertOrder(o)
					if err != nil {
						config.WarnF("插入订单失败: %s", err)
					} else {
						config.DebugF(
							"插入订单: %f, %f, %s, %s", o.OrderQty(), o.OrderPrice(), o.OrderDirection(),
							o.OrderTime().String(),
						)
					}
				}
			}

			b.currCloseTick = *d.Value
		}
	}
}

func (b *NextMode) IsFinished() bool {
	select {
	case v, ok := <-b.finish: // 从ch1接收数据
		if !ok {
			return false
		}

		return v
	}
}

func (b *NextMode) Matcher() handler.Matcher {
	return b.matcher
}

func (b *NextMode) Init(resource setting.Resource, options ...setting.WithResource) error {
	b.Resource = resource

	b.account = account2.NewDefaultHandler()
	options = append(options, setting.WithAccountHandler(b.account))
	options = append(options, setting.WithFramework(b))

	for _, option := range options {
		option(&b.Resource)
	}

	b.account.SetResource(b.Resource)

	b.evalFunc = func(values model.InputValues) model.OutputValues {
		return model.OutputValues(values)
	}

	if b.Config().Mode == config.TrainMode {
		b.trainMode = true
	} else {
		b.trainMode = false
	}

	b.beginTime = b.Config().Framework.Begin
	b.endTime = b.Config().Framework.End.Add(time.Hour * 17)

	// 设置一下开盘时间
	b.nextMarketOpenTime = b.beginTime.Add(time.Hour * 8)

	// 设置一下结算时间
	b.nextSettleTime = b.beginTime.Add(time.Hour * 17)

	if !b.trainMode {
		recorders := make(map[config.RecordType]recorder.Handler)
		recorders[config.RecordTypeAsset] = recorder.NewRecorder[recorder.AssetRecord](
			b.Config().System.RecordHandlerType,
			recorder.WithFilePath(b.Dir().AccountResultFile),
		)

		recorders[config.RecordTypePosition] = recorder.NewRecorder[recorder.PositionRecord](
			b.Config().System.RecordHandlerType,
			recorder.WithFilePath(b.Dir().PositionResultFile),
		)

		recorders[config.RecordTypeOrder] = recorder.NewRecorder[recorder.OrderRecord](
			b.Config().System.RecordHandlerType,
			recorder.WithFilePath(b.Dir().OrderResultFile),
			// recorder.WithPlusMode(),
		)

		err := b.account.Init(b, recorders)
		if err != nil {
			return err
		}
	} else {
		recorders := make(map[config.RecordType]recorder.Handler)
		b.trainRecorder = recorder.NewMemoryRecorder[recorder.AssetRecord]()

		// 只记录资产
		recorders[config.RecordTypeAsset] = b.trainRecorder

		err := b.account.Init(b, recorders)
		if err != nil {
			return err
		}
	}

	stockSlippage := b.Config().Framework.Stock.Slippage
	futureSlippage := b.Config().Framework.Future.Slippage

	b.matcher = &matcher2.NextQuote{StockSlippage: stockSlippage, FutureSlippage: futureSlippage}
	b.finish = make(chan bool, 1)

	// 指标计算模块加载
	b.calc = indicator.StreamLoadCalculator{}
	err := b.calc.Init(formula.WithRuntime(*b.Config()))
	if err != nil {
		return err
	}

	return nil
}

func (b *NextMode) SetStrategy(strategy handler.Strategy) {
	b.strategy = strategy
}

func (b *NextMode) Evaluate(values model.InputValues) model.OutputValues {
	return b.evalFunc(values)
}

func init() {
	setting.RegisterReplay(&NextMode{}, config.HandlerTypeNextMatch)
}
