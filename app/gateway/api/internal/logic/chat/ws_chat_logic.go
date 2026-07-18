// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"

	logger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/app/gateway/api/internal/svc"
)

type WsChatLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWsChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WsChatLogic {
	return &WsChatLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *WsChatLogic) WsChat() error {
	// todo: add your logic here and delete this line

	return nil
}
