package config

import (
	"os"
	"path"
	"path/filepath"
)

type Path struct {
	Root      string // 根目录
	Strategy  string // 策略目录
	Common    string // 公共参数根目录
	Run       string // 运行参数根目录
	Download  string // 下载数据根目录
	Base      string // 基础数据根目录
	Indicator string // 计算数据根目录
	Output    string // 计算结果根目录
	Input     string // 配置文件根目录
	Result    string // 结果文件根目录
	State     string // 状态文件根目录

	// 以下为策略文件的一些命名
	// 策略文件
	StrategyFile string // 策略文件

	// 使用到的基础数据文件
	XrxdFile string // 股票除权除息文件

	// 2. 公共参数文件
	IndicatorFile string // 指标计算文件
	ContractFile  string // 合约属性文件

	// 3. 运行参数文件
	FrameworkConfigFile string // 框架配置文件
	DataConfigFile      string // 下载数据文件
	ModelConfigFile    string // 模型配置文件
	TrainConfigFile    string // 训练配置文件
	BackTestConfigFile string // 回测配置文件
	RunTimeConfigFile  string // 运行配置文件

	// 4. 运行中描述文件
	PidFile    string // PID文件
	StatusFile string // 状态日志文件
	ErrorFile  string // 错误日志文件
	InfoFile   string // 信息日志文件

	// 5. 运行结果文件
	KarvaExpressionFile string // 卡尔瓦表达式文件
	AccountResultFile   string // 账户结果文件
	OrderResultFile     string // 订单结果文件
	PositionResultFile  string // 持仓结果文件
}

type WithOption func(*Path)

func WithRoot(rootDir string) WithOption {
	return func(p *Path) {
		p.Root = rootDir
	}
}

func WithStrategyID(rootDir, ID string) WithOption {
	return func(p *Path) {
		// 如果没有设置策略目录，则使用根目录
		if rootDir == "" {
			p.Strategy = p.Root
		} else
		// 如果设置了绝对路径，则使用绝对路径
		if filepath.IsAbs(rootDir) {
			p.Strategy = rootDir
		} else
		// 如果设置了相对路径，则使用相对路径，且加入ID
		{
			p.Strategy = path.Join(p.Root, rootDir, ID)
		}
	}
}

func WithRunDir(rootDir string) WithOption {
	return func(p *Path) {
		// 如果没有设置策略目录，则使用根目录
		if rootDir == "" {
			p.Run = p.Strategy
		} else
		// 如果设置了绝对路径，则使用绝对路径
		if filepath.IsAbs(rootDir) {
			p.Run = rootDir
		} else
		// 如果设置了相对路径，则使用相对路径
		{
			p.Run = path.Join(p.Strategy, rootDir)
		}
	}
}

func WithCommonDir(rootDir string) WithOption {
	return func(p *Path) {
		if rootDir == "" {
			p.Common = p.Root
		} else
		// 如果设置了绝对路径，则使用绝对路径
		if filepath.IsAbs(rootDir) {
			p.Common = rootDir
		} else
		// 如果设置了相对路径，则使用相对路径
		{
			p.Common = path.Join(p.Root, rootDir)
		}
	}
}

func WithDownloadDir(rootDir string) WithOption {
	return func(p *Path) {
		if rootDir == "" {
			p.Download = p.Strategy
		} else
		// 如果设置了绝对路径，则使用绝对路径
		if filepath.IsAbs(rootDir) {
			p.Download = rootDir
		} else
		// 如果设置了相对路径，则使用相对路径
		{
			p.Download = path.Join(p.Strategy, rootDir)
		}
	}
}

func WithBaseDir(rootDir string) WithOption {
	return func(p *Path) {
		if rootDir == "" {
			p.Base = p.Download
		} else
		// 如果设置了绝对路径，则使用绝对路径
		if filepath.IsAbs(rootDir) {
			p.Base = rootDir
		} else
		// 如果设置了相对路径，则使用相对路径
		{
			p.Base = path.Join(p.Download, rootDir)
		}
	}
}

func WithIndicatorDir(rootDir string) WithOption {
	return func(p *Path) {
		if rootDir == "" {
			p.Indicator = p.Strategy
		} else
		// 如果设置了绝对路径，则使用绝对路径
		if filepath.IsAbs(rootDir) {
			p.Indicator = rootDir
		} else
		// 如果设置了相对路径，则使用相对路径
		{
			p.Indicator = path.Join(p.Strategy, rootDir)
		}
	}
}

func WithInputPrefix(prefix string) WithOption {
	return func(p *Path) {
		p.Input = path.Join(p.Run, prefix)
	}
}

func WithOutputPrefix(prefix string) WithOption {
	return func(p *Path) {
		p.Output = path.Join(p.Run, prefix)
	}
}

func WithExpressionFileExport() WithOption {
	return func(p *Path) {
		p.KarvaExpressionFile = path.Join(p.Output, "expression.yaml")
	}
}

func WithExpressionFileImport() WithOption {
	return func(p *Path) {
		p.KarvaExpressionFile = path.Join(p.Input, "expression.yaml")
	}
}

func WithStrategyFile(fileName string) WithOption {
	return func(p *Path) {
		p.StrategyFile = fileName
	}
}

func WithIndicatorFile(fileName string) WithOption {
	return func(p *Path) {
		p.IndicatorFile = fileName
	}
}

func WithContractFile(fileName string) WithOption {
	return func(p *Path) {
		p.ContractFile = fileName
	}
}

func WithModelConfigFile(fileName string) WithOption {
	return func(p *Path) {
		p.ModelConfigFile = fileName
	}
}

func WithTrainConfigFile(fileName string) WithOption {
	return func(p *Path) {
		p.TrainConfigFile = fileName
	}
}

func WithBackTestConfigFile(fileName string) WithOption {
	return func(p *Path) {
		p.BackTestConfigFile = fileName
	}
}

func WithRunTimeConfigFile(fileName string) WithOption {
	return func(p *Path) {
		p.RunTimeConfigFile = fileName
	}
}



// NewPath 创建路径信息
func NewPath(mode Mode, options ...WithOption) *Path {
	p := Path{}

	for _, option := range options {
		option(&p)
	}

	cwd, _ := os.Getwd()
	// 生成其他目录信息
	if p.Root == "" {
		p.Root = cwd
	}

	if p.Run == "" {
		p.Run = p.Root
	}

	if p.Download == "" {
		p.Download = p.Run
	}

	if p.Base == "" {
		p.Base = p.Download
	}

	if p.Input == "" {
		p.Input = p.Run
	}

	if p.Output == "" {
		p.Output = p.Run
	}

	if p.Common == "" {
		p.Common = p.Strategy
	}

	if p.State == "" {
		p.State = path.Join(p.Run, "log")
	}

	if p.StrategyFile == "" {
		p.StrategyFile = path.Join(p.Strategy, "strategy.yaml")
	}

	if p.XrxdFile == "" {
		p.XrxdFile = path.Join(p.Base, "xrxd.csv")
	}

	if p.IndicatorFile == "" {
		p.IndicatorFile = path.Join(p.Common, "indicator.yaml")
	}

	if p.ContractFile == "" {
		p.ContractFile = path.Join(p.Common, "contract.yaml")
	}

	if p.ModelConfigFile == "" {
		p.ModelConfigFile = path.Join(p.Input, "model.yaml")
	}
	// todo 这部分为wonderstone临时添加
	if p.FrameworkConfigFile == "" {
		p.FrameworkConfigFile = path.Join(p.Input, "framework.yaml")
	}
	// todo 结束

	if p.TrainConfigFile == "" {
		if mode == TrainMode {
			p.TrainConfigFile = path.Join(p.Run, "train.yaml")
		} else {
			p.TrainConfigFile = path.Join(p.Input, "train.yaml")
		}
	}

	if p.BackTestConfigFile == "" {
		if mode == BTMode {
			p.BackTestConfigFile = path.Join(p.Run, "backtest.yaml")
		} else {
			p.BackTestConfigFile = path.Join(p.Input, "backtest.yaml")
		}
	}

	if p.RunTimeConfigFile == "" {
		if mode == RunMode {
			p.RunTimeConfigFile = path.Join(p.Run, "runtime.yaml")
		} else {
			p.RunTimeConfigFile = path.Join(p.Input, "runtime.yaml")
		}
	}

	
	p.PidFile = path.Join(p.State, ".pid")
	p.StatusFile = path.Join(p.State, ".status")
	p.ErrorFile = path.Join(p.State, "error.log")
	p.InfoFile = path.Join(p.State, "info.log")

	if p.KarvaExpressionFile == "" {
		p.KarvaExpressionFile = path.Join(p.Output, "expression.yaml")
	}

	if p.AccountResultFile == "" {
		p.AccountResultFile = path.Join(p.Output, "account")
	}

	if p.OrderResultFile == "" {
		p.OrderResultFile = path.Join(p.Output, "order")
	}

	if p.PositionResultFile == "" {
		p.PositionResultFile = path.Join(p.Output, "position")
	}

	return &p
}

func NewDefaultPath(mode Mode, options ...WithOption) *Path {
	dir := NewPath(mode, options...)
	switch mode {
	case CalcMode:
		return dir
	case TrainMode:
		WithExpressionFileExport()(dir)
		return dir
	case BTMode:
		WithExpressionFileImport()(dir)
		return dir
	case RunMode:
		WithExpressionFileImport()(dir)
		return dir
	}

	panic("未知的运行模式: " + mode)
}
