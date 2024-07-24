package indicator

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	_ "github.com/wonderstone/QuantKit/framework/logic/formula/inner"
	"github.com/wonderstone/QuantKit/tools/container/btree"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dag"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Cell struct {
	NodeID  int64 // node id
	Config  *config.Formula
	Formula formula.Formula
}

func (c Cell) ID() int64 {
	return c.NodeID
}

type StreamCalcGraph struct {
	calculator *StreamLoadCalculator
	instID     string // 合约ID
	calcCell   []Cell

	quotes []dataframe.StreamingRecord
	record dataframe.StreamingRecord
}

// StreamLoadCalculator 流式读取指标计算器
type StreamLoadCalculator struct {
	*simple.DirectedGraph
	indicatorDataPath string
	config            config.Runtime
	// 拓扑排序后的节点
	sortedNodes    []graph.Node
	indicators     map[string]config.Formula
	nodeNumber     int64
	node2Indicator map[int64]*Cell
	indicator2Node map[string]*Cell
	indicator      []string // 指标名称

	inst2Graph map[string]*StreamCalcGraph

	loadOnce sync.Once

	xrxd map[string]*btree.MapIterG[time.Time, *config.Xrxd]

	settleTimeQueue map[time.Time][]string
}

func (f *StreamLoadCalculator) NewCell(conf config.Formula) (cell *Cell) {
	if _, ok := f.indicator2Node[conf.Name]; ok {
		config.ErrorF("指标名称重复: %s", conf.Name)
	}

	// id自增
	f.nodeNumber += 1
	cell = &Cell{
		NodeID:  f.nodeNumber,
		Config:  &conf,
		Formula: nil,
	}

	f.AddNode(cell)
	return cell
}

func (f *StreamLoadCalculator) Find(name string) (cell *Cell) {
	if cell, ok := f.indicator2Node[name]; ok {
		return cell
	} else {
		config.ErrorF("指标不存在: %s", name)
	}

	return
}

func (f *StreamLoadCalculator) buildDAG(property config.IndicatorProperty) {
	for _, p := range property.Indicator {
		cell := f.NewCell(p)

		f.indicator = append(f.indicator, p.Name)
		f.node2Indicator[cell.ID()] = cell
		f.indicator2Node[p.Name] = cell
	}

	for _, p := range property.Indicator {
		cell := f.Find(p.Name)
		for dep, _ := range p.Input {
			depCell := f.Find(dep)
			f.SetEdge(f.NewEdge(depCell, cell))
		}

		for _, dep := range p.Depend {
			depCell := f.Find(dep)
			f.SetEdge(f.NewEdge(depCell, cell))
		}
	}

	// 检查是否回环
	if !dag.IsDAG(f.DirectedGraph) {
		config.ErrorF("指标计算器存在回环")
	}

	sorted, err := topo.Sort(f)
	if err != nil {
		panic(err)
	}

	f.sortedNodes = sorted
}

func (f *StreamLoadCalculator) Init(option ...formula.WithOption) error {
	f.inst2Graph = make(map[string]*StreamCalcGraph)
	f.settleTimeQueue = make(map[time.Time][]string)

	f.DirectedGraph = simple.NewDirectedGraph()
	f.node2Indicator = make(map[int64]*Cell)
	f.indicator2Node = make(map[string]*Cell)

	op := formula.NewOp(option...)

	f.config = op.Config

	f.SetSourceDataPath(filepath.Join(op.Config.Path.Download, string(op.Config.Framework.Frequency)))

	f.buildDAG(*op.Config.Indicator)

	for _, inst := range op.Config.Framework.Instrument {
		// 检查文件是否存在
		p := filepath.Join(f.indicatorDataPath, inst+".csv")
		if _, err := os.Stat(p); err != nil {
			config.ErrorF("文件不存在: %s，请下载 %s", p, inst)
		}

		// 构建拓扑排序
		g := &StreamCalcGraph{
			calculator: f,
			instID:     inst,
			record: dataframe.StreamingRecord{
				Data:    make([]string, len(f.sortedNodes)),
				Headers: make(map[string]int),
			},
		}

		for i, node := range f.sortedNodes {
			g.record.Headers[node.(*Cell).Config.Name] = i
			cell := *node.(*Cell)
			if cell.Config.Func != "" {
				cell.Formula = formula.NewFormula(cell.Config.Func)
				cell.Config.InstID = g.instID
				cell.Formula.DoInit(*cell.Config)
			}
			g.calcCell = append(g.calcCell, cell)
		}

		f.inst2Graph[inst] = g
	}

	return nil
}

func (f *StreamCalcGraph) GetColumns() map[string]int {
	return f.record.Headers
}

func (f *StreamCalcGraph) GetIndicator(name string, row int) float64 {
	return f.record.ConvertToFloat(name)
}

func (f *StreamLoadCalculator) SetSourceDataPath(dir string) {
	f.indicatorDataPath = dir
}

func (f *StreamLoadCalculator) StartCalc() {
}

func (g *StreamCalcGraph) calcInstIDOneLine(
	tm time.Time,
	record dataframe.StreamingRecord,
) {
	for _, cell := range g.calcCell {
		if cell.Config.Func == "" {
			g.record.Update(cell.Config.Name, record.Val(cell.Config.Name))
		} else {
			g.record.Update(cell.Config.Name, cell.Formula.DoCalculate(tm, g.record))
		}
	}
}

func (f *StreamLoadCalculator) Calculate(
	tm time.Time,
	indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) orderedmap.OrderedMap[string, dataframe.StreamingRecord] {

	records := orderedmap.New[string, dataframe.StreamingRecord]()
	for curr := indicators.Oldest(); curr != nil; curr = curr.Next() {
		instID := curr.Key
		var quoteRecord dataframe.StreamingRecord
		quoteRecord.Headers = curr.Value.Headers
		quoteRecord.Data = make([]string, len(curr.Value.Data))
		copy(quoteRecord.Data, curr.Value.Data)
		g := f.inst2Graph[instID]
		// 记录一下目前的合约行情，用于计算除权除息
		g.quotes = append(g.quotes, quoteRecord)

		g.calcInstIDOneLine(tm, quoteRecord)

		records.Set(instID, g.record)
	}

	return *records
}

func (f *StreamLoadCalculator) getXrxds(base handler.Basic, tm time.Time) {
	f.xrxd = make(map[string]*btree.MapIterG[time.Time, *config.Xrxd])
	for _, inst := range f.config.Framework.Instrument {
		if xrxd, need := base.GetXrxd(inst, tm); need {
			if xrxd.Key().Before(tm) {
				if !xrxd.Next() {
					continue
				}
			}

			f.xrxd[inst] = &xrxd

			if _, ok := f.settleTimeQueue[xrxd.Key()]; !ok {
				f.settleTimeQueue[xrxd.Key()] = make([]string, 0)
			}

			f.settleTimeQueue[xrxd.Key()] = append(f.settleTimeQueue[xrxd.Key()], inst)
		}
	}
}

func (v *StreamCalcGraph) doSettle(tm time.Time, exFactor float64) {
	for _, cell := range v.calcCell {
		if cell.Formula != nil {
			cell.Formula.DoReset()
		}
	}

	for i, quote := range v.quotes {
		v.quotes[i].Update(
			"Close", fmt.Sprintf("%.2f", quote.ConvertToFloat("Close")*exFactor),
		)
		v.quotes[i].Update(
			"Open", fmt.Sprintf("%.2f", quote.ConvertToFloat("Open")*exFactor),
		)
		v.quotes[i].Update(
			"High", fmt.Sprintf("%.2f", quote.ConvertToFloat("High")*exFactor),
		)
		v.quotes[i].Update(
			"Low", fmt.Sprintf("%.2f", quote.ConvertToFloat("Low")*exFactor),
		)
		v.quotes[i].Update(
			"Volume", fmt.Sprintf("%.2f", quote.ConvertToFloat("Volume")/exFactor),
		)
		v.quotes[i].Update(
			"Amount", fmt.Sprintf("%.2f", quote.ConvertToFloat("Amount")/exFactor),
		)
	}

	for _, quoteRecord := range v.quotes {
		v.calcInstIDOneLine(tm, quoteRecord)
	}
}

func (f *StreamLoadCalculator) DoSettle(
	tm time.Time,
	base handler.Basic,
) {
	// 获取除权除息数据
	if base.NeedReload() {
		f.getXrxds(base, tm)
	} else {
		f.loadOnce.Do(
			func() {
				// fmt.Printf("加载除权除息数据\n")
				f.getXrxds(base, tm)
			},
		)
	}

	if settles, ok := f.settleTimeQueue[tm]; !ok {
		// config.DebugF("没有找到对应的除权除息数据: %s", tm)

		for _, xrxd := range f.xrxd {
			for xrxd.Next() {
				if xrxd.Key().After(tm) {
					break
				}
			}
		}
		return
	} else {
		// config.DebugF("找到对应的除权除息数据: %s", tm)
		for _, inst := range settles {
			if xrxd, ok := f.xrxd[inst]; ok {
				f.inst2Graph[inst].doSettle(tm, xrxd.Value().ExFactor)
				if !xrxd.Next() {
					delete(f.xrxd, inst)
				}
			}
		}
	}
}

func init() {
	formula.RegisterNewCalculator(
		&StreamLoadCalculator{},
		config.HandlerTypeStream,
		config.HandlerTypeDefault,
	)
}
