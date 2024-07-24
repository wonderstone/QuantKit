package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type StrategyParamConfig struct {
	Name              string            `yaml:"name"`
	Type              StrategyParamType `yaml:"type"`
	Def               any               `yaml:"default"`
	DefaultSliceInt   []int64
	DefaultSliceFloat []float64
	DefaultSliceStr   []string
	DefaultInt        int64
	DefaultFloat      float64
	DefaultStr        string
	Comment           string `yaml:"comment"`
}

type StrategyDefConfig struct {
	Name string   `yaml:"name"`
	Tag  []string `yaml:"tag"`
	Desc string   `yaml:"desc"`
}

type StrategyConfig struct {
	Define StrategyDefConfig     `yaml:"strategy"`
	Param  []StrategyParamConfig `yaml:"strategy-param"`
}

func anyCastInt(t any) int64 {
	switch t.(type) {
	case int8, int16, int32, int, int64:
		return int64(t.(int))
	case float32, float64:
		return int64(t.(float64))
	case string:
		val, err := strconv.ParseInt(t.(string), 10, 64)
		if err != nil {
			ErrorF("参数 %s 的返回值类型不是int", t)
		}

		return val
	default:
		ErrorF("参数 %s 的返回值类型不是int", t)
	}
	return 0
}

func anyCastFloat(t any) float64 {
	switch tVal := t.(type) {
	case int8, int16, int32, int, int64:
		return float64(t.(int))
	case float32, float64:
		return tVal.(float64)
	case string:
		val, err := strconv.ParseFloat(t.(string), 64)
		if err != nil {
			ErrorF("参数 %s 的返回值类型不是float", t)
		}

		return val
	default:
		ErrorF("参数 %s 的返回值类型不是float", t)
	}
	return 0
}

func NewStrategyConfig(configFile string) (StrategyConfig, error) {
	f, err := os.ReadFile(configFile)
	if err != nil {
		return StrategyConfig{}, err
	}

	c := yaml.Node{}
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return StrategyConfig{}, err
	}

	if len(c.Content) != 1 {
		return StrategyConfig{}, fmt.Errorf("策略配置文件格式错误")
	}

	var node *yaml.Node = nil
	for i, v := range c.Content[0].Content {
		if v.Value == "strategy-param" {
			if len(c.Content[0].Content) < i+2 {
				return StrategyConfig{}, fmt.Errorf("策略配置文件格式错误")
			}

			node = c.Content[0].Content[i+1]
			break
		}
	}

	if node == nil {
		return StrategyConfig{}, fmt.Errorf("策略配置文件格式错误")
	}

	conf := StrategyConfig{}
	for i := 0; i < len(node.Content); i += 1 {
		paramNodes := node.Content[i].Content

		param := StrategyParamConfig{}
		val := ""
		vals := []string{}
		for i := 0; i < len(paramNodes); i += 2 {
			switch paramNodes[i].Value {
			case "name":
				param.Name = paramNodes[i+1].Value
			case "type":
				param.Type = StrategyParamType(paramNodes[i+1].Value)
			case "default":
				switch paramNodes[i+1].Kind {
				case yaml.ScalarNode:
					val = paramNodes[i+1].Value
				case yaml.SequenceNode:
					for _, v := range paramNodes[i+1].Content {
						vals = append(vals, v.Value)
					}
				}
			}
		}

		switch param.Type {
		case StrategyParamTypeInt:
			param.DefaultInt = anyCastInt(val)
		case StrategyParamTypeFloat:
			param.DefaultFloat = anyCastFloat(val)
		case StrategyParamTypeString:
			param.DefaultStr = val
		case StrategyParamTypeArrayInt:
			for _, v := range vals {
				param.DefaultSliceInt = append(param.DefaultSliceInt, anyCastInt(v))
			}
		case StrategyParamTypeArrayFloat:
			for _, v := range vals {
				param.DefaultSliceFloat = append(param.DefaultSliceFloat, anyCastFloat(v))
			}
		case StrategyParamTypeArrayString:
			param.DefaultSliceStr = vals
		default:
			ErrorF("参数 %s 的类型 %s 不支持", param.Name, param.Type)
		}

		conf.Param = append(conf.Param, param)
	}

	return conf, nil
}

func GetParam[T any](c StrategyConfig, name string, defaultVal ...T) (t T) {
	for _, p := range c.Param {
		if p.Name == name {
			switch p.Type {
			case StrategyParamTypeArrayString:
				tType, ok := any(t).([]string)
				if !ok {
					ErrorF("参数 %s 的返回值类型不是数组", name)
					return
				}

				for _, v := range p.DefaultSliceStr {
					tType = append(tType, v)
				}

				t = any(tType).(T)
				return
			case StrategyParamTypeArrayInt:
				switch any(t).(type) {
				case []int8:
					var vals []int8
					for _, val := range p.DefaultSliceInt {
						vals = append(vals, int8(val))
					}

					t = any(vals).(T)
				case []int16:
					var vals []int16
					for _, val := range p.DefaultSliceInt {
						vals = append(vals, int16(val))
					}

					t = any(vals).(T)
				case []int32:
					var vals []int32
					for _, val := range p.DefaultSliceInt {
						vals = append(vals, int32(val))
					}

					t = any(vals).(T)
				case []int:
					var vals []int
					for _, val := range p.DefaultSliceInt {
						vals = append(vals, int(val))
					}

					t = any(vals).(T)

				case []int64:
					var vals []int64

					for _, val := range p.DefaultSliceInt {
						vals = append(vals, val)
					}

					t = any(vals).(T)
				default:
					ErrorF("参数 %s 的返回值类型不是数组", name)
				}

				return
			case StrategyParamTypeArrayFloat:
				switch any(t).(type) {
				case []float32:
					var vals []float32
					for _, val := range p.DefaultSliceFloat {
						vals = append(vals, float32(val))
					}

					t = any(vals).(T)
				case []float64:
					var vals []float64

					for _, val := range p.DefaultSliceFloat {
						vals = append(vals, val)
					}

					t = any(vals).(T)
				default:
					ErrorF("参数 %s 的返回值类型不是数组", name)
				}

			case StrategyParamTypeInt, StrategyParamTypeFloat, StrategyParamTypeString:
				switch any(t).(type) {
				case int8:
					i := int8(p.DefaultInt)
					t = any(i).(T)
				case int16:
					i := int16(p.DefaultInt)
					t = any(i).(T)
				case int32:
					i := int32(p.DefaultInt)
					t = any(i).(T)
				case int:
					i := int(p.DefaultInt)
					t = any(i).(T)
				case int64:
					i := p.DefaultInt
					t = any(i).(T)
				case float32:
					i := float32(p.DefaultFloat)
					t = any(i).(T)
				case float64:
					i := p.DefaultFloat
					t = any(i).(T)
				case string:
					t = any(p.DefaultStr).(T)
				default:
					ErrorF("参数 %s 的返回值类型不支持，目前支持[int, float, string]", name)
				}

				return
			default:
				ErrorF("参数 %s 的类型 %s 不支持", name, p.Type)
			}
		}
	}

	if len(defaultVal) > 0 {
		return defaultVal[0]
	}

	ErrorF("没有找到该策略参数: %s", name)
	return
}
