package setting

import (
	"reflect"

	"github.com/wonderstone/QuantKit/framework/entity/handler"
)

var typeRegistry = make(map[string]reflect.Type)

func RegisterAccount(elem interface{}, name ...string) {
	if len(name) > 0 {
		typeRegistry[name[0]] = reflect.TypeOf(elem).Elem()
		return
	}

	t := reflect.TypeOf(elem).Elem()
	typeRegistry[t.Name()] = t
}

func NewAccount(accountType string) handler.Account2 {
	elem, ok := typeRegistry[accountType]
	if !ok {
		panic("未知的账户类型: " + accountType)
	}
	return reflect.New(elem).Interface().(handler.Account2)
}

