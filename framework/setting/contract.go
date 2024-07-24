package setting

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
)

var contractCreator = make(map[config.HandlerType]reflect.Type)

func RegisterContractHandler(elem interface{}, name ...config.HandlerType) {
	for _, n := range name {
		contractCreator[n] = reflect.TypeOf(elem).Elem()
	}
}

func NewContractHandler(name config.HandlerType, options ...contract.WithOption) (handler.Contract, error) {
	if t, ok := contractCreator[name]; ok {
		contractHandler := reflect.New(t).Interface().(handler.Contract)
		return contractHandler.Init(options...)
	}

	support := make([]string, 0, len(contractCreator))
	for k := range contractCreator {
		support = append(support, string(k))
	}
	return nil, fmt.Errorf("没有找到对应的合约处理器，类型: %s, 目前支持: %v", name, support)
}

func MustNewContractHandler(name config.HandlerType, options ...contract.WithOption) handler.Contract {
	f, err := NewContractHandler(name, options...)
	if err != nil {
		config.ErrorF(err.Error())
	}
	return f
}




