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

// GetReportsAfterID 按 id 游标读取待聚合的上报记录。
func (r *PlaybackQoSRepo) GetReportsAfterID(ctx context.Context, lastID int64, limit int32) ([]*types.PlaybackQoSReport, error) {
	rows, err := qosreporttable.GetReportsAfterID(ctx, r.db, lastID, limit)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "PlaybackQoSRepo.GetReportsAfterID")
	}
	return r.rowsToReports(rows)
}

// GetReportsByVideoIDs 批量读取指定视频的全部上报记录（用于重算指标）。
func (r *PlaybackQoSRepo) GetReportsByVideoIDs(ctx context.Context, videoIDs []int64) ([]*types.PlaybackQoSReport, error) {
	rows, err := qosreporttable.GetReportsByVideoIDs(ctx, r.db, videoIDs)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "PlaybackQoSRepo.GetReportsByVideoIDs")
	}
	return r.rowsToReports(rows)
}

func (r *PlaybackQoSRepo) rowsToReports(rows []qosreporttable.PlaybackQoSReport) ([]*types.PlaybackQoSReport, error) {
	result := make([]*types.PlaybackQoSReport, 0, len(rows))
	for i := range rows {
		report, err := parseReportDataJSON(&rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, report)
	}
	return result, nil
}

// buildReportDataJSON 将上报指标序列化为 JSON，存入 report_data 字段。
func buildReportDataJSON(report *types.PlaybackQoSReport) (string, error) {
	data := map[string]any{
		"event_type":     report.EventType,
		"duration_ms":    report.DurationMs,
		"played_ms":      report.PlayedMs,
		"buffered_ms":    report.BufferedMs,
		"stall_count":    report.StallCount,
		"stall_total_ms": report.StallTotalMs,
		"resolution":     report.Resolution,
		"bitrate_kbps":   report.BitrateKbps,
		"fps":            report.Fps,
		"error_code":     report.ErrorCode,
		"error_msg":      report.ErrorMsg,
		"network_type":   report.NetworkType,
		"device_info":    report.DeviceInfo,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// parseReportDataJSON 从数据库行解析上报指标。
func parseReportDataJSON(row *qosreporttable.PlaybackQoSReport) (*types.PlaybackQoSReport, error) {
	report := &types.PlaybackQoSReport{
		ID:             row.ID,
		UserID:         row.UserID,
		VideoID:        row.VideoID,
		IdempotencyKey: row.IdempotencyKey,
	}
	if row.ReportData == "" {
		return report, nil
	}
	var data struct {
		EventType    string `json:"event_type"`
		DurationMs   int64  `json:"duration_ms"`
		PlayedMs     int64  `json:"played_ms"`
		BufferedMs   int64  `json:"buffered_ms"`
		StallCount   int32  `json:"stall_count"`
		StallTotalMs int64  `json:"stall_total_ms"`
		Resolution   string `json:"resolution"`
		BitrateKbps  int32  `json:"bitrate_kbps"`
		Fps          int32  `json:"fps"`
		ErrorCode    int32  `json:"error_code"`
		ErrorMsg     string `json:"error_msg"`
		NetworkType  string `json:"network_type"`
		DeviceInfo   string `json:"device_info"`
	}
	if err := json.Unmarshal([]byte(row.ReportData), &data); err != nil {
		return nil, pkgerrors.WithMessage(err, "parse playback qos report data failed")
	}
	report.EventType = data.EventType
	report.DurationMs = data.DurationMs
	report.PlayedMs = data.PlayedMs
	report.BufferedMs = data.BufferedMs
	report.StallCount = data.StallCount
	report.StallTotalMs = data.StallTotalMs
	report.Resolution = data.Resolution
	report.BitrateKbps = data.BitrateKbps
	report.Fps = data.Fps
	report.ErrorCode = data.ErrorCode
	report.ErrorMsg = data.ErrorMsg
	report.NetworkType = data.NetworkType
	report.DeviceInfo = data.DeviceInfo
	return report, nil
}
