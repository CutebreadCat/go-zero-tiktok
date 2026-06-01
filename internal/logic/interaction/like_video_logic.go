package interaction

import (
	"context"

	videopb "go_zero-tiktok/app/video/rpc/video_pb/video_pb"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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

	_, err = l.svcCtx.VideoRpc.LikeVideo(l.ctx, &videopb.LikeVideoRequest{
		UserId:     userID,
		VideoId:    req.VideoID,
		ActionType: req.ActionType,
	})
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
