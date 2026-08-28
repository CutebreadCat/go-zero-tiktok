package svc

import (
	"context"
	"time"

	communicationservice "go_zero-tiktok/app/communication/rpc/communicationservice"
	"go_zero-tiktok/app/gateway/api/internal/config"
	"go_zero-tiktok/app/gateway/api/internal/middleware"
	interactionservice "go_zero-tiktok/app/interaction/rpc/interactionservice"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/app/video/rpc/videoservice"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                 config.Config
	UserRpc                userservice.UserService
	VideoRpc               videoservice.VideoService
	InteractionRpc         interactionservice.InteractionService
	CommunicationRpc       communicationservice.CommunicationService
	RateLimit              rest.Middleware
	hotScoreRecalcProducer HotScoreRecalcProducer
	TrackingEventProducer  TrackingEventProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := &ServiceContext{
		Config:           c,
		UserRpc:          userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
		VideoRpc:         videoservice.NewVideoService(zrpc.MustNewClient(c.VideoRpc)),
		InteractionRpc:   interactionservice.NewInteractionService(zrpc.MustNewClient(c.InteractionRpc)),
		CommunicationRpc: communicationservice.NewCommunicationService(zrpc.MustNewClient(c.CommunicationRpc)),
		RateLimit:        middleware.NewRateLimitMiddleware().Handle,
	}

	// Kafka 热度分重算事件生产者：未启用或未配置时回退到同步 RPC。
	if c.Kafka.Enable && len(c.Kafka.Brokers) > 0 && c.Kafka.Brokers[0] != "" {
		producer := NewKafkaHotScoreRecalcProducer(c.Kafka.Brokers, c.Kafka.Topic)
		producer.WatchErrors()
		ctx.hotScoreRecalcProducer = producer

		trackingProducer := NewKafkaTrackingEventProducer(c.Kafka.Brokers, c.Kafka.TrackingTopic)
		trackingProducer.WatchErrors()
		ctx.TrackingEventProducer = trackingProducer
	}

	return ctx
}

// TriggerHotScoreRecalc 触发视频热度分重算（非关键路径）。
// 优先通过 Kafka 发送事件解耦 Gateway 与 video.rpc；Kafka 未启用时回退到同步 RPC。
// 由 gateway 在点赞/收藏/评论等互动事件成功后调用，失败仅记日志不影响主流程。
func (s *ServiceContext) TriggerHotScoreRecalc(ctx context.Context, videoID int64) {
	if videoID == 0 {
		return
	}

	if s.hotScoreRecalcProducer != nil {
		if err := s.hotScoreRecalcProducer.Send(ctx, videoID); err != nil {
			logx.Errorf("send hot score recalc event failed, video_id=%d: %v", videoID, err)
		}
		return
	}

	// 降级：Kafka 未启用时同步调用 video.rpc（仍不阻塞请求路径外的地方，但避免业务 goroutine）。
	callCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	if _, err := s.VideoRpc.RecalculateHotScore(callCtx, &videoservice.RecalculateHotScoreRequest{
		VideoId: videoID,
	}); err != nil {
		logx.Errorf("RecalculateHotScore sync fallback failed, video_id=%d: %v", videoID, err)
	}
}

// Close 优雅关闭 Gateway 侧资源。
func (s *ServiceContext) Close() {
	if s.hotScoreRecalcProducer != nil {
		_ = s.hotScoreRecalcProducer.Close()
	}
	if s.TrackingEventProducer != nil {
		_ = s.TrackingEventProducer.Close()
	}
}
