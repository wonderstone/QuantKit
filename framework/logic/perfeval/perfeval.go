package perfeval

import (
	"math"
	"sort"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/perf"
	"github.com/wonderstone/QuantKit/tools/recorder"
)

type PerfEval struct {
	Records []recorder.AssetRecord
	sorted  bool
}

type Op struct {
	Tag          perf.IndicateType
	RiskFreeRate float64
}

type WithOption func(*Op)

func WithPerformanceIndicateType(tag perf.IndicateType) WithOption {
	return func(op *Op) {
		op.Tag = tag
	}
}

func WithRiskFreeRate(rf float64) WithOption {
	return func(op *Op) {
		op.RiskFreeRate = rf
	}
}

func NewOp(options ...WithOption) *Op {
	op := &Op{}
	for _, opt := range options {
		opt(op)
	}
	return op
}

func (p *PerfEval) Less(i, j int) bool {
	return p.Records[i].Date < p.Records[j].Date
}

// NewPerfEval 新建一个PerfEval， 如果sorted为true，则表示已经排序好了，默认为false
func NewPerfEval(Accounts []recorder.AssetRecord, sorted ...bool) *PerfEval {
	if len(sorted) > 0 {
		return &PerfEval{
			Records: Accounts,
			sorted:  sorted[0],
		}
	}

	return &PerfEval{
		Records: Accounts,
		sorted:  false,
	}
}

func (p *PerfEval) CalcPerfEvalResult(options ...WithOption) float64 {
	op := NewOp(options...)

	switch op.Tag {
	case perf.TotalReturn:
		return p.TotalReturn()
	case perf.AnnualizedReturn:
		return p.AnnualizedReturn()
	case perf.MaxDrawdown:
		return p.MaxDrawDown()
	// case perf.InformationRatio:
	// 	return p.AnnualizedReturn() / p.MaxDrawDown()
	case perf.SharpeRatio:
		return p.SharpeRatio(op.RiskFreeRate)
	default:
		config.ErrorF("未知的性能指标: %s", op.Tag)
	}

	return 0
}

func (p *PerfEval) RateOfReturns() (RoRs []float64) {
	RoRs = make([]float64, p.Len()-1)
	for i := 1; i < p.Len(); i++ {
		RoRs[i-1] = (p.Records[i].TotalAsset / p.Records[i-1].TotalAsset) - 1
	}
	return
}

func (p *PerfEval) TotalReturn() (TR float64) {
	if p.Len() == 0 {
		config.ErrorF("没有交易记录")
	}

	if !p.sorted {
		p.Sort()
	}

	return p.Records[p.Len()-1].TotalAsset / p.Records[0].TotalAsset
}

func (p *PerfEval) AnnualizedReturn() (AR float64) {
	// 默认了日线级别 偷懒做法  后期有空精细化吧
	return math.Pow(p.TotalReturn(), float64(252/p.Len()))
}

func (p *PerfEval) MaxDrawDown() (maxDrawDown float64) {
	if !p.sorted {
		p.Sort()
	}
	maxVal := 0.0
	for i := 0; i < p.Len(); i++ {
		if p.Records[i].TotalAsset > maxVal {
			maxVal = p.Records[i].TotalAsset
		}
		drawDown := 1.0 - (p.Records[i].TotalAsset / maxVal)
		if drawDown > 0 && drawDown > maxDrawDown {
			maxDrawDown = drawDown
		}

	}
	return
}

func (p *PerfEval) SharpeRatio(Rf float64) (SR float64) {
	// 默认了日线级别 偷懒做法  后期有空精细化吧
	std := Std(p.RateOfReturns(), 1)
	if std == 0 {
		return 0
	}
	return (p.AnnualizedReturn() - Rf) / (math.Sqrt(252) * std)
}

func (p *PerfEval) Len() int {
	return len(p.Records)
}

func (p *PerfEval) Swap(i, j int) {
	p.Records[i], p.Records[j] = p.Records[j], p.Records[i]
}

func (p *PerfEval) Add(time string, acc recorder.AssetRecord) {
	p.Records = append(p.Records, acc)
	p.sorted = false
}

func (p *PerfEval) Sort() {
	if !p.sorted {
		sort.Sort(p)
		p.sorted = true
	}
}
