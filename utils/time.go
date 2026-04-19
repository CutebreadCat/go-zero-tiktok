package myutils

import (
	"time"
)

func TsToStr(ts int64, layout string) string {
	if layout == "" {
		layout = "2006-01-02 15:04:05" // 默认格式
	}
	return time.Unix(ts, 0).Format(layout)
} //将时间戳直接转化为字符串
