package communication

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type CountUnreadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCountUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CountUnreadLogic {
	return &CountUnreadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CountUnreadLogic) CountUnread(req *types.CountUnreadRequest) (resp *types.CountUnreadResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	result, err := l.svcCtx.CommunicationRpc.CountUnread(l.ctx, &communication_pb.CountUnreadRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "CountUnread")
	}

	return &types.CountUnreadResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		UnreadCount: result.UnreadCount,
	}, nil
}
