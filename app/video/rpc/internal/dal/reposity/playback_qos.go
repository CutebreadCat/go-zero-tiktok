package reposity

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	qosreporttable "go_zero-tiktok/app/video/rpc/internal/dal/tables/playback_qos_reports"
	"go_zero-tiktok/pkg/contract"

	"github.com/go-sql-driver/mysql"
	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

// PlaybackQoSRepo 播放质量上报仓库，负责幂等落库。
type PlaybackQoSRepo struct {
	db *gorm.DB
}

// NewPlaybackQoSRepo 创建播放质量上报仓库。
func NewPlaybackQoSRepo(db *gorm.DB) *PlaybackQoSRepo {
	return &PlaybackQoSRepo{db: db}
}

// CreateReport 创建一条播放质量上报记录。
// 对 (user_id, idempotency_key) 唯一键冲突视为已上报，直接返回成功（幂等）。
func (r *PlaybackQoSRepo) CreateReport(ctx context.Context, report *types.PlaybackQoSReport) error {
	if report == nil {
		return pkgerrors.New("PlaybackQoSRepo.CreateReport: report is nil")
	}

	reportData, err := buildReportDataJSON(report)
	if err != nil {
		return pkgerrors.WithMessage(err, "PlaybackQoSRepo.CreateReport")
	}

	row := &qosreporttable.PlaybackQoSReport{
		UserID:         report.UserID,
		VideoID:        report.VideoID,
		ReportData:     reportData,
		IdempotencyKey: report.IdempotencyKey,
	}

	if err := qosreporttable.CreateReport(ctx, r.db, row); err != nil {
		if isDuplicateKeyError(err) {
			// 唯一键冲突：幂等返回成功
			return nil
		}
		return pkgerrors.WithMessage(err, "PlaybackQoSRepo.CreateReport")
	}
	return nil
}

// isDuplicateKeyError 判断是否为唯一键冲突。
// 生产环境使用 MySQL（Error 1062），单测使用 SQLite（UNIQUE constraint failed）。
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// buildReportDataJSON 将上报指标序列化为 JSON，存入 report_data 字段。
func buildReportDataJSON(report *types.PlaybackQoSReport) (string, error) {
	data := map[string]any{
		"event_type":       report.EventType,
		"duration_ms":      report.DurationMs,
		"played_ms":        report.PlayedMs,
		"buffered_ms":      report.BufferedMs,
		"stall_count":      report.StallCount,
		"stall_total_ms":   report.StallTotalMs,
		"resolution":       report.Resolution,
		"bitrate_kbps":     report.BitrateKbps,
		"fps":              report.Fps,
		"error_code":       report.ErrorCode,
		"error_msg":        report.ErrorMsg,
		"network_type":     report.NetworkType,
		"device_info":      report.DeviceInfo,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
