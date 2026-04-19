// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"go_zero-tiktok/internal/dal"

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
		return nil, err
	}
	if relationsIDPeople, total, err := dal.GetFriendByUserID(l.ctx, user_id, req.PageNumber, req.PageSize); err != nil {
		logx.Errorf("failed to get friend list: %v", err)
		return nil, err
	} else {
		var relationsID []string
		for _, relation := range relationsIDPeople {
			relationsID = append(relationsID, relation.FollowerID)
		}
		relations := make([]types.UserBaseinfo, 0, len(relationsID))
		if relations, err = dal.GetUsersByIDs(l.ctx, relationsID); err != nil {
			logx.Errorf("failed to get user base info: %v", err)
			return nil, err
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
