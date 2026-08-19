package domain

import (
	"context"

	"go_zero-tiktok/pkg/contract"
)

type UserFollowService struct {
	followRepo IUserFollowRepo
}

func NewUserFollowService(followRepo IUserFollowRepo) *UserFollowService {
	return &UserFollowService{
		followRepo: followRepo,
	}
}

// GetFansList 获取粉丝 ID 列表（gateway 负责调 user rpc 水合用户详情）。
func (s *UserFollowService) GetFansList(ctx context.Context, userID int64, pageNum, pageSize int32) ([]int64, int64, error) {
	relations, total, err := s.followRepo.GetFansByUserID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return extractUserIDs(relations, func(r types.UserFollow) int64 { return r.FollowerID }), total, nil
}

// GetSubscriberList 获取关注 ID 列表（gateway 负责调 user rpc 水合用户详情）。
func (s *UserFollowService) GetSubscriberList(ctx context.Context, userID int64, pageNum, pageSize int32) ([]int64, int64, error) {
	relations, total, err := s.followRepo.GetFollowingByFollowerID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return extractUserIDs(relations, func(r types.UserFollow) int64 { return r.UserID }), total, nil
}

// GetFriendList 获取好友 ID 列表（gateway 负责调 user rpc 水合用户详情）。
func (s *UserFollowService) GetFriendList(ctx context.Context, userID int64, pageNum, pageSize int32) ([]int64, int64, error) {
	relations, total, err := s.followRepo.GetFriendByUserID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return extractUserIDs(relations, func(r types.UserFollow) int64 { return r.UserID }), total, nil
}

// FollowUser 关注用户
func (s *UserFollowService) FollowUser(ctx context.Context, followerID, userID int64) error {
	return s.followRepo.FollowUser(ctx, followerID, userID)
}

// UnfollowUser 取消关注
func (s *UserFollowService) UnfollowUser(ctx context.Context, followerID, userID int64) error {
	return s.followRepo.UnfollowUser(ctx, followerID, userID)
}

func extractUserIDs(relations []types.UserFollow, pickID func(types.UserFollow) int64) []int64 {
	ids := make([]int64, 0, len(relations))
	for _, relation := range relations {
		ids = append(ids, pickID(relation))
	}
	return ids
}
