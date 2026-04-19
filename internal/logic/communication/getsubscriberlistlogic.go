// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubscriberListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubscriberListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubscriberListLogic {
	return &GetSubscriberListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubscriberListLogic) GetSubscriberList(req *types.GetSubscriberListRequest) (resp *types.GetSubscriberListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
