package setting

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
)

type ReplayFramework interface {
	handler.Framework

	Init(resource Resource, option ...WithResource) error

	SetGenome(genome *genome.Genome)

	SetGenomeSet(genomeSet *genomeset.GenomeSet)

	SetStrategy(strategy handler.Strategy)

	// Matcher 获取匹配器
	Matcher() handler.Matcher

	// SubscribeData 订阅数据
	SubscribeData()

	// Run 运行
	Run()

	// IsFinished 是否完成
	IsFinished() bool

	// GetPerformance 获取性能指标结果
	GetPerformance() float64
}

var replayCreator = make(map[config.HandlerType]reflect.Type)
var rtExecCreator = make(map[config.HandlerType]reflect.Type)

func RegisterReplay(elem interface{}, name config.HandlerType) {
	replayCreator[name] = reflect.TypeOf(elem).Elem()
}

func NewReplay(name config.HandlerType) (ReplayFramework, error) {
	if t, ok := replayCreator[name]; ok {
		return reflect.New(t).Interface().(ReplayFramework), nil
	}

	support := make([]string, 0, len(replayCreator))
	for k := range replayCreator {
		support = append(support, string(k))
	}
	return nil, fmt.Errorf("没有找到对应的回放器，类型: %s, 目前支持: %v", name, support)
}

func MustNewReplay(name config.HandlerType) ReplayFramework {
	f, err := NewReplay(name)
	if err != nil {
		config.ErrorF(err.Error())
	}
	return f
}

type RTFramework interface {
	handler.Framework

	Init(resource Resource, option ...WithResource) error

	SetGenome(genome *genome.Genome)

	SetGenomeSet(genomeSet *genomeset.GenomeSet)

	SetStrategy(strategy handler.Strategy)

	// Matcher 获取匹配器
	Matcher() handler.Matcher

	// SubscribeData 订阅数据
	SubscribeData()

	// Run 运行
	Run()

	// IsFinished 是否完成
	IsFinished() bool

	// GetPerformance 获取性能指标结果
	GetPerformance() float64
}

func RegisterRTExecutor(elem interface{}, name config.HandlerType) {
	rtExecCreator[name] = reflect.TypeOf(elem).Elem()
}

func NewRTExecutor(name config.HandlerType) (RTFramework, error) {
	if t, ok := rtExecCreator[name]; ok {
		return reflect.New(t).Interface().(RTFramework), nil
	}

	support := make([]string, 0, len(rtExecCreator))
	for k := range rtExecCreator {
		support = append(support, string(k))
	}
	return nil, fmt.Errorf("没有找到对应的实盘执行器，类型: %s, 目前支持: %v", name, support)
}

func MustNewRTExecutor(name config.HandlerType) RTFramework {
	f, err := NewRTExecutor(name)
	if err != nil {
		config.ErrorF(err.Error())
	}
	return f
}
