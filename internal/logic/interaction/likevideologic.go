// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeVideoLogic {
	return &LikeVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeVideoLogic) LikeVideo(req *types.LikeVideoRequest) (resp *types.LikeVideoResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	switch req.ActionType {
	case 1:
		err = l.svcCtx.Dal.VideoLiker.LikeVideo(l.ctx, userID, req.VideoID)
		if err == nil {
			err = l.svcCtx.Dal.Popular.UpdateVideoLikeCount(l.ctx, req.VideoID, 1)
		}
	case 2:
		err = l.svcCtx.Dal.VideoLiker.CancelLikeVideo(l.ctx, userID, req.VideoID)
		if err == nil {
			err = l.svcCtx.Dal.Popular.UpdateVideoLikeCount(l.ctx, req.VideoID, -1)
		}
	default:
		err = xerr.New(400, "操作类型无效，仅支持1(点赞)或2(取消点赞)")
	}
	if err != nil {
		return nil, err
	}

	resp = &types.LikeVideoResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
	}

	return
}
