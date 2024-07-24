package formula

import (
	"os"
	"path/filepath"
	"time"

	"github.com/wonderstone/QuantKit/framework/entity/formula"
	_ "github.com/wonderstone/QuantKit/framework/logic/formula/inner"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/dag"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/times"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Cell struct {
	NodeID  int64 // node id
	Name    string
	InstID  string
	Config  config.Formula
	Formula formula.Formula
}

func (c Cell) ID() int64 {
	return c.NodeID
}

func (f *FullLoadCalculator) NewCell(name string) (cell *Cell, new bool) {
	if cell, ok := f.indicator2Node[name]; ok {
		return cell, false
	}

	// id自增
	f.nodeNumber += 1
	cell = &Cell{
		NodeID:  f.nodeNumber,
		Name:    name,
		Formula: nil,
	}

	f.AddNode(cell)
	return cell, true
}

// FullLoadCalculator 全量读取指标计算器
type FullLoadCalculator struct {
	quoteDataPath     string
	indicatorDataPath string
	financialDataPath string
	outputPath        string
	config            config.Runtime

	instID2Path map[string]string // 合约ID->路径
	indicator   []string          // 指标名称

	*simple.DirectedGraph
	node2Indicator map[int64]*Cell
	indicator2Node map[string]*Cell
	nodeNumber     int64

	df dataframe.DataFrame
}

func (f *FullLoadCalculator) Init(option ...formula.WithOption) error {
	f.DirectedGraph = simple.NewDirectedGraph()
	f.node2Indicator = make(map[int64]*Cell)
	f.indicator2Node = make(map[string]*Cell)
	f.instID2Path = make(map[string]string)

	op := formula.NewOp(option...)

	f.config = op.Config
	f.quoteDataPath = filepath.Join(op.Config.Path.Download, string(op.Config.Framework.Frequency))
	f.outputPath = op.Config.Path.Indicator

	for _, inst := range op.Config.Framework.Instrument {
		// 检查文件是否存在
		p := filepath.Join(f.quoteDataPath, inst+".csv")
		if _, err := os.Stat(p); err == nil {
			f.instID2Path[inst] = p
		}
	}

	indicatorConfig, err := config.NewIndicatorConfig(op.Config.Path.IndicatorFile)
	if err != nil {
		return err
	}

	f.buildDAG(*indicatorConfig)

	return nil
}

func (f *FullLoadCalculator) GetColumns() map[string]int {
	return f.df.HeaderToIndex
}

func (f *FullLoadCalculator) GetIndicator(name string, row int) float64 {
	return f.df.FrameRecords[row].ConvertToFloat(name, f.GetColumns())
}

func (f *FullLoadCalculator) buildDAG(property config.IndicatorProperty) {
	for _, p := range property.Indicator {
		// 检查是否有公式
		if p.Func == "" {
			continue
		}

		cell, _ := f.NewCell(p.Name)
		cell.Config = p

		f.indicator = append(f.indicator, p.Name)
		f.node2Indicator[cell.ID()] = cell
		f.indicator2Node[p.Name] = cell

		for dep, _ := range p.Input {
			depCell, _ := f.NewCell(dep)
			f.SetEdge(f.NewEdge(depCell, cell))
		}
	}

	// 检查是否回环
	if !dag.IsDAG(f.DirectedGraph) {
		panic("指标计算器存在回环")
	}

}

func (f *FullLoadCalculator) StartCalc() {
	// 进行拓扑排序
	sorted, err := topo.Sort(f)
	if err != nil {
		panic(err)
	}

	// 计算指标
	for instID, file := range f.instID2Path {
		// 读取已有数据，包括行情的高开低收，成交量，成交额
		f.df = dataframe.CreateDataFrame(filepath.Dir(file), filepath.Base(file))
		for _, indicator := range f.indicator {
			f.df.NewField(indicator)
		}

		for _, indicator := range sorted {
			cell := indicator.(*Cell)
			cell.Formula = formula.NewFormula(cell.Config.Func)
			cell.Config.InstID = instID
			cell.Formula.DoInit(cell.Config)
		}

		// 计算指标
		for _, row := range f.df.FrameRecords {
			// 2019.01.03T14:50:00.000
			// 将时间分割出来
			result := times.MustDuration(config.TimeFormatTime, row.Val("Time", f.df.HeaderToIndex)[11:])
			tm, err := time.Parse(config.TimeFormatDefault, row.Val("Time", f.df.HeaderToIndex))
			if err != nil {
				config.ErrorF("时间格式错误: %s", err)
			}
			if f.config.Framework.Frequency == config.Frequency1Day && result.Minutes() != f.config.Framework.DailyTriggerTime.Minutes() {
				continue
			}

			for _, indicator := range sorted {
				cell := indicator.(*Cell)
				if cell.Formula != nil {
					row.Update(cell.Name, cell.Formula.DoCalculate(tm, row))
				}
			}
		}

		// 保存数据
		err := os.MkdirAll(f.outputPath, os.ModePerm)
		if err != nil {
			config.ErrorF("创建计算结果目录失败: %s", err)
		}

		f.df.SaveDataFrame(f.outputPath, instID)
	}

}

func init() {
	formula.RegisterNewCalculator(
		&FullLoadCalculator{},
		config.HandlerTypeFullLoad,
		config.HandlerTypeDefault,
	)
}
