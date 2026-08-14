package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetLikeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetLikeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeListLogic {
	return &GetLikeListLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetLikeListLogic) GetLikeList(in *interaction_pb.GetLikeListRequest) (*interaction_pb.GetLikeListResponse, error) {
	videoIDs, total, err := l.svcCtx.InteractionService.GetLikedVideoIDs(l.ctx, in.UserId, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetLikeList")
	}

	return &interaction_pb.GetLikeListResponse{
		VideoIds: videoIDs,
		Total:    total,
	}, nil
}
