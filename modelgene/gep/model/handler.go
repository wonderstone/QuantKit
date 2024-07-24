package model

import (
	"os"

	functions2 "github.com/wonderstone/QuantKit/modelgene/gep/functions"
	"github.com/wonderstone/QuantKit/modelgene/gep/genome"
	"github.com/wonderstone/QuantKit/modelgene/gep/genomeset"
	"github.com/wonderstone/QuantKit/config"
	"gopkg.in/yaml.v3"
)

type InputValues []float64
type OutputValues []float64

type EvaluateFunc func(InputValues) OutputValues

// GetEvaluateFunc 用于单个基因组的评估
func GetEvaluateFunc(g *genome.Genome) EvaluateFunc {
	return func(input InputValues) OutputValues {
		return OutputValues{g.EvalMath(input)}
	}
}

// GetEvaluateFunc2 用于多个基因组的评估
func GetEvaluateFunc2(g *genomeset.GenomeSet) EvaluateFunc {
	return func(input InputValues) OutputValues {
		o := OutputValues{}
		for _, g := range g.Genomes {
			o = append(o, g.EvalMath(input))
		}

		return o
	}
}

type Handler interface {
	Init(option ...WithOption) error

	Evolve() *Record  // 用于多次评估
	RunOnce() *Record // 用于单次运行
}

type TrainInfo struct {
	Genome    *genome.Genome       // 当前基因组
	GenomeSet *genomeset.GenomeSet // 当前基因组集合
}

type GepRecord struct {
	Mode  string     `yaml:"mode"`
	Score float64    `yaml:"score"`
	KES   [][]string `yaml:"kes"`
}

type Record struct {
	TrainId   string `yaml:"train-id,omitempty"`
	TrainName string `yaml:"train-name,omitempty"`
	ModelId   string `yaml:"model-id,omitempty"`
	ModelName string `yaml:"model-name,omitempty"`

	Gep GepRecord `yaml:"gep"`
}

type Op struct {
	Conf                   *config.GepModel
	Perf                   PerformanceFunc
	Perf2                  PerformanceFunc2
	GenomeF                GenomeFunc
	GenomeSetF             GenomeSetFunc
	Record                 Record
	NumTerminal            int
	FuncType               functions2.FuncType
	Indicator2FormulaIndex map[string]int // 指标名称到公式索引的映射
}

type WithOption func(op *Op)

// PerformanceFunc 优化函数，返回最优基因组，以及是否达到预期
type PerformanceFunc func(int, []*genome.Genome) (g *genome.Genome, accomplished bool)
type PerformanceFunc2 func(int, []*genomeset.GenomeSet) (g *genomeset.GenomeSet, accomplished bool)

type GenomeFunc func(*genome.Genome)
type GenomeSetFunc func(*genomeset.GenomeSet)

func WithOp(opt *Op) WithOption {
	return func(op *Op) {
		*op = *opt
	}
}

func WithModelConfig(model config.GepModel) WithOption {
	return func(op *Op) {
		op.Conf = &model
	}
}

func WithGenome(f GenomeFunc) WithOption {
	return func(op *Op) {
		op.GenomeF = f
	}
}

func WithGenomeSet(f GenomeSetFunc) WithOption {
	return func(op *Op) {
		op.GenomeSetF = f
	}
}

func WithIndicator2FormulaIndex(m map[string]int) WithOption {
	return func(op *Op) {
		op.Indicator2FormulaIndex = m
	}
}

func WithKarvaExpressionFile(f string) WithOption {
	return func(op *Op) {
		rec := Record{
			Gep: GepRecord{
				KES: [][]string{},
			},
		}
		content, err := os.ReadFile(f)
		if err != nil {
			config.ErrorF("读取karva表达式文件失败: %v", err)
		}

		err = yaml.Unmarshal(content, &rec)
		if err != nil {
			config.ErrorF("解析karva表达式文件失败: %v", err)
		}

		op.Record = rec
	}
}

func WithPerformance(perf PerformanceFunc) WithOption {
	return func(op *Op) {
		op.Perf = perf
	}
}

func WithPerformanceSet(perf PerformanceFunc2) WithOption {
	return func(op *Op) {
		op.Perf2 = perf
	}
}

func WithNumTerminal(num int) WithOption {
	return func(op *Op) {
		op.NumTerminal = num
	}
}

func WithFuncType(funcType functions2.FuncType) WithOption {
	return func(op *Op) {
		op.FuncType = funcType
	}
}

func NewOp(option ...WithOption) *Op {
	op := &Op{
		FuncType: functions2.Float64,
	}

	for _, opt := range option {
		opt(op)
	}

	if op.Conf == nil {
		config.ErrorF("未指定模型配置")
	}

	if op.Perf == nil && op.Perf2 == nil && op.GenomeF == nil && op.GenomeSetF == nil {
		config.ErrorF("未指定筛选函数 或者 评估函数")
	}

	if op.NumTerminal == 0 {
		config.ErrorF("未指定模型参数数量，此参数由指标数量决定")
	}

	return op
}

func NewHandler(mode config.ModelType, option ...WithOption) Handler {
	switch mode {
	case config.ModelTypeGenome:
		handler := &GenomeHandler{}
		if err := handler.Init(option...); err != nil {
			config.ErrorF("初始化Genome模型失败: %v", err)
		}

		return handler
	case config.ModelTypeGenomeSet:
		handler := &GenomeSetHandler{}
		if err := handler.Init(option...); err != nil {
			config.ErrorF("初始化GenomeSet模型失败: %v", err)
		}
		return handler
	default:
		config.ErrorF("未知模型类型: %s", mode)
	}

	return nil
}
