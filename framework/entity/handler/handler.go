package handler

import (
	"github.com/wonderstone/QuantKit/config"
)


type Resource interface {
	// Dir 获取目录
	Dir() *config.Path
	// Config 获取配置
	Config() *config.Runtime
	// Base 获取基础处理器
	Base() Basic
	// Contract 获取合约处理器
	Contract() Contract
	// Indicator 获取指标处理器
	Quote() Quote
	// Account 获取账户处理器
	Account() Account
	// Creator 获取创建者
	Creator() StrategyCreator
	// Framework 获取框架
	Framework() Framework
}
