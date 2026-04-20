// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

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
	// todo: add your logic here and delete this line
	var user_id string
	if userID, ok := l.ctx.Value("user_id").(string); ok {
		user_id = userID
	} else {
		logx.Errorf("failed to get user_id from context")
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}
	if relationsIDPeople, total, err := l.svcCtx.Dal.UserFollow.GetFriendByUserID(l.ctx, user_id, req.PageNumber, req.PageSize); err != nil {
		logx.Errorf("failed to get friend list: %v", err)
		return nil, xerr.New(1002, "获取好友列表失败，请稍后重试")
	} else {
		var relationsID []string
		for _, relation := range relationsIDPeople {
			relationsID = append(relationsID, relation.FollowerID)
		}
		relations := make([]types.UserBaseinfo, 0, len(relationsID))
		if relations, err = l.svcCtx.Dal.User.GetUsersByIDs(l.ctx, relationsID); err != nil {
			logx.Errorf("failed to get user base info: %v", err)
			return nil, xerr.New(1002, "获取好友信息失败，请稍后重试")
		}
		resp = &types.GetFriendListResponse{
			BaseResponse: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
			FriendCount:  int32(total),
		}
		for i := range relations {
			resp.FriendList = append(resp.FriendList, relations[i])
		}
	}

	return
}
