package logic

import (
	"context"
	"strings"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type ReplyCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewReplyCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplyCommentLogic {
	return &ReplyCommentLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *ReplyCommentLogic) ReplyComment(in *interaction_pb.ReplyCommentRequest) (*interaction_pb.ReplyCommentResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	commentText := strings.TrimSpace(in.CommentText)
	if commentText == "" {
		return nil, xerr.NewInvalidParam("评论内容不能为空")
	}
	if in.ParentCommentId == 0 {
		return nil, xerr.NewInvalidParam("父评论ID不能为空")
	}

	commentID, err := l.svcCtx.CommentService.ReplyParentComment(l.ctx, in.UserId, commentText, in.ParentCommentId)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "ReplyComment")
	}

	return &interaction_pb.ReplyCommentResponse{
		CommentId: commentID,
	}, nil
}
