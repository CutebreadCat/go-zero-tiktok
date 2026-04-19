// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublishVideoLogic) PublishVideo(req *types.PublishVideoRequest) (resp *types.PublishVideoResponse, err error) {
	authorID, err := getVideoUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	video := &types.VideoBaseinfo{
		VideoID:     myutils.GenerateVideoID(),
		AuthorID:    authorID,
		VideoURL:    "",
		CoverURL:    "",
		Title:       req.Title,
		Description: req.Description,
	}

	if err := dal.CreateVideo(l.ctx, video); err != nil {
		return nil, err
	}

	resp = &types.PublishVideoResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		VideoID: video.VideoID,
	}

	return
}

func getVideoUserIDFromContext(ctx context.Context) (string, error) {
	keys := []string{"user_id", "userId", "uid", "UserID"}
	for _, key := range keys {
		if v := ctx.Value(key); v != nil {
			if uid, ok := v.(string); ok && uid != "" {
				return uid, nil
			}
		}
	}

	return "", errors.New("user id not found in context")
}
