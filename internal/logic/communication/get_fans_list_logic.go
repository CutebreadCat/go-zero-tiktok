package communication

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
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

	relations, total, err := l.svcCtx.Dal.UserFollow.GetFansByUserID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, err
	}

	fansIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		fansIDs = append(fansIDs, relation.FollowerID)
	}

	fansList, err := l.svcCtx.Dal.User.GetUsersByIDs(l.ctx, fansIDs)
	if err != nil {
		return nil, err
	}

	resp = &types.GetFansListResponse{
		BaseResponse: types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		FansList:     l.svcCtx.Dal.User.UsersToResponse(fansList),
		FansCount:    int32(total),
	}

	return resp, nil
}
