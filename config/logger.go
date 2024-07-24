package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type StatusEvent string

const (
	StartingEvent StatusEvent = "启动中"
	RunningEvent  StatusEvent = "运行中"
	FinishEvent   StatusEvent = "已完成"
	FailEvent     StatusEvent = "已失败"
)

var statusLog zerolog.Logger
var infoLog zerolog.Logger
var errorLog zerolog.Logger
var panicable bool

func NewLogger(
	infoFile, errorFile, statusFile string, panicExport bool, prefixMsg map[string]any,
) {
	panicable = panicExport
	timeFormat := "2006-01-02 15:04:05"
	zerolog.TimeFieldFormat = timeFormat
	// 创建状态目录
	// 文件夹不存在则创建
	dir := filepath.Dir(infoFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			panic(fmt.Sprintf("创建日志文件夹失败: %v", err))
		}
	}

	// 文件夹不存在则创建
	dir = filepath.Dir(errorFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			panic(fmt.Sprintf("创建错误记录文件夹失败: %v", err))
		}
	}

	// 文件夹不存在则创建
	dir = filepath.Dir(statusFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			panic(fmt.Sprintf("创建状态记录文件夹失败: %v", err))
		}
	}

	// 普通日志
	logFile, _ := os.Create(infoFile)
	multi := zerolog.MultiLevelWriter(logFile)
	infoLog = zerolog.New(multi).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	// 错误日志
	logFile, _ = os.Create(errorFile)
	multi = zerolog.MultiLevelWriter(logFile)
	errorLog = zerolog.New(multi).Level(zerolog.ErrorLevel).With().Timestamp().Logger()

	// 状态日志
	logFile, _ = os.Create(statusFile)
	multi = zerolog.MultiLevelWriter(logFile)
	ctx := zerolog.New(multi).Level(zerolog.NoLevel).With()
	for k, v := range prefixMsg {
		ctx = ctx.Interface(k, v)
	}
	statusLog = ctx.Logger()

}

func DebugF(format string, a ...any) {
	infoLog.Debug().Msgf(format, a...)
	fmt.Printf(format+"\n", a...)
}

func InfoF(format string, a ...any) {
	infoLog.Info().Msgf(format, a...)
}

func WarnF(format string, a ...any) {
	infoLog.Warn().Msgf(format, a...)
}

func ErrorF(format string, a ...any) {
	infoLog.Error().Msgf(format, a...)
	errorLog.Error().Msgf(format, a...)
	statusLog.Log().Str("event", string(FailEvent)).Float64("progress", 100).Msgf(format, a...)

	panic(fmt.Sprintf(format, a...))
}

func StatusLog(event StatusEvent, progress float64, msg ...map[string]any) {
	if len(msg) == 0 {
		statusLog.Log().
			Str("event", string(event)).
			Float64("progress", progress).
			Msg("")
	} else {
		statusLog.Log().
			Str("event", string(event)).
			Float64("progress", progress).
			Fields(msg[0]).
			Msg("")
	}
}
