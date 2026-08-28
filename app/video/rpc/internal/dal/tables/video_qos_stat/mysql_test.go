package video_qos_stat

import (
	"context"
	"testing"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/testhelpers"
)

func TestUpdateQoSAggregates(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	metrics := types.VideoQoSMetrics{
		CompletionRate: 8500,
		StallRate:      120,
		ErrorRate:      5,
		AvgBitrateKbps: 2500,
		AvgBufferedMs:  300,
		AvgStallCount:  1,
		ReportCount:    100,
	}

	testhelpers.AssertNoErr(t, UpdateQoSAggregates(ctx, db, 1, metrics))

	got, err := GetQoSMetricsByVideoIDs(ctx, db, []int64{1})
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, got[1].CompletionRate, metrics.CompletionRate)
	testhelpers.AssertEqual(t, got[1].StallRate, metrics.StallRate)
	testhelpers.AssertEqual(t, got[1].AvgBitrateKbps, metrics.AvgBitrateKbps)
}

func TestGetQoSMetricsByVideoIDs_Empty(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	got, err := GetQoSMetricsByVideoIDs(ctx, db, []int64{})
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(got)), int64(0))
}

func TestGetQoSMetricsByVideoIDs_Missing(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	got, err := GetQoSMetricsByVideoIDs(ctx, db, []int64{999})
	testhelpers.AssertNoErr(t, err)
	if _, ok := got[999]; ok {
		t.Fatalf("expected missing video to not be present")
	}
}
