// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package video

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VideoSearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoSearchLogic {
	return &VideoSearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VideoSearchLogic) VideoSearch(req *types.VideoSearchRequest) (resp *types.VideoSearchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
