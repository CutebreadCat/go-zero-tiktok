// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"go_zero-tiktok/app/gateway/api/internal/svc"
)

type WsChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWsChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WsChatLogic {
	return &WsChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WsChatLogic) WsChat() error {
	// todo: add your logic here and delete this line

	return nil
}
