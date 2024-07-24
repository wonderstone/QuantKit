package setting

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
)

type Runner interface {
	handler.Global
	// Init 初始化
	Init(creator ...WithResource) error

	// SetStrategyCreator(creator runner.StrategyCreator)

	Start() error
}

var runnerCreator = make(map[config.Mode]reflect.Type)

func RegisterRunner(elem interface{}, name config.Mode) {
	runnerCreator[name] = reflect.TypeOf(elem).Elem()
}

func NewRunner(name config.Mode) (Runner, error) {
	if t, ok := runnerCreator[name]; ok {
		return reflect.New(t).Interface().(Runner), nil
	}

	return nil, fmt.Errorf("没有找到对应的Runner，类型: %s", name)
}

func MustNewRunner(name config.Mode) Runner {
	if t, ok := runnerCreator[name]; ok {
		return reflect.New(t).Interface().(Runner)
	}

	names := make([]string, 0, len(runnerCreator))
	for n := range runnerCreator {
		names = append(names, string(n))
	}

	config.ErrorF("没有找到对应的Runner，类型: %s, 支持的运行模式: %v", name, names)
	return nil
}
