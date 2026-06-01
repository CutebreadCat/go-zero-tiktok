package communication

import (
	"context"

	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb/communication_pb"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFansListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFansListLogic {
	return &GetFansListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFansListLogic) GetFansList(req *types.GetFansListRequest) (resp *types.GetFansListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	rpcResp, err := l.svcCtx.CommunicationRpc.GetFansList(l.ctx, &communicationpb.GetFansListRequest{
		UserId:   userID,
		PageNum:  req.PageNumber,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	fansList := make([]types.UserBaseinfo, 0, len(rpcResp.Users))
	for _, u := range rpcResp.Users {
		fansList = append(fansList, types.UserBaseinfo{
			UserID:   u.UserId,
			Username: u.Username,
			PhotoURL: u.PhotoUrl,
		})
	}

	resp = &types.GetFansListResponse{
		BaseResponse: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FansList:     fansList,
		FansCount:    int32(rpcResp.Total),
	}

	return resp, nil
}
