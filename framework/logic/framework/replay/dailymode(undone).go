package replay

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	account2 "github.com/wonderstone/QuantKit/framework/logic/account"

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

// DailyMode 日频回放，只能当前根撮合
// 对于日频回测模式，每天下午5点进行结算，可以简化步骤，直接每放完一天就结算一次，单日处理流程如下：
// 1. OnDayOpen
// 2. OnTick
// 3. Match
// 3. OnDayClose
// 4. OnSettle

type DailyMode struct {
	setting.Resource

	trainMode bool // 是否记录
	// config      config.ReplayFramework
	// performance config.Performance
	//
	// base      handler.Basic
	// contract  handler.Indicator
	// indicator handler.Indicator
	// dir       *config.Path

	currIndicators *orderedmap.OrderedMap[string, dataframe.StreamingRecord]

	trainRecorder *recorder.MemoryRecorder[recorder.AssetRecord] // 记录训练结果，用于计算最终的结果
	// OrderRecorder *recorder.MemoryRecorder[account.OrderRecord] // 记录训练结果，用于计算最终的结果

	account handler.Accounts
	matcher handler.Matcher

	strategy handler.Strategy

	gn    *genome.Genome       // 当前基因组, 与gnSet二选一
	gnSet *genomeset.GenomeSet // 当前基因组集合

	evalFunc model.EvaluateFunc

	currTime       time.Time // 当前时间
	nextSettleTime time.Time // 下次结算时间
	beginTime      time.Time // 启动时间
	endTime        time.Time // 停止时间

	ch *handler.Channel

	finish          chan bool
	processChan     chan float64
	enableStatusLog bool
}

func (b *DailyMode) SetEvaluateFunc(f func(values model.InputValues) model.OutputValues) {
	b.evalFunc = f
}

func (b *DailyMode) SetGenome(genome *genome.Genome) {
	b.gn = genome
	b.evalFunc = model.GetEvaluateFunc(genome)
}

func (b *DailyMode) SetGenomeSet(genomeSet *genomeset.GenomeSet) {
	b.gnSet = genomeSet
	b.evalFunc = model.GetEvaluateFunc2(genomeSet)
}

func (b *DailyMode) ConfigStrategyFromFile(file ...string) config.StrategyConfig {
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

func (b *DailyMode) TrainInfo() (*model.TrainInfo, bool) {
	if !b.trainMode {
		return nil, false
	}

	return &model.TrainInfo{
		Genome:    b.gn,
		GenomeSet: b.gnSet,
	}, true
}

func (b *DailyMode) GetPerformance() (result float64) {
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

func (b *DailyMode) CurrTime() *time.Time {
	return &b.currTime
}

func (b *DailyMode) CurrDate() time.Time {
	return time.Date(
		b.currTime.Year(),
		b.currTime.Month(), b.currTime.Day(), 0, 0, 0, 0, b.currTime.Location(),
	)
}

func (b *DailyMode) SubscribeData() {
	b.ch = b.Quote().Subscribe()
}

func (b *DailyMode) Run() {
	// defer config.CatchPanic()

	// 初始化
	b.strategy.OnInitialize(b)
	endFlag := false
	for {
		select {
		case d, ok := <-b.ch.DataChan:
			if !ok || endFlag {
				// config.InfoF("数据回放完成")
				// 结束释放资源
				b.account.Release()

				b.strategy.OnEnd(b)

				// config.InfoF("回测结束")
				b.finish <- true
				close(b.finish)
				return
			}

			if d.Key.Before(b.beginTime) {
				continue
			}

			if d.Key.After(b.endTime) {
				close(b.ch.StopChan)
				endFlag = true
				continue
			}

			b.currTime = time.Date(
				d.Key.Year(),
				d.Key.Month(),
				d.Key.Day(), 8, 0, 0, 0,
				d.Key.Location(),
			)

			b.strategy.OnDailyOpen(b, config.MarketTypeStock, b.Account().GetAccount(config.MarketTypeStock)...)
			b.strategy.OnDailyOpen(b, config.MarketTypeFuture, b.Account().GetAccount(config.MarketTypeFuture)...)

			// 设置当前时间
			b.currTime = d.Key

			orders := b.strategy.OnTick(b, d.Key, *d.Value)
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

			// 撮合
			b.account.DoMatch(d.Key, *d.Value)
			// 计算当前持仓的指标
			b.account.CalcPositionPnL(d.Key, *d.Value)

			b.strategy.OnDailyClose(b, b.Account().GetAccounts())
			b.account.DoSettle(b.currTime, b.currIndicators)

		}
	}
}

func (b *DailyMode) IsFinished() bool {
	select {
	case v, ok := <-b.finish: // 从ch1接收数据
		if !ok {
			return false
		}

		return v
	}
}

func (b *DailyMode) Matcher() handler.Matcher {
	return b.matcher
}

func (b *DailyMode) Init(resource setting.Resource, options ...setting.WithResource) error {
	// 初始化账户处理器
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
	b.currTime = b.beginTime.Add(time.Hour * 8)

	// 设置一下结算时间和完全结束时间
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
		)

		b.trainRecorder = recorder.NewMemoryRecorder[recorder.AssetRecord]()
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

	// 设置撮合器，对于日频及以上回测，只能当前根撮合
	b.matcher = setting.MustNewMatcher(
		config.HandlerTypeCurrMatch,
		handler.WithStockSlippage(b.Config().Framework.Stock.Slippage),
		handler.WithFutureSlippage(b.Config().Framework.Future.Slippage),
	)

	b.finish = make(chan bool, 1)
	return nil
}

func (b *DailyMode) SetStrategy(strategy handler.Strategy) {
	b.strategy = strategy
}

func (b *DailyMode) Evaluate(values model.InputValues) model.OutputValues {
	return b.evalFunc(values)
}

func init() {
	setting.RegisterReplay(&DailyMode{}, config.HandlerTypeDailyMode)
}
