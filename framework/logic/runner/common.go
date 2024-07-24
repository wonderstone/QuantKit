package runner

import (
	"fmt"

	"github.com/wonderstone/QuantKit/framework/entity/handler"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/base"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/framework/entity/quote"
	_ "github.com/wonderstone/QuantKit/framework/logic/base"
	_ "github.com/wonderstone/QuantKit/framework/logic/contract"
	_ "github.com/wonderstone/QuantKit/framework/logic/framework"
	_ "github.com/wonderstone/QuantKit/framework/logic/indicator"
	_ "github.com/wonderstone/QuantKit/framework/logic/quote"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
	"github.com/wonderstone/QuantKit/tools/common"
)

type Common struct {
	*setting.Resource

	process *common.Process

	params []string
}

func (r *Common) RunMode() config.Mode {
	return r.Config().Mode
}

type WithCalculator struct {
	calculator formula.CalculatorModule
}

type RunOp struct {
	resource  handler.Resource
	genome    *genome.Genome
	genomeSet *genomeset.GenomeSet
}

type WithRunOption func(op *RunOp)

func WithResource(r handler.Resource) WithRunOption {
	return func(op *RunOp) {
		op.resource = r
	}
}

func WithGenomeModel(genome *genome.Genome) WithRunOption {
	return func(op *RunOp) {
		op.genome = genome
	}
}

func WithGenomeSetModel(genomeSet *genomeset.GenomeSet) WithRunOption {
	return func(op *RunOp) {
		op.genomeSet = genomeSet
	}
}

func NewRunOp(option ...WithRunOption) *RunOp {
	r := &RunOp{}

	for _, runOption := range option {
		runOption(r)
	}

	return r
}

func (r *Common) SetGEPInputParams(params []string) {
	r.params = params
}

func (r *Common) NewReplay(option ...WithRunOption) setting.ReplayFramework {
	op := NewRunOp(option...)
	var f setting.ReplayFramework
	// 如果没有设置回测类型，则根据回测频率设置回测类型
	if r.Config().System.ReplayHandlerType == config.HandlerTypeDefault {
		f = setting.MustNewReplay(config.HandlerTypeNextMatch)
	} else {
		f = setting.MustNewReplay(r.Config().System.ReplayHandlerType)
	}

	if err := f.Init(
		*r.Resource,
	); err != nil {
		config.ErrorF("初始化运行状态失败, %e", err)
	}

	if op.genomeSet != nil {
		f.SetGenomeSet(op.genomeSet)
	} else if op.genome != nil {
		f.SetGenome(op.genome)
	}

	f.SetStrategy(r.Creator()())
	f.SubscribeData()

	return f
}

func (r *Common) NewRealtime(option ...WithRunOption) setting.ReplayFramework {
	op := NewRunOp(option...)
	var f setting.RTFramework
	// 如果没有设置回测类型，则根据回测频率设置回测类型
	if r.Config().System.ReplayHandlerType == config.HandlerTypeDefault {
		f = setting.MustNewRTExecutor(config.HandlerTypeNextMatch)
	} else {
		f = setting.MustNewRTExecutor(r.Config().System.ReplayHandlerType)
	}

	if err := f.Init(
		*r.Resource,
	); err != nil {
		config.ErrorF("初始化运行状态失败, %e", err)
	}

	if op.genomeSet != nil {
		f.SetGenomeSet(op.genomeSet)
	} else if op.genome != nil {
		f.SetGenome(op.genome)
	}

	f.SetStrategy(r.Creator()())
	f.SubscribeData()

	return f
}

func (r *Common) newProcess() {
	r.process = common.NewProcess(1, 0, 94.99)
}

func (r *Common) GetProgress() float64 {
	return r.process.GetProgress()
}

func (r *Common) newContract() {
	config.StatusLog(
		config.StartingEvent, r.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("初始化合约")},
	)

	setting.WithContractHandler(
		setting.MustNewContractHandler(
			r.Config().System.ContractHandlerType,
			contract.WithProperty(r.Config().Contract),
		),
	)(
		r.Resource,
	)

	// 先加载一部分标的的合约信息
	if len(r.Config().Framework.Instrument) == 0 {
		for _, instId := range r.Config().Framework.Instrument {
			r.Contract().GetContract(instId)
		}
	}

}

func (r *Common) newQuote() {
	config.StatusLog(
		config.StartingEvent, r.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("初始化指标: %s", r.Config().ID)},
	)

	quoteType := config.HandlerTypeReplay
	if r.Config().Mode == config.RunMode {
		quoteType = config.HandlerTypeRealtime
	}

	setting.WithQuoteHandler(
		setting.MustNewQuote(
			quoteType,
			quote.WithConfig(*r.Config()),
		),
	)(
		r.Resource,
	)

	config.StatusLog(
		config.StartingEvent, r.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("启动数据加载...")},
	)

	r.Quote().LoadData()

	config.StatusLog(
		config.StartingEvent, r.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("初始化模型")},
	)
}

func (t *WithCalculator) runCalc(process *common.Process, options *config.Runtime) {
	config.StatusLog(
		config.StartingEvent, process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("初始化指标计算器")},
	)

	t.calculator = formula.MustNewCalculator(options.System.FormulaHandlerType)

	err := t.calculator.Init(
		// formula.WithInstID(options.Framework.Instrument),
		// formula.WithDir(options.Path),
formula.WithRuntime(*options),
	)

	if err != nil {
		config.ErrorF("初始化指标处理器失败: %s", err)
	}

	defer func() {
		if err := recover(); err != nil {
			config.ErrorF("指标计算启动失败: %s", err)
		}
	}()

	config.StatusLog(
		config.StartingEvent, process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("指标计算中")},
	)

	t.calculator.StartCalc()
	config.StatusLog(
		config.StartingEvent, process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("指标计算完成, 指标输出到: %s", options.Path.Indicator)},
	)
}

func (r *Common) newBasic() {
	config.StatusLog(
		config.StartingEvent, r.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("初始化基础数据")},
	)

	setting.WithBaseHandler(
		setting.MustBaseNewHandler(
			config.HandlerTypeFullLoad,
			r.Resource,
			base.WithXrxdHandler(r.Config().System.XrxdHandlerType),
			base.WithPath(r.Config().Path),
			base.WithInstID(
				func() []string {
					// 如果有组合合约设置，则不能设置合约
					if len(r.Config().Framework.GroupInstrument) > 0 {
						return []string{}
					}

					return r.Config().Framework.Instrument
				}(),
			),
			base.WithTimeRange(r.Config().Framework.Begin, r.Config().Framework.End),
		),
	)(r.Resource)
}

func (r *Common) ConfigStrategyFromFile(file ...string) config.StrategyConfig {
	config.ErrorF("计算模式(calc)下不支持从文件中获取策略配置")
	return config.StrategyConfig{}
}

func (r *Common) Init(creator ...setting.WithResource) {
	r.newProcess()
	config.StatusLog(
		config.StartingEvent, r.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("初始化")},
	)

	r.Resource = setting.NewResource(
		creator...,
	)

	config.StatusLog(
		config.StartingEvent, r.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("调用用户自定义初始化配置")},
	)

	// 调用全局初始化函数，设置的优先级高于配置文件
	r.Creator()().OnGlobalOnce(r)

	// 初始化基础数据处理器
	r.newBasic()

	// 初始化合约处理器
	r.newContract()

	// 初始化行情处理器
	r.newQuote()
}
