package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCommentLogic) DeleteComment(in *interaction_pb.DeleteCommentRequest) (*interaction_pb.DeleteCommentResponse, error) {
	if in.CommentId == "" {
		return nil, xerr.NewInvalidParam("评论ID不能为空")
	}
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}

	if err := l.svcCtx.CommentService.DeleteComment(l.ctx, in.CommentId, in.UserId); err != nil {
		return nil, xerr.HandleDaoError(err, "DeleteComment")
	}

	return &interaction_pb.DeleteCommentResponse{}, nil
}
