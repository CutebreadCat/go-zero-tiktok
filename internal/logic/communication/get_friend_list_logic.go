package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/shared/xerr"
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
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}

	relations, total, err := l.svcCtx.Dal.UserFollow.GetFriendByUserID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, err
	}

	friendIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		friendIDs = append(friendIDs, relation.FollowerID)
	}

	friendList, err := l.svcCtx.Dal.User.GetUsersByIDs(l.ctx, friendIDs)
	if err != nil {
		return nil, err
	}

	resp = &types.GetFriendListResponse{
		BaseResponse: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FriendList:   friendList,
		FriendCount:  int32(total),
	}

	return resp, nil
}
