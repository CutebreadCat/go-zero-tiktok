// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"
	"strings"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentVideoLogic {
	return &CommentVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentVideoLogic) CommentVideo(req *types.CommentVideoRequest) (resp *types.CommentVideoResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.VideoID == "" {
		return nil, xerr.New(400, "视频ID不能为空")
	}
	commentText := strings.TrimSpace(req.CommentText)
	if commentText == "" {
		return nil, xerr.New(400, "评论内容不能为空")
	}

	comment := &types.CommentBaseinfo{
		CommentID: myutils.GenerateCommentID(),
		UserID:    userID,
		VideoID:   req.VideoID,
		Content:   commentText,
	}

	if err := l.svcCtx.Dal.Comment.CreateComment(l.ctx, comment); err != nil {
		return nil, err
	}

	resp = &types.CommentVideoResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "评论发布成功"},
		CommentID: comment.CommentID,
	}

	return resp, nil
}
