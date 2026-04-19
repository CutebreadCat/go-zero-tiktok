// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubscribeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeLogic {
	return &SubscribeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubscribeLogic) Subscribe(req *types.SubscribeRequest) (resp *types.SubscribeResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
