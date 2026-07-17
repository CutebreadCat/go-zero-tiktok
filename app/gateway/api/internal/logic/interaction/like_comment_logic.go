package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb/interaction_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCommentLogic {
	return &LikeCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeCommentLogic) LikeComment(req *types.LikeCommentRequest) (resp *types.LikeCommentResponse, err error) {
	userId, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户未登录或登录已过期")
	}
	if userId == "" {
		return nil, xerr.NewUnauthorized("用户未登录或登录已过期")
	}
	if req.Liketype != 1 && req.Liketype != 0 {
		return nil, xerr.NewInvalidParam("无效的点赞类型")
	}

	_, err = l.svcCtx.InteractionRpc.LikeComment(l.ctx, &interactionpb.LikeCommentRequest{
		CommentId: req.CommentID,
		UserId:    userId,
		LikeType:  req.Liketype,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.LikeCommentResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "操作成功",
		},
	}
	return resp, nil
}
