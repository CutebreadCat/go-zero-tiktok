// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *WsChatLogic) WsChat(req *types.WsChatRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
