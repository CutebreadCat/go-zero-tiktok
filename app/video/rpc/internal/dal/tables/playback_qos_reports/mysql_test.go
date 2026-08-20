package playback_qos_reports

import (
	"context"
	"testing"

	"go_zero-tiktok/testhelpers"
)

func newReport(userID, videoID int64, idempotencyKey string) *PlaybackQoSReport {
	return &PlaybackQoSReport{
		UserID:         userID,
		VideoID:        videoID,
		ReportData:     `{"event_type":"complete","duration_ms":1000}`,
		IdempotencyKey: idempotencyKey,
	}
}

// TestCreateReport 正常写入一条上报记录
func TestCreateReport(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, CreateReport(ctx, db, newReport(1, 100, "key-1")))

	reports, err := GetReportsByVideoID(ctx, db, 100, 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(reports)), int64(1))
	testhelpers.AssertEqual(t, reports[0].IdempotencyKey, "key-1")
}

// TestCreateReport_DuplicateIdempotencyKey 相同 (user_id, idempotency_key) 重复写入应被数据库拒绝
func TestCreateReport_DuplicateIdempotencyKey(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	testhelpers.AssertNoErr(t, CreateReport(ctx, db, newReport(1, 100, "key-1")))

	err := CreateReport(ctx, db, newReport(1, 200, "key-1"))
	if err == nil {
		t.Fatalf("expected duplicate key error, got nil")
	}
}

// TestGetReportsByVideoID 按 video_id 分页查询
func TestGetReportsByVideoID(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	for i := int64(1); i <= 3; i++ {
		testhelpers.AssertNoErr(t, CreateReport(ctx, db, newReport(i, 100, "key-"+string(rune('0'+i)))))
	}

	reports, err := GetReportsByVideoID(ctx, db, 100, 1, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(reports)), int64(2))
}
