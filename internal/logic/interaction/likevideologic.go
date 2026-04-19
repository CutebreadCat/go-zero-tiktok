// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"
	"errors"

	"go_zero-tiktok/internal/dal"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeVideoLogic {
	return &LikeVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeVideoLogic) LikeVideo(req *types.LikeVideoRequest) (resp *types.LikeVideoResponse, err error) {
	userID, err := getUserIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	switch req.ActionType {
	case 1:
		err = dal.LikeVideo(l.ctx, userID, req.VideoID)
		if err == nil {
			err = dal.UpdateVideoLikeCount(l.ctx, req.VideoID, 1)
		}
	case 2:
		err = dal.CancelLikeVideo(l.ctx, userID, req.VideoID)
		if err == nil {
			err = dal.UpdateVideoLikeCount(l.ctx, req.VideoID, -1)
		}
	default:
		err = errors.New("invalid action_type")
	}
	if err != nil {
		return nil, err
	}

	resp = &types.LikeVideoResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
	}

	return
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	keys := []string{"user_id", "userId", "uid", "UserID"}
	for _, key := range keys {
		if v := ctx.Value(key); v != nil {
			if uid, ok := v.(string); ok && uid != "" {
				return uid, nil
			}
		}
	}

	return "", errors.New("user id not found in context")
}
