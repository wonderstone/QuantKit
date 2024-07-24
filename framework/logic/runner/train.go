package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/wonderstone/QuantKit/config"
	model2 "github.com/wonderstone/QuantKit/framework/logic/model"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
)

type Train struct {
	Common
	WithCalculator
}

func (t *Train) RunMode() config.Mode {
	return config.TrainMode
}

func (t *Train) Init(creator ...setting.WithResource) error {
	t.Common.Init(creator...)

	// 初始化参数
	// 当参数为空时，使用配置文件中的指数作为训练入参
	if len(t.params) == 0 {
		t.params = t.Config().Framework.Indicator
	}

	config.StatusLog(
		config.RunningEvent, t.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("初始化完毕: %s", t.Config().ID)},
	)

	return nil
}

func (t *Train) validFunc(iterate int, gs []*genome.Genome) (*genome.Genome, bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	doneChan := make(chan struct{})

	fs := make([]setting.ReplayFramework, len(gs))

	config.StatusLog(
		config.RunningEvent,
		t.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("迭代: %d / %d", iterate, t.Config().Model.Gep.Iteration)},
	)

	wg := sync.WaitGroup{}
	for i := range gs {
		fs[i] = t.NewReplay(WithGenomeModel(gs[i]))
	}

	t.Quote().Run()

	for i := range gs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			fs[index].Run()
		}(i)
	}

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		t.Quote().WaitForShutdown()

		var bestGenome *genome.Genome = nil
		var bestGenomeIndex = 0

		for i := 0; i < len(gs); i++ {
			if fs[i].IsFinished() {
				gs[i].Score = fs[i].GetPerformance()
			} else {
				config.ErrorF("回测失败")
			}

			if bestGenome == nil || gs[i].Score > bestGenome.Score {
				bestGenome = gs[i]
				bestGenomeIndex = i
			}
		}

		config.InfoF(
			"第%d代最优基因组序号: %d, 得分: %f, Exp: %s", iterate, bestGenomeIndex+1, bestGenome.Score,
			bestGenome.StringSlice(),
		)

		return bestGenome, bestGenome.Score >= t.Config().Performance.ExpectFitness
	case <-ctx.Done():
		config.InfoF(
			"中断，当前最优基因组序号: %d, 得分: %f, Exp: %s", 1, gs[0].Score,
			gs[0].StringSlice(),
		)

		return gs[0], true
	}
}

func (t *Train) validFunc2(iterate int, gs []*genomeset.GenomeSet) (*genomeset.GenomeSet, bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	doneChan := make(chan struct{})

	fs := make([]setting.ReplayFramework, len(gs))

	config.StatusLog(
		config.RunningEvent,
		t.process.GetProgress(),
		map[string]any{"msg": fmt.Sprintf("迭代: %d / %d", iterate+1, t.Config().Model.Gep.Iteration)},
	)

	wg := sync.WaitGroup{}
	for i := range gs {
		fs[i] = t.NewReplay(WithGenomeSetModel(gs[i]))
	}

	t.Quote().Run()

	for i := range gs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			fs[index].Run()
		}(i)
	}

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		t.Quote().WaitForShutdown()

		var bestGenome *genomeset.GenomeSet = nil
		var bestGenomeIndex int = 0

		for i := 0; i < len(gs); i++ {
			if fs[i].IsFinished() {
				gs[i].Score = fs[i].GetPerformance()
			} else {
				config.ErrorF("回测失败")
			}

			if bestGenome == nil || gs[i].Score > bestGenome.Score {
				bestGenome = gs[i]
				bestGenomeIndex = i
			}
		}

		config.InfoF(
			"第%d代最优基因组序号: %d, 得分: %f, Exp: %s", iterate, bestGenomeIndex+1, bestGenome.Score,
			bestGenome.StringSlice(),
		)
		return bestGenome, bestGenome.Score >= t.Config().Performance.ExpectFitness
	case <-ctx.Done():
		config.InfoF(
			"中断，当前最优基因组序号: %d, 得分: %f, Exp: %s", 1, gs[0].Score,
			gs[0].StringSlice(),
		)

		return gs[0], true
	}
}

func (t *Train) Start() error {
	if err := model2.Run(
		t.Config().System.ModelHandlerType,
		model2.WithRunner(t),
		model2.WithModelOption(
			model.WithModelConfig(*t.Config().Model.Gep),
			model.WithIndicator2FormulaIndex(t.Config().Indicator2FormulaVarIndex),
			model.WithNumTerminal(len(t.params)),
			model.WithPerformance(t.validFunc),
			model.WithPerformanceSet(t.validFunc2),
		),
	); err != nil {
		config.ErrorF("模型训练失败: %s", err)
	}

	return nil
}

func (t *Train) Exit() {

}

func init() {
	setting.RegisterRunner((*Train)(nil), config.TrainMode)
}
