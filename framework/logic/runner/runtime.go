package runner

import (
	"fmt"
	"time"

	"github.com/wonderstone/QuantKit/config"
	model2 "github.com/wonderstone/QuantKit/framework/logic/model"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
)

// Realtime 实时运行
type Realtime struct {
	Common

	startTime time.Time
}

func (t *Realtime) RunMode() config.Mode {
	return config.RunMode
}

func (t *Realtime) Init(creator ...setting.WithResource) error {
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

func (t *Realtime) validFunc(g *genome.Genome) {
	f := t.NewRealtime(
		WithResource(t.Resource),
		WithGenomeModel(g),
	)
	t.Quote().Run()
	f.Run()

	t.Quote().WaitForShutdown()

	if !f.IsFinished() {
		config.ErrorF("失败")
	}
}

func (t *Realtime) validFunc2(gs *genomeset.GenomeSet) {
	f := t.NewRealtime(WithGenomeSetModel(gs))
	t.Quote().Run()

	f.Run()

	t.Quote().WaitForShutdown()

	if !f.IsFinished() {
		config.ErrorF("回测失败")
	}
}

func (t *Realtime) Start() error {
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
	setting.RegisterRunner((*Realtime)(nil), config.RunMode)
}
