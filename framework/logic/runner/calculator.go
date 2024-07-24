package runner

import (
	"fmt"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	_ "github.com/wonderstone/QuantKit/framework/logic/formula"
	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/config"
)

type Calculator struct {
	handler.Resource

	handler formula.CalculatorModule
}

func (c *Calculator) Init(sources ...setting.WithResource) error {
	c.Resource = setting.NewResource(sources...)
	c.handler = formula.MustNewCalculator(c.Config().System.FormulaHandlerType)
	config.StatusLog(config.StartingEvent, 0)

	return c.handler.Init(
		formula.WithRuntime(*c.Config()),
	)
}

func (c *Calculator) SetGEPInputParams(params []string) {
	// do nothing
}

func (c *Calculator) ConfigStrategyFromFile(file ...string) config.StrategyConfig {
	config.ErrorF("计算模式(calc)下不支持从文件中获取策略配置")
	return config.StrategyConfig{}
}

func (c *Calculator) GetProgress() float64 {
	return 0.0
}

func (c *Calculator) RunMode() config.Mode {
	return config.CalcMode
}

func (c *Calculator) SetStrategyCreator(creator handler.StrategyCreator) {}

func (c *Calculator) Start() error {
	config.StatusLog(config.RunningEvent, 20)
	c.handler.StartCalc()
	config.StatusLog(
		config.FinishEvent, 100,
		map[string]any{"msg": fmt.Sprintf("指标计算完成, 指标输出到: %s", c.Config().Path.Indicator)},
	)

	return nil
}

func init() {
	setting.RegisterRunner((*Calculator)(nil), config.CalcMode)
}
