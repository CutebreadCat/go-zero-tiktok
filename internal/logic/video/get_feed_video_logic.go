// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFeedVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedVideoLogic {
	return &GetFeedVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFeedVideoLogic) GetFeedVideo(req *types.FeedVideoRequest) (resp *types.FeedVideoResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
