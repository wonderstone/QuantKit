package setting

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/base"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
)

var registryBase = make(map[config.HandlerType]reflect.Type)

// RegisterBase 注册新的公式
func RegisterBase(elem interface{}, name config.HandlerType) {
	t := reflect.TypeOf(elem).Elem()
	registryBase[name] = t
}

func NewBaseHandler(
	handlerName config.HandlerType, resource handler.Resource, options ...base.WithOption,
) (handler.Basic, error) {
	elem, ok := registryBase[handlerName]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("未知的基础数据处理器类型: %s", handlerName))
	}

	return reflect.New(elem).Interface().(handler.Basic).Init(resource, options...)
}

func MustBaseNewHandler(
	handlerName config.HandlerType, resource handler.Resource, options ...base.WithOption,
) handler.Basic {
	baseHandler, err := NewBaseHandler(handlerName, resource, options...)
	if err != nil {
		config.ErrorF(err.Error())
	}

	return baseHandler
}
