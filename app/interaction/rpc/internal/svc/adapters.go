package svc

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/internal/config"
	"go_zero-tiktok/app/video/rpc/videoservice"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

// VideoVisitAdapter 视频访问量适配器（通过 video RPC 累加访问量）
type VideoVisitAdapter struct {
	videoRpc videoservice.VideoService
}

func NewVideoVisitAdapter(c config.Config) *VideoVisitAdapter {
	return &VideoVisitAdapter{
		videoRpc: videoservice.NewVideoService(zrpc.MustNewClient(c.VideoRpc)),
	}
}

func (a *VideoVisitAdapter) IncreaseVideoVisitCount(ctx context.Context, videoID int64, delta int64) error {
	_, err := a.videoRpc.IncreaseVideoVisitCount(ctx, &videoservice.IncreaseVideoVisitCountRequest{
		VideoId: videoID,
		Delta:   delta,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("IncreaseVideoVisitCount failed, videoID=%d, delta=%d, err=%v", videoID, delta, err)
	}
	return err
}
