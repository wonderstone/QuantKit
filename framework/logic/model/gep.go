package model

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/tools/qk"
	"gopkg.in/yaml.v3"
)

type Gep struct {
	*Op

	handler model.Handler
}

func (g *Gep) Run(option ...WithOption) error {
	g.Op = NewOp(option...)

	g.handler = model.NewHandler(g.runner.Config().Model.Gep.Mode, model.WithOp(g.gepOps))

	switch g.runner.RunMode() {
	case config.TrainMode:
		g.runTrain()
	case config.BTMode:
		g.runBT()
	case config.RunMode:
		g.runRT()
	default:
		return qk.ErrInvalidRunMode{Mode: g.runner.RunMode()}
	}

	return nil
}

func (g *Gep) runTrain() {
	// 初始化模型处理器
	conf := g.runner.Config()
	g.handler = model.NewHandler(
		conf.Model.Gep.Mode,
		model.WithOp(g.gepOps),
	)

	// 开始演化
	modelRecord := g.handler.Evolve()
	modelRecord.TrainId = conf.ID
	modelRecord.TrainName = conf.Name
	modelRecord.ModelId = conf.Model.ID
	modelRecord.ModelName = conf.Model.Name

	// 写入模型记录文件
	data, err := yaml.Marshal(modelRecord)
	if err != nil {
		config.ErrorF("序列化模型记录失败: %s", err)
	}

	// 没有则创建目录
	dir := filepath.Dir(conf.Path.KarvaExpressionFile)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			config.ErrorF("创建模型记录目录失败: %s", err)
		}
	}

	err = os.WriteFile(conf.Path.KarvaExpressionFile, data, 0644)

	config.StatusLog(
		config.FinishEvent, 100,
		map[string]any{"msg": fmt.Sprintf(`训练完成, 模型输出到文件：%s`, conf.Path.KarvaExpressionFile)},
	)
}

func (g *Gep) runBT() {
	// 初始化模型
	conf := g.runner.Config()

	g.handler = model.NewHandler(
		conf.Model.Gep.Mode,
		model.WithOp(g.gepOps),
	)

	finish := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-finish:
				return
			case <-time.After(100 * time.Millisecond):
				config.StatusLog(
					config.RunningEvent, g.runner.GetProgress(),
					map[string]any{"msg": fmt.Sprintf("回测中")},
				)
			}
		}
	}()

	_ = g.handler.RunOnce()

	finish <- true
}

func (g *Gep) runRT() {
	// 初始化模型
	conf := g.runner.Config()

	g.handler = model.NewHandler(
		conf.Model.Gep.Mode,
		model.WithOp(g.gepOps),
	)

	finish := make(chan bool, 1)

	_ = g.handler.RunOnce()

	finish <- true
}

func init() {
	RegisterNewHandler(
		&Gep{},
		config.HandlerTypeGepMode,
		config.HandlerTypeDefault,
	)
}
