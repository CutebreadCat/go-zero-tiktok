package communication

import (
	"context"

	communicationpb "go_zero-tiktok/app/communication/rpc/communication_pb"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"
	userpb "go_zero-tiktok/app/user/rpc/user_pb"

	logger "go_zero-tiktok/pkg/logger"
)

type GetFansListLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFansListLogic {
	return &GetFansListLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetFansListLogic) GetFansList(req *types.GetFansListRequest) (resp *types.GetFansListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	rpcResp, err := l.svcCtx.CommunicationRpc.GetFansList(l.ctx, &communicationpb.GetFansListRequest{
		UserId:   userID,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetFansList.GetFansList")
	}

	fansList := l.hydrateUsers(rpcResp.UserIds)

	return &types.GetFansListResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FansList:  fansList,
		FansCount: rpcResp.Total,
	}, nil
}

func (l *GetFansListLogic) hydrateUsers(userIDs []int64) []types.UserBaseinfo {
	if len(userIDs) == 0 {
		return []types.UserBaseinfo{}
	}

	userResp, err := l.svcCtx.UserRpc.BatchGetUserInfo(l.ctx, &userpb.BatchGetUserInfoRequest{
		UserIds: userIDs,
	})
	if err != nil {
		// 用户 RPC 失败时降级返回 ID 列表，避免关系流完全不可用。
		l.ContextLogger.Errorf("BatchGetUserInfo failed, userIDs=%v, err=%v", userIDs, err)
		users := make([]types.UserBaseinfo, 0, len(userIDs))
		for _, id := range userIDs {
			users = append(users, types.UserBaseinfo{UserID: id})
		}
		return users
	}

	users := make([]types.UserBaseinfo, 0, len(userResp.Users))
	for _, u := range userResp.Users {
		users = append(users, types.UserBaseinfo{
			UserID:   u.UserId,
			Username: u.Username,
			PhotoURL: u.PhotoUrl,
		})
	}
	return users
}
