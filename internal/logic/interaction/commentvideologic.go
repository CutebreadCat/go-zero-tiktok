// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"
	"strings"
	"time"

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
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
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
		CreatedAt: myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05"),
		UpdatedAt: myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05"),
		DeletedAt: myutils.TsToStr(time.Now().Unix(), "2006-01-02 15:04:05"),
	}

	if err := l.svcCtx.Dal.Comment.CreateComment(l.ctx, comment); err != nil {
		return nil, xerr.New(1002, "发布评论失败，请稍后重试")
	}

	resp = &types.CommentVideoResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "评论发布成功"},
		CommentID: comment.CommentID,
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := l.svcCtx.Dal.Popular.IncreaseVideoVisitCount(ctx, req.VideoID, 1); err != nil {
			logx.Errorf("increment visit count failed for video %s: %v", req.VideoID, err)
		}

	}()

	return resp, nil
}
