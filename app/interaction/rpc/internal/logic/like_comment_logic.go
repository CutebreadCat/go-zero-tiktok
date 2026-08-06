package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type LikeCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCommentLogic {
	return &LikeCommentLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *LikeCommentLogic) LikeComment(in *interaction_pb.LikeCommentRequest) (*interaction_pb.LikeCommentResponse, error) {
	if in.CommentId == 0 {
		return nil, xerr.NewInvalidParam("评论ID不能为空")
	}
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.LikeType != 0 && in.LikeType != 1 {
		return nil, xerr.NewInvalidParam("操作类型无效，仅支持1(点赞)或0(取消点赞)")
	}

	if err := l.svcCtx.CommentService.LikeComment(l.ctx, in.CommentId, in.UserId, in.LikeType); err != nil {
		return nil, xerr.HandleDaoError(err, "LikeComment")
	}

	return &interaction_pb.LikeCommentResponse{}, nil
}
