package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentPareantCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentPareantCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentPareantCommentLogic {
	return &CommentPareantCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentPareantCommentLogic) CommentPareantComment(req *types.CommentPareantCommentRequest) (resp *types.CommentPareantCommentResponse, err error) {
	UserId, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户未登录或登录已过期")
	}

	var CommentId string
	if CommentId, err = l.svcCtx.Dal.Comment.CommentParentComent(l.ctx, UserId, req.CommentText, req.ParentCommentID); err != nil {
		return nil, err
	}
	resp = &types.CommentPareantCommentResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "评论发布成功",
		},
		CommentID: CommentId,
	}
	return resp, nil
}
