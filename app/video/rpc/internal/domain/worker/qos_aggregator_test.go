package worker

import (
	"testing"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/testhelpers"
)

func TestCalcVideoQoSMetrics(t *testing.T) {
	reports := []*types.PlaybackQoSReport{
		{VideoID: 1, EventType: "complete", DurationMs: 10000, PlayedMs: 10000, BitrateKbps: 2000, BufferedMs: 100, StallCount: 0, ErrorCode: 0},
		{VideoID: 1, EventType: "complete", DurationMs: 10000, PlayedMs: 5000, BitrateKbps: 3000, BufferedMs: 200, StallCount: 1, ErrorCode: 0},
		{VideoID: 1, EventType: "error", DurationMs: 0, PlayedMs: 0, BitrateKbps: 0, BufferedMs: 0, StallCount: 0, ErrorCode: 500},
	}

	m := calcVideoQoSMetrics(reports)

	// completion = (1 + 0.5) / 2 = 0.75 -> 7500
	testhelpers.AssertEqual(t, m.CompletionRate, int32(7500))
	// stall rate = 1/3 -> 3333
	testhelpers.AssertEqual(t, m.StallRate, int32(3333))
	// error rate = 1/3 -> 3333
	testhelpers.AssertEqual(t, m.ErrorRate, int32(3333))
	// avg bitrate = (2000+3000)/2 = 2500
	testhelpers.AssertEqual(t, m.AvgBitrateKbps, int32(2500))
	// avg buffered = (100+200+0)/3 = 100
	testhelpers.AssertEqual(t, m.AvgBufferedMs, int64(100))
	// avg stall count = (0+1+0)/3 = 0
	testhelpers.AssertEqual(t, m.AvgStallCount, int32(0))
	testhelpers.AssertEqual(t, m.ReportCount, int64(3))
}

func TestCalcVideoQoSMetrics_Empty(t *testing.T) {
	m := calcVideoQoSMetrics(nil)
	testhelpers.AssertEqual(t, m.ReportCount, int64(0))
}
