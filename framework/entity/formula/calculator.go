package formula

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
)

// Calculator 指标计算接口
type Calculator interface {
	// GetColumns 获取指标列名
	GetColumns() map[string]int

	// GetIndicator 获取指标
	GetIndicator(name string, row int) float64
}

// CalculatorModule 指标计算接口
type CalculatorModule interface {
	// Init 初始化
	Init(option ...WithOption) error

	StartCalc()
}

var calculators = make(map[config.HandlerType]reflect.Type)

// RegisterNewCalculator 注册新的计算器
func RegisterNewCalculator(elem interface{}, name ...config.HandlerType) {
	for _, n := range name {
		calculators[n] = reflect.TypeOf(elem).Elem()
	}
}

// NewCalculator 新建计算器
func NewCalculator(calculatorName config.HandlerType) (CalculatorModule, error) {
	elem, ok := calculators[calculatorName]
	if !ok {
		return nil, fmt.Errorf("未知的计算器类型: %s", calculatorName)
	}

	return reflect.New(elem).Interface().(CalculatorModule), nil
}

func MustNewCalculator(calculatorName config.HandlerType) CalculatorModule {
	calculator, err := NewCalculator(calculatorName)
	if err != nil {
		config.ErrorF(err.Error())
	}

	return calculator
}
