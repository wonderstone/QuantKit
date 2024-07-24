package framework

import (
	"path/filepath"

	"github.com/wonderstone/QuantKit/framework/entity/handler"
	_ "github.com/wonderstone/QuantKit/framework/logic/runner"
	"github.com/wonderstone/QuantKit/framework/logic/strategy"
	"github.com/wonderstone/QuantKit/framework/setting"
	// _ "github.com/wonderstone/QuantKit/ksft"
	"github.com/wonderstone/QuantKit/config"
)

type Op struct {
	Arg
	RunID          string           // RunID
	StrategyID     string           // 策略ID
	Mode           config.Mode      // 模式 calc|train|bt|runtime
	PathStyle      PathStyle        // 路径风格 vmt|none
	Freq           config.Frequency // 频率
	InstID         []string         // 合约ID
	WorkRoot       string           // 工作根路径 vmt风格: ${workdir}
	StrategyPrefix string           // 策略根路径 vmt风格: ${workdir}/strategy
	CommonRoot     string           // 公共根路径 vmt风格: ${workdir}/common
	DataPrefix     string           // 数据根路径 vmt风格: ${workdir}/strategy/${StrategyID}/.vqt
	DownloadPrefix string           // 下载路径 vmt风格: ${workdir}/strategy/${StrategyID}/.vqt/download
	BasePrefix     string           // 基础数据路径 vmt风格: ${workdir}/strategy/${StrategyID}/.vqt/download/base
	CalcPrefix     string           // 数据路径 vmt风格: ${workdir}/strategy/${StrategyID}/.vqt/calc
	InputPrefix    string           // 输入参数路径 vmt风格: ${workdir}/strategy/${StrategyID}/.vqt/${ModePrefix}/input
	OutputPrefix   string           // 输出参数路径 vmt风格: ${workdir}/strategy/${StrategyID}/.vqt/${ModePrefix}/output
	ModePrefix     bool             // 模式路径前缀 vmt风格: ${workdir}/strategy/${StrategyID}/.vqt/${Mode}

	StrategyCreator handler.StrategyCreator // 策略创建器
}

type WithOption func(*Op)

func WithMode(mode config.Mode) WithOption {
	return func(op *Op) {
		op.Mode = mode
	}
}

func WithFrequency(freq config.Frequency) WithOption {
	return func(op *Op) {
		op.Freq = freq
	}
}

func WithInstID(instID []string) WithOption {
	return func(op *Op) {
		op.InstID = instID
	}
}

func WithPanicExport() WithOption {
	return func(op *Op) {
		op.PanicExport = true
	}
}

func WithWorkRoot(workRoot string) WithOption {
	return func(op *Op) {
		op.WorkRoot = workRoot
	}
}

func WithPathStyle(pathStyle PathStyle) WithOption {
	return func(op *Op) {
		op.PathStyle = pathStyle
	}
}

func WithStrategyPrefix(dir string) WithOption {
	return func(op *Op) {
		op.StrategyPrefix = dir
	}
}

func WithStrategyID(strategyID string) WithOption {
	return func(op *Op) {
		op.StrategyID = strategyID
	}
}

func WithDataPrefix(dataRoot ...string) WithOption {
	return func(op *Op) {
		if len(dataRoot) == 0 {
			op.DataPrefix = ".vqt"
		} else {
			op.DataPrefix = dataRoot[0]
		}
	}
}

func WithDownloadPrefix(downloadDir ...string) WithOption {
	return func(op *Op) {
		if len(downloadDir) == 0 {
			op.DownloadPrefix = "download"
		} else {
			op.DownloadPrefix = downloadDir[0]
		}
	}
}

func WithInputPrefix(input ...string) WithOption {
	return func(op *Op) {
		if len(input) == 0 {
			op.InputPrefix = "input"
		} else {
			op.InputPrefix = input[0]
		}
	}
}

func WithOutPrefix(output ...string) WithOption {
	return func(op *Op) {
		if len(output) == 0 {
			op.OutputPrefix = "output"
		} else {
			op.OutputPrefix = output[0]
		}
	}
}

func WithCalcPrefix(output ...string) WithOption {
	return func(op *Op) {
		if len(output) == 0 {
			op.CalcPrefix = "calc"
		} else {
			op.CalcPrefix = output[0]
		}
	}
}

func WithModePrefix() WithOption {
	return func(op *Op) {
		op.ModePrefix = true
	}
}

func WithCommonPrefix(commonRoot ...string) WithOption {
	return func(op *Op) {
		if len(commonRoot) == 0 {
			op.CommonRoot = "common"
		} else {
			op.CommonRoot = commonRoot[0]
		}
	}
}

func WithRunID(runID string) WithOption {
	return func(op *Op) {
		op.RunID = runID
	}
}

func WithStrategyCreator(creator handler.StrategyCreator) WithOption {
	return func(op *Op) {
		op.StrategyCreator = creator
	}
}

func WithStrategyDemo(creator string) WithOption {
	return func(op *Op) {
		switch creator {
		case "T0":
			op.StrategyCreator = func() handler.Strategy {
				return &strategy.T0{}
			}
		}
	}
}

func NewOp(options ...WithOption) *Op {
	op := &Op{
		Arg:            args,
		RunID:          args.RID,
		StrategyID:     args.SID,
		Mode:           args.Mode,
		WorkRoot:       args.WorkDir,
		PathStyle:      args.PathStyle,
		DownloadPrefix: "download",
		BasePrefix:     "base",
		CalcPrefix:     "calc",
	}

	// 如果设置了策略演示模式，则使用演示模式
	if args.Strategy != "" {
		WithStrategyDemo(args.Strategy)(op)
	}

	for _, opt := range options {
		opt(op)
	}
	return op
}

func Run(options ...WithOption) {
	defer config.CatchPanic()

	op := NewOp(options...)

	// 检查参数
	if op.Mode == "" {
		panic("未设置运行模式, 可选择的模式为[calc, bt, train, runtime]")
	}

	modePrefix := ""

	switch op.Mode {
	case config.TrainMode, config.BTMode, config.RunMode:
		if op.StrategyCreator == nil {
			panic("未设置策略创建器或者没有选择策略演示模式[T0, DMT, ...]")
		}
		fallthrough
	case config.CalcMode:
	default:
		panic("未设置正确运行模式(mode), 可选择的模式为[calc, bt, train, runtime]")
	}

	if op.ModePrefix {
		modePrefix = string(op.Mode)
	}

	var dir *config.Path = nil
	switch op.PathStyle {
	case PathStyleNone:
		dir = config.NewDefaultPath(
			op.Mode,
			config.WithRoot(op.WorkRoot),
			config.WithStrategyID(op.StrategyPrefix, op.StrategyID),
			config.WithRunDir(filepath.Join(op.DataPrefix, modePrefix, op.RunID)),
			config.WithDownloadDir(filepath.Join(op.DataPrefix, op.DownloadPrefix)),
			config.WithBaseDir(filepath.Join(op.DataPrefix, op.BasePrefix)),
			config.WithIndicatorDir(filepath.Join(op.DataPrefix, op.CalcPrefix)),
			config.WithInputPrefix(op.InputPrefix),
			config.WithOutputPrefix(op.OutputPrefix),
		)
	case PathStyleVMT:
		if op.StrategyID == "" {
			panic("未设置策略ID(--sid)")
		}

		dir = config.NewDefaultPath(
			op.Mode,
			config.WithRoot(op.WorkRoot),
			config.WithStrategyID("strategy", op.StrategyID),
			config.WithRunDir(filepath.Join(".vqt", string(op.Mode), op.RunID)),
			config.WithDownloadDir(filepath.Join(".vqt", "download")),
			config.WithBaseDir("base"),
			config.WithIndicatorDir(filepath.Join(".vqt", "calc")),
			config.WithInputPrefix("input"),
			config.WithOutputPrefix("output"),
		)
	}

	// 初始化日志，至此可以使用logger进行日志输出
	config.NewLogger(
		dir.InfoFile, dir.ErrorFile, dir.StatusFile, op.PanicExport, map[string]any{
			"id":   op.RunID,
			"mode": op.Mode,
		},
		// ! 不一致
		// op.Mode == config.RunMode,
	)

	conf := config.New(op.Mode, dir)
if op.RunID != "" {
		conf.ID = op.RunID
	}

	config.DebugF("启动运行模式: %s", conf.Mode)
	config.DebugF("启动运行id: %s", conf.ID)
	config.DebugF("启动运行路径: %s", conf.Path.Root)

	// 处理器设置
	if len(op.Handlers) > 0 {
		for k, v := range op.Handlers {
			switch k {
			case "account":
				conf.System.AccountHandlerType = config.HandlerType(v)
			case "formula":
				conf.System.FormulaHandlerType = config.HandlerType(v)
			case "record":
				conf.System.RecordHandlerType = config.HandlerType(v)
			}
		}
	}

	runner := setting.MustNewRunner(op.Mode)

	err := runner.Init(setting.WithRuntimeConfig(conf), setting.WithStrategyCreator(op.StrategyCreator))
	if err != nil {
		config.ErrorF("初始化失败 %e", err)
	}

	err = runner.Start()
	if err != nil {
		config.ErrorF("运行失败 %e", err)
	}
}
