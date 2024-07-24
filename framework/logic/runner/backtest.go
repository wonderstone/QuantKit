package runner

import (
	"fmt"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	model2 "github.com/wonderstone/QuantKit/framework/logic/model"
	"github.com/wonderstone/QuantKit/framework/logic/perfeval"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/perf"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

type Backtest struct {
	Common

	WithCalculator

	startTime time.Time
}

func (t *Backtest) RunMode() config.Mode {
	return config.BTMode
}

func (t *Backtest) Init(creator ...setting.WithResource) error {
	t.Common.Init(creator...)

	config.StatusLog(config.StartingEvent, t.process.GetProgress())
	t.startTime = time.Now()

	config.StatusLog(config.StartingEvent, t.process.GetProgress())

	// 初始化参数
	// 当参数为空时，使用配置文件中的指数作为训练入参
	if len(t.params) == 0 {
		t.params = t.Config().Framework.Indicator
	}

	config.StatusLog(
		config.RunningEvent, t.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("%s初始化完毕", t.Config().ID)},
	)
	return nil
}

func (t *Backtest) calcPerfResult(f handler.Framework) {
	// 读取account.csv 计算收益率
	records := make([]recorder.AssetRecord, 0)

	err := f.Account().(handler.Accounts).GetHistoryAssetRecord(&records)
	if err != nil {
		config.ErrorF("读取回测结果失败，请检查运行参数，可能是由于回测区间无数据引起")
	}

	// err := config.ReadCsvFile(t.Config().Path.AccountResultFile, &records)
	// if err != nil {
	// 	config.ErrorF("读取回测结果失败，请检查运行参数，可能是由于回测区间无数据引起")
	// }

	pe := perfeval.NewPerfEval(records, true)
	tr := pe.CalcPerfEvalResult(
		perfeval.WithPerformanceIndicateType(perf.TotalReturn),
	)

	ar := pe.CalcPerfEvalResult(
		perfeval.WithPerformanceIndicateType(perf.AnnualizedReturn),
	)

	md := pe.CalcPerfEvalResult(
		perfeval.WithPerformanceIndicateType(perf.MaxDrawdown),
	)

	config.StatusLog(
		config.FinishEvent, 100, map[string]any{
			"total_return":  tr - 1.0,
			"annual_return": ar - 1.0,
			"max_drawdown":  md,
			"cost_sec":      time.Since(t.startTime).Seconds(),
		},
	)
}

func (t *Backtest) validFunc(g *genome.Genome) {
	f := t.NewReplay(
		WithResource(t.Resource),
		WithGenomeModel(g),
	)
	t.Quote().Run()
	f.Run()

	t.Quote().WaitForShutdown()

	if !f.IsFinished() {
		config.ErrorF("回测失败")
	}

	t.calcPerfResult(f)
}

func (t *Backtest) validFunc2(gs *genomeset.GenomeSet) {
	f := t.NewReplay(WithGenomeSetModel(gs))
	t.Quote().Run()

	f.Run()

	t.Quote().WaitForShutdown()

	if !f.IsFinished() {
		config.ErrorF("回测失败")
	}

	t.calcPerfResult(f)
}

func (t *Backtest) Start() error {
	return model2.Run(
		t.Config().System.ModelHandlerType,
		model2.WithRunner(t),
		model2.WithModelOption(
			model.WithModelConfig(*t.Config().Model.Gep),
			model.WithKarvaExpressionFile(t.Config().Path.KarvaExpressionFile),
			model.WithIndicator2FormulaIndex(t.Config().Indicator2FormulaVarIndex),
			model.WithNumTerminal(len(t.params)),
			model.WithGenome(t.validFunc),
			model.WithGenomeSet(t.validFunc2),
		),
	)
}

func init() {
	setting.RegisterRunner((*Backtest)(nil), config.BTMode)
}
