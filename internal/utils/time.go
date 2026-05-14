package myutils

import (
	"database/sql"
	"time"
)

const DefaultTimeLayout = "2006-01-02 15:04:05"

func TsToStr(ts int64, layout string) string {
	if layout == "" {
		layout = DefaultTimeLayout
	}
	return time.Unix(ts, 0).Format(layout)
}

// TimeToStr 将 time.Time 转换为字符串
func TimeToStr(t time.Time, layout string) string {
	if layout == "" {
		layout = DefaultTimeLayout
	}
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

// NullTimeToStr 将 sql.NullTime 转换为字符串
func NullTimeToStr(nt sql.NullTime, layout string) string {
	if !nt.Valid {
		return ""
	}
	return TimeToStr(nt.Time, layout)
}

// StrToTime 将字符串转换为 time.Time
func StrToTime(str string, layout string) (time.Time, error) {
	if layout == "" {
		layout = DefaultTimeLayout
	}
	return time.Parse(layout, str)
}

// TimeToNullTime 将 *time.Time 转换为 sql.NullTime
func TimeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// StrToNullTime 将字符串转换为 sql.NullTime
func StrToNullTime(str string, layout string) (sql.NullTime, error) {
	if str == "" {
		return sql.NullTime{}, nil
	}
	t, err := StrToTime(str, layout)
	if err != nil {
		return sql.NullTime{}, err
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

// NowPtr 返回当前时间的指针
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}
