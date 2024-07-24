package handler

import (
	"time"

	"github.com/wonderstone/QuantKit/modelgene/gep/model"
	"github.com/wonderstone/QuantKit/config"
)

type Global interface {
	// SetGEPInputParams 设置GEP输入参数(只在GEP模式有用)
	SetGEPInputParams(params []string)

	// Dir 获取目录
	Dir() *config.Path

	// Config 获取配置
	Config() *config.Runtime

	// Contract 获取合约处理器
	Contract() Contract

	// Quote 获取行情处理器
	Quote() Quote

	// ConfigStrategyFromFile 从文件中获取策略配置
	ConfigStrategyFromFile(file ...string) config.StrategyConfig

	// RunMode 获取运行类型
	RunMode() config.Mode

	// GetProgress 获取进度
	GetProgress() float64
}

type Setting interface {
	// SetLog 设置日志
	// level: 日志级别
	// path: 日志路径
	SetLog(
		level config.LogLevel,
		path ...string,
	)

	// SetStockParam 设置股票参数
	SetStockParam(
		indicate []string,
		instID []string,
	)

	// SetFutureParam 设置期货参数
	// indicates: 指标
	// stocks: 合约
	SetFutureParam(
		indicate []string,
		instID []string,
	)

	// SetMatchParam 设置撮合参数
	// matchFreq: 撮合频率
	// startTime: 撮合开始时间
	// endTime: 撮合结束时间, 如果不设置则一直撮合到无数据为止
	SetMatchParam(
		matchFreq config.Frequency,
		startTime time.Time,
		endTime ...time.Time,
	)

	// SetTrainParam 设置训练参数
	SetTrainParam(options ...model.WithOption)
}
