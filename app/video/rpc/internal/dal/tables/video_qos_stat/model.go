package video_qos_stat

import "time"

// VideoQoSStat 视频播放质量聚合指标模型。
type VideoQoSStat struct {
	VideoID        int64     `gorm:"primaryKey;type:bigint;column:video_id"`
	CompletionRate int32     `gorm:"default:0;type:int;column:completion_rate"`
	StallRate      int32     `gorm:"default:0;type:int;column:stall_rate"`
	ErrorRate      int32     `gorm:"default:0;type:int;column:error_rate"`
	AvgBitrateKbps int32     `gorm:"default:0;type:int;column:avg_bitrate_kbps"`
	AvgBufferedMs  int64     `gorm:"default:0;type:bigint;column:avg_buffered_ms"`
	AvgStallCount  int32     `gorm:"default:0;type:int;column:avg_stall_count"`
	ReportCount    int64     `gorm:"default:0;type:bigint;column:report_count"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (VideoQoSStat) TableName() string {
	return "video_qos_stat"
}
