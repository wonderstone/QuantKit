package handler

import (
	"github.com/wonderstone/QuantKit/framework/entity/contract"
)

type Contract interface {
	Resource

	// Init 初始化合约处理器
	Init(option ...contract.WithOption) (Contract, error)

	// GetContract 获取合约接口
	GetContract(instId string) contract.Contract
}
