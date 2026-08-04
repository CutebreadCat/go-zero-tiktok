package svc

import (
	"context"

	"go_zero-tiktok/app/communication/rpc/internal/config"
	"go_zero-tiktok/app/user/rpc/userservice"
	"go_zero-tiktok/pkg/contract"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

// UserRepoAdapter 用户仓储适配器（通过 user RPC 获取用户信息）
type UserRepoAdapter struct {
	userRpc userservice.UserService
}

func NewUserRepoAdapter(c config.Config) *UserRepoAdapter {
	return &UserRepoAdapter{
		userRpc: userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
	}
}

func (a *UserRepoAdapter) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]types.UserBaseinfo, error) {
	if len(userIDs) == 0 {
		return []types.UserBaseinfo{}, nil
	}

	resp, err := a.userRpc.BatchGetUserInfo(ctx, &userservice.BatchGetUserInfoRequest{
		UserIds: userIDs,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("BatchGetUserInfo failed, userIDs=%v, err=%v", userIDs, err)
		return nil, err
	}

	result := make([]types.UserBaseinfo, 0, len(resp.Users))
	for _, u := range resp.Users {
		result = append(result, types.UserBaseinfo{
			UserID:    u.UserId,
			Username:  u.Username,
			PhotoURL:  u.PhotoUrl,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}
	return result, nil
}
