package interaction

import (
	"context"
	"strings"
	"time"

	"fmt"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

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
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.VideoID == "" {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}
	commentText := strings.TrimSpace(req.CommentText)
	if commentText == "" {
		return nil, xerr.NewInvalidParam("评论内容不能为空")
	}

	commentID := myutils.GenerateCommentID()
	if err := l.svcCtx.Dal.Comment.CreateCommentFromParams(l.ctx, commentID, userID, req.VideoID, commentText, ""); err != nil {
		return nil, xerr.HandleDaoError(err, "CommentVideo.CreateComment")
	}

	resp = &types.CommentVideoResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "评论发布成功"},
		CommentID: commentID,
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := l.svcCtx.Dal.Popular.IncreaseVideoVisitCount(ctx, req.VideoID, 1); err != nil {
			fmt.Printf("increment visit count failed for video %s: %v\n", req.VideoID, err)
		}
	}()

	return resp, nil
}
