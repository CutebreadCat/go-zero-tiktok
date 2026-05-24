package interaction

import (
	"context"
	"fmt"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"
	"time"

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
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.VideoID == "" {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	switch req.ActionType {
	case 1:
		err = l.svcCtx.Dal.VideoLiker.LikeVideo(l.ctx, userID, req.VideoID)
		if err == nil {
			err = l.svcCtx.Dal.Popular.UpdateVideoLikeCount(l.ctx, req.VideoID, 1)
		}
	case 0:
		err = l.svcCtx.Dal.VideoLiker.CancelLikeVideo(l.ctx, userID, req.VideoID)
		if err == nil {
			err = l.svcCtx.Dal.Popular.UpdateVideoLikeCount(l.ctx, req.VideoID, -1)
		}
	default:
		return nil, xerr.NewInvalidParam("操作类型无效，仅支持1(点赞)或0(取消点赞)")
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
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := l.svcCtx.Dal.Popular.IncreaseVideoVisitCount(ctx, req.VideoID, 1); err != nil {
			fmt.Printf("increment visit count failed for video %s: %v\n", req.VideoID, err)
		}
	}()

	return
}
