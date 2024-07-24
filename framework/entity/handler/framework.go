package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/modelgene/gep/model"
)

type Framework interface {
	Resource

	// CurrTime 获取当前时间
	CurrTime() *time.Time

	// CurrDate 获取当前日期
	CurrDate() time.Time

	// TrainInfo 获取训练信息
	TrainInfo() (*model.TrainInfo, bool)

	// ConfigStrategyFromFile 从文件中获取策略配置
	ConfigStrategyFromFile(file ...string) config.StrategyConfig

	// Evaluate 模型评估函数
	Evaluate(values model.InputValues) model.OutputValues
}
