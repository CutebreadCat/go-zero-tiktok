// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentRequest) (resp *types.DeleteCommentResponse, err error) {
	if req.CommentID == "" {
		return nil, xerr.New(400, "评论ID不能为空")
	}

	if err := l.svcCtx.Dal.Comment.DeleteCommentByID(l.ctx, req.CommentID); err != nil {
		return nil, err
	}

	resp = &types.DeleteCommentResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "删除评论成功"},
	}

	return resp, nil
}
