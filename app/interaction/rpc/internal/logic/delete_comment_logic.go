package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *DeleteCommentLogic) DeleteComment(in *interaction_pb.DeleteCommentRequest) (*interaction_pb.DeleteCommentResponse, error) {
	if in.CommentId == 0 {
		return nil, xerr.NewInvalidParam("评论ID不能为空")
	}
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}

	if err := l.svcCtx.CommentService.DeleteComment(l.ctx, in.CommentId, in.UserId); err != nil {
		return nil, xerr.HandleDaoError(err, "DeleteComment")
	}

	return &interaction_pb.DeleteCommentResponse{}, nil
}
