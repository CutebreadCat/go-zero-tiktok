package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetFavoriteListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetFavoriteListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFavoriteListLogic {
	return &GetFavoriteListLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetFavoriteListLogic) GetFavoriteList(in *interaction_pb.GetFavoriteListRequest) (*interaction_pb.GetFavoriteListResponse, error) {
	videoIDs, total, err := l.svcCtx.InteractionService.GetFavoritedVideoIDs(l.ctx, in.UserId, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetFavoriteList")
	}

	return &interaction_pb.GetFavoriteListResponse{
		VideoIds: videoIDs,
		Total:    total,
	}, nil
}
