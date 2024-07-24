package setting

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/entity/quote"
)

var quoteCreator = make(map[config.HandlerType]reflect.Type)

// RegisterNewQuote 注册新的公式
func RegisterNewQuote(elem interface{}, names ...config.HandlerType) {
	t := reflect.TypeOf(elem).Elem()
	for _, name := range names {
		quoteCreator[name] = t
	}
}

func NewQuote(handlerName config.HandlerType, option ...quote.WithOption) (handler.Quote, error) {

	elem, ok := quoteCreator[handlerName]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("未知的指标处理器类型: %s", handlerName))
	}

	quoteHandler := reflect.New(elem).Interface().(handler.Quote)
	err := quoteHandler.Init(option...)
	if err != nil {
		return nil, err
	}

	return quoteHandler, nil
}

func MustNewQuote(handlerName config.HandlerType, option ...quote.WithOption) handler.Quote {
	quoteHandler, err := NewQuote(handlerName, option...)
	if err != nil {
		config.ErrorF(err.Error())
	}

	return quoteHandler
}
