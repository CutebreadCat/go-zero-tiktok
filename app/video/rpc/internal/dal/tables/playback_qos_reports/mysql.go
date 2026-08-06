package playback_qos_reports

import (
	"context"

	"go_zero-tiktok/pkg/xerr"

	"gorm.io/gorm"
)

// CreateReport 写入一条播放质量上报记录
func CreateReport(ctx context.Context, db *gorm.DB, report *PlaybackQoSReport) error {
	if report == nil {
		return xerr.NewInvalidParam("上报数据为空")
	}
	if err := db.WithContext(ctx).Create(report).Error; err != nil {
		return xerr.Wrap(err, "create playback qos report failed")
	}
	return nil
}

// GetReportsByVideoID 按视频分页查询播放质量上报记录
func GetReportsByVideoID(ctx context.Context, db *gorm.DB, videoID int64, pageNum, pageSize int32) ([]PlaybackQoSReport, error) {
	var rows []PlaybackQoSReport
	if err := db.WithContext(ctx).
		Where("video_id = ?", videoID).
		Order("id DESC").
		Limit(int(pageSize)).
		Offset(int((pageNum - 1) * pageSize)).
		Find(&rows).Error; err != nil {
		return nil, xerr.Wrap(err, "get playback qos reports failed")
	}
	return rows, nil
}