package interaction

import (
	"context"

	videodomain "go_zero-tiktok/app/interaction/rpc/internal/domain"
	"go_zero-tiktok/pkg/kafka"
	appLogger "go_zero-tiktok/pkg/logger"
)

// LikeEventHandler 消费 Kafka 互动事件并持久化到 MySQL。
// 注意：like 关系与计数已在请求路径写入 Redis，因此 video-rpc 内部 consumer
// 只负责把事件最终落库，不再更新 Redis；同时保留日志点供消息中心/推荐权重使用。
type LikeEventHandler struct {
	interactionRepo videodomain.IVideoInteractionRepo
}

func NewLikeEventHandler(interactionRepo videodomain.IVideoInteractionRepo) *LikeEventHandler {
	return &LikeEventHandler{interactionRepo: interactionRepo}
}

func (h *LikeEventHandler) Consume(ctx context.Context, event *kafka.Event) error {
	e, err := LikeEventFromKafkaEvent(event)
	if err != nil {
		appLogger.Warnf("LikeEventHandler parse event failed: %v", err)
		return nil
	}

	return h.interactionRepo.ApplyLikeEvent(ctx, string(e.Action), e.UserID, e.VideoID)
}

// Compile-time check: LikeEventHandler 实现 kafka.ConsumerHandler。
var _ kafka.ConsumerHandler = (*LikeEventHandler)(nil)
