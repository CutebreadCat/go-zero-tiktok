package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
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

	likeCounts, err := l.svcCtx.InteractionService.GetLikeCounts(l.ctx, in.VideoIds)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "BatchGetVideoInteractionStats.GetLikeCounts")
	}

	favoriteCounts, err := l.svcCtx.InteractionService.GetFavoriteCounts(l.ctx, in.VideoIds)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "BatchGetVideoInteractionStats.GetFavoriteCounts")
	}

	stats := make([]*interaction_pb.VideoInteractionStat, 0, len(in.VideoIds))
	for _, videoID := range in.VideoIds {
		stats = append(stats, &interaction_pb.VideoInteractionStat{
			VideoId:       videoID,
			LikeCount:     likeCounts[videoID],
			FavoriteCount: favoriteCounts[videoID],
		})
	}

	return &interaction_pb.BatchGetVideoInteractionStatsResponse{Stats: stats}, nil
}
