// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendListLogic) GetFriendList(req *types.GetFriendListRequest) (resp *types.GetFriendListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}

	relations, total, err := l.svcCtx.Dal.UserFollow.GetFriendByUserID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, xerr.New(1002, "获取好友列表失败，请稍后重试")
	}

	friendIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		friendIDs = append(friendIDs, relation.FollowerID)
	}

	friendList, err := l.svcCtx.Dal.User.GetUsersByIDs(l.ctx, friendIDs)
	if err != nil {
		return nil, xerr.New(1002, "获取好友信息失败，请稍后重试")
	}

	resp = &types.GetFriendListResponse{
		BaseResponse: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FriendList:   l.svcCtx.Dal.User.UsersToResponse(friendList),
		FriendCount:  int32(total),
	}

	return resp, nil
}
