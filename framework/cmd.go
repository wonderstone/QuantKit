package framework

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wonderstone/QuantKit/config"
)

type PathStyle string

const (
	PathStyleNone PathStyle = ""
	PathStyleVMT  PathStyle = "vmt"
)

// Arg 量化命令
// WorkDir: 工作目录 和 RID 二选一
type Arg struct {
		// 以下参数如果没有指定任意一个，则不会使用对应模式，而是使用手动模式，用户需要自行指定各种运行的初始信息
	RID      string // 运行ID
	SID      string // 策略ID
	WorkDir  string // 工作目录
	Strategy string // 策略

	PathStyle PathStyle // 样式,支持vmt客户端风格的路径布局

	Mode config.Mode // 模式

	Handlers map[string]string // 处理器列表, key: [account, formula, record]

	PanicExport bool // 是否导出panic信息
}

var args Arg

func parseCmd(vqt *Arg) error {
	cmd := &cobra.Command{
		Use: "vqt [flags]...",
		// SilenceUsage: true,
		DisableSuggestions: true,
		Example: `  vqt --mode=calc        计算指标
  vqt --mode=train       训练模型
  vqt --mode=bt          回测运行
  vqt --mode=runtime     实盘运行`,
	}

	var pwd, _ = os.Getwd()
	cmd.Flags().StringVarP(&vqt.WorkDir, "workdir", "w", pwd, "工作目录")

	cmd.Flags().StringVarP(&vqt.RID, "rid", "", "", "运行ID(可选)")
	cmd.Flags().StringVarP(&vqt.SID, "sid", "", "", "策略ID(可选)")

	var mode string
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "运行模式(指标计算: calc, 训练: train, 回测: bt, 运行: runtime)")

	var pathStyle string
	cmd.Flags().StringVarP(&pathStyle, "style", "s", "", "路径样式")
	err := cmd.Flags().MarkHidden("style")
	if err != nil {
		return err
	}

	cmd.Flags().StringVarP(&vqt.Strategy, "strategy", "", "", "策略实例")
	err = cmd.Flags().MarkHidden("strategy")
	if err != nil {
		return err
	}

	cmd.Flags().BoolVar(&vqt.PanicExport, "panic", true, "是否导出panic信息")

	cmd.Flags().StringToStringVarP(
		&vqt.Handlers, "handler", "H", nil,
		"处理器列表, key: [account, formula, record], value: [default, dailymode, csv, sqlite, ...]",
	)
	err = cmd.Flags().MarkHidden("handler")
	if err != nil {
		return err
	}

	help := false

	cmd.Flags().BoolVarP(&help, "help", "h", false, "帮助")

	err = cmd.Execute()
	if err != nil {
		return err
	}

	if help {
		println(cmd.UsageString())
		os.Exit(0)
	}

	switch pathStyle {
	case "vmt":
		vqt.PathStyle = PathStyleVMT
	default:
		vqt.PathStyle = PathStyleNone
	}

	switch mode {
	case "calc":
		vqt.Mode = config.CalcMode
	case "train":
		vqt.Mode = config.TrainMode
	case "bt":
		vqt.Mode = config.BTMode
	case "rt":
		vqt.Mode = config.RunMode
	default:

	}

	return nil
}

func init() {
	err := parseCmd(&args)
	if err != nil {
		panic(err)
	}
}
