package model

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
)

type Handler interface {
	// Run 初始化
	Run(option ...WithOption) error
}

var registry = make(map[config.HandlerType]reflect.Type)

// RegisterNewHandler 注册新的公式
func RegisterNewHandler(elem interface{}, names ...config.HandlerType) {
	t := reflect.TypeOf(elem).Elem()
	for _, name := range names {
		registry[name] = t
	}
}

func Run(handlerName config.HandlerType, option ...WithOption) error {
	elem, ok := registry[handlerName]
	if !ok {
		return fmt.Errorf(fmt.Sprintf("未知的模型处理器类型: %s", handlerName))
	}

	handler := reflect.New(elem).Interface().(Handler)
	err := handler.Run(option...)
	if err != nil {
		return err
	}

	return nil
}
