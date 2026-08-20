package domain

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"
)

type fakePlaybackQoSRepo struct{}

func (f *fakePlaybackQoSRepo) CreateReport(ctx context.Context, report *types.PlaybackQoSReport) error {
	return nil
}

func TestPlaybackQoSService_ReportPlaybackQoS_Validation(t *testing.T) {
	svc := NewPlaybackQoSService(&fakePlaybackQoSRepo{})
	ctx := context.Background()

	tests := []struct {
		name    string
		report  *types.PlaybackQoSReport
		wantErr bool
	}{
		{
			name: "正常上报",
			report: &types.PlaybackQoSReport{
				UserID:         1,
				VideoID:        100,
				IdempotencyKey: "idem-1",
				EventType:      "complete",
			},
			wantErr: false,
		},
		{
			name: "video_id 无效",
			report: &types.PlaybackQoSReport{
				UserID:         1,
				VideoID:        0,
				IdempotencyKey: "idem-1",
				EventType:      "complete",
			},
			wantErr: true,
		},
		{
			name: "幂等键为空",
			report: &types.PlaybackQoSReport{
				UserID:    1,
				VideoID:   100,
				EventType: "complete",
			},
			wantErr: true,
		},
		{
			name: "事件类型无效",
			report: &types.PlaybackQoSReport{
				UserID:         1,
				VideoID:        100,
				IdempotencyKey: "idem-1",
				EventType:      "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ReportPlaybackQoS(ctx, tt.report)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr {
				var codeErr *xerr.CodeError
				if !errors.As(err, &codeErr) || codeErr.Code != xerr.InvalidParam {
					t.Fatalf("expected invalid param error, got %v", err)
				}
			}
		})
	}
}
