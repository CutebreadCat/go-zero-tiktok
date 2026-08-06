package logic

import (
	"context"
	"strings"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type CommentVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewCommentVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentVideoLogic {
	return &CommentVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *CommentVideoLogic) CommentVideo(in *interaction_pb.CommentVideoRequest) (*interaction_pb.CommentVideoResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}
	commentText := strings.TrimSpace(in.CommentText)
	if commentText == "" {
		return nil, xerr.NewInvalidParam("评论内容不能为空")
	}

	commentID := myutils.GenerateCommentID()
	if err := l.svcCtx.CommentService.CreateComment(l.ctx, commentID, in.UserId, in.VideoId, commentText, 0); err != nil {
		return nil, xerr.HandleDaoError(err, "CommentVideo.CreateComment")
	}

	return &interaction_pb.CommentVideoResponse{
		CommentId: commentID,
	}, nil
}
