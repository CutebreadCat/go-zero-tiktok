package logic

import (
	"context"

	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type GetCommentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetCommentListLogic) GetCommentList(in *interaction_pb.GetCommentListRequest) (*interaction_pb.GetCommentListResponse, error) {
	if in.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}
	if in.PageSize <= 0 || in.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	comments, total, err := l.svcCtx.CommentService.GetCommentsByVideoID(l.ctx, in.VideoId, in.PageNum, in.PageSize)
	if err != nil {
		return nil, xerr.Wrap(err, "GetCommentList")
	}

	commentInfos := make([]*interaction_pb.CommentInfo, 0, len(comments))
	for _, c := range comments {
		commentInfos = append(commentInfos, &interaction_pb.CommentInfo{
			CommentId:       c.CommentID,
			UserId:          c.UserID,
			VideoId:         c.VideoID,
			Content:         c.Content,
			ParentCommentId: c.ParentCommentID,
			CreatedAt:       c.CreatedAt,
		})
	}

	return &interaction_pb.GetCommentListResponse{
		Comments: commentInfos,
		Total:    total,
	}, nil
}
