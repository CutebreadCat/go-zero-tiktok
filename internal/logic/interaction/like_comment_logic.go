package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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
	if err = l.svcCtx.Dal.Comment.LikeComment(l.ctx, req.CommentID, userId, req.Liketype); err != nil {
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
