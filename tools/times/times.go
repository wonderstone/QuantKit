package times

import "time"

var Zero = time.Date(0, 1, 1, 0, 0, 0, 0, time.Local)

func Duration(layout, value string) (time.Duration, error) {
	t, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		return 0, err
	}

	return t.Sub(Zero), nil
}

// MustDuration 输出时间间隔，如果解析失败则 panic
func MustDuration(layout, value string) time.Duration {
	t, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		panic(err)
	}

	return t.Sub(Zero)
}
