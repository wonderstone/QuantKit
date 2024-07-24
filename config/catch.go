package config

func CatchPanic() {
	if r := recover(); r != nil {
		ErrorF("运行失败: %v", r)
	}
}
