package setting

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
)

var matchCreator = make(map[config.HandlerType]reflect.Type)

func RegisterMatcher(elem interface{}, names ...config.HandlerType) {
	t := reflect.TypeOf(elem).Elem()
	for _, name := range names {
		matchCreator[name] = t
	}
}

func NewMatcher(matcherType config.HandlerType, opts ...handler.WithMatcherOption) (handler.Matcher, error) {
	elem, ok := matchCreator[matcherType]
	if !ok {
		return nil, fmt.Errorf("未知的撮合器类型: %s", matcherType)
	}

	return reflect.New(elem).Interface().(handler.Matcher).Init(opts...), nil
}

func MustNewMatcher(matcherType config.HandlerType, opts ...handler.WithMatcherOption) handler.Matcher {
	matcher, err := NewMatcher(matcherType, opts...)
	if err != nil {
		config.ErrorF("创建撮合器失败: %s", err)
	}

	return matcher
}

