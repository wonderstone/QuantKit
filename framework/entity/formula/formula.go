package formula

import (
	"fmt"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"reflect"
	"time"
)

type Formula interface {
	DoInit(config config.Formula)
	DoCalculate(tm time.Time, row dataframe.RecordFunc) string
	DoReset()
}

var formulas = make(map[string]reflect.Type)

// RegisterNewFormula 注册新的公式
func RegisterNewFormula(elem interface{}, name ...string) {
	t := reflect.TypeOf(elem).Elem()
	if len(name) > 0 {
		formulas[name[0]] = t
		return
	}

	formulas[t.Name()] = t
}

// NewFormula 新建公式
func NewFormula(formulaName string) Formula {
	elem, ok := formulas[formulaName]
	if !ok {
		panic(fmt.Sprintf("未知的公式: %s\n", formulaName))
	}

	return reflect.New(elem).Interface().(Formula)
}
