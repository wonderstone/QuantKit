package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wonderstone/QuantKit/tools/container/slice"
	"github.com/wonderstone/QuantKit/tools/times"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

// - Formula is actually the basic component of indicator yaml file
// ? InstID is
type Formula struct {
	Name   string            `yaml:"name"` // 指标名称
	Func   string            `yaml:"func"` // 指标函数
	InstID string            // 合约代码
	Param  map[string]string `yaml:"param,omitempty"`  // 指标参数
	Input  map[string]string `yaml:"input,omitempty"`  // 输入指标信息, key为指标名称，value为指标数量
	Depend []string          `yaml:"depend,omitempty"` // 输入指标信息, key为指标名称，value为指标数量
}

// - MustGetParamType funcs are used in formulas
// - kind of sth u can do it urself but provided by us
func MustGetParamFloat64(param map[string]string, key string) float64 {
	if v, ok := param[key]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			ErrorF("参数[%s]值[%s]不是数字", key, v)
		}

		return f
	}

	ErrorF("参数[%s]不存在", key)

	return 0
}

func MustGetParamInt(param map[string]string, key string) int64 {
	if v, ok := param[key]; ok {
		f, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			ErrorF("参数[%s]值[%s]不是数字", key, v)
		}

		return f
	}

	ErrorF("参数[%s]不存在", key)

	return 0
}

func MustGetParamString(param map[string]string, key string) string {
	if v, ok := param[key]; ok {
		return v
	}

	ErrorF("参数[%s]不存在", key)

	return ""
}

// - Indicator is based on Formula slice
type IndicatorProperty struct {
	Indicator []Formula `yaml:"indicator"`
}

func NewIndicatorConfig(configFile string) (*IndicatorProperty, error) {
	f, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// 默认指标
	c := IndicatorProperty{
		Indicator: []Formula{
			{
				Name: "Open",
			},
			{
				Name: "Close",
			},
			{
				Name: "High",
			},
			{
				Name: "Low",
			},
			{
				Name: "Volume",
			},
			{
				Name: "Amount",
			},
		},
	}
	// - if unmarshal has no error, c will be filled by yaml
	// - and the default []formula will be removed
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return nil, err
	}
	// - therefore, need to add the default again!
	defaultFormula := []Formula{
		{
			Name: "Open",
		},
		{
			Name: "Close",
		},
		{
			Name: "High",
		},
		{
			Name: "Low",
		},
		{
			Name: "Volume",
		},
		{
			Name: "Amount",
		},
	}

	c.Indicator = append(defaultFormula, c.Indicator...)

	return &c, nil
}

type FutureContract struct {
	Name string `yaml:"name"` // 合约名称

	Basic ContractBasic `yaml:",inline"` // 合约基本信息

	MarginRate MarginRate `yaml:",inline"` // 保证金比例

	FutureFee FutureFee `yaml:",inline"` // 期货费用
}

func (s *FutureContract) InitializeObject() error {
	s.Basic = ContractBasic{
		ContractSize: 100,
		MinOrderVol:  100,
		TickSize:     0.01,
		TPlus:        1,
	}

	s.MarginRate = MarginRate{
		Long:   0.08,
		Short:  0.08,
		Broker: 0.02,
	}

	s.FutureFee = FutureFee{
		Broker:        0,
		Open:          0,
		CloseToday:    0,
		ClosePrevious: 0,
	}

	return nil
}

type MarginRate struct {
	Long   float64 `yaml:"margin-long,omitempty"`   // 多头保证金比例
	Short  float64 `yaml:"margin-short,omitempty"`  // 空头保证金比例
	Broker float64 `yaml:"margin-broker,omitempty"` // 佣金保证金比例
}

type FutureFee struct {
	Broker            float64 `yaml:"comm-broker,omitempty"`              // 佣金
	Open              float64 `yaml:"comm-open,omitempty"`                // 开仓佣金
	CloseToday        float64 `yaml:"comm-close-today,omitempty"`         // 平今佣金
	ClosePrevious     float64 `yaml:"comm-close-previous,omitempty"`      // 平昨佣金
	BrokerRate        float64 `yaml:"comm-broker-rate,omitempty"`         // 佣金率
	OpenRate          float64 `yaml:"comm-open-rate,omitempty"`           // 开仓佣金率
	CloseTodayRate    float64 `yaml:"comm-close-today-rate,omitempty"`    // 平今佣金率
	ClosePreviousRate float64 `yaml:"comm-close-previous-rate,omitempty"` // 平昨佣金率
}

type StockFee struct {
	TransferFeeRate float64 `yaml:"transfer-fee-rate,omitempty"` // 过户费率
	TaxRate         float64 `yaml:"tax-rate,omitempty"`          // 印花税率
	CommBrokerRate  float64 `yaml:"comm-broker-rate,omitempty"`  // 佣金率
	MinFees         float64 `yaml:"min-fees,omitempty"`          // 最低佣金
}

type ContractBasic struct {
	ContractSize float64 `yaml:"contract-size,omitempty"`    // 一手数目，主板股票、ETF为100，科创板、创业板为1
	MinOrderVol  float64 `yaml:"min-order-volume,omitempty"` // 最小下单单位，主板股票、ETF为100，科创板、创业板为200
	TickSize     float64 `yaml:"tick-size,omitempty"`        // 最小变动价位 股票 0.01 etf 0.001
	TPlus        int8    `yaml:"t-plus"`                     // t+x交易 t+1
}

type StockContract struct {
	Name  string        `yaml:"name"`    // 合约名称
	Basic ContractBasic `yaml:",inline"` // 合约基本信息

	StockFee StockFee `yaml:",inline"` // 股票费用
}

func (s *StockContract) InitializeObject() error {
	s.Basic = ContractBasic{
		ContractSize: 100,
		MinOrderVol:  100,
		TickSize:     0.01,
		TPlus:        1,
	}

	s.StockFee = StockFee{
		TransferFeeRate: 0.00001,
		TaxRate:         0.001,
		CommBrokerRate:  0.0003,
		MinFees:         5,
	}

	return nil
}

type ContractProperty struct {
	TargetProp struct {
		Stock  []*StockContract  `yaml:"stock"`  // 股票合约属性
		Future []*FutureContract `yaml:"future"` // 期货合约属性
	} `yaml:"target-prop"`
}

func NewContractPropertyConfig(configFile string) (*ContractProperty, error) {
	f, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	c := ContractProperty{}
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

type ModelFunc struct {
	Symbol string `yaml:"func"`
	Weight int    `yaml:"weight"`
}

type GepModel struct {
	Function               []ModelFunc `yaml:"input-function"`           // [func][weight] 注入的函数权重
	Iteration              int         `yaml:"iteration"`                // 迭代次数
	PMutate                float64     `yaml:"pmutate"`                  // 变异概率
	Pis                    float64     `yaml:"pis"`                      // 插入概率
	Glis                   int         `yaml:"glis"`                     // 插入长度
	Pris                   float64     `yaml:"pris"`                     // 重复概率
	Glris                  int         `yaml:"glris"`                    // 重复长度
	PGene                  float64     `yaml:"pgene"`                    // 基因突变概率
	P1p                    float64     `yaml:"p1p"`                      // 一点交叉概率
	P2p                    float64     `yaml:"p2p"`                      // 两点交叉概率
	Pr                     float64     `yaml:"pr"`                       // 交换概率
	NumGenomeSet           int         `yaml:"num-genomeset"`            // 基因组集合数量
	NumGenomes             int         `yaml:"num-genome"`               // 基因组数量
	HeadSize               int         `yaml:"head-size"`                // 头部长度
	NumGenomesPerGenomeSet int         `yaml:"num-genome-per-genomeset"` // 每个基因组集合中基因组数量
	NumGenesPerGenome      int         `yaml:"num-gene-per-genome"`      // 每个基因组中基因数量
	// NumTerminals           int         // 终端数量，即输入指标数量
	NumConstants int       `yaml:"num-constants"` // 常量数量
	LinkFunc     string    `yaml:"link-func"`     // 连接函数
	Mode         ModelType `yaml:"mode"`          // 模式
}

type Model struct {
	ID          string `yaml:"model-id"`          // 模型ID
	Name        string `yaml:"model-name"`        // 模型名称
	Description string `yaml:"model-description"` // 模型描述

	Gep *GepModel `yaml:"gep"` // GEP模型参数

}

func NewModelConfig(configFile string) (*Model, error) {
	f, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	c := Model{}
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

type Performance struct {
	RiskFreeRate    float64 `yaml:"risk-free-rate"`   // 无风险利率
	PerformanceType string  `yaml:"performance-type"` // 评估指标类型
	ExpectFitness   float64 `yaml:"expect-fitness"`   // 满意预期，达到即终止，可设定大一些，迫使算法持续搜索
}

type TradeAcc struct {
	Username string      `yaml:"username,omitempty"` // 用户名
	Password string      `yaml:"password,omitempty"` // 密码
	URL      string      `yaml:"url,omitempty"`      // 地址
	Tunnel   HandlerType `yaml:"tunnel,omitempty"`   // 协议: 期货CTP，股票VMT
}

type Account struct {
	Cash     float64   `yaml:"cash,omitempty"`     // 初始资金
	Count    int       `yaml:"count,omitempty"`    // 拉取
	Slippage float64   `yaml:"slippage,omitempty"` // 滑点
	Account  *TradeAcc `yaml:"account,omitempty"`  // 账户信息
}

type Framework struct {
	Stock  Account `yaml:"stock,omitempty"`  // 股票
	Future Account `yaml:"future,omitempty"` // 期货

	GroupInstrument []string `yaml:"group,omitempty"`      // 股票组合
	Instrument      []string `yaml:"instrument,omitempty"` // 合约标的
	Indicator       []string `yaml:"indicator,omitempty"`  // 参与的指标

	Realtime            bool          `yaml:"realtime,omitempty"`           // 是否实盘，对于回测无效
	Frequency           Frequency     `yaml:"frequency,omitempty"`          // 1min, 5min, 15min, 30min, 60min, 1day, 1week, 1month
	BeginTime           string        `yaml:"begin-time,omitempty"`         // 启动时间
	EndTime             string        `yaml:"end-time,omitempty"`           // 结束时间
	BeginDate           string        `yaml:"begin,omitempty"`              // 启动日期
	EndDate             string        `yaml:"end,omitempty"`                // 结束日期
	DailyTriggerTimeStr string        `yaml:"daily-trigger-time,omitempty"` // 每日触发时间
	Begin               time.Time     `yaml:",inline"`                      // 启动
	End                 time.Time     `yaml:",inline"`                      // 结束
	DailyTriggerTime    time.Duration // 每日触发时间, 默认为14:50:00
}

type System struct {
	ReplayMatcher        HandlerType `yaml:"replay-matcher,omitempty"`    // 回放匹配器
	AccountHandlerType   HandlerType `yaml:"account-handler,omitempty"`   // 账户处理器类型
	FormulaHandlerType   HandlerType `yaml:"formula-handler,omitempty"`   // 公式处理器类型
	QuoteHandlerType     HandlerType `yaml:"quote-handler,omitempty"`     // 行情处理器类型
	IndicatorHandlerType HandlerType `yaml:"indicator-handler,omitempty"` // 指标处理器类型
	ReplayHandlerType    HandlerType `yaml:"replay-handler,omitempty"`    // 回放处理器类型
	ContractHandlerType  HandlerType `yaml:"contract-handler,omitempty"`  // 合约处理器类型
	XrxdHandlerType      HandlerType `yaml:"xrxd-handler,omitempty"`      // 除权除息处理器类型
	ModelHandlerType     HandlerType `yaml:"model-handler,omitempty"`     // 模型处理器类型 gep|manual
	RecordHandlerType    HandlerType `yaml:"record-handler,omitempty"`    // 记录处理器类型 csv|memory|sqlite
	DataType             HandlerType `yaml:"data-type,omitempty"`         // 数据模式 全复权模式|前复权模式
	TunnelType           HandlerType `yaml:"tunnel-type,omitempty"`       // 隧道模式 默认=vmt
}

type Tunnel struct {
	Host string `yaml:"host"`
	Port int32  `yaml:"port"`
}

func (t *Tunnel) InitializeObject() error {
	t.Host = "127.0.0.1"
	t.Port = 20613

	return nil
}

type DataSource struct {
	Name   string            `yaml:"name"`   // 数据源名称
	Type   string            `yaml:"type"`   // 数据源类型
	Url    string            `yaml:"url"`    // 数据源地址
	Params map[string]string `yaml:"params"` // 数据源参数
}

type Runtime struct {
	Mode Mode   // 运行模式
	ID   string // 运行ID
	Name string // 运行名称
	Path *Path  // 运行路径

	System                    System         // 系统配置
	Indicator2FormulaVarIndex map[string]int // 指标名称到公式的映射，如果缺失则无法计算，可以多余公式需要的指标，但不能少于公式需要的指标

	TrainID      string `yaml:"train-id"`      // 训练ID
	TrainName    string `yaml:"train-name"`    // 训练名称
	BacktestID   string `yaml:"backtest-id"`   // 回测ID
	BacktestName string `yaml:"backtest-name"` // 回测名称
	RuntimeID    string `yaml:"runtime-id"`    // 运行ID
	RuntimeName  string `yaml:"runtime-name"`  // 运行名称

	Model     *Model             // 模型
	Contract  *ContractProperty  // 合约
	Indicator *IndicatorProperty // 指标

	Tunnel      *Tunnel      `yaml:"tunnel,omitempty"`    // 隧道
	Performance Performance  `yaml:"performance"`         // 评估指标
	Framework   Framework    `yaml:"framework,omitempty"` // 运行参数
	DataSource  []DataSource `yaml:"datasource"`          // 数据源
}

func (rt *Runtime) NewConfig(configFile string) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	f, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(f, rt)
	if err != nil {
		return err
	}

	return nil
}

func checkFramework(framework *Framework) {
	if framework.Frequency == "" {
		ErrorF("频率不能为空")
	}

	if framework.Frequency != Frequency1Min && framework.Frequency != Frequency5Min &&
		framework.Frequency != Frequency15Min && framework.Frequency != Frequency30Min &&
		framework.Frequency != Frequency1Day && framework.Frequency != Frequency120Min {
		ErrorF("频率[%s]不支持", framework.Frequency)
	}

	if framework.BeginDate != "" {
		begin, err := time.ParseInLocation(TimeFormatDate, framework.BeginDate, time.Local)
		if err != nil {
			ErrorF("开始日期格式错误：%v", err)
		}

		framework.Begin = begin
	} else if framework.BeginTime != "" {
		begin, err := time.ParseInLocation(TimeFormatDate, framework.BeginTime[:8], time.Local)
		if err != nil {
			ErrorF("开始时间格式错误：%v", err)
		}

		framework.Begin = begin
	} else {
		ErrorF("开始时间不能为空")
	}

	if framework.EndDate != "" {
		end, err := time.ParseInLocation(TimeFormatDate, framework.EndDate, time.Local)
		if err != nil {
			ErrorF("结束日期格式错误：%v", err)
		}

		framework.End = end
	} else if framework.EndTime != "" && framework.EndDate == "" {
		end, err := time.ParseInLocation(TimeFormatDate, framework.EndTime[:8], time.Local)
		if err != nil {
			ErrorF("结束时间格式错误：%v", err)
		}

		framework.End = end
	} else {
		ErrorF("结束时间不能为空")
	}

	if framework.Begin.After(framework.End) {
		ErrorF("开始时间不能晚于结束时间")
	}

	// 每日触发时间
	if framework.DailyTriggerTimeStr != "" {
		framework.DailyTriggerTime = times.MustDuration(TimeFormatHHMM, framework.DailyTriggerTimeStr)
	} else {
		framework.DailyTriggerTime = times.MustDuration(TimeFormatHHMM, "14:50")
	}

}

func checkModel(model *Model) {
	if model.Gep.NumGenesPerGenome < 2 {
		ErrorF("每个基因组的基因数量不能小于2")
	}

	if model.Gep.NumGenomesPerGenomeSet < 2 {
		ErrorF("每个基因组合的基因组数量不能小于2")
	}

	if model.Gep.NumGenomeSet < 2 {
		ErrorF("基因组集合数量不能小于2")
	}

	if model.Gep.NumGenomes < 2 {
		ErrorF("基因组数量不能小于2")
	}
}

func randomID() string {
	return strings.Replace(time.Now().Format(TimeFormatRange), ".", "-", 1)
}

func New(mode Mode, dir *Path) *Runtime {
	r := Runtime{
		Path:                      dir,
		Indicator2FormulaVarIndex: make(map[string]int),
		System: System{
			ReplayMatcher:        HandlerTypeDefault,
			AccountHandlerType:   HandlerTypeDefault,
			FormulaHandlerType:   HandlerTypeDefault,
			QuoteHandlerType:     HandlerTypeDefault,
			IndicatorHandlerType: HandlerTypeDefault,
			ReplayHandlerType:    HandlerTypeDefault,
			ContractHandlerType:  HandlerTypeDefault,
			XrxdHandlerType:      HandlerTypeDefault,
			ModelHandlerType:     HandlerTypeDefault,
			RecordHandlerType:    HandlerTypeDefault,
			DataType:             HandlerTypeDefault,
			TunnelType:           HandlerTypeDefault,
		},
		Tunnel: &Tunnel{
			Host: "127.0.0.1",
			Port: 20613,
		},
	}

	// 读取配置
	if mode != CalcMode {
		model, err := NewModelConfig(dir.ModelConfigFile)
		if err != nil {
			WarnF("读取模型配置失败: %s", err)
		} else {
			checkModel(model)
			r.Model = model
		}

		// 读取合约配置
		contract, err := NewContractPropertyConfig(dir.ContractFile)
		if err != nil {
			ErrorF("读取合约配置失败: %s", err)
		}

		r.Contract = contract
	}

	// 读取指标配置
	indicator, err := NewIndicatorConfig(dir.IndicatorFile)
	if err != nil {
		ErrorF("读取指标配置失败: %s", err)
	}

	r.Indicator = indicator

	// 读取框架配置, FrameworkConfigFile可能不存在
	err = r.NewConfig(dir.FrameworkConfigFile)
	if err != nil {
		ErrorF("读取vqt配置失败: %s", err)
	}

	err = r.NewConfig(dir.DataConfigFile)
	if err != nil {
		ErrorF("读取数据配置失败: %s", err)
	}

	switch mode {
	case TrainMode:
		// 读取训练配置
		err := r.NewConfig(dir.TrainConfigFile)
		if err != nil {
			ErrorF("读取训练配置失败: %s", err)
		}

		r.ID = r.TrainID
		r.Name = r.TrainName
		r.Mode = mode

		for i, v := range r.Framework.Indicator {
			r.Indicator2FormulaVarIndex[v] = i
		}

	case BTMode:
		// 读取训练配置
		err := r.NewConfig(dir.TrainConfigFile)
		if err != nil {
			ErrorF("读取训练配置失败: %s", err)
		}

		// 建立指标名称到公式参数index的映射
		for i, v := range r.Framework.Indicator {
			r.Indicator2FormulaVarIndex[v] = i
		}

		// 读取回测配置
		err = r.NewConfig(dir.BackTestConfigFile)
		if err != nil {
			ErrorF("读取回测配置失败: %s", err)
		}

		r.ID = r.BacktestID
		r.Name = r.BacktestName
		r.Mode = mode

		// 检查训练配置的指标是否在回测配置中
		for _, v := range r.Framework.Indicator {
			if _, ok := r.Indicator2FormulaVarIndex[v]; !ok {
				ErrorF("训练集中的指标[%s]在回测配置中不存在，无法完成模型回测", v)
			}
		}
	case RunMode:
		// 读取训练配置
		err := r.NewConfig(dir.TrainConfigFile)
		if err != nil {
			ErrorF("读取训练配置失败: %s", err)
		}

		// 读取回测配置
		err = r.NewConfig(dir.BackTestConfigFile)
		if err != nil {
			ErrorF("读取回测配置失败: %s", err)
		}

		// 读取运行配置
		err = r.NewConfig(dir.RunTimeConfigFile)
		if err != nil {
			ErrorF("读取运行配置失败: %s", err)
		}

		r.ID = r.RuntimeID
		r.Name = r.RuntimeName
		r.Mode = mode

		// 建立指标名称到公式参数index的映射
		for i, v := range r.Framework.Indicator {
			r.Indicator2FormulaVarIndex[v] = i
		}

		if r.Framework.Stock.Account != nil {
			if r.Framework.Stock.Account.Tunnel == "" {
				r.Framework.Stock.Account.Tunnel = HandlerTypeKsft
			}

			if r.Framework.Stock.Account.URL == "" {
				r.Framework.Stock.Account.URL = "ksft://127.0.0.1:3333"
			}
		}
	}

	// 日期和时间
	checkFramework(&r.Framework)

	// 检查指标是否在指标配置中存在或者为常规高开低收、成交量、成交价指标
	if invalidKey, ok := slice.ContainsFunc(
		r.Framework.Indicator,
		r.Indicator.Indicator,
		func(a string) string {
			return a
		},
		func(v Formula) string {
			return v.Name
		},
	); !ok {
		ErrorF("指标[%s]在指标配置中不存在，无法运行", *invalidKey)
	}

	// 去重
	r.Framework.Indicator = slices.CompactFunc(
		r.Framework.Indicator, func(i, j string) bool {
			return i == j
		},
	)

	r.Indicator.Indicator = slices.CompactFunc(
		r.Indicator.Indicator, func(i, j Formula) bool {
			return i.Name == j.Name
		},
	)

	r.ID = randomID()
	r.Mode = mode

	return &r
}
