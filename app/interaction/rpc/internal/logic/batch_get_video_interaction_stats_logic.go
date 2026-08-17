package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	interactiondomain "go_zero-tiktok/app/interaction/rpc/internal/domain/interaction"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type BatchGetVideoInteractionStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewBatchGetVideoInteractionStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetVideoInteractionStatsLogic {
	return &BatchGetVideoInteractionStatsLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *BatchGetVideoInteractionStatsLogic) BatchGetVideoInteractionStats(in *interaction_pb.BatchGetVideoInteractionStatsRequest) (*interaction_pb.BatchGetVideoInteractionStatsResponse, error) {
	if len(in.VideoIds) == 0 {
		return &interaction_pb.BatchGetVideoInteractionStatsResponse{Stats: []*interaction_pb.VideoInteractionStat{}}, nil
	}

	statsMap, err := l.svcCtx.InteractionService.BatchGetVideoInteractionStats(l.ctx, in.UserId, in.VideoIds)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "BatchGetVideoInteractionStats")
	}

	stats := make([]*interaction_pb.VideoInteractionStat, 0, len(in.VideoIds))
	for _, videoID := range in.VideoIds {
		s := statsMap[videoID]
		if s == nil {
			s = &interactiondomain.VideoInteractionStat{VideoID: videoID}
		}
		stats = append(stats, &interaction_pb.VideoInteractionStat{
			VideoId:       videoID,
			LikeCount:     s.LikeCount,
			FavoriteCount: s.FavoriteCount,
			CommentCount:  s.CommentCount,
			Liked:         s.Liked,
			Favorited:     s.Favorited,
		})
	}

	return &interaction_pb.BatchGetVideoInteractionStatsResponse{Stats: stats}, nil
}
