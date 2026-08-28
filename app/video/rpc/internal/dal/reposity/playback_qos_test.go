package reposity

import (
	"context"
	"testing"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/testhelpers"
)

func TestPlaybackQoSRepo_CreateReport_Idempotent(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	repo := NewPlaybackQoSRepo(db)
	ctx := context.Background()

	report := &types.PlaybackQoSReport{
		UserID:         1,
		VideoID:        100,
		IdempotencyKey: "idem-1",
		EventType:      "complete",
		DurationMs:     1000,
	}

	testhelpers.AssertNoErr(t, repo.CreateReport(ctx, report))

	// 同一 (user_id, idempotency_key) 再次上报应幂等成功
	report.VideoID = 200
	testhelpers.AssertNoErr(t, repo.CreateReport(ctx, report))

	// 数据库应只有一条记录
	rows, err := db.Table("playback_qos_reports").Where("user_id = ? AND idempotency_key = ?", 1, "idem-1").Rows()
	testhelpers.AssertNoErr(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	testhelpers.AssertEqual(t, int64(count), int64(1))
}
