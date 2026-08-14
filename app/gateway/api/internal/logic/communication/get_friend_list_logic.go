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

type GetFriendListLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetFriendListLogic) GetFriendList(req *types.GetFriendListRequest) (resp *types.GetFriendListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	rpcResp, err := l.svcCtx.CommunicationRpc.GetFriendList(l.ctx, &communicationpb.GetFriendListRequest{
		UserId:   userID,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetFriendList.GetFriendList")
	}

	friendList := l.hydrateUsers(rpcResp.UserIds)

	return &types.GetFriendListResponse{
		Base:        types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FriendList:  friendList,
		FriendCount: rpcResp.Total,
	}, nil
}

func (l *GetFriendListLogic) hydrateUsers(userIDs []int64) []types.UserBaseinfo {
	if len(userIDs) == 0 {
		return []types.UserBaseinfo{}
	}

	userResp, err := l.svcCtx.UserRpc.BatchGetUserInfo(l.ctx, &userpb.BatchGetUserInfoRequest{
		UserIds: userIDs,
	})
	if err != nil {
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
