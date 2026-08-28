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

type MarkReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkReadLogic) MarkRead(req *types.MarkReadRequest) (resp *types.MarkReadResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	result, err := l.svcCtx.CommunicationRpc.MarkRead(l.ctx, &communication_pb.MarkReadRequest{
		UserId:     userID,
		MessageIds: req.MessageIDs,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "MarkRead")
	}

	return &types.MarkReadResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		UpdatedCount: result.UpdatedCount,
	}, nil
}
