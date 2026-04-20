// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLikeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLikeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeListLogic {
	return &GetLikeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLikeListLogic) GetLikeList(req *types.GetLikeListRequest) (resp *types.GetLikeListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	videoIDs, total, err := l.svcCtx.Dal.VideoLiker.GetLikedVideoIDsByUserID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, err
	}

	videos, err := l.svcCtx.Dal.Video.GetVideosByIDs(l.ctx, videoIDs)
	if err != nil {
		return nil, err
	}

	resp = &types.GetLikeListResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		VideoList: videos,
		LikeCount: int32(total),
	}

	return
}
